# -*- coding: utf-8 -*-
#  Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
import os
from tempfile import NamedTemporaryFile

import pytest
from pytest_mock import MockerFixture

from common.db.database import DataBase
from common.db.migrate import Migrate
from om_event_subscription import subscription_mgr
from om_event_subscription.models import SubscriptionCert, SubsPreCert
from om_event_subscription.subscription_mgr import PreSubCertMgr
from om_redfish_db.init_structure import INIT_COLUMNS


class TestMigrate(Migrate):

    @classmethod
    def instance(cls):
        return cls._instances.get(cls)


@pytest.fixture(scope="package")
def database(package_mocker: MockerFixture) -> DataBase:
    with NamedTemporaryFile(suffix=".db") as tmp_file:
        db_path = tmp_file.name
    TestMigrate.register_models(SubscriptionCert, SubsPreCert)
    TestMigrate.execute_on_install(db_path, INIT_COLUMNS)
    test_db = TestMigrate.instance()
    package_mocker.patch.object(subscription_mgr, "session_maker", test_db.session_maker)
    yield test_db
    os.remove(db_path)


class TestPreSubCertMgr:

    @staticmethod
    def test_del_by_name(database: DataBase):
        with database.session_maker() as session:
            session.query(SubscriptionCert).delete()
            session.add(SubscriptionCert(id=1))
        assert PreSubCertMgr().get_using_cert()

    @staticmethod
    def test_backup_pre_subs_cert(database: DataBase):
        with database.session_maker() as session:
            session.query(SubsPreCert).delete()
            session.query(SubscriptionCert).delete()
            session.add(SubscriptionCert(id=1))
            PreSubCertMgr().backup_pre_subs_cert()
            assert session.query(SubsPreCert)

    @staticmethod
    def test_get_pre_subs_cert(database: DataBase):
        with database.session_maker() as session:
            session.query(SubsPreCert).delete()
            session.add(SubsPreCert(id=1))
        assert PreSubCertMgr().get_pre_subs_cert()

    @staticmethod
    def test_delete_cert(database: DataBase):
        with database.session_maker() as session:
            session.query(SubsPreCert).delete()
            session.add(SubsPreCert(id=1))
        assert PreSubCertMgr().delete_cert() == 1

    @staticmethod
    def test_check_cert(mocker: MockerFixture, database: DataBase):
        mocker.patch.object(PreSubCertMgr, "model_checker").locked.return_value = True
        assert not PreSubCertMgr().check_cert("test")

    @staticmethod
    def test_restore_pre_subs_cert(mocker: MockerFixture, database: DataBase):
        mocker.patch.object(PreSubCertMgr, "check_cert")
        with database.session_maker() as session:
            session.query(SubscriptionCert).delete()
            session.add(SubsPreCert(cert_contents="test"))
            PreSubCertMgr().restore_pre_subs_cert()
            assert session.query(SubscriptionCert)
