# Copyright (c) Huawei Technologies Co., Ltd. 2023-2025. All rights reserved.
import copy
import time
from abc import ABC, abstractmethod
from functools import partial
from typing import NoReturn

from common.log.logger import run_log
from common.schema import AdapterResult
from common.utils.app_common_method import AppCommonMethod
from lib_restful_adapter import LibRESTfulAdapter
from mef_msg_process.mef_proc import MefProc
from net_manager.common_methods import docker_root_is_mounted
from net_manager.constants import NetManagerConstants
from net_manager.exception import NetSwitchException
from net_manager.manager.fd_cfg_manager import FdConfigData, FdCfgManager
from net_manager.manager.net_cfg_manager import NetCfgManager
from net_manager.models import NetManager
from net_manager.schemas import SystemInfo, WebNetInfo
from wsclient.connect_status import FdConnectStatus
from wsclient.fd_connect_check import FdConnectCheck
from wsclient.ws_client_mef import WsClientMef
from wsclient.ws_client_mgr import WsClientMgr
from wsclient.ws_monitor import WsMonitor


class NetSwitchManagerBase(ABC):
    """网管切换管理基类"""

    def __init__(self, config_param: dict):
        self.config_param = config_param
        self.net_cfg_manager = NetCfgManager()
        self.last_net_cfg = self.net_cfg_manager.get_net_cfg_info()

    @abstractmethod
    def switch_deal(self):
        """网管切换抽象方法"""
        pass

    def update_net_info(self, data: dict):
        self.net_cfg_manager.update_net_cfg_info(data)


class WebNetSwitchManager(NetSwitchManagerBase):
    """Web网管切换管理类"""

    adapter = partial(LibRESTfulAdapter.lib_restful_interface, "EtcHostsManager")

    def update_host_info(self, request_data: dict) -> NoReturn:
        try:
            result = AdapterResult.from_dict(self.adapter("POST", request_data, False))
        except Exception as err:
            raise NetSwitchException(f"Update host info failed, {err}") from err

        if result.status != AppCommonMethod.OK:
            raise NetSwitchException(f"Update host info failed, {result.message}")

    def switch_deal(self):
        # 切换成Web管理，当存在server_name时，需要清空/etc/hosts
        server_name = self.last_net_cfg.server_name
        if server_name:
            run_log.info("Switch Web manage execute clear hosts.")
            self.update_host_info({"operate_type": "clear", "fd_server_name": server_name})
        # 更新数据库
        self.update_net_info(WebNetInfo().to_dict())
        # 停止之前的所有协程任务
        run_log.info("Try to stop current loops when switching to Web manage.")
        WsMonitor.stop_all_connect_threads()
        WsClientMef.stop_mef_process()
        FdConnectStatus().trans_to_not_configured()
        # 切换成web时，消除掉所有fd证书的告警
        FdCfgManager.clean_fd_cert_alarm()
        run_log.info("Switch Web manage success.")
        return [0, ""]


class FdNetSwitchManager(NetSwitchManagerBase):
    """FusionDirector网管切换管理类"""

    adapter = partial(LibRESTfulAdapter.lib_restful_interface, "System")

    def connect_test(self) -> bool:
        net_manager: NetManager = NetManager.from_dict(copy.deepcopy(self.config_param))
        sys_info: SystemInfo = self.get_sys_info()
        result = FdConnectCheck.connect_test(FdConfigData.from_dict(net=net_manager, sys=sys_info))
        if not result:
            run_log.error("Connect test failed, %s", result.error)
            return False

        run_log.info("The connectivity test success!")
        return True

    def get_sys_info(self) -> SystemInfo:
        try:
            result = AdapterResult.from_dict(self.adapter("GET", None))
        except Exception as err:
            raise NetSwitchException(f"Get system info failed, {err}") from err

        if result.status != AppCommonMethod.OK:
            raise NetSwitchException(f"Get system info failed, {result.message}")

        if not result.message.get("SerialNumber"):
            run_log.info("Get serial number is null, will use node id [%s].", self.config_param.get("NodeId"))
            result.message["SerialNumber"] = self.config_param.get("NodeId")

        try:
            return SystemInfo.from_dict(result.message)
        except Exception as err:
            raise NetSwitchException(f"Get system info failed, {err}") from err

    def get_net_info(self, roll_back: bool = True) -> NetManager:
        if roll_back:
            # 回滚操作返回上一次配置实例
            return self.last_net_cfg

        # 更新操作返回当前配置实例
        config_net_info = copy.deepcopy(self.config_param)
        config_net_info["ServerName"] = config_net_info.get("ServerName") or NetManagerConstants.SERVER_NAME
        # 手动更新认证信息场景，纳管状态不变
        is_manual_update_pwd = all(
            (
                self.last_net_cfg.ip == config_net_info.get("NetIP"),
                self.last_net_cfg.cloud_user == config_net_info.get("NetAccount"),
                self.last_net_cfg.cloud_pwd != config_net_info.get("NetPassword")
            )
        )
        if is_manual_update_pwd:
            config_net_info["Status"] = self.last_net_cfg.status

        return NetManager.from_dict(config_net_info)

    def switch_deal(self):
        WsClientMgr().clear_connect_result()
        if not docker_root_is_mounted():
            run_log.error("Switch FusionDirector manage failed: docker root dir is not mounted.")
            return [110226, "docker root dir is not mounted."]

        if self.config_param.get("test"):
            # 进行连接性测试
            run_log.info("Start the connectivity test.")
            try:
                result = self.connect_test()
            except Exception as err:
                raise NetSwitchException(f"Switch FusionDirector manage failed, {err}") from err

            if not result:
                if WsClientMgr().check_account_info_invalid():
                    run_log.error("Switch FusionDirector manage failed: invalid account info.")
                    WsClientMgr().clear_connect_result()
                    return [110207, "The user name or password error."]

                if WsClientMgr().check_ip_locked():
                    run_log.error("Switch FusionDirector manage failed: ip locked.")
                    WsClientMgr().clear_connect_result()
                    return [110225, "IP locked."]
                WsClientMgr().clear_connect_result()
                raise NetSwitchException("Switch FusionDirector manage failed.")

        # 更新数据库
        self.update_net_info(self.get_net_info(False).to_dict_for_update())
        run_log.info("Please wait for the delivery of one device and one secret.")
        # 停止之前的所有协程任务
        run_log.info("Try to stop current loops when switching to FD %s.", self.config_param.get("NetIP"))
        WsMonitor.stop_all_connect_threads()

        FdConnectStatus().trans_to_connecting()
        WsMonitor.start_fd_connect_monitor()
        MefProc.start_mef_connect_timer()
        # 等待一机一密更新，最大等待2min后超时,一般正常1分钟左右可以就绪
        for _ in range(1, 6 * 4):
            time.sleep(5)
            # 每5秒检查一次状态是否就绪
            status: str = self.net_cfg_manager.get_net_cfg_info().status
            if status == "ready" and WsClientMgr().ready_for_send_msg():
                FdConnectStatus().trans_to_connected()
                run_log.info("Switch FusionDirector manage success.")
                return [0, ""]

            if WsClientMgr().check_account_info_invalid():
                run_log.error("Switch FusionDirector manage failed: invalid account info.")
                FdConnectStatus().trans_to_err_configured()
                WsMonitor.stop_all_connect_threads()
                return [110207, "The user name or password error."]

            if WsClientMgr().check_ip_locked():
                run_log.error("Switch FusionDirector manage failed: ip locked.")
                FdConnectStatus().trans_to_err_configured()
                WsMonitor.stop_all_connect_threads()
                return [110225, "IP locked."]

            if WsClientMgr().check_cert_is_invalid() or WsClientMgr().check_ip_port_is_invalid():
                break

        run_log.info("Save FusionDirector manage config params successfully.")
        return [206, "save config success"]
