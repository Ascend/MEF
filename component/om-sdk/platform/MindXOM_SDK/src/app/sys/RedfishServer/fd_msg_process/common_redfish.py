#!/usr/bin/python
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.

"""
功    能：Redfish Server公共变量及接口模块
"""

import glob
import json
import os
import re
import threading

from flask import make_response

from common.constants.config_constants import ConfigPathConstants
from common.file_utils import FileCheck
from common.log.logger import run_log
from common.utils.app_common_method import AppCommonMethod
from common.utils.result_base import Result
from lib_restful_adapter import LibRESTfulAdapter
from net_manager.checkers.fd_param_checker import FdConfigNetManagerInfoChecker
from net_manager.constants import NetManagerConstants
from net_manager.manager.fd_cfg_manager import FdCfgManager
from net_manager.manager.fd_cfg_manager import FdConfigData
from net_manager.manager.fd_cfg_manager import FdMsgData
from wsclient.fd_connect_check import FdConnectCheck


class CommonRedfish(object):
    PRIMARY_EMMC_LIST = (
        "mmcblk0p1", "mmcblk0p2", "mmcblk0p3", "mmcblk0p4", "mmcblk0p5", "mmcblk0p6", "mmcblk0p7"
    )
    # 遍历目录时，最大允许的文件数目
    MAX_CYCLE_NUM = 30
    # 内部服务器错误
    internalErrInfo = "Internal server error"

    SYS_CRITIC_LOCK = threading.Lock()

    # 内部服务器错误对应的HTTP错误码
    internalErrCode = AppCommonMethod.INTERNAL_ERROR

    # 创建资源成功对应的响应码
    inputParamErrCode = AppCommonMethod.ERROR

    @staticmethod
    def get_inactive_profile():
        profile_name = ""
        profile_dir = ConfigPathConstants.SYS_CONFIG_PATH
        res = FileCheck.check_path_is_exist_and_valid(profile_dir)
        if not res:
            run_log.error("%s path invalid : %s", profile_dir, res.error)
            return profile_name

        for index, prf_file in enumerate(glob.iglob(os.path.join(profile_dir, "*.prf"))):
            # /home/data/config/redfish/*.prf文件有且只能有一个存在，存在多个时profile_name赋值为空
            if index > 0:
                profile_name = ""
                run_log.warning("invalid count of inactive profiles")
                break

            profile_name = os.path.splitext(os.path.basename(prf_file))[0]

        return profile_name

    @staticmethod
    def replace_kv(adict, key, value):
        AppCommonMethod.replace_kv(adict, key, value)

    @staticmethod
    def replace_kv_list(adict, copy):
        AppCommonMethod.replace_kv_list(adict, copy)

    @staticmethod
    def add_itmes(adict, k, items):
        """
        功能描述：根据key值，增加键值对
        参数：adict JSON模板
             k Key值
             items 增加的键值对
        返回值：无
        异常描述：NA
        """
        if k in adict:
            adict[k].append(items)

    @staticmethod
    def ret_dict_is_error_proc(ret_dict, input_err_info):
        """
        功能描述：调用LibRESTfulAdapter
        中的接口获取相关信息时返回的
        retDict字典中key(status)
        为非200错误处理(生成对应的http response)
        参数：retDict字典中包含key(status)和 key(message)
        inputErrInfo为错误描述信息.
        返回值：http response
        异常描述：NA
        """

        # retDict字典中同时包含key(status)和key(message), 且retDict中的status不等于200,
        # 记录错误描述信息
        run_log.error("%s", input_err_info)

        # 生成对应的http response
        res = make_response(json.dumps(ret_dict), ret_dict["status"])
        return res

    @staticmethod
    def ret_dict_is_exception_proc(input_err_info):
        """
        功能描述：调用LibRESTfulAdapter中的接口获取相关信息时返回
        的retDict字典中未同时包含key(status)
        和key(message)的异常处理(生成对应的http response)
        参数：retDict字典中正常情况下包含key(status)
        和 key(message)
        inputErrInfo为错误描述信息.
        """

        # retDict字典中未同时包含key(status)和key(message),记录inputErrInfo为错误描述信息
        run_log.error("%s", input_err_info)

        ret = {"status": CommonRedfish.internalErrCode, "message": CommonRedfish.internalErrInfo}

        # 生成对应的http response
        res = make_response(json.dumps(ret), CommonRedfish.internalErrCode)
        return res

    @staticmethod
    def update_json_of_list(resp_json, ret_dict, input_err_info):
        """
        功能描述：调用LibRESTfulAdapter中的接口获取相关信息时
        返回的retDict字典中包含key(status)
        和key(message),若retDict数据正常则
        更新resp_json的对应数据
        参数：resp_json为对应的资源模板,
        retDict字典中正常情况下包含key(status)和 key(message)
        inputErrInfo为错误描述信息.
        返回值：http response
        异常描述：NA
        """
        if ret_dict is None or not isinstance(ret_dict, dict):
            # retDict字典中未同时包含key(status)
            # 和key(message)的异常处理(生成对应的http response)
            return CommonRedfish.ret_dict_is_exception_proc(input_err_info)

        # retDict字典中是否包含key(status)
        has_status_key = "status" in ret_dict

        # retDict字典中是否包含key(message)
        has_message_key = "message" in ret_dict

        # 检测retDict字典中是否同时包含key(status)和key(message)
        if has_status_key and has_message_key:
            # 检测retDict字典的key(status)是否为HTTP OK
            if ret_dict["status"] == 200:
                # 更新resp_json的对应数据
                CommonRedfish.replace_kv_list(resp_json, ret_dict["message"])
                return json.dumps(resp_json)
            else:
                # retDict字典中key(status)为非200错误处理
                # (生成对应的http response)
                return CommonRedfish.ret_dict_is_error_proc(ret_dict, input_err_info)
        else:
            # retDict字典中未同时包含key(status)
            # 和key(message)的异常处理(生成对应的http response)
            return CommonRedfish.ret_dict_is_exception_proc(input_err_info)

    @staticmethod
    def update_json_of_mem_count_and_odata_id(resp_json, ret_dict, input_err_info, prifix_id=""):
        '''
                         功能描述：调用LibAdapter中的接口获取相关信息时
                         返回的retDict字典中包含key(status)
                         和key(message),若retDict数据正常则
                         更新resp_json中的Members@odata.count
                         和 Members的@odata.id
                         参数：resp_json为对应的资源模板,
                  retDict字典中正常情况下包含key(status)和 key(message)
                  inputErrInfo为错误描述信息.
                  prifixId为  Members的@odata.id前缀ID
                         返回值：http response
                         异常描述：NA
        '''
        if ret_dict is None or not isinstance(ret_dict, dict):
            # retDict字典中未同时包含key(status)
            # 和key(message)的异常处理(生成对应的http response)
            return CommonRedfish.ret_dict_is_exception_proc(input_err_info)

        # retDict字典中是否包含key(status)
        has_status_key = "status" in ret_dict

        # retDict字典中是否包含key(message)
        has_message_key = "message" in ret_dict

        # 检测retDict字典中是否同时包含key(status)和key(message)
        if has_status_key and has_message_key:
            # 检测retDict字典的key(status)是否为HTTP OK
            if ret_dict["status"] == 200:
                # 更新resp_json的对应数据
                value_list = ret_dict["message"]
                if len(value_list) == 0:
                    CommonRedfish.replace_kv(resp_json, "Members@odata.count", 0)
                    del resp_json['Members'][0]['@odata.id']
                    return json.dumps(resp_json)
                else:
                    CommonRedfish.replace_kv(resp_json, "Members@odata.count", len(value_list))
                    member = resp_json['Members'][0]['@odata.id']
                    if prifix_id != "":
                        member = member.replace('oDataPrefixID', prifix_id)

                    for i, _ in enumerate(value_list):
                        if i == 0:
                            resp_json['Members'][0]['@odata.id'] = member. \
                                replace('entityID', value_list[i])
                        else:
                            url = member.replace('entityID', value_list[i])
                            CommonRedfish.add_itmes(resp_json, 'Members', {'@odata.id': url})

                    return json.dumps(resp_json)
            else:
                # retDict字典中key(status)
                # 为非200错误处理(生成对应的http response)
                return CommonRedfish.ret_dict_is_error_proc(ret_dict, input_err_info)
        else:
            # retDict字典中未同时包含key(status)
            # 和key(message)的异常处理(生成对应的http response)
            return CommonRedfish.ret_dict_is_exception_proc(input_err_info)

    @staticmethod
    def is_legal_of_input_param(input_param):
        """
        功能描述：检测inputParam 是否符合“A-Z,a-z,0-9_:.”
        参数：inputParam 输入参数
        返回值：inputParam 输入参数合法,返回True,否则返回False
        异常描述：NA
        """
        reg_str = "^[a-z0-9A-Z_:.]+$"
        pattern_str = re.compile(reg_str)
        # redfish接口限制..使用
        if ".." in input_param:
            return False
        if pattern_str.fullmatch(input_param):
            return True
        else:
            run_log.error("Fail because of inputParam: %s", input_param)
            return False

    @staticmethod
    def check_status_is_ok(ret_dict):
        return AppCommonMethod.check_status_is_ok(ret_dict)

    @staticmethod
    def get_dev_site_name_by_location(dev_location):
        """
        根据设备位置信息获取设备的名称
        :param dev_location: 存储设备在/run/formated_hw_1.info文件中的位置信息
        :return: 经过转换的存储名称
        """
        ret = LibRESTfulAdapter.lib_restful_interface("Partition", "GET", None, False, "", dev_location)
        is_ok = CommonRedfish.check_status_is_ok(ret)
        if not is_ok:
            return Result(False, err_msg="function call failed")

        site_name = ret.get("message", {}).get("DeviceName")
        if not site_name:
            return Result(False, err_msg="result is invalid")

        return Result(result=True, data=site_name)

    @staticmethod
    def modify_host_name(hostname):
        request_data_dict = {
            "HostName": hostname,
            "_User": "FD",
            "_Xip": FdCfgManager.get_cur_fd_ip()
        }
        # 根据调用接口LibAdapter返回的retDict字典，生成对应的json,返回给iBMC
        ret_dict = LibRESTfulAdapter.lib_restful_interface("System", "PATCH", request_data_dict, False)
        ret = LibRESTfulAdapter.check_status_is_ok(ret_dict)
        if not ret:
            err_msg = ret_dict.get("message")
            return Result(result=False, err_msg=err_msg)

        return Result(result=True)

    @staticmethod
    def check_net_manager_info(msg: FdMsgData) -> list:
        """
        FD下发的配置网管消息，检查一机一密参数
        :param msg: FdMsgData对象
        :return: list
        """
        payload = msg.content
        ret = FdConfigNetManagerInfoChecker().check(payload)
        if not ret:
            run_log.error("invalid parameters: %s", ret.reason)
            return [-1, f"invalid parameters: {ret.reason}"]

        fd_cfg = FdCfgManager()
        # 等待一机一密阶段、更新认证信息场景下不允许下发的账号与初始账号相同
        if not payload.get("address") and payload.get("account") == NetManagerConstants.FD_INIT_ACCOUNT:
            run_log.error("invalid parameters when updating device account info.")
            return [-1, "invalid parameters when updating device account info"]

        sys_info = fd_cfg.get_sys_info()
        if not sys_info:
            run_log.error("check net manager info failed, %s", sys_info.error)
            return [-1, f"check net manager info failed, {sys_info.error}"]

        if payload.get("address"):
            fd_cfg.net_manager.ip = payload.get("address")

        cfg: FdConfigData = FdConfigData.from_dict(net=fd_cfg.net_manager, sys=sys_info.data)
        cfg.cloud_user = payload.get("account")
        cfg.cloud_pwd = payload.get("password")
        ret = FdConnectCheck.connect_test(cfg)
        if not ret:
            run_log.error("connect fd test failed: %s", ret.error)
            return [-1, f"connect fd test failed: {ret.error}"]

        run_log.info("connect fd test success")
        return [0, ""]
