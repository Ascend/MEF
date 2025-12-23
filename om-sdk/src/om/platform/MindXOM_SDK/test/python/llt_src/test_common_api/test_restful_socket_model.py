from collections import namedtuple

from common.restfull_socket_model import RestFullSocketModel


class TestRestFullSocketModel:
    CheckSocketModelCase = namedtuple("CheckSocketModelCase", "excepted, json_data")

    use_cases = {
        "test_check_socket_model": {
            "not_contain_method": (False, {"test": 1}),
            "contain_method": (True, {"method": 1})
        }
    }

    def test_check_socket_model(self, model: CheckSocketModelCase):
        ret = RestFullSocketModel.check_socket_model(model.json_data)
        assert model.excepted == ret

    def test_get_socket_model(self):
        ret = RestFullSocketModel.get_socket_model({"method": "test"})
        assert ret.method == "test"

    def test_get_socket_info(self):
        model = RestFullSocketModel("method1", "model_name", "request_type", "request_data",
                                    "need_list", "item1", "item2", "item3", "item4")
        ret = model.get_socket_info()
        assert "method1" in ret
