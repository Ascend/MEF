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
from datetime import datetime

from cert_manager.parse_tools import ParseCertInfo
from cert_manager.parse_tools import ParseCrlInfo
from common.constants.base_constants import CommonConstants
from common.constants.base_constants import PubkeyType
from common.constants.error_codes import SecurityServiceErrorCodes
from common.checkers import CheckResult
from common.checkers import StringLengthChecker


class CrlContentsChecker(StringLengthChecker):
    def __init__(self, attr_name=None, min_len: int = 1, max_len: int = CommonConstants.MAX_CRL_LIMIT):
        super().__init__(attr_name, min_len, max_len)

    def check_dict(self, data: dict) -> CheckResult:
        result = super().check_dict(data)
        if not result.success:
            return result

        crl_buffer = self.raw_value(data)
        if not crl_buffer:
            return CheckResult.make_success()

        time_now = datetime.utcnow()
        try:
            with ParseCrlInfo(crl_buffer) as parse_tool:
                if time_now <= parse_tool.last_update or time_now >= parse_tool.next_update:
                    msg_format = "Crl contents checker: check last update time or next update time is error."
                    return CheckResult.make_failed(msg_format)
        except Exception as err:
            msg_format = f"Crl contents checker: invalid crl of {err}"
            return CheckResult.make_failed(msg_format)

        return CheckResult.make_success()


class CertContentsChecker(StringLengthChecker):
    MAX_CHAIN_NUMS = 10
    X509_V3 = 3
    RSA_LEN_LIMIT = 3072
    # 椭圆曲线密钥长度
    EC_LEN_LIMIT = 256
    # 允许的签名算法
    SAFE_SIGNATURE_ALGORITHM = (
        "sha256WithRSAEncryption", "sha384WithRSAEncryption", "sha512WithRSAEncryption", "ecdsa-with-SHA256"
    )

    def __init__(self, attr_name=None, min_len: int = 1, max_len: int = CommonConstants.MAX_CERT_LIMIT):
        super().__init__(attr_name, min_len, max_len)

    def check_cert_info(self, cert_info: ParseCertInfo) -> CheckResult:
        if cert_info.is_ca == 0:
            msg_format = "Cert contents checker: cert is not CA."
            err_code = SecurityServiceErrorCodes.ERROR_CERTIFICATE_IS_NOT_CA.code
            return CheckResult.make_failed(msg_format, err_code=err_code)

        if cert_info.chain_nums > self.MAX_CHAIN_NUMS:
            msg_format = f"Cert contents checker: the number of CA cert chains is greater than {self.MAX_CHAIN_NUMS}."
            err_code = SecurityServiceErrorCodes.ERROR_CERTIFICATE_CHAIN_NUMS_MAX.code
            return CheckResult.make_failed(msg_format, err_code=err_code)

        if cert_info.key_cert_sign == 0:
            msg_format = "Cert contents checker: cannot used for verifying signatures on public key certificates."
            err_code = SecurityServiceErrorCodes.ERROR_CERTIFICATE_KEY_SIGN.code
            return CheckResult.make_failed(msg_format, err_code=err_code)

        time_now = datetime.utcnow()
        if time_now <= cert_info.start_time or time_now >= cert_info.end_time:
            msg_format = "Cert contents checker: invalid cert validity period."
            err_code = SecurityServiceErrorCodes.ERROR_CERTIFICATE_HAS_EXPIRED.code
            return CheckResult.make_failed(msg_format, err_code=err_code)

        if cert_info.cert_version != self.X509_V3:
            msg_format = f"Cert contents checkers: check cert version '{cert_info.cert_version}' is not safe."
            err_code = SecurityServiceErrorCodes.ERROR_CERTIFICATE_IS_NOT_X509V3.code
            return CheckResult.make_failed(msg_format, err_code=err_code)

        if cert_info.pubkey_type not in (PubkeyType.EVP_PKEY_RSA.value, PubkeyType.EVP_PKEY_EC.value):
            msg_format = "Cert contents checkers: check cert pubkey type is not safe."
            err_code = SecurityServiceErrorCodes.ERROR_CERTIFICATE_IS_NOT_RSA_EC.code
            return CheckResult.make_failed(msg_format, err_code=err_code)

        if cert_info.pubkey_type == PubkeyType.EVP_PKEY_RSA.value and cert_info.signature_len < self.RSA_LEN_LIMIT:
            msg_format = "Cert contents checkers: check cert pubkey length is not safe."
            err_code = SecurityServiceErrorCodes.ERROR_CERTIFICATE_RSA_LEN_INVALID.code
            return CheckResult.make_failed(msg_format, err_code=err_code)

        if cert_info.pubkey_type == PubkeyType.EVP_PKEY_EC.value and cert_info.signature_len < self.EC_LEN_LIMIT:
            msg_format = "Cert contents checkers: check cert pubkey length is not safe."
            err_code = SecurityServiceErrorCodes.ERROR_CERTIFICATE_EC_LEN_INVALID.code
            return CheckResult.make_failed(msg_format, err_code=err_code)

        if cert_info.signature_algorithm not in self.SAFE_SIGNATURE_ALGORITHM:
            msg_format = "Cert contents checkers: check signature algorithm is not safe."
            err_code = SecurityServiceErrorCodes.ERROR_CERTIFICATE_SIGN_ALG_INVALID.code
            return CheckResult.make_failed(msg_format, err_code=err_code)

        # 放在最后，该项校验失败不影响证书链整个链校验
        if not cert_info.ca_sign_valid:
            msg_format = "Cert contents checker: cert signature is invalid."
            err_code = SecurityServiceErrorCodes.ERROR_CERTIFICATE_CA_SIGNATURE_INVALID.code
            return CheckResult.make_failed(msg_format, err_code=err_code)

        return CheckResult.make_success()

    def check_dict(self, data: dict) -> CheckResult:
        result = super().check_dict(data)
        if not result.success:
            return result

        cert_buffer = self.raw_value(data)
        if not cert_buffer:
            return CheckResult.make_success()

        try:
            with ParseCertInfo(cert_buffer) as cert_info:
                return self.check_cert_info(cert_info)
        except Exception as err:
            msg_format = f"Cert contents checkers: invalid cert of {err}."
            return CheckResult.make_failed(msg_format)
