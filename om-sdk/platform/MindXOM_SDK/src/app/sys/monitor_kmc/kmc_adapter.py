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

import os
from typing import Iterable

from common.constants.base_constants import CommonConstants
from common.kmc_lib.kmc_adapter import TableAdapter, FilesAdapter
from devm.devm_configs import DevmConfig
from monitor_db.session import session_maker


class NginxPsdAdapter(FilesAdapter):

    def encrypted_files(self) -> Iterable[str]:
        yield os.path.join(CommonConstants.NGINX_KS_DIR, "server_kmc.psd")


class UdsPsdAdapter(FilesAdapter):

    def encrypted_files(self) -> Iterable[str]:
        yield os.path.join(CommonConstants.UDS_KS_DIR, "server_kmc.psd")
        # uds client_kmc.psd也需要同步更新，如不更新，会导致mk重置且重启服务后，导致redfish进程解密client_kmc.psd失败
        yield os.path.join(CommonConstants.UDS_KS_DIR, "client_kmc.psd")


class DevmConfigAdapter(TableAdapter):
    model = DevmConfig
    session = session_maker
    filter_by = "filename"
    cols = ("cfg_sha256",)
