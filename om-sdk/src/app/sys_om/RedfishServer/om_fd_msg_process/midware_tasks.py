# coding: utf-8
# Copyright (C) 2023.Huawei Technologies Co., Ltd. All rights reserved.
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
