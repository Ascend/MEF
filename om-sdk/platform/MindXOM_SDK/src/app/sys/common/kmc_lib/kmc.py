# coding: utf-8
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.

import json
import os
import threading

from collections import namedtuple
from datetime import datetime
from enum import Enum

RET_SUCCESS = 0
DEFAULT_ENCODING = "UTF-8"


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

    def __enter__(self):
        for keystore_file in (self.primary_key_store_file, self.standby_key_store_file):
            if not os.path.exists(keystore_file) or os.path.getsize(keystore_file) == 0:
                with open(keystore_file, "w") as fp:
                    fp.write('\n')
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        pass

    def encrypt(self, plain_text, sdp_alg_id=None):
        return plain_text

    def decrypt(self, cipher_text):
        return cipher_text

    def update_rk(self, advance_update_rk_days):
        pass

    def instant_update_rk(self):
        pass

    def update_mk(self, advance_update_mk_days):
        pass

    def instant_update_mk(self):
        pass

    def get_mk_count(self):
        return 0

    def check_and_remove_oldest_mk(self):
        pass

    def get_rk_latest_create_time(self):
        return datetime.fromtimestamp(0)

    def get_rk_expire_time(self):
        return datetime.fromtimestamp(0)

    def get_mk_latest_create_time(self):
        return datetime.fromtimestamp(0)

    def get_mk_expire_time(self):
        return datetime.fromtimestamp(0)

    def get_key_latest_create_time(self, key_type):
        return datetime.fromtimestamp(0)

    def get_interval_since_key_expire(self, key_type):
        return -1

    def get_interval_since_key_create(self, key_type):
        return 1

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
        return False

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
