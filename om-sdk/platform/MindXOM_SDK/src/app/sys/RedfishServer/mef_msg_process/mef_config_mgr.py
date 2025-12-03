# coding: utf-8
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import os.path
import ssl
from dataclasses import dataclass
from typing import Iterable

from common.constants.base_constants import CommonConstants
from common.file_utils import FileCheck
from common.log.logger import run_log
from common.schema import BaseModel
from common.schema import field
from common.kmc_lib.kmc import Kmc
from common.kmc_lib.tlsconfig import TlsConfig
from common.utils.result_base import Result
from net_manager.checkers.contents_checker import CertContentsChecker

# 证书路径配置文件路径
PATH_PREFIX = "/usr/local/mindx/MindXOM/software/RedfishServer/cert"


@dataclass
class MefConfigData(BaseModel):
    """MEF配置数据"""
    host: str = field(default="127.0.0.1")
    port: str = field(default="10010")
    wss_uri: str = field(init=False)
    root_crt_path: str = field(default=os.path.join(PATH_PREFIX, "root_ca_mef.cert"), comment="MEF根证书")
    client_crt_path: str = field(default=os.path.join(PATH_PREFIX, "client_kmc.cert"))
    client_pri_path: str = field(default=os.path.join(PATH_PREFIX, "client_kmc.priv"))
    client_pwd_path: str = field(default=os.path.join(PATH_PREFIX, "client_kmc.psd"))
    ks_path: str = field(default=os.path.join(PATH_PREFIX, "om_cert.keystore"))
    ks_backup_path: str = field(default=os.path.join(PATH_PREFIX, "om_cert_backup.keystore"))
    alg_json_path: str = field(default=os.path.join(PATH_PREFIX, "om_alg.json"))

    def __post_init__(self):
        self.wss_uri = f"wss://{self.host}:{self.port}/device_om"

    def gen_client_ssl_context(self) -> Result:
        """生成连接MEF服务端的安全SSLContext"""
        for file in self._gen_necessary_files():
            res = FileCheck.check_path_is_exist_and_valid(file)
            if not res:
                return Result(False, err_msg=f"{file} path invalid : {res.error}")

        # 校验MEF的根证书内容合法性
        ret = self._check_mef_root_ca_valid()
        if not ret:
            run_log.error("MEF root ca cert %s invalid: %s.", self.root_crt_path, ret.error)
            return ret

        kmc = Kmc(self.ks_path, self.ks_backup_path, self.alg_json_path)
        with open(self.client_pwd_path, "r") as psd:
            encrypted = psd.read()
        decrypted = kmc.decrypt(encrypted)
        is_normal_value, ctx = TlsConfig.get_ssl_context(self.root_crt_path, self.client_crt_path, self.client_pri_path,
                                                         pwd=decrypted)
        if not is_normal_value:
            err_info = "get client ssl context error."
            run_log.error(err_info)
            return Result(False, err_msg=err_info)
        # 需要校验对端证书
        ctx.verify_mode = ssl.CERT_REQUIRED
        decrypted = None
        return Result(result=True, data=ctx)

    def _gen_necessary_files(self) -> Iterable[str]:
        yield from (
            self.root_crt_path, self.client_crt_path, self.client_pri_path, self.client_pwd_path,
            self.ks_path, self.ks_backup_path, self.alg_json_path
        )

    def _check_mef_root_ca_valid(self) -> Result:
        if not os.path.exists(self.root_crt_path):
            return Result(False, err_msg="not exist")

        with open(self.root_crt_path) as file:
            file_info = os.stat(file.fileno())
            # 校验文件是否是软连接
            if os.path.islink(self.root_crt_path):
                return Result(False, err_msg="is link")

            if file_info.st_size > CommonConstants.MAX_CERT_LIMIT:
                return Result(False, err_msg="content is too large")

            if file_info.st_uid != os.geteuid():
                return Result(False, err_msg="owner is not right")

            data = {"cert": file.read(CommonConstants.MAX_CERT_LIMIT)}
            ret = CertContentsChecker("cert").check_dict(data)
            if not ret:
                return Result(False, err_msg="invalid cert content")

            return Result(True)
