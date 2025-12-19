# coding: utf-8
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
import ssl
from socket import create_connection
from typing import Tuple

import urllib3

from common.utils.result_base import Result

from common.log.logger import run_log
from net_manager.manager.fd_cfg_manager import FdCfgManager
from net_manager.manager.fd_cfg_manager import FdConfigData
from wsclient.ws_client_mgr import WsClientMgr


class FdConnectCheck:
    """FD连接测试检查"""
    LIMIT_1M_SIZE = 1 * 1024 * 1024
    SOCKET_CONNECT_TIMEOUT: int = 5

    @staticmethod
    def connect_test(config: FdConfigData) -> Result:
        """
        websocket客户端尝试连接FD服务端认证用户名密码是否正确
        :param config: FdConfigData
        :return: Result
        """
        try:
            return FdConnectCheck._connect_test(config)
        except Exception as err:
            if isinstance(err, ssl.SSLCertVerificationError):
                WsClientMgr().set_connect_result(WsClientMgr.ERR_INVALID_FD_CERT)
            run_log.error("connect test failed")
            return Result(False, err_msg="connect test failed")

    @staticmethod
    def get_fd_config_data() -> Result:
        """
        获取FD的配置参数
        :return: FdConfigData
        """
        try:
            ret = FdCfgManager().get_ws_config()
        except Exception as err:
            return Result(result=False, err_msg=f"get fd config data failed, caught exception: {err}")

        return ret

    @staticmethod
    def check_fd_connect_status_ready():
        try:
            return FdCfgManager().check_status_is_ready()
        except Exception as err:
            run_log.error("caught exception: %s", err)
            return False

    @staticmethod
    def is_switch_fd(target_fd_ip):
        try:
            return FdCfgManager().check_is_switch_fd(target_fd_ip)
        except Exception as err:
            run_log.error("caught exception: %s", err)
            return False

    @staticmethod
    def is_ip_port_available(ip_port: Tuple[str, int]):
        """判断ip和端口是否可达"""
        try:
            with create_connection(address=ip_port, timeout=FdConnectCheck.SOCKET_CONNECT_TIMEOUT):
                return True
        except ConnectionRefusedError:
            # ip不可达情况
            WsClientMgr().set_connect_result(WsClientMgr.ERR_INVALID_FD_IP_PORT)
            return False

    @staticmethod
    def _connect_test(config: FdConfigData) -> Result:
        if not FdConnectCheck.is_ip_port_available((config.server_ip, config.server_port)):
            return Result(False, err_msg="connect test failed")

        headers = config.gen_extra_headers()
        if not headers:
            return Result(False, err_msg="connect test failed")

        ssl_context_ret = config.gen_client_ssl_context()
        if not ssl_context_ret:
            WsClientMgr().set_connect_result(WsClientMgr.ERR_INVALID_FD_CERT)
            return Result(result=False, err_msg="connect test failed")

        ssl_context = ssl_context_ret.data
        https = urllib3.PoolManager(ssl_context=ssl_context, assert_hostname=False, timeout=5, retries=1)
        resp = https.request("GET", config.test_url, headers=headers, preload_content=False)
        try:
            res = FdConnectCheck._check_resp_valid(resp)
        except Exception as err:
            if isinstance(err, ssl.SSLCertVerificationError):
                WsClientMgr().set_connect_result(WsClientMgr.ERR_INVALID_FD_CERT)
            run_log.error("Check response valid failed, caught exception: %s.", err)
            return Result(False)
        finally:
            if isinstance(resp, urllib3.response.HTTPResponse):
                resp.release_conn()
        if not res:
            return Result(False, err_msg=f"invalid response from fd: {res.error}")

        return Result(True)

    @staticmethod
    def _check_resp_valid(response) -> Result:
        try:
            content_length = int(response.headers.get("Content-Length"))
        except Exception as abn:
            run_log.error("get content length info failed, find exception: %s", abn)
            return Result(result=False, err_msg="get content length info failed, find exception")

        if content_length > FdConnectCheck.LIMIT_1M_SIZE:
            run_log.error("Request failed: payload too large")
            return Result(result=False, err_msg="Request failed: payload too large")

        run_log.info(f"receive fd response code: {response.status}")
        if response.status == 200:
            return Result(result=True)

        if response.status == 404:
            run_log.error("fd request failed: can not request the fd server")
            return Result(result=False, err_msg="Https request failed: can not request the fd server")

        try:
            response_data = response.read(amt=FdConnectCheck.LIMIT_1M_SIZE)
        except Exception as abn:
            run_log.error("fd connect test failed, find exception: %s", abn)
            return Result(result=False, err_msg="fd connect test failed, find exception")

        return FdConnectCheck._check_status_code(response_data)

    @staticmethod
    def _check_status_code(data) -> Result:
        error_mes = ""
        try:
            res = json.loads(data.decode("utf-8"))
        except Exception as abn:
            run_log.error("json loads response data failed, find exception: %s", abn)
            return Result(result=False, err_msg="Json loads response data failed")

        if "error" in res.keys():
            error_mes = res.get("error").get("@Message.ExtendedInfo")[0].get("MessageId")

        if error_mes in WsClientMgr.ERR_INVALID_ACCOUNT:
            WsClientMgr().set_connect_result(error_mes)
            run_log.error("Test user failed: %s", error_mes)
            return Result(result=False, err_msg="Test config failed. account or pword is wrong.")

        if error_mes == WsClientMgr.ERR_IP_LOCKED:
            WsClientMgr().set_connect_result(error_mes)
            run_log.error("Test user failed: %s", error_mes)
            return Result(result=False, err_msg="Test config failed. IP locked.")

        if error_mes == "EdgeDevMgmt.1.0.InternalError":
            run_log.error("FD InternalError: EdgeDevMgmt.1.0.InternalError")
            return Result(result=False, err_msg="FD InternalError")

        if error_mes == "EdgeDevMgmt.1.0.NodeIDExist":
            run_log.error("NodeID exist: EdgeDevMgmt.1.0.NodeIDExist")
            return Result(result=False, err_msg="Test nodeid failed, nodeid exist in FD: EdgeDevMgmt.1.0.NodeIDExist.")

        if error_mes == "EdgeDevMgmt.1.0.TheSameDevice":
            WsClientMgr().connected = True
            run_log.info("The same device: EdgeDevMgmt.1.0.TheSameDevice")
            return Result(result=True, err_msg="The same device: EdgeDevMgmt.1.0.TheSameDevice")

        run_log.error("can not check fd return message")
        return Result(result=False, err_msg="The check internal error")
