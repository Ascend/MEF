# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import json
from copy import deepcopy
from unittest.mock import patch

from flask.testing import FlaskClient
from pytest_mock import MockerFixture

from ibma_redfish_globals import RedfishGlobals
from lib_restful_adapter import LibRESTfulAdapter
from om_system_service.default_config import DefaultConfig
from user_manager.user_manager import UserManager
from ut_utils.create_client import get_client
from ut_utils.models import MockPrivilegeAuth


with patch("token_auth.get_privilege_auth", return_value=MockPrivilegeAuth):
    from om_system_service.om_actions_views import om_actions_service_bp


class TestOmActionViews:
    uri = "/redfish/v1/Systems/Actions/RestoreDefaults.Config"
    body = {"ReserveIP": True, "Password": "test10086"}
    headers = {"X-Real-Ip": "127.0.0.1", "X-Auth-Token": "abc"}
    client: FlaskClient = get_client(om_actions_service_bp)
    expect = {
        "error": {
            "code": "Base.1.0.GeneralError",
            "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
            "@Message.ExtendedInfo": [
                {
                    "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                    "Description": "Indicates that a general error has occurred.",
                    "Message": "",
                    "Severity": "Critical",
                    "NumberOfArgs": None,
                    "ParamTypes": None,
                    "Resolution": 'None',
                    "Oem": {
                        "status": 100028,
                    },
                },
            ]
        },
    }

    def test_rf_restore_default_configuration_is_locked(self, mocker: MockerFixture):
        mocker.patch.object(RedfishGlobals, "high_risk_exclusive_lock").locked().return_value = True
        resp = self.client.post(path=self.uri, headers=self.headers, data=json.dumps(self.body))
        tmp_expect = deepcopy(self.expect)
        tmp_expect.get("error").get("@Message.ExtendedInfo")[0]["Message"] = "The operation is busy."
        assert resp.get_json(force=True) == tmp_expect
        assert resp.get_json(force=True)["error"]["@Message.ExtendedInfo"][0]["Oem"]["status"] == 100028

    def test_rf_restore_default_configuration_get_exclusive_status_fail(self, mocker: MockerFixture):
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface").return_value = []
        tmp_expect = deepcopy(self.expect)
        extended_info = tmp_expect.get("error").get("@Message.ExtendedInfo")[0]
        extended_info["Message"] = "Internal server error"
        extended_info.get("Oem")["status"] = 100011
        resp = self.client.post(path=self.uri, headers=self.headers, data=json.dumps(self.body))
        assert resp.get_json(force=True) == tmp_expect

    def test_rf_restore_default_configuration_ret_msg_fail(self, mocker: MockerFixture):
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface").return_value = {
            "status": 200,
            "message": [],
        }
        tmp_expect = deepcopy(self.expect)
        extended_info = tmp_expect.get("error").get("@Message.ExtendedInfo")[0]
        extended_info["Message"] = "The operation is busy."
        resp = self.client.post(path=self.uri, headers=self.headers, data=json.dumps(self.body))
        assert resp.get_json(force=True) == tmp_expect

    def test_rf_restore_default_configuration_check_req_fail(self, mocker: MockerFixture):
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface").return_value = {
            "status": 200,
            "message": {
                "system_busy": False,
            },
        }
        tmp_expect = deepcopy(self.expect)
        err = "error"
        extended_info = tmp_expect.get(err).get("@Message.ExtendedInfo")[0]
        extended_info["Description"] = 'Indicates that the request body was malformed JSON.  Could be duplicate, ' \
                                       'syntax error,etc.'
        extended_info["Message"] = 'The request body submitted was malformed JSON and could not be parsed by ' \
                                   'the receiving service.'
        extended_info["NumberOfArgs"] = 0
        extended_info["Oem"] = {'status': None}
        extended_info["Resolution"] = "Ensure that the request body is valid JSON and resubmit the request."
        error_info = tmp_expect.get(err)
        error_info["code"] = "Base.1.0.MalformedJSON"
        error_info["message"] = 'A MalformedJSON has occurred. See ExtendedInfo for more information.'
        resp = self.client.post(path=self.uri, headers=self.headers, data=json.dumps(""))
        assert resp.get_json(force=True) == tmp_expect

    def test_rf_restore_default_configuration_check_parm_fail(self, mocker: MockerFixture):
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface").return_value = {
            "status": 200,
            "message": {
                "system_busy": False,
            },
        }
        body = {"ReserveIP": "", "Password": "test10086"}
        tmp_expect = deepcopy(self.expect)
        extended_info = tmp_expect.get("error").get("@Message.ExtendedInfo")[0]
        extended_info["Message"] = "Parameter is invalid."
        extended_info["Oem"] = {'status': 100024}
        resp = self.client.post(path=self.uri, headers=self.headers, data=json.dumps(body))
        assert resp.get_json(force=True) == tmp_expect

    def test_rf_restore_default_configuration_pwd_fail(self, mocker: MockerFixture):
        status = "status"
        msg = "message"
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface").return_value = {
            status: 200,
            msg: {
                "system_busy": False,
            },
        }
        mocker.patch.object(UserManager, "get_all_info").return_value = {
            status: 400,
            msg: [110206, "Secondary authentication failed."],
        }
        tmp_expect = deepcopy(self.expect)
        extended_info = tmp_expect.get("error").get("@Message.ExtendedInfo")[0]
        extended_info["Description"] = 'Indicates that a general error has occurred.'
        extended_info["Message"] = 'Secondary authentication failed.'
        extended_info["Oem"] = {'status': 110206}
        resp = self.client.post(path=self.uri, headers=self.headers, data=json.dumps(self.body))
        assert resp.get_json(force=True) == tmp_expect

    def test_rf_restore_default_configuration_del_req_fail(self, mocker: MockerFixture):
        status = "status"
        msg = "message"
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface").return_value = {
            status: 200,
            msg: {
                "system_busy": False,
            },
        }
        mocker.patch.object(UserManager, "get_all_info").return_value = {
            status: 200,
            "msg": "success",
        }
        mocker.patch.object(DefaultConfig, "deal_request").return_value = {
            status: 400,
            msg: "deal request fail.",
        }
        tmp_expect = deepcopy(self.expect)
        extended_info = tmp_expect.get("error").get("@Message.ExtendedInfo")[0]
        extended_info["Message"] = "deal request fail."
        extended_info["Oem"] = {'status': None}
        resp = self.client.post(path=self.uri, headers=self.headers, data=json.dumps(self.body))
        assert resp.get_json(force=True) == tmp_expect

    def test_rf_restore_default_configuration_success(self, mocker: MockerFixture):
        status = "status"
        msg = "message"
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface").return_value = {
            status: 200,
            msg: {
                "system_busy": False,
            },
        }
        mocker.patch.object(UserManager, "get_all_info").return_value = {
            status: 200,
            msg: "success",
        }
        mocker.patch.object(DefaultConfig, "deal_request").return_value = {
            status: 200,
            msg: "success",
        }
        tmp_expect = deepcopy(self.expect)
        err = "error"
        extended_info = tmp_expect.get(err).get("@Message.ExtendedInfo")[0]
        extended_info["Description"] = "Indicates that no error has occurred."
        extended_info["Message"] = "Restore defaults configuration successfully."
        extended_info["Severity"] = "OK"
        error_info = tmp_expect.get(err)
        error_info["code"] = "Base.1.0.Success"
        error_info["message"] = "Operation success. See ExtendedInfo for more information."
        del extended_info["Oem"]
        resp = self.client.post(path=self.uri, headers=self.headers, data=json.dumps(self.body))
        assert resp.get_json(force=True) == tmp_expect

    def test_rf_restore_default_configuration_exception(self, mocker: MockerFixture):
        status = "status"
        msg = "message"
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface").return_value = {
            status: 200,
            msg: {
                "system_busy": False,
            },
        }
        mocker.patch.object(UserManager, "get_all_info").return_value = {
            status: 200,
            msg: "success",
        }
        mocker.patch.object(DefaultConfig, "deal_request").return_value = {
            status: 200,
            msg: "success",
        }
        mocker.patch("ibma_redfish_serializer.SuccessMessageResourceSerializer.service.get_resource")\
            .return_value = "\ufeff6666"
        tmp_expect = deepcopy(self.expect)
        extended_info = tmp_expect.get("error").get("@Message.ExtendedInfo")[0]
        extended_info["Oem"] = {'status': None}
        extended_info["Message"] = "Restore defaults configuration failed."
        resp = self.client.post(path=self.uri, headers=self.headers, data=json.dumps(self.body))
        assert resp.get_json(force=True) == tmp_expect
