import mock
from fd_msg_process.common_redfish import CommonRedfish


class TestCommonRedfish:
    FAILED = False
    SUCCESS = True

    @mock.patch.object(CommonRedfish, 'ret_dict_is_exception_proc', mock.Mock(return_value="dict_none"))
    def test_update_json_of_list_should_return_none_when_dict_is_none(self):
        ret = CommonRedfish.update_json_of_list("", None, "error_info")
        assert ret == "dict_none"

    @mock.patch.object(CommonRedfish, 'ret_dict_is_exception_proc', mock.Mock(return_value="key_not_exists"))
    def test_update_json_of_list_should_return_error_when_dict_key_not_exists(self):
        ret = CommonRedfish.update_json_of_list("", {'status': 100}, "error_info")
        assert ret == "key_not_exists"

    @mock.patch.object(CommonRedfish, 'ret_dict_is_error_proc', mock.Mock(return_value="status_wrong"))
    def test_update_json_of_list_should_return_error_when_dict_status_wrong(self):
        ret = CommonRedfish.update_json_of_list("", {"status": 100, "message": "mess"}, "error_info")
        assert ret == "status_wrong"

    @mock.patch.object(CommonRedfish, 'replace_kv_list', mock.Mock(return_value=True))
    def test_update_json_of_list_should_return_ok(self):
        ret = CommonRedfish.update_json_of_list({"status": 100, "message": "ms"}, {"status": 200, "message": "mess"},
                                                "error_info")
        assert ret == '{"status": 100, "message": "ms"}'

    @mock.patch.object(CommonRedfish, 'ret_dict_is_exception_proc', mock.Mock(return_value="dict_none"))
    def test_update_json_of_mem_count_and_odata_id_should_return_none_when_dict_is_none(self):
        ret = CommonRedfish.update_json_of_mem_count_and_odata_id("", None, "error_info")
        assert ret == "dict_none"

    @mock.patch.object(CommonRedfish, 'ret_dict_is_exception_proc', mock.Mock(return_value="key_not_exists"))
    def test_update_json_of_mem_count_and_odata_id_should_return_error_when_dict_key_not_exists(self):
        ret = CommonRedfish.update_json_of_mem_count_and_odata_id("", {'status': 100}, "error_info")
        assert ret == "key_not_exists"

    @mock.patch.object(CommonRedfish, 'ret_dict_is_error_proc', mock.Mock(return_value="status_wrong"))
    def test_update_json_of_mem_count_and_odata_id_should_return_error_when_dict_status_wrong(self):
        ret = CommonRedfish.update_json_of_mem_count_and_odata_id("", {"status": 100, "message": "mess"}, "error_info")
        assert ret == "status_wrong"

    @mock.patch.object(CommonRedfish, 'replace_kv', mock.Mock(return_value=True))
    def test_update_json_of_mem_count_and_odata_id_should_return_json_when_message_len_is_0(self):
        ret = CommonRedfish.update_json_of_mem_count_and_odata_id(
            {"status": 100, "message": "mess", "Members": [{'@odata.id': 1}]}, {"status": 200, "message": ""},
            "error_info")
        assert ret == '{"status": 100, "message": "mess", "Members": [{}]}'

    @mock.patch.object(CommonRedfish, 'replace_kv', mock.Mock(return_value=True))
    def test_update_json_of_mem_count_and_odata_id_should_return_json(self):
        ret = CommonRedfish.update_json_of_mem_count_and_odata_id(
            {"status": 100, "message": ['1', '2'], "Members": [{'@odata.id': "entityID"}]},
            {"status": 200, "message": ['1', '2']},
            "error_info")
        assert ret == '{"status": 100, "message": ["1", "2"], "Members": [{"@odata.id": "1"}, {"@odata.id": "2"}]}'

    def test_is_legal_of_input_param_should_failed_when_param_not_illegal(self):
        ret = CommonRedfish.is_legal_of_input_param("...")
        assert ret == self.FAILED

    def test_is_legal_of_input_param_should_ok_when_param_is_true(self):
        ret = CommonRedfish.is_legal_of_input_param("abs")
        assert ret == self.SUCCESS

    def test_is_legal_of_input_param_should_ok_when_param_contains_illegal_char(self):
        ret = CommonRedfish.is_legal_of_input_param("abs-+")
        assert ret == self.FAILED
