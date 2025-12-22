from common.file_utils import FileCheck


class TestMidwareurlsRealarm:
    def check_path_should_ok_when_path_is_ok(self):
        ret = FileCheck.check_path_is_exist_and_valid("/run/restarting_flag")
        assert ret

    def check_path_should_failed_when_path_is_failed(self):
        ret = FileCheck.check_path_is_exist_and_valid("/run/rest")
        assert not ret
