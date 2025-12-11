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
import fcntl
import os
import re
import threading
import time

from common.file_utils import FileCheck
from common.log.logger import run_log
from common.common_methods import CommonMethods

ALARM_FILE_EXTEND = "/run/all_active_alarm_extend"
CERT_ALARM_ID = "00180000"
FILE_MAX_SIZE = 10 * 1024  # 10KB
FILE_MAX_LINE = 256


class ModifyAlarm:
    """
    功能描述：修改告警文件内容
    """
    SECURITY_LOCK = threading.Lock()

    @staticmethod
    def rewrite_alarm(cert_name_list: list):
        """
        Rewrite the run alarm file.
        :param cert_name_list:Certificate name list
        """
        lines = []
        alarm_time = int(time.time())
        for cert_name in cert_name_list:
            alarm_content = f"{CERT_ALARM_ID}@Certificate alarm@{cert_name} FD CERT@{alarm_time}@1@aabb\n"
            lines.append(alarm_content)
            run_log.info("Add %s FD certificate about to expire alarm", cert_name)

        with os.fdopen(os.open(ALARM_FILE_EXTEND, os.O_RDWR | os.O_CREAT | os.O_TRUNC, 0o600), "w") as alarm_file:
            fcntl.flock(alarm_file.fileno(), fcntl.LOCK_EX)
            for line in lines:
                alarm_file.write(line)
            fcntl.flock(alarm_file.fileno(), fcntl.LOCK_UN)
        run_log.info("Modify FD cert alarm done.")

    @staticmethod
    def check_parameter(cert_name_list: list) -> bool:
        """
        Checking parameter validity
        :param cert_name_list:Certificate name list
        """
        if not isinstance(cert_name_list, list):
            run_log.error("Modify alarm failed, cert_name_list is not list")
            return False
        if len(cert_name_list) > FILE_MAX_LINE:
            run_log.error("Cert name list length too long")
            return False
        pattern_str = re.compile(r"^[a-zA-Z0-9_.]{4,64}$")
        for cert_name in cert_name_list:
            if not cert_name:
                run_log.error("Modify alarm failed, cert_name is empty")
                return False
            if pattern_str.fullmatch(cert_name) is None:
                run_log.error("Incorrect parameter cert_name")
                return False
        return True

    def modify_alarm_file(self, cert_name_list: list):
        """
        Modify the run alarm file.
        :param cert_name_list:Certificate name list
        """
        if not FileCheck.check_is_link(ALARM_FILE_EXTEND):
            return
        if not FileCheck.is_exists(ALARM_FILE_EXTEND):
            try:
                with os.fdopen(os.open(ALARM_FILE_EXTEND, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600), "w"):
                    run_log.info("Create all_active_alarm_extend success.")
            except Exception as err:
                run_log.error("Create all_active_alarm_extend failed. %s", err)
                return
        if os.path.getsize(ALARM_FILE_EXTEND) > FILE_MAX_SIZE:
            run_log.error("This %s file is too large.", ALARM_FILE_EXTEND)
            return
        self.rewrite_alarm(cert_name_list)

    def post_request(self, request_data):
        if ModifyAlarm.SECURITY_LOCK.locked():
            run_log.warning("Modify alarm is busy")
            return [CommonMethods.ERROR, "Modify alarm failed"]

        with ModifyAlarm.SECURITY_LOCK:
            cert_name_list = request_data.get("CertNameList")
            if not self.check_parameter(cert_name_list):
                return [CommonMethods.ERROR, "Modify alarm failed"]

            self.modify_alarm_file(cert_name_list)
            run_log.info("Check the validity period of the FD certificate succeed")
            return [CommonMethods.OK, "Modify alarm succeed"]
