# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import os
import socket

from bin.HiSecAdap.heart_beat import MsgHeartBeatReq, MsgHeartBeatRsp
from bin.HiSecAdap.sec_event import MsgSecEventReq, MsgSecEventRsp
from common.file_utils import FileCheck, FilePermission
from common.log.logger import run_log
from common.common_methods import CommonMethods


class MsgDeal:
    MSG_HANDLER_MAP = (
        [MsgHeartBeatReq, MsgHeartBeatRsp],
        [MsgSecEventReq, MsgSecEventRsp],
    )

    @staticmethod
    def deal_msg(send_func, one_msg_data):
        ret = MsgDeal._get_rsp(one_msg_data)
        if not ret[0]:
            run_log.warning("deal message: %s", ret[1])
            return

        send_func(ret[1])

    @staticmethod
    def _get_rsp(one_msg_data):
        if len(one_msg_data) < 16:
            return [False, "invalid msg"]
        msg_type = int.from_bytes(one_msg_data[4:8], byteorder='little')

        for msg_handler in MsgDeal.MSG_HANDLER_MAP:
            if msg_handler[0].MSG_ID == msg_type:
                msg_req_instance = msg_handler[0](one_msg_data)
                if not msg_req_instance.is_parse_ok:
                    run_log.info("parse msg failed. %s", msg_req_instance.MSG_ID)
                ret_code = msg_req_instance.deal_with()

                msg_rsp_instance = msg_handler[1](msg_req_instance.get_msg_map())
                msg_rsp_instance.deal_with(ret_code)
                if not msg_rsp_instance.is_packer_ok:
                    run_log.info("msg packe failed : %s", msg_rsp_instance.MSG_ID)
                    return [False, "msg packe failed"]

                return [True, msg_rsp_instance.get_msg_bytes()]

        return [False, "msg not support now"]


class UdsServer:
    MSG_HEADER_LEN = 16
    MSG_BODY_LEN_START_INDEX = 12
    MAX_MSG_BODY_LEN = 1 * 1024 * 1024  # 1MB
    MAX_MSG_BUFFER_LEN = 10 * 1024 * 1024  # 10MB

    def __init__(self, **kwargs):
        self._msg_buffer: bytes = b''
        self._max_link: int = kwargs.get("max_link", 100)
        self._data_len: int = kwargs.get("data_len", 1024)
        self._socket_path: str = ""
        self._deal_func = None
        self.client_sock = None

    def init_server(self, socket_path="", func=None):
        if not isinstance(socket_path, str) or socket_path == "":
            return [False, "socketpath is invalid"]

        self._socket_path = socket_path
        if not callable(func):
            return [False, "deal func is invalid"]
        self._deal_func = func

        self._start_socket_server()
        return [True, ""]

    def _start_socket_server(self):
        with socket.socket(socket.AF_UNIX, socket.SOCK_STREAM) as sock:
            sock.bind(self._socket_path)
            FilePermission.set_path_owner_group(self._socket_path, "root")
            FilePermission.set_path_permission(self._socket_path, 0o600)
            sock.listen(self._max_link)
            while True:
                self._msg_buffer = b''
                self.client_sock, _ = sock.accept()
                self._receive_and_handle_msg()

    def _get_one_msg_and_deal(self):
        cur_msg_buffer_len = len(self._msg_buffer)
        if cur_msg_buffer_len > self.MAX_MSG_BUFFER_LEN:
            self._msg_buffer = b''
            raise ValueError("current msg buffer len is too large")
        if cur_msg_buffer_len >= self.MSG_HEADER_LEN:
            cur_msg_body_len = int.from_bytes(
                self._msg_buffer[self.MSG_BODY_LEN_START_INDEX:self.MSG_HEADER_LEN],
                byteorder='little')
            if cur_msg_body_len > self.MAX_MSG_BODY_LEN or cur_msg_body_len < 0:
                raise ValueError("current msg body len is too large")
            if cur_msg_buffer_len >= self.MSG_HEADER_LEN + cur_msg_body_len and self._deal_func is not None:
                # 单线程处理消息
                self._deal_func(self._send_msg, self._msg_buffer[:self.MSG_HEADER_LEN + cur_msg_body_len])
                self._msg_buffer = self._msg_buffer[self.MSG_HEADER_LEN + cur_msg_body_len:]

    def _receive_and_handle_msg(self):
        try:
            while True:
                byte_data_tmp = self.client_sock.recv(self._data_len)
                if len(byte_data_tmp) == 0:
                    run_log.info("socket link break")
                    break
                self._msg_buffer = b''.join([self._msg_buffer, byte_data_tmp])
                try:
                    self._get_one_msg_and_deal()
                except Exception as ex:
                    run_log.error("deal msg failed : %s", ex)
        except Exception as except_info:
            run_log.error("receive dada error %s", except_info)
        finally:
            try:
                # 先 shutdown, 然后 close.
                self.client_sock.shutdown(socket.SHUT_RDWR)
            except Exception:
                run_log.error("socket shutdown failed.")
            try:
                self.client_sock.close()
            except Exception:
                run_log.error("Deal request socket close failed.")

    def _send_msg(self, data: bytes):
        self.client_sock.sendall(data)


def start_uds():
    socket_path = "/usr/local/mindx/MindXOM/software/sec_agent/server.socket"
    run_log.info("socket_path : %s", socket_path)
    if os.path.exists(socket_path):
        run_log.info("Delete hisec server.socket")
        os.remove(socket_path)

    ret = FileCheck.check_input_path_valid(socket_path)
    if not ret:
        err_msg = f"socket path {socket_path} invalid, {ret.error}"
        run_log.error("Server start failed, message: %s ", err_msg)
        return

    try:
        ret = UdsServer().init_server(socket_path, MsgDeal.deal_msg)
    except Exception as err:
        run_log.error("start_uds exception:%s", err)
        return

    if not ret[0]:
        run_log.error("Server start failed, message: %s", ret)
