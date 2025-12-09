# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
from bin.HiSecAdap.constant import SecEyeAgentMsgType, MsgErrorCode
from bin.HiSecAdap.msg_header import MsgHeaderParse, MsgHeaderPacker
from common.log.logger import run_log


class MsgHeartBeatReq(MsgHeaderParse):
    """parse heartbeat message request from sec eye agent"""
    MSG_ID = SecEyeAgentMsgType.MSG_TYPE_SEND_HEARTBEAT_REQ.value

    def __init__(self, origin_data: bytes):
        super().__init__(origin_data)
        self._parse_msg()

    def deal_with(self) -> MsgErrorCode:
        if not self.is_parse_ok:
            run_log.error("hisec heart beat parse failed")
            return MsgErrorCode.MSG_PARSE_ERROR

        return MsgErrorCode.MSG_PARSE_OR_PACKER_OK

    def _parse_msg(self):
        self._pasre_msg_header()


class MsgHeartBeatRsp(MsgHeaderPacker):
    """packer heartbeat message response to sec eye agent"""
    MSG_ID = SecEyeAgentMsgType.MSG_TYPE_SEND_HEARTBEAT_RSP.value

    def __init__(self, msg_header_map: dict):
        super().__init__(msg_header_map)

    def deal_with(self, ret_code: MsgErrorCode):
        if ret_code != MsgErrorCode.MSG_PARSE_OR_PACKER_OK:
            self.is_packer_ok = False
            return
        self.msg_type: int = self.MSG_ID
        self._packer_msg()

    def _packer_msg(self):
        self._packer_msg_header()
