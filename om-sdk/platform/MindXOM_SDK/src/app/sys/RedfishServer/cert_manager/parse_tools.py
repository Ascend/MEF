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
import ssl
from contextlib import AbstractContextManager, ExitStack
from datetime import datetime
from functools import cached_property
from tempfile import NamedTemporaryFile
from typing import Union, Iterable, Optional, List, Callable

from cert_manager.cert_clib_mgr import CertClibMgr
from cert_manager.schemas import CertInfoSchema, CrlInfoSchema
from common.constants.base_constants import CommonConstants
from common.constants.error_codes import SecurityServiceErrorCodes
from common.log.logger import run_log
from net_manager.exception import DataCheckException, InvalidCertInfo


class Parser(AbstractContextManager):
    """证书、吊销列表解析基类"""
    clib_mgr: CertClibMgr
    schema: Union[CertInfoSchema, CrlInfoSchema]

    def __init__(self, buffer: str, work_dir: str = CommonConstants.OM_WORK_DIR_PATH):
        if not buffer:
            raise ValueError("Buffer is null.")
        self.buffer = buffer
        self.tmp_file = NamedTemporaryFile("w", dir=CommonConstants.REDFISH_TMP_DIR)
        self.work_dir = work_dir

    def __enter__(self):
        self.tmp_file.__enter__()
        self._parse()
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        self.tmp_file.__exit__(exc_type, exc_val, exc_tb)

    def _parse(self):
        self.tmp_file.write(self.buffer)
        self.tmp_file.flush()
        self.clib_mgr = CertClibMgr(self.tmp_file.name, self.work_dir)


class CertParser(Parser):

    def overdue_validate(self, target_datetime: datetime) -> bool:
        return target_datetime >= self.schema.end_time

    def _parse(self):
        super()._parse()
        self.schema = self.clib_mgr.get_cert_info()


class CrlParser(Parser):

    def verify_crl_by_cert(self, cert: str) -> bool:
        with NamedTemporaryFile("w", dir=CommonConstants.REDFISH_TMP_DIR) as tmp_file:
            tmp_file.write(cert)
            tmp_file.flush()
            ret = self.clib_mgr.verify_cert_available(tmp_file.name)
            return ret == 0

    def validate(self, time_now: datetime):
        if time_now <= self.schema.last_update or time_now >= self.schema.next_update:
            raise DataCheckException("Crl checker: check last update time or next update time is error.")

    def overdue_validate(self, target_datetime: datetime) -> bool:
        return target_datetime >= self.schema.next_update

    def _parse(self):
        super()._parse()
        self.schema = self.clib_mgr.get_crl_info()


class ChainParser:
    MAX_CHAIN_NUMS = 10
    sep: str

    def __init__(self, buffer: str, work_dir: str = CommonConstants.OM_WORK_DIR_PATH):
        if not buffer:
            raise DataCheckException("Buffer is null.")
        self.buffer = buffer
        self.work_dir = work_dir

    @cached_property
    def node_num(self) -> int:
        return 0

    def node_generator(self) -> Iterable[str]:
        for node in self.buffer.split(self.sep):
            if not node.split():
                continue
            yield "".join((node, self.sep))

    def _ssl_ctx(self) -> ssl.SSLContext:
        ctx = ssl.SSLContext()
        try:
            with NamedTemporaryFile("w", dir=CommonConstants.REDFISH_TMP_DIR) as tmp_file:
                tmp_file.write(self.buffer)
                tmp_file.flush()
                ctx.load_verify_locations(tmp_file.name)
        except Exception as err:
            raise DataCheckException("load verify locations failed.") from err
        return ctx


class CertChainParser(ChainParser):
    sep = "-----END CERTIFICATE-----"

    @cached_property
    def node_num(self) -> int:
        try:
            ctx = self._ssl_ctx()
        except Exception as err:
            run_log.error("load cert failed, %s", err)
            raise InvalidCertInfo(SecurityServiceErrorCodes.ERROR_CERTIFICATE_IS_INVALID) from err
        stats = ctx.cert_store_stats()
        if stats.get("crl"):
            run_log.error("Certificate and revocation list concatenation file not supported.")
            raise InvalidCertInfo(SecurityServiceErrorCodes.ERROR_CERTIFICATE_IS_INVALID)
        return len(ctx.get_ca_certs())

    def cert_schema_generator(self):
        if self.node_num != self.buffer.count(self.sep):
            run_log.error("Duplicate chain in root certificate.")
            raise InvalidCertInfo(SecurityServiceErrorCodes.ERROR_CERTIFICATE_IS_INVALID)

        if self.node_num > self.MAX_CHAIN_NUMS:
            raise InvalidCertInfo(SecurityServiceErrorCodes.ERROR_CERTIFICATE_CHAIN_NUMS_MAX)

        root_ca = 0
        for cert in self.node_generator():
            with CertParser(cert, self.work_dir) as parser:
                root_ca = root_ca + 1 if parser.schema.ca_sign_valid else root_ca
            if root_ca > 1:
                run_log.error("The certificate chain contains multiple root certificates.")
                raise InvalidCertInfo(SecurityServiceErrorCodes.ERROR_CERTIFICATE_IS_INVALID)
            yield parser.schema

    def verify_cert_chain(self):
        """校验证书链是否合法"""
        with NamedTemporaryFile("w", dir=CommonConstants.REDFISH_TMP_DIR) as tmp_file:
            tmp_file.write(self.buffer)
            tmp_file.flush()
            if CertClibMgr(tmp_file.name, self.work_dir).cert_chain_verify() != 0:
                raise InvalidCertInfo(SecurityServiceErrorCodes.ERROR_CERTIFICATE_IS_INVALID)

    def get_root_ca_schema(self, cert_check_func: Callable[[CertInfoSchema], None]) -> CertInfoSchema:
        root_cert_schema: Optional[CertInfoSchema] = None
        end_time_list: List[datetime] = []
        for cert_schema in self.cert_schema_generator():
            cert_check_func(cert_schema)
            root_cert_schema = cert_schema if cert_schema.ca_sign_valid else root_cert_schema
            end_time_list.append(cert_schema.end_time)
        if not root_cert_schema:
            raise InvalidCertInfo(SecurityServiceErrorCodes.ERROR_CERTIFICATE_CA_SIGNATURE_INVALID)
        self.verify_cert_chain()
        cert_schema = root_cert_schema
        cert_schema.end_time = min(end_time_list)
        cert_schema.chain_num = self.node_num
        return cert_schema

    def verify_cert_chain_overdue(self, target_datetime: datetime) -> bool:
        """判断证书是否已过期"""
        # 采用ExitStack使每个证书对应的CertParser实例化一次并验证证书有效性
        with ExitStack() as stack:
            for cert in self.node_generator():
                parser = stack.enter_context(CertParser(cert, self.work_dir))
                if not parser.overdue_validate(target_datetime):
                    return False
        return True


class CrlChainParser(ChainParser):
    sep = "-----END X509 CRL-----"

    @cached_property
    def node_num(self) -> int:
        stats = self._ssl_ctx().cert_store_stats()
        if stats.get("x509"):
            raise DataCheckException("Certificate and revocation list concatenation file not supported.")
        return stats.get("crl") or 0

    def verify_crl_chain_by_cert_chain(self, cert_contents: str) -> bool:
        """判断吊销链与证书链是否匹配：证书链（校验过的）中的证书都能在吊销链中找到对应的吊销列表"""
        cert_chain = CertChainParser(cert_contents, self.work_dir)
        if self.node_num != cert_chain.node_num:
            return False

        # 采用ExitStack使每个crl对应的CrlParser实例化一次并与多个证书比较
        with ExitStack() as stack:
            time_now = datetime.utcnow()
            crl_chain: List[CrlParser] = []
            for crl in self.node_generator():
                parser = stack.enter_context(CrlParser(crl, self.work_dir))
                parser.validate(time_now)
                crl_chain.append(parser)
            return all(
                any(parser.verify_crl_by_cert(cert) for parser in crl_chain)
                for cert in cert_chain.node_generator()
            )

    def verify_crl_chain_overdue(self, target_datetime: datetime) -> bool:
        """判断吊销链是否已过期"""
        # 采用ExitStack使每个crl对应的CrlParser实例化一次并验证证书有效性
        with ExitStack() as stack:
            for crl in self.node_generator():
                parser = stack.enter_context(CrlParser(crl, self.work_dir))
                if not parser.overdue_validate(target_datetime):
                    return False
        return True
