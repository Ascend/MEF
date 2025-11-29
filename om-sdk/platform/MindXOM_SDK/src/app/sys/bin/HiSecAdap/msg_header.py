# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from common.log.logger import run_log


class MsgHeaderParse(object):
    """parse message header from sec eye agent"""
    MSG_INTER_VERION_SIZE = 4
    MSG_TYPE_SIZE = 4
    MSG_SEQ_NUM_SIZE = 4
    MSG_BODY_LENGTH_SIZE = 4

    def __init__(self, origin_data: bytes):
        self.is_parse_ok: bool = True
        self.byte_data: bytes = origin_data
        self.byte_data_len: int = len(self.byte_data)
        self.index: int = 0
        self.intfVersion: int = 0
        self.msg_type: int = 0
        self.msg_seq_num: int = 0
        self.msg_body_len: int = 0

    def get_msg_map(self):
        return {"intfVersion": self.intfVersion, "msg_type": self.msg_type, "msg_seq_num": self.msg_seq_num,
                "msg_body_len": self.msg_body_len}

    def _pasre_msg_header(self):
        self._pasre_interface_version()
        self._pasre_msg_type()
        self._pasre_msg_seq_num()
        self._pasre_msg_body_len()

    def _pasre_interface_version(self):
        if self.byte_data_len < self.index + self.MSG_INTER_VERION_SIZE:
            self.intfVersion = 0
            self.is_parse_ok = False
            run_log.error("hisec parse intfVersion error")
        else:
            self.intfVersion = int.from_bytes(self.byte_data[self.index:self.index + self.MSG_INTER_VERION_SIZE],
                                              byteorder='little')
            self.index += self.MSG_INTER_VERION_SIZE

    def _pasre_msg_type(self):
        if self.byte_data_len < self.index + self.MSG_TYPE_SIZE:
            self.msg_type = 0
            self.is_parse_ok = False
            run_log.error("hisec parse msg_type error")
        else:
            self.msg_type = int.from_bytes(self.byte_data[self.index:self.index + self.MSG_TYPE_SIZE],
                                           byteorder='little')
            self.index += self.MSG_TYPE_SIZE

    def _pasre_msg_seq_num(self):
        if self.byte_data_len < self.index + self.MSG_SEQ_NUM_SIZE:
            self.msg_seq_num = 0
            self.is_parse_ok = False
            run_log.error("hisec parse msg_seq_num error")
        else:
            self.msg_seq_num = int.from_bytes(self.byte_data[self.index:self.index + self.MSG_SEQ_NUM_SIZE],
                                              byteorder='little')
            self.index += self.MSG_SEQ_NUM_SIZE

    def _pasre_msg_body_len(self):
        if self.byte_data_len < self.index + self.MSG_BODY_LENGTH_SIZE:
            self.msg_body_len = 0
            self.is_parse_ok = False
            run_log.error("hisec parse msg_body_len error")
        else:
            self.msg_body_len = int.from_bytes(self.byte_data[self.index:self.index + self.MSG_BODY_LENGTH_SIZE],
                                               byteorder='little')
            self.index += self.MSG_BODY_LENGTH_SIZE


class MsgHeaderPacker:
    """packer response message header to sec eye agent"""
    MSG_INTER_VERION_SIZE = 4
    MSG_TYPE_SIZE = 4
    MSG_SEQ_NUM_SIZE = 4
    MSG_BODY_LENGTH_SIZE = 4

    def __init__(self, msg_header_map: dict):
        self.is_packer_ok: bool = True
        self.intfVersion: int = msg_header_map.get("intfVersion")
        self.msg_type: int = 0
        self.msg_seq_num: int = msg_header_map.get("msg_seq_num")
        self.msg_body_len: int = 0
        self.msg_bytes: bytes = b''

    def get_msg_bytes(self):
        return self.msg_bytes

    def _packer_msg_header(self):
        self._packer_msg_interface_version()
        self._packer_msg_type()
        self._packer_msg_seq_num()
        self._packer_msg_body_len()

    def _packer_msg_interface_version(self):
        if self.intfVersion is not None:
            try:
                self.msg_bytes = b''.join(
                    [self.msg_bytes, int(self.intfVersion).to_bytes(self.MSG_INTER_VERION_SIZE, byteorder='little')])
            except Exception as err:
                self.is_packer_ok = False
                run_log.error(f"get data failed.{err}")
        else:
            self.is_packer_ok = False
            run_log.error("hisec packer intfVersion error")

    def _packer_msg_type(self):
        if self.msg_type is not None:
            try:
                self.msg_bytes = b''.join(
                    [self.msg_bytes, int(self.msg_type).to_bytes(self.MSG_TYPE_SIZE, byteorder='little')])
            except Exception as err:
                self.is_packer_ok = False
                run_log.error(f"get data failed.{err}")
        else:
            self.is_packer_ok = False
            run_log.error("hisec packer msg_type error")

    def _packer_msg_seq_num(self):
        if self.msg_seq_num is not None:
            try:
                self.msg_bytes = b''.join(
                    [self.msg_bytes, int(self.msg_seq_num).to_bytes(self.MSG_SEQ_NUM_SIZE, byteorder='little')])
            except Exception as err:
                run_log.error(f"get data failed.{err}")
                self.is_packer_ok = False
        else:
            self.is_packer_ok = False
            run_log.error("hisec packer msg_seq_num error")

    def _packer_msg_body_len(self):
        if self.msg_body_len is not None:
            try:
                self.msg_bytes = b''.join(
                    [self.msg_bytes, int(self.msg_body_len).to_bytes(self.MSG_BODY_LENGTH_SIZE, byteorder='little')])
            except Exception as err:
                run_log.error(f"get data failed.{err}")
                self.is_packer_ok = False
        else:
            self.is_packer_ok = False
            run_log.error("hisec packer msg_body_len error")
