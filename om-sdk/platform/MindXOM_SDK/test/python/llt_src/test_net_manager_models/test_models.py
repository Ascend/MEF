# -*- coding: utf-8 -*-
#  Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

from net_manager.models import FdPreCert, CertInfo, PreCertInfo, CertManager


class TestCertManager:

    @staticmethod
    def test_to_obj():
        assert isinstance(CertManager().to_obj({"name": "test"}), CertManager)

    @staticmethod
    def test_to_dict():
        assert CertManager(name="test").to_dict().get("name") == "test"

    @staticmethod
    def test_get_name():
        assert CertManager(name="test").get_name() == "test"


class TestCertInfo:

    @staticmethod
    def test_to_obj():
        assert isinstance(CertInfo().to_obj({"name": "test"}), CertInfo)

    @staticmethod
    def test_to_cert_info_dict():
        assert CertInfo(name="test").to_cert_info_dict().get("name") == "test"

    @staticmethod
    def test_get_cert_info():
        assert CertInfo(name="test").get_cert_info().get("name") == "test"

    @staticmethod
    def test_to_dict():
        assert CertInfo(subject="test").to_dict().get("Subject") == "test"


class TestPreCertInfo:

    @staticmethod
    def test_to_obj():
        assert isinstance(PreCertInfo().to_obj({"name": "test"}), PreCertInfo)

    @staticmethod
    def test_to_dict():
        assert PreCertInfo(name="test").to_dict().get("name") == "test"

    @staticmethod
    def test_get_cert_info():
        assert PreCertInfo(name="test").get_cert_info().get("name") == "test"

    @staticmethod
    def test_get_name():
        assert PreCertInfo(name="test").get_name() == "test"


class TestFdPreCert:

    @staticmethod
    def test_to_obj():
        assert isinstance(FdPreCert().to_obj({"name": "test"}), FdPreCert)

    @staticmethod
    def test_to_dict():
        assert FdPreCert(name="test").to_dict().get("name") == "test"

    @staticmethod
    def test_get_name():
        assert FdPreCert(name="test").get_name() == "test"

    @staticmethod
    def test_get_source():
        assert FdPreCert(source="test").get_source() == "test"
