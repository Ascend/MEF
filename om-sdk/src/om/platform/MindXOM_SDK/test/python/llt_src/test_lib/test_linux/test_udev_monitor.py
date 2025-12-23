# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
from functools import partial
from typing import NamedTuple, Dict, Callable

import pytest
from pytest_mock import MockerFixture

from lib.Linux.systems.disk import udev_monitor
from lib.Linux.systems.disk.errors import DevError
from lib.Linux.systems.disk.schema import UdevMonitorInfo
from lib.Linux.systems.disk.udev_monitor import terminate_by_gpt, terminate_while_add_part, terminate_by_path

ONE_UDEV = (
    b"UDEV  [11856.245467] change",
    b"ACTION=change",
    b"DEVPATH=/devices/platform/a6000000.sata/ata2/host1/target1:0:0/1:0:0:0/block/sda/sda9",
    b"DEVNAME=/dev/sda9",
    b"ID_PART_TABLE_TYPE=gpt",
    b"\n"
)

ADD_PART_INFO = (
    b"UDEV  [11856.245467] add",
    b"ACTION=add",
    b"DEVPATH=/devices/platform/a6000000.sata/ata2/host1/target1:0:0/1:0:0:0/block/sda/sda9",
    b"DEVNAME=/dev/sda9",
    b"\n"
)


class DevPropertyDictModel(NamedTuple):
    expect: Dict
    reader: Callable[[None], bytes]


class TerminateMonitorModel(NamedTuple):
    expect: bytes
    reader: Callable[[None], bytes]
    terminate: Callable[[UdevMonitorInfo], bool]


class MockDisk(NamedTuple):
    name: str = ""
    pt_type: str = ""
    path: str = ""


class TestUdevMonitor:
    use_cases = {
        "test_dev_property_dict": {
            "normal_udev_info": DevPropertyDictModel(
                expect={
                    "ACTION": "change",
                    "DEVPATH": "/devices/platform/a6000000.sata/ata2/host1/target1:0:0/1:0:0:0/block/sda/sda9",
                    "DEVNAME": "/dev/sda9",
                    "ID_PART_TABLE_TYPE": "gpt"
                },
                reader=partial(next, iter(ONE_UDEV))
            ),
            "not udev": DevPropertyDictModel(expect={}, reader=partial(next, iter((b"a", b"b")))),
        },
        "test_terminate_monitor": {
            "terminate_by_gpt": TerminateMonitorModel(
                expect=b"gpt",
                reader=partial(next, iter(ONE_UDEV + (b"gpt",) + ONE_UDEV)),
                terminate=partial(terminate_by_gpt, disk=MockDisk(path="/dev/sda9", pt_type="gpt", name="sda"))),
            "terminate_while_add_part": TerminateMonitorModel(
                expect=b"add",
                reader=partial(next, iter(ONE_UDEV + ADD_PART_INFO + (b"add",) + ONE_UDEV)),
                terminate=partial(terminate_while_add_part, disk=MockDisk(name="sda", ))
            ),
            "terminate_by_path": TerminateMonitorModel(
                expect=b"path",
                reader=partial(next, iter(ONE_UDEV + (b"path",) + ADD_PART_INFO)),
                terminate=partial(terminate_by_path, path="/dev/sda9")
            )
        }
    }

    @staticmethod
    def test_dev_property_dict(mocker: MockerFixture, model: DevPropertyDictModel):
        proc = mocker.MagicMock()
        proc.stdout.readline = model.reader
        assert udev_monitor.dev_property_dict(proc) == model.expect

    @staticmethod
    def test_dev_property_dict_time_out(mocker: MockerFixture):
        proc = mocker.MagicMock()
        proc.stdout.readline = partial(next, iter((b"",)))
        with pytest.raises(DevError):
            udev_monitor.dev_property_dict(proc)

    @staticmethod
    def test_terminate_monitor(mocker: MockerFixture, model: TerminateMonitorModel):
        proc = mocker.MagicMock()
        proc.stdout.readline = model.reader
        udev_monitor.terminate_monitor(proc, model.terminate)
        assert proc.stdout.readline() == model.expect
