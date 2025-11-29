# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import json
from collections import namedtuple
from unittest.mock import patch
from pytest_mock import MockerFixture
from flask.testing import FlaskClient


from user_manager.user_manager import SessionManager
from ut_utils.models import MockPrivilegeAuth
from test_bp_api.create_client import get_client

with patch("token_auth.get_privilege_auth", return_value=MockPrivilegeAuth):
    from session_service.session_blueprint import session_service_bp

GetSessionService = namedtuple("GetSessionService", "expect, return_value")
PatchSessionService = namedtuple("PatchSessionService", "expect, data")
DeleteSession = namedtuple("DeleteSession", "expect, request_obj, patch_return")


class TestSessionViews:
    client: FlaskClient = get_client(session_service_bp)
    use_cases = {
        "test_rf_get_session_service_collection": {
            "success": (
                {
                    "@odata.context": "/redfish/v1/$metadata#SessionService",
                    "@odata.id": "/redfish/v1/SessionService",
                    "@odata.type": "#SessionService.v1_1_8.SessionService",
                    "Id": "SessionService",
                    "Name": "Session Service",
                    "SessionTimeout": 300,
                    "Sessions": {
                        "@odata.id": "/redfish/v1/SessionService/Sessions"
                    }}, {"status": 200, "message": {"SessionTimeout": 300}})
        },
        "test_rf_patch_session_service_collection": {
            "test patch sessions failed due to data is not json!": (
                {"error": {"code": "Base.1.0.MalformedJSON",
                           "message": "A MalformedJSON has occurred. "
                                      "See ExtendedInfo for more information.",
                           "@Message.ExtendedInfo": [
                               {
                                   "@odata.type":
                                       "#MessageRegistry.v1_0_0.MessageRegistry",
                                   "Description": "Indicates that the request body was malformed JSON.  "
                                                  "Could be duplicate, syntax error,etc.",
                                   "Message": "The request body submitted was malformed JSON "
                                              "and could not be parsed by the receiving service.",
                                   "Severity": "Critical", "NumberOfArgs": 0, "ParamTypes": None,
                                   "Resolution": "Ensure that the request body is valid JSON and resubmit the request.",
                                   "Oem": {"status": None}}]
                           }
                 }, "test"
            ),
            "test patch sessions failed due to SessionTimeout is wrong!": (
                {"error": {"code": "Base.1.0.GeneralError",
                           "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                           "@Message.ExtendedInfo": [
                               {"@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Parameter is invalid.",
                                "Severity": "Critical", "NumberOfArgs": None, "ParamTypes": None,
                                "Resolution": "None",
                                "Oem": {"status": 100024}}]}}, {"SessionTimeout": 4, "Password": "test"}),
        },
        "test_rf_session_id": {
            "test delete specified session failed due to index is invalid": (
                {"error": {"code": "Base.1.0.GeneralError",
                           "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                           "@Message.ExtendedInfo": [
                               {"@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Parameter is invalid.",
                                "Severity": "Critical", "NumberOfArgs": None, "ParamTypes": None,
                                "Resolution": "None",
                                "Oem": {"status": 100024}}]}},
                {
                    "code": 400,
                    "index": '9ca9ac8d44d7353b11e985ef3d97d56abfbf0a1d5333b30'
                }, None
            ),
            "test delete specified session success": (
                {
                    "error": {
                        "code": "Base.1.0.Success",
                        "message": "Operation success. See ExtendedInfo for more information.",
                        "@Message.ExtendedInfo": [
                            {
                                "@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that no error has occurred.",
                                "Message": "Logout Success",
                                "Severity": "OK",
                                "NumberOfArgs": None,
                                "ParamTypes": None,
                                "Resolution": "None"
                            }
                        ]
                    }
                }, {
                    "code": 200,
                    "index": '9ca9ac8d44d7353b11e985ef3d97d56abfbf0a1d5333b300'
                }, [{"status": 200, "message": "test"}]
            ),
            "test delete specified session exception": (
                {"error": {"code": "Base.1.0.GeneralError",
                           "message": "A GeneralError has occurred. See ExtendedInfo for more information.",
                           "@Message.ExtendedInfo": [
                               {"@odata.type": "#MessageRegistry.v1_0_0.MessageRegistry",
                                "Description": "Indicates that a general error has occurred.",
                                "Message": "Logout failed.",
                                "Severity": "Critical", "NumberOfArgs": None, "ParamTypes": None,
                                "Resolution": "None",
                                "Oem": {"status": None}}]}},
                {
                    "code": 400,
                    "index": '9ca9ac8d44d7353b11e985ef3d97d56abfbf0a1d5333b300'
                }, Exception()
            )
        }
    }

    def test_rf_get_session_service_collection(self, mocker: MockerFixture, model: GetSessionService):
        mocker.patch.object(SessionManager, "get_all_info", return_value=model.return_value)
        response = self.client.get("/redfish/v1/SessionService")
        assert response.get_json(force=True) == model.expect

    def test_rf_patch_session_service_collection(self, model: PatchSessionService):
        response = self.client.patch("/redfish/v1/SessionService", data=json.dumps(model.data))
        assert response.get_json(force=True) == model.expect

    def test_rf_session_id(self, mocker: MockerFixture, model: DeleteSession):
        mocker.patch.object(SessionManager, "get_all_info", side_effect=model.patch_return)
        response = self.client.delete("/redfish/v1/SessionService/Sessions/{}".format(model.request_obj["index"]))
        assert response.get_json(force=True) == model.expect
