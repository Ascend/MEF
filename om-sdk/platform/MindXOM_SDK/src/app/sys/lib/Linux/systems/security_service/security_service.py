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
import os
import re
import shutil
import threading
from datetime import datetime

import common.common_methods as commMethods
from common.constants.error_codes import SecurityServiceErrorCodes
from common.constants.upload_constants import UploadConstants
from common.file_utils import FileCheck
from common.file_utils import FileCopy
from common.file_utils import FileCreate
from common.file_utils import FileUtils
from common.log.logger import run_log
from common.kmc_lib.kmc import Kmc
from common.kmc_lib.tlsconfig import TlsConfig

CERT_PATH = "/home/data/config/default"
CERT_BACKUP_PATH = "/home/data/cert_backup"
COPY_CUSTOM_CERT_PATH = "/home/data/cert_backup/default"


class SecurityService:
    """
    功能描述：证书信息
    接口：NA
    """
    SECURITY_LOCK = threading.Lock()
    CertupdateImportFile = "/run/certupdate"
    DEFAULT_CERT_PATH = "/home/data/config/default/server_kmc.cert"

    def __init__(self):
        """
        功能描述：初始化函数
        参数：NA
        返回值：无
        异常描述：NA
        """

        self.Subject = None
        self.Issuer = None
        self.ValidNotBefore = None
        self.ValidNotAfter = None
        self.SerialNumber = None
        self.FingerPrint = None
        self.HttpsCertEable = None
        self.ExpiredDayRemaining = None
        self.items = None

    @staticmethod
    def cheack_parameter(custom_certificate_filename):
        # 任意值为空则错
        if custom_certificate_filename is None:
            run_log.error("Import custom certificate failed.")
            return [commMethods.CommonMethods.ERROR, "Import custom certificate failed."]
        reg_str = "^[a-zA-Z0-9_.-]{0,255}$"
        pattern_str = re.compile(reg_str)
        if pattern_str.fullmatch(custom_certificate_filename) is None or ".." in custom_certificate_filename:
            run_log.error("Incorrect parameter name : %s", custom_certificate_filename)
            return [commMethods.CommonMethods.ERROR, "Incorrect parameter name"]
        if custom_certificate_filename[-4:] not in (".crt", ".cer"):
            run_log.error("Incorrect parameter type : %s", custom_certificate_filename[-4:])
            return [commMethods.CommonMethods.ERROR, "Incorrect parameter type"]
        return [commMethods.CommonMethods.OK, ""]

    @staticmethod
    def create_certupdate():
        if os.path.exists(SecurityService.CertupdateImportFile):
            run_log.info("%s has exists" % SecurityService.CertupdateImportFile)
        else:
            if not FileCheck.check_is_link(SecurityService.CertupdateImportFile):
                run_log.error("path might be link")
                return
            try:
                with os.fdopen(os.open(SecurityService.CertupdateImportFile,
                                       os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600), "w+"):
                    run_log.info("%s has exists" % SecurityService.CertupdateImportFile)
            except Exception as err:
                run_log.error(f'{err}')

    @staticmethod
    def import_certificate(custom_certificate_filepath):
        primary_ksf = os.path.join(COPY_CUSTOM_CERT_PATH, "om_cert.keystore")
        standby_ksf = os.path.join(COPY_CUSTOM_CERT_PATH, "om_cert_backup.keystore")
        cfg_path = os.path.join(COPY_CUSTOM_CERT_PATH, "om_alg.json")
        kmc = Kmc(primary_ksf, standby_ksf, cfg_path)
        # 调用导入方法
        result = SecurityService.import_cert(custom_certificate_filepath, kmc)
        if result[0] == 0:
            SecurityService.create_certupdate()
            SecurityService._set_cert_sum()
            return [commMethods.CommonMethods.OK, result]
        else:
            if result[0] == SecurityServiceErrorCodes.ERROR_IMPORT_CERTIFICATE_LEGALITY_RISKY.code:
                run_log.error("Import custom certificate successfully but legality is risky")
                SecurityService.create_certupdate()
                SecurityService._set_cert_sum()
            else:
                run_log.error("Importing a custom certificate failed %s" % result[0])
            return [commMethods.CommonMethods.ERROR, result]

    @staticmethod
    def import_cert(custom_certificate_filepath, kmc: Kmc):
        try:
            from lib.Linux.systems.security_service.security_service_clib import certificate_verification, \
                ERR_SIGNATURE_LEN_NOT_ENOUGH, check_cert_expired
        except Exception as err:
            run_log.error("Import custom certificate failed. The reason is: %s", err)
            return [commMethods.CommonMethods.ERROR, "Import custom certificate failed."]

        # 判断证书有效期
        check_ret = check_cert_expired(custom_certificate_filepath)
        if not check_ret:
            if SecurityServiceErrorCodes.ERROR_CERTIFICATE_IS_INVALID.code == check_ret.error_code:
                run_log.error("The certificate is invalid.")
                return [SecurityServiceErrorCodes.ERROR_CERTIFICATE_IS_INVALID.code, "The certificate is invalid"]
            run_log.error("The certificate has expired.")
            return [SecurityServiceErrorCodes.ERROR_CERTIFICATE_HAS_EXPIRED.code, "the certificate has expired"]

        # 判断证书私钥是否匹配
        pkey_file = os.path.join(COPY_CUSTOM_CERT_PATH, "server_kmc.priv")
        psd_file = os.path.join(COPY_CUSTOM_CERT_PATH, "server_kmc.psd")
        if not SecurityService.check_cert_pkey_match(kmc, custom_certificate_filepath, pkey_file, psd_file):
            run_log.error("Import custom certificate failed: the certificate does not match the private key.")
            return [SecurityServiceErrorCodes.ERROR_CERTIFICATE_DOES_NOT_MATCH.code,
                    "the certificate does not match the private key"]

        # 判断证书合法性
        cert_check_result = certificate_verification(custom_certificate_filepath)
        if cert_check_result not in (0, ERR_SIGNATURE_LEN_NOT_ENOUGH,):
            run_log.error("Import custom certificate type error.")
            return [commMethods.CommonMethods.ERROR, "Import custom certificate type error."]

        # 复制证书到目录
        file_list = (
            "server_kmc.priv", "server_kmc.psd", "om_cert.keystore", "om_cert_backup.keystore", "om_alg.json"
        )
        for file_name in file_list:
            file_path = os.path.join(COPY_CUSTOM_CERT_PATH, file_name)
            new_file_path = os.path.join(CERT_PATH, file_name)
            FileCopy.copy_file(file_path, new_file_path, 0o600, "nobody", "nobody")
            FileUtils.delete_file_or_link(file_path)

        new_cert_path = os.path.join(CERT_PATH, "server_kmc.cert")
        FileCopy.copy_file(custom_certificate_filepath, new_cert_path, 0o600, "nobody", "nobody")
        run_log.info("Importing a custom certificate success")

        # 删除csr下载标志和csr文件
        is_download_csr = os.path.join(COPY_CUSTOM_CERT_PATH, "csr_flag")
        FileUtils.delete_file_or_link(is_download_csr)
        x509_csr = os.path.join(COPY_CUSTOM_CERT_PATH, "x509Req.pem")
        FileUtils.delete_file_or_link(x509_csr)

        if cert_check_result == ERR_SIGNATURE_LEN_NOT_ENOUGH:
            run_log.info("Legality of Certificate is risky")
            # 导入证书成功但存在合法性问题
            return [SecurityServiceErrorCodes.ERROR_IMPORT_CERTIFICATE_LEGALITY_RISKY.code,
                    "Importing a custom certificate success but legality is risky"]
        return [0, "Importing a custom certificate success"]

    @staticmethod
    def check_cert_pkey_match(kmc_inst: Kmc, cert_file: str, pkey_file: str, psd_file: str) -> bool:
        """
        检查证书和私钥是否匹配
        :param kmc_inst: kmc实例
        :param cert_file: 证书文件
        :param pkey_file: 私钥文件
        :param psd_file: 口令密文文件
        :return: Result
        """
        try:
            with open(psd_file, "r") as file:
                pwd = kmc_inst.decrypt(file.read())
        except Exception:
            run_log.error("decrypt pwd failed, caught exception")
            return False

        ret, err_msg = TlsConfig.get_ssl_context(None, cert_file, pkey_file, pwd)
        if not ret:
            run_log.error("check cert and pkey matching failed: %s", err_msg)
            return ret

        return True

    @staticmethod
    def check_upload_file_size(file_path):
        file_size = os.path.getsize(file_path)
        file_max_size = UploadConstants.FILE_MAX_SIZE.get("crt")
        if file_size > file_max_size:
            run_log.error("upload cert failed: over max size")
            return False
        return True

    @staticmethod
    def _set_cert_sum():
        cert_flag = os.path.join(CERT_BACKUP_PATH, "cert_flag")

        ret = FileCheck.check_input_path_valid(cert_flag)
        if not ret:
            run_log.error("cert_flag path is invalid.")
            return False

        try:
            with os.fdopen(os.open(cert_flag, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600), "w"):
                run_log.info("creat cert_flag success.")
        except Exception as err:
            run_log.error(f'{err}')

        return True

    def get_all_info(self, cert_type=None):
        if cert_type is None:
            self.items = ["HttpsCert"]

        self.get_https_cert_info()

    # 获取证书信息
    def get_https_cert_info(self):
        try:
            from lib.Linux.systems.security_service.security_service_clib import get_cert_info, \
                get_http_cert_effective, get_verify_cert_info
        except Exception as err:
            run_log.error(f'{err}')
            return

        try:
            https_cert_info_list = get_cert_info(SecurityService.DEFAULT_CERT_PATH)
        except Exception as err:
            run_log.error("Get httpsCert_info failed, reason: %s", err)
            return

        try:
            https_cert_info_list = [item.decode("utf-8") for item in https_cert_info_list]
        except Exception as err:
            run_log.info("HttpsCert_info_list" % err)
        self.Subject, self.Issuer, self.ValidNotBefore, self.ValidNotAfter, self.SerialNumber = https_cert_info_list

        try:
            _, _, _, _, fingerprint = get_verify_cert_info(SecurityService.DEFAULT_CERT_PATH)
        except Exception as err:
            run_log.error("Get fingerprint failed, %s", err)
            return

        if fingerprint is None:
            run_log.error("Get fingerprint for libcertmanage.so faild")
            return
        self.FingerPrint = fingerprint

        not_after_time = datetime.strptime(self.ValidNotAfter, "%b %d %H:%M:%S %Y %Z")
        not_before_time = datetime.strptime(self.ValidNotBefore, "%b %d %H:%M:%S %Y %Z")
        cur_time = datetime.utcnow()
        if cur_time > not_after_time:
            run_log.error("Certificate has expired")
            time_diff = cur_time - not_after_time
            self.ExpiredDayRemaining = time_diff.days
            self.HttpsCertEable = "Certificate has expired"
        elif not_before_time > cur_time:
            run_log.error("Certificate has not effected")
            self.HttpsCertEable = "Certificate has not effected"
            time_diff = not_before_time - cur_time
            self.ExpiredDayRemaining = time_diff.days
        else:
            time_diff = not_after_time - cur_time
            self.ExpiredDayRemaining = time_diff.days
            self.HttpsCertEable = "Certificate has not expired"

    # 导入自定义证书
    def post_request(self, request_data):
        if SecurityService.SECURITY_LOCK.locked():
            run_log.warning("SecurityService import is busy")
            return [commMethods.CommonMethods.ERROR, "Importing a custom certificate failed"]
        with SecurityService.SECURITY_LOCK:
            custom_certificate_file_name = request_data.get("FileName")

            ret = SecurityService.cheack_parameter(custom_certificate_file_name)
            if isinstance(ret, list) and ret[0] == commMethods.CommonMethods.ERROR:
                return ret

            custom_certificate_file_path = os.path.join(UploadConstants.CERT_UPLOAD_DIR, custom_certificate_file_name)
            ret = FileCheck.check_path_is_exist_and_valid(custom_certificate_file_path)
            if not ret:
                run_log.error("Importing a custom certificate failed. %s", ret.error)
                return [commMethods.CommonMethods.ERROR, "Importing a custom certificate failed"]

            if not SecurityService.check_upload_file_size(custom_certificate_file_path):
                run_log.error("Importing a custom certificate failed: over max size.")
                return [commMethods.CommonMethods.ERROR, "Importing a custom certificate failed"]

            return self.import_certificate(custom_certificate_file_path)
