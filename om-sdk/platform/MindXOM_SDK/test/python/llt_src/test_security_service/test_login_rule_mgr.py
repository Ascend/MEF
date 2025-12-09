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
from lib.Linux.systems.security_service.models import LoginRules
from lib.Linux.systems.security_service.login_rule_mgr import LoginRuleManager
from lib.Linux.systems.security_service import login_rule_mgr
from monitor_db.init_structure import INIT_COLUMNS


@pytest.fixture(scope="module")
def database() -> DataBase:
    with NamedTemporaryFile(suffix=".db") as tmp_file:
        db_path = tmp_file.name
    type("TestMigrate", (Migrate,),
         {"models": {LoginRules.__tablename__: LoginRules}}
         ).execute_on_install(db_path, INIT_COLUMNS)
    test_db = type("TestDB", (DataBase,), {})(db_path)
    with patch.object(login_rule_mgr, "session_maker", test_db.session_maker):
        yield test_db
    os.remove(db_path)


class TestLoginRuleManager:
    @staticmethod
    def test_get_all(database: DataBase):
        with database.session_maker() as session:
            assert not LoginRuleManager().get_all()
            session.add(LoginRules(enable="enable", start_time="start_time",
                                   end_time="end_time", ip_addr="ip_addr", mac_addr="mac_addr"))
            assert len(LoginRuleManager().get_all()) == 1
            assert session.query(LoginRules).delete()

    @staticmethod
    def test_over_write_database(database: DataBase):
        with database.session_maker() as session:
            session.add(LoginRules(enable="enable1", start_time="start time",
                                   end_time="end time", ip_addr="ip addr", mac_addr="mac addr"))
            obj_list = {
                LoginRules(enable="enable2", start_time="start_time2",
                           end_time="end_time2", ip_addr="ip_addr2", mac_addr="mac_addr2"),
                LoginRules(enable="enable3", start_time="start_time3",
                           end_time="end_time3", ip_addr="ip_addr3", mac_addr="mac_addr3"),
            }
            LoginRuleManager().over_write_database(obj_list)
            for item in LoginRuleManager().get_all():
                assert item["enable"] != "enable1"
                assert item["start_time"] != "start time"
                assert item["end_time"] != "end time"
                assert item["ip_addr"] != "ip addr"
                assert item["mac_addr"] != "mac addr"
            assert session.query(LoginRules).delete()
