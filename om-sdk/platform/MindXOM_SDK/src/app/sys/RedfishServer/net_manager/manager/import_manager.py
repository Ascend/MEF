# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
from abc import ABC, abstractmethod
from typing import NoReturn, List, Type, Iterable

from common.checkers import StringLengthChecker
from cert_manager.parse_tools import ParseCertInfo, ParseCrlInfo
from common.constants.error_codes import SecurityServiceErrorCodes
from net_manager.checkers.contents_checker import CertContentsChecker, CrlContentsChecker
from net_manager.constants import NetManagerConstants
from net_manager.exception import DataCheckException
from net_manager.manager.fd_cert_manager import FdCertManager
from net_manager.manager.net_cfg_manager import NetCertManager
from net_manager.manager.net_cfg_manager import NetCfgManager
from net_manager.models import CertManager


class ImportManagerBase(ABC):
    """导入管理基类"""

    checker_class: Type[StringLengthChecker]
    contents_name: str

    def __init__(self, contents: str, source: str):
        self.contents = contents
        self.source = source
        self.cert_manager = NetCertManager()

    @abstractmethod
    def import_deal(self) -> dict:
        pass

    @abstractmethod
    def update_contents(self) -> NoReturn:
        pass

    def check_contents(self) -> NoReturn:
        try:
            check_ret = self.checker_class(self.contents_name).check({self.contents_name: self.contents})
        except Exception as err:
            raise DataCheckException(f"{self.contents_name} check failed, {err}") from err

        if not check_ret.success:
            raise DataCheckException(f"{self.contents_name} check failed, {check_ret.reason}",
                                     err_code=check_ret.err_code)


class CertImportManager(ImportManagerBase):
    """根证书导入管理类"""

    checker_class = CertContentsChecker
    contents_name = "cert_contents"
    sep = "-----END CERTIFICATE-----"

    def __init__(self, contents: str, source=NetManagerConstants.WEB, cert_name=NetManagerConstants.FD_CERT_NAME):
        super().__init__(contents, source)
        self.cert_name = cert_name

    @staticmethod
    def check_status_is_ready() -> bool:
        return NetCfgManager().get_net_cfg_info().status == "ready"

    def check_contents(self):
        self.check_multi_cert()
        if self.source == NetManagerConstants.WEB:
            return

        # 新导入证书和已存在证书指纹比较，来源于FD的不允许导入相同指纹的证书
        with ParseCertInfo(self.contents) as cert_info:
            new_fingerprint = cert_info.fingerprint

        cert_managers: List[CertManager] = self.cert_manager.get_all()
        for cert in cert_managers:
            with ParseCertInfo(cert.cert_contents) as cert_info:
                old_fingerprint = cert_info.fingerprint

            if new_fingerprint == old_fingerprint:
                raise DataCheckException(f"import cert finger is the same with local cert name [{cert.name}].")

    def import_deal(self) -> dict:
        if self.source == NetManagerConstants.WEB and self.check_status_is_ready():
            raise DataCheckException("Current net manage status is 'ready', upload certificate is not allowed.")

        self.check_contents()
        self.update_contents()
        with ParseCertInfo(self.contents) as cert_info:
            return cert_info.to_dict()

    def update_contents(self) -> NoReturn:
        data = {
            "CertName": self.cert_name,
            "Source": self.source,
            "CertContents": self.contents,
        }
        # 根据证书名删除已存在名字的对象
        self.cert_manager.delete_obj_by_name(self.cert_name)
        # 插入新的数据对象
        self.cert_manager.add_objs(self.cert_manager.model.from_dict(data))

    def node_generator(self) -> Iterable[str]:
        for node in self.contents.split(self.sep):
            if not node.split():
                continue
            yield "".join((node, self.sep))

    def check_multi_cert(self):
        root_ca = 0
        for cert in self.node_generator():
            try:
                check_ret = self.checker_class(self.contents_name).check({self.contents_name: cert})
            except Exception as err:
                raise DataCheckException(f"{self.contents_name} check failed, {err}") from err

            if check_ret.success:
                root_ca += 1
            elif check_ret.err_code != SecurityServiceErrorCodes.ERROR_CERTIFICATE_CA_SIGNATURE_INVALID.code:
                raise DataCheckException(f"{self.contents_name} check failed, {check_ret.reason}",
                                         err_code=check_ret.err_code)

        if root_ca == 0:
            raise DataCheckException("The certificate chain has no root certificates.")
        elif root_ca > 1:
            raise DataCheckException("The certificate chain contains multiple root certificates.")


class CrlImportManager(ImportManagerBase):
    """吊销列表导入管理类"""
    checker_class = CrlContentsChecker
    contents_name = "crl_contents"

    def __init__(self, contents: str, source: str = NetManagerConstants.WEB):
        super().__init__(contents, source)
        self.cert_id = 0

    @staticmethod
    def compare_crl_issuing_date(old_last_update, new_last_update) -> NoReturn:
        if old_last_update >= new_last_update:
            err_msg = f"The last update time [{new_last_update}] of the uploaded CRL file is earlier " \
                      f"than or equal to the last update time [{old_last_update}] of the existing CRL file."
            raise DataCheckException(err_msg)

    def check_contents(self):
        super().check_contents()
        # 验证吊销列表是否匹配到对应根证书
        verify_result = FdCertManager().verify_crl_against_ca(self.contents, self.source)
        if not verify_result:
            raise DataCheckException(verify_result.error)

        self.cert_id = verify_result.data
        # 检查该根证书是否已上传吊销列表，若已上传则和新上传的时间作对比
        crl_contents = self.cert_manager.get_obj_by_id(self.cert_id).crl_contents
        if not crl_contents:
            return

        with ParseCrlInfo(crl_contents) as parse_crl:
            old_last_update = parse_crl.last_update

        with ParseCrlInfo(self.contents) as parse_crl:
            new_last_update = parse_crl.last_update

        self.compare_crl_issuing_date(old_last_update, new_last_update)

    def import_deal(self) -> dict:
        self.check_contents()
        self.update_contents()
        return {"Message": "import crl success."}

    def update_contents(self) -> NoReturn:
        # 根据其CA证书数据对象id插入crl_contents
        self.cert_manager.update_obj_by_id(self.cert_id, {"crl_contents": self.contents})
