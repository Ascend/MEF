# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import os
from ctypes import CDLL, POINTER, pointer, cast, c_char, c_ulong, c_int, c_char_p, create_string_buffer
from datetime import datetime

from cert_manager.schemas import CertInfoSchema, CrlInfoSchema
from common.constants.base_constants import CommonConstants
from common.file_utils import FileCheck

# 时间转化格式
TM_FMT = "%b %d %H:%M:%S %Y"


class CertClibMgr:
    CLIB_PATH = "software/RedfishServer/lib/c/libcertmanage.so"

    def __init__(self, file_path: str, work_dir: str = CommonConstants.OM_WORK_DIR_PATH):
        self.clib_path = os.path.join(work_dir, self.CLIB_PATH)
        self.file_path = file_path
        self.file_path_check()
        self.clib_cert = CDLL(self.clib_path)

    def file_path_check(self):
        if not os.path.exists(self.clib_path):
            raise FileNotFoundError(f"Load clib filed, check [{self.clib_path}] does not exist.")

        if not FileCheck.check_path_is_exist_and_valid(self.file_path):
            raise ValueError(f"Load clib filed, check path [{self.file_path}] is invalid.")

    def get_cert_info(self) -> CertInfoSchema:
        cert_path = self.file_path.encode(encoding="utf-8")
        cert_file = cast(cert_path, POINTER(c_char * 64))

        clib_get_cert_info = self.clib_cert.cert_getcertinfo
        clib_get_cert_info.argtypes = [
            POINTER(c_char * 64), POINTER(c_char * 256), c_ulong, POINTER(c_char * 256), c_ulong,
            POINTER(c_char * 64), c_int, POINTER(c_char * 64), c_int, POINTER(c_char * 50), c_ulong,
        ]
        clib_get_cert_info.restype = c_int
        subj, subj_length = (c_char * 256)(), c_ulong(256)
        issuer, issuer_len = (c_char * 256)(), c_ulong(256)
        not_begin, not_begin_len = (c_char * 64)(), c_int(64)
        not_after, not_after_len = (c_char * 64)(), c_int(64)
        serial_num, serial_num_len = (c_char * 50)(), c_ulong(50)
        ret_value = clib_get_cert_info(
            cert_file, pointer(subj), subj_length, pointer(issuer), issuer_len, pointer(not_begin), not_begin_len,
            pointer(not_after), not_after_len, pointer(serial_num), serial_num_len,
        )
        if ret_value != 0:
            raise ValueError(f"Load clib filed, clib function [cert_getcertinfo] return value is [{ret_value}].")

        clib_get_extend_cert_info = self.clib_cert.get_extend_certinfo
        clib_get_extend_cert_info.argtypes = [
            c_char_p, POINTER(c_char * 64), c_ulong, POINTER(c_char * 128), c_ulong,
            POINTER(c_ulong), POINTER(c_ulong), POINTER(c_ulong), POINTER(c_ulong), POINTER(c_ulong), POINTER(c_ulong)
        ]
        clib_get_extend_cert_info.restype = c_int
        signature_algorithm, signature_algorithm_len = (c_char * 64)(), c_ulong(64)
        fingerprint, fingerprint_len = (c_char * 128)(), c_ulong(128)
        pubkey_len, pubkey_type, cert_version, key_usage, is_ca, chain_nums = (c_ulong() for _ in range(6))
        ret_value = clib_get_extend_cert_info(
            cert_path, pointer(signature_algorithm), signature_algorithm_len, fingerprint, fingerprint_len,
            pointer(pubkey_len), pointer(pubkey_type), pointer(cert_version), pointer(key_usage), pointer(is_ca),
            pointer(chain_nums)
        )
        if ret_value != 0:
            raise ValueError(f"Load clib filed, clib function [get_extend_certinfo] return value is [{ret_value}].")

        return CertInfoSchema(
            subject=subj.value.decode(), issuer=issuer.value.decode(),
            serial_num=hex(int(serial_num.value.decode()))[2:].upper(),
            signature_algorithm=signature_algorithm.value.decode(), signature_len=pubkey_len.value,
            cert_version=cert_version.value + 1, pubkey_type=pubkey_type.value,
            fingerprint=fingerprint.value.decode(), key_cert_sign=key_usage.value, chain_num=chain_nums.value,
            is_ca=is_ca.value, ca_sign_valid=is_ca.value and self.verify_ca_signature_valid(),
            start_time=datetime.strptime(" ".join(not_begin.value.decode().split()[:-1]), TM_FMT),
            end_time=datetime.strptime(" ".join(not_after.value.decode().split()[:-1]), TM_FMT),
        )

    def get_crl_info(self) -> CrlInfoSchema:
        crl_path = self.file_path.encode(encoding="utf-8")
        clib_get_crl_info = self.clib_cert.get_crl_info
        clib_get_crl_info.argtypes = [
            c_char_p, POINTER(c_ulong), POINTER(c_char * 1024), POINTER(c_char * 1024), POINTER(c_int * 64), c_ulong,
            POINTER(c_char * 64), c_int, POINTER(c_char * 64), c_int,
        ]
        clib_get_crl_info.restype = c_int
        issuer_len = c_ulong()
        issuer_obj_name = (c_char * 1024)()
        issuer_value_name = (c_char * 1024)()
        issuer_value_length = (c_int * 64)()
        row_len = c_ulong(1024)
        last_update = (c_char * 64)()
        last_update_len = c_int(64)
        next_update = (c_char * 64)()
        next_update_len = c_int(64)

        ret_value = clib_get_crl_info(
            crl_path, pointer(issuer_len), pointer(issuer_obj_name), pointer(issuer_value_name),
            pointer(issuer_value_length), row_len, pointer(last_update), last_update_len, pointer(next_update),
            next_update_len,
        )
        if ret_value != 0:
            raise ValueError(f"Load clib filed, clib function [get_crl_info] return value is [{ret_value}].")

        return CrlInfoSchema(
            last_update=datetime.strptime(" ".join(last_update.value.decode().split()[:-1]), TM_FMT),
            next_update=datetime.strptime(" ".join(next_update.value.decode().split()[:-1]), TM_FMT),
        )

    def verify_cert_available(self, cert_path) -> int:
        crl_path = self.file_path.encode(encoding="utf-8")
        cert_file = cert_path.encode(encoding="utf-8")
        verify_cert = self.clib_cert.verify_cert_by_crl
        verify_cert.argtypes = [c_char_p, c_char_p]
        verify_cert.restype = c_int
        return verify_cert(cert_file, crl_path)

    def verify_ca_signature_valid(self) -> bool:
        """
        校验根CA证书的签名是否正确
        :return: True-是根CA证书，且签名信息正确；False-不是根CA证书或者签名信息不正确
        """
        verify_ca_cert_sign = self.clib_cert.verify_ca_cert_sign
        verify_ca_cert_sign.args = [c_char_p]
        verify_ca_cert_sign.restype = c_int
        ca_path_p = create_string_buffer(self.file_path.encode(encoding="utf-8"))
        return verify_ca_cert_sign(ca_path_p) == 0

    def cert_chain_verify(self) -> int:
        chain_verify_func = self.clib_cert.CertChainVerify
        chain_verify_func.args = [c_char_p]
        chain_verify_func.restype = c_int
        return chain_verify_func(create_string_buffer(self.file_path.encode(encoding="utf-8")))
