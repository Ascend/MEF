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
from lib.Linux.systems.nfs.models import NfsCfg
from lib.Linux.systems.nfs.cfg_mgr import query_nfs_configs
from lib.Linux.systems.nfs.cfg_mgr import get_nfs_config_count
from lib.Linux.systems.nfs.cfg_mgr import del_nfs_config_by_mount_path
from lib.Linux.systems.nfs.cfg_mgr import mount_path_already_exists
from lib.Linux.systems.nfs.cfg_mgr import save_nfs_config
from lib.Linux.systems.nfs import cfg_mgr
from monitor_db.init_structure import INIT_COLUMNS


@pytest.fixture(scope="module")
def database() -> DataBase:
    with NamedTemporaryFile(suffix=".db") as tmp_file:
        db_path = tmp_file.name
    type("TestMigrate", (Migrate,),
         {"models": {NfsCfg.__tablename__: NfsCfg}}
         ).execute_on_install(db_path, INIT_COLUMNS)
    test_db = type("TestDB", (DataBase,), {})(db_path)
    with patch.object(cfg_mgr, "session_maker", test_db.session_maker):
        yield test_db
    os.remove(db_path)


class TestNfsCfgMgr:
    @staticmethod
    def test_query_nfs_configs(database: DataBase):
        with database.session_maker() as session:
            assert session.query(NfsCfg).count() == 0
            session.add(NfsCfg(server_ip="server_ip", server_dir="server_dir",
                               local_dir="mount_path", fs_type="fs_type"))
            session.add(NfsCfg(server_ip="server_ip2", server_dir="server_dir2",
                               local_dir="mount_path2", fs_type="fs_type2"))
            session.add(NfsCfg(server_ip="server_ip3", server_dir="server_dir3",
                               local_dir="mount_path3", fs_type="fs_type3"))
            for index, item in enumerate(query_nfs_configs(limit=2)):
                if index == 0:
                    assert item.server_ip == "server_ip"
                else:
                    assert item.server_ip == "server_ip2"

            assert session.query(NfsCfg).delete()

    @staticmethod
    def test_get_nfs_config_count(database: DataBase):
        with database.session_maker() as session:
            assert session.query(NfsCfg).count() == 0
            session.add(NfsCfg(server_ip="server_ip", server_dir="server_dir",
                               local_dir="mount_path", fs_type="fs_type"))
            session.add(NfsCfg(server_ip="server_ip2", server_dir="server_dir2",
                               local_dir="mount_path2", fs_type="fs_type2"))
            session.add(NfsCfg(server_ip="server_ip3", server_dir="server_dir3",
                               local_dir="mount_path3", fs_type="fs_type3"))
            assert get_nfs_config_count() == 3
            assert session.query(NfsCfg).delete()

    @staticmethod
    def test_del_nfs_config_by_mount_path(database: DataBase):
        with database.session_maker() as session:
            assert session.query(NfsCfg).count() == 0
            local_dir = "mount_path"
            session.add(NfsCfg(server_ip="server_ip", server_dir="server_dir",
                               local_dir=local_dir, fs_type="fs_type"))
            session.add(NfsCfg(server_ip="server_ip2", server_dir="server_dir2",
                               local_dir=local_dir, fs_type="fs_type2"))
            session.add(NfsCfg(server_ip="server_ip3", server_dir="server_dir3",
                               local_dir="mount_path3", fs_type="fs_type3"))
            del_nfs_config_by_mount_path(local_dir)
            assert session.query(NfsCfg).count() == 1
            assert session.query(NfsCfg).delete()

    @staticmethod
    def test_mount_path_already_exists(database: DataBase):
        with database.session_maker() as session:
            assert session.query(NfsCfg).count() == 0
            session.add(NfsCfg(server_ip="server_ip", server_dir="server_dir",
                               local_dir="mount_path", fs_type="fs_type"))
            assert mount_path_already_exists("mount_path")
            assert not mount_path_already_exists("mount path")
            assert session.query(NfsCfg).delete()

    @staticmethod
    def test_save_nfs_config(database: DataBase):
        with database.session_maker() as session:
            assert session.query(NfsCfg).count() == 0
            save_nfs_config(NfsCfg(server_ip="server_ip", server_dir="server_dir",
                                   local_dir="mount_path", fs_type="fs_type"))
            assert session.query(NfsCfg).count() == 1
            assert session.query(NfsCfg).delete()
