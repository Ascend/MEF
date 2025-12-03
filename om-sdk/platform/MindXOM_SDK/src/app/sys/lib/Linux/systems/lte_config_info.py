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

import configparser
import os
import threading

import common.common_methods as commMethods
from common.checkers.param_checker import LTEConfigInfoChecker
from common.constants.error_codes import CommonErrorCodes
from common.file_utils import FileCheck
from common.log.logger import run_log
from devm.device_mgr import DEVM
from devm.exception import DeviceManagerError

IBMA_EDGE_SERVICE_PATH = '/home/data/ies/ibma_edge_service.ini'
APN_SET_RESULT_SUCCESS = [0, 'Set APN success']
APN_SET_RESULT_OTHER = [-1, 'Setting LTE APN failed']


class LteConfigInfo:
    """
    功能描述：Lte配置信息
    接口：NA
    修改记录：2019-9-25 创建
    """
    LTE_CONFIG_LOCK = threading.Lock()

    def __init__(self):
        """
         功能描述：初始化函数
         参数：
         返回值：无
         异常描述：NA
        """
        # apn
        self.apn_name = None
        # 拨号的用户名
        self.apn_user = None
        # 身份验证类型：0:NONE, 1：PAP, 2：CHAP, 3：PAP or CHAP
        self.auth_type = "0"
        # 0:华为模块，1：移远模块
        self.mode_type = None

    @staticmethod
    def check_apn_name(mode_type, app_name_len):
        # A2仅支持 EC200T 和 RM500U 的模块类型，即 mode_type 取值为 2 或 3.
        # 为适配 FD，apn_name 最大长度均限制为 39.
        if app_name_len > 39 or mode_type not in (2, 3):
            return [commMethods.CommonMethods.ERROR, CommonErrorCodes.ERROR_PARAMETER_INVALID.messageKey]
        return [commMethods.CommonMethods.OK, '']

    @staticmethod
    def check_apn_user(apn_user, apn_passwd):
        # 满足用户名、密码长度不超过64个字符
        exists = apn_user and apn_passwd
        if exists and (len(apn_user) > 64 or len(apn_passwd) > 64):
            run_log.error("Parameter is invalid.")
            return [commMethods.CommonMethods.ERROR, CommonErrorCodes.ERROR_PARAMETER_INVALID.messageKey]
        return [commMethods.CommonMethods.OK, '']

    def get_all_info(self):
        # 校验配置文件路径是否存在及合法性
        ret = FileCheck.check_path_is_exist_and_valid(IBMA_EDGE_SERVICE_PATH)
        if not ret:
            run_log.error(ret.error)
            return [commMethods.CommonMethods.ERROR, 'Lte config reads failed']

        # 若路径合法，从配置文件中获取auth_type
        config_raw = configparser.RawConfigParser()
        config_raw.read(IBMA_EDGE_SERVICE_PATH)
        section_list = config_raw.sections()
        if 'lte_apn' in section_list:
            self.auth_type = config_raw.get('lte_apn', 'auth_type')  # auth_type，存放在文件ibma_edge_service.ini中

        # 底层接口中获取apn_name、apn_user
        try:
            apn_info = DEVM.get_device('Wireless_Module').get_attribute('apn_info')
        except DeviceManagerError:
            run_log.warning('get lte apn info failed')
            return [commMethods.CommonMethods.OK, ""]
        self.apn_name = apn_info.get('apn_name')
        self.apn_user = apn_info.get('apn_user')

        try:
            self.mode_type = DEVM.get_device('Wireless_Module').get_attribute('type')
        except DeviceManagerError:
            run_log.warning('get lte type failed')
            return [commMethods.CommonMethods.ERROR, "get lte type failed"]
        return [commMethods.CommonMethods.OK, 'get lte apn info success']

    def patch_request(self, request_dict, name=None, item=None):
        """LTE APN 的配置"""
        if LteConfigInfo.LTE_CONFIG_LOCK.locked():
            run_log.warning("Lte config is busy")
            return [commMethods.CommonMethods.ERROR, 'Lte config is busy']

        with LteConfigInfo.LTE_CONFIG_LOCK:
            try:
                lte_present = DEVM.get_device('Wireless_Module').get_attribute('present')
            except DeviceManagerError:
                run_log.error('get lte present failed')
                return [commMethods.CommonMethods.ERROR, "get lte present failed"]

            if not lte_present:
                return [commMethods.CommonMethods.ERROR, "Lte does not exist."]

            # 在配置文件中修改参数
            # 校验配置文件路径是否存在及合法性
            ret = FileCheck.check_path_is_exist_and_valid(IBMA_EDGE_SERVICE_PATH)
            if not ret:
                run_log.error("file path %s invalid: %s", IBMA_EDGE_SERVICE_PATH, ret.error)
                return [commMethods.CommonMethods.ERROR, 'Lte config reads failed']

            config_raw = configparser.RawConfigParser()  # 创建对象
            config_raw.read(IBMA_EDGE_SERVICE_PATH)
            section_list = config_raw.sections()
            # 判断LTE的数据流量开关是否为开，如果为打开的情况不允许操作APN
            state_lte_file = config_raw.getboolean('lte_state', 'state_lte')
            state_data_file = config_raw.getboolean('lte_state', 'state_data')

            # 当移动打开，并且sim卡在位时不让修改apn。
            try:
                lte_sim_state = DEVM.get_device('Wireless_Module').get_attribute('sim_state')
            except DeviceManagerError:
                run_log.error('get lte sim state failed')
                return [commMethods.CommonMethods.ERROR, 'get lte sim state failed']

            if state_lte_file and state_data_file and lte_sim_state:
                return [commMethods.CommonMethods.ERROR, 'ERR.007,%s' % 'Lte data is open, can not operate APN!']
            # 对request_dict进行参数校验
            try:
                check_ret = LTEConfigInfoChecker().check(request_dict)
            except Exception:
                run_log.error("Parameter is invalid.")
                return [commMethods.CommonMethods.ERROR, 'Parameter is invalid.']
            if not check_ret.success:
                run_log.error("Parameter is invalid.")
                return [commMethods.CommonMethods.ERROR, 'Parameter is invalid.']

            # 获取输入参数
            apn_name = request_dict.get("apn_name", None)
            apn_user = request_dict.get("apn_user", None)
            apn_passwd = request_dict.get("apn_passwd", None)
            auth_type = request_dict.get("auth_type", None)

            result = LteConfigInfo.check_apn_user(apn_user, apn_passwd)
            if result[0] != commMethods.CommonMethods.OK:
                return result

            # 根据LTE设备类型检查apn_name长度
            try:
                lte_mode_type = DEVM.get_device('Wireless_Module').get_attribute('type')
            except DeviceManagerError:
                run_log.warning('get lte type failed')
                return [commMethods.CommonMethods.ERROR, 'get lte type failed']

            result = LteConfigInfo.check_apn_name(lte_mode_type, len(apn_name))
            if result[0] != commMethods.CommonMethods.OK:
                return result

            # 调用底层c库保存apn信息
            try:
                DEVM.get_device('Wireless_Module').set_attribute(
                    'apn_info', {'apn_name': apn_name,
                                 'apn_user': apn_user,
                                 'apn_password': apn_passwd,
                                 'auth_type': int(auth_type)})
                run_log.info('Setting LTE APN success')
            except (DeviceManagerError, ValueError):
                run_log.warning('Setting LTE APN failed')
                return [commMethods.CommonMethods.ERROR, 'ERR.006,%s' % APN_SET_RESULT_OTHER[1]]

            if 'lte_apn' not in section_list:
                config_raw.add_section('lte_apn')
            config_raw.set('lte_apn', 'auth_type', str(auth_type))
            try:
                FileCheck.check_is_link_exception(IBMA_EDGE_SERVICE_PATH)
                with os.fdopen(os.open(
                        IBMA_EDGE_SERVICE_PATH, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o640), "w+") as file:
                    config_raw.write(file)
            except Exception as err:
                run_log.error("failed:{}".format(err))
            return [commMethods.CommonMethods.OK, ]
