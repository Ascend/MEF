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
from lib.Linux.systems.security_service.models import PunyDictSign
from lib.Linux.systems.security_service import puny_dict_sign_mgr
from lib.Linux.systems.security_service.puny_dict_sign_mgr import get_puny_dict_sign
from lib.Linux.systems.security_service.puny_dict_sign_mgr import set_puny_dict_sign
from lib.Linux.systems.security_service.puny_dict import PunyDict
from monitor_db.init_structure import INIT_COLUMNS


@pytest.fixture(scope="module")
def database() -> DataBase:
    with NamedTemporaryFile(suffix=".db") as tmp_file:
        db_path = tmp_file.name
    type("TestMigrate", (Migrate,),
         {"models": {PunyDictSign.__tablename__: PunyDictSign}}
         ).execute_on_install(db_path, INIT_COLUMNS)
    test_db = type("TestDB", (DataBase,), {})(db_path)
    with patch.object(puny_dict_sign_mgr, "session_maker", test_db.session_maker):
        yield test_db
    os.remove(db_path)


class TestPunyDictSignMgr:
    @staticmethod
    def test_get_puny_dict_sign(database: DataBase):
        with database.session_maker() as session:
            assert session.query(PunyDictSign).count() == 0
            assert not get_puny_dict_sign()
            session.add(PunyDictSign(operation=PunyDict.OPERATION_TYPE_IMPORT))
            assert get_puny_dict_sign() == PunyDict.OPERATION_TYPE_IMPORT
            assert session.query(PunyDictSign).delete()

    @staticmethod
    def test_set_puny_dict_sign(database: DataBase):
        with database.session_maker() as session:
            assert session.query(PunyDictSign).count() == 0
            session.add(PunyDictSign(operation=PunyDict.OPERATION_TYPE_IMPORT))
            assert session.query(PunyDictSign).count() == 1
            set_puny_dict_sign(operation=PunyDict.OPERATION_TYPE_DELETE)
            assert get_puny_dict_sign() == PunyDict.OPERATION_TYPE_DELETE
            assert session.query(PunyDictSign).delete()


