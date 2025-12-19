#!/usr/bin/python
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import ast
import os
import threading
import time

import common.common_methods as CommMethods
from common.constants.base_constants import CommonConstants
from common.file_utils import FileCheck
from common.file_utils import FilePermission
from common.log.logger import run_log
from common.utils.exec_cmd import ExecCmd
from common.checkers import IPV4Checker
from lib.Linux.systems.nic.config_web_ip import NginxConfig


class NTPService:
    """
    功能描述：NTP配置信息
    接口：NA
    """
    NTP_LOCK = threading.Lock()
    TIMER_RUN_CNT = 0
    file_path = "/home/data/ies/NTPEnable.ini"

    def __init__(self):
        """
        功能描述：初始化函数
        参数：
        返回值：无
        异常描述：NA
        """
        self.Target = None
        self.NTPRemoteServers = None
        self.NTPRemoteServersbak = None
        self.NTPLocalServers = None
        self.ntp_client_enable = None
        self.NTPServerEnable = None
        self.ClientEnabled = None
        self.ServerEnabled = None

    @staticmethod
    def get_ntp_config():
        result = []
        ntp_config = "/etc/ntp.conf"
        res = FileCheck.check_path_is_exist_and_valid(ntp_config)
        if not res:
            run_log.error("%s path invalid : %s", ntp_config, res.error)
            return result

        # 校验文件大小
        filesize = os.path.getsize(ntp_config)
        if filesize > CommonConstants.OM_READ_FILE_MAX_SIZE_BYTES:
            run_log.error("File: %s is too large.", ntp_config)
            return result

        with open(ntp_config, "r") as config_file:
            for line in config_file:
                if line.startswith("server") and len(line.split(" ")) > 1:
                    result.append(line.split(" ")[1].strip())

        return result

    @staticmethod
    def stop_ntp_server():
        ntpscript = "/usr/local/mindx/MindXOM/scripts/ntp_service.sh"
        cmd = [ntpscript, "stop"]
        sign = ExecCmd.exec_cmd(cmd, wait=30)
        if sign != 0:
            run_log.error("Stop ntp service failed.")
            return [1, "Failed to stop ntp service"]

        return [0, "stop shell ok"]

    @staticmethod
    def ntpserver_format_check(ntpserver_str):
        if ntpserver_str is None:
            return [1, "Ntpserver is none"]
        elif IPV4Checker("ipv4_data").check({"ipv4_data": ntpserver_str}):
            return [0, " NTP format is ok"]
        else:
            return [1, "ntpserver is illegal"]

    @staticmethod
    def get_local_server_ip():
        try:
            _, nginx_listen_ipv4 = NginxConfig.get_nginx_listen_ipv4()
        except Exception as err:
            run_log.warning("get nginx listen ipv4 failed. %s", err)
            return ""

        return nginx_listen_ipv4

    def get_all_info(self):
        self.ClientEnabled = self.get_ntp_info()
        self.ServerEnabled = False
        self.NTPLocalServers = ""
        self.get_ntp_servers()

    def get_ntp_servers(self):
        ret_list = NTPService.get_ntp_config()

        # 排除元素127.127.1.0
        ip_address = "127.127.1.0"
        if ip_address in ret_list:
            ret_list.remove(ip_address)

        if len(ret_list) == 1:
            self.NTPRemoteServers = ret_list[0]
            self.NTPRemoteServersbak = None
        elif len(ret_list) == 2:
            self.NTPRemoteServers = ret_list[0]
            self.NTPRemoteServersbak = ret_list[1]

    def get_ntp_info(self):
        ret = FileCheck.check_path_is_exist_and_valid(self.file_path)
        if not ret:
            return False

        try:
            FileCheck.check_is_link_exception(self.file_path)
            with open(self.file_path, "r") as file:
                nt_pfile = file.read()
            ntp_enable_dict = ast.literal_eval(nt_pfile)
        except Exception as err:
            run_log.info(f"{err}")
            return False

        return ntp_enable_dict.get("NTPClientEnable", False)

    def patch_request(self, request_dict):
        """NTP配置"""
        if NTPService.NTP_LOCK.locked():
            run_log.warning("NTP modify is busy")
            return [CommMethods.CommonMethods.ERROR, "NTP modify is busy"]

        with NTPService.NTP_LOCK:
            self.get_all_info()
            ret = self.check_request_dict(request_dict)
            if ret[0] != CommMethods.CommonMethods.OK:
                return ret

            # 开启了client，默认监听web service ip
            self.NTPLocalServers = self.get_local_server_ip()

            # 客户端标志打开 或者 服务端标志打开
            if self.ntp_client_enable:
                ret = self.remote_server_check(self.NTPRemoteServers, self.NTPRemoteServersbak)
                if ret[0] != CommMethods.CommonMethods.OK:
                    return ret
                ntp_remote_servers = ret[1]

                # 写标志文件
                self.write_ntp_enable(self.ntp_client_enable, self.NTPServerEnable, self.NTPLocalServers)

                ret_i = self.update_ntp_server(ntp_remote_servers, self.NTPLocalServers)
                if ret_i[0] == 0:
                    time.sleep(0.1)
                    self.get_all_info()
                    run_log.info("Start NTP service successfully!")
                    return [CommMethods.CommonMethods.OK, "NTP set success"]
                else:
                    self.write_ntp_enable(self.ClientEnabled, self.ServerEnabled, self.NTPLocalServers)
                    return [CommMethods.CommonMethods.ERROR, str(ret_i[1])]

            else:
                ret = self.stop_ntp_server()
                if ret[0] == 0:
                    self.write_ntp_enable(False, False, "")
                    time.sleep(0.1)
                    self.get_all_info()
                    run_log.info("Stop ntp service successfully!")
                    return [CommMethods.CommonMethods.OK, "NTP set success"]
                else:
                    return [CommMethods.CommonMethods.ERROR, str(ret[1])]

    def check_request_dict(self, request_dict):
        self.NTPServerEnable = request_dict.get("ServerEnabled", False)  # 小站的NTP服务端是否开启标志
        self.ntp_client_enable = request_dict.get("ClientEnabled", None)  # 小站是否从别的服务器同步时间的开启标志
        self.Target = request_dict.get("Target", None)

        if not isinstance(self.NTPServerEnable, bool):
            return [CommMethods.CommonMethods.ERROR, "NTPServerEnable wrong format"]

        if not isinstance(self.ntp_client_enable, bool):
            return [CommMethods.CommonMethods.ERROR, "ntp_client_enable wrong format"]

        if not self.Target:
            self.Target = "Client"

        if self.Target != "Client":
            return [CommMethods.CommonMethods.ERROR, "Ntp target is not Client"]

        if self.NTPServerEnable:
            return [CommMethods.CommonMethods.ERROR, "no support NTPServerEnable config"]

        # 作为client时，远端服务器IP
        self.NTPRemoteServers = request_dict.get("NTPRemoteServers", None)
        # 作为client时，远端服务器备选IP
        self.NTPRemoteServersbak = request_dict.get("NTPRemoteServersbak", None)
        # 开启客户端，至少配置RemoteServers
        if self.ntp_client_enable and self.NTPRemoteServers is None:
            return [CommMethods.CommonMethods.ERROR,
                    "The NTP client is enabled. At least NTPRemoteServer must be configured"]

        for req_ip in request_dict.get("NTPLocalServers", ""), self.NTPRemoteServers, self.NTPRemoteServersbak:
            if req_ip == "0.0.0.0":
                run_log.error("Check request server ip is invalid.")
                return [CommMethods.CommonMethods.ERROR, "Servers ip wrong format"]

        return [CommMethods.CommonMethods.OK, "check request dict success."]

    def remote_server_check(self, server, server_back):
        ret = self.ntpserver_format_check(server)
        if ret[0] == 1:
            run_log.error("NTPRemoteServers: ntp remote server is illegal")
            return [CommMethods.CommonMethods.ERROR, "NTPRemoteServers :ntp remote server is illegal"]
        if server_back:
            ret = self.ntpserver_format_check(server_back)
            if ret[0] == 1:
                run_log.error("NTPRemoteServersbak : ntp remote back server is illegal")
                return [CommMethods.CommonMethods.ERROR, "NTPRemoteServersbak: ntp remote back server is illegal"]
            if server == server_back:
                run_log.error("The primary remote NTP address is the same as the secondary address.")
                return [CommMethods.CommonMethods.ERROR,
                        "The primary remote NTP address is the same as the secondary address."]
            else:
                return [CommMethods.CommonMethods.OK, "%s;%s" % (server, server_back)]
        else:
            return [CommMethods.CommonMethods.OK, server]

    # 保存标志文件
    def write_ntp_enable(self, client_state, server_state, localservers):
        result = self.ntpserver_format_check(localservers)
        if localservers != "" and result[0] == 1:
            run_log.error("NTPLocalServers : [%s] is illegal" % localservers)
            return

        run_log.info("Initialize the NTP enable start.")
        enable_content = {
            "NTPClientEnable": client_state,
            "NTPServerEnable": server_state,
            "NTPLocalServers": localservers
        }
        try:
            FileCheck.check_is_link_exception(self.file_path)
            with os.fdopen(os.open(self.file_path, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o640), "w") as enable:
                enable.write(str(enable_content))
        except IOError as err:
            run_log.error(f"{err}")
        except Exception as err:
            run_log.error(f"{err}")

        # 父目录权限为700, os.fdopen指定权限也无法把文件所属组权限设置成功, 此处单独再设置一次权限
        FilePermission.set_path_permission(self.file_path, 0o640)
        run_log.info("Initialize the NTP enable done.")

    def update_ntp_server(self, remote_servers, local_servers):
        result = self.ntpserver_format_check(local_servers)
        if local_servers != "" and result[0] == 1:
            return [1, "NTPLocalServers : [%s] is illegal" % local_servers]
        self.stop_ntp_server()
        time.sleep(0.1)

        ntpscript = "/usr/local/mindx/MindXOM/scripts/ntp_service.sh"
        cmd = [ntpscript, "start", remote_servers, local_servers, "-f"]
        run_log.info("start cmd: %s" % (cmd))

        sign = ExecCmd.exec_cmd(cmd, wait=30)
        if sign != 0:
            run_log.error("Start NTP service failed.")
            return [1, "Failed to start ntp service"]

        return [0, "start ntp service ok"]
