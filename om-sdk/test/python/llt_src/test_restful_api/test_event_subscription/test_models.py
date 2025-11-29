# -*- coding: utf-8 -*-
#  Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

# -*- coding: utf-8 -*-
#  Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.


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
