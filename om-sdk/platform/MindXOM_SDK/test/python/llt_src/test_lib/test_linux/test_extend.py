# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
from pathlib import Path
from typing import NamedTuple, Optional, List

from pytest_mock import MockerFixture

from lib.Linux.systems.disk.device_loader import const
from lib.Linux.systems.extend import Extend
from common.common_methods import CommonMethods


class GetAllInfoModel(NamedTuple):
    expect: List[str]
    dev_name: Optional[str]
    start_from_m2: bool


class TestExtend:
    use_cases = {
        "test_get_all_info": {
            "get_extend_names": GetAllInfoModel(expect=["eth0", "eth1", "eMMC", "disk0"], dev_name=None,
                                                start_from_m2=False),
            "m2_get_extend_names": GetAllInfoModel(expect=["eth0", "eth1", "disk0"], dev_name=None, start_from_m2=True),
            "get_eMMC_detail": GetAllInfoModel(expect=[CommonMethods.OK, "get extend info success."], dev_name="eMMC",
                                               start_from_m2=False),
            "m2_get_eMMC_detail": GetAllInfoModel(expect=[CommonMethods.ERROR, "eMMC invalid."], dev_name="eMMC",
                                                  start_from_m2=True),
            "not_exists": GetAllInfoModel(expect=[CommonMethods.NOT_EXIST, "eMMC2 is fail"], dev_name="eMMC2",
                                          start_from_m2=False)
        }
    }

    @staticmethod
    def test_get_all_info(mocker: MockerFixture, model: GetAllInfoModel):
        mocker.patch("lib.Linux.systems.extend.Env").return_value.start_from_m2 = model.start_from_m2
        mocker.patch("lib.Linux.systems.disk.device_loader.SystemUtils").return_value.is_a500 = True
        const.FORMAT_INFO = Path(__file__).parent.joinpath("format_hw.info").as_posix()
        extend = Extend()
        if model.dev_name is None:
            extend.get_all_info(model.dev_name)
            assert extend.items == model.expect
        else:
            assert extend.get_all_info(model.dev_name) == model.expect
