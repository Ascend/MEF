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
import ctypes
import os
import threading
import time

from common.constants.base_constants import CertInfo
from common.constants.base_constants import CommonConstants
from common.constants.error_codes import SecurityServiceErrorCodes
from common.file_utils import FileCheck
from common.file_utils import FileCopy
from common.file_utils import FileCreate
from common.file_utils import FileUtils
from common.log.logger import run_log
from lib.Linux.systems.security_service.security_service_clib import cert_manage_finalize
from lib.Linux.systems.security_service.security_service_clib import cert_manage_init
from lib.Linux.systems.security_service.security_service_clib import create_cert_sign_request
from common.common_methods import CommonMethods


class SecurityServiceCSR:
    SECURITY_LOCK = threading.Lock()
    COPY_CUSTOM_CERT_PATH = "/home/data/cert_backup/default"
    CSR_DOWNLOAD_FLAG_TIMEOUT = 2 * 60 * 60  # 下载超时时间2小时

    def __init__(self):
        self.x509rep_content = ""

    @staticmethod
    def create_csr(cert_path):
        om_alg_path = os.path.join(CommonConstants.OM_WORK_DIR_PATH, "software/ibma/cert/om_alg.json")
        dst_path = os.path.join(cert_path, "om_alg.json")
        try:
            FileCopy.copy_file(om_alg_path, dst_path)
            ret_init = cert_manage_init(os.path.join(cert_path, "om_cert.keystore").encode("utf-8"),
                                        os.path.join(cert_path, "om_cert_backup.keystore").encode("utf-8"),
                                        os.path.join(cert_path, "om_alg.json").encode("utf-8"))
            if ret_init != 0:
                run_log.error("init kmc failed %s", ret_init)
                return False

            domain_name_str = CertInfo.gen_domain_name_with_uuid()
            domain_name = ctypes.create_string_buffer(domain_name_str.encode(encoding="utf-8"))
            create_cert_sign_request(cert_path.encode("utf-8"), domain_name)

            cert_manage_finalize()
            return True
        except Exception as err:
            run_log.error("Create csr failed. %s", err)
            return False

    @staticmethod
    def set_csr_flag(csr_path):
        csr_flag = os.path.join(csr_path, "csr_flag")
        ret = FileCheck.check_input_path_valid(csr_flag)
        if not ret:
            run_log.error("csr_flag path is invalid.")
            return False
        try:
            with os.fdopen(os.open(csr_flag, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600), "w"):
                run_log.info("create csr_flag success.")
        except Exception as err:
            run_log.error("Set csr download flag failed. %s", err)
            return False
        return True

    def read_csr(self, csr_dir, file_name):
        # 获取证书内容，并发送给redfish
        try:
            with open(os.path.join(csr_dir, file_name), "r") as file:
                self.x509rep_content = file.read()
                return True
        except Exception as err:
            run_log.error("read file %s failed, %s", file_name, err)
            return False

    def check_upload_mark_file_is_timeout(self, csr_download_flag_path):
        # 判断csr下载标志 超时时间
        if not os.path.exists(csr_download_flag_path):
            return [0, ""]
        modify_time = os.path.getmtime(csr_download_flag_path)
        current_time = time.time()
        if (current_time - modify_time) > self.CSR_DOWNLOAD_FLAG_TIMEOUT:
            run_log.info("upload mark flag %s timeout, delete it", csr_download_flag_path)
            FileUtils.delete_file_or_link(csr_download_flag_path)  # 超时清除标记
            return [0, ""]
        run_log.info("upload mark flag %s is not timeout", csr_download_flag_path)
        return [SecurityServiceErrorCodes.ERROR_CERTIFICATE_WAIT_TIME.code,
                "Please wait 2 hours and download it again"]

    def post_request(self, request_data):
        if SecurityServiceCSR.SECURITY_LOCK.locked():
            run_log.warning("SecurityServiceCSR is busy")
            return [CommonMethods.ERROR, "Create CSR failed, SecurityServiceCSR.SECURITY_LOCK.locked"]
        with SecurityServiceCSR.SECURITY_LOCK:
            if not os.path.exists(self.COPY_CUSTOM_CERT_PATH):
                ret = FileCreate.create_dir(self.COPY_CUSTOM_CERT_PATH, 0o700)
                if not ret:
                    run_log.error("Creat csr_path failed.")
                    return [CommonMethods.ERROR, "Creat csr_path failed."]

            # 判断csr下载标记
            flag_file = os.path.join(self.COPY_CUSTOM_CERT_PATH, "csr_flag")
            ret = self.check_upload_mark_file_is_timeout(flag_file)
            if ret[0] != 0:
                run_log.error("Please wait 2 hours and download it again")
                return [CommonMethods.ERROR, ret]

            if not self.create_csr(self.COPY_CUSTOM_CERT_PATH):
                return [CommonMethods.ERROR, "Download CSR failed."]

            if not self.read_csr(self.COPY_CUSTOM_CERT_PATH, "x509Req.pem"):
                return [CommonMethods.ERROR, "Create CSR failed, copy failed"]

            if not self.set_csr_flag(self.COPY_CUSTOM_CERT_PATH):
                return [CommonMethods.ERROR, "Add csr download flag failed"]

            return [CommonMethods.OK, "Create CSR Succeed"]
