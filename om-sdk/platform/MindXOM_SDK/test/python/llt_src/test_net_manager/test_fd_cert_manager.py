#  Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
import contextlib
import socket
import ssl
import threading
from pathlib import Path
from typing import NamedTuple, Iterable, Optional, List

import pytest
from _pytest.logging import LogCaptureFixture
from pytest_mock import MockerFixture

from common.db.database import DataBase
from net_manager.constants import NetManagerConstants
from net_manager.manager.fd_cert_manager import FdCertManager
from net_manager.manager.import_manager import ImportCert, ImportCrl
from test_net_manager.conftest import chain_contents, cert_p

HOST, PORT = "127.0.0.1", 5678


class TlsSrv:

    def __init__(self, host: str, port: int, cert_file: Path, key_file: Path):
        self.host = host
        self.port = port
        self.cert_file: str = cert_file.as_posix()
        self.key_file: str = key_file.as_posix()
        self.started_event = threading.Event()
        self.closed_event = threading.Event()

    def start(self):
        srv = threading.Thread(target=self._server)
        srv.setDaemon(True)
        srv.start()
        if not self.started_event.wait(20):
            raise ValueError("env may be error.")

    def close(self):
        self.closed_event.set()

    def _server(self):
        svr = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        svr.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEPORT, 1)
        svr = ssl.wrap_socket(svr, server_side=True, keyfile=self.key_file, certfile=self.cert_file)
        svr.bind((self.host, self.port))
        svr.listen(0)
        self.started_event.set()
        while True:
            if self.closed_event.is_set():
                svr.close()
                break
            with contextlib.suppress():
                svr.accept()


class CertPoolFull(NamedTuple):
    expect: bool = False
    certs: Iterable[Iterable[Path]] = []


class ClientSslCtx(NamedTuple):
    err_msg: str = "No available certificates found."
    logs: Iterable[str] = []
    server_crt: Optional[Path] = None
    server_key: Optional[Path] = None
    cert_chains: Iterable[Iterable[Path]] = []
    crl_chains: Iterable[Iterable[Path]] = []


class CertInUsing(NamedTuple):
    expect: int = 0
    cert_name: str = "0"
    using_cert: str = "0"
    cert_chains: Iterable[Iterable[Path]] = []


class GetCertInfo(NamedTuple):
    mgr_type: str = NetManagerConstants.FUSION_DIRECTOR
    status: str = "ready"
    server_crt: Optional[Path] = None
    server_key: Optional[Path] = None
    cert_chains: Iterable[Iterable[Path]] = []
    crl_chains: Iterable[Iterable[Path]] = []
    using_cert_name: str = ""


class CertInfoForFD(NamedTuple):
    cert_chains: Iterable[Iterable[Path]] = []
    crl_chains: Iterable[Iterable[Path]] = []
    cert_crl_index: List[int] = []


class TestFdCertManager:
    use_cases = {
        "test_is_cert_pool_full": {
            "empty pool": CertPoolFull(),
            "full pool": CertPoolFull(expect=True, certs=(cert_p.single_root_ca(1), cert_p.single_root_ca(2)))
        },
        "test_available_cert_and_ctx": {
            "empty certs": ClientSslCtx(),
            "cert 0 revoked": ClientSslCtx(
                logs=("0 revoked.",),
                server_crt=cert_p.server_crt(1, "04"),
                server_key=cert_p.server_keyfile(1, "04"),
                cert_chains=(cert_p.normal_chain(1),),
                crl_chains=(cert_p.normal_crl_chain(1),)
            ),
            "not revoke should be success": ClientSslCtx(
                err_msg="",
                server_crt=cert_p.server_crt(1, "04"),
                server_key=cert_p.server_keyfile(1, "04"),
                cert_chains=(cert_p.normal_chain(1),),
            ),
        },
        "test_cert_is_in_using": {
            "empty certs not in using": CertInUsing(),
            "cert 0 in using": CertInUsing(
                expect=1,
                cert_chains=(cert_p.single_root_ca(1), cert_p.single_root_ca(2), cert_p.single_root_ca(2)),
            ),
            "cert 0 not in using": CertInUsing(
                expect=0,
                cert_name="1",
                cert_chains=(cert_p.single_root_ca(1), cert_p.single_root_ca(2), cert_p.single_root_ca(2)),
            ),
        },
        "test_get_cert_info": {
            "fd ready should get cur using cert": GetCertInfo(
                status="ready",
                mgr_type=NetManagerConstants.FUSION_DIRECTOR,
                cert_chains=(cert_p.normal_chain(1),),
                using_cert_name="0"
            ),
            "fd not ready should get available cert": GetCertInfo(
                status="",
                mgr_type=NetManagerConstants.FUSION_DIRECTOR,
                cert_chains=(cert_p.normal_chain(1),),
                server_crt=cert_p.server_crt(1, "04"),
                server_key=cert_p.server_keyfile(1, "04"),
            ),
            "fd not ready and no available cert should get fist web cert": GetCertInfo(
                status="",
                mgr_type=NetManagerConstants.FUSION_DIRECTOR,
                cert_chains=(cert_p.normal_chain(1),),
                server_crt=cert_p.server_crt(1, "04"),
                server_key=cert_p.server_keyfile(1, "04"),
                crl_chains=(cert_p.normal_crl_chain(1),),
            ),
        },
        "test_cert_info_for_fd_generator": {
            "empty certs": CertInfoForFD(),
            "with no crl": CertInfoForFD(cert_chains=(cert_p.normal_chain(1), cert_p.normal_chain(2))),
            "cert index 0 with crl": CertInfoForFD(
                cert_chains=(cert_p.normal_chain(1), cert_p.normal_chain(2)),
                crl_chains=(cert_p.normal_crl_chain(1),),
                cert_crl_index=[0, ]
            )
        }
    }

    @staticmethod
    def test_is_cert_pool_full(model: CertPoolFull, init_certs_db: DataBase, mocker: MockerFixture):
        mocker.patch.object(NetManagerConstants, "CERT_FROM_FD_LIMIT_NUM", 2)
        for index, cert_chain in enumerate(model.certs):
            ImportCert("FusionDirector", str(index)).import_deal(chain_contents(cert_chain))
        assert model.expect == FdCertManager().is_cert_pool_full()

    @staticmethod
    def test_available_cert_and_ctx(model: ClientSslCtx, init_certs_db: DataBase, caplog: LogCaptureFixture):
        for index, cert_chain in enumerate(model.cert_chains):
            ImportCert(cert_name=str(index)).import_deal(chain_contents(cert_chain))
        for crl_chain in model.crl_chains:
            ImportCrl().import_deal(chain_contents(crl_chain))
        mgr = FdCertManager(HOST, PORT)
        srv: Optional[TlsSrv] = None
        if model.server_crt:
            srv = TlsSrv(HOST, PORT, model.server_crt, model.server_key)
            srv.start()
        if model.err_msg:
            with pytest.raises(Exception, match=model.err_msg):
                mgr.get_client_ssl_context(True)
        else:
            mgr.get_client_ssl_context(True)
            assert "cert_contents" in mgr.cert_for_restore_mini_os()
            assert len(mgr.cert_to_mef())
        for msg in model.logs:
            assert msg in caplog.text
        if srv:
            srv.close()

    @staticmethod
    def test_cert_is_in_using(init_certs_db: DataBase, model: CertInUsing):
        for index, cert_chain in enumerate(model.cert_chains):
            ImportCert(cert_name=str(index)).import_deal(chain_contents(cert_chain))
        FdCertManager()._update_cert_usage_status(model.using_cert)
        assert model.expect == FdCertManager().cert_is_in_using(model.cert_name)

    @staticmethod
    def test_get_cert_info(init_certs_db: DataBase, model: GetCertInfo):
        mgr = FdCertManager(HOST, PORT)
        for index, cert_chain in enumerate(model.cert_chains):
            ImportCert(cert_name=str(index)).import_deal(chain_contents(cert_chain))
        for crl_chain in model.crl_chains:
            ImportCrl().import_deal(chain_contents(crl_chain))
        mgr._update_cert_usage_status(model.using_cert_name)
        srv: Optional[TlsSrv] = None
        if model.server_crt:
            srv = TlsSrv(HOST, PORT, model.server_crt, model.server_key)
            srv.start()
        info = mgr.get_cert_info(model.mgr_type, model.status)
        assert isinstance(info, dict) and info
        if srv:
            srv.close()

    @staticmethod
    def test_cert_info_for_fd_generator(init_certs_db: DataBase, model: CertInfoForFD):
        index = 0
        for index, cert_chain in enumerate(model.cert_chains):
            ImportCert(cert_name=str(index)).import_deal(chain_contents(cert_chain))
        for crl_chain in model.crl_chains:
            ImportCrl().import_deal(chain_contents(crl_chain))
        certs = list(FdCertManager().cert_info_for_fd_generator())
        assert index + 1 == len(certs) if index else not certs
        for index in model.cert_crl_index:
            assert certs[index]["is_import_crl"]
