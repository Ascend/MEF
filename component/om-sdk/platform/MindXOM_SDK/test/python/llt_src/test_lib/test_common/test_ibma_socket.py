import socket
from collections import namedtuple

import mock
from pytest_mock import MockFixture

from common.utils.result_base import Result
from bin.ibma_server import Server

Start = namedtuple("Start", "expect, check_path_is_exist_and_valid")
ReceiveSocketMessage = namedtuple("ReceiveSocketMessage", "expect, recv")
DealRequest = namedtuple("DealRequest", "expect, recv")
Worker = namedtuple("Worker", "expect")
InitServer = namedtuple("InitServer", "expect, input_para1, input_para2")
WaitClient = namedtuple("WaitClient", "expect, recv, input_para2")


class TestIbmaSocket:
    use_cases = {
        "test_start": {
            "normal": (None, Result(True)),
            "check_path_is_exist_and_valid_is_false": (None, Result(False)),
        },
        "test_init_server": {
            "socket_path_is_invalid": ([1, "socketPath is invalid."], 123, int),
            "func_is_invalid": ([1, "Socket server start failed."], "test", 123),
            "func_is_none": ([1, "Function is invalid."], "test", None)
        },
        "test_receive_socket_message": {
            "normal": ("test", b"test;;over;"),
        },
        "test_deal_request": {
            "normal": (None, b"test;;over;"),
        },
        "test_worker": {
            "normal": (None, ),
        },
        "test_wait_client": {
            "normal": (None, b"test;;over;", int),
        },
    }

    def test_socket_set_so_linger(self):
        with socket.socket(socket.AF_UNIX, socket.SOCK_STREAM) as sock:
            ret = Server.socket_set_so_linger(sock)
            assert ret is None

    @staticmethod
    def test_init_server(mocker: MockFixture, model: InitServer):
        server = Server()
        assert model.expect == server.init_server(model.input_para1, model.input_para2)

    @staticmethod
    def test_deal_request(mocker: MockFixture, model: DealRequest):
        server = Server()
        with mock.patch("socket.socket") as mock_socket:
            mock_socket.return_value.recv.return_value = model.recv
            assert model.expect == server.deal_request(mock_socket.return_value, str)

    @staticmethod
    def test_receive_socket_message(model: ReceiveSocketMessage):
        server = Server()
        with mock.patch("socket.socket") as mock_socket:
            mock_socket.return_value.recv.return_value = model.recv
        assert model.expect == server.receive_socket_message(mock_socket.return_value)
