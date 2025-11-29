# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from itertools import islice
from pathlib import Path
from typing import Iterable

from common.constants.base_constants import CommonConstants
from common.kmc_lib.kmc_adapter import TableAdapter, FilesAdapter
from net_manager.models import NetManager
from redfish_db.session import session_maker


class NetMgrKmcAdapter(TableAdapter):
    session = session_maker
    model = NetManager
    filter_by = "id"
    cols = ("cloud_pwd",)


class ProfileKmcAdapter(FilesAdapter):
    """FD配置导入的Kmc密钥更新适配器"""

    def encrypted_files(self) -> Iterable[str]:
        for path in islice(Path(CommonConstants.REDFISH_KS_DIR).glob("*.prf"), CommonConstants.MAX_ITER_LIMIT):
            yield path.as_posix()


class ClientPsdAdapter(FilesAdapter):
    """客户端密码文件的Kmc密钥更新适配器"""

    def encrypted_files(self) -> Iterable[str]:
        yield Path(CommonConstants.UDS_CLIENT_KS_DIR).joinpath("client_kmc.psd").as_posix()


class FlaskPsdAdapter(FilesAdapter):
    def encrypted_files(self) -> Iterable[str]:
        yield Path(CommonConstants.FLASK_KS_DIR).joinpath("server_kmc.psd").as_posix()
