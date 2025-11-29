# -*- coding: UTF-8 -*-
# Copyright (C) 2023.Huawei Technologies Co., Ltd. All rights reserved.
import contextlib
import ssl
from tempfile import NamedTemporaryFile
from typing import Iterable, NoReturn
from typing import Tuple
from typing import Union

from urllib3.util.connection import create_connection

from common.constants.base_constants import CommonConstants
from common.kmc_lib.tlsconfig import TlsConfig
from common.log.logger import run_log
from net_manager.constants import NetManagerConstants
from net_manager.exception import NetManagerException
from net_manager.manager.net_cfg_manager import CertMgr
from net_manager.models import CertManager, CertInfo


class FdCertManager(CertMgr):
    SOCKET_CONNECT_TIMEOUT: int = 5

    def __init__(self, server_name: str = "", port: Union[int, str] = ""):
        self.addr: Tuple[str, int] = (
            server_name or NetManagerConstants.SERVER_NAME,
            int(port or NetManagerConstants.PORT),
        )

    @staticmethod
    def _load_crl(cert: CertManager, ctx: ssl.SSLContext) -> NoReturn:
        if not cert.crl_contents:
            return

        with NamedTemporaryFile("w", dir=CommonConstants.REDFISH_TMP_DIR) as tmp_file:
            tmp_file.write(cert.crl_contents)
            tmp_file.flush()
            ctx.load_verify_locations(tmp_file.name)
            ctx.verify_flags |= ssl.VERIFY_CRL_CHECK_CHAIN

    def get_client_ssl_context(self, set_state=False) -> ssl.SSLContext:
        """获取客户端SSLContext"""
        _, cert, ctx = self._available_cert_and_ctx()
        if set_state:
            self._update_cert_usage_status(cert.name)
        return ctx

    def is_cert_pool_full(self) -> bool:
        """
        证书池是否已满
        :return: True: 证书池已满
        """
        return self._number_of_certs(NetManagerConstants.FUSION_DIRECTOR) >= NetManagerConstants.CERT_FROM_FD_LIMIT_NUM

    def cert_is_in_using(self, cert_name: str) -> int:
        """判断证书是否正在使用"""
        with self.session_maker() as session:
            return session.query(self.cert_info).filter_by(name=cert_name, in_use=True).count()

    def get_cert_info(self, net_mgmt_type: str, status: str) -> dict:
        """
        返给前端纳管页面展示的证书信息：
            网管模式：
                -- 就绪前，IP修改后返回可用的证书信息，IP修改前或无可用返回web证书；
                -- 就绪后，返回in_use为True的证书信息；
            非网管模式，返回web导入的证书信息，不管是否过期；
        """
        if net_mgmt_type == NetManagerConstants.FUSION_DIRECTOR:
            if status == "ready":
                return self._get_current_using_cert()

            if self.addr[0] != NetManagerConstants.SERVER_NAME:
                try:
                    return self._available_cert_and_ctx()[0].to_dict()
                except NetManagerException as err:
                    run_log.warning(err)

        run_log.info("Check current net mode is 'Web', get source is 'Web' cert info.")
        # Web导入的FD根证书最多1个，异常时记录warning日志
        if self._number_of_certs(source=NetManagerConstants.WEB) > 1:
            run_log.warning("Current fd certs larger than one, please check or upload certificate again.")
        return self._get_first_cert_info_by_source(source=NetManagerConstants.WEB)

    def cert_info_for_fd_generator(self) -> Iterable[dict]:
        """返回给FD的证书信息"""
        with self.session_maker() as session:
            for cert, cert_info in session.query(self.cert, self.cert_info).join(
                    self.cert_info, self.cert.name == self.cert_info.name
            ).limit(NetManagerConstants.CERT_FORM_FD_AND_WEB_LIMIT_NUM):
                yield cert_info.content_to_dict(bool(cert.crl_contents))

    def cert_for_restore_mini_os(self) -> dict:
        """获取恢复最小系统的证书"""
        return self._available_cert_and_ctx()[1].to_dict()

    def cert_to_mef(self) -> str:
        """发送给mef的证书内容"""
        return self._available_cert_and_ctx()[1].cert_contents

    def _get_current_using_cert(self) -> dict:
        """
        获取当前正在使用的证书，仅供纳管ready后会调用，故必然存在一条正在使用的记录
        :return: 当前正在使用证书
        """
        with self.session_maker() as session:
            return session.query(self.cert_info).filter_by(in_use=True).first().to_dict()

    def _get_first_cert_info_by_source(self, source: str) -> dict:
        with self.session_maker() as session:
            info = session.query(self.cert_info).join(self.cert, self.cert.name == self.cert_info.name).filter(
                self.cert.source == source
            ).first()
            return info.to_dict() if info else {}

    def _is_socket_connect_pass(self, context: ssl.SSLContext, cert_name: str) -> bool:
        """
        根据ip和端口进行socket连接测试
        :param context: ssl context
        :param cert_name: 证书名
        :return: True: 连接检查通过
        """
        try:
            with contextlib.ExitStack() as stack:
                sock = stack.enter_context(create_connection(address=self.addr, timeout=self.SOCKET_CONNECT_TIMEOUT))
                ssl_sock = stack.enter_context(context.wrap_socket(sock))
                ssl_sock.getpeercert(True)
        except Exception as err:
            if "revoked" in str(err):
                run_log.warning("%s revoked.", cert_name)
            return False

        return True

    def _cert_within_the_validity_period_generator(self) -> Iterable[Tuple[CertManager, CertInfo]]:
        with self.session_maker() as session:
            for cert, cert_info in self._certs_within_the_validity_period(session).limit(
                    NetManagerConstants.CERT_FORM_FD_AND_WEB_LIMIT_NUM
            ):
                session.expunge(cert)
                session.expunge(cert_info)
                yield cert, cert_info

    def _update_cert_usage_status(self, cert_name: str) -> NoReturn:
        """正在使用的证书最多只有一个，建立连接时设置"""
        with self.session_maker() as session:
            session.query(self.cert_info).update({"in_use": False})
            session.query(self.cert_info).filter_by(name=cert_name).update({"in_use": True})

    def _available_cert_and_ctx(self) -> Tuple[CertInfo, CertManager, ssl.SSLContext]:
        """可用的证书"""
        for cert, cert_info in self._cert_within_the_validity_period_generator():
            res, ctx = TlsConfig.get_client_context_with_cadata(cert.cert_contents)
            if not res:
                run_log.warning("Get client context with ca data failed, cert name is [%s]", cert.name)
                continue
            # 进行ssl握手前加载吊销列表内容，并开启吊销判断开关，如果遇到异常，则忽略掉对应证书
            try:
                self._load_crl(cert, ctx)
            except Exception as err:
                run_log.warning("load crl failed, catch %s", err.__class__.__name__)
                continue
            if self._is_socket_connect_pass(ctx, cert.name):
                return cert_info, cert, ctx

        raise NetManagerException("No available certificates found.")
