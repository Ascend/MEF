#!/usr/bin/python
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from bin.monitor_config import SystemSetting

from devm.device_mgr import DEVM
from devm.exception import DeviceManagerError
from lib.Linux.EdgeSystem import log_helper
from lib.Linux.EdgeSystem.log_helper import log_and_ret_formatted_err


class AiProcessorInfo:
    """
    功能描述：AI处理器资源信息
    接口：NA
    修改记录：2018-3-20 创建
    """

    def __init__(self):
        """
        功能描述：初始化函数
        参数：
        返回值：无
        异常描述：NA
        """
        self.Manufacturer = None
        self.Model = None
        self.NpuVersion = None
        self.Count = None

        self.Calc = None
        self.Ddr = None

        self.Health = None
        self.State = None

        self.AiCore = None
        self.AiCpu = None
        self.CtrlCpu = None
        self.DdrUsage = None
        self.DdrBandWidth = None

    @staticmethod
    def get_usage_rate(attribute_name):
        """AI处理器占用率"""
        try:
            return DEVM.get_device('davinci0').get_attribute('utilization_rate').get(attribute_name)
        except DeviceManagerError:
            return log_and_ret_formatted_err(log_helper.ERR_GET_DAVINCI)

    def get_all_info(self):
        # AI处理器信息
        self.get_chip_info()
        # AI处理器个数
        self.get_ai_count()
        # AI处理器内存
        self.get_memory_size()
        # AI处理器的状态
        # 健康状态
        self.get_minid_health_status()

        # Ai处理器版本
        self.NpuVersion = SystemSetting().npu_version

        # 处理器资源占用率
        # 20200609:902版本的dsmi接口获取npu使用率直接获取即可，不需要除2
        ai_core = self.get_usage_rate('ai_core')
        if isinstance(ai_core, int):
            self.AiCore = ai_core
        # AI CPU占用率
        ai_cpu = self.get_usage_rate('ai_cpu')
        if isinstance(ai_cpu, int):
            self.AiCpu = ai_cpu
        # Ctrl CPU占用率
        ctrl_cpu = self.get_usage_rate('ctrl_cpu')
        if isinstance(ctrl_cpu, int):
            self.CtrlCpu = ctrl_cpu
        # DDR内存占用率
        ddr_usage = self.get_usage_rate('ddr')
        if isinstance(ddr_usage, int):
            self.DdrUsage = ddr_usage
        # 内存带宽占用率
        ddr_usage = self.get_usage_rate('ctrl_ddr_bandwidth')
        if isinstance(ddr_usage, int):
            self.DdrBandWidth = ddr_usage

    def get_memory_size(self):
        try:
            self.Ddr = round(DEVM.get_device('davinci0').get_attribute('memory') / 1000, 2)
        except DeviceManagerError:
            log_and_ret_formatted_err(log_helper.ERR_GET_CHIP_MEM_SIZE)

    def get_vendor(self):
        try:
            self.Manufacturer = DEVM.get_device('davinci0').get_attribute('chip_info').get('vendor')
        except DeviceManagerError:
            log_and_ret_formatted_err(log_helper.ERR_GET_CHIP_INFO)

    def get_model(self):
        try:
            self.Model = DEVM.get_device('davinci0').get_attribute('chip_info').get('model')
        except DeviceManagerError:
            log_and_ret_formatted_err(log_helper.ERR_GET_CHIP_INFO)

    def get_ability_desc(self):
        self.Calc = SystemSetting().calculate_ability

    def get_chip_info(self):
        self.get_model()
        self.get_vendor()
        self.get_ability_desc()

    def get_ai_count(self):
        """AI处理器数量"""
        try:
            self.Count = DEVM.count_devices_by_module('npu')
        except DeviceManagerError:
            log_and_ret_formatted_err(log_helper.ERR_GET_CHIP_COUNT)

    def get_minid_health_status(self):
        """AI处理器健康状态"""
        try:
            res = DEVM.get_device('davinci0').get_attribute('health')
        except DeviceManagerError:
            self.State = None
            self.Health = None
            log_and_ret_formatted_err(log_helper.ERR_GET_HEALTH)
            return
        self.State = True
        if res == 0:
            self.Health = "OK"
        elif res == 1:
            self.Health = "General warning"
        elif res == 2:
            self.Health = "Important warning"
        elif res == 3:
            self.Health = "Urgent warning"
        # 假如mindi数量为0，则健康/使能状态不可知
        elif res == -2:
            self.State = None
            self.Health = None
        else:
            self.Health = "Unknown mistake"
        return
