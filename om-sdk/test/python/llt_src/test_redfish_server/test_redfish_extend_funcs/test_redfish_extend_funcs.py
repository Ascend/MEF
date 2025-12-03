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

from pytest_mock import MockerFixture

from om_event_subscription.models import SubsPreCert
from om_event_subscription.subscription_mgr import PreSubCertMgr
from redfish_extend_funcs import get_unused_cert, restore_cert, delete_cert


class TestGetUnusedCert:
    use_cases = {
        "test_get_unused_cert": {
            "raise exception": (False, Exception()),
            "normal": (True, [SubsPreCert(), ]),
        },
        "test_restore_cert": {
            "raise exception": (False, Exception()),
            "normal": (True, [True, ]),
        },
        "test_delete_cert": {
            "no unused cert": (False, False, None),
            "succeed": (True, True, [True, ]),
            "failed": (False, True, [False, ]),
        },
    }

    GetUnusedCertCase = namedtuple("GetUnusedCertCase", "expect, get_pre_subs_cert")
    RestoreCertCase = namedtuple("RestoreCertCase", "expect, restore_pre_subs_cert")
    DeleteCertCase = namedtuple("RestoreCertCase", "expect, get_unused_cert, delete_cert")

    @staticmethod
    def test_delete_cert(mocker: MockerFixture, model: DeleteCertCase):
        mocker.patch("redfish_extend_funcs.get_unused_cert", return_value=model.get_unused_cert)
        mocker.patch.object(PreSubCertMgr, "delete_cert", side_effect=model.delete_cert)
        assert model.expect == bool(delete_cert())

    @staticmethod
    def test_restore_cert(mocker: MockerFixture, model: RestoreCertCase):
        mocker.patch.object(PreSubCertMgr, "restore_pre_subs_cert", side_effect=model.restore_pre_subs_cert)
        assert model.expect == bool(restore_cert())

    @staticmethod
    def test_get_unused_cert(mocker: MockerFixture, model: GetUnusedCertCase):
        mocker.patch.object(PreSubCertMgr, "get_pre_subs_cert", side_effect=model.get_pre_subs_cert)
        assert model.expect == bool(get_unused_cert())
