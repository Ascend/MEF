#!/usr/bin/python
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

import os
import threading
from functools import partial

from common.file_utils import FileCheck
from common.file_utils import FilePermission
from common.log.logger import run_log
from common.checkers import IntegerChecker
from common.common_methods import CommonMethods


class CertAlarmTime:
    """
    功能描述：证书过期提醒时间
    接口：NA
    修改记录：2019-12-23 创建
    """
    CERT_ALARM_TIME_MANAGE_LOCK = threading.Lock()
    CommonCertAlarmTimeFile = "/home/data/ies/certWarnTime.ini"
    DefaultCertAlarmTime = "10"
    Max_cert_alarm_time = 180
    Min_cert_alarm_time = 7

    def __init__(self):
        """
        功能描述：初始化函数
        参数：NA
        返回值：无
        异常描述：NA
        修改记录：2019-3-20 创建
        """

        self.CertAlarmTime = int(self.DefaultCertAlarmTime)

    def get_all_info(self):
        self.get_https_cert_alarm_time_info()

    def get_https_cert_alarm_time_info(self):
        if FileCheck.check_path_is_exist_and_valid(self.CommonCertAlarmTimeFile):
            try:
                with open(self.CommonCertAlarmTimeFile, "r") as file:
                    self.CertAlarmTime = int(file.readline().strip())
            except Exception as err:
                run_log.error("Open CommonCertAlarmTimeFile Failed %s", err)
            return

        try:
            flags = os.O_WRONLY | os.O_CREAT | os.O_TRUNC
            with os.fdopen(os.open(self.CommonCertAlarmTimeFile, flags, 0o640), "w") as file:
                file.write(self.DefaultCertAlarmTime)
        except Exception as err:
            run_log.error("Create CommonCertAlarmTimeFile Failed %s" % err)

        # 父目录权限为700, os.fdopen指定权限也无法把文件所属组权限设置成功, 此处单独再设置一次权限
        FilePermission.set_path_permission(self.CommonCertAlarmTimeFile, 0o640)

    def patch_request(self, request_data_dict):
        if request_data_dict is None or not isinstance(request_data_dict, dict):
            return [1, "Request parameter is empty or not dict"]
        cert_alarm_time = request_data_dict.get('CertAlarmTime', None)
        ret = IntegerChecker(
            "CertAlarmTime",
            min_value=self.Min_cert_alarm_time,
            max_value=self.Max_cert_alarm_time).check({"CertAlarmTime": cert_alarm_time})
        if not ret.success:
            run_log.error("Parameter CertAlarmTime cheack error")
            return [CommonMethods.ERROR, "Parameter CertAlarmTime error"]
        if CertAlarmTime.CERT_ALARM_TIME_MANAGE_LOCK.locked():
            run_log.warning("CertAlarm modify is busy")
            return [CommonMethods.ERROR, "CertAlarm modify is busy"]
        with CertAlarmTime.CERT_ALARM_TIME_MANAGE_LOCK:
            for filename in (self.CommonCertAlarmTimeFile, "/run/certWarnTimeUpdate"):
                if not FileCheck.check_is_link(filename):
                    run_log.error("path might link.")
                    return [CommonMethods.ERROR, "Setting CertAlarmTime Failed"]

            os_open = partial(os.open, flags=os.O_WRONLY | os.O_CREAT | os.O_TRUNC, mode=0o640)
            try:
                with os.fdopen(os_open(self.CommonCertAlarmTimeFile), "w") as file:
                    file.write(str(cert_alarm_time))
            except Exception as err:
                run_log.error("Setting CertAlarmTime Failed %s" % err)
                return [CommonMethods.ERROR, "Setting CertAlarmTime Failed"]

            # 父目录权限为700, os.fdopen指定权限也无法把文件所属组权限设置成功, 此处单独再设置一次权限
            FilePermission.set_path_permission(self.CommonCertAlarmTimeFile, 0o640)
            # 创建标识文件
            try:
                with os.fdopen(os_open("/run/certWarnTimeUpdate"), "w+") as file:
                    run_log.info("CertWarnTimeUpdate has exists")
            except Exception as err:
                run_log.error("Setting CertAlarmTime Failed %s" % err)
                return [CommonMethods.ERROR, "Setting CertAlarmTime Failed"]

            self.get_https_cert_alarm_time_info()
            return [CommonMethods.OK, "Setting CertAlarmTime success"]
