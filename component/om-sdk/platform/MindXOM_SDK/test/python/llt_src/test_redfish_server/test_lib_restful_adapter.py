from collections import namedtuple
from pytest_mock import MockerFixture

from common.restfull_socket_model import RestFullSocketModel
from lib_restful_adapter import LibRESTfulAdapter
import ibma_client
from common.utils.app_common_method import AppCommonMethod

SendAndParseCase = namedtuple("SendAndParseCase", "excepted, send_msg, loads, literal_eval")
CheckStatusCase = namedtuple("CheckStatusCase", "excepted, ret_dict")


class TestLibRESTfulAdapter:
    use_cases = {
        "test_send_and_parse_result": {
            "send_msg_failed": ({"status": 400, "message": "wrong"}, [-1, {"status": 400, "message": "wrong"}], "", ""),
            "loads_success": ({"status": 200, "message": "success"}, [0, {"status": 200, "message": "success"}],
                              [{"status": 200, "message": "success"}], ""),
            "loads_exception": ("test", [0, b'test'], Exception(), ["test"]),
            "eval_exception": ("test", [0, b'test'], Exception(), Exception())
        },
        "test_check_status_is_ok": {
            "ok": (True, {"status": 200}),
            "fail": (False, {"status1": 200})
        }
    }

    def test_lib_restful_interface(self, mocker: MockerFixture):
        mocker.patch.object(LibRESTfulAdapter, "send_and_parse_result", return_value="success")
        ret = LibRESTfulAdapter.lib_restful_interface("test", "nic", None)
        assert ret == "success"

    def test_start_timer(self, mocker: MockerFixture):
        mocker.patch.object(LibRESTfulAdapter, "send_and_parse_result", return_value="success")
        ret = LibRESTfulAdapter.start_timer()
        assert ret == "success"

    def test_send_and_parse_result(self, mocker: MockerFixture, model: SendAndParseCase):
        module = RestFullSocketModel("lib_restful_interface", "test")
        mocker.patch.object(ibma_client.Client, "send_msg", return_value=model.send_msg)
        mocker.patch.object(AppCommonMethod, "get_json_error_by_array",
                            return_value={"status": 400, "message": "wrong"})
        mocker.patch("json.loads", side_effect=model.literal_eval)
        mocker.patch("ast.literal_eval", side_effect=model.literal_eval)
        ret = LibRESTfulAdapter.send_and_parse_result(module)
        assert model.excepted == ret

    def test_check_status_is_ok(self, model: CheckStatusCase):
        ret = LibRESTfulAdapter.check_status_is_ok(model.ret_dict)
        assert model.excepted == ret