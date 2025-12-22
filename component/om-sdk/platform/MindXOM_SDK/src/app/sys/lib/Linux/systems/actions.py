# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import os
import shlex
import subprocess
import threading
import time
from typing import AnyStr, Dict, NoReturn

from bin.global_exclusive_control import GlobalExclusiveController
from common.constants import error_codes
from common.constants.base_constants import ResetActionConst
from common.constants.product_constants import REDFISH_RESTART_TYPE
from common.constants.upgrade_constants import UpgradeState
from common.exception.biz_exception import Exceptions
from common.init_cmd import cmd_constants
from common.log.logger import run_log
from common.utils.common_check import CommonCheck
from common.utils.exception_utils import ExceptionUtils
from common.utils.exec_cmd import ExecCmd
from common.checkers import IPV4Checker
from lib.Linux.systems import restore_defaults_action_clib
from lib.Linux.upgrade.upgrade_effect import UpgradeEffect
from lib.Linux.upgrade.upgrade_new import Upgrade
from common.common_methods import CommonMethods


class SystemAction:
    """
    功能描述：系统action
    接口：NA
    """
    EDGE_SYSTEM_ACTION_LOCK = GlobalExclusiveController()
    LOCK_TIMEOUT = 60
    RESTART_DELAY_TIME = 5
    GRACEFUL_RESTART_CMD = [cmd_constants.OS_CMD_SYSTEMCTL, "reboot"]
    FORCE_RESTART_CMD = [cmd_constants.OS_CMD_SYSTEMCTL, "reboot", "-f"]
    LOG_SAVE_CMD = ["/usr/local/scripts/backup_restart_log.sh"]

    @staticmethod
    def post_request(request_data_dict):
        if SystemAction.EDGE_SYSTEM_ACTION_LOCK.locked():
            run_log.error("Edge system action is busy")
            ret_msg = ResetActionConst.EFFECT_CODE_MAP.get(ResetActionConst.ERR_RESET_FAILED, "Unknown result")
            return [CommonMethods.ERROR, ret_msg]

        SystemAction.EDGE_SYSTEM_ACTION_LOCK.acquire(SystemAction.LOCK_TIMEOUT)
        action = request_data_dict.get("ResetType", None)
        if action not in REDFISH_RESTART_TYPE:
            ret_msg = ResetActionConst.EFFECT_CODE_MAP.get(ResetActionConst.ERR_RESET_FAILED, "Unknown result")
            SystemAction.EDGE_SYSTEM_ACTION_LOCK.release()
            return [CommonMethods.ERROR, ret_msg]

        run_log.info("Start to reset system by %s.", action)
        # 如果需要生效，则需要调用生效接口重启
        if Upgrade.allow_effect():
            try:
                UpgradeEffect().effect_firmware()
            except Exception as err:
                run_log.error("upgrade effect failed: %s", err)
                ret_msg = ResetActionConst.EFFECT_CODE_MAP.get(ResetActionConst.ERR_RESET_FAILED, "Unknown result")
                SystemAction.EDGE_SYSTEM_ACTION_LOCK.release()
                return [CommonMethods.ERROR, ret_msg]

        threading.Thread(target=SystemAction.graceful_restart).start()
        ret_msg = ResetActionConst.EFFECT_CODE_MAP.get(ResetActionConst.ERR_RESET_SUCCEED, "Unknown result")
        return [CommonMethods.OK, ret_msg]

    @classmethod
    def graceful_restart(cls):
        """
        注意：此操作一定会成功。
        """
        try:
            ExecCmd.exec_cmd(cls.LOG_SAVE_CMD)
            os.sync()
            time.sleep(cls.RESTART_DELAY_TIME)
            ExecCmd.exec_cmd(cls.GRACEFUL_RESTART_CMD)
        except Exception as err:
            run_log.error("reboot system failed, because unknown error %s", err)
        finally:
            SystemAction.EDGE_SYSTEM_ACTION_LOCK.release()


class RestoreDefaultsAction:
    """
    功能描述：恢复出厂设置
       接口：NA
    """
    RESTORE_LOCK = GlobalExclusiveController()
    IP_LIST = None
    LOCK_TIMEOUT = 120

    def __init__(self):
        """
         功能描述：初始化函数
         参数：
         返回值：无
         异常描述：NA
        """
        self.ip_addr_list = None

    @staticmethod
    def check_root_pwd(root_pwd: AnyStr) -> NoReturn:
        """
        用户输入的root密码
        :param root_pwd:
        :return: True:密码校验正确；False：密码校验错误
        """
        # 1、判断root密码标致位是否存在，如果存在则抛出异常
        if os.path.exists("/run/root_pwd_init"):
            raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_ROOT_PASS_WORD_FLAG_EXIST_CHECK)
        # 2、判断密码是否为空
        if not root_pwd:
            raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_ARGUMENT_VALUE_NOT_EXIST, "root_pwd")

        # 3、校验密码是否匹配
        if not isinstance(root_pwd, str) or len(root_pwd) > 20:
            raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_PARAMETER_INVALID)

        b_ok = restore_defaults_action_clib.authenticate("root", root_pwd)
        if not b_ok:
            raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_USER_AUTH_FAILED)

    @staticmethod
    def record_restore_log(user_name, user_ip):

        dev_p1 = "/dev/mmcblk0p1"
        dev_p1_mount_path = "/mnt/p1"
        restore_log_path = f"{dev_p1_mount_path}/om_restore_msg.log"

        # p1分区设为可写模式
        cmd = (cmd_constants.OS_CMD_BLOCKDEV, "--setrw", dev_p1)
        if ExecCmd.exec_cmd(cmd) != 0:
            run_log.error("exec cmd : %s failed", cmd)
            raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_CMD_EXEC_WRONG)

        # 挂载p1分区
        cmd = (cmd_constants.OS_CMD_MOUNT, dev_p1, dev_p1_mount_path)
        if ExecCmd.exec_cmd(cmd) != 0:
            run_log.error("exec cmd : %s failed", cmd)
            raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_CMD_EXEC_WRONG)

        # 记录日志
        log_msg = f'{[time.strftime("%Y-%m-%d %H:%M:%S")]} [{user_name}@{user_ip}] Restore defaults system success.\n'
        try:
            with os.fdopen(os.open(restore_log_path, os.O_WRONLY | os.O_CREAT, 0o600), "w") as file:
                file.write(log_msg)
        except Exception as err:
            run_log.error("write restore msg failed, %s", err)
            raise Exceptions.biz_exception(error_codes.FileErrorCodes.ERROR_FILE_OPEN_ERROR)

        # p1分区恢复只读模式
        cmd = (cmd_constants.OS_CMD_BLOCKDEV, "--setro", dev_p1)
        if ExecCmd.exec_cmd(cmd) != 0:
            run_log.error("exec cmd : %s failed", cmd)
            raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_CMD_EXEC_WRONG)

        # 解除p1挂载
        cmd = (cmd_constants.OS_CMD_UMOUNT, dev_p1)
        if ExecCmd.exec_cmd(cmd) != 0:
            run_log.error("exec cmd : %s failed", cmd)
            raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_CMD_EXEC_WRONG)

    @staticmethod
    def get_eth_ip_map() -> Dict[AnyStr, AnyStr]:
        """
        根据命令ifconfig获取eth与ip对应的dict集合
        :return:  返回集合列表
        """
        RestoreDefaultsAction.IP_LIST = []
        cmd_shell = f"{cmd_constants.OS_CMD_IFCONFIG} | grep flags -A 1 | grep ^eth -A 1 | grep -v '-' | awk " \
                    "'NR%2{printf \"%s \",$0;next;}1' | awk -F ' ' '{OFS=\"-\";print $1,$6}'" \
                    " | rev | sed -e 's/://' | rev"
        ret = ExecCmd.exec_cmd_use_pipe_symbol(cmd_shell, wait=30)
        if ret[0] != 0:
            raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_CMD_EXEC_WRONG, ret[1])
        eth_ip_list = ret[1].strip().split()
        eth_ip_map = {}
        for eth_ip in eth_ip_list:
            split_ip = eth_ip.split("-")
            eth = split_ip[0]
            ip_addr = split_ip[1]
            if not IPV4Checker("ip").check({"ip": ip_addr}):
                continue
            if "." in eth and ":" in eth:
                continue

            eth0_check = (eth == "eth0" and ip_addr == "192.168.2.111")
            eth1_check = (eth == "eth1" and ip_addr == "192.168.3.111")
            if eth0_check or eth1_check:
                # 如果IP地址为192.168.2.111、192.168.3.111并且在默认的网口上，则不在列表出现，web不需要使用
                continue
            eth_ip_map[eth] = ip_addr
            RestoreDefaultsAction.IP_LIST.append(ip_addr)
        return eth_ip_map

    @staticmethod
    def get_eth_gateway():
        """
        通过位与计算可恢复出厂的ip与默认网关是否匹配
        :return:  返回与默认网关不匹配的ip列表
        """
        import ipaddress
        no_gateway_ip_list = []
        for ip in RestoreDefaultsAction.IP_LIST:
            cmd_shell = "%s addr show | grep -w %s | awk -F ' ' '{ print $2}'" \
                        % (cmd_constants.OS_CMD_IP, shlex.quote(ip))
            ret = ExecCmd.exec_cmd_use_pipe_symbol(cmd_shell, wait=30)
            if ret[0] != 0:
                raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_CMD_EXEC_WRONG, ret[1])
            ip_network = ret[1].strip().split()[0]
            subnet_mask = ip_network.split("/")[1]
            gateway_list = RestoreDefaultsAction.getgateway()
            if not gateway_list:
                return no_gateway_ip_list
            ip_network_segment = ipaddress.ip_network(ip_network, strict=False).network_address
            gateway_segment_list = []
            for gateway in gateway_list:
                gateway_segment = ipaddress.ip_network('%s/%s' % (gateway, subnet_mask), strict=False).network_address
                gateway_segment_list.append(gateway_segment)
            if ip_network_segment in gateway_segment_list:
                continue
            no_gateway_ip_list.append(ip)
        return no_gateway_ip_list

    @staticmethod
    def getgateway():
        """
        根据命令ip route 获取 默认网关
        :return:  默认网关
        """
        cmd_shell = "%s route | grep '^default' | awk -F ' ' '{print $3}'" % cmd_constants.OS_CMD_IP
        ret = ExecCmd.exec_cmd_use_pipe_symbol(cmd_shell, wait=30)
        if ret[0] != 0:
            raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_CMD_EXEC_WRONG, ret[1])
        if ret[1] == "":
            return list()
        else:
            gateway_list = ret[1].strip().split("\n")
            return gateway_list

    def get_all_info(self) -> NoReturn:
        try:
            # 获取ip地址列表并返回
            self.ip_addr_list = RestoreDefaultsAction.get_eth_ip_map()
            self.noGatewayIP = RestoreDefaultsAction.get_eth_gateway()
        except Exception as ex:
            run_log.error("Get eth ip list failed.")
            return [CommonMethods.ERROR, ExceptionUtils.exception_process(ex)]
        return [CommonMethods.OK, ]

    def post_request(self, request_data_dict):
        if RestoreDefaultsAction.RESTORE_LOCK.locked():
            run_log.warning("Restore default is busy")
            return [400, [error_codes.CommonErrorCodes.ERROR_ROOT_OPERATE_LOCKED.code,
                          error_codes.CommonErrorCodes.ERROR_ROOT_OPERATE_LOCKED.messageKey]]

        RestoreDefaultsAction.RESTORE_LOCK.acquire(self.LOCK_TIMEOUT)
        # 1、获取网卡信息、root密码、用户名和IP
        root_pwd = request_data_dict.get("root_pwd")
        ethernet = request_data_dict.get("ethernet")
        user_name = request_data_dict.get("_User")
        user_ip = request_data_dict.get("_Xip")
        operator_check = CommonCheck.check_operator(user_name, user_ip)
        if not operator_check:
            run_log.error("The operator is illegal, %s", operator_check.error)
            RestoreDefaultsAction.RESTORE_LOCK.release()
            return [400, [error_codes.CommonErrorCodes.ERROR_ROOT_OPERATE_ILLEGAL.code,
                          error_codes.CommonErrorCodes.ERROR_ROOT_OPERATE_ILLEGAL.messageKey]]

        try:
            # 2、校验root密码是否合法
            RestoreDefaultsAction.check_root_pwd(root_pwd)
            # 3、恢复出厂
            self.restore_defaults(ethernet, user_name, user_ip)
        except Exception as ex:
            run_log.error("Restore default failed.")
            return [CommonMethods.ERROR, ExceptionUtils.exception_process(ex)]
        finally:
            RestoreDefaultsAction.RESTORE_LOCK.release()
        return [CommonMethods.OK, ]

    def restore_defaults(self, ethernet: str, user_name: str, user_ip: str) -> NoReturn:
        """
        调用脚本执行恢复出厂设置
        :param ethernet:要保留的网口
        :param user_name: 操作用户
        :param user_ip: 操作ip
        """
        # 要保留网口信息时，校验网口的合法性，并指定保留的网口信息
        cmd_shell = '%s restore_factory' % cmd_constants.OS_CMD_IES_TOOL
        if ethernet:
            eth_ip_map: Dict[AnyStr, AnyStr] = RestoreDefaultsAction.get_eth_ip_map()
            if not eth_ip_map.__contains__(ethernet):
                raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_ARGUMENT_NOT_EXIST, "ethernet")
            cmd_shell = cmd_shell + " -e " + shlex.quote(ethernet)

        if Upgrade.upgrade_state == UpgradeState.UPGRADE_RUNNING_STATE:
            # 系统正在升级
            run_log.error("The system is upgrading, can not restore factory")
            raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_SYSTEM_UPGRADED_WRONG)

        # 挂载p1分区将操作日志信息写入，然后取消挂载
        self.record_restore_log(user_name, user_ip)

        # 恢复出厂设置
        try:
            with subprocess.Popen(cmd_shell.split(), stdin=subprocess.PIPE, stdout=subprocess.PIPE, shell=False) as s:
                s.stdin.write(b"Y\n")
                s.stdin.flush()
                msg_str = ",".join(str(msg.strip()) for msg in s.stdout.readlines())
                s.communicate(timeout=600)
        except Exception as ex:
            run_log.error("restore_defaults subprocess Popen error %s.", ex)
            raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_CMD_EXEC_WRONG)
        # 判断恢复出厂是否成功
        if "The system will be automatically reset" in msg_str:
            # 开始执行恢复出厂
            return
        elif "device is upgrading" in msg_str:
            # 系统正在升级
            raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_SYSTEM_UPGRADED_WRONG)
        else:
            # 其它原因恢复出厂失败
            raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_RESTORE_DEFAULTS_WRONG)