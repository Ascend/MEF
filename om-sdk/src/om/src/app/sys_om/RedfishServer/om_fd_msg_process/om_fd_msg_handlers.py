# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import json
import os
import shutil
import time

from common.checkers.fd_param_checker import ComputerSystemResetChecker
from common.checkers.fd_param_checker import SetDflcInfoChecker
from common.constants.base_constants import CommonConstants
from common.constants.base_constants import RecoverMiniOSConstants
from common.file_utils import FileUtils
from common.log.logger import run_log
from common.utils.app_common_method import AppCommonMethod
from fd_msg_process.common_redfish import CommonRedfish
from fd_msg_process.fd_common_methods import FdCommonMethod
from fd_msg_process.fd_common_methods import MidwareErrCode
from fd_msg_process.fd_common_methods import publish_ws_msg
from ibma_redfish_globals import RedfishGlobals
from lib_restful_adapter import LibRESTfulAdapter
from net_manager.constants import NetManagerConstants
from net_manager.manager.fd_cfg_manager import FdCfgManager
from net_manager.manager.fd_cfg_manager import FdMsgData
from net_manager.manager.net_cfg_manager import NetCertManager
from net_manager.manager.net_cfg_manager import NetCfgManager
from net_manager.models import NetManager
from net_manager.schemas import PayloadPublish


class RecoverError(Exception):
    pass


class OMFDMessageHandler:

    @staticmethod
    def check_external_param(checker_class, param_data, payload_publish, err_info):
        check_ret = checker_class().check(param_data)
        if not check_ret.success:
            run_log.error(err_info % check_ret.reason)
            if "restartable" in payload_publish.keys():
                payload_publish["restartable"] = "false"
            if "result" in payload_publish.keys():
                payload_publish["result"] = "failed"
            payload_publish["reason"] = MidwareErrCode.midware_generate_err_msg(
                MidwareErrCode.midware_input_parameter_invalid, check_ret.reason)
            return [-1, payload_publish]
        else:
            return [0, ""]

    @staticmethod
    @FdCommonMethod.fd_operational_log("Set dflc info")
    def config_netmanager_dflc(payload):
        input_opera_log = "Set dflc info failed."
        try:
            if not isinstance(payload, dict):
                return [-1, "FAILED: wrong param type"]

            if "start_point" not in payload or "life_span" not in payload:
                return [-1, "FAILED: need config start_point and life_span at same time"]

            check_ret = SetDflcInfoChecker().check(payload)
            if not check_ret.success:
                run_log.error(f"{input_opera_log} wrong param reason is {check_ret.reason}")
                return [-1, "FAILED: input param check failed"]

            dflc_dict = {"StartPoint": payload["start_point"], "LifeSpan": payload["life_span"]}
            ret_dict = LibRESTfulAdapter.lib_restful_interface("dflc_info", "POST", dflc_dict, False)
            ret = AppCommonMethod.check_status_is_ok(ret_dict)
            if not ret:
                run_log.error("set dflc info failed.")
                return [-1, "FAILED: set dflc info failed"]

            input_opera_log = "Set dflc info successfully."
            run_log.info(input_opera_log)
            return [0, "SUCCESS"]
        except Exception as err:
            run_log.error(f"{input_opera_log} reason is: {err}")
            return [-1, "FAILED: set dflc info failed"]

    @staticmethod
    def handle_msg_config_dflc(msg: FdMsgData):
        run_log.info("Receive topic: %s", msg.topic)
        ret = OMFDMessageHandler.config_netmanager_dflc(msg.content)
        ret_valid = isinstance(ret, list) and len(ret) == 2
        if not ret_valid:
            run_log.error("%s result invalid", msg.deal_func_name)
            publish_content = "FAILED: deal result invalid"
        else:
            publish_content = ret[1]

        parent_msg_id = msg.header.msg_id
        sync = msg.header.sync
        resp_msg_obj = FdMsgData.gen_ws_msg_obj(publish_content, msg.up_resource, parent_msg_id, sync)
        publish_ws_msg(resp_msg_obj)
        run_log.info("Done topic: %s", msg.topic)

    # 系统复位入口
    @staticmethod
    @FdCommonMethod.fd_operational_log("Restart system")
    def computer_system_reset(payload):
        payload_publish = {
            "restartable": "false",
            "reason": "resource is busy"
        }
        message = "Restart system failed."
        try:
            param_check_error = "Restart system check param failed. %s"
            check_ret = OMFDMessageHandler.check_external_param(
                ComputerSystemResetChecker, payload, payload_publish, param_check_error)
            if check_ret[0] != 0:
                return check_ret

            reset_type = payload.get("restart_method", None)

            if reset_type is None:
                run_log.error("Input parameter is invalid.")
                payload_publish["restartable"] = "false"
                payload_publish["reason"] = MidwareErrCode.midware_generate_err_msg(
                    MidwareErrCode.midware_input_parameter_invalid, "Input parameter invalid!")
                return [-1, payload_publish]

            recv_conf_dict = {}
            if reset_type == "Graceful":
                recv_conf_dict["ResetType"] = "GracefulRestart"
            elif reset_type == "Force":
                recv_conf_dict["ResetType"] = "ForceRestart"
            # 冷复位
            elif reset_type == "ColdReset":
                recv_conf_dict["ResetType"] = "Restart"
            else:
                run_log.error("Parameter reset_type error.")
                payload_publish["restartable"] = "false"
                payload_publish["reason"] = MidwareErrCode.midware_generate_err_msg(
                    MidwareErrCode.midware_input_parameter_invalid, "Parameter reset_type error.")
                return [-1, payload_publish]

            # 获取资源模板
            if CommonRedfish.SYS_CRITIC_LOCK.locked():
                if reset_type == "Graceful" or reset_type == "ColdReset":
                    payload_publish["reason"] = MidwareErrCode.midware_generate_err_msg(
                        MidwareErrCode.midware_resource_busy, "Resource is busy")
                    return [-1, payload_publish]
                elif reset_type == "Force":
                    payload_publish["restartable"] = "true"
                    resp_msg_obj = FdMsgData.gen_ws_msg_obj(payload_publish, "websocket/restart_result")
                    publish_ws_msg(resp_msg_obj)
                    time.sleep(5)
                    run_log.info("Force system to restart now.")

                    # 注意：此接口内部用定时器延迟重启，所以总能收到返回值
                    ret_dict = LibRESTfulAdapter.lib_restful_interface("Actions", "POST", recv_conf_dict, False)
                    run_log.info(f'{ret_dict}')
                    ret = AppCommonMethod.check_status_is_ok(ret_dict)
                    if ret:
                        message = f"Restart system ({reset_type}) successfully."
                        run_log.info(message)
                        return [0, payload_publish]

                    message = f"Restart system ({reset_type}) failed."
                    run_log.error(message)
                    payload_publish["reason"] = MidwareErrCode.midware_generate_err_msg(
                        MidwareErrCode.midware_common_err, ret_dict['message'])
                    return [-1, payload_publish]

            with CommonRedfish.SYS_CRITIC_LOCK:
                payload_publish["restartable"] = "true"
                payload_publish["reason"] = "System is restartable"
                resp_msg = FdMsgData.gen_ws_msg_obj(payload_publish, "websocket/restart_result")
                publish_ws_msg(resp_msg)
                time.sleep(5)
                run_log.info("System restart now.")

                # 注意：此接口内部用定时器延迟重启，所以总能收到返回值
                ret_dict = LibRESTfulAdapter.lib_restful_interface("Actions", "POST", recv_conf_dict, False)
                run_log.info(f'{ret_dict}')
                ret = AppCommonMethod.check_status_is_ok(ret_dict)
                if ret:
                    message = f"Restart system ({reset_type}) successfully."
                    run_log.info(message)
                    return [0, payload_publish]

                message = f"Restart system ({reset_type}) failed."
                run_log.error(message)
                payload_publish["restartable"] = "false"
                payload_publish["reason"] = MidwareErrCode.midware_generate_err_msg(
                    MidwareErrCode.midware_common_err, ret_dict['message'])
                return [-1, payload_publish]
        except Exception as err:
            run_log.error(f"{message}, error: {err}")
            payload_publish["reason"] = MidwareErrCode.midware_generate_err_msg(
                MidwareErrCode.midware_common_err, "System restart error.")
            return [-1, payload_publish]

    @staticmethod
    def handle_computer_system_reset_msg_from_fd_by_mqtt(msg: FdMsgData):
        run_log.info("Receive topic: %s", msg.topic)
        res = OMFDMessageHandler.computer_system_reset(msg.content)
        resp_msg_obj = FdMsgData.gen_ws_msg_obj(res[1], msg.up_resource)
        publish_ws_msg(resp_msg_obj)
        run_log.info("Done topic: %s", msg.topic)

    @staticmethod
    def fetch_net_manager_config():
        try:
            net_manager: NetManager = NetCfgManager().get_net_cfg_info()
        except Exception:
            raise RecoverError("Get netmanager info from db failed")

        if net_manager.net_mgmt_type != NetManagerConstants.FUSION_DIRECTOR:
            raise RecoverError("The net manager type is not FD mode.")
        return net_manager

    @staticmethod
    def fetch_cert_manager_config():
        try:
            cert_list = NetCertManager().get_all()
        except Exception:
            raise RecoverError("Get certmanager info from db failed")

        if not cert_list:
            raise RecoverError("The net manager cert list is empty.")

        return cert_list

    @staticmethod
    def save_fd_config_files(net_manager, cert_manager):
        net_manager_dict = {
            "config": net_manager.to_dict_for_update(),
            "cert": [cert.to_dict() for cert in cert_manager]
        }

        redfish_path = os.path.join(CommonConstants.CONFIG_HOME_PATH, "redfish")
        ksf_path = "redfish_encrypt.keystore"
        bak_ksf_path = "redfish_encrypt_backup.keystore"
        alg_path = "om_alg.json"

        # 拷贝 kmc 文件到 /home/package/config 目录下
        for file in (ksf_path, bak_ksf_path, alg_path):
            src_file = os.path.join(redfish_path, file)
            dest_file = os.path.join(RecoverMiniOSConstants.CONFIG_PATH, file)

            try:
                shutil.copyfile(src_file, dest_file)
            except Exception:
                raise RecoverError("Copy kmc file failed.")

        # 保存对接信息和证书列表，在 /home/package/config 目录下
        net_manage_json = os.path.join(RecoverMiniOSConstants.CONFIG_PATH, RecoverMiniOSConstants.NET_MANAGER_CONFIG)
        with os.fdopen(os.open(net_manage_json, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600), "w") as file:
            file.write(json.dumps(net_manager_dict))

    @staticmethod
    def handle_recover_mini_os(msg: FdMsgData):
        run_log.info("Receive topic: %s", msg.topic)
        ret_dict = OMFDMessageHandler().recover_mini_os()
        resp_msg_obj = FdMsgData.gen_ws_msg_obj(ret_dict[1], msg.up_resource)
        publish_ws_msg(resp_msg_obj)
        run_log.info("Done topic: %s", msg.topic)

    @FdCommonMethod.fd_operational_log("Prepare for recover mini os")
    def recover_mini_os(self):
        payload_publish = PayloadPublish(topic="min_recovery", reason="Recover mini os failed")
        payload_publish.percentage = "0%"
        payload_publish.result = "failed"

        if RedfishGlobals.high_risk_exclusive_lock.locked():
            run_log.error("The operation is busy.")
            return [-1, payload_publish.to_dict()]

        with RedfishGlobals.high_risk_exclusive_lock:
            ret_dict = LibRESTfulAdapter.lib_restful_interface("ExclusiveStatus", "GET", None, False)
            if not LibRESTfulAdapter.check_status_is_ok(ret_dict):
                return [-1, payload_publish.to_dict()]

            if not isinstance(ret_dict.get("message"), dict) or not \
                    isinstance(ret_dict.get("message").get("system_busy"), bool) or \
                    ret_dict.get("message").get("system_busy"):
                return [-1, payload_publish.to_dict()]

            try:
                # recover_mini_os_flag 存放在 /home/package/config 目录下
                with os.fdopen(os.open(RecoverMiniOSConstants.RECOVER_FLAG, os.O_CREAT, 0o400), "w"):
                    pass

                # 获取对接信息和证书列表，保存下来
                net_manager = self.fetch_net_manager_config()
                cert_list = self.fetch_cert_manager_config()
                self.save_fd_config_files(net_manager, cert_list)
                run_log.info("Save config for recover mini os successfully.")

                fd_ip = FdCfgManager.get_cur_fd_ip()
                if not fd_ip:
                    raise RecoverError("Get fd ip failed.")

                ret_dict = LibRESTfulAdapter.lib_restful_interface("RecoverMiniOS", "POST", {"fd_ip": fd_ip}, False)
                ret = AppCommonMethod.check_status_is_ok(ret_dict)
                if not ret:
                    raise RecoverError("Prepare for recover mini os failed.")

                run_log.info("Prepare for recover mini os successfully.")
                payload_publish.percentage = "100%"
                payload_publish.result = "success"
                payload_publish.reason = "Prepare for recover mini os successfully."
                return [0, payload_publish.to_dict()]
            except RecoverError as err:
                FileUtils.delete_dir_content(RecoverMiniOSConstants.CONFIG_PATH)
                run_log.error(err)
                payload_publish.percentage = "0%"
                payload_publish.result = "failed"
                payload_publish.reason = "Prepare for recover mini os has exception"
                return [-1, payload_publish.to_dict()]
            except Exception:
                FileUtils.delete_dir_content(RecoverMiniOSConstants.CONFIG_PATH)
                run_log.error("Prepare for recover mini os has internal exception")
                payload_publish.percentage = "0%"
                payload_publish.result = "failed"
                payload_publish.reason = "Prepare for recover mini os has internal exception"
                return [-1, payload_publish.to_dict()]
