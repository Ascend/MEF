# coding: utf-8
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
"""
功 能：KMC加解密接口
版权信息：华为技术有限公司，版本所有(C) 2020-2029
"""

import ctypes
import json
import os
import sys
import threading
import time

from collections import namedtuple
from datetime import datetime
from enum import Enum

RET_SUCCESS = 0
DEFAULT_ENCODING = "UTF-8"
KMC_LIB_NAME = "libkmcext.so"
KMC_DEP_LIBS = ["libkmcext.so", "libkmc.so", "libsdp.so", "libsecurec.so", "libcrypto.so"]


class KmcEnum(Enum):

    @classmethod
    def value_list(cls):
        return [x.value for x in cls]


class Role(KmcEnum):
    agent = 0
    master = 1


class CryptoAlgorithm(KmcEnum):
    AES128_GCM = 8
    AES256_GCM = 9


class SignAlgorithm(KmcEnum):
    HMAC_SHA384 = 2053
    HMAC_SHA512 = 2054


class KeyType(KmcEnum):
    ROOT_KEY = "root key"
    MASTER_KEY = "master key"


class KmcConfig(ctypes.Structure):
    _fields_ = [
        ("primaryKeyStoreFile", ctypes.c_char * 4096),
        ("standbyKeyStoreFile", ctypes.c_char * 4096),
        ("domainCount", ctypes.c_int),
        ("role", ctypes.c_int),
        ("procLockPerm", ctypes.c_int),
        ("sdpAlgId", ctypes.c_int),
        ("hmacAlgId", ctypes.c_int),
        ("semKey", ctypes.c_int),
        ("innerSymmAlgId", ctypes.c_int),
        ("innerHashAlgId", ctypes.c_int),
        ("innerHmacAlgId", ctypes.c_int),
        ("innerKdfAlgId", ctypes.c_int),
        ("workKeyIter", ctypes.c_int),
        ("rootKeyIter", ctypes.c_int),
        ("version", ctypes.c_ushort),
    ]


class KmcWsecSysTime(ctypes.Structure):
    _fields_ = [
        ("kmcYear", ctypes.c_ushort),
        ("kmcMonth", ctypes.c_ubyte),
        ("kmcDate", ctypes.c_ubyte),
        ("kmcHour", ctypes.c_ubyte),
        ("kmcMinute", ctypes.c_ubyte),
        ("kmcSecond", ctypes.c_ubyte),
        ("kmcWeek", ctypes.c_ubyte)
    ]


class KeyAdaptorError(Exception):
    """KeyAdaptorError"""

    def __init__(self, error_msg="", error_code=None):
        super().__init__()
        self.error_msg = error_msg
        self.error_code = error_code


class KmcError(KeyAdaptorError):
    def __str__(self):
        if self.error_code:
            return f"KmcError[{self.error_code}] {self.error_msg}"
        return self.error_msg


def str_to_bytes(s):
    if isinstance(s, bytes):
        return s
    if isinstance(s, str):
        return s.encode(sys.getfilesystemencoding())
    raise KeyAdaptorError("Value must be represented as bytes or unicode string")


def free_char_buffer(char_p):
    char_p.value = b"\x00" * len(char_p)


class KmcWrapper:
    _DEFAULT_DOMAIN_ID = 0
    _DEFAULT_DOMAIN_COUNT = 2
    _DEFAULT_ROLE = Role.master.value
    _DEFAULT_PROC_LOCK_PERM = 0o0600
    _DEFAULT_RETRY_TIMES = 3

    def __init__(self, sdp_alg_id=CryptoAlgorithm.AES256_GCM.value,
                 hmac_alg_id=SignAlgorithm.HMAC_SHA512.value):
        self._has_inited = False
        self._kmc_config = KmcConfig()
        self.sdp_alg_id = sdp_alg_id
        self.hmac_alg_id = hmac_alg_id
        self._load_so()

    @staticmethod
    def _convert_wsec_time(wsec_time):
        return datetime(
            year=wsec_time.kmcYear,
            month=wsec_time.kmcMonth,
            day=wsec_time.kmcDate,
            hour=wsec_time.kmcHour,
            minute=wsec_time.kmcMinute,
            second=wsec_time.kmcSecond
        )

    @staticmethod
    def _get_interval_time(one_time, other_time):
        return (one_time - other_time).total_seconds()

    def finalize(self):
        if self._has_inited:
            self._finalize()

    def initialize(self, primary_key_store_file, standby_key_store_file):
        """
        Initialize kmc component
        Args:
            primary_key_store_file: <bytes> Primary KSF file path
            standby_key_store_file: <bytes> Standby KSF file path

        Returns:
            None
        """
        self._kmc_config.primaryKeyStoreFile = primary_key_store_file
        self._kmc_config.standbyKeyStoreFile = standby_key_store_file
        self._kmc_config.domainCount = self._DEFAULT_DOMAIN_COUNT
        self._kmc_config.role = self._DEFAULT_ROLE
        self._kmc_config.procLockPerm = self._DEFAULT_PROC_LOCK_PERM
        self._kmc_config.sdpAlgId = self.sdp_alg_id
        self._kmc_config.hmacAlgId = self.hmac_alg_id
        self._kmc_config.semKey = 0
        ret = RET_SUCCESS
        for _ in range(self._DEFAULT_RETRY_TIMES):
            ret = self._kmc_ext_dll.SeInitialize(ctypes.byref(self._kmc_config))
            if ret == RET_SUCCESS:
                break
            time.sleep(1)

        if ret != RET_SUCCESS:
            raise KmcError("Initialize error", ret)
        self._has_inited = True

    def encrypt_mem_base64(self, plain_text_str, cipher_alg_id):
        """
        Encrypt Memory Data With base64 encoding
        Args:
            plain_text_str: <str> Data for encryption
            cipher_alg_id:  <int> AES Algorithm ID

        Returns:
            Encrypted data with base64 encoding
        """
        plain_text = str_to_bytes(plain_text_str)
        plain_text_len = len(plain_text)
        cipher_text_len = self._get_cipher_data_len(plain_text)
        cipher_text_len += cipher_text_len
        cipher_text_len_p = ctypes.c_uint32(cipher_text_len)
        plain_text_p = ctypes.create_string_buffer(plain_text)
        cipher_text_p = ctypes.create_string_buffer(cipher_text_len)
        self._kmc_ext_dll.SdpMemEncryptBase64.argtypes = [
            ctypes.c_uint32,
            ctypes.c_uint32,
            ctypes.c_char_p,
            ctypes.c_uint32,
            ctypes.c_char_p,
            ctypes.POINTER(ctypes.c_uint32)
        ]
        ret = self._kmc_ext_dll.SdpMemEncryptBase64(
            ctypes.c_uint32(self._DEFAULT_DOMAIN_ID),
            ctypes.c_uint32(cipher_alg_id),
            plain_text_p,
            ctypes.c_uint32(plain_text_len),
            cipher_text_p,
            ctypes.byref(cipher_text_len_p)
        )
        if ret != RET_SUCCESS:
            free_char_buffer(plain_text_p)
            raise KmcError("Encrypt error", ret)
        cipher_text_len = cipher_text_len_p.value
        cipher_text_base64 = str(cipher_text_p.raw[:cipher_text_len],
                                 encoding=DEFAULT_ENCODING)
        free_char_buffer(plain_text_p)
        return cipher_text_base64

    def decrypt_mem_base64(self, cipher_text_str):
        """
        Decrypt Memory Data With base64 encoding
        Args:
            cipher_text_str: <str> cipher text with base64 encoding

        Returns:
            Decrypted data

        """
        cipher_text = str_to_bytes(cipher_text_str)
        cipher_text_len = len(cipher_text)
        cipher_text_p = ctypes.create_string_buffer(cipher_text)
        plain_text_p = ctypes.create_string_buffer(cipher_text_len)
        plain_text_len_p = ctypes.c_uint32(cipher_text_len)
        self._kmc_ext_dll.SdpMemDecryptBase64.argtypes = [
            ctypes.c_uint32,
            ctypes.c_char_p,
            ctypes.c_uint32,
            ctypes.c_char_p,
            ctypes.POINTER(ctypes.c_uint32)
        ]
        ret = self._kmc_ext_dll.SdpMemDecryptBase64(
            ctypes.c_uint32(self._DEFAULT_DOMAIN_ID),
            cipher_text_p,
            ctypes.c_uint32(cipher_text_len),
            plain_text_p,
            ctypes.byref(plain_text_len_p)
        )
        if ret != RET_SUCCESS:
            free_char_buffer(plain_text_p)
            raise KmcError("Decrypt error", ret)
        plain_text_len = plain_text_len_p.value
        plain_text = str(
            plain_text_p.raw[:plain_text_len], encoding=DEFAULT_ENCODING)
        free_char_buffer(plain_text_p)
        return plain_text

    def update_root_key(self, advance_update_rk_days):
        self._kmc_ext_dll.SeAutoUpdateRk.argtypes = [
            ctypes.c_int
        ]
        ret = self._kmc_ext_dll.SeAutoUpdateRk(
            ctypes.c_int(advance_update_rk_days)
        )
        if ret != RET_SUCCESS:
            raise KmcError("Update root key error", ret)

    def instant_update_root_key(self):
        ret = self._kmc_ext_dll.SeUpdateRootKey()
        if ret != RET_SUCCESS:
            raise KmcError("Update root key error", ret)

    def update_master_key(self, advance_update_mk_days):
        self._kmc_ext_dll.SeCheckAndUpdateMk.argtypes = [
            ctypes.c_uint32,
            ctypes.c_int
        ]
        ret = self._kmc_ext_dll.SeCheckAndUpdateMk(
            ctypes.c_uint32(self._DEFAULT_DOMAIN_ID),
            ctypes.c_int(advance_update_mk_days)
        )
        if ret != RET_SUCCESS:
            raise KmcError("Update master key error", ret)

    def instant_update_master_key(self):
        self._kmc_ext_dll.SeUpdateMasterKey.argtypes = [
            ctypes.c_uint32
        ]
        ret = self._kmc_ext_dll.SeUpdateMasterKey(
            ctypes.c_uint32(self._DEFAULT_DOMAIN_ID)
        )
        if ret != RET_SUCCESS:
            raise KmcError("Update master key error", ret)

    def get_master_key_count(self):
        self._kmc_ext_dll.SeGetMkCountByDomain.argtypes = [
            ctypes.c_uint32
        ]
        mk_counts = self._kmc_ext_dll.SeGetMkCountByDomain(
            ctypes.c_uint32(self._DEFAULT_DOMAIN_ID)
        )
        return mk_counts

    def check_and_remove_oldest_master_key(self):
        self._kmc_ext_dll.SeCheckAndRmvOldestMk.argtypes = [
            ctypes.c_uint32
        ]
        ret = self._kmc_ext_dll.SeCheckAndRmvOldestMk(
            ctypes.c_uint32(self._DEFAULT_DOMAIN_ID)
        )
        if ret != RET_SUCCESS:
            raise KmcError("Remove oldest master key error", ret)

    def get_utc_date_time(self):
        kmc_utc_date_time = KmcWsecSysTime()
        ret = self._kmc_ext_dll.SeGetUtcDateTime(
            ctypes.byref(kmc_utc_date_time)
        )
        if ret != RET_SUCCESS:
            raise KmcError("Get the utc date time error", ret)

        return self._convert_wsec_time(kmc_utc_date_time)

    def get_latest_root_key_create_time(self):
        kmc_rk_create_time_latest = KmcWsecSysTime()
        ret = self._kmc_ext_dll.SeGetRkCreateTimeLatest(
            ctypes.byref(kmc_rk_create_time_latest)
        )
        if ret != RET_SUCCESS:
            raise KmcError("Get the latest create time of the root key error", ret)

        return self._convert_wsec_time(kmc_rk_create_time_latest)

    def get_root_key_expire_time(self):
        kmc_rk_expire_time = KmcWsecSysTime()
        ret = self._kmc_ext_dll.SeGetRkExpireTime(
            ctypes.byref(kmc_rk_expire_time)
        )
        if ret != RET_SUCCESS:
            raise KmcError("Get the expire time of the root key error", ret)

        return self._convert_wsec_time(kmc_rk_expire_time)

    def get_latest_master_key_create_time(self):
        kmc_mk_create_time_latest = KmcWsecSysTime()
        ret = self._kmc_ext_dll.SeGetMkCreateTimeLatest(
            ctypes.c_uint32(self._DEFAULT_DOMAIN_ID),
            ctypes.byref(kmc_mk_create_time_latest)
        )
        if ret != RET_SUCCESS:
            raise KmcError("Get the latest create time of the domain's master keys error", ret)

        return self._convert_wsec_time(kmc_mk_create_time_latest)

    def get_master_key_expire_time(self):
        kmc_mk_expire_time = KmcWsecSysTime()
        ret = self._kmc_ext_dll.SeGetMkExpireTime(
            ctypes.c_uint32(self._DEFAULT_DOMAIN_ID),
            ctypes.byref(kmc_mk_expire_time)
        )
        if ret != RET_SUCCESS:
            raise KmcError("Get the expire time of the domain's master keys error", ret)

        return self._convert_wsec_time(kmc_mk_expire_time)

    def get_key_latest_create_time(self, key_type):
        if key_type == KeyType.MASTER_KEY.value:
            mk_create_time_latest = self.get_latest_master_key_create_time()
            return mk_create_time_latest

        if key_type == KeyType.ROOT_KEY.value:
            rk_create_time_latest = self.get_latest_root_key_create_time()
            return rk_create_time_latest

    def get_interval_since_key_expire(self, key_type):
        utc_date_time = self.get_utc_date_time()

        if key_type == KeyType.MASTER_KEY.value:
            mk_expire_time = self.get_master_key_expire_time()

            return self._get_interval_time(utc_date_time, mk_expire_time)

        if key_type == KeyType.ROOT_KEY.value:
            rk_expire_time = self.get_root_key_expire_time()

            return self._get_interval_time(utc_date_time, rk_expire_time)

    def get_interval_since_last_key_create(self, key_type):
        utc_date_time = self.get_utc_date_time()

        if key_type == KeyType.MASTER_KEY.value:
            mk_create_time_latest = self.get_latest_master_key_create_time()

            return self._get_interval_time(utc_date_time, mk_create_time_latest)

        if key_type == KeyType.ROOT_KEY.value:
            rk_create_time_latest = self.get_latest_root_key_create_time()

            return self._get_interval_time(utc_date_time, rk_create_time_latest)

    def _load_so(self):
        try:
            self._kmc_ext_dll = ctypes.CDLL(KMC_LIB_NAME)
        except Exception as err:
            raise KmcError(f"Load {KMC_LIB_NAME} error {err}") from err

    def _get_cipher_data_len(self, plain_text):
        plain_text_len = len(plain_text)
        cipher_text_len_p = ctypes.c_uint32(0)
        ret = self._kmc_ext_dll.SeGetCipherDataLen(
            ctypes.c_uint32(plain_text_len),
            ctypes.byref(cipher_text_len_p)
        )
        if ret != RET_SUCCESS:
            raise KmcError("Get cipher text length error", ret)
        return cipher_text_len_p.value

    def _finalize(self):
        ret = self._kmc_ext_dll.SeFinalize(ctypes.byref(self._kmc_config))
        if ret != RET_SUCCESS:
            raise KmcError("Finalize error", ret)
        self._has_inited = False


class KmcUtil:
    _SDP_ALG_ID = "sdp_alg_id"
    _HMAC_ALG_ID = "hmac_alg_id"
    _ALG_CONFIG = namedtuple("_ALG_CONFIG", ["sdp_alg_id", "hmac_alg_id"])
    # 读取配置文件最大1MB
    _OM_ALG_MAX_SIZE_BYTES = 1 * 1024 * 1024

    def __init__(self, primary_key_store_file, standby_key_store_file, alg_cfg_file=None):
        self.primary_key_store_file = primary_key_store_file
        self.standby_key_store_file = standby_key_store_file
        self.alg_cfg = self._load_alg_from_json(alg_cfg_file)
        self._crypto_inst = None

    def __enter__(self):
        self._crypto_inst = KmcWrapper(self.alg_cfg.sdp_alg_id, self.alg_cfg.hmac_alg_id)
        self._crypto_inst.initialize(
            str_to_bytes(self.primary_key_store_file),
            str_to_bytes(self.standby_key_store_file)
        )
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        self._crypto_inst.finalize()

    def encrypt(self, plain_text, sdp_alg_id=None):
        if not sdp_alg_id:
            sdp_alg_id = self.alg_cfg.sdp_alg_id
        return self._crypto_inst.encrypt_mem_base64(plain_text, sdp_alg_id)

    def decrypt(self, cipher_text):
        return self._crypto_inst.decrypt_mem_base64(cipher_text)

    def update_rk(self, advance_update_rk_days):
        self._crypto_inst.update_root_key(advance_update_rk_days)

    def instant_update_rk(self):
        self._crypto_inst.instant_update_root_key()

    def update_mk(self, advance_update_mk_days):
        self._crypto_inst.update_master_key(advance_update_mk_days)

    def instant_update_mk(self):
        self._crypto_inst.instant_update_master_key()

    def get_mk_count(self):
        return self._crypto_inst.get_master_key_count()

    def check_and_remove_oldest_mk(self):
        self._crypto_inst.check_and_remove_oldest_master_key()

    def get_rk_latest_create_time(self):
        return self._crypto_inst.get_latest_root_key_create_time()

    def get_rk_expire_time(self):
        return self._crypto_inst.get_root_key_expire_time()

    def get_mk_latest_create_time(self):
        return self._crypto_inst.get_latest_master_key_create_time()

    def get_mk_expire_time(self):
        return self._crypto_inst.get_master_key_expire_time()

    def get_key_latest_create_time(self, key_type):
        return self._crypto_inst.get_key_latest_create_time(key_type)

    def get_interval_since_key_expire(self, key_type):
        return self._crypto_inst.get_interval_since_key_expire(key_type)

    def get_interval_since_key_create(self, key_type):
        return self._crypto_inst.get_interval_since_last_key_create(key_type)

    def _load_alg_from_json(self, alg_json_file):
        default_alg_cfg = self._ALG_CONFIG(CryptoAlgorithm.AES256_GCM.value, SignAlgorithm.HMAC_SHA512.value)
        if alg_json_file is None or not os.path.exists(alg_json_file):
            return default_alg_cfg
        if os.path.getsize(alg_json_file) > self._OM_ALG_MAX_SIZE_BYTES:
            return default_alg_cfg
        try:
            real_json_file = os.path.realpath(alg_json_file)
            with open(real_json_file, "r", encoding=DEFAULT_ENCODING) as cfg_file:
                data = json.load(cfg_file)
                sdp_alg_id = int(data.get(self._SDP_ALG_ID))
                if sdp_alg_id not in CryptoAlgorithm.value_list():
                    sdp_alg_id = default_alg_cfg.sdp_alg_id

                hmac_alg_id = int(data.get(self._HMAC_ALG_ID))
                if hmac_alg_id not in SignAlgorithm.value_list():
                    hmac_alg_id = default_alg_cfg.hmac_alg_id

                return self._ALG_CONFIG(sdp_alg_id, hmac_alg_id)
        except Exception:
            return default_alg_cfg


class Kmc(object):
    _mutex = threading.Lock()

    def __init__(self, primary_ksf, standby_ksf, cfg_path=None):
        self.primary_ksf = primary_ksf
        self.standby_ksf = standby_ksf
        self.cfg_path = cfg_path

    @staticmethod
    def available():
        """
        kmc功能是否可用
        ctypes.util.find_library不可靠
        CDLL会尝试使用dlopen函数打开动态库，找不到时抛OSError异常
        """
        try:
            for lib in KMC_DEP_LIBS:
                ctypes.CDLL(lib)
        except OSError:
            return False
        return True

    def encrypt(self, content):
        with self._mutex:
            with KmcUtil(self.primary_ksf, self.standby_ksf, self.cfg_path) as crypto:
                return crypto.encrypt(content)

    def decrypt(self, content):
        with self._mutex:
            with KmcUtil(self.primary_ksf, self.standby_ksf, self.cfg_path) as crypto:
                return crypto.decrypt(content)

    def update_rk(self, advance_update_rk_days):
        with self._mutex:
            with KmcUtil(self.primary_ksf, self.standby_ksf, self.cfg_path) as crypto:
                crypto.update_rk(advance_update_rk_days)

    def instant_update_rk(self):
        with self._mutex:
            with KmcUtil(self.primary_ksf, self.standby_ksf, self.cfg_path) as crypto:
                crypto.instant_update_rk()

    def update_mk(self, advance_update_mk_days):
        with self._mutex:
            with KmcUtil(self.primary_ksf, self.standby_ksf, self.cfg_path) as crypto:
                crypto.update_mk(advance_update_mk_days)

    def instant_update_mk(self):
        with self._mutex:
            with KmcUtil(self.primary_ksf, self.standby_ksf, self.cfg_path) as crypto:
                crypto.instant_update_mk()

    def get_mk_count(self):
        with self._mutex:
            with KmcUtil(self.primary_ksf, self.standby_ksf, self.cfg_path) as crypto:
                return crypto.get_mk_count()

    def check_and_remove_oldest_mk(self):
        with self._mutex:
            with KmcUtil(self.primary_ksf, self.standby_ksf, self.cfg_path) as crypto:
                crypto.check_and_remove_oldest_mk()

    def get_rk_latest_create_time(self):
        with self._mutex:
            with KmcUtil(self.primary_ksf, self.standby_ksf, self.cfg_path) as crypto:
                return crypto.get_rk_latest_create_time()

    def get_rk_expire_time(self):
        with self._mutex:
            with KmcUtil(self.primary_ksf, self.standby_ksf, self.cfg_path) as crypto:
                return crypto.get_rk_expire_time()

    def get_mk_latest_create_time(self):
        with self._mutex:
            with KmcUtil(self.primary_ksf, self.standby_ksf, self.cfg_path) as crypto:
                return crypto.get_mk_latest_create_time()

    def get_mk_expire_time(self):
        with self._mutex:
            with KmcUtil(self.primary_ksf, self.standby_ksf, self.cfg_path) as crypto:
                return crypto.get_mk_expire_time()

    def get_key_latest_create_time(self, key_type):
        with self._mutex:
            with KmcUtil(self.primary_ksf, self.standby_ksf, self.cfg_path) as crypto:
                return crypto.get_key_latest_create_time(key_type)

    def get_interval_since_key_expire(self, key_type):
        with self._mutex:
            with KmcUtil(self.primary_ksf, self.standby_ksf, self.cfg_path) as crypto:
                return crypto.get_interval_since_key_expire(key_type)

    def get_interval_since_key_create(self, key_type):
        with self._mutex:
            with KmcUtil(self.primary_ksf, self.standby_ksf, self.cfg_path) as crypto:
                return crypto.get_interval_since_key_create(key_type)
