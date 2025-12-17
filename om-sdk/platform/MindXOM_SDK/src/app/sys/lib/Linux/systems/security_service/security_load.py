#!/usr/bin/python
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import os
import threading
import time
from typing import Set

from common.checkers import SecurityLoadCfgChecker
from common.file_utils import FileCheck
from common.file_utils import FileUtils
from common.log.logger import run_log
from common.net_check import NetCheck
from common.utils.app_common_method import AppCommonMethod
from common.utils.result_base import Result
from lib.Linux.systems.security_service.login_rule_mgr import LoginRuleManager
from lib.Linux.systems.security_service.models import LoginRules
from common import common_methods as commMethods


class SecurityLoad:
    """
    功能描述：安全登录规则
    接口：NA
    """

    CHECK_REQUEST_DATA_FAIL = 701
    SAVE_CONDIF_DATA_FAIL = 702
    IMPORT_PARAM_INVALID = 703
    IMPORT_FILE_NOT_EXIST = 704
    IMPORT_CFG_INVALID = 705
    IMPORT_CFG_FILE_WOUNDED = 706
    LOCAL_CFG_FILE_WOUNDED = 707
    CREATE_EMPTY_FILE_FAIL = 708
    IMPORT_FILE_TOO_LARGE = 709
    REQUEST_DATA_REPEATED = 710
    ERROR_ROOT_OPERATE_LOCKED = 711
    FILE_IS_INVALID = 712
    COPY_FAILED = 713

    SECURITY_LOCK = threading.Lock()

    def __init__(self):
        """
         功能描述：初始化函数
         参数：
         返回值：无
         异常描述：NA
        """
        self.load_cfg = []

    @staticmethod
    def get_all_arp_item():
        arp_ip_list = []
        arp_mac_list = []
        with open("/proc/net/arp", "r") as file_read:
            for line in file_read:
                if line.startswith("IP address"):
                    continue
                line = line.split()
                if len(line) >= 4:
                    arp_ip_list.append(line[0])
                    arp_mac_list.append(line[3])

        return arp_ip_list, arp_mac_list

    @staticmethod
    def check_remote_request_ip(remote_ip, expect_ip) -> bool:
        if '/' not in expect_ip:
            return remote_ip == expect_ip
        cfg_ip, cfg_mask = expect_ip.split("/")[:2]
        return NetCheck.is_same_network_segment(remote_ip, cfg_ip, cfg_mask)

    @staticmethod
    def check_remote_request_mac(remote_ip, expect_mac):
        comp_len = len(expect_mac)
        arp_ip, arp_mac = SecurityLoad.get_all_arp_item()
        if not arp_ip:
            return True
        if remote_ip in arp_ip:
            idx = arp_ip.index(remote_ip)
            if expect_mac.upper() == arp_mac[idx].upper()[:comp_len]:
                return True
        run_log.info("remote ip(%s) not meet cfg mac(%s)" % (remote_ip, expect_mac.upper()))
        return False

    @staticmethod
    def check_usr_load_time(cfg_item) -> bool:
        start = cfg_item['start_time']
        end = cfg_item['end_time']
        if start is None or end is None:
            return True
        try:
            start_hour, start_min, end_hour, end_min = map(int, (*start.split(':')[:2], *end.split(':')[:2]))
        except Exception as err:
            run_log.error("convert time error: %s" % err)
            return True
        curr = time.localtime()
        run_log.info("curr time %s:%s, cfg time %s - %s" % (curr[3], curr[4], start, end))

        start_time = start_hour * 60 + start_min
        end_time = end_hour * 60 + end_min
        cur_time = curr[3] * 60 + curr[4]
        if start_time <= cur_time <= end_time:
            return True
        # 跨零时
        if end_time < start_time:
            if cur_time >= start_time or cur_time <= end_time:
                return True
        return False

    @staticmethod
    def check_one_session_sec_cfg(request_ip, cfg_item):
        if not SecurityLoad.check_usr_load_time(cfg_item):
            return False
        run_log.info("load cfg, check expect ip: %s" % cfg_item['ip_addr'])
        ip_addr = cfg_item['ip_addr']
        if ip_addr and not SecurityLoad.check_remote_request_ip(request_ip, ip_addr):
            return False
        mac_addr = cfg_item['mac_addr']
        run_log.info(f"load cfg, check expect mac: {mac_addr}")
        if mac_addr and not SecurityLoad.check_remote_request_mac(request_ip, mac_addr):
            return False
        run_log.info("load cfg met.")
        return True

    @staticmethod
    def get_usr_session_sec_cfg(spec_cfg_file=None) -> list:
        sec_cfg = []
        try:
            cfg_parser = FileUtils.get_config_parser(spec_cfg_file)
        except Exception as err:
            run_log.error("get session sec cfg error: %s" % err)
            return []
        for section in cfg_parser.sections():
            single_cfg = {}
            for option in ("enable", "start_time", "end_time", "ip_addr", "mac_addr"):
                try:
                    value = cfg_parser.get(section, option)
                except Exception as err:
                    run_log.error("get session sec cfg error: %s" % err)
                    return []
                single_cfg[option] = value if value != "None" else None
            sec_cfg.append(single_cfg)
        return sec_cfg

    @staticmethod
    def check_session_request_usr_data(request_ip):
        """
        功能描述：检查登录用户是否满足登录规则
        参数：
        返回值：无
        异常描述：NA
        """
        enable_num = 0
        sec_cfg = LoginRuleManager().get_all()
        if not sec_cfg:
            run_log.info("no available session sec cfg")
            return True

        for cfg_item in sec_cfg:
            if cfg_item['enable'] != 'true':
                continue
            enable_num += 1
            if SecurityLoad.check_one_session_sec_cfg(request_ip, cfg_item):
                return True
        if enable_num == 0:
            run_log.info("no session sec cfg is enable")
            return True
        run_log.info("remote ip is %s, no session sec cfg is met" % request_ip)
        return False

    @staticmethod
    def check_patch_request_data(request_dict):
        """
        功能描述：验证patch请求的请求参数
        参数：
        返回值：无
        异常描述：NA
        """
        if request_dict is None or not isinstance(request_dict, dict):
            return Result(result=False, err_msg="Request parameter is empty or not dict")

        try:
            req_data = request_dict.get("load_cfg", None)
        except Exception as err:
            return Result(result=False, err_msg=f"Request parameter get load_cfg {err}")

        if req_data is None or not isinstance(req_data, list):
            return Result(result=False, err_msg="Request sub parameter is empty or not list")
        if len(req_data) > 30:
            return Result(result=False, err_msg="Request parameter list len exceeds")

        for item in req_data:
            try:
                check_ret = SecurityLoadCfgChecker().check(item)
            except Exception as err:
                return Result(result=False, err_msg=f"Request sub parameter is not dict {err}")
            if not check_ret.success:
                return Result(result=False, err_msg="SecurityLoadCfg Parameter error")

        return Result(result=True)

    @staticmethod
    def check_post_request_data(request_dict):
        """
        功能描述：验证post请求的请求参数
        参数：
        返回值：无
        异常描述：NA
        """
        if request_dict is None:
            return Result(result=False, err_msg="request parameter is empty")
        action = request_dict.get("action", None)
        if action is None or action not in ("import", "export"):
            return Result(result=False, err_msg="action is invalid")

        if action == 'import':
            file_name = request_dict.get("file_name", None)
            if file_name is None or AppCommonMethod.check_input_parm(file_name) is False:
                return Result(result=False, err_msg="file_name is none or invalid")
            if file_name[-4:] != '.ini':
                return Result(result=False, err_msg="file type invalid")
        return Result(result=True)

    @staticmethod
    def get_part_same_db_rule(req_item: LoginRules, db_data: Set[LoginRules]):
        for db_item in db_data:
            if req_item.is_part_same(db_item):
                return [0, db_item]
        return [-1, None]

    @staticmethod
    def record_diff_rule_opera_type(req_data: Set[LoginRules], db_data: Set[LoginRules]):
        add_list = []
        modify_list = []
        for item in req_data:
            ret = SecurityLoad.get_part_same_db_rule(item, db_data)
            if ret[0] == 0:
                modify_list.append(ret[1].to_dict())
                db_data.remove(ret[1])
                continue

            add_list.append(item.to_dict())

        delete_list = [rule.to_dict() for rule in db_data]

        if add_list:
            run_log.warning("those login rule will be added: %s", add_list)

        if modify_list:
            run_log.warning("those login rule will be modified: %s", modify_list)

        if delete_list:
            run_log.warning("those login rule will be deleted: %s", delete_list)

    def get_all_info(self):
        """
        功能描述：解析get请求
        参数：
        返回值：无
        异常描述：NA
        """
        self.load_cfg = LoginRuleManager().get_all()

    def patch_request(self, request_dict):
        """
        功能描述：解析patch请求,用于web界面保存登录规则。
        参数：
        返回值：无
        异常描述：NA
        """
        if SecurityLoad.SECURITY_LOCK.locked():
            run_log.warning("SecurityLoad modify is busy")
            return [commMethods.CommonMethods.ERROR,
                    "ERR.%s, SecurityLoad modify is busy." % SecurityLoad.ERROR_ROOT_OPERATE_LOCKED]
        with SecurityLoad.SECURITY_LOCK:
            ret = self.check_patch_request_data(request_dict)
            if not ret:
                run_log.error("check patch request data fail: %s", ret.error)
                return [
                    commMethods.CommonMethods.ERROR,
                    "ERR.%s, Check patch request data fail" % SecurityLoad.CHECK_REQUEST_DATA_FAIL
                ]

            cfg_list = request_dict.get("load_cfg", None)
            req_rules = set(LoginRules.to_obj(cfg) for cfg in cfg_list)
            if len(req_rules) != len(cfg_list):
                run_log.error("Patch request data repeated")
                return [
                    commMethods.CommonMethods.ERROR,
                    "ERR.%s, Patch request data repeated" % SecurityLoad.REQUEST_DATA_REPEATED
                ]
            manager = LoginRuleManager()
            db_rules = {LoginRules.to_obj(rule) for rule in manager.get_all()}
            same_rules = req_rules.intersection(db_rules)
            deduplication_req_rules = req_rules.difference(same_rules)
            deduplication_db_rules = db_rules.difference(same_rules)

            self.record_diff_rule_opera_type(deduplication_req_rules, deduplication_db_rules)

            try:
                manager.over_write_database(req_rules)
            except Exception as err:
                run_log.error("set config data failed: %s" % err)
                return [
                    commMethods.CommonMethods.ERROR,
                    "ERR.%s, Set config data fail" % SecurityLoad.SAVE_CONDIF_DATA_FAIL
                ]

            return [commMethods.CommonMethods.OK, "set config data ok"]

    def post_request(self, request_dict):
        """
        功能描述：解析post请求
        参数：
        返回值：无
        异常描述：NA
        """
        if SecurityLoad.SECURITY_LOCK.locked():
            run_log.warning("SecurityLoad modify is busy")
            return [commMethods.CommonMethods.ERROR,
                    "ERR.%s, SecurityLoad modify is busy." % SecurityLoad.ERROR_ROOT_OPERATE_LOCKED]
        with SecurityLoad.SECURITY_LOCK:
            ret = self.check_post_request_data(request_dict)
            if not ret:
                run_log.error("Check upload para failed, %s", ret.error)
                return [commMethods.CommonMethods.ERROR,
                        "ERR.%s, Upload parameter error." % SecurityLoad.IMPORT_PARAM_INVALID]
            action = request_dict.get("action", None)

            if action == 'import':
                file_name = request_dict.get("file_name", None)
                ret = self.import_session_cfg_file(file_name)
                if not ret:
                    run_log.error("Import Security load config failed %s", ret.error)
                    return [commMethods.CommonMethods.ERROR, ret.error]
                run_log.info("Import Security load config ok")
                return [commMethods.CommonMethods.OK, "Import security load config ok"]

            elif action == 'export':
                # 将获取到的登录规则返回给redfish，由redfish进程持久化到文件中，提供用户下载
                self.load_cfg = LoginRuleManager().get_all()
                run_log.info("get Security load config ok")
                return [commMethods.CommonMethods.OK, "get Security load config ok"]
            return [commMethods.CommonMethods.ERROR, "action not support"]

    def import_session_cfg_file(self, file_name):
        import_file = os.path.join('/run/web/ini', file_name)
        if not FileCheck.check_path_is_exist_and_valid(import_file):
            return Result(result=False, err_msg=f"ERR.{SecurityLoad.FILE_IS_INVALID}")

        file_size = os.path.getsize(import_file)
        if file_size > 10 * 1024:
            run_log.error("Import file too large(%s)." % file_size)
            return Result(result=False, err_msg=f"ERR.{SecurityLoad.IMPORT_FILE_TOO_LARGE}, Import file is too large.")

        # 校验字段
        login_rules_list = self.get_usr_session_sec_cfg(import_file)
        cfg_dict = {"load_cfg": login_rules_list}
        ret = self.check_patch_request_data(cfg_dict)
        if not ret:
            run_log.error("check import file config failed: %s", ret.error)
            return Result(result=False,
                          err_msg=f"ERR.{SecurityLoad.IMPORT_CFG_INVALID}, Check import file configuration failed.")

        # 将数据保存至数据库中
        obj_list = {LoginRules.to_obj(item) for item in login_rules_list}
        try:
            LoginRuleManager().over_write_database(obj_list)
        except Exception as err:
            run_log.error("Import login rule file failed: %s" % err)
            return Result(result=False, err_msg=f"ERR.{SecurityLoad.COPY_FAILED}, Import login rule file failed.")

        os.sync()
        time.sleep(1)
        return Result(result=True)
