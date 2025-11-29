from collections import namedtuple
from configparser import ConfigParser

import pytest
from pytest_mock import MockerFixture

from common.file_utils import FileCheck
from common.utils.result_base import Result
from ibma_redfish_globals import RedfishGlobals

ReturnInfoMessageCase = namedtuple("ReturnInfoMessageCase", "excepted, error_message, make_response")
UpdateJsonOfListCase = namedtuple("UpdateJsonOfListCase", "excepted, ret_dict")
MakeNetConfigResponseCase = namedtuple("MakeNetConfigResponseCase", "input_info")
MakeResponseOfConfigNetCase = namedtuple("MakeResponseOfConfigNetCase", "req_dict")
CheckJsonRequestCase = namedtuple("CheckJsonRequestCase", "excepted, request_data")
MakeEdgeTimeCase = namedtuple("MakeEdgeTimeCase", "cmd")


class TestRedfishGlobals:
    use_cases = {
        "test_return_info_message": {
            "status_null": ({"message": "Internal server error", "status": 100011}, {"status": None, "message": None},
                            [{"message": "Internal server error", "status": 100011}]),
            "not_list": ({"message": {"test": "test"}, "status": 500},
                         {"status": 500, "message": {"test": "test"}},
                         [{"message": {"test": "test"}, "status": 500}]),
            "exception": ({"message": {"test": "test"}, "status": 500},
                          {"status": 500, "message": {"test": "test"}},
                          [Exception(), {"message": {"test": "test"}, "status": 500}]),
            "list": ({"message": [500, "test"], "status": 500},
                     {"status": 500, "message": [500, "test"]},
                     [Exception(), {"message": [500, "test"], "status": 500}]),
            "exception2": ({"message": [500, "test"], "status": 500},
                           {"status": 500, "message": {"test": "test"}},
                           [Exception(), {"message": [500, "test"], "status": 500}])
        },
        "test_update_json_of_list": {
            "not_has": ({"message": {"test": "2"}}, {"message": {"test": "2"}})
        },
        "test_make_net_config_response": {
            "in": ("223",),
            "out": ("366",)
        },
        "test_make_response_of_config_net": {
            "with_out_message": ({"status": "test"},),
            "status_ok": ({"status": 200, "message": "test"},),
            "status_not_exist": ({"status": 404, "message": "test"},),
            "status_else": ({"status": 500, "message": "test"},)
        },
        "test_make_response_of_upgrade": {
            "with_out_message": ({"status": "test"},),
            "status_ok": ({"status": 200, "message": "test"},),
        },
        "test_make_reboot_failed_reason_response": {
            "223": ("223",),
            "229": ("229",),
        },
        "test_make_response_of_reset": {
            "with_out_message": ({"status": "test"},),
            "status_ok": ({"status": 200, "message": "test"},),
            "status_else": ({"status": 500, "message": "test"},)
        },
        "test_update_json_of_target": {
            "not_dict": ([],),
            "status_200": ({"status": 200, "message": {"Id": "1"}},),
            "status_not_200": ({"status": 300, "message": ["0", "test"]},),
            "not_has": ({"status": 300},),
        },
        "test_update_json_of_error_info": {
            "not_dict": ([],),
            "status_200": ({"status": 200, "message": ["0", "test"]},),
            "status_not_200": ({"status": 300, "message": "GeneralError", "errorKey": "key",
                                "errorValue": "value", "ParamTypes": "type", "NumberOfArgs": "arg"},),
            "not_has": ({"status": 300},),
        },
        "test_replace_resp_json_id": {
            "not_has": ({"status": 300},),
            "message_null": ({"status": 200, "message": ""},),
            "message_token_null": ({"status": 200, "message": {"result": {"Id": "test"}}},),
        },
        "test_update_update_model_json_of_list": {
            "not_dict": ([],),
            "not_has": ({"status": 200},),
            "status_200": ({"status": 200, "message": ""},),
            "status_not_200": ({"status": 300, "message": {"result": {"Id": "test"}}},),
        },
        "test_update_update_reset_model_json_of_list": {
            "not_dict": ([],),
            "not_has": ({"status": 200},),
            "status_200": ({"status": 200, "message": ""},),
            "status_not_200": ({"status": 300, "message": {"result": {"Id": "test"}}},),
        },
        "test_update_disk_json_of_list": {
            "not_dict": ([],),
            "not_has": ({"status": 200},),
            "status_200": ({"status": 200, "message": ""},),
            "status_not_200": ({"status": 300, "message": {"result": {"Id": "test"}}},),
        },
        "test_check_json_request": {
            "exception": (1, "test"),
            "not_dict": (1, b'{}'),
            "normal": (0, b'{"a": "test"}')
        },
        "test_make_edge_time": {
            "failed": ([-1, "failed"],),
            "success": ([0, "success"],)
        },
        "test_update_alarm_time_stamp_failed": {
            "not_dict": (None, []),
            "has_not_message": (None, {"status": 400}),
            "status_not_200": (None, {"status": 400, "message": "test"}),
            "success": ("1970-04-27 01:46:40", {"status": 200, "message": {"AlarMessages": [{"Timestamp": 10000000}]}})
        }
    }

    def test_init_http_server_param(self, mocker: MockerFixture):
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid", return_value=Result(False))
        mocker.patch("os.path.getsize", return_value=1024)
        mocker.patch.object(ConfigParser, "read", side_effect=Exception())
        mocker.patch.object(ConfigParser, "get",
                            result_value={"iBMA_user": "test1", "iBMA_http_server_ip": "test2", "iBMA_nic": "test3"})
        with pytest.raises(Exception):
            RedfishGlobals.init_http_server_param()

    def test_get_http_port(self):
        ret = RedfishGlobals.get_http_port()
        assert ret == 36665

    def test_set_http_port(self):
        RedfishGlobals.set_http_port(36665)
        assert RedfishGlobals.rfPort == 36665

    def test_get_http_nic_name(self):
        RedfishGlobals.rfNicName = "eth1"
        ret = RedfishGlobals.get_http_nic_name()
        assert ret == "eth1"

    def test_get_http_user(self):
        RedfishGlobals.rfUser = "MindXOM"
        RedfishGlobals.get_http_user()
        assert RedfishGlobals.rfUser == "MindXOM"

    def test_replace_kv(self):
        test_dict = {"test": "value"}
        RedfishGlobals.replace_kv(test_dict, "test", "case")
        assert test_dict.get("test") == "case"

    def test_replace_kv_list(self):
        test_dict = {"test": "value"}
        RedfishGlobals.replace_kv_list(test_dict, {"test": "case"}, )
        assert test_dict.get("test") == "case"

    def test_add_itmes(self):
        test_dict = {"test": ["value"]}
        RedfishGlobals.add_itmes(test_dict, "test", "case")
        assert test_dict.get("test") == ["value", "case"]

    def test_host_name_check(self):
        ret = RedfishGlobals.host_name_check("test")
        assert ret.group() == "test"

    def test_replace_id(self):
        resp_json = {"Id": "1", "@odata.id": "oDataID2"}
        RedfishGlobals.replace_id(resp_json, "test")
        assert resp_json["Id"] == "test"

    def test_replaceinfo(self):
        resp_json = {"Id": "1", "@odata.id": "oDataID2", "IPv4Addresses": ["1"],
                     "Oem": {"StartTime": 1, "TaskState": 1, "TaskPercentage": 1}}
        RedfishGlobals.replaceinfo(resp_json, "test", "2", {"IPv4Addresses": ["1"]}, "2")
        assert resp_json["Id"] == "test"
        assert resp_json["Oem"]["StartTime"] == "2"
        assert resp_json["@odata.id"] == "test2"

    def test_make_response(self, mocker: MockerFixture):
        mocker.patch("ibma_redfish_globals.make_response", return_value="test")
        ret = RedfishGlobals.make_response("test", "test")
        assert ret == "test"

    def test_make_update_file_response(self, mocker: MockerFixture):
        with mocker.patch("ibma_redfish_globals.jsonify"):
            ret = RedfishGlobals.make_update_file_response("test", "test")
        assert ret.status_code == "test"

    def test_make404_error_resp(self, mocker: MockerFixture):
        mocker.patch("ibma_redfish_globals.make_response", return_value="test")
        ret = RedfishGlobals.make404_error_resp()
        assert ret == "test"

    def test_replace_target_id(self):
        resp_json = {"Id": "1", "@odata.id": "oDataID1"}
        RedfishGlobals.replace_target_id(resp_json, "test")
        assert resp_json["@odata.id"] == "test1"

    def test_update_json_of_error_info(self, mocker: MockerFixture, model: MakeResponseOfConfigNetCase):
        mocker.patch.object(RedfishGlobals, "replace_kv_list")
        mocker.patch("ibma_redfish_globals.make_response", return_value="test")
        ret = RedfishGlobals.update_json_of_error_info(model.req_dict, "test")
        assert ret == "test"

    def test_check_json_request(self, model: CheckJsonRequestCase):
        ret = RedfishGlobals.check_json_request(model.request_data)
        assert model.excepted == ret[0]

    def test_check_input_parm(self):
        ret = RedfishGlobals.check_input_parm("text..")
        assert ret is False

    def test_update_alarm_time_stamp_failed(self, model: UpdateJsonOfListCase):
        ret = RedfishGlobals.update_alarm_time_stamp(model.ret_dict)
        if model.excepted:
            assert model.excepted == model.ret_dict["message"]["AlarMessages"][0]["Timestamp"]
        else:
            assert ret is None