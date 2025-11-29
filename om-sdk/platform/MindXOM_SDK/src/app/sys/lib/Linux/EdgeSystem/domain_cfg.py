#!/usr/bin/python
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

import os
import time

from common.checkers import LocalIpChecker
from common.constants.base_constants import CommonConstants
from common.constants.config_constants import ConfigPathConstants
from common.file_utils import FileCheck
from common.log.logger import run_log
from common.utils.app_common_method import AppCommonMethod
from common.checkers import IPV4Checker
from common.common_methods import CommonMethods


class DomainCfg:

    @staticmethod
    def host_config(host_list, net_manager_domain):
        system_cfg_info = ""
        local_name_server = [net_manager_domain]
        exist_count = 0
        try:
            with open("/etc/hosts", "r") as f:
                for line in f:
                    ip_name_list = line.split()
                    if len(ip_name_list) > 1 and (ip_name_list[0] in ("127.0.0.1", "::1") or
                                                  ip_name_list[1] == net_manager_domain):
                        system_cfg_info = system_cfg_info + line
                        local_name_server.extend(line.split(" ")[1:])
                        exist_count += 1

                    if exist_count >= 128:
                        run_log.error("/etc/hosts lines are excessive.")
                        return

        except Exception as e:
            run_log.error(f"Read /etc/hosts file failed, {e}")
            return

        for index in host_list:
            if index["name"] not in local_name_server:
                system_cfg_info = system_cfg_info + index["ip_address"] + " " + index["name"] + "\n"

        if not FileCheck.check_is_link("/etc/hosts"):
            run_log.error("Check /etc/hosts is link file.")
            return

        try:
            with os.fdopen(os.open("/etc/hosts", os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o644), "w+") as f:
                f.write(system_cfg_info)
        except Exception as e:
            run_log.error("failed:{}".format(e))

    @staticmethod
    def dns_config(name_server):
        system_cfg_info = ""
        exist_count = 0
        resolv_config = ConfigPathConstants.ETC_RESOLV_PATH
        if not os.path.exists(resolv_config):
            # openEuler22.03默认不存在该文件，配置时先创建该文件
            with os.fdopen(os.open(resolv_config, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o644), "w"):
                run_log.info("The file %s not exists, create it success", resolv_config)

        ret = FileCheck.check_input_path_valid(resolv_config)
        if not ret:
            run_log.error("Check %s path invalid: %s", resolv_config, ret.error)
            return

        if os.path.getsize(resolv_config) > CommonConstants.OM_READ_FILE_MAX_SIZE_BYTES:
            run_log.error("Check %s path invalid: file size is too large", resolv_config)
            return

        try:
            with open(resolv_config, "r") as f:
                for line in f:
                    if "nameserver" not in line:
                        system_cfg_info = system_cfg_info + line
                        exist_count += 1

                    if exist_count >= 128:
                        run_log.error("%s lines are excessive.", resolv_config)
                        return

        except Exception as e:
            run_log.error("Read %s file failed, caught exception: %s", resolv_config, e)
            return

        for index in name_server:
            system_cfg_info = system_cfg_info + "nameserver" + " " + index["ip_address"] + "\n"

        if not FileCheck.check_is_link(resolv_config):
            run_log.error("Check %s is link file.", resolv_config)
            return

        try:
            with os.fdopen(os.open(resolv_config, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o644), "w") as f:
                f.write(system_cfg_info)
        except Exception as e:
            run_log.error("Write %s file failed, caught exception: %s", resolv_config, e)

    @staticmethod
    def get_host_list():
        host_list = []
        try:
            with open("/etc/hosts", "r") as f:
                for line in f:
                    if "127.0.0.1" not in line and "::1" not in line:
                        host_element = {}
                        line_list = line.split(" ")
                        if len(line_list) != 2:
                            continue
                        host_element["ip_address"] = line_list[0]
                        host_element["name"] = line_list[1].strip("\n")
                        if AppCommonMethod.check_ipv4_format(host_element["ip_address"]) \
                                and IPV4Checker("domain").domain_check(host_element["name"]):
                            host_list.append(host_element)
        except Exception as e:
            run_log.error("failed:{}".format(e))
        return host_list

    @staticmethod
    def get_dns_list():
        dns_list = []
        resolv_config = ConfigPathConstants.ETC_RESOLV_PATH
        # openEuler22.03可能不存在该文件，返回默认值，ubuntu该文件为软链接
        if not os.path.exists(resolv_config):
            return dns_list

        try:
            with open(resolv_config, "r") as f:
                for line in f:
                    dns_element = {}
                    line_list = line.split(" ")
                    if len(line_list) != 2:
                        continue
                    nameserver = line_list[0]
                    dns_element["ip_address"] = line_list[1].strip("\n")
                    if dns_element["ip_address"] == "0.0.0.0" or dns_element["ip_address"] == "255.255.255.255":
                        continue
                    if AppCommonMethod.check_ipv4_format(dns_element["ip_address"]) and nameserver == "nameserver":
                        dns_list.append(dns_element)
        except Exception as e:
            run_log.error("Read %s file failed, caught exception: %s", resolv_config, e)
        return dns_list

    @staticmethod
    def request_data_check(request_data_dict):
        if request_data_dict is None:
            return [1, "request parameter is empty"]

        static_host_list = request_data_dict.get("static_host_list", [])
        name_server = request_data_dict.get("name_server", [])
        net_manager_domain = request_data_dict.get("net_manager_domain")
        if not isinstance(static_host_list, list) or not isinstance(name_server, list):
            return [2, "invalid parameter"]

        if len(static_host_list) > 128 or len(name_server) > 3:
            return [3, "invalid list len"]

        if not IPV4Checker("domain").domain_check(net_manager_domain):
            return [3, "invalid net manager domain"]

        for index in static_host_list:
            ip_address = index["ip_address"]
            if not AppCommonMethod.check_ipv4_format(ip_address) or "127.0.0.1" == ip_address or "::1" == ip_address:
                return [3, "invalid ip format"]

            ret = LocalIpChecker("ip_address").check_dict({"ip_address": ip_address})
            if not ret:
                return [3, f"invalid static host: {ret.reason}"]

            if not IPV4Checker("domain").domain_check(index["name"]) or \
                    DomainCfg.is_forbidden_domain(index["name"], net_manager_domain):
                return [4, "invalid domain format"]

        for index in name_server:
            ip_address = index["ip_address"]
            if not AppCommonMethod.check_ipv4_format(ip_address):
                return [5, "invalid ip format"]

            if ip_address == "0.0.0.0" or ip_address == "255.255.255.255":
                return [6, "invalid name server"]

            ret = LocalIpChecker("ip_address").check_dict({"ip_address": ip_address})
            if not ret:
                return [6, f"invalid name server: {ret.reason}"]

        return [0, "ok"]

    @staticmethod
    def is_forbidden_domain(domain, net_manager_domain):
        return domain == net_manager_domain or domain in AppCommonMethod.FORBIDDEN_DOMAINS

    @staticmethod
    def post_request(request_data_dict):
        try:
            ret = DomainCfg.request_data_check(request_data_dict)
            if ret[0] != 0:
                run_log.error("Check input para failed, %s" % ret[1])
                error_message = "ERR.300, Check input para failed, {}".format(ret[1])
                return [CommonMethods.ERROR, error_message]

            static_host_list = request_data_dict.get("static_host_list")
            if static_host_list:
                DomainCfg.host_config(static_host_list, request_data_dict.get("net_manager_domain"))

            name_server = request_data_dict.get("name_server")
            if name_server:
                DomainCfg.dns_config(name_server)

            run_log.info("Config domain success: host:{} name_server:{}".format(static_host_list, name_server))
            return [CommonMethods.OK, "OK"]
        except Exception as e:
            run_log.error("post_request ManageInfo: " + str(e))
            return [CommonMethods.ERROR, "ERR.400 PostRequest DomainCfg failed."]
        finally:
            os.sync()
            time.sleep(3)

    def get_all_info(self):
        self.static_host_list = DomainCfg.get_host_list()
        self.name_server = DomainCfg.get_dns_list()
