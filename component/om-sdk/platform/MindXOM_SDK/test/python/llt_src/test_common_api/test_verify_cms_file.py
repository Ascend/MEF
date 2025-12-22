from collections import namedtuple

from pytest_mock import MockerFixture

from common.verify_cms_file import check_parameter


class TestVerifyCmsFile:
    CheckParamCase = namedtuple("CheckParamCase", "excepted, test_agr, isfile")

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
        }

    }

    def test_check_parameter(self, mocker: MockerFixture, model: CheckParamCase):
        mocker.patch("os.path.isfile", side_effect=model.isfile)
        ret = check_parameter(model.test_agr)
        assert model.excepted == ret
