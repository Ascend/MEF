#!/usr/bin/python
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

import ctypes
from ctypes import c_char
from ctypes import c_int
from ctypes import pointer
from datetime import datetime

from common.constants.error_codes import SecurityServiceErrorCodes
from common.file_utils import FileCheck
from common.log.logger import run_log
from common.utils.result_base import Result

SUC_CERT_IMPORT_SUCCESS = '0x0'  # 产生CSR文件失败
ERR_CERT_GENERATECSR_FAILED = '0x20140001'  # 产生CSR文件失败
ERR_CERT_LOADCERT_FAILED = '0x20140002'  # 加载证书文件失败
ERR_CERT_NOT_MATCHING = '0x20140003'  # 证书和私钥不匹配
ERR_CERT_FILELEN_ERROR = '0x20140004'  # 证书文件过大
ERR_CERT_PFX2PEM_FAILED = '0x20140005'  # PFX格式证书转换成PEM格式失败
ERR_CERT_NO_PRIVKEY_EXIST = '0x20140006'  # 私钥证书不存在
ERR_CERT_DECRTYPEDPRIVK_FAILED = '0x20140007'  # 私钥证书解密失败
ERR_CERT_DECRYPT_FAILED = '0x20140008'  # 证书解密失败，用于解密自定义证书失败场景
ERR_CERT_VALIDITY_WARN = '0x20140009'  # 证书有效期告警
ERR_CERT_EARLY_USE_WARN = '0x20140010'  # 证书过早使用
ERR_CERT_PASSWD_TOO_LONG = '0x20140011'  # 证书密码过长

CERT_CHECK_SUCCESS = 0
ERR_GET_CERT_INFO = -1
ERR_ALGORITHM = 1
ERR_SIGNATURE_LEN = 2
ERR_ALGORITHM_NOT_SAFE = 3
ERR_SIGNATURE_LEN_NOT_ENOUGH = 4
ERR_SIGNATURE_VERSION_NOT_SAFE = 5
ERR_CERT_TYPE_CA = 6

X509_V3 = 3
RSA_LEN_LIMIT = 4096
ECDSA_LEN_LIMIT = 256

CERT_MANAGE_LIB_NAME = "libcertmanage.so"
libgetcert = ctypes.CDLL(CERT_MANAGE_LIB_NAME)
getcertinfo = libgetcert.cert_getcertinfo
getcertinfo.argtypes = [
    ctypes.POINTER(ctypes.c_char * 64),
    ctypes.POINTER(ctypes.c_char * 256), ctypes.c_ulong,
    ctypes.POINTER(ctypes.c_char * 256), ctypes.c_ulong,
    ctypes.POINTER(ctypes.c_char * 64), ctypes.c_int,
    ctypes.POINTER(ctypes.c_char * 64), ctypes.c_int,
    ctypes.POINTER(ctypes.c_char * 50), ctypes.c_ulong
]
getcertinfo.restype = ctypes.c_int

get_extend_cert_info = libgetcert.get_extend_certinfo
get_extend_cert_info.argtypes = [
    ctypes.c_char_p,
    ctypes.POINTER(ctypes.c_char * 64), ctypes.c_ulong,
    ctypes.POINTER(ctypes.c_char * 128), ctypes.c_ulong,
    ctypes.POINTER(ctypes.c_ulong), ctypes.POINTER(ctypes.c_ulong),
    ctypes.POINTER(ctypes.c_ulong), ctypes.POINTER(ctypes.c_ulong),
    ctypes.POINTER(ctypes.c_ulong), ctypes.POINTER(ctypes.c_ulong)
]
get_extend_cert_info.restype = ctypes.c_int

# 证书是否过期
get_cert_effective = libgetcert.certPeriodicalCheck
get_cert_effective.argtypes = [ctypes.POINTER(ctypes.c_char * 64), ctypes.POINTER(ctypes.c_int)]
get_cert_effective.restype = ctypes.c_int

# 自定义证书导入初始化
cert_manage_init = libgetcert.cert_manage_init
cert_manage_init.argtypes = (ctypes.c_char_p, ctypes.c_char_p, ctypes.c_char_p)
cert_manage_init.restype = c_int

# 自定义证书导入去初始化
cert_manage_finalize = libgetcert.cert_manage_finalize
cert_manage_finalize.restype = c_int

# create cert by path
create_server_cert = libgetcert.create_server_cert
create_server_cert.argtypes = (ctypes.c_char_p, ctypes.c_ulong, ctypes.c_char_p, ctypes.c_char_p, ctypes.c_char_p)
create_server_cert.restype = ctypes.c_int

# create cert sign request
create_cert_sign_request = libgetcert.create_cert_sign_request
create_cert_sign_request.argtypes = (ctypes.c_char_p, ctypes.c_char_p)
create_cert_sign_request.restype = ctypes.c_int

CERT_PATH = "/home/data/config/default"
CERT_PRIMARY_KSF = "/home/data/config/default/om_cert.keystore"
CERT_STANDBY_KSF = "/home/data/config/default/om_cert_backup.keystore"
ALG_CONFIG_FILE = "/home/data/config/default/om_alg.json"


def check_lib_available(lib_name: str) -> bool:
    """
    检查lib_name的功能是否可用
    CDLL会尝试使用dlopen函数打开动态库，找不到时抛OSError异常
    :param lib_name: 动态库名字
    :return: True-可用；False-不可用
    """
    try:
        ctypes.CDLL(lib_name)
    except OSError as err:
        run_log.error("lib name %s not available: %s", lib_name, err)
        return False
    return True


# 证书信息
def get_cert_info(cert_file_path):
    """
    功能描述：证书信息
    参数：无
    返回值：无
    异常描述：NA
    修改记录：2019-3-20 创建
    """
    if not check_lib_available(CERT_MANAGE_LIB_NAME):
        raise IOError(f"Check so file {CERT_MANAGE_LIB_NAME} not available.")

    ret = FileCheck.check_path_is_exist_and_valid(cert_file_path)
    if not ret:
        raise IOError(f"Check cert path failed! because of {ret.error}")

    cert_path = cert_file_path.encode(encoding='utf-8')
    certinfo_file = ctypes.cast(cert_path, ctypes.POINTER(ctypes.c_char * 64))
    subj = (c_char * 256)()
    subj_length = ctypes.c_ulong(256)
    issuer = (c_char * 256)()
    issuer_len = ctypes.c_ulong(256)
    not_begin = (c_char * 64)()
    not_begin_len = ctypes.c_int(64)
    not_after = (c_char * 64)()
    not_after_len = ctypes.c_int(64)
    serial_num = (c_char * 50)()
    serial_num_len = ctypes.c_ulong(50)

    ret = getcertinfo(
        certinfo_file,
        pointer(subj), subj_length,
        pointer(issuer), issuer_len,
        pointer(not_begin), not_begin_len,
        pointer(not_after), not_after_len,
        pointer(serial_num), serial_num_len,
    )
    if ret < 0:
        raise ValueError("Get https cert info failed")

    return [subj.value, issuer.value, not_begin.value, not_after.value, serial_num.value]


def get_verify_cert_info(cert_file_path):
    """
    功能描述：证书信息
    参数：无
    返回值：无
    异常描述：NA
    修改记录：2019-3-20 创建
    """
    if not check_lib_available(CERT_MANAGE_LIB_NAME):
        raise IOError(f"so file {CERT_MANAGE_LIB_NAME} not available.")

    ret = FileCheck.check_path_is_exist_and_valid(cert_file_path)
    if not ret:
        raise IOError(f"Check cert path failed! because of {ret.error}")

    cert_path = cert_file_path.encode(encoding='utf-8')
    signature_algorithm = (c_char * 64)()
    signature_algorithm_len = ctypes.c_ulong(64)
    fingerprint = (c_char * 128)()
    fingerprint_len = ctypes.c_ulong(128)
    pubkey_len = ctypes.c_ulong()
    pubkey_type = ctypes.c_ulong()
    cert_version = ctypes.c_ulong()
    key_usage = ctypes.c_ulong()
    is_ca = ctypes.c_ulong()
    chain_nums = ctypes.c_ulong()

    ret = get_extend_cert_info(cert_path, pointer(signature_algorithm), signature_algorithm_len, fingerprint,
                               fingerprint_len, pointer(pubkey_len), pointer(pubkey_type), pointer(cert_version),
                               pointer(key_usage), pointer(is_ca), pointer(chain_nums))
    if ret < 0:
        raise ValueError("Get https cert info for libcertmanage.so failed")

    return [signature_algorithm.value.decode("utf-8"), pubkey_len.value, cert_version.value, is_ca.value,
            fingerprint.value.decode("utf-8")]


# 证书是否过期
def get_http_cert_effective(default_cert_path):
    ret = FileCheck.check_path_is_exist_and_valid(default_cert_path)
    if not ret:
        raise IOError(f"check default cert path failed! because {ret.error}")

    cert_path = default_cert_path.encode(encoding='utf-8')
    certinfo_file = ctypes.cast(cert_path, ctypes.POINTER(ctypes.c_char * 64))
    day = c_int()

    ret = get_cert_effective(certinfo_file, pointer(day))
    if ret == -1:
        raise ValueError("Get_cert_effective function error")
    elif ret == 0x20140009:
        run_log.error("Certificate will expire")
        return [0, -day.value]

    return [0, day.value]


def certificate_verification(cert_path):
    """
    功能描述：校验web证书是否符合安全要求
    参数：无
    返回值：0,1,2,3,4,5,6,-1
    异常描述：NA
    """
    try:
        cert_signature, cert_signature_len, cert_version, is_ca, _ = get_verify_cert_info(cert_path)
    except Exception as err:
        run_log.error("load certificate error, because {}.".format(err))
        return ERR_GET_CERT_INFO

    cert_version = cert_version + 1

    if not cert_signature:
        run_log.error("get signature algorithm failed.")
        return ERR_ALGORITHM

    if not cert_signature_len:
        run_log.error("get signature len failed.")
        return ERR_SIGNATURE_LEN

    if cert_signature != "sha256WithRSAEncryption" and cert_signature != "ecdsa-with-SHA256":
        run_log.error("signature algorithm not safe.")
        return ERR_ALGORITHM_NOT_SAFE

    if cert_version != X509_V3:
        run_log.error("x509 version is not safe.")
        return ERR_SIGNATURE_VERSION_NOT_SAFE

    if is_ca:
        run_log.error("web cert can not be ca.")
        return ERR_CERT_TYPE_CA

    # 允许的导入的风险提示场景的校验需要放在最后
    if cert_signature == "sha256WithRSAEncryption" and cert_signature_len < RSA_LEN_LIMIT:
        run_log.error("signature sha256WithRSAEncryption length is not safe.")
        return ERR_SIGNATURE_LEN_NOT_ENOUGH

    if cert_signature == "ecdsa-with-SHA256" and cert_signature_len < ECDSA_LEN_LIMIT:
        run_log.error("signature ecdsa-with-SHA256 length is not safe.")
        return ERR_SIGNATURE_LEN_NOT_ENOUGH

    run_log.info("signature algorithm and signature length and signature version is ok.")
    return CERT_CHECK_SUCCESS


def check_cert_expired(cert_path: str) -> Result:
    """检查证书是否未过期"""
    try:
        https_cert_info_list = get_cert_info(cert_path)
    except Exception as err:
        run_log.error("Get cert info failed, reason: %s", err)
        return Result(result=False, err_code=SecurityServiceErrorCodes.ERROR_CERTIFICATE_IS_INVALID.code)

    try:
        https_cert_info_list = [item.decode("utf-8") for item in https_cert_info_list]
    except Exception as err:
        run_log.error("HttpsCert_info_list: %s", err)
    _, _, not_before, not_after, _ = https_cert_info_list

    run_log.info("not_before: %s, not_after: %s", not_before, not_after)
    time_format = "%b %d %H:%M:%S %Y %Z"
    not_before_time = datetime.strptime(not_before, time_format)
    not_after_time = datetime.strptime(not_after, time_format)

    cur_time = datetime.utcnow()
    run_log.info("cur_time: %s", cur_time)
    return Result(not_before_time <= cur_time <= not_after_time)
