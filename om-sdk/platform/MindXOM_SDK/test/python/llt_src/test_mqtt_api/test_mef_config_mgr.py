# -*- coding: utf-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

from collections import namedtuple
from unittest.mock import MagicMock
from unittest.mock import mock_open

from pytest_mock import MockerFixture

from common.utils.result_base import Result
from mef_msg_process.mef_config_mgr import MefConfigData
from net_manager.checkers.contents_checker import CertContentsChecker

data = {'Module': ('Module', ), 'Version': ('Version', )}
mock_stat_result = MagicMock()
StatTuple = namedtuple('StatTuple',
                       ['st_mode', 'st_ino', 'st_dev', 'st_nlink', 'st_uid', 'st_gid', 'st_size', 'st_atime',
                        'st_mtime', 'st_ctime'])


class TestGetJsonInfoObj:
    CheckMefRootCaValidCase = namedtuple("CheckMefRootCaValidCase",
                                         "expected, stat_sequence, islink, exists, check_dict")

    use_cases = {
        "test_check_mef_root_ca_valid": {
            "first": (Result(False), StatTuple(0, 0, 0, 0, 0, 0, 20 * 1024 + 1, 0, 0, 0), True, False, False),
            "second": (Result(False), StatTuple(0, 0, 0, 0, 0, 0, 20 * 1024 + 1, 0, 0, 0), True, True, False),
            "third": (Result(False), StatTuple(0, 0, 0, 0, 0, 0, 20 * 1024 + 1, 0, 0, 0), False, True, False),
            "forth": (Result(False), StatTuple(0, 0, 0, 0, 1, 0, 20 * 1024, 0, 0, 0), False, True, False),
            "fifth": (Result(False), StatTuple(0, 0, 0, 0, 0, 0, 20 * 1024, 0, 0, 0), False, True, False),
            "sixth": (Result(True), StatTuple(0, 0, 0, 0, 0, 0, 20 * 1024, 0, 0, 0), False, True, True),
        },
    }

    def test_gen_necessary_files(self):
        assert len(list(MefConfigData._gen_necessary_files(MefConfigData))) == 7

    def test_check_mef_root_ca_valid(self, mocker: MockerFixture, model: CheckMefRootCaValidCase):
        stat_sequence = model.stat_sequence
        mock_stat_result.st_size = stat_sequence.st_size
        mock_stat_result.st_uid = stat_sequence.st_uid
        CopyFile1 = namedtuple("CopyFile1", ["root_crt_path", ])
        backup_restore_base_two = CopyFile1(*(r"/usr/local/mindx/MindXOM/software/RedfishServer/cert/root_ca_mef.cert",
                                              ))
        mocker.patch("os.path.exists", return_value=True)
        mocker.patch("os.path.islink", return_value=model.islink)
        mocker.patch("os.stat", return_value=mock_stat_result)
        mocker.patch("builtins.open", mock_open(read_data="test"))
        mocker.patch.object(CertContentsChecker, "check_dict", return_value=model.check_dict)
        ret = MefConfigData._check_mef_root_ca_valid(backup_restore_base_two)
        assert bool(ret) == bool(model.expected)
