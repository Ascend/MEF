# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2022-2023. All rights reserved.
import threading
from typing import Literal, NoReturn, Iterable, List, Tuple

from cert_manager.parse_tools import CertChainParser, CrlChainParser
from cert_manager.schemas import CertInfoSchema
from common.log.logger import run_log
from net_manager.checkers.contents_checker import CertInfoValidator
from net_manager.constants import NetManagerConstants
from net_manager.exception import DataCheckException, LockedError
from net_manager.manager.net_cfg_manager import CertMgr
from net_manager.manager.net_cfg_manager import NetCfgManager
from net_manager.models import CertManager, CertInfo


class ImportCert(CertMgr, CertInfoValidator):
    """根证书导入管理类"""
    lock = threading.Lock()

    def __init__(self, source: Literal["Web", "FusionDirector"] = NetManagerConstants.WEB,
                 cert_name: str = NetManagerConstants.FD_CERT_NAME):
        self.source = source
        self.cert_name = cert_name
        self._necessary()

    def check_cert_info(self, cert_info: CertInfoSchema) -> NoReturn:
        super().check_cert_info(cert_info)
        if not cert_info.ca_sign_valid:
            return
        # 新导入证书和已存在证书指纹比较，来源于FD的不允许导入相同指纹的证书
        if self.source == NetManagerConstants.FUSION_DIRECTOR and self._finger_already_existed(cert_info.fingerprint):
            raise DataCheckException("import cert finger already existed.")

    def import_deal(self, contents: str) -> dict:
        if self.lock.locked():
            raise LockedError("Import cert is busy.")
        with self.lock:
            cert_chain_parser = CertChainParser(contents)
            return self._add_cert(cert_chain_parser.get_root_ca_schema(self.check_cert_info), contents)

    def _finger_already_existed(self, finger: str) -> int:
        with self.session_maker() as session:
            return session.query(self.cert_info).filter_by(fingerprint=finger).count()

    def _necessary(self) -> NoReturn:
        if self.source == NetManagerConstants.WEB and NetCfgManager().get_net_cfg_info().status == "ready":
            raise DataCheckException("Current net manage status is 'ready', upload certificate is not allowed.")

    def _add_cert(self, cert_schema: CertInfoSchema, cert_contents: str) -> dict:
        with self.session_maker() as session:
            # 备份当前使用的证书
            self.backup_previous_cert()
            # 保存新导入的证书
            self.del_by_name(name=self.cert_name, session=session)
            cert_info = self.cert_info(name=self.cert_name)
            cert_info.update_by_schema(cert_schema)
            session.bulk_save_objects((
                self.cert(name=self.cert_name, source=self.source, cert_contents=cert_contents),
                cert_info,
            ))
            return cert_info.to_dict()


class ImportCrl(CertMgr):
    """导入crl处理类"""
    lock = threading.Lock()

    def __init__(self):
        super().__init__()
        self._necessary()

    def import_deal(self, contents: str) -> dict:
        if self.lock.locked():
            raise LockedError("Import crl is busy.")
        with self.lock:
            return self._import_deal(contents)

    def _import_deal(self, contents: str) -> dict:
        crl_chain_parser = CrlChainParser(contents)
        if not crl_chain_parser.node_num:
            raise DataCheckException("The number of crl is 0.")

        if crl_chain_parser.node_num != crl_chain_parser.buffer.count(crl_chain_parser.sep):
            raise DataCheckException("Duplicate chain in crl list.")

        if crl_chain_parser.node_num > crl_chain_parser.MAX_CHAIN_NUMS:
            raise DataCheckException(f"The number of crl is greater than {crl_chain_parser.MAX_CHAIN_NUMS}.")
        cert_names = list(self._verified_cert_names(crl_chain_parser))
        if not cert_names:
            raise DataCheckException("Verify CRL does not match against the CA certificate.")
        self._update_to_certs(cert_names, contents)
        return {"Message": "import crl success."}

    def _necessary(self) -> NoReturn:
        if not self._number_of_certs():
            raise DataCheckException("Check certificate is null. Please check and upload certificate.")

    def _certs_for_against_crl(self, chain_num: int) -> Iterable[Tuple[CertManager, CertInfo]]:
        with self.session_maker() as session:
            yield from self._certs_within_the_validity_period(session).filter(
                self.cert_info.chain_num == chain_num,
            ).limit(NetManagerConstants.CERT_FORM_FD_AND_WEB_LIMIT_NUM)

    def _verified_cert_names(self, parser: CrlChainParser) -> Iterable[str]:
        for cert, _ in self._certs_for_against_crl(parser.node_num):
            if not parser.verify_crl_chain_by_cert_chain(cert.cert_contents):
                run_log.warning("%s does not match CRL.", cert.name)
                continue
            yield cert.name

    def _update_to_certs(self, names: List[str], contents: str) -> NoReturn:
        with self.session_maker() as session:
            session.query(self.cert).filter(self.cert.name.in_(names)).update({"crl_contents": contents})
