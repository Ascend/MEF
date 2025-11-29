# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2022-2023. All rights reserved.
from datetime import datetime

from cert_manager.parse_tools import CertChainParser
from cert_manager.schemas import CertInfoSchema
from common.checkers import CheckResult
from common.checkers import StringLengthChecker
from common.constants.base_constants import CommonConstants
from common.constants.base_constants import PubkeyType
from common.constants.error_codes import SecurityServiceErrorCodes
from net_manager.exception import InvalidCertInfo


class CertInfoValidator:
    X509_V3 = 3
    RSA_LEN_LIMIT = 3072
    # 椭圆曲线密钥长度
    EC_LEN_LIMIT = 256
    # 允许的签名算法
    SAFE_SIGNATURE_ALGORITHM = (
        "sha256WithRSAEncryption", "sha384WithRSAEncryption", "sha512WithRSAEncryption", "ecdsa-with-SHA256"
    )

    def check_cert_info(self, cert_info: CertInfoSchema):
        if cert_info.is_ca == 0:
            raise InvalidCertInfo(SecurityServiceErrorCodes.ERROR_CERTIFICATE_IS_NOT_CA)
        if cert_info.key_cert_sign == 0:
            raise InvalidCertInfo(SecurityServiceErrorCodes.ERROR_CERTIFICATE_KEY_SIGN)
        time_now = datetime.utcnow()
        if time_now <= cert_info.start_time or time_now >= cert_info.end_time:
            raise InvalidCertInfo(SecurityServiceErrorCodes.ERROR_CERTIFICATE_HAS_EXPIRED)
        if cert_info.cert_version != self.X509_V3:
            raise InvalidCertInfo(SecurityServiceErrorCodes.ERROR_CERTIFICATE_IS_NOT_X509V3)
        if cert_info.pubkey_type not in (PubkeyType.EVP_PKEY_RSA.value, PubkeyType.EVP_PKEY_EC.value):
            raise InvalidCertInfo(SecurityServiceErrorCodes.ERROR_CERTIFICATE_IS_NOT_RSA_EC)
        if cert_info.pubkey_type == PubkeyType.EVP_PKEY_RSA.value and cert_info.signature_len < self.RSA_LEN_LIMIT:
            raise InvalidCertInfo(SecurityServiceErrorCodes.ERROR_CERTIFICATE_RSA_LEN_INVALID)
        if cert_info.pubkey_type == PubkeyType.EVP_PKEY_EC.value and cert_info.signature_len < self.EC_LEN_LIMIT:
            raise InvalidCertInfo(SecurityServiceErrorCodes.ERROR_CERTIFICATE_EC_LEN_INVALID)
        if cert_info.signature_algorithm not in self.SAFE_SIGNATURE_ALGORITHM:
            raise InvalidCertInfo(SecurityServiceErrorCodes.ERROR_CERTIFICATE_SIGN_ALG_INVALID)


class CertContentsChecker(StringLengthChecker, CertInfoValidator):
    def __init__(self, attr_name=None, min_len: int = 1, max_len: int = CommonConstants.MAX_CERT_LIMIT):
        super().__init__(attr_name, min_len, max_len)

    def check_dict(self, data: dict) -> CheckResult:
        result = super().check_dict(data)
        if not result.success:
            return result

        cert_buffer = self.raw_value(data)
        if not cert_buffer:
            return CheckResult.make_success()

        try:
            cert_chain_parser = CertChainParser(cert_buffer)
            for cert_schema in cert_chain_parser.cert_schema_generator():
                self.check_cert_info(cert_schema)
            cert_chain_parser.verify_cert_chain()
            return CheckResult.make_success()
        except Exception as err:
            msg_format = f"Cert contents checkers: invalid cert of {err}."
            return CheckResult.make_failed(msg_format)
