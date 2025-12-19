# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
from collections import namedtuple

from pytest_mock import MockerFixture

from common.utils.exec_cmd import ExecCmd
from lib.Linux.systems import time_zone_mgr
from modify_time_zone import TimeZone

ReadTimeZoneCfgCase = namedtuple("ReadTimeZoneCfgCase", "expected, get_time_zone_offset")
SetTimeZoneCase = namedtuple("SetTimeZoneCase", "expected, read_time_zone_cfg, exec_cmd_get_output")


class TestModifyTimeZone:
    use_cases = {
        "test_read_time_zone_cfg": {
            "normal": ("", (None,)),
        },
        "test_set_time_zone": {
            "normal": (True, True, [0, ]),
            "time_zone_is_false": (True, False, [0, ]),
            "ret[0] != 0": (False, True, [1, ])
        },
    }

    def test_read_time_zone_cfg(self, mocker: MockerFixture, model: ReadTimeZoneCfgCase):
        mocker.patch.object(time_zone_mgr, "get_time_zone_offset", side_effect=model.get_time_zone_offset)
        ret = TimeZone.read_time_zone_cfg()
        assert model.expected == ret

    def test_set_time_zone(self, mocker: MockerFixture, model: SetTimeZoneCase):
        mocker.patch.object(TimeZone, "read_time_zone_cfg", return_value=model.read_time_zone_cfg)
        mocker.patch.object(ExecCmd, "exec_cmd_get_output", return_value=model.exec_cmd_get_output)
        ret = TimeZone.set_time_zone(TimeZone())
        assert model.expected == ret
