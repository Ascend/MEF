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
import fcntl
import threading

from common.file_utils import FileCheck
from common.file_utils import FileUtils
from common.log.logger import run_log
from common.utils.app_common_method import AppCommonMethod
from common.checkers import RegexStringChecker
from common.utils.result_base import Result
from common.common_methods import CommonMethods


class EtcHostsManager(object):
    ETC_HOSTS_FILE = "/etc/hosts"
    ETC_HOSTS_LOCK = threading.Lock()
    ETC_HOSTS_LINE_LIMIT = 512

    @staticmethod
    def clear_etc_host(server_name: str) -> Result:
        server_name = server_name.strip()
        if not server_name:
            return Result(result=True)

        if not RegexStringChecker("server_name", "^[A-Za-z0-9-.]{1,64}$").check({"server_name": server_name}):
            err_msg = "sever name is not valid, can not clear"
            run_log.error(err_msg)
            return Result(result=False, err_msg=err_msg)

        ret = AppCommonMethod.get_etc_host_content_lines()
        if not ret:
            run_log.error("get etc host content lines failed: %s", ret.error)
            return ret

        try:
            hosts_info = EtcHostsManager._parse_etc_host(ret.data, server_name)
        except Exception as err:
            run_log.error("set etc hosts failed, find exception: %s", err)
            return Result(result=False, err_msg="set etc hosts failed, find exception")

        if not EtcHostsManager._write_etc_host_content(hosts_info):
            err_msg = "clear etc hosts failed, write etc host content failed"
            run_log.error(err_msg)
            return Result(result=False, err_msg=err_msg)

        return Result(result=True)

    @staticmethod
    def set_etc_host(net_ip, server_name):
        if not RegexStringChecker("server_name", "^[A-Za-z0-9-.]{1,64}$").check({"server_name": server_name}):
            run_log.warning("sever name is not valid, can not set etc hosts")
            return False

        if not net_ip or not isinstance(net_ip, str) or not AppCommonMethod.check_ipv4_format(net_ip.strip()):
            run_log.warning("net_ip is none or not valid")
            return False

        server_name = server_name.strip()
        net_ip = net_ip.strip()

        ret = AppCommonMethod.get_etc_host_content_lines()
        if not ret:
            run_log.error("get etc host content lines failed: %s", ret.error)
            return ret

        try:
            hosts_info = EtcHostsManager._parse_etc_host(ret.data, server_name)
        except Exception as err:
            run_log.error("set etc hosts failed: %s", err)
            return False

        write_content = "{}{} {}\n".format(hosts_info, net_ip, server_name)
        return EtcHostsManager._write_etc_host_content(write_content)

    @staticmethod
    def _parse_etc_host(lines, server_name):
        if len(lines) > EtcHostsManager.ETC_HOSTS_LINE_LIMIT:
            raise ValueError("lines exceeded limit.")

        hosts_info = ""
        for line in lines:
            if line == "\n":
                continue

            # server_name 不在此行，保留此行
            if server_name not in line.split():
                hosts_info = "{}{}\n".format(hosts_info, line.split("\n")[0])
                continue

            # server_name 在此行内，仅保留此行其他信息， 如果无其他信息，不保留该行
            split_str = line.split()
            # 分割后大于2 表示有2个以上域名对应一个ip 保留其他的域名 ip 映射
            if len(split_str) > 2:
                split_str.remove(server_name)
                hosts_info = "{}{}\n".format(hosts_info, " ".join(split_str))

        if not hosts_info:
            run_log.warning("hosts info is null")

        return hosts_info

    @staticmethod
    def _get_etc_host_content_lines():
        if not FileCheck.check_path_is_exist_and_valid(EtcHostsManager.ETC_HOSTS_FILE):
            run_log.error("get etc host content failed, invalid etc hosts path")
            return []

        try:
            with open(EtcHostsManager.ETC_HOSTS_FILE) as file:
                fcntl.flock(file.fileno(), fcntl.LOCK_EX)
                return file.readlines()

        except Exception as err:
            run_log.error("get etc host content failed, find exception: %s", err)
            return []

    @staticmethod
    def _write_etc_host_content(write_content):
        def inner_write_function(file, content):
            file.write(content)

        try:
            FileUtils.write_file_with_lock(EtcHostsManager.ETC_HOSTS_FILE, inner_write_function, write_content)
        except Exception as err:
            run_log.error("write etc host content failed, find exception: %s", err)
            return False

        return True

    def post_request(self, request_dict):
        if EtcHostsManager.ETC_HOSTS_LOCK.locked():
            run_log.error("config /etc/hosts is busy")
            return [CommonMethods.ERROR, "config /etc/hosts is busy"]

        with EtcHostsManager.ETC_HOSTS_LOCK:
            operate_type = request_dict.get("operate_type")
            if operate_type == "clear":
                result = self.clear_etc_host(request_dict.get("fd_server_name"))
                if not result:
                    run_log.error("clear /etc/hosts failed")
                    return [CommonMethods.ERROR, "clear /etc/hosts failed"]
            elif operate_type == "set":
                ret = self.set_etc_host(request_dict.get("fd_ip"), request_dict.get("fd_server_name"))
                if not ret:
                    run_log.error("config /etc/hosts failed")
                    return [CommonMethods.ERROR, "config /etc/hosts failed"]
            else:
                run_log.error("operate_type is invalid.")
                return [CommonMethods.ERROR, "config /etc/hosts failed, operate_type is invalid."]

            return [CommonMethods.OK, ""]
