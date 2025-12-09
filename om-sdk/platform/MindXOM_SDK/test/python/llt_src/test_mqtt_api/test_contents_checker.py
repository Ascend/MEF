# -*- coding: utf-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.

from collections import namedtuple
from datetime import datetime

from net_manager.checkers.contents_checker import CertInfoValidator, CertContentsChecker


class TestUtils:
    @staticmethod
    def raw_value():
        return True


class TestGetJsonInfoObj:
    CheckCertInfoCase = namedtuple("CheckCertInfoCase",
                                   "expected, is_ca, ca_sign_valid, key_cert_sign, start_time, end_time, cert_version,"
                                   "pubkey_type, signature_len, chain_nums, signature_algorithm")
    CheckDictCase1 = namedtuple("CheckDictCase1", "tested_class, param_key, input_param, excepted_success")
    use_cases = {
        "test_check_cert_info": {
            "is_ca": (False, 0, True, 1, datetime(2020, 7, 3, 6, 36, 39, 135677),
                      datetime(2029, 7, 3, 6, 36, 39, 135677), 1, 0, 1, 10, ""),
            "ca_sign_invalid": (False, 1, False, 1, datetime(2020, 7, 3, 6, 36, 39, 135677),
                                datetime(2029, 7, 3, 6, 36, 39, 135677), 1, 0, 1, 10, ""),
            "key_cert_sign": (False, 1, True, 0, datetime(2020, 7, 3, 6, 36, 39, 135677),
                              datetime(2029, 7, 3, 6, 36, 39, 135677), 1, 0, 1, 10, ""),
            "start_time": (False, 1, True, 1, datetime(2029, 7, 3, 6, 36, 39, 135677),
                           datetime(2029, 7, 3, 6, 36, 39, 135677), 1, 0, 1, 10, ""),
            "cert_version": (False, 1, True, 1, datetime(2020, 7, 3, 6, 36, 39, 135677),
                             datetime(2029, 7, 3, 6, 36, 39, 135677), 1, 0, 1, 10, ""),
            "pubkey_type": (False, 1, True, 1, datetime(2020, 7, 3, 6, 36, 39, 135677),
                            datetime(2029, 7, 3, 6, 36, 39, 135677), 3, 0, 1, 10, ""),
            "signature_len": (False, 1, True, 1, datetime(2020, 7, 3, 6, 36, 39, 135677),
                              datetime(2029, 7, 3, 6, 36, 39, 135677), 3, 6, 1, 10, ""),
            "EVP_PKEY_EC": (False, 1, True, 1, datetime(2020, 7, 3, 6, 36, 39, 135677),
                            datetime(2029, 7, 3, 6, 36, 39, 135677), 3, 408, 1, 10, ""),
            "signature_algorithm": (False, 1, True, 1, datetime(2020, 7, 3, 6, 36, 39, 135677),
                                    datetime(2029, 7, 3, 6, 36, 39, 135677), 3, 408, 256, 10, ""),
            "chain_nums": (False, 0, True, 1, datetime(2020, 7, 3, 6, 36, 39, 135677),
                           datetime(2029, 7, 3, 6, 36, 39, 135677), 1, 0, 1, 10, ""),
            "normal": (True, 1, True, 1, datetime(2020, 7, 3, 6, 36, 39, 135677),
                       datetime(2029, 7, 3, 6, 36, 39, 135677), 3, 408, 256, 9, "sha256WithRSAEncryption"),
        },
        "test_check_dict1": {
            "not_exists_checker_is_existed": (CertContentsChecker, "1", {"1": "1"}, False),
            "not_exists_checker_not_existed": (CertContentsChecker, "2", {"1": "1"}, False),
        },
    }

    def test_check_cert_info(self, model: CheckCertInfoCase):
        class TestUtilsCertVversion:
            is_ca = model.is_ca
            key_cert_sign = model.key_cert_sign
            start_time = model.start_time
            end_time = model.end_time
            cert_version = model.cert_version
            pubkey_type = model.pubkey_type
            signature_len = model.signature_len
            signature_algorithm = model.signature_algorithm
            chain_num = model.chain_nums
            ca_sign_valid = model.ca_sign_valid

        ret = True
        try:
            CertInfoValidator().check_cert_info(TestUtilsCertVversion)
        except Exception:
            ret = False
        assert ret == model.expected

    def test_check_dict1(self, model: CheckDictCase1):
        ret = model.tested_class(model.param_key).check_dict(model.input_param)
        assert model.excepted_success == ret.success
