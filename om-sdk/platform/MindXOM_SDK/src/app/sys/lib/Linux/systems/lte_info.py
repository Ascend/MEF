#!/usr/bin/python
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

import configparser
import os
import re
import threading
import time
from typing import List
from typing import Optional
from typing import Union

import common.common_methods as commMethods
from common.checkers.param_checker import FdIpChecker
from common.constants.error_codes import FileErrorCodes
from common.exception.biz_exception import BizException
from common.file_utils import FileCheck
from common.file_utils import FileCopy
from common.file_utils import FileUtils
from common.init_cmd import cmd_constants
from common.log.logger import operate_log
from common.log.logger import run_log
from common.utils.common_check import CommonCheck
from common.utils.exec_cmd import ExecCmd
from common.utils.policy_based_route import PolicyBasedRouting
from common.utils.policy_based_route import RouteCfgError
from devm.device_mgr import DEVM
from devm.exception import DeviceManagerError
from devm.exception import DeviceNotExistError


IBMA_EDGE_SERVICE_PATH = '/home/data/ies/ibma_edge_service.ini'

LTE_DIAL_UP_AUTO = 0  # 自动拨号
LTE_DIAL_UP_MANUAL = 1  # 手动拨号

SIGNAL_STRENGTH = {
    0: ['2G', 'ACT_GSM'],
    1: ['2G', 'ACT_GSM_COMPACT'],
    2: ['3G', 'ACT_UTRAN'],
    3: ['2G', 'ACT_EGPRS'],
    4: ['3G', 'ACT_HSDPA'],
    5: ['3G', 'ACT_HSUPA'],
    6: ['3G', 'ACT_HSDPA_HSUPA'],
    7: ['4G', 'ACT_LTE'],
    8: ['-1', 'ACT_MAX'],
    10: ['5G', 'ACT_5GCN_E_UTRAN'],
    11: ['5G', 'ACT_5GCN_NR'],
    12: ['5G', 'ACT_NG_RAN'],
    13: ['5G', 'ACT_E_UTRAN_NR'],
    15: ['4G', 'ACT_HSPA'],
}


def set_lte_dial_up_authtype():
    """有0、1、2、3"""
    ret = FileCheck.check_path_is_exist_and_valid(IBMA_EDGE_SERVICE_PATH)
    if not ret:
        run_log.error(ret.error)
        return

    config_raw = configparser.RawConfigParser()
    config_raw.read(IBMA_EDGE_SERVICE_PATH)
    section_list = config_raw.sections()
    if 'lte_apn' not in section_list:
        return

    auth_type = config_raw.get('lte_apn', 'auth_type')  # auth_type，存放在文件ibma_edge_service.ini中
    if auth_type is None or len(auth_type) == 0:
        auth_type = '255'  # 当没有配置APN时，下发255，底层用

    if not auth_type.isdigit():
        run_log.error("set lte dial up auth type failed !%s", auth_type)
        return

    DEVM.get_device('Wireless_Module').set_attribute('dial_up_authtype', int(auth_type))
    run_log.info("set lte dial up auth type info!")


class LteInfo:
    """
    功能描述：Lte配置信息
    接口：NA
    修改记录：2019-9-25 创建
    """
    LTE_LOCK = threading.Lock()
    STATE_INIT = False

    def __init__(self):
        """
        功能描述：初始化函数
        参数：
        返回值：无
        异常描述：NA
        """
        self.default_gateway = False  # 判断当前是否有eth0与eth1配置的默认网关，如果有为true，否则就为false
        self.lte_enable = False  # lte是否使能，默认值为false。True 使能  false 不可用
        self.sim_exist = False  # sim卡是否在位。true：在位，false：不在位
        self.state_data = None  # 移动数据的开关状态 true 打开 false 关闭
        self.state_lte = None  # lte的开关状态 true 打开 false 关闭
        self.network_signal_level = None  # 信号的强度 取值范围：0-5级，取值0时，代表网络断开，这时network_type的取值为null
        self.network_type = None  # 网络状态取值范围：4g、3g、2g
        self.ip_addr = None  # 网络状态取值范围：4g、3g、2g

    @staticmethod
    def set_lte_state(state_lte, state_lte_file, state_data, username, request_ip):
        if state_lte != state_lte_file:
            if state_lte:  # Lte开关打开
                try:
                    DEVM.get_device('Wireless_Module').set_attribute('dial_up_type', LTE_DIAL_UP_MANUAL)
                    DEVM.get_device('Wireless_Module').set_attribute('switch', True)
                except DeviceManagerError:
                    operate_log.info(f"[{username}@{request_ip}] lte switch on failed.")
                    return [commMethods.CommonMethods.ERROR, 'ERR.001,LTE open failed!']
                operate_log.info(f"[{username}@{request_ip}] lte switch on success.")
                time.sleep(5)
            else:  # Lte关闭，同时要关闭数据流量
                try:
                    DEVM.get_device('Wireless_Module').set_attribute('switch', False)
                except DeviceManagerError:
                    operate_log.info(f"[{username}@{request_ip}] lte switch off failed.")
                    return [commMethods.CommonMethods.ERROR, 'ERR.002,LTE close failed!']
                operate_log.info(f"[{username}@{request_ip}] lte switch off success.")
        if state_lte:
            if state_data:  # 打开数据流量
                try:
                    DEVM.get_device('Wireless_Module').set_attribute('dial_up_type', LTE_DIAL_UP_MANUAL)
                    DEVM.get_device('Wireless_Module').set_attribute('data', True)
                except DeviceManagerError:
                    operate_log.info(f"[{username}@{request_ip}] lte_data switch on failed.")
                    return [commMethods.CommonMethods.ERROR, 'ERR.003,data open failed!']
                operate_log.info(f"[{username}@{request_ip}] lte_data switch on success.")
            else:
                try:
                    DEVM.get_device('Wireless_Module').set_attribute('data', False)
                except DeviceManagerError:
                    operate_log.info(f"[{username}@{request_ip}] lte_data switch off failed.")
                    return [commMethods.CommonMethods.ERROR, 'ERR.004,data close failed!']
                operate_log.info(f"[{username}@{request_ip}] lte_data switch off success.")
        return [commMethods.CommonMethods.OK, '']

    @staticmethod
    def modify_lte(state_lte, state_data):
        # 策略路由配置可能修改了配置文件，故此处需要重新加载，否则数据会被覆盖不一致了
        try:
            config_raw = FileUtils.get_config_parser(IBMA_EDGE_SERVICE_PATH)
        except BizException as error:
            run_log.error("get lte config info from file failed: %s", error)
            return

        config_raw.set('lte_state', 'state_lte', str(state_lte))  # lte的开关状态，在python端存放
        config_raw.set('lte_state', 'state_data', str(state_data))  # 移动数据的开关状态，在python端存放
        flags = os.O_WRONLY | os.O_CREAT | os.O_TRUNC
        try:
            FileCheck.check_is_link_exception(IBMA_EDGE_SERVICE_PATH)
            with os.fdopen(os.open(IBMA_EDGE_SERVICE_PATH, flags, 0o640), "w+") as file:
                config_raw.write(file)
        except Exception as err:
            run_log.error("modify lte config file failed: %s", err)

    @staticmethod
    def get_ip_address():
        result = ''
        # 当数据流量为打开状态时，获取ip地址
        shell_str = "%s usb0 | grep 'inet ' | awk -F'[ :]+' '{print $3}'" % cmd_constants.OS_CMD_IFCONFIG
        ret = ExecCmd.exec_cmd_use_pipe_symbol(shell_str, wait=10)
        if isinstance(ret, list) and ret[0] == 0:
            result = ret[1].strip()
        return result

    @staticmethod
    def lte_base_init():
        try:
            if not os.path.exists(IBMA_EDGE_SERVICE_PATH):
                inifile = '/usr/local/mindx/MindXOM/software/ibma/lib/Linux/config/ibma_edge_service.ini'
                ret = FileCopy.copy_file(inifile, IBMA_EDGE_SERVICE_PATH)
                if not ret:
                    run_log.error("Copy file error: %s", ret.error)
                    return False
        except Exception as err:
            run_log.error("Init file operation error: %s.", err)

        try:
            if not DEVM.get_device('Wireless_Module').get_attribute('present'):
                return LteInfo.STATE_INIT

            # 校验配置文件路径是否存在及合法性
            ret = FileCheck.check_path_is_exist_and_valid(IBMA_EDGE_SERVICE_PATH)
            if not ret:
                run_log.error(ret.error)
                return LteInfo.STATE_INIT

            LteInfo.STATE_INIT = True
            # 从配置文件中查询state_lte、state_data的取值
            config_raw = configparser.ConfigParser()  # 创建对象
            config_raw.read(IBMA_EDGE_SERVICE_PATH)
            # lte的开关状态，存放在文件ibma_edge_service.ini中
            state_lte = config_raw.getboolean('lte_state', 'state_lte')
            # 移动数据的开关状态，存放在文件ibma_edge_service.ini中
            state_data = config_raw.getboolean('lte_state', 'state_data')

            if not state_lte or not DEVM.get_device('Wireless_Module').get_attribute('sim_state'):
                return LteInfo.STATE_INIT

            # lte开关打开 并且 sim卡在位
            if LteInfo._get_ethernet_default_gateway():
                LteInfo().modify_lte(False, False)
                run_log.info(
                    "Unable to open settings, close Lte and data due to configuration default gateway")
                return LteInfo.STATE_INIT

            try:
                DEVM.get_device('Wireless_Module').set_attribute('dial_up_type', LTE_DIAL_UP_AUTO)
                DEVM.get_device('Wireless_Module').set_attribute('switch', True)
            except DeviceManagerError:
                run_log.error("Init open lte switch failed.")
                return False
            time.sleep(5)

            if state_data:
                try:
                    DEVM.get_device('Wireless_Module').set_attribute('dial_up_type', LTE_DIAL_UP_AUTO)
                    set_lte_dial_up_authtype()
                    DEVM.get_device('Wireless_Module').set_attribute('data', True)
                except DeviceManagerError:
                    run_log.error("Init open lte data switch failed.")
                    return False
            else:
                try:
                    DEVM.get_device('Wireless_Module').set_attribute('data', False)
                except DeviceManagerError:
                    run_log.error("Init close the lte data switch failed.")
                    return False
            return LteInfo.STATE_INIT
        except Exception as ex:
            run_log.error("Init set lte error: %s.", ex)
            return False

    @staticmethod
    def _get_ethernet_default_gateway() -> List[str]:
        for eth_name in "eth0", "eth1":
            default_gateway = LteInfo._get_ethernet_gateway(eth_name)
            if default_gateway:
                return [eth_name, default_gateway]

        return []

    @staticmethod
    def _get_ethernet_gateway(eth_name):
        ret = ExecCmd.exec_cmd_use_pipe_symbol(f"{cmd_constants.OS_CMD_IP} route | grep '^default'")
        if ret[0] != 0:
            return ""

        match_res = re.findall(rf"default via ([\d.]+) dev {eth_name}", ret[1], re.M)
        return match_res[0] if match_res else ""

    @staticmethod
    def _config_policy_routing(fd_server_ip: Optional[str], web_request_ip: Optional[str]) -> List[Union[int, str]]:
        """
        对dest_ips中的ip地址配置策略路由。
        配置步骤如下：
            选择空的路由表
            配置一条到dest_ipv4的路由规则到该路由表
            新增该路由表的策略路由规则
            路由表内容保存至文件并持久化
            策略路由规则保存至文件并持久化
            检查有线网络默认路由是否存在，若有，则删除该默认路由
            检查默认网关配置项，如存在，则需要置空，并持久化到OM的配置文件中，后续切回有线可以恢复
        :param fd_server_ip: FD网管服务端IP地址
        :param web_request_ip: WEB近端访问客户端IP地址
        :return: List[err_code，err_msg]
        """
        eth_gw = LteInfo._get_ethernet_default_gateway()
        # 没有默认网关则不配置策略路由
        if not eth_gw:
            run_log.info("ethernet default gateway ip is empty, skip configuring policy based routing")
            return [commMethods.CommonMethods.OK, ""]

        route_table_id = ""
        try:
            route_table_id = FileUtils.get_option(IBMA_EDGE_SERVICE_PATH, "policy_route", "route_table_id")
        except BizException as error:
            # 不影响配置策略路由的逻辑，只记录日志
            run_log.warning("get route table id from config file failed: %s", error)

        eth_name, eth_gw_ip = eth_gw
        try:
            router = PolicyBasedRouting(eth_name, eth_gw_ip, route_table_id)
        except RouteCfgError as error:
            run_log.error("init PolicyBasedRouting caught exception: %s", error)
            return [commMethods.CommonMethods.ERROR, "init PolicyBasedRouting caught exception"]

        dest_ips = [dest_ip for dest_ip in [fd_server_ip, web_request_ip] if dest_ip]
        if not router.config_policy_route(dest_ips):
            run_log.error("config policy route to dest ips %s via gw %s failed", dest_ips, eth_gw_ip)
            return [commMethods.CommonMethods.ERROR, "config policy routes failed before open LTE"]

        run_log.info("config policy routes success before open LTE")
        return [commMethods.CommonMethods.OK, ""]

    @staticmethod
    def _cancel_policy_routing(fd_server_ip: Optional[str], web_request_ip: Optional[str]):
        """
        取消对dest_ips中的ip地址的策略路由。
        配置步骤如下：
            从配置文件中读取策略路由的相关信息
            恢复默认网关配置项
            检查有线网络默认路由是否存在，若无，则添加该默认路由
            取消策略路由规则文件持久化
            取消路由表内容文件持久化
            删除路由表的策略路由规则
            删除配置到dest_ipv4的路由规则到该路由表
            清空路由表
        :param fd_server_ip: FD网管服务端IP地址
        :param web_request_ip: WEB近端访问客户端IP地址
        :return: List[err_code，err_msg]
        """
        try:
            policy_route_options = FileUtils.get_option_list(IBMA_EDGE_SERVICE_PATH, "policy_route")
        except BizException as error:
            if error.error_code != FileErrorCodes.ERROR_FILE_NO_SECTION:
                run_log.warning("get route table id from config file failed: %s", error)
            # 读取参数失败可以不取消配置，返回OK
            return [commMethods.CommonMethods.OK, ""]

        eth_name = policy_route_options.get("eth_name")
        eth_gw_ip = policy_route_options.get("eth_gw_ip")
        route_table_id = policy_route_options.get("route_table_id")
        if not eth_name or not route_table_id or not eth_gw_ip:
            run_log.info("no policy route info, no need rollback")
            return [commMethods.CommonMethods.OK, ""]

        try:
            router = PolicyBasedRouting(eth_name, eth_gw_ip, route_table_id)
        except RouteCfgError as error:
            run_log.error("init PolicyBasedRouting caught exception: %s", error)
            return [commMethods.CommonMethods.ERROR, "init PolicyBasedRouting caught exception"]

        dest_ips = [dest_ip for dest_ip in [fd_server_ip, web_request_ip] if dest_ip]
        router.cancel_policy_route(dest_ips)
        return [commMethods.CommonMethods.OK, ""]

    @staticmethod
    def _check_input_params(username, request_ip, fd_server_ip, state_data, state_lte):
        operator_check = CommonCheck.check_operator(username, request_ip)
        if not operator_check:
            run_log.error("The operator is illegal, %s", operator_check.error)
            return [commMethods.CommonMethods.ERROR, "username or request_ip wrong format"]

        if fd_server_ip:
            ret = FdIpChecker("fd_server_ip").check_dict({"fd_server_ip": fd_server_ip})
            if not ret:
                run_log.error("The fd_server_ip is invalid: %s", ret.reason)
                return [commMethods.CommonMethods.ERROR, "The fd_server_ip is invalid"]

        if not isinstance(state_data, bool):
            run_log.error("LTE state_data wrong format")
            return [commMethods.CommonMethods.ERROR, "LTE state_data wrong format"]

        if not isinstance(state_lte, bool):
            run_log.error("LTE state_lte wrong format")
            return [commMethods.CommonMethods.ERROR, "LTE state_lte wrong format"]

        if not state_lte and state_data:
            run_log.error("Could not open data when lte is closed")
            return [commMethods.CommonMethods.ERROR, "Could not open data when lte is closed"]
        return [commMethods.CommonMethods.OK, ""]

    def get_all_info(self):
        """获取Lte的状态"""
        # 当前是否有eth0、eth1的默认网关配置
        if self._get_ethernet_default_gateway():
            self.default_gateway = True

        # 如果当前没有Lte模组的设备,则直接返回，避免日志刷屏
        if not DEVM.has_device("Wireless_Module"):
            return

        # 如果Lte不在位直接返回
        try:
            if not DEVM.get_device('Wireless_Module').get_attribute('present'):
                return
        except DeviceNotExistError:
            run_log.error("device Wireless_Module not exists")
            return

        self.lte_enable = True  # Lte在位，则lte_enable=True

        # 判断初始化lte是否成功，如果没有成功重新初始化，如果还是失败，则直接返回
        if LteInfo.STATE_INIT is False:
            if LteInfo.lte_base_init() is not True:
                return

        if DEVM.get_device('Wireless_Module').get_attribute('sim_state') is False:  # sim卡是否在位
            return
        self.sim_exist = True  # sim卡在位。则sim_exist=True

        # 校验配置文件路径是否存在及合法性
        ret = FileCheck.check_path_is_exist_and_valid(IBMA_EDGE_SERVICE_PATH)
        if not ret:
            run_log.error("file path %s invalid: %s", IBMA_EDGE_SERVICE_PATH, ret.error)
            return

        # 若路径合法，从配置文件中读取state_lte、state_data状态
        config_raw = configparser.ConfigParser()  # 创建对象
        config_raw.read(IBMA_EDGE_SERVICE_PATH)
        self.state_lte = config_raw.getboolean('lte_state', 'state_lte')  # lte的开关状态，存放在文件ibma_edge_service.ini中
        self.state_data = config_raw.getboolean('lte_state', 'state_data')  # 移动数据的开关状态，存放在文件ibma_edge_service.ini中
        if self.state_lte is False:  # Lte为开状态，查询信号强度与信号类型
            return
        # 当数据流量为打开状态，才去获取ip地址，有了ip地址才能上网
        if self.state_data:
            ip_address = LteInfo.get_ip_address()
            if len(ip_address) != 0:
                self.ip_addr = ip_address
        # 获取信号强度与制式
        try:
            signal_info = DEVM.get_device('Wireless_Module').get_attribute('signal_info')
            # 信号的强度 取值范围：0-5级，取值0时，代表网络断开，这时network_type的取值为null
            self.network_signal_level = signal_info.get('signal_strength', 0)
            if self.network_signal_level:  # 当信号强度不为0时，才判断信号类型
                type_ = SIGNAL_STRENGTH.get(signal_info.get('signal_type'))
                if type_:
                    self.network_type = type_[0]  # 信号类型取值范围：4g、3g、2g
        except DeviceManagerError:
            run_log.error('get signal strength failed.')
            return

    def patch_request(self, request_dict):
        """LTE status配置"""
        if LteInfo.LTE_LOCK.locked():
            run_log.warning("Lte modify is busy")
            return [commMethods.CommonMethods.ERROR, 'Lte modify is busy']
        with LteInfo.LTE_LOCK:
            try:
                if DEVM.get_device('Wireless_Module').get_attribute('sim_state') is False:  # sim卡是否在位
                    return [commMethods.CommonMethods.ERROR, 'ERR.005,sim card is not exist!']
            except DeviceNotExistError as err:
                run_log.error("Get sim_state failed, reason: %s", err)
                return [commMethods.CommonMethods.ERROR, "Device Wireless_Module not exists!"]
            # 校验输入的参数
            state_data = request_dict.get("state_data")
            state_lte = request_dict.get("state_lte")
            username = request_dict.get("_User")
            request_ip = request_dict.get("_Xip")
            fd_server_ip = request_dict.get("fd_server_ip")
            ret = self._check_input_params(username, request_ip, fd_server_ip, state_data, state_lte)
            if ret[0] != commMethods.CommonMethods.OK:
                return ret

            # 根据前端传过来的state_lte、state_data对lte进行执行
            # 校验配置文件路径是否存在及合法性
            ret = FileCheck.check_path_is_exist_and_valid(IBMA_EDGE_SERVICE_PATH)
            if not ret:
                run_log.error("file path invalid: %s", ret.error)
                return [commMethods.CommonMethods.ERROR, 'Lte config reads failed']

            config_raw = configparser.ConfigParser()  # 创建对象
            config_raw.read(IBMA_EDGE_SERVICE_PATH)
            state_lte_file = config_raw.getboolean(
                'lte_state', 'state_lte')  # lte的开关状态，存放在文件ibma_edge_service.ini中
            switch_to_lte = state_lte and state_data
            # 从有线切换到无线网络之前，需要先配置策略路由，防脱管
            if switch_to_lte:
                ret = self._config_policy_routing(fd_server_ip, request_ip)
                if ret[0] != commMethods.CommonMethods.OK:
                    run_log.error("config policy routes failed before open LTE: %s", ret[1])
                    return ret

            ret = LteInfo.set_lte_state(state_lte, state_lte_file, state_data, username, request_ip)
            if ret[0] == commMethods.CommonMethods.ERROR:
                return ret
            LteInfo.modify_lte(state_lte, state_data)
            if not switch_to_lte:
                # 成功关闭无线网络之后，需要取消策略路由配置，恢复有线网络原来的状态
                ret = self._cancel_policy_routing(fd_server_ip, request_ip)
                if ret[0] != commMethods.CommonMethods.OK:
                    run_log.error("cancel policy routes failed before close LTE: %s", ret[1])
                    return ret

            self.get_all_info()
            return [commMethods.CommonMethods.OK, ]
