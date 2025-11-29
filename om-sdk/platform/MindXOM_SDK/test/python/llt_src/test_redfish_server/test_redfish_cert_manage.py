# -*- coding: utf-8 -*-
#  Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
from collections import namedtuple

from pytest_mock import MockerFixture

from certs_manage.redfish_cert_manage import delete_unused_cert, get_unused_cert, restore_cert, extend_cert_nameger, \
    Component
from common.utils.result_base import Result
from net_manager.manager.net_cfg_manager import CertMgr


class TestRedfishCertManage:
    use_cases = {
        "test_restore_cert": {
            "fd exception": (False, Component.FD.value, Exception()),
            "fd normal": (True, Component.FD.value, [True, ]),
            "https": (False, Component.HTTPS.value, [True, ]),
        },
        "test_get_unused_cert": {
            "fd exception": (False, Component.FD.value, Exception()),
            "fd get failed": (False, Component.FD.value, ['{"unused cert": "unused cert not exist."}', ]),
            "fd normal": (True, Component.FD.value, ['{"unused cert": "test"}', ]),
            "https": (False, Component.HTTPS.value, None),
        },
        "test_delete_unused_cert": {
            "fd exception": (False, Component.FD.value, Exception()),
            "fd delete failed": (False, Component.FD.value, [Result(False, err_msg="test"), ]),
            "fd normal": (True, Component.FD.value, [Result(True, data="test"), ]),
            "https": (False, Component.HTTPS.value, None),
        },
    }

    DeleteCertCase = namedtuple("RestoreCertCase", "expect, component, del_fd_unused_cert_by_name")
    GetUnusedCertCase = namedtuple("GetUnusedCertCase", "expect, component, get_fd_unused_cert")
    RestoreCertCase = namedtuple("RestoreCertCase", "expect, component, restore_fd_pre_cert")

    @staticmethod
    def test_delete_unused_cert(mocker: MockerFixture, model: DeleteCertCase):
        mocker.patch.object(CertMgr, "del_fd_unused_cert_by_name", side_effect=model.del_fd_unused_cert_by_name)
        assert model.expect == bool(delete_unused_cert(model.component, ""))

    @staticmethod
    def test_get_unused_cert(mocker: MockerFixture, model: GetUnusedCertCase):
        mocker.patch.object(CertMgr, "get_fd_unused_cert", side_effect=model.get_fd_unused_cert)
        assert model.expect == bool(get_unused_cert(model.component, ""))

    @staticmethod
    def test_restore_cert(mocker: MockerFixture, model: RestoreCertCase):
        mocker.patch.object(CertMgr, "restore_fd_pre_cert", side_effect=model.restore_fd_pre_cert)
        assert model.expect == bool(restore_cert(model.component, ""))

    @staticmethod
    def test_extend_cert_nameger():
        assert not bool(extend_cert_nameger("test", "test"))
