from collections import namedtuple

from pytest_mock import MockerFixture

from common.utils.result_base import Result
from upload_mark_file import UploadMarkFile
from common.utils.app_common_method import AppCommonMethod


class TestUploadMarkFile:
    CheckCreateUploadMarkFileCase = namedtuple("CheckCreateUploadMarkFileCase", "excepted, filename")
    CheckUploadCase = namedtuple("CheckUploadCase", "excepted, exists")
    CheckTimeoutCase = namedtuple("CheckTimeoutCase", "excepted, getmtime, cur_time")
    ClearCase = namedtuple("ClearCase", "exists, force_remove_file")

    use_cases = {
        "test_check_upload_mark_file_is_invalid_by_filename": {
            "cer": (True, "test.cer"),
            "crt": (True, "test.crt"),
            "zip": (True, "test.zip"),
            "conf": (True, "test.conf"),
            "ini": (True, "test.ini"),
            "other": (False, "test.txt")
        },
        "test_check_upload_mark_file_is_invalid": {
            "exist": (False, True),
            "not_exist": (True, False)
        },
        "test_check_upload_mark_file_is_timeout": {
            "timeout": (True, 1, 602),
            "not_timeout": (False, 1, 500)
        },
        "test_clear_upload_mark_file": {
            "not_exist": (False, None),
            "remove_failed": (True, Result(result=False)),
            "remove_success": (True, Result(result=True))
        },
        "test_create_upload_mark_file_by_filename": {
            "cer": (True, "test.cer"),
            "crt": (True, "test.crt"),
            "zip": (True, "test.zip"),
            "conf": (True, "test.conf"),
            "ini": (True, "test.ini"),
            "other": (False, "test.txt")
        },
    }

    def test_check_upload_mark_file_is_invalid_by_filename(self, mocker: MockerFixture,
                                                           model: CheckCreateUploadMarkFileCase):
        mocker.patch.object(UploadMarkFile, "check_upload_mark_file_is_invalid", return_value=True)
        ret = UploadMarkFile.check_upload_mark_file_is_invalid_by_filename(model.filename)
        assert model.excepted == ret

    def test_check_upload_mark_file_is_invalid(self, mocker: MockerFixture, model: CheckUploadCase):
        mocker.patch("os.path.exists", return_value=model.exists)
        mocker.patch.object(UploadMarkFile, "check_upload_mark_file_is_timeout", return_value=False)
        ret = UploadMarkFile.check_upload_mark_file_is_invalid("/home/test")
        assert model.excepted == ret

    def test_check_upload_mark_file_is_timeout(self, mocker: MockerFixture, model: CheckTimeoutCase):
        mocker.patch("os.path.getmtime", return_value=model.getmtime)
        mocker.patch("time.time", return_value=model.cur_time)
        mocker.patch.object(UploadMarkFile, "clear_upload_mark_file")
        ret = UploadMarkFile.check_upload_mark_file_is_timeout("/home/test")
        assert model.excepted == ret

    def test_clear_upload_mark_file(self, mocker: MockerFixture, model: ClearCase):
        mocker.patch("os.path.exists", return_value=model.exists)
        mocker.patch.object(AppCommonMethod, "force_remove_file", return_value=model.force_remove_file)
        ret = UploadMarkFile.clear_upload_mark_file("/home/test")
        assert ret is None

    def test_create_upload_mark_file_by_filename(self, mocker: MockerFixture, model: CheckCreateUploadMarkFileCase):
        mocker.patch.object(UploadMarkFile, "create_upload_mark_file", return_value=True)
        ret = UploadMarkFile.create_upload_mark_file_by_filename(model.filename)
        assert model.excepted == ret

    def test_create_upload_mark_file(self, mocker: MockerFixture):
        mocker.patch("os.path.exists", return_value=False)
        mocker.patch("os.mknod")
        mocker.patch("os.chmod")
        ret = UploadMarkFile.create_upload_mark_file("/home/test")
        assert ret is True

    def test_clear_upload_mark_file_all(self, mocker: MockerFixture):
        mocker.patch("os.path.exists", side_effect=[False, False, False, False, True, True])
        mocker.patch.object(AppCommonMethod, "force_remove_file", return_value=[False, True])
        ret = UploadMarkFile.clear_upload_mark_file_all()
        assert ret is None