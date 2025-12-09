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

from common.file_utils import FileCheck
from common.log.logger import run_log


class Alarm:
    """
    功能描述：告警信息
    接口：NA
    修改记录：2019-3-20 创建
    """
    MAX_ALARM_COUNT = 512
    ALARM_FILE = "/run/all_active_alarm"
    ALARM_FILE_EXTEND = "/run/all_active_alarm_extend"
    SHIELD_ALARM_FILE_PATH = "/home/data/ies/shield_alarm.ini"
    FD_CERT_SIGN = "FD CERT"

    def __init__(self):
        """
        功能描述：初始化函数
        参数：NA
        返回值：无
        异常描述：NA
        修改记录：2019-3-20 创建
        """
        self.AlarMessages = []

    @staticmethod
    def getallalarmdev():
        ret = FileCheck.check_path_is_exist_and_valid("/run/all_active_alarm")
        if not ret:
            run_log.warning("all_active_alarm file invalid: %s", ret.error)
            return []

        dev_list = []
        with open("/run/all_active_alarm", "r") as file:
            for line in file:
                if line is None:
                    continue
                list_alrm = line.split("@")
                if len(list_alrm) != 6:
                    continue
                terminator = list_alrm[5].strip('\n')
                if "aabb" in terminator and list_alrm[2] not in dev_list:
                    dev_list.append(list_alrm[2])
                if len(dev_list) >= Alarm.MAX_ALARM_COUNT:
                    run_log.warning(f"alarm_dev_list exceeding the max_alarm_count")
                    break

        return dev_list

    def get_shield_flag_by_inner_id(self, inner_id: str) -> bool:
        ret = FileCheck.check_path_is_exist_and_valid(self.SHIELD_ALARM_FILE_PATH)
        if not ret:
            run_log.warning("%s path invalid : %s.", self.SHIELD_ALARM_FILE_PATH, ret.error)
            return False

        with open(self.SHIELD_ALARM_FILE_PATH, "r") as shield_file:
            for line_count, line in enumerate(shield_file):
                if not line or len(line) > 256 or line_count > 256:
                    break
                list_alarm = line.split("@")
                if len(list_alarm) != 6 or "aabb" not in list_alarm[5]:
                    continue
                if list_alarm[0] == inner_id:
                    return True
        return False

    def get_all_info(self):
        self.get_file_info(self.ALARM_FILE)
        self.get_file_info(self.ALARM_FILE_EXTEND)

    def get_file_info(self, alarm_file):
        # 告警文件不存在时，直接返回，避免日志刷屏
        if not os.path.exists(alarm_file):
            return

        ret = FileCheck.check_input_path_valid(alarm_file)
        if not ret:
            run_log.warning("%s file invalid: %s.", alarm_file, ret.error)
            return

        with open(alarm_file, "r") as file:
            for line_count, line in enumerate(file):
                if line is None or len(line) > 256 or line_count > 256:
                    break
                list_alarm = line.split("@")
                if len(list_alarm) != 6 or "aabb" not in list_alarm[5] or self._is_shield_alarm(line):
                    continue
                message_dict = {
                    "AlarmId": list_alarm[0],
                    "AlarmName": list_alarm[1],
                    "AlarmInstance": list_alarm[2],
                    "Timestamp": list_alarm[3],
                    "PerceivedSeverity": list_alarm[4],
                }
                self.AlarMessages.append(message_dict)

    def _is_shield_alarm(self, line: str) -> bool:
        if self.get_shield_flag_by_inner_id('w000000000') and self.FD_CERT_SIGN in line:
            return True
        return False
