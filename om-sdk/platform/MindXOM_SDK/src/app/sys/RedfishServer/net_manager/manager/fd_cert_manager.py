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
import contextlib
import os
import ssl
from functools import partial
from typing import List
from typing import Optional
from typing import Tuple
from typing import Union

from urllib3.util.connection import create_connection

from cert_manager.parse_tools import ParseCertInfo
from cert_manager.parse_tools import ParseCrlInfo
from common.file_utils import FileCopy
from common.log.logger import run_log
from common.kmc_lib.tlsconfig import TlsConfig
from common.utils.result_base import Result
from net_manager.constants import NetManagerConstants
from net_manager.manager.net_cfg_manager import NetCertManager
from net_manager.models import CertManager


class FdCertManager:
    SOCKET_CONNECT_TIMEOUT: int = 5

    def __init__(self, server_name: str = "", port: Union[int, str] = ""):
        self.cert_manager: NetCertManager = NetCertManager()
        self.addr: Tuple[str, int] = (
            server_name or NetManagerConstants.SERVER_NAME,
            int(port or NetManagerConstants.PORT),
        )

    def is_socket_connect_pass(self, context: ssl) -> bool:
        """
        根据ip和端口进行socket连接测试
        :param context: ssl context
        :return: True: 连接检查通过
        """
        try:
            with contextlib.ExitStack() as stack:
                sock = stack.enter_context(create_connection(address=self.addr, timeout=self.SOCKET_CONNECT_TIMEOUT))
                ssl_sock = stack.enter_context(context.wrap_socket(sock))
                ssl_sock.getpeercert(True)
        except Exception:  # 无需获取异常内容，抛异常均认为是测试不通过
            return False

        return True

    def is_cert_pool_full(self) -> bool:
        """
        证书池是否已满
        :return: True: 证书池已满
        """
        cert_managers: List[CertManager] = self.cert_manager.get_obj_by_source(NetManagerConstants.FUSION_DIRECTOR)
        return len(cert_managers) >= NetManagerConstants.CERT_FROM_FD_LIMIT_NUM

    def get_current_using_cert(self) -> Optional[CertManager]:
        """
        获取当前正在使用的证书
        :return: 当前正在使用证书
        """
        cert_managers: List[CertManager] = self.cert_manager.get_all()
        for cert_manager in cert_managers:
            res, ssl_context = TlsConfig.get_client_context_with_cadata(ca_data=cert_manager.cert_contents)
            if not res:
                run_log.warning("Get client context with ca data failed, cert_manager_id is [%s]", cert_manager.id)
                continue

            if not self.is_socket_connect_pass(ssl_context):
                continue

            return cert_manager

    def get_cert_info(self, net_mgmt_type) -> dict:
        """
        获取当前FD证书信息
        先判断是否纳管到FD，如果已纳管，则获取当前正在使用的证书信息，否则获取来源于Web导入的证书信息
        :return:
        """
        if net_mgmt_type == NetManagerConstants.FUSION_DIRECTOR:
            cert_manager: CertManager = self.get_current_using_cert()
            # 获取到有使用的证书时才去解析证书信息，否则还是去查询web导入的证书信息
            if cert_manager:
                with ParseCertInfo(cert_manager.cert_contents) as cert_info:
                    return cert_info.to_dict()

        # 当前未纳管到FD，获取来源于Web导入的证书
        run_log.info("Check current net mode is 'Web', get source is 'Web' cert info.")
        cert_managers: List[CertManager] = self.cert_manager.get_obj_by_source(NetManagerConstants.WEB)
        if not cert_managers:
            return {}

        # Web导入的FD根证书只能1个，异常时记录warning日志
        if len(cert_managers) != 1:
            run_log.warning("Current fd certs larger than one, please check or upload certificate again.")

        with ParseCertInfo(cert_managers[0].cert_contents) as cert_info:
            return cert_info.to_dict()

    def get_client_ssl_context(self) -> Result:
        tmp_cert_file = "/run/web/tmp_fd_crt.crt"
        tmp_crl_file = "/run/web/tmp_fd_crl.crl"
        try:
            return self._gen_ssl_context(tmp_cert_file, tmp_crl_file)
        except Exception as err:
            return Result(result=False, err_msg=str(err))
        finally:
            for file in tmp_cert_file, tmp_crl_file:
                FileCopy.remove_path(file)

    def verify_crl_against_ca(self, crl_contents: str, source: str) -> Result:
        try:
            cert_managers: List[CertManager] = {
                NetManagerConstants.WEB: partial(self.cert_manager.get_obj_by_source, NetManagerConstants.WEB),
                NetManagerConstants.FUSION_DIRECTOR: partial(self.cert_manager.get_all),
            }.get(source)()
        except Exception as err:
            return Result(result=False, err_msg=f"Verify crl against CA certificate failed, reason is {err}")

        if not cert_managers:
            return Result(result=False, err_msg="Check certificate is null. Please check and upload certificate.")

        if source == NetManagerConstants.WEB and len(cert_managers) != 1:
            return Result(result=False, err_msg=f"Check from {source} certificate failed. Please check the quantity.")

        try:
            with ParseCrlInfo(crl_contents) as parse_crl:
                return parse_crl.verify_crl_buffer_by_ca(cert_managers)
        except Exception as err:
            return Result(result=False, err_msg=f"Verify crl against CA certificate failed, reason is {err}")

    def _gen_ssl_context(self, tmp_cert_file, tmp_crl_file) -> Result:
        cert_manager = self.get_current_using_cert()
        if not cert_manager or not cert_manager.cert_contents:
            return Result(False, err_msg="certificate is null.")

        flags = os.O_WRONLY | os.O_CREAT | os.O_TRUNC
        with os.fdopen(os.open(tmp_cert_file, flags=flags, mode=0o600), "w") as tmp_file:
            tmp_file.write(cert_manager.cert_contents)

        if not cert_manager.crl_contents:
            res, ssl_context = TlsConfig.get_client_ssl_context(tmp_cert_file)
        else:
            with os.fdopen(os.open(tmp_crl_file, flags=flags, mode=0o600), "w") as tmp_file:
                tmp_file.write(cert_manager.crl_contents)

            res, ssl_context = TlsConfig.get_client_ssl_context(tmp_cert_file, tmp_crl_file)

        return Result(True, data=ssl_context) if res else Result(False, err_msg=ssl_context)
