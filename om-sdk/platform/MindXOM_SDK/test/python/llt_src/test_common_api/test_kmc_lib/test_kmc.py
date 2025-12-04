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
import ctypes
from collections import namedtuple
from datetime import datetime

import pytest
from pytest_mock import MockerFixture

from common.kmc_lib.kmc import KmcEnum, KeyAdaptorError, Role, CryptoAlgorithm, SignAlgorithm, KeyType, KmcConfig, \
    KmcWsecSysTime, KmcError, str_to_bytes, free_char_buffer, KmcWrapper
from conftest import TestBase


class TestKmcEnum(TestBase):

    TestKmcEnumCase = namedtuple("TestKmcEnumCase", "expect, data")
    TestKeyAdaptorErrorCase = namedtuple("TestKeyAdaptorError", "expect, error_msg, error_code")
    TestRoleCase = namedtuple("TestRoleCase", "expect")
    TestCryptoAlgorithmCase = namedtuple("TestCryptoAlgorithmCase", "expect")
    TestSignAlgorithmCase = namedtuple("TestSignAlgorithmCase", "expect")
    TestKeyTypeCase = namedtuple("TestKeyTypeCase", "expect")
    TestKmcConfigCase = namedtuple("TestKmcConfigCase", "expect")
    TestKmcWsecSysTimeCase = namedtuple("TestKmcWsecSysTimeCase", "expect")
    TestKmcErrorCase = namedtuple("TestKmcErrorCase", "expect, error_msg, error_code")
    TestStrToBytes = namedtuple("TestStrToBytes", "expect, data")
    TestFreeCharBuffer = namedtuple("TestFreeCharBuffer", "expect, data")

    use_cases = {
        "test_value_list": {
            "None": ([], [])
        },
        "test_key_adapt_error": {
            "no_param": (["", 0], "", 0),
            "key": (["error message", 1], "error message", 1),
        },
        "test_role": {
            "normal": ([0, 1], ),
        },
        "test_crypto_algorithm": {
            "normal": ([8, 9], ),
        },
        "test_sign_algorithm": {
            "normal": ([2053, 2054], ),
        },
        "test_key_type": {
            "normal": (["root key", "master key"], ),
        },
        "test_kmc_config": {
            "normal": ([b"", b"", 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0], ),
        },
        "test_kmc_wsec_sys_time": {
            "normal": ([0, 0, 0, 0, 0, 0, 0],),
        },
        "test_kmc_error": {
            "no_code": ("error message", "error message", None),
            "with_code": ("KmcError[1] error message", "error message", 1),
        },
        "test_str_to_bytes": {
            "bytes": (b"1", b"1"),
            "string": (b"a", "a"),
            "exception": (None, {1: 1}),
        },
        "test_free_char_buffer": {
            "bytes": (b"", b"Hello, world!"),
            "int": (b"", 111),
        },
    }

    @staticmethod
    def test_value_list(model: TestKmcEnumCase):
        assert model.expect == KmcEnum.value_list()

    @staticmethod
    def test_key_adapt_error(model: TestKeyAdaptorErrorCase):
        assert model.expect[0] == KeyAdaptorError(error_msg=model.error_msg, error_code=model.error_code).error_msg
        assert model.expect[1] == KeyAdaptorError(error_msg=model.error_msg, error_code=model.error_code).error_code

    @staticmethod
    def test_role(model: TestRoleCase):
        assert model.expect == Role.value_list()

    @staticmethod
    def test_crypto_algorithm(model: TestCryptoAlgorithmCase):
        assert model.expect == CryptoAlgorithm.value_list()

    @staticmethod
    def test_sign_algorithm(model: TestSignAlgorithmCase):
        assert model.expect == SignAlgorithm.value_list()

    @staticmethod
    def test_key_type(model: TestKeyTypeCase):
        assert model.expect == KeyType.value_list()

    @staticmethod
    def test_kmc_config(model: TestKmcConfigCase):
        assert model.expect[0] == KmcConfig().primaryKeyStoreFile
        assert model.expect[1] == KmcConfig().standbyKeyStoreFile
        assert model.expect[2] == KmcConfig().domainCount
        assert model.expect[3] == KmcConfig().role
        assert model.expect[4] == KmcConfig().procLockPerm
        assert model.expect[5] == KmcConfig().sdpAlgId
        assert model.expect[6] == KmcConfig().hmacAlgId
        assert model.expect[7] == KmcConfig().semKey
        assert model.expect[8] == KmcConfig().innerSymmAlgId
        assert model.expect[9] == KmcConfig().innerHashAlgId
        assert model.expect[10] == KmcConfig().innerHmacAlgId
        assert model.expect[11] == KmcConfig().innerKdfAlgId
        assert model.expect[12] == KmcConfig().workKeyIter
        assert model.expect[13] == KmcConfig().rootKeyIter
        assert model.expect[14] == KmcConfig().version

    @staticmethod
    def test_kmc_wsec_sys_time(model: TestKmcWsecSysTimeCase):
        assert model.expect[0] == KmcWsecSysTime().kmcYear
        assert model.expect[1] == KmcWsecSysTime().kmcMonth
        assert model.expect[2] == KmcWsecSysTime().kmcDate
        assert model.expect[3] == KmcWsecSysTime().kmcHour
        assert model.expect[4] == KmcWsecSysTime().kmcMinute
        assert model.expect[5] == KmcWsecSysTime().kmcSecond
        assert model.expect[6] == KmcWsecSysTime().kmcWeek

    @staticmethod
    def test_kmc_error(model: TestKmcErrorCase):
        assert model.expect == str(KmcError(error_msg=model.error_msg, error_code=model.error_code))

    @staticmethod
    def test_str_to_bytes(model: TestStrToBytes):
        if isinstance(model.data, bytes) or isinstance(model.data, str):
            assert model.expect == str_to_bytes(model.data)
        else:
            with pytest.raises(KeyAdaptorError):
                str_to_bytes(model.data)

    @staticmethod
    def test_free_char_buffer(model: TestFreeCharBuffer):
        char_p = ctypes.create_string_buffer(model.data)
        free_char_buffer(char_p)
        assert model.expect == char_p.value


class TestKmcWrapper(TestBase):
    KmcWrapperCaseWithPara1 = namedtuple("KmcWrapperCaseWithPara1", "expect, para1")
    KmcWrapperCaseWithPara2 = namedtuple("KmcWrapperCaseWithPara1", "expect, para1, para2")

    use_cases = {
        "test_convert_wsec_time": {
            "test1": ("1970-01-01 00:00:00", (1970, 1, 1, 0, 0, 0, 0)),
            "test2": ("2024-09-01 00:00:00", (2024, 9, 1, 0, 0, 0, 0)),
        },
        "test_get_interval_time": {
            "test1": (31536000.0, (2023, 1, 1), (2024, 1, 1)),
        },
    }

    @staticmethod
    def test_convert_wsec_time(mocker: MockerFixture, model: KmcWrapperCaseWithPara1):
        mocker.patch.object(KmcWrapper, '_load_so', return_value="")
        instance = KmcWrapper()
        assert str(instance._convert_wsec_time(KmcWsecSysTime(*model.para1))) == model.expect

    @staticmethod
    def test_get_interval_time(mocker: MockerFixture, model: KmcWrapperCaseWithPara2):
        mocker.patch.object(KmcWrapper, '_load_so', return_value="")
        instance = KmcWrapper()
        assert instance._get_interval_time(datetime(*model.para2), datetime(*model.para1)) == model.expect
