from collections import namedtuple

import mock
from pytest_mock import MockerFixture

from common.utils.app_common_method import AppCommonMethod
from ut_utils.mock_utils import mock_cdll

from bin.lib_adapter import LibAdapter

with mock_cdll():
    from om_lib.Linux.om_systems.dflc_info import DflcInfo
    from devm.device_mgr import DEVM
    from devm.exception import DeviceManagerError

PostRequest = namedtuple("PostRequest",
                         "expect, payload, locked, get_format_time, set_attribute")
GetStartPoint = namedtuple("GetStartPoint", "expect, get_attribute")
GetLifeSpan = namedtuple("GetLifeSpan", "expect, get_life_span")
GetManufactureDate = namedtuple("GetManufactureDate", "expect, get_manufacture_date")
GetItemId = namedtuple("GetItemId", "expect, get_item_id")
GetEdgeSystemInfo = namedtuple("GetEdgeSystemInfo", "expect, lib_restful_interface")


class TestDflcInfo:
    use_cases = {
        "test_get_item_id": {
            "normal": ("1", "1"),
        },
        "test_get_manufacture_date": {
            "invalid_when_get_sys_elabel_is_minus_one": ("2022-07-11", "2022-07-11"),
            "date_is_null": ("", ""),
        },
        "test_get_life_span": {
            "invalid_when_service_life_is_minus_one": (100, 100),
        },
        "test_get_start_point": {
            "get_start_time_exception": ("", DeviceManagerError()),
            "start_time_is_NA": ("2022-07-11", ["NA", "2022-07-11"]),
            "start_time_not_NA": ("2022-07-11", ["2022-07-11", "2022-07-11"]),
            "power_on_time_is_NA": ("", ["NA", "NA"]),
        },
        "test_post_request": {
            "normal": ([200, ''], {"StartPoint": "2022-07-11", "LifeSpan": 1}, False, "2022-07-11", [0, 0]),
            "invalid_when_locked": ([400, 'Config life_span is busy'], {"StartPoint": "2022-07-11", "LifeSpan": 1},
                                    True, "2022-07-11", [0, 0]),
            "invalid_when_life_span_is_not_int":
                ([400, 'invalid life_span'], {"StartPoint": "2022-07-11", "LifeSpan": ""}, False, "2022-07-11", [0, 0]),
            "invalid_when_life_span_smaller_than_zero":
                ([400, 'invalid life_span'], {"StartPoint": "2022-07-11", "LifeSpan": -1}, False, "2022-07-11", [0, 0]),
            "invalid_when_start_point_is_not_str":
                ([400, 'invalid start_point'], {"StartPoint": 1, "LifeSpan": 1}, False, "2022-07-11", [0, 0]),
            "invalid_when_len_start_point_larger_than_32":
                ([400, 'invalid start_point'], {"StartPoint": "2022-07-112022-07-112022-07-112022", "LifeSpan": 1},
                 False, "2022-07-11", [0, 0]),
            "invalid_when_get_format_time_is_null":
                ([400, 'invalid start_point'], {"StartPoint": "2022-07-11", "LifeSpan": 1}, False, "", [0, 0]),
            "set_start_time_exception":
                ([400, 'config start_point failed'], {"StartPoint": "2022-07-11", "LifeSpan": 1},
                 False, "2022-07-11", [DeviceManagerError(), 0]),
            "set_service_life_exception":
                ([400, 'config life_span failed'], {"StartPoint": "2022-07-11", "LifeSpan": 1},
                 False, "2022-07-11", [0, DeviceManagerError()]),
        },
        "test_get_edge_system_info": {
            "ok": (10, {
                "message": {"ProductName": "date", "SerialNumber": "point"},
                "status": 200
            }),
            "error": ("", {
                "message": {"ProductName": "date", "SerialNumber": "point"},
                "status": 400
            })

        }
    }

    @mock.patch.object(LibAdapter, 'lib_restful_interface', mock.Mock(
        return_value={"message": {"ManufactureDate": "mdate", "StartPoint": "spoint", "LifeSpan": 10, }, }))
    @mock.patch.object(AppCommonMethod, 'check_status_is_ok', mock.Mock(return_value=True))
    def test_get_edge_system_info_should_ok(self):
        ret = DflcInfo.get_edge_system_info(DflcInfo())
        assert not ret

    @mock.patch.object(LibAdapter, 'lib_restful_interface',
                       mock.Mock(return_value={"message": {"ManufactureDate": "mdate",
                                                           "StartPoint": "spoint", "LifeSpan": 10, }, }))
    @mock.patch.object(AppCommonMethod, 'check_status_is_ok', mock.Mock(return_value=False))
    def test_get_edge_system_info_should_failed(self):
        ret = DflcInfo.get_edge_system_info(DflcInfo())
        assert not ret

    def test_get_item_id(self, mocker: MockerFixture, model: GetItemId):
        mocker.patch.object(DEVM, "get_device").return_value.get_attribute.return_value = model.get_item_id
        dflc_info = DflcInfo()
        dflc_info.get_item_id()
        assert dflc_info.ItemId == model.expect

    def test_get_manufacture_date(self, mocker: MockerFixture, model: GetManufactureDate):
        mocker.patch.object(DEVM, "get_device").return_value.get_attribute.return_value = model.get_manufacture_date
        dflc_info = DflcInfo()
        dflc_info.get_manufacture_date()
        assert dflc_info.ManufactureDate == model.expect

    def test_get_life_span(self, mocker: MockerFixture, model: GetLifeSpan):
        mocker.patch.object(DEVM, "get_device").return_value.get_attribute.return_value = model.get_life_span
        dflc_info = DflcInfo()
        dflc_info.get_life_span()
        assert dflc_info.LifeSpan == model.expect

    def test_post_request(self, mocker: MockerFixture, model: PostRequest):
        mocker.patch.object(DflcInfo, "DFLC_LOCK").locked.return_value = model.locked
        mocker.patch.object(DflcInfo, "get_format_time", return_value=model.get_format_time)
        mocker.patch.object(DEVM, "get_device").return_value.set_attribute.side_effect = model.set_attribute
        assert DflcInfo.post_request(model.payload) == model.expect

    def test_get_start_point(self, mocker: MockerFixture, model: GetStartPoint):
        mocker.patch.object(DEVM, "get_device").return_value.get_attribute.side_effect = model.get_attribute

        dflc_info = DflcInfo()
        dflc_info.get_start_point()
        assert dflc_info.StartPoint == model.expect
