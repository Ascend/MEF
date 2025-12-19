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
import threading
from datetime import datetime

from common.file_utils import FileCopy
from common.utils.result_base import Result

from common.constants.base_constants import CommonConstants
from cert_manager.cert_clib_mgr import CertClibMgr

# FD 显示时间格式
GREENWICH_MEAN_TIME = "%b %d %Y CMT"
# 时间转化格式
TIME_FORMAT = "%b %d %H:%M:%S %Y"

PARSE_CERT_LOCK = threading.Lock()
PARSE_CRL_LOCK = threading.Lock()


class ParseCertInfo:
    """解析根证书信息类"""

    def __init__(self, cert_buffer: str):
        if PARSE_CERT_LOCK.locked():
            raise ResourceWarning(f"Resource {__class__.__name__} is in progress.")

        with PARSE_CERT_LOCK:
            if not cert_buffer:
                raise ValueError("Cert buffer is null.")

            self.cert_buffer = cert_buffer
            self.cert_file_path = CommonConstants.REDFISH_CERT_TMP_FILE
            self.write_cert_tmp_file()
            self.cert_mgr = CertClibMgr(self.cert_file_path)
            self.cert_info = self.cert_mgr.get_cert_info()
            self.subject = self.cert_info[0].decode("utf-8")
            self.issuer = self.cert_info[1].decode("utf-8")
            self.serial_num = hex(int(self.cert_info[4].decode("utf-8")))[2:].upper()
            self.signature_algorithm = self.cert_info[5].decode("utf-8")
            self.signature_len = self.cert_info[6]
            self.cert_version = self.cert_info[7] + 1
            self.pubkey_type = self.cert_info[8]
            self.fingerprint = self.cert_info[9].decode("utf-8")[:-1]
            self.key_cert_sign = self.cert_info[10]
            self.is_ca = self.cert_info[11]
            self.chain_nums = self.cert_info[12]

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        FileCopy.remove_path(self.cert_file_path)

    @property
    def start_time(self):
        return datetime.strptime(" ".join(item for item in self.cert_info[2].decode("utf-8").split()[:-1]), TIME_FORMAT)

    @property
    def end_time(self):
        return datetime.strptime(" ".join(item for item in self.cert_info[3].decode("utf-8").split()[:-1]), TIME_FORMAT)

    @property
    def ca_sign_valid(self):
        return self.is_ca and self.cert_mgr.verify_ca_signature_valid()

    def write_cert_tmp_file(self):
        with os.fdopen(os.open(self.cert_file_path, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600), "w") as tmp_file:
            tmp_file.write(self.cert_buffer)

    def to_dict(self) -> dict:
        return {
            "SerialNum": self.serial_num,
            "Subject": self.subject,
            "Issuer": self.issuer,
            "Fingerprint": self.fingerprint,
            "Date": f"{self.start_time}--{self.end_time}",
        }

    def content_to_dict(self, cert_name: str, is_import_crl: bool) -> dict:
        return {
            "cert_type": "FDRootCert",
            "cert_name": cert_name,
            "issuer": self.issuer,
            "subject": self.subject,
            "valid_not_before": self.start_time.strftime(GREENWICH_MEAN_TIME),
            "valid_not_after": self.end_time.strftime(GREENWICH_MEAN_TIME),
            "serial_number": self.serial_num,
            "is_import_crl": is_import_crl,
            "signature_algorithm": self.signature_algorithm,
            "fingerprint": self.fingerprint,
            "key_usage": "Signing, CRL Sign",
            "public_key_length_bits2": f"{self.signature_len}",
        }


class ParseCrlInfo:
    """解析吊销列表信息类"""

    def __init__(self, crl_buffer: str):
        if PARSE_CRL_LOCK.locked():
            raise ResourceWarning(f"Resource {__class__.__name__} is in progress.")

        with PARSE_CRL_LOCK:
            if not crl_buffer:
                raise ValueError("Crl buffer is null.")

            self.crl_buffer = crl_buffer
            self.crl_file_path = CommonConstants.REDFISH_CRL_TMP_FILE
            self.cert_file_path = CommonConstants.REDFISH_CERT_TMP_FILE
            self.write_crl_tmp_file()
            self.cert_lib_manager = CertClibMgr(self.crl_file_path)
            self.crl_info = self.cert_lib_manager.get_crl_info()

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        FileCopy.remove_path(self.crl_file_path)
        FileCopy.remove_path(self.cert_file_path)

    @property
    def last_update(self):
        return datetime.strptime(" ".join(item for item in self.crl_info.last_update.split()[:-1]), TIME_FORMAT)

    @property
    def next_update(self):
        return datetime.strptime(" ".join(item for item in self.crl_info.next_update.split()[:-1]), TIME_FORMAT)

    def verify_crl_buffer_by_ca(self, cert_managers) -> Result:
        for cert_manager in cert_managers:
            if not cert_manager.cert_contents:
                continue
            if PARSE_CERT_LOCK.locked():
                return Result(result=False, err_msg="parse cert in progress.")

            with PARSE_CERT_LOCK:
                with os.fdopen(os.open(self.cert_file_path, os.O_WRONLY | os.O_CREAT | os.O_TRUNC,
                                       0o600), "w") as tmp_file:
                    tmp_file.write(cert_manager.cert_contents)

                ret = self.cert_lib_manager.verify_cert_available(self.cert_file_path)
                if ret != 0:
                    continue
                return Result(result=True, data=cert_manager.id)

        err_msg = "Verify CRL does not match against the CA certificate, please upload the CRL file again."
        return Result(result=False, err_msg=err_msg)

    def write_crl_tmp_file(self):
        with os.fdopen(os.open(self.crl_file_path, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600), "w") as tmp_file:
            tmp_file.write(self.crl_buffer)
