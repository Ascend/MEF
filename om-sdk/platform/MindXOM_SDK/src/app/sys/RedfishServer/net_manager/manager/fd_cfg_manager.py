# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import json
import os
import ssl
import time
import uuid
from dataclasses import dataclass, field
from functools import partial
from typing import Optional

from common.constants.base_constants import RecoverMiniOSConstants
from common.file_utils import FileUtils, FileCopy, FileCheck
from common.log.logger import run_log
from common.schema import AdapterResult
from common.schema import BaseModel
from common.utils.app_common_method import AppCommonMethod
from common.utils.result_base import Result
from common.utils.timer import RepeatingTimer
from fd_msg_process.config import Config
from lib_restful_adapter import LibRESTfulAdapter
from net_manager.checkers.fd_param_checker import FdMsgChecker
from net_manager.constants import NetManagerConstants
from net_manager.manager.fd_cert_manager import FdCertManager
from net_manager.manager.import_manager import ImportCert, ImportCrl
from net_manager.manager.net_cfg_manager import NetCfgManager, CertMgr
from net_manager.models import NetManager
from net_manager.schemas import HeaderData
from net_manager.schemas import RouteData
from net_manager.schemas import SystemInfo


class FdCfgManager:
    """FusionDirector配置管理类"""
    EXPIRATION_THRESHOLD_DAY: int = 10
    MONITOR_PERIOD = 24 * 60 * 60  # one day 24 * 60 * 60

    adapter = partial(LibRESTfulAdapter.lib_restful_interface, "System")

    def __init__(self):
        self.net_manager = NetCfgManager().get_net_cfg_info()
        self.cert_mgr = CertMgr()

    @staticmethod
    def get_cur_fd_ip():
        """获取当前FusionDirector网管IP地址，若发生异常，则返回空字符串"""
        try:
            net_cfg = NetCfgManager().get_net_cfg_info()
        except Exception as err:
            run_log.error("get net manager info from db failed: %s", err)
            return ""

        return net_cfg.ip

    @staticmethod
    def get_cur_fd_server_name():
        """从数据库获取当前FusionDirector网管域名，若发生异常，则使用默认FD域名"""
        try:
            net_cfg = NetCfgManager().get_net_cfg_info()
        except Exception as err:
            run_log.warning("Get net manger info from db failed, will use default server name. Reason: %s", err)
            return NetManagerConstants.SERVER_NAME

        return net_cfg.server_name or NetManagerConstants.SERVER_NAME

    @staticmethod
    def check_fd_mode_and_status_ready() -> bool:
        """检查当前是否为FD纳管模式，且status为ready状态"""
        try:
            net_cfg = NetCfgManager().get_net_cfg_info()
        except Exception as err:
            run_log.warning("Get net manger info from db failed, will use default mode web. Reason: %s", err)
            return False

        return net_cfg.net_mgmt_type == NetManagerConstants.FUSION_DIRECTOR and net_cfg.status == "ready"

    @staticmethod
    def check_manager_type_is_web() -> bool:
        """
        检查当前网管模式是否为WEB模式
        :return: bool
        """
        try:
            net_cfg = NetCfgManager().get_net_cfg_info()
        except Exception as err:
            run_log.warning("Get net manger info from db failed, will use default mode web. Reason: %s", err)
            return True

        return net_cfg.net_mgmt_type == NetManagerConstants.WEB

    @staticmethod
    def get_fd_info_to_mef() -> dict:
        """获取MEF需要的FD配置信息"""
        try:
            net_cfg = NetCfgManager().get_net_cfg_info()
            return {
                "ip": net_cfg.ip,
                "domain": net_cfg.server_name,
                "port": net_cfg.port,
                "ca_content": FdCertManager(net_cfg.ip, net_cfg.port).cert_to_mef()
            }
        except Exception as err:
            run_log.warning("Get net manger info from db failed, will use default server name. Reason: %s", err)
            return {}

    @staticmethod
    def update_etc_hosts(fd_ip, old_server_name, new_server_name):
        """
        更新/etc/hosts文件中的fd网管域名配置信息，域名通过接口保存至db,这里直接从db查询
        :param fd_ip: 网管IP
        :param old_server_name: 原有的FD域名
        :param new_server_name: 新的的FD域名
        :return: Result实例
        """
        # 先清除原有配置，再更新配置
        clear_request = {
            "operate_type": "clear",
            "fd_server_name": old_server_name  # 实例中保存的是旧的fd域名
        }
        ret = LibRESTfulAdapter.lib_restful_interface("EtcHostsManager", "POST", clear_request, False)
        if not AppCommonMethod.check_status_is_ok(ret):
            return Result(False, err_msg="clear old /etc/hosts failed")

        update_request = {
            "operate_type": "set",
            "fd_ip": fd_ip,
            "fd_server_name": new_server_name,
        }
        ret = LibRESTfulAdapter.lib_restful_interface("EtcHostsManager", "POST", update_request, False)
        if not AppCommonMethod.check_status_is_ok(ret):
            return Result(False, err_msg="update new /etc/hosts failed")
        return Result(True)

    @staticmethod
    def restore_mini_os_config():
        """
        恢复最小系统启动后，需恢复FD配置
        """
        try:
            if not os.path.exists(RecoverMiniOSConstants.RECOVER_FLAG):
                run_log.info("Recover mini os flag not exist.")
                return

            run_log.info("Restore fd config after recover mini os and start redfish.")

            ksf_path = "redfish_encrypt.keystore"
            bak_ksf_path = "redfish_encrypt_backup.keystore"
            alg_path = "om_alg.json"

            net_manage_file = os.path.join(RecoverMiniOSConstants.CONFIG_PATH,
                                           RecoverMiniOSConstants.NET_MANAGER_CONFIG)

            redfish_path = "/home/data/config/redfish"
            for file in (ksf_path, bak_ksf_path, alg_path):
                FileCopy.copy_file(os.path.join(RecoverMiniOSConstants.CONFIG_PATH, file),
                                   os.path.join(redfish_path, file))

            with open(net_manage_file, "r") as file:
                net_manage_json = json.loads(file.read())
            for cert in net_manage_json["cert"]:
                ImportCert(cert["source"], cert["name"]).import_deal(cert["cert_contents"])
                if not cert["crl_contents"]:
                    continue
                try:
                    ImportCrl().import_deal(cert["crl_contents"])
                except Exception as err:
                    run_log.error("Restore crl failed, %s.", err)
            NetCfgManager().update_net_cfg_info(net_manage_json["config"])
            run_log.info("Restore fd config after recover mini os successfully.")
        except Exception:
            run_log.error("Restore fd config failed.")
        finally:
            FileUtils.delete_dir_content(RecoverMiniOSConstants.CONFIG_PATH)

    @staticmethod
    def check_and_reset_status():
        """进程启动之后，协程客户端启动之前调用，用于重置Database中的无效状态"""
        try:
            net_manager: NetManager = NetCfgManager().get_net_cfg_info()
        except Exception as err:
            run_log.warning("Get netmanager info from db failed, caught exception: %s.", err)
            return

        # 异常场景：还在连接状态，系统重启了，需要重置为""
        if net_manager.status == "connecting":
            NetCfgManager().update_net_cfg_info({"status": ""})
            run_log.info("Reset netmanager status successfully.")

    @staticmethod
    def modify_alarm(fd_cert_name_list: list):
        request_data_dict = {"FdCertNameList": fd_cert_name_list}
        ret = LibRESTfulAdapter.lib_restful_interface("ModifyAlarm", "POST", request_data_dict)
        if LibRESTfulAdapter.check_status_is_ok(ret):
            run_log.info("Modify alarm succeed")

    @staticmethod
    def clean_fd_cert_alarm():
        request_data_dict = {}
        ret = LibRESTfulAdapter.lib_restful_interface("ModifyAlarm", "POST", request_data_dict)
        if LibRESTfulAdapter.check_status_is_ok(ret):
            run_log.info("Clean FD certificate alarm succeed")

    def check_status_is_ready(self):
        return self.net_manager.status == "ready"

    def check_is_switch_fd(self, target_ip):
        return self.net_manager.ip != target_ip

    def get_sys_info(self) -> Result:
        try:
            result = AdapterResult.from_dict(self.adapter("GET", None))
        except Exception as err:
            return Result(result=False, err_msg=f"Get system info failed, {err}")

        if result.status != AppCommonMethod.OK:
            return Result(result=False, err_msg=f"Get system info failed, {result.message}")

        if not result.message.get("SerialNumber"):
            result.message["SerialNumber"] = self.net_manager.node_id

        try:
            return Result(result=True, data=SystemInfo.from_dict(result.message))
        except Exception as err:
            return Result(result=False, err_msg=f"Get system info failed, {err}")

    def get_ws_config(self) -> Result:
        """获取WebSocket连接相关的配置信息"""
        if self.net_manager.net_mgmt_type != NetManagerConstants.FUSION_DIRECTOR:
            return Result(result=True, data=FdConfigData(net_mgmt_type=self.net_manager.net_mgmt_type))

        result = self.get_sys_info()
        if not result:
            return result

        return Result(True, FdConfigData.from_dict(net=self.net_manager, sys=result.data))

    def check_cert_status(self):
        try:
            self._check_cert_status()
        except Exception as err:
            run_log.error("Failed to verify the validity period of certificates: %s", err)

    def start_cert_status_monitor(self):
        RepeatingTimer(self.MONITOR_PERIOD, self.check_cert_status).start()

    def _check_cert_status(self):
        """
        检查证书有效期
        """
        fd_cert_name_list = []
        expiration_time = self._get_cert_expiration_time()
        run_log.info("Start to check certificates for expire alarm, current expiration time [%d] days", expiration_time)

        # 判断是FD纳管就绪状态
        if self.check_status_is_ready():
            for cert_info in self.cert_mgr.get_expire_certs(expiration_time):
                fd_cert_name_list.append(cert_info.name)
                run_log.info("Add %s FD certificate about to expire alarm", cert_info.name)

            if self.cert_mgr.is_crl_expired(expiration_time):
                fd_cert_name_list.append(self.cert_mgr.FD_CRL_NAME)
                run_log.info("Add %s FD CRL about to expire alarm", self.cert_mgr.FD_CRL_NAME)

        self.modify_alarm(fd_cert_name_list)
        run_log.info("Check the validity period of the FD certificate succeed")

    def _get_cert_expiration_time(self) -> int:
        ret_dict = LibRESTfulAdapter.lib_restful_interface("CertAlarmTime", "GET", None, False)
        if LibRESTfulAdapter.check_status_is_ok(ret_dict) and isinstance(ret_dict.get("message"), dict):
            return ret_dict.get("message").get("CertAlarmTime")
        return self.EXPIRATION_THRESHOLD_DAY


@dataclass
class FdConfigData:
    """FD连接参数的数据类"""
    cloud_pwd: Optional[str] = field(default=None)
    product_name: str = field(default="")
    asset_tag: str = field(default="")
    serial_number: str = field(default="")
    node_id: str = field(default="")
    net_mgmt_type: str = field(default="")
    # 独立对接上报FD时需要固定AtlasEdge
    dev_mgmt_type: str = field(default="AtlasEdge")
    server_name: str = field(default=NetManagerConstants.SERVER_NAME)
    server_ip: str = field(default="")
    server_port: int = field(default=443)
    cloud_user: str = field(default="")
    status: str = field(default="")
    wss_uri: str = field(init=False)
    test_url: str = field(init=False)

    def __post_init__(self):
        self.wss_uri = f"wss://{self.server_ip}:{self.server_port}/websocket/{self.node_id}/events"
        self.test_url = f"https://{self.server_ip}:{self.server_port}/websocket/{self.node_id}/AccountCheck"

    @classmethod
    def from_dict(cls, **kwargs):
        """"将数据转成FdConfigData对象"""
        net_manager: NetManager = kwargs.get("net")
        system_info: SystemInfo = kwargs.get("sys")
        return cls(
            product_name=system_info.product_name or 'Atlas 500',
            asset_tag=system_info.asset_tag or "",
            serial_number=system_info.serial_number,
            node_id=net_manager.node_id,
            net_mgmt_type=net_manager.net_mgmt_type,
            server_name=net_manager.server_name,
            server_ip=net_manager.ip,
            server_port=int(net_manager.port) if net_manager.port else 443,
            cloud_user=net_manager.cloud_user,
            cloud_pwd=net_manager.decrypt_cloud_pwd(),
            status=net_manager.status,
        )

    def gen_extra_headers(self) -> dict:
        """
        生成headers
        :return: dict
        """
        authorization = AppCommonMethod.make_authentication_string(self.cloud_user, self.cloud_pwd)
        headers = {
            "Authorization": authorization,
            "ProductName": self.product_name,
            "SerialNumber": self.serial_number,
            "AssetTag": self.asset_tag,
            "DevMgmtType": self.dev_mgmt_type
        }
        return headers

    def gen_client_ssl_context(self, set_state=False) -> Optional[ssl.SSLContext]:
        """生成连接服务端的安全SSLContext"""
        return FdCertManager(self.server_ip, self.server_port).get_client_ssl_context(set_state)


class FdMsgException(Exception):
    """FD消息解析异常"""
    pass


@dataclass
class FdMsgData(BaseModel):
    """FD下发消息"""
    header: HeaderData
    route: RouteData
    content: dict
    # 向FD返回消息的resource信息
    up_resource: str = ""
    # FD消息的处理函数名，对应midware_urls.py中函数的route
    deal_func_name: str = ""
    topic: str = field(init=False)
    payload: str = field(init=False)

    def __post_init__(self):
        self.topic = f"$hw/edge/v1/hardware/operate/{self.route.resource[len('websocket/'):]}"
        self.payload = json.dumps(self.content)

    @classmethod
    def from_str(cls, msg):
        """解析下发消息内容"""
        if len(msg) >= Config.mqtt_max_msg_payload_size:
            raise FdMsgException("msg size is too large")

        try:
            data = json.loads(msg)
        except Exception as err:
            raise FdMsgException("json loads failed") from err

        ret = FdMsgChecker().check(data)
        if not ret:
            raise FdMsgException(f"invalid Fd msg format: {ret.reason}")

        return FdMsgData.from_dict(data)

    @classmethod
    def gen_ws_msg_obj(cls, content, resource, parent_msg_id="", sync=False, group="hub", operation="update",
                       source="hardware"):
        """
        根据参数生成websocket消息实例
        :param content: 具体的消息内容，由各个组件自行定义。
        :param resource: 资源信息，字符串格式，具体请参见各个接口的具体定义。
        :param parent_msg_id: 该消息的父消息，格式为UUID，可以此表示request/response关系。
                            对于请求消息request，该字段为空。对于应答消息response，该字段为请求消息的msg_id。
        :param sync: 表示同步还是异步消息，取值为：
                false：表示异步消息，不回应答。默认为false。true：表示同步消息，要求边缘回应答。
        :param group: 消息分类，参考消息分组定义。
        :param operation: 消息操作类型，字符串格式，取值为insert、update、delete、query等。
        :param source: 消息源，即发送消息的模块，从ESPManager上报的消息，统一填写hardware。
        :return: FdMsgData实例
        """
        header = HeaderData(msg_id=str(uuid.uuid4()), parent_msg_id=parent_msg_id, sync=sync,
                            timestamp=round(time.time() * 1000))
        route = RouteData(group=group, operation=operation, resource=resource, source=source)
        return cls(header, route, content=content)

    def to_ws_msg_str(self):
        """websocket消息字符串格式"""
        data = {
            "header": self.header.to_dict(),
            "route": self.route.to_dict(),
            "content": self.content if isinstance(self.content, str) else json.dumps(self.content)
        }
        return json.dumps(data)
