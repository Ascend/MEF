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
import json
import os
import threading

import common.common_methods as CommonMethods
from common.file_utils import FileCheck
from common.log.logger import run_log

MAX_ALARM_SHIELD_LEN = 256


class AlarmShield(object):
    """
    功能描述：告警屏蔽
    接口：NA
    """
    sysLock = threading.Lock()
    shield_alarm_file_path = "/home/data/ies/shield_alarm.ini"
    shield_alarm_json_file_path = "/usr/local/mindx/MindXOM/software/ibma/config/all_alarm_for_manager.json"

    @staticmethod
    def split_alarm_shield(line):
        if not isinstance(line, str):
            run_log.error("line type invalidation!")
            return {}
        list_alrm = line.split("@")
        terminator = ""
        message_dict = {}
        if len(list_alrm) == 6:
            main_id = hex(int(list_alrm[1])).upper()[2:]
            sub_id = hex(int(list_alrm[2])).upper()[2:]
            message_dict = {
                "UniquelyIdentifies": list_alrm[0],
                "AlarmId": f"{main_id.zfill(4)}{sub_id.zfill(4)}",
                "PerceivedSeverity": list_alrm[3],
                "AlarmInstance": list_alrm[4]
            }
            terminator = list_alrm[5].strip('\n')
        if "aabb" in terminator:
            return message_dict
        return {}

    @staticmethod
    def check_alarm_shield(alarm_shield_messages_list):
        if alarm_shield_messages_list is None or alarm_shield_messages_list == "" or isinstance(
                alarm_shield_messages_list, int):
            return [400, "Alarm masking rules are illegal,AlarmShieldMessagesList is None"]
        all_alarm_shield_list = AlarmShield.get_all_alarm_role()
        if len(all_alarm_shield_list) > MAX_ALARM_SHIELD_LEN or len(alarm_shield_messages_list) > MAX_ALARM_SHIELD_LEN:
            return [400, "Alarm shield list oversize"]
        for alarm_clear_messages_dict in alarm_shield_messages_list:
            if not isinstance(alarm_clear_messages_dict, dict):
                return [400, "Alarm masking rules are illegal,alarm_clear_messages_dict is not dict"]
            for key in alarm_clear_messages_dict:
                if key not in ["UniquelyIdentifies", "AlarmId", "PerceivedSeverity", "AlarmInstance"]:
                    return [400, "Alarm masking key are illegal"]
            if alarm_clear_messages_dict not in all_alarm_shield_list:
                return [400, "Alarm masking rules are illegal,not in rules"]
        return [200, "Verify successfully"]

    @staticmethod
    def get_all_alarm_role():
        if not os.path.exists(AlarmShield.shield_alarm_json_file_path):
            run_log.error("shield_alarm_json_file_path is not exist")
            return []

        res = FileCheck.check_input_path_valid(AlarmShield.shield_alarm_json_file_path)
        if not res:
            run_log.error("shield_alarm_json_file_path is invalid : %s", res.error)
            return []

        with open(AlarmShield.shield_alarm_json_file_path) as fd:
            alarm_info_dict = json.loads(fd.read())

        all_alarm_shield_list = []
        try:
            for item in alarm_info_dict["LANG"]["EventSuggestion"]["item"]:
                alarm_shield_dict = {
                    "UniquelyIdentifies": item["@innerid"],
                    "AlarmId": item["id"],
                    "PerceivedSeverity": item["level"],
                    "AlarmInstance": item["AlarmInstance"]
                }
                all_alarm_shield_list.append(alarm_shield_dict)
        except Exception as e:
            run_log.error(f'{e}')

        return all_alarm_shield_list

    def get_all_info(self):
        self.read_alarm_shield()

    def read_alarm_shield(self):
        ret = FileCheck.check_path_is_exist_and_valid(AlarmShield.shield_alarm_file_path)
        if not ret:
            run_log.error("shield_alarm.ini invalid")
            return

        message_list = []
        with open(AlarmShield.shield_alarm_file_path, "r") as file:
            line_count = 0
            for line in file:
                if line is None or len(line) > 256 or line_count > 256:
                    break
                line_count += 1
                ret_dict = AlarmShield.split_alarm_shield(line)
                if ret_dict:
                    message_list.append(ret_dict)
            self.AlarmShieldMessages = message_list

    def patch_request(self, request_dict):
        if AlarmShield.sysLock.locked():
            run_log.warning("Alarm config modify is busy")
            return [400, "Alarm config modify is busy"]
        with AlarmShield.sysLock:
            if not isinstance(request_dict, dict):
                run_log.error("request_dict type invalidation!")
                return [400, "request_dict type invalidation!"]

            req_alarm_shield_list = request_dict.get("AlarmShieldMessages")
            ret = AlarmShield.check_alarm_shield(req_alarm_shield_list)
            if ret[0] != 200:
                run_log.error(ret)
                return ret

            if request_dict.get("Type") == "Increase":
                return self.increase_alarm_shield(req_alarm_shield_list)

            if request_dict.get("Type") == "Decrease":
                return self.decrease_alarm_shield(req_alarm_shield_list)

            message = "Alarm config modify failed because of invalid request."
            run_log.error(message)
            return [500, message]

    def write_alarm_shield(self, alarm_shield_messages_list) -> list:
        if alarm_shield_messages_list is None:
            return [400, "alarm shield messages list is None"]

        result_str = ""
        flags = os.O_WRONLY | os.O_CREAT | os.O_TRUNC
        try:
            for alarm_shield_dict in alarm_shield_messages_list:
                alarm_id = alarm_shield_dict.get("AlarmId")
                result_str += "%s@%s@%s@%s@%s@aabb\n" % (
                    alarm_shield_dict["UniquelyIdentifies"],
                    int(alarm_id[:4], base=16),
                    int(alarm_id[4:], base=16),
                    alarm_shield_dict["PerceivedSeverity"],
                    alarm_shield_dict["AlarmInstance"])

            FileCheck.check_is_link_exception(AlarmShield.shield_alarm_file_path)
            with os.fdopen(os.open(AlarmShield.shield_alarm_file_path, flags, 0o640), "w") as f:
                f.write(result_str)
                # 写入磁盘
                os.fsync(f)
        except Exception as err:
            run_log.error("write alarm shield failed, %s", err)
            return [400, "shield_alarm.ini Write filed"]

        self.read_alarm_shield()
        return [200, "shield_alarm.ini Write successful"]

    def split_different_shield_rule(self, req_alarm_shield_list):
        self.read_alarm_shield()
        shield_rule = self.AlarmShieldMessages if self.AlarmShieldMessages else []

        overlap_shield_rule = [alarm for alarm in req_alarm_shield_list if alarm in shield_rule]

        shielded_rule_minus_overlap = [alarm for alarm in shield_rule if alarm not in overlap_shield_rule]

        unshielded_rule_minus_overlap = [alarm for alarm in req_alarm_shield_list if alarm not in overlap_shield_rule]

        return overlap_shield_rule, shielded_rule_minus_overlap, unshielded_rule_minus_overlap

    def increase_alarm_shield(self, req_alarm_shield_list):
        overlap_shield_rule, _, unshielded_rule = \
            self.split_different_shield_rule(req_alarm_shield_list)
        if not unshielded_rule:
            run_log.warning("All of the alarm shield rules to be shielded in the request have been shield, "
                            "there is no need to create new the shield rule.")
            return [200, "No need to create new the shield rule."]

        if overlap_shield_rule:
            run_log.warning("There are some duplicate alarm shield rules in request and shielding list, "
                            "the system will remove the duplicate rules in the request and continue.")

        shield_rule = self.AlarmShieldMessages if self.AlarmShieldMessages else []
        need_create_shield_list = shield_rule + unshielded_rule
        return self.write_alarm_shield(need_create_shield_list)

    def decrease_alarm_shield(self, req_alarm_shield_list):
        overlap_shield_rule, need_create_shield_list, unshielded_rule = \
            self.split_different_shield_rule(req_alarm_shield_list)

        if not overlap_shield_rule:
            run_log.warning("All of the alarm shield rules to be unshielded in the request have not been shield, "
                            "there is no need to cancel the shield rule.")
            return [200, "No need to cancel the shield rule."]

        if unshielded_rule:
            run_log.warning("There are some alarm shield rules have not been shield in the request, "
                            "the system will ignore those unshielded rules and continue.")

        return self.write_alarm_shield(need_create_shield_list)
