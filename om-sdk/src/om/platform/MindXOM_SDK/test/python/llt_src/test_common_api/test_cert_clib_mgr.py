# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
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
