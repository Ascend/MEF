# coding: utf-8
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
from typing import Dict

from common.log.logger import run_log
from fd_msg_process.common_redfish import CommonRedfish
from lib_restful_adapter import LibRESTfulAdapter


def get_digital_warranty() -> Dict:
    ret_dict = LibRESTfulAdapter.lib_restful_interface("dflc_info", "GET", None, False)
    ret = CommonRedfish.check_status_is_ok(ret_dict)
    if not ret:
        run_log.error("Get dflc info failed.")
        return {}

    ret = ret_dict.get("message")
    return {
        "digital_warranty": {
            "manufacture_date": ret.get("ManufactureDate"),
            "start_point": ret.get("StartPoint"),
            "life_span": ret.get("LifeSpan")
        }
    }
