# -*- coding: utf-8 -*-
#  Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
from collections import namedtuple

from pytest_mock import MockerFixture

from common.file_utils import FileCheck, FileReader, FileUtils
from common.kmc_lib.kmc import Kmc
from common.kmc_lib.tlsconfig import TlsConfig
from common.utils.result_base import Result
from nginx_cert_manage import NginxCertMgr


class TestNginxCertMgr:
    use_cases = {
        "test_check_cert_pkey_match": {
            "decrypt failed": (False, Exception, None),
            "get context failed": (False, ["test", ], [False, ""]),
            "succeed": (True, ["test", ], [True, ""]),
        },
        "test_check_cert_is_valid": {
            "invalid file": (False, [Result(False, err_msg="test"), ], None, None),
            "pkey not match": (
                False,
                [Result(True), Result(True), Result(True), Result(True), Result(True), Result(True)], False, None
            ),
            "expired": (
                False,
                [Result(True), Result(True), Result(True), Result(True), Result(True), Result(True)], True, False
            ),
            "succeed": (
                True,
                [Result(True), Result(True), Result(True), Result(True), Result(True), Result(True)], True, True
            ),
        },
        "test_restore_pre_cert": {
            "invalid file": (False, Result(False, err_msg="test"), None),
            "copy exception": (False, Result(True), Exception),
            "normal": (True, Result(True), [True, ]),
        },
        "test_get_unused_cert": {
            "invalid file": (False, Result(False, err_msg="test"), None),
            "read failed": (False, Result(True), [Result(False, err_msg="test"), ]),
            "normal": (True, Result(True), [Result(True, data="test"), ]),
        },
        "test_delete_unused_cert": {
            "no unused cert": (False, Result(False, err_msg="test"), None),
            "delete exception": (False, Result(True), [False, ]),
            "normal": (True, Result(True), [True, True, True, True, True, True, True]),
        },
    }

    CheckCertPkeyMatchCase = namedtuple("CheckCertPkeyMatchCase", "expect, decrypt, get_ssl_context")
    CheckCertValidCase = namedtuple("CheckCertValidCase",
                                    "expect, check_path_is_exist_and_valid, check_cert_pkey_match, check_cert_expired")
    RestorePreCertCase = namedtuple("RestorePreCertCase", "expect, check_cert_is_valid, copy_pre_cert_to_work_dir")
    GetUnusedCertCase = namedtuple("GetUnusedCertCase", "expect, check_cert_is_valid, read")
    DeleteUnusedCertCase = namedtuple("GetUnusedCertCase", "expect, get_unused_cert, delete_file_or_link")

    @staticmethod
    def test_delete_unused_cert(mocker: MockerFixture, model: DeleteUnusedCertCase):
        mocker.patch.object(NginxCertMgr, "get_unused_cert", return_value=model.get_unused_cert)
        mocker.patch.object(FileUtils, "delete_file_or_link", side_effect=model.delete_file_or_link)
        assert model.expect == bool(NginxCertMgr().delete_unused_cert())

    @staticmethod
    def test_get_unused_cert(mocker: MockerFixture, model: GetUnusedCertCase):
        mocker.patch.object(NginxCertMgr, "check_cert_is_valid", return_value=model.check_cert_is_valid)
        mocker.patch.object(FileReader, "read", side_effect=model.read)
        assert model.expect == bool(NginxCertMgr().get_unused_cert())

    @staticmethod
    def test_restore_pre_cert(mocker: MockerFixture, model: RestorePreCertCase):
        mocker.patch.object(NginxCertMgr, "check_cert_is_valid", return_value=model.check_cert_is_valid)
        mocker.patch.object(NginxCertMgr, "copy_pre_cert_to_work_dir", side_effect=model.copy_pre_cert_to_work_dir)
        assert model.expect == bool(NginxCertMgr().restore_pre_cert())

    @staticmethod
    def test_check_cert_is_valid(mocker: MockerFixture, model: CheckCertValidCase):
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid", side_effect=model.check_path_is_exist_and_valid)
        mocker.patch("os.path.getsize", return_value=1024)
        mocker.patch.object(NginxCertMgr, "check_cert_pkey_match", return_value=model.check_cert_pkey_match)
        mocker.patch("lib.Linux.systems.security_service.security_service_clib.check_cert_expired",
                     return_value=model.check_cert_expired)
        assert model.expect == bool(NginxCertMgr().check_cert_is_valid())

    @staticmethod
    def test_check_cert_pkey_match(mocker: MockerFixture, model: CheckCertPkeyMatchCase):
        mocker.patch("builtins.open")
        mocker.patch.object(Kmc, "decrypt", side_effect=model.decrypt)
        mocker.patch.object(TlsConfig, "get_ssl_context", return_value=model.get_ssl_context)
        assert model.expect == bool(NginxCertMgr().check_cert_pkey_match())
