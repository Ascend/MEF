# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import json
import queue

from bin.HiSecAdap.constant import SecEyeAgentMsgType, SecEyeAgentTlvTag, MsgErrorCode
from bin.HiSecAdap.msg import MsgParse, MsgPacker
from common.log.logger import run_log

hisec_event_message_que = queue.Queue(maxsize=128)


class HiSecEvent:
    EVENT_DETAIL_MAX_LEN = 4096
    EVENT = {
        "Rootkit attack": {
            "event_id": "0x01000001",
            "name": "rootkit attack"
        },
        "Unauthorized root user": {
            "event_id": "0x01000002",
            "name": "unauthorized root user"
        },
        "Key file tampering": {
            "event_id": "0x01000003",
            "name": "key file tampering"
        },
        "Shell file tampering": {
            "event_id": "0x01000004",
            "name": "shell file tampering"
        },
        "File privilege escalation": {
            "event_id": "0x01000005",
            "name": "file privilege escalation"
        },
    }


class HiSecEventProc:

    @staticmethod
    def deal_hisec_message(message: dict):
        event_type = message.get("eventType")
        if event_type == "File privilege escalation":
            return HiSecEventProc.deal_hisec_file_privilege_escalation(message)
        elif event_type == "Key file tampering" or event_type == "Shell file tampering":
            return HiSecEventProc.deal_hisec_key_file_or_shell_file_tampering(message)
        elif event_type == "Unauthorized root user":
            return HiSecEventProc.deal_hisec_unauthorized_root_user(message)
        elif event_type == "Rootkit attack":
            return HiSecEventProc.deal_hisec_rootkit_attack(message)
        else:
            run_log.error("deal hisec message error")
            return ""

    @staticmethod
    def deal_hisec_file_privilege_escalation(message: dict):
        if not message or not isinstance(message, dict):
            run_log.error("deal hisec file privilege escalation error: invalid message format")
            return ""

        event_level = "[MINOR]"
        event_type = message.get("eventType")
        event_name = message.get("eventName", event_type)
        event_method = message.get("method")
        event_filepath = message.get("path")
        detail = f"{event_level} {event_name}, the method is {event_method}, the file is {event_filepath}."
        if len(detail) > HiSecEvent.EVENT_DETAIL_MAX_LEN:
            run_log.warning("hisec file privilege escalation event detail exceeds the length limit.")
            detail = f"{event_level} {event_name}"
        return detail

    @staticmethod
    def deal_hisec_key_file_or_shell_file_tampering(message: dict):
        event_level = "[MINOR]"
        event_type = message.get("eventType")
        event_name = message.get("eventName", event_type)
        event_filepath = message.get("path")
        detail = f"{event_level} {event_name}, file is {event_filepath}."

        evidence = message.get("evidence", {})
        if not evidence:
            return detail if len(detail) < HiSecEvent.EVENT_DETAIL_MAX_LEN else f"{event_level} {event_name}"

        if not isinstance(evidence, dict):
            run_log.error("deal hisec key file or shell file tampering error: message invalid format")
            return ""

        attribute = evidence.get("attribute")
        if attribute and isinstance(attribute, list):
            if not attribute[0] or not isinstance(attribute[0], dict):
                run_log.error("deal hisec key file or shell file tampering error: invalid attribute")
                return ""
            event_from = attribute[0].get("from")
            event_to = attribute[0].get("to")
            attribute_type = attribute[0].get("type")
            detail = f"{event_level} {event_name}, change {attribute_type} from {event_from} " \
                     f"to {event_to}, the file is {event_filepath}."
        event_associated_path = evidence.get("associatedPath")
        if event_associated_path:
            detail = f"{event_level} {event_name}, {event_filepath} has been moved to {event_associated_path}."
        if len(detail) > HiSecEvent.EVENT_DETAIL_MAX_LEN:
            run_log.warning(
                "hisec key file or shell file tampering event detail exceeds the length limit.")
            detail = f"{event_level} {event_name}"
        return detail

    @staticmethod
    def deal_hisec_unauthorized_root_user(message: dict):
        if not message or not isinstance(message, dict):
            run_log.error("deal hisec unauthorized root user error: invalid message format")
            return ""
        event_level = "[MINOR]"
        event_type = message.get("eventType")
        event_name = message.get("eventName", event_type)
        event_user = message.get("unauthorizedUser")
        detail = f"{event_level} {event_name}, the unauthorized root user is {event_user}."
        if len(detail) > HiSecEvent.EVENT_DETAIL_MAX_LEN:
            run_log.warning("hisec unauthorized root user event detail exceeds the length limit.")
            detail = f"{event_level} {event_name}"
        return detail

    @staticmethod
    def deal_hisec_rootkit_attack(message: dict):
        event_level = "[MINOR]"
        event_type = message.get("eventType")
        event_name = message.get("eventName", event_type)
        event_rootkit_name = message.get("rootkitName")
        if not message.get("evidence", None):
            detail = f"{event_level} {event_name}, the rootkit name is {event_rootkit_name}."
            if len(detail) > HiSecEvent.EVENT_DETAIL_MAX_LEN:
                run_log.warning("hisec rootkit attack event detail exceeds the length limit.")
                detail = f"{event_level} {event_name}"
            return detail

        feature = message.get("evidence").get("feature")
        try:
            if len(feature) > 1:
                event_filepath_all = list()
                HiSecEventProc._deal_msg_rootkit_attack(message, event_filepath_all)
                event_filepath = ",".join(event_filepath_all)
                detail = f"{event_level} {event_name}, the rootkit name is {event_rootkit_name}, " \
                         f"the file is {event_filepath}."
            else:
                event_filepath = feature[0].get("detectionSource")
                detail = f"{event_level} {event_name}, the rootkit name is {event_rootkit_name}, " \
                         f"the file is {event_filepath}."
        except Exception as err:
            run_log.error("deal hisec rootkit attack error: %s", err)
            return ""

        if len(detail) > HiSecEvent.EVENT_DETAIL_MAX_LEN:
            run_log.warning("hisec rootkit attack event detail exceeds the length limit.")
            detail = f"{event_level} {event_name}"
        return detail

    @staticmethod
    def push_hisec_event_task(msg):
        if hisec_event_message_que.full():
            run_log.warning("hisec event message queue is full, will be cleared")
            hisec_event_message_que.queue.clear()
        try:
            hisec_event_message_que.put(msg, False)
        except Exception as err:
            run_log.error("put hisec event message to queue failed: %s", err)

    @staticmethod
    def report_hisec_event(message: dict):
        if not isinstance(message, dict) or len(message) < 1:
            run_log.error("Invalid HiSec event message.")
            return

        event_type = message.get("eventType")
        event_const = HiSecEvent.EVENT
        if not isinstance(event_type, str) or not event_type or event_type not in event_const:
            run_log.error("Get HiSec event type failed.")
            return

        try:
            event_detail = HiSecEventProc.deal_hisec_message(message)
        except Exception as ex:
            run_log.error("deal hisec event failed, %s.", ex)
            return
        if not event_detail:
            run_log.error("Query event detail failed.")
            return

        event_id = event_const.get(event_type, {}).get("event_id")
        event_name = event_const.get(event_type, {}).get("name")
        payload_publish = {
            "alarm": [
                {
                    "type": "event",
                    "alarmId": event_id,
                    "alarmName": event_name,
                    "resource": "system",
                    "perceivedSeverity": "MINOR",
                    "timestamp": "",
                    "notificationType": "",
                    "detailedInformation": event_detail,
                    "suggestion": "",
                    "reason": "",
                    "impact": ""
                }
            ]
        }
        HiSecEventProc.push_hisec_event_task(payload_publish)

    @staticmethod
    def _deal_msg_rootkit_attack(message, event_filepath_all):
        for msg_dict in message.get("evidence").get("feature"):
            msg = msg_dict.get("detectionSource")
            event_filepath_all.append(msg)


class MsgSecEventReq(MsgParse):
    """parse sec event message from sec eye agent"""
    MSG_ID = SecEyeAgentMsgType.MSG_TYPE_SEND_JSON_SEC_EVENT_REQ.value

    SEC_EYE_EVENT_TYPE = (
        SecEyeAgentTlvTag.TLV_TAG_JSON_SEC_EVENT.value,
    )

    def __init__(self, origin_data: bytes):
        super().__init__(origin_data)
        self._is_check_ok: bool = True
        self._parse_msg()

    def deal_with(self) -> MsgErrorCode:
        if not self.is_parse_ok:
            return MsgErrorCode.MSG_PARSE_ERROR

        self._check_msg_req_para()
        if not self._is_check_ok:
            run_log.error("hisec check sec event failed")
            return MsgErrorCode.MSG_PARA_CHECK_ERROR

        self._report_to_fd()
        return MsgErrorCode.MSG_PARSE_OR_PACKER_OK

    def _parse_msg(self):
        self._pasre_msg_header()
        self.msg_body = self._parse_tlv_list(self.index, self.msg_body_len)

    def _check_msg_req_para(self):
        for one_event in self.msg_body:
            if one_event["tag"] not in self.SEC_EYE_EVENT_TYPE:
                self._is_check_ok = False
                run_log.error("hisec check sec event tag failed")
                return

    def _report_to_fd(self):
        for one_event in self.msg_body:
            value_str = one_event['value']
            value_json = json.loads(value_str.strip(b'\x00'.decode()))
            if not self._check_log_content_valid(value_json):
                run_log.error("hi sec sec event value invalid")
                continue
            run_log.info(f"hisec event: {value_json}")
            HiSecEventProc.report_hisec_event(value_json)


class MsgSecEventRsp(MsgPacker):
    """packer sec event message response to sec eye agent"""
    MSG_ID = SecEyeAgentMsgType.MSG_TYPE_SEND_JSON_SEC_EVENT_RSP.value

    def __init__(self, msg_map: dict):
        super().__init__(msg_map)

    def deal_with(self, ret_code: MsgErrorCode):
        self.msg_type = self.MSG_ID
        self.msg_body = {"tag": SecEyeAgentTlvTag.TLV_TAG_RETURN_CODE.value, "value": ret_code.value}
        self._packer_msg()

    def _packer_msg(self):
        self._packer_msg_header()
        self._packer_msg_body()
        self._assemble_msg_header_and_body()
