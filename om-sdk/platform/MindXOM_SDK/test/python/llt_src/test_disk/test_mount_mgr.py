# Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
import os
from tempfile import NamedTemporaryFile

import pytest
from mock.mock import patch

from common.db.database import DataBase
from common.db.migrate import Migrate
from lib.Linux.systems.disk.models import MountWhitelistPath
from lib.Linux.systems.disk import mount_mgr
from lib.Linux.systems.disk.mount_mgr import get_whitelist_path_count
from lib.Linux.systems.disk.mount_mgr import query_whitelist_path
from lib.Linux.systems.disk.mount_mgr import path_in_whitelist
from lib.Linux.systems.disk.mount_mgr import add_whitelist_paths
from lib.Linux.systems.disk.mount_mgr import add_whitelist_path
from lib.Linux.systems.disk.mount_mgr import delete_mount_white_path

from monitor_db.init_structure import INIT_COLUMNS


@pytest.fixture(scope="module")
def database() -> DataBase:
    with NamedTemporaryFile(suffix=".db") as tmp_file:
        db_path = tmp_file.name
    type("TestMigrate", (Migrate,),
         {"models": {MountWhitelistPath.__tablename__: MountWhitelistPath}}
         ).execute_on_install(db_path, INIT_COLUMNS)
    test_db = type("TestDB", (DataBase,), {})(db_path)
    with patch.object(mount_mgr, "session_maker", test_db.session_maker):
        yield test_db
    os.remove(db_path)


class TestMountMgr:
    @staticmethod
    def test_get_whitelist_path_count(database: DataBase):
        with database.session_maker() as session:
            assert get_whitelist_path_count() == 0
            session.add(MountWhitelistPath(path="path"))
            assert get_whitelist_path_count() == 1
            assert session.query(MountWhitelistPath).delete()

    @staticmethod
    def test_query_whitelist_path(database: DataBase):
        with database.session_maker() as session:
            assert get_whitelist_path_count() == 0
            session.add(MountWhitelistPath(path="path"))
            for item in query_whitelist_path():
                assert item.path == "path"
            assert session.query(MountWhitelistPath).delete()

    @staticmethod
    def test_path_in_whitelist(database: DataBase):
        with database.session_maker() as session:
            assert get_whitelist_path_count() == 0
            session.add(MountWhitelistPath(path="path"))
            assert path_in_whitelist("path")
            assert not path_in_whitelist("path2")
            assert session.query(MountWhitelistPath).delete()

    @staticmethod
    def test_add_whitelist_paths(database: DataBase):
        with database.session_maker() as session:
            assert get_whitelist_path_count() == 0
            add_whitelist_paths(*{"path", "path2", "path3"})
            assert get_whitelist_path_count() == 3
            assert session.query(MountWhitelistPath).delete()

    @staticmethod
    def test_add_whitelist_path(database: DataBase):
        with database.session_maker() as session:
            assert get_whitelist_path_count() == 0
            add_whitelist_path(path="path")
            assert get_whitelist_path_count() == 1
            assert session.query(MountWhitelistPath).delete()

    @staticmethod
    def test_delete_mount_white_path(database: DataBase):
        with database.session_maker() as session:
            assert get_whitelist_path_count() == 0
            add_whitelist_path(path="path")
            assert get_whitelist_path_count() == 1
            delete_mount_white_path(path="path")
            assert get_whitelist_path_count() == 0
