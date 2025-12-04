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
import queue
from pathlib import Path
from tempfile import NamedTemporaryFile
from typing import NamedTuple, Optional, Tuple, Iterable, List

import pytest
from mock.mock import patch
from pytest_mock import MockerFixture

from common.db.database import DataBase
from common.db.migrate import Migrate
from common.file_utils import FileCheck
from common.utils.exec_cmd import ExecCmd
from common.utils.result_base import Result
from lib.Linux.EdgeSystem import hdd_info_mgr, event
from lib.Linux.EdgeSystem.event import Event
from lib.Linux.EdgeSystem.models import HddInfo
from monitor_db.init_structure import INIT_COLUMNS


class GetAllInfo(NamedTuple):
    expect: tuple[str, list] = ("", [])
    event_type: str = ""
    hisec: str = ""
    event_time: str = ""


class GetEventGenTime(NamedTuple):
    expect: str = ""
    path_valid: Result = Result(False)
    start_time: str = ""


class RenameRestartingFlag(NamedTuple):
    expect: int = 0
    valid: Result = Result(False)
    rename: Optional[Exception] = None


class GetHddRemovalEvent(NamedTuple):
    expect: bool = False
    cmd: Tuple[int, str] = (1, "")


class GetHddReplacementEvent(NamedTuple):
    expect: str = "sn"
    hdd_his: str = "NULL,NULL,NULL"
    hdd_cur: str = "sn"


class GetHddDevSn(NamedTuple):
    expect: str = "NULL,NULL,NULL"
    blocks: Iterable[Path] = []
    cmd: List[Tuple[int, str]] = []


@pytest.fixture(scope="module")
def database() -> DataBase:
    with NamedTemporaryFile(suffix=".db") as tmp_file:
        db_path = tmp_file.name
    type("TestMigrate", (Migrate,), {"models": {HddInfo.__tablename__: HddInfo}})\
        .execute_on_install(db_path, INIT_COLUMNS)
    test_db = type("TestDB", (DataBase,), {})(db_path)
    with patch.object(hdd_info_mgr, "session_maker", test_db.session_maker):
        yield test_db
    os.remove(db_path)


class TestEvent:
    use_cases = {
        "test_get_all_info": {
            "event type invalid": GetAllInfo(),
            "hisec": GetAllInfo(("", ["hi"]), "hisec", "hi"),
            "all": GetAllInfo(("time", []), "all", event_time="time")
        },
        "test_get_event_generation_time": {
            "path invalid": GetEventGenTime(),
            "invalid utc": GetEventGenTime(path_valid=Result(True)),
            "valid utc": GetEventGenTime("2024-02-21T05:40:07+00:00", Result(True), "1708494007")
        },
        "test_rename_restarting_flag": {
            "invalid path": RenameRestartingFlag(),
            "rename exception": RenameRestartingFlag(valid=Result(True), rename=Exception()),
            "success": RenameRestartingFlag(1, Result(True)),
        },
        "test_get_hdd_removal_event": {
            "exec cmd failed": GetHddRemovalEvent(),
            "invalid tec tem": GetHddRemovalEvent(cmd=(0, ":inv")),
            "valid tec tem": GetHddRemovalEvent(True, (0, ""))
        },
        "test_get_hdd_replacement_event": {
            "new hdd in": GetHddReplacementEvent(),
            "hdd not replace": GetHddReplacementEvent("his", "his", "his"),
            "hdd replace": GetHddReplacementEvent("new", "his", "new"),
        },
        "test_get_hdd_dev_sn": {
            "null blocks": GetHddDevSn(),
            "no sata": GetHddDevSn(blocks=(Path("/sys/block/abc"),)),
            "2:0:0:0 sn failed": GetHddDevSn(blocks=(Path("/sata/2:0:0:0"),), cmd=[(1, ""), ]),
            "2:0:0:0 sn": GetHddDevSn("NULL,NULL,sn", (Path("/sata/2:0:0:0"),), [(0, ":sn"), ]),
            "1:0:0:0 sn": GetHddDevSn("NULL,sn,NULL", (Path("/sata/1:0:0:0"),), [(0, ":sn"), ]),
            "0:0:0:0 sn": GetHddDevSn("sn,NULL,NULL", (Path("/sata/0:0:0:0"),), [(0, ":sn"), ]),
            "three sn": GetHddDevSn("sn0,sn1,sn2",
                                    (Path("/sata/1:0:0:0"), Path("/sata/0:0:0:0"), Path("/sata/2:0:0:0")),
                                    [(0, ":sn1"), (0, ":sn0"), (0, ":sn2")]),
        }
    }

    @staticmethod
    def test_clear_or_write_hdd_config(database: DataBase):
        session = database._scoped_session
        Event().clear_or_write_hdd_config("sn")
        assert session.query(HddInfo).first().serial_number == "sn"
        Event().clear_or_write_hdd_config()
        assert not session.query(HddInfo).count()

    @staticmethod
    def test_on_start_with_sn(database: DataBase):
        with database.session_maker() as session:
            session.add(HddInfo(serial_number="sn"))
        Event().on_start()
        with database.session_maker() as session:
            assert session.query(HddInfo).delete() == 1

    @staticmethod
    def test_on_start_without_sn(database: DataBase, mocker: MockerFixture):
        mocker.patch.object(Event, "get_hdd_dev_sn", return_value=None)
        Event().on_start()
        assert not database._scoped_session.query(HddInfo).count()

    @staticmethod
    def test_on_start_save_sn(database: DataBase, mocker: MockerFixture):
        mocker.patch.object(Event, "get_hdd_dev_sn", return_value="sn")
        Event().on_start()
        with database.session_maker() as session:
            assert session.query(HddInfo).delete() == 1

    @staticmethod
    def test_get_all_info(mocker: MockerFixture, model: GetAllInfo):
        if model.event_type == "hisec":
            mocker.patch.object(event, "hisec_event_message_que", queue.Queue(128))
            event.hisec_event_message_que.put(model.hisec)
        mocker.patch.object(Event, "get_event_generation_time", return_value=model.event_time)
        mocker.patch.object(Event, "get_hdd_removal_event")
        mocker.patch.object(Event, "get_hdd_replacement_event")
        mocker.patch.object(Event, "rename_restarting_flag")
        evt = Event()
        evt.get_all_info(model.event_type)
        assert model.expect == (evt.event_time, evt.result)

    @staticmethod
    def test_get_event_generation_time(mocker: MockerFixture, model: GetEventGenTime):
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid", return_value=model.path_valid)
        if not model.path_valid:
            assert model.expect == Event().get_event_generation_time()
            return

        with NamedTemporaryFile("w") as tmp_file:
            mocker.patch.object(Event, "RESTARTING_FLAG", tmp_file.name)
            tmp_file.write(model.start_time)
            tmp_file.flush()
            assert model.expect == Event().get_event_generation_time()

    @staticmethod
    def test_rename_restarting_flag(mocker: MockerFixture, model: RenameRestartingFlag):
        mocker.patch.object(FileCheck, "check_input_path_valid", return_value=model.valid)
        mocker.patch("os.rename", side_effect=model.rename)
        evt = Event()
        evt.rename_restarting_flag()
        assert model.expect == len(evt.result)

    @staticmethod
    def test_get_hdd_removal_event(mocker: MockerFixture, model: GetHddRemovalEvent):
        mocker.patch.object(ExecCmd, "exec_cmd_use_pipe_symbol", return_value=model.cmd)
        mocker.patch.object(Event, "clear_or_write_hdd_config")
        assert model.expect == Event().get_hdd_removal_event()

    @staticmethod
    def test_get_hdd_replacement_event(database: DataBase, mocker: MockerFixture, model: GetHddReplacementEvent):
        mocker.patch.object(Event, "get_hdd_dev_sn", return_value=model.hdd_cur)
        with database.session_maker() as session:
            session.add(HddInfo(serial_number=model.hdd_his))
        Event().get_hdd_replacement_event()
        with database.session_maker() as session:
            sn = session.query(HddInfo).first().serial_number
            session.query(HddInfo).delete()
        assert model.expect == sn

    @staticmethod
    def test_get_hdd_dev_sn(mocker: MockerFixture, model: GetHddDevSn):
        mocker.patch.object(Path, "glob", return_value=model.blocks)
        mocker.patch.object(ExecCmd, "exec_cmd_use_pipe_symbol", side_effect=model.cmd)
        assert model.expect == Event().get_hdd_dev_sn()
