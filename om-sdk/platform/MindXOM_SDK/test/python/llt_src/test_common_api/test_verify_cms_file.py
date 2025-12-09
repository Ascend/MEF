from collections import namedtuple

from pytest_mock import MockerFixture

from common.verify_cms_file import check_parameter, verify_cms_file


class TestVerifyCmsFile:
    CheckParamCase = namedtuple("CheckParamCase", "excepted, test_agr, isfile")
    VerifyCmsFileCase = namedtuple("VerifyCmsFileCase", "excepted, prepare")

    use_cases = {
        "test_check_parameter": {
            "len_wrong": (False, ["test", "/home/test1", "/home/test2", "/home/test3"], [False, False, False, False]),
            "not_is_file_1": (False, ["test", "/home/test1", "/home/test2", "/home/test3", "/home/test4"],
                              [False, False, False, False]),
            "not_is_file_2": (False, ["test", "/home/test1", "/home/test2", "/home/test3", "/home/test4"],
                              [True, False, False, False]),
            "not_is_file_3": (False, ["test", "/home/test1", "/home/test2", "/home/test3", "/home/test4"],
                              [True, True, False, False]),
            "not_is_file_4": (False, ["test", "/home/test1", "/home/test2", "/home/test3", "/home/test4"],
                              [True, True, True, False]),
            "normal": (True, ["test", "/home/test1", "/home/test2", "/home/test3", "/home/test4"],
                       [True, True, True, True]),
        },
        "test_verify_cms_file": {
            "failed": (False, 1),
            "success": (True, 0)
        }

    }

    def test_check_parameter(self, mocker: MockerFixture, model: CheckParamCase):
        mocker.patch("os.path.isfile", side_effect=model.isfile)
        ret = check_parameter(model.test_agr)
        assert model.excepted == ret

    def test_verify_cms_file(self, mocker: MockerFixture, model: VerifyCmsFileCase):
        mocker.patch("ctypes.CDLL").return_value.prepareUpgradeImageCms.return_value = model.prepare
        ret = verify_cms_file("/test/1.so", "/home/test1", "/home/test2", "/home/test3")
        assert model.excepted == ret
