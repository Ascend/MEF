import socket
import ssl
from collections import namedtuple
from ssl import SSLContext
from ssl import SSLSocket

import mock
from pytest_mock import MockerFixture

from common.file_utils import FileCheck
from common.kmc_lib.kmc import Kmc
from common.kmc_lib.tlsconfig import TlsConfig
from common.utils.result_base import Result
from ibma_client import Client
from ibma_client import socket_set_so_linger

RecvCase = namedtuple("RecvCase", "excepted, recv")
SendCase = namedtuple("SendCase", "excepted, exists, get_client_ssl_context, connect")


class TestClient:
    use_cases = {
        "test_receive_socket_message": {
            "small_len": ("test", b"test;;over;"),
        },
        "test_send_msg": {
            "not_exists": ([400, "Socket path is not exist."], False, None, None),
            "get_client_failed": ([400, "Send message failed."], True, [False, ssl.SSLContext()], None),
            "connect_failed": (["failed", "Connect server failed."], True, [True, ssl.SSLContext()], "failed"),
            "success": ([0, "test"], True, [True, ssl.SSLContext()], None),
        }
    }

    recv_msg = ""

    @staticmethod
    def recv_test():
        return TestClient.recv_msg

    def test_socket_set_so_linger(self):
        sock = None
        try:
            sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
            ret = socket_set_so_linger(sock)
            assert ret is None
        finally:
            if sock:
                sock.close()

    def test_receive_socket_message(self, model: RecvCase):
        with mock.patch("socket.socket") as mock_socket:
            mock_socket.return_value.recv.return_value = model.recv
            ret = Client.receive_socket_message(mock_socket.return_value)
            assert model.excepted == ret

    def test_send_msg(self, mocker: MockerFixture, model: SendCase):
        mocker.patch("os.path.exists", return_value=model.exists)
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid", return_value=Result(result=True))
        mocker.patch.object(TlsConfig, "get_ssl_context", return_value=model.get_client_ssl_context)
        mocker.patch.object(socket.socket, "connect", return_value=model.connect)
        mocker.patch("builtins.open")
        mocker.patch.object(Kmc, "decrypt")
        mocker.patch.object(SSLContext, "wrap_socket")
        mocker.patch.object(SSLSocket, "sendall")
        mocker.patch.object(Client, "receive_socket_message", return_value="test")
        mocker.patch.object(SSLSocket, "shutdown")
        ret = Client.send_msg("test")
        assert model.excepted == ret
