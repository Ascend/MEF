# coding: UTF-8
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
import time

from common.constants.upload_constants import UploadConstants
from common.log.logger import run_log
from common.utils.app_common_method import AppCommonMethod


class UploadMarkFile:
    @staticmethod
    def check_upload_mark_file_is_invalid_by_filename(upload_filename: str) -> bool:
        """
        校验上传标记是否失效
        不存在或者存在但超时返回True，否则返回False
        :param upload_filename: 上传的文件名
        :return: boolean
        """
        file_format = upload_filename.rsplit('.', 1)[1] if '.' in upload_filename else ""
        mark_type = UploadConstants.EXT_MARK_TYPE.get(file_format)
        if mark_type in UploadConstants.MARK_FILES:
            return UploadMarkFile.check_upload_mark_file_is_invalid(UploadConstants.MARK_FILES[mark_type])

        run_log.error("Check upload mark flag failed: invalid upload file format")
        return False

    @staticmethod
    def check_upload_mark_file_is_invalid(mark_file: str):
        mark_file_basename = os.path.basename(mark_file)
        if os.path.exists(mark_file):
            run_log.info("upload mark flag (%s) is exist", mark_file_basename)
            return UploadMarkFile.check_upload_mark_file_is_timeout(mark_file)
        else:
            run_log.info("upload mark flag (%s) is not exist", mark_file_basename)
            return True

    @staticmethod
    def check_upload_mark_file_is_timeout(mark_file: str):
        mark_file_basename = os.path.basename(mark_file)
        modify_time = os.path.getmtime(mark_file)
        current_time = time.time()
        if abs(current_time - modify_time) > UploadConstants.TIMEOUT:
            run_log.info(f"upload mark flag({mark_file_basename}) timeout, delete it")
            UploadMarkFile.clear_upload_mark_file(mark_file)  # 超时清楚标记
            return True
        else:
            run_log.info(f"upload mark flag({mark_file_basename}) is not timeout")
            return False

    @staticmethod
    def clear_upload_mark_file(mark_file: str):
        mark_file_basename = os.path.basename(mark_file)
        if not os.path.exists(mark_file):
            run_log.info(f"upload mark flag({mark_file_basename}) not exist, no need to clear")
            return
        ret = AppCommonMethod.force_remove_file(mark_file)
        if not ret:
            run_log.error(f"clear upload mark flag({mark_file_basename}) failed: {ret.error}")
        else:
            run_log.info(f"clear upload mark flag({mark_file_basename}) success")

    @staticmethod
    def create_upload_mark_file_by_filename(upload_filename: str) -> bool:
        file_format = upload_filename.rsplit(".", 1)[1] if "." in upload_filename else ""
        mark_type = UploadConstants.EXT_MARK_TYPE.get(file_format)
        if mark_type in UploadConstants.MARK_FILES:
            return UploadMarkFile.create_upload_mark_file(UploadConstants.MARK_FILES[mark_type])

        run_log.error("Check upload mark flag failed: invalid upload file format(%s)", file_format)
        return False

    @staticmethod
    def create_upload_mark_file(mark_file: str):
        mark_file_basename = os.path.basename(mark_file)
        if not os.path.exists(mark_file):
            os.mknod(mark_file)
            os.chmod(mark_file, 0o600)
        run_log.info(f"Create upload mark flag({mark_file_basename}) success")
        return True

    @staticmethod
    def clear_upload_mark_file_all():
        for mark_file in UploadConstants.MARK_FILES.values():
            if not os.path.exists(mark_file):
                run_log.info("upload mark flag(%s) not exist, no need to clear", os.path.basename(mark_file))
                continue
            ret = AppCommonMethod.force_remove_file(mark_file)
            if not ret:
                run_log.error("clear upload mark flag %s failed: %s", os.path.basename(mark_file), ret.err_msg)
        run_log.info(f"clear all upload mark flag success")
