# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from bin.HiSecAdap.constant import SecEyeAgentTlvTag
from bin.HiSecAdap.msg_header import MsgHeaderParse, MsgHeaderPacker
from common.log.log_handlers import LOG_INJECTION_VALUES
from common.log.logger import run_log


class MsgParse(MsgHeaderParse):
    """parse message from sec eye agent"""
    TLV_TAG_SIZE = 2
    TLV_LENGTH_SIZE = 2
    INVALID_VALUE = -1
    MAX_LEN_VAL = 1 * 1024 * 1024
    MAX_STR_VAL_LEN = 1 * 1024 * 1024
    MAX_ALARM_COUNT = 64

    INT_VALUE = (
        SecEyeAgentTlvTag.TLV_TAG_SYS_LOG_LOGID.value,
        SecEyeAgentTlvTag.TLV_TAG_SYS_LOG_LOGLEVEL.value,
        SecEyeAgentTlvTag.TLV_TAG_SYS_LOG_LOGTYPE.value,
        SecEyeAgentTlvTag.TLV_TAG_SYS_LOG_PROC_ID.value,
    )
    STRING_VALUE = (
        SecEyeAgentTlvTag.TLV_TAG_SYS_LOG_TIMESTAMP.value,
        SecEyeAgentTlvTag.TLV_TAG_SYS_LOG_APP_NAME.value,
        SecEyeAgentTlvTag.TLV_TAG_SYS_LOG_DESCRIPTION.value,
        SecEyeAgentTlvTag.TLV_TAG_SYS_LOG_CONTEXT.value,
        SecEyeAgentTlvTag.TLV_TAG_SYS_LOG_LOGCONTENT.value,
        SecEyeAgentTlvTag.TLV_TAG_SYS_LOG_RAWCONTENT.value,
        SecEyeAgentTlvTag.TLV_TAG_JSON_SEC_EVENT.value,
    )

    def __init__(self, origin_data: bytes):
        super().__init__(origin_data)
        self.msg_body = []

    def get_msg_map(self):
        return {"intfVersion": self.intfVersion, "msg_type": self.msg_type, "msg_seq_num": self.msg_seq_num,
                "msg_body_len": self.msg_body_len, "msg_body": self.msg_body}

    def _check_log_content_valid(self, log_content):
        if len(log_content) > self.MAX_STR_VAL_LEN:
            return False

        for ch in LOG_INJECTION_VALUES:
            if ch in log_content:
                return False

        return True

    def _parse_tlv_double_list(self, tlv_list_start_index, tlv_list_len):
        tlv_double_list = []
        sub_tlv_index = tlv_list_start_index

        if tlv_list_len > self.MAX_LEN_VAL:
            self.is_parse_ok = False
            run_log.error("tlv_list_len is too large")
            return tlv_double_list

        count = 0
        while sub_tlv_index < tlv_list_start_index + tlv_list_len and count < self.MAX_ALARM_COUNT:

            tag = self._parse_tag(sub_tlv_index)
            sub_tlv_index += self.TLV_TAG_SIZE

            length = self._parse_length(sub_tlv_index)
            sub_tlv_index += self.TLV_LENGTH_SIZE

            sub_tlv_list = self._parse_tlv_list(sub_tlv_index, length)

            if not self.is_parse_ok:
                run_log.error("hisec parse msg body double list sub tlv failed")
                return []
            sub_tlv_index += length
            count += 1
            tlv_double_list.append({"tag": tag, "length": length, "value": sub_tlv_list})

        if sub_tlv_index != tlv_list_start_index + tlv_list_len:
            self.is_parse_ok = False
            run_log.error("hisec parse msg body double list failed")
            return []

        return tlv_double_list

    def _parse_tlv_list(self, tlv_list_start_index, tlv_list_len):
        tlv_list = []
        sub_tlv_index = tlv_list_start_index

        if tlv_list_len > self.MAX_LEN_VAL or tlv_list_len < 0:
            self.is_parse_ok = False
            run_log.error("tlv_list_len is too large")
            return tlv_list

        count = 0
        while sub_tlv_index < tlv_list_start_index + tlv_list_len and count < self.MAX_ALARM_COUNT:
            tag = self._parse_tag(sub_tlv_index)
            sub_tlv_index += self.TLV_TAG_SIZE

            length = self._parse_length(sub_tlv_index)
            sub_tlv_index += self.TLV_LENGTH_SIZE

            value = self._parse_value(sub_tlv_index, length, tag)

            if not self.is_parse_ok:
                run_log.error("hisec parse msg body list sub tlv failed")
                return []
            sub_tlv_index += length
            count += 1
            tlv_list.append({"tag": tag, "length": length, "value": value})

        if sub_tlv_index != tlv_list_start_index + tlv_list_len:
            run_log.error("hisec parse msg body list failed")
            self.is_parse_ok = False
            return []

        return tlv_list

    def _parse_tlv(self, tlv_index):
        tag = self._parse_tag(tlv_index)
        tlv_index += self.TLV_TAG_SIZE

        length = self._parse_length(tlv_index)
        tlv_index += self.TLV_LENGTH_SIZE

        value = self._parse_value(tlv_index, length, tag)

        return {"tag": tag, "length": length, "value": value}

    def _parse_tag(self, tlv_tag_index):
        end_idx = tlv_tag_index + self.TLV_TAG_SIZE

        if end_idx > self.byte_data_len:
            # 返回无效值-1，有效值都是非负数，而且_parse_value还会校验tag的具体取值范围
            run_log.error("invalid tag parameter")
            return self.INVALID_VALUE

        ret = int.from_bytes(self.byte_data[tlv_tag_index:end_idx], byteorder='big')
        if ret > SecEyeAgentTlvTag.TLV_TAG_CHECK_BENCHMARK_RESULT.value or ret < 0:
            run_log.error("invalid length parameter")
            return self.INVALID_VALUE

        return ret

    def _parse_length(self, tlv_length_index):
        end_idx = tlv_length_index + self.TLV_LENGTH_SIZE

        if end_idx > self.byte_data_len:
            # 返回无效值-1，有效值都是非负数，而且while循环会直接退出
            run_log.error("invalid length parameter")
            return self.INVALID_VALUE

        ret = int.from_bytes(self.byte_data[tlv_length_index:end_idx], byteorder='big')
        if ret > self.MAX_LEN_VAL or ret < 0:
            run_log.error("invalid length parameter")
            return self.INVALID_VALUE

        return ret

    def _parse_value(self, tlv_value_index, value_len, tag):
        value = 0
        end_idx = tlv_value_index + value_len
        if tag in self.INT_VALUE:
            if end_idx > self.byte_data_len:
                run_log.error("invalid value parameter")
                return self.INVALID_VALUE
            value = int.from_bytes(self.byte_data[tlv_value_index:end_idx], byteorder='little')

        elif tag in self.STRING_VALUE:
            if end_idx > self.byte_data_len:
                run_log.error("invalid value parameter")
                return ""
            value = self.byte_data[tlv_value_index:end_idx].decode()
            if len(value) > self.MAX_STR_VAL_LEN:
                run_log.error("length of string type value is too large")
                return ""

        else:
            self.is_parse_ok = False
            run_log.error("hisec parse tlv value failed")

        return value


class MsgPacker(MsgHeaderPacker):
    """packer response message to sec eye agent"""
    TLV_TAG_SIZE = 2
    TLV_LENGTH_SIZE = 2

    def __init__(self, msg_map: dict):
        super().__init__(msg_map)
        self.msg_body = msg_map.get("msg_body")
        self.msg_body_bytes: bytes = b''

    def packer_list_map(self, list_map):
        list_map_bytes = b''
        for tlv_map in list_map:
            list_map_bytes = b''.join([list_map_bytes, self.packer_map(tlv_map)])

        return list_map_bytes

    def packer_map(self, tlv_map):
        map_bytes = b''
        if tlv_map.get("tag", None) is None:
            self.is_packer_ok = False
            run_log.error("hisec packer tlv tag failed")
            return b''

        tag = tlv_map["tag"]
        tag_bytes = self.packer_tag(tag)

        if tlv_map.get("value", None) is None:
            self.is_packer_ok = False
            run_log.error("hisec packer tlv value failed")
            return b''

        value = tlv_map.get("value")
        value_bytes = self.packer_value(value)
        value_bytes_len = len(value_bytes)

        map_bytes = b''.join([map_bytes, tag_bytes])
        map_bytes = b''.join([map_bytes, int(value_bytes_len).to_bytes(self.TLV_LENGTH_SIZE, byteorder='big')])
        map_bytes = b''.join([map_bytes, value_bytes])

        return map_bytes

    def packer_tag(self, tag: int):
        try:
            data_value = int(tag).to_bytes(self.TLV_TAG_SIZE, byteorder='big')
        except Exception as err:
            run_log.error(f"get data failed {err}")
            data_value = 0

        return data_value

    def packer_value(self, value):
        value_bytes = b''
        if isinstance(value, list):
            value_bytes = self.packer_list_map(value)
        elif isinstance(value, dict):
            value_bytes = self.packer_map(value)
        elif isinstance(value, int):
            value_bytes = int(value).to_bytes(4, byteorder='little')
        elif isinstance(value, str):
            value_bytes = value.encode()
        else:
            self.is_packer_ok = False
            run_log.error("hisec packer value failed")
            return value_bytes

        return value_bytes

    def get_msg_bytes(self):
        return self.msg_bytes

    def _assemble_msg_header_and_body(self):
        msg_body_byte_len = len(self.msg_body_bytes)

        self.msg_bytes = self.msg_bytes[:-4]
        self.msg_bytes = b''.join(
            [self.msg_bytes, int(msg_body_byte_len).to_bytes(self.MSG_BODY_LENGTH_SIZE, byteorder='little')])
        self.msg_bytes = b''.join([self.msg_bytes, self.msg_body_bytes])

    def _packer_msg_body(self):
        if isinstance(self.msg_body, list):
            self.msg_body_bytes = self.packer_list_map(self.msg_body)
        elif isinstance(self.msg_body, dict):
            self.msg_body_bytes = self.packer_map(self.msg_body)
        else:
            self.is_packer_ok = False
