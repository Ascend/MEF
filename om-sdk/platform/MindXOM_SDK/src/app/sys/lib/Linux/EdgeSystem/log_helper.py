#!/usr/bin/python
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.

from common.log.logger import run_log
from common.log.logger import operate_log
import common.common_methods as commMethods


# NOTE: when adding a new error code, be sure to also change the value of
# MINID_ERROR_CODE_BEGIN and XXX_ERROR_CODE_END to be correct!!!
MINID_ERROR_CODE_BEGIN = 130
ERR_GET_CHIP_COUNT = 130
ERR_GET_CHIP_INFO = 131
ERR_GET_CHIP_MEM_SIZE = 132
ERR_GET_NPU_LIST = 133
ERR_GET_MINIDI_TEMPERATURE = 134
ERR_GET_HEALTH = 135
ERR_DEVICE_TYPE = 136
ERR_GET_DAVINCI = 137
MINID_ERROR_CODE_END = 137

MCU_ERROR_CODE_BEGIN = 138
ERR_GET_TEMPERATURE = 138
ERR_GET_VOLTAGE = 139
ERR_GET_TIME = 140
ERR_SET_TIME_FAILED = 141
ERR_GET_BOARD_INFO = 142
ERR_GET_OS_VERSION = 143
ERR_GET_KERNEL_VERSION = 144
ERR_GET_UPTIME = 145
ERR_GET_MCU_HEALTH = 146
ERR_GET_MCU_HEATING_COOL = 147
ERR_GET_POWER = 148
MCU_ERROR_CODE_END = 148

H3559_ERROR_CODE_BEGIN = 152
ERR_GET_MEMORY_USAGE = 152
ERR_GET_ELECTRONIC_TAGS = 154
ERR_SET_LABEL_FAILED = 155
ERR_GET_HOSTNAME = 156
ERR_GET_PRODUCT_NUM = 157
ERR_SET_PRODUCT_NUM = 158
ERR_GET_CPU_ERROR = 159
ERR_GET_UUID_TAGS = 160
ERR_GET_BOARD_TYPE = 161
ERR_GET_BOARD_ID = 162
H3559_ERROR_CODE_END = 162

EDGE_ERROR_BEGIN = 1004
ERR_SET_TIME_TIMEZONES = 1004
ERR_DONT_HAVA_TIME_TIMEZONES = 1005
ERR_SET_HOSTNAME = 1006
EDGE_ERROR_END = 1006

MiniDErrorCode = {
    130: "Failed to get chip count.",
    131: "Failed to get chip info.",
    132: "Failed to get memory size.",
    133: "Failed to get npu list or npu count is 0.",
    134: "Failed to get MiniD temperature.",
    135: "Failed to get the overall health of the device.",
    136: "Device type range is 1 to 5.",
    137: "Failed to get Davinci occupancy.",
}

H3559ErrorCode = {
    152: "Failed to get h3559 memory usage.",
    154: "Failed to get electronic tags.",
    155: "Failed to set electronic tags.",
    156: "Failed to get hostname",
    157: "Failed to get product num.",
    158: "Failed to set product num.",
    159: "Failed to get h3559 info error.",
    160: "Failed to get uuid error.",
    161: "Failed to get board type.",
    162: "Failed to get board id."
}

McuErrorCode = {
    138: "Failed to get system temperature.",
    139: "Failed to get voltage.",
    140: "Failed to get time.",
    141: "Failed to Setting mcu time.",
    142: "Failed to get board info.",
    143: "Failed to get os version.",
    144: "Failed to get kernel version.",
    145: "Failed to get uptime.",
    146: "Failed to get mcu health status.",
    147: "Failed to get mcu heating cooling status.",
    148: "Failed to get mcu power status."
}

EdgeErrorCode = {
    1004: "Set timezones failed",
    1005: "System don't hava this timezones",
    1006: "Set hostname failed"
}


def operation_log_info(request_data, msg):
    user_name = request_data.get("_User", None)
    request_ip = request_data.get("_Xip", None)
    operate_log.info("[%s@%s] %s" % (user_name, request_ip, msg))


def log_formatted_err(err_code, extra_msg=None):
    if err_code >= H3559_ERROR_CODE_BEGIN and err_code <= H3559_ERROR_CODE_END:
        msg = H3559ErrorCode.get(err_code, None)
    elif err_code >= MCU_ERROR_CODE_BEGIN and err_code <= MCU_ERROR_CODE_END:
        msg = McuErrorCode.get(err_code, None)
    elif err_code >= MINID_ERROR_CODE_BEGIN and err_code <= MINID_ERROR_CODE_BEGIN:
        msg = MiniDErrorCode.get(err_code, None)
    elif err_code >= EDGE_ERROR_BEGIN and err_code <= EDGE_ERROR_END:
        msg = EdgeErrorCode.get(err_code, None)
    else:
        return ""

    if extra_msg is not None:
        run_log.error("ERR.0%s, %s" % (str(err_code) + msg, extra_msg))
    else:
        run_log.error("ERR.0%s," % str(err_code) + msg)
    return msg


def log_and_ret_formatted_err(err_code, extra_msg=None):
    msg = log_formatted_err(err_code, extra_msg)
    if not msg:
        return [
            commMethods.CommonMethods.INTERNAL_ERROR,
            "wrong err type <%s> or sub err type <%s>" % (str(err_code), err_code)
        ]

    return [commMethods.CommonMethods.ERROR, "ERR.0%s," % str(err_code) + msg]
