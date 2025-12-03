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

from pytest_mock import MockerFixture

from common.utils.app_common_method import AppCommonMethod
from fd_msg_process.schemas import Firmware


data = {'Module': ('Module', ), 'Version': ('Version', )}


class TestUtils:
    @staticmethod
    def some_function():
        return "some result"


class TestGetJsonInfoObj:
    def test_get_socket_info(self, mocker: MockerFixture):
        mocker.patch.object(AppCommonMethod, "get_json_info", side_effect={"header": TestUtils.some_function(),
                                                                           "route": ""})
        assert Firmware.firmware_list([data, ]) == [{'name': ('Module',), 'version': ('Version',),
                                                     'inactive_version': '', 'active_method': 'inband',
                                                     'board_id': '', 'upgrade_agent': 'OM'}]

    def test_to_dict(self):
        f_f = Firmware('Module', 'Version')
        assert f_f.to_dict() == {'name': 'Module', 'version': 'Version', 'inactive_version': '',
                               'active_method': 'inband', 'board_id': '', 'upgrade_agent': 'OM'}
