# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import os
from tempfile import NamedTemporaryFile

import pytest
from mock.mock import patch

from common.db.database import DataBase
from common.db.migrate import Migrate
from lib.Linux.systems.nic import nic_mgr
from lib.Linux.systems.nic.models import NetConfig
from lib.Linux.systems.nic.nic_mgr import NetConfigMgr
from monitor_db.init_structure import INIT_COLUMNS


@pytest.fixture(scope="module")
def database() -> DataBase:
    with NamedTemporaryFile(suffix=".db") as tmp_file:
        db_path = tmp_file.name
    type("TestMigrate", (Migrate,),
         {"models": {NetConfig.__tablename__: NetConfig}}
         ).execute_on_install(db_path, INIT_COLUMNS)
    test_db = type("TestDB", (DataBase,), {})(db_path)
    with patch.object(nic_mgr, "session_maker", test_db.session_maker):
        yield test_db
    os.remove(db_path)


class TestNetConfigMgr:
    @staticmethod
    def test_query_info_with_condition(database: DataBase):
        with database.session_maker() as session:
            tag = "web"
            session.add(NetConfig(name="2", ipv4="ipv4", tag=tag))
            session.add(NetConfig(name="1", ipv4="ip", tag=tag))
        for index, item in enumerate(NetConfigMgr.query_info_with_condition(tag=tag)):
            if index == 0:
                assert item.name == "1" and item.ipv4 == "ip"
            else:
                assert item.name == "2" and item.ipv4 == "ipv4"

        with database.session_maker() as session:
            assert session.query(NetConfig).delete()

    @staticmethod
    def test_query_tag_from_ip(database: DataBase):
        with database.session_maker() as session:
            ipv4 = "ipv4"
            session.add(NetConfig(name="2", ipv4=ipv4, tag="web"))
            session.add(NetConfig(name="1", ipv4="dhcp", tag="test tag"))
        assert NetConfigMgr.query_tag_from_ip(ipv4, False) == "web"
        assert NetConfigMgr.query_tag_from_ip(ipv4, True) == "test tag"
        with database.session_maker() as session:
            assert session.query(NetConfig).delete()

    @staticmethod
    def test_delete_specific_eth_config(database: DataBase):
        with database.session_maker() as session:
            session.add(NetConfig(name="2", ipv4="ipv4", tag="web"))
            session.add(NetConfig(name="1", ipv4="dhcp", tag="test tag"))
        NetConfigMgr.delete_specific_eth_config(tag="web")
        with database.session_maker() as session:
            assert session.query(NetConfig).count() == 1
            assert session.query(NetConfig).delete()

    @staticmethod
    def test_save_net_config(database: DataBase):
        with database.session_maker() as session:
            assert session.query(NetConfig).count() == 0
        NetConfigMgr.save_net_config([NetConfig(name="eth_name", ipv4="dhcp", tag="web")])
        with database.session_maker() as session:
            assert session.query(NetConfig).count() == 1
            assert session.query(NetConfig).delete()

