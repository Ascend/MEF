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

"""
功    能：system service的工具函數
"""
import time

from common.log.logger import run_log
from lib_restful_adapter import LibRESTfulAdapter


def make_error_dict(input_info, input_status):
    return {"status": input_status, "message": input_info}


def lib_rest_check_status(request_data_dict, ret_dict):
    ret_status_is_ok = LibRESTfulAdapter.check_status_is_ok(ret_dict)
    if ret_status_is_ok:
        date_time_local_offset = request_data_dict.get("DateTimeLocalOffset", None)
        if date_time_local_offset:
            # 根据环境变量 重新初始化该进程的时间相关设置。
            time.tzset()
        return True
    return False


def modify_null_tag_to_default(request_dict: dict):
    addresses = request_dict.get("IPv4Addresses", [])
    if not isinstance(addresses, list):
        run_log.error(f"Modify null tag to web failed")
        return request_dict

    modified_addresses = []
    for address in addresses:
        if not isinstance(address, dict):
            run_log.error(f"Modify null tag to web failed")
            return request_dict
        address["Tag"] = address.get("Tag") or "default"
        modified_addresses.append(address)
    return {"IPv4Addresses": modified_addresses}
