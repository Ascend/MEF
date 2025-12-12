# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from collections import namedtuple

import pytest
from pytest_mock import MockerFixture

from cert_manager.cert_clib_mgr import CertClibMgr
from common.constants.base_constants import CommonConstants

CheckDestPathCase = namedtuple("CheckDestPathCase", "expected, path")


class TestCertClibMgr:
    use_cases = {
        "test_check_dest_path": {
            "normal": ("/", "/"),
        },
    }

    def test_file_path_check_first_raise(self, mocker: MockerFixture):
        mocker.patch("os.path.exists", return_value=False)
        with pytest.raises(FileNotFoundError):
            CertClibMgr.file_path_check(CertClibMgr(CommonConstants.REDFISH_CERT_TMP_FILE))
