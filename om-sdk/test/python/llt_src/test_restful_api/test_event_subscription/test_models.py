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


from om_event_subscription.models import SubsPreCert


class TestSubsPreCert:

    @staticmethod
    def test_to_obj():
        assert isinstance(SubsPreCert().from_dict({"name": "test"}), SubsPreCert)

    @staticmethod
    def test_to_dict():
        assert SubsPreCert(root_cert_id=1).to_dict().get("root_cert_id") == 1

    @staticmethod
    def test_get_name():
        assert SubsPreCert(cert_contents="test", crl_contents="test").get_cert_crl() == {
            "cert contents": "test",
            "crl contents": "test",
        }
