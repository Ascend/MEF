from collections import namedtuple
from unittest.mock import mock_open

import pytest
from pytest_mock import MockerFixture

from common.file_utils import FileCheck
from common.utils.version_xml_file_manager import VersionXmlManager
from common.utils.result_base import Result
from conftest import TestBase

ParseXmlCase = namedtuple("ParseXmlCase", "exists, realpath, check_path, check_xml")


class TestVersionXmlManager(TestBase):
    CheckXmlCase = namedtuple("CheckXmlCase", "expected, get_size, read_data")

    use_cases = {
        "test_check_xml_is_safe": {
            "size_too_large": (False, 600 * 1024, ""),
            "key_wrong": (False, 300, "<!DOCTYPEabc"),
            "normal": (True, 300, "abc"),
        },
        "test_parse_xml_exception": {
            "exists_failed": (False, "", None, None),
            "start_with_check_path_failed": (True, "abc", Result(result=False), None),
            "check_safe_failed": (True, "/usr/local/mindx/AtlasEdge_Work_A",
                                  Result(result=True), Result(result=False)),
        },
        "test_parse_xml_normal": {
            "normal": (True, "/usr/local/mindx/AtlasEdge_Work_A", Result(result=True), Result(result=True)),
        },
    }

    @staticmethod
    def mock_param_xml_funcs(mocker, model):
        mocker.patch("os.path.exists", return_value=model.exists)
        mocker.patch("os.path.realpath", return_value=model.realpath)
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid", return_value=model.check_path)
        mocker.patch.object(VersionXmlManager, "_check_xml_is_safe", return_value=model.check_xml)
        mocker.patch("common.utils.version_xml_file_manager.xml_parser", return_value=None)

    def test_check_xml_is_safe(self, mocker: MockerFixture, model: CheckXmlCase):
        mocker.patch("os.path.getsize", return_value=model.get_size)
        mocker.patch("builtins.open", mock_open(read_data=model.read_data))
        ret = VersionXmlManager._check_xml_is_safe("/test")
        assert model.expected == bool(ret)

    def test_check_xml_is_safe_exception(self, mocker: MockerFixture):
        mocker.patch("os.path.getsize", return_value=300)
        mocker.patch("builtins.open", side_effect=ValueError())
        ret = VersionXmlManager._check_xml_is_safe("/test")
        assert not bool(ret)

    def test_parse_xml_exception(self, mocker: MockerFixture, model: ParseXmlCase):
        self.mock_param_xml_funcs(mocker, model)
        with pytest.raises(ValueError):
            VersionXmlManager()._parse_xml()

    def test_parse_xml_normal(self, mocker: MockerFixture, model: ParseXmlCase):
        self.mock_param_xml_funcs(mocker, model)
        ret = VersionXmlManager()._parse_xml()
        assert not ret
