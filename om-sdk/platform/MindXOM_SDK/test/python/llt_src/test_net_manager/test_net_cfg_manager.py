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
import json
import uuid

import pytest
from pytest_mock import MockerFixture

from common.db.database import DataBase
from net_manager.constants import NetManagerConstants
from net_manager.manager.net_cfg_manager import CertMgr
from net_manager.models import CertManager, CertInfo, FdPreCert, PreCertInfo, NetManager


class TestCertMgr:

    @staticmethod
    def init_database(database: DataBase):
        with database.session_maker() as session:
            session.query(CertManager).delete()
            session.query(CertInfo).delete()
            session.query(FdPreCert).delete()
            session.query(PreCertInfo).delete()

    @staticmethod
    def test_del_by_name(database: DataBase):
        with database.session_maker() as session:
            session.query(CertManager).delete()
            session.query(CertInfo).delete()
            session.add(CertManager(name="test"))
            session.add(CertInfo(name="test"))
        assert CertMgr().del_by_name("test") == 1

    @staticmethod
    def test_get_used_cert_name(database: DataBase):
        with database.session_maker() as session:
            session.query(CertInfo).delete()
            session.add(CertInfo(name="test", in_use=True))
            assert CertMgr().get_used_cert_name(session) == "test"

    @staticmethod
    def test_get_used_cert(database: DataBase):
        with database.session_maker() as session:
            session.query(CertInfo).delete()
            session.query(CertManager).delete()
            session.add(CertInfo(name="test", in_use=True))
            session.add(CertManager(name="test"))
            assert CertMgr().get_used_cert()

    @staticmethod
    def test_get_cert_info_by_name(database: DataBase):
        with database.session_maker() as session:
            session.query(CertInfo).delete()
            session.add(CertInfo(name="test"))
            assert CertMgr().get_cert_info_by_name("test")

    @staticmethod
    def test_del_cert_info_by_name(database: DataBase):
        with database.session_maker() as session:
            session.query(CertInfo).delete()
            session.add(CertInfo(name="test"))
            assert CertMgr().del_cert_info_by_name("test") == 1

    @staticmethod
    def test_get_fd_pre_cert(database: DataBase):
        with database.session_maker() as session:

            session.add(FdPreCert(name="test"))
            assert CertMgr().get_fd_pre_cert()

    @staticmethod
    def test_get_fd_pre_cert_info(database: DataBase):
        with database.session_maker() as session:
            session.query(PreCertInfo).delete()
            session.add(PreCertInfo(name="test"))
            assert CertMgr().get_fd_pre_cert_info()

    @staticmethod
    def test_get_unused_cert_info(database: DataBase):
        with database.session_maker() as session:
            session.query(CertInfo).delete()
            session.add(CertInfo(name="test", in_use=False))
            assert CertMgr().get_unused_cert_info()

    @staticmethod
    def test_get_fd_pre_cert_name(database: DataBase):
        with database.session_maker() as session:
            session.query(FdPreCert).delete()
            session.add(FdPreCert(name="test"))
            assert CertMgr().get_fd_pre_cert_name() == "test"

    @staticmethod
    def test_check_pre_cert_exists(database: DataBase):
        with database.session_maker() as session:
            session.query(FdPreCert).delete()
            session.query(CertManager).delete()
            session.add(CertManager(name="test"))
            session.add(FdPreCert(name="test"))
            assert CertMgr().check_pre_cert_exists()

    @staticmethod
    def test_backup_previous_cert_with_used_cert(database: DataBase):
        TestCertMgr.init_database(database)
        with database.session_maker() as session:
            session.add(CertManager(name="test"))
            session.add(CertInfo(name="test", in_use=True))
            CertMgr().backup_previous_cert()
            assert session.query(FdPreCert)

    @staticmethod
    def test_backup_previous_cert_with_no_web_source(database: DataBase):
        TestCertMgr.init_database(database)
        with database.session_maker() as session:
            session.add(CertManager(name="test", source="test"))
            session.add(CertInfo(name="test"))
            assert not CertMgr().backup_previous_cert()

    @staticmethod
    def test_backup_previous_cert_without_used_cert(database: DataBase):
        TestCertMgr.init_database(database)
        with database.session_maker() as session:
            session.add(CertManager(name="test", source=NetManagerConstants.WEB))
            session.add(CertInfo(name="test"))
            session.add(FdPreCert(name="pre"))
            session.add(PreCertInfo(name="pre"))
            CertMgr().backup_previous_cert()
            assert session.query(FdPreCert).first().get_name() == "test"

    @staticmethod
    def test_check_fd_connection_failed(database: DataBase):
        TestCertMgr.init_database(database)
        with database.session_maker() as session:
            session.query(NetManager).delete()
        assert not CertMgr().check_fd_connection()

    @staticmethod
    def test_check_fd_connection_succeed(database: DataBase):
        TestCertMgr.init_database(database)
        with database.session_maker() as session:
            session.query(NetManager).delete()
            session.add(NetManager(
                net_mgmt_type=NetManagerConstants.FUSION_DIRECTOR,
                node_id=str(uuid.uuid4()),
                server_name="",
                ip="1.1.1.1",
                port="443",
                cloud_user="test",
                cloud_pwd="test",
                status="ready"
            ))
        assert CertMgr().check_fd_connection()

    @staticmethod
    def test_restore_fd_pre_cert_with_unconnected(mocker: MockerFixture, database: DataBase):
        mocker.patch.object(CertMgr, "check_pre_cert_is_invalid")
        TestCertMgr.init_database(database)
        with database.session_maker() as session:
            session.query(NetManager).delete()
            session.add(FdPreCert(name="pre"))
            session.add(PreCertInfo(name="pre"))
            CertMgr().restore_fd_pre_cert()
            assert session.query(CertInfo) and session.query(CertManager)

    @staticmethod
    def test_restore_fd_pre_cert_with_exist_cert(mocker: MockerFixture, database: DataBase):
        mocker.patch.object(CertMgr, "check_pre_cert_is_invalid")
        mocker.patch.object(CertMgr, "check_fd_connection", return_value=True)
        mocker.patch.object(CertMgr, "check_pre_cert_exists", return_value=True)
        TestCertMgr.init_database(database)
        with pytest.raises(Exception):
            with database.session_maker() as session:
                session.add(CertManager(name="test"))
                session.add(CertInfo(name="test"))
                session.add(FdPreCert(name="pre"))
                session.add(PreCertInfo(name="pre"))
                CertMgr().restore_fd_pre_cert()

    @staticmethod
    def test_restore_fd_pre_cert_succeed(mocker: MockerFixture, database: DataBase):
        mocker.patch.object(CertMgr, "check_pre_cert_is_invalid")
        mocker.patch.object(CertMgr, "check_fd_connection", return_value=True)
        TestCertMgr.init_database(database)
        with database.session_maker() as session:
            session.add(CertManager(name="test", source="test"))
            session.add(CertInfo(name="test"))
            session.add(FdPreCert(name="pre", source="test"))
            session.add(PreCertInfo(name="pre"))
            CertMgr().restore_fd_pre_cert()
            assert session.query(CertInfo).first().get_cert_name() == "pre" \
                   and session.query(CertManager).first().get_name() == "pre"

    @staticmethod
    def test_get_fd_unused_cert_with_no_pre(mocker: MockerFixture, database: DataBase):
        mocker.patch.object(CertMgr, "get_unused_cert_info", return_value={"test": "test"})
        TestCertMgr.init_database(database)
        with database.session_maker() as session:
            session.add(CertManager(name="test"))
            session.add(CertInfo(name="test"))
        assert CertMgr().get_fd_unused_cert() == json.dumps({CertMgr.UNUNSED_KEY: ["test"]})

    @staticmethod
    def test_get_fd_unused_cert_succeed(mocker: MockerFixture, database: DataBase):
        mocker.patch.object(CertMgr, "get_unused_cert_info", return_value={"test": "test"})
        TestCertMgr.init_database(database)
        with database.session_maker() as session:
            session.add(FdPreCert(name="test"))
            session.add(PreCertInfo(name="test"))
        assert CertMgr().get_fd_unused_cert() == json.dumps({
            CertMgr.UNUNSED_KEY: ["test"],
            CertMgr.PRE_KEY: {"name": "test",
                              "subject": None,
                              "issuer": None,
                              "serial num": None,
                              "signature algorithm": None,
                              "signature length": None,
                              "pubkey type": None,
                              "fingerprint": None}
        })

    @staticmethod
    def test_del_pre_cert_by_name(database: DataBase):
        TestCertMgr.init_database(database)
        with database.session_maker() as session:
            session.add(PreCertInfo(name="test"))
            session.add(FdPreCert(name="test"))
        assert CertMgr().del_pre_cert_by_name("test") == 1

    @staticmethod
    def test_check_cert_exist_by_name_with_cert(database: DataBase):
        TestCertMgr.init_database(database)
        with database.session_maker() as session:
            session.add(CertInfo(name="test"))
        assert CertMgr().check_cert_exist_by_name("test")

    @staticmethod
    def test_check_cert_exist_by_name_with_pre_cert(database: DataBase):
        TestCertMgr.init_database(database)
        with database.session_maker() as session:
            session.add(PreCertInfo(name="test"))
        assert CertMgr().check_cert_exist_by_name("test")

    @staticmethod
    def test_check_cert_exist_by_name_false(database: DataBase):
        TestCertMgr.init_database(database)
        with database.session_maker() as session:
            session.add(PreCertInfo(name="test"))
            session.add(CertInfo(name="test"))
        assert not CertMgr().check_cert_exist_by_name("dddd")

    @staticmethod
    def test_del_fd_unused_cert_by_name_with_invalid_name(database: DataBase):
        TestCertMgr.init_database(database)
        with pytest.raises(Exception):
            CertMgr().del_fd_unused_cert_by_name("test..crt")

    @staticmethod
    def test_del_fd_unused_cert_by_name_with_cert_not_exist(mocker: MockerFixture, database: DataBase):
        mocker.patch.object(CertMgr, "check_cert_exist_by_name", return_value=False)
        TestCertMgr.init_database(database)
        assert not bool(CertMgr().del_fd_unused_cert_by_name("test.crt"))

    @staticmethod
    def test_del_fd_unused_cert_by_name_with_delete_pre_failed(mocker: MockerFixture, database: DataBase):
        mocker.patch.object(CertMgr, "check_cert_exist_by_name", return_value=True)
        TestCertMgr.init_database(database)
        with database.session_maker() as session:
            session.add(CertInfo(name="test.crt", in_use=True))
            session.add(CertManager(name="test.crt"))
            session.add(PreCertInfo(name="pre.crt"))
            session.add(FdPreCert(name="pre.crt"))
        assert not bool(CertMgr().del_fd_unused_cert_by_name("test.crt"))

    @staticmethod
    def test_del_fd_unused_cert_by_name_with_delete_pre_succeed(mocker: MockerFixture, database: DataBase):
        mocker.patch.object(CertMgr, "check_cert_exist_by_name", return_value=True)
        TestCertMgr.init_database(database)
        with database.session_maker() as session:
            session.add(CertInfo(name="test.crt", in_use=True))
            session.add(CertManager(name="test.crt"))
            session.add(PreCertInfo(name="test.crt"))
            session.add(FdPreCert(name="test.crt"))
        assert bool(CertMgr().del_fd_unused_cert_by_name("test.crt"))

    @staticmethod
    def test_del_fd_unused_cert_by_name_succeed(mocker: MockerFixture, database: DataBase):
        mocker.patch.object(CertMgr, "check_cert_exist_by_name", return_value=True)
        TestCertMgr.init_database(database)
        with database.session_maker() as session:
            session.add(CertInfo(name="test.crt"))
            session.add(CertManager(name="test.crt"))
            session.add(PreCertInfo(name="test.crt"))
            session.add(FdPreCert(name="test.crt"))
        assert bool(CertMgr().del_fd_unused_cert_by_name("test.crt"))
