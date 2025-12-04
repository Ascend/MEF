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
import json
import os
import re
import socket
import threading
import time
from typing import List, Dict, NoReturn, Optional, Any

from bin.monitor_config import SystemSetting
from common.checkers.param_checker import ProductNameChecker
from common.constants.base_constants import CommonConstants
from common.constants.upgrade_constants import OMUpgradeConstants
from common.file_utils import FileCheck
from common.init_cmd import cmd_constants
from common.log.logger import run_log
from common.utils.app_common_method import AppCommonMethod
from common.utils.common_check import CommonCheck
from common.utils.exec_cmd import ExecCmd
from common.utils.system_utils import SystemUtils
from common.utils.version_xml_file_manager import VersionXmlManager
from devm import config
from devm.device_mgr import DEVM
from devm.exception import DeviceManagerError
from common.utils.result_base import Result
from lib.Linux.EdgeSystem import log_helper
from lib.Linux.EdgeSystem.log_helper import log_and_ret_formatted_err
from lib.Linux.EdgeSystem.log_helper import log_formatted_err
from lib.Linux.EdgeSystem.log_helper import operation_log_info
from lib.Linux.systems.disk.contants import STATUS_INACTIVE
from lib.Linux.systems.extend import Extend
from lib.Linux.systems.schemas import FirmwareInfo
from lib.Linux.systems.time_zone_mgr import set_time_zone_offset
from lib.Linux.upgrade.module_type import ModuleType
from lib.Linux.upgrade.upgrade_new import Upgrade
from common.common_methods import CommonMethods


def get_main_board_instance():
    try:
        dev_instance = DEVM.get_device("mainboard0")
    except DeviceManagerError as dev_err:
        run_log.error(dev_err)
        dev_instance = None

    return dev_instance


class SystemInfo:
    """
    功功能描述：查询系统资源信息
    接口：NA
    修改记录：2018-3-18 创建
    """
    SYSTEM_LOCK = threading.Lock()
    SYSTEM_RESOURCES_LIST = {"HostName", "AssetTag", "DateTime", "DateTimeLocalOffset", "_User", "_Xip"}
    UPTIME_PATH = "/proc/uptime"

    CpuHeatingState = {
        0: "Heating",
        1: "Stop",
    }

    DiskHeatingState = {
        0: "Heating",
        1: "Refrigeration",
        2: "Stop",
    }

    UsbHubHeatingState = {
        0: "Heating",
        1: "Stop"
    }

    # 内存信息文件
    _local__MenInfoFile = "/proc/meminfo"
    # CPU状态文件
    _local__CPUStatFile = "/proc/stat"
    # VERSION 文件
    _local__SysVerionFile = "/usr/local/lib/release_file"

    # 当前系统信息初始化一次，接口获取时，只需更新待生效相关信息
    firmware_info: List[FirmwareInfo] = [FirmwareInfo(name=module) for module in ModuleType.values()]
    main_board = get_main_board_instance()

    def __init__(self):
        """
        功能描述：初始化函数
        参数：
        返回值：无
        异常描述：NA
        """
        # 初始化父项的属性
        super().__init__()
        self.HostName = None
        self.UUID = None
        # 上报给FD用于显示型号，如果存在前端配置，则从配置中获取，保持web与FD显示一致；
        self.Model: Optional[str] = None
        # 上报给FD用于匹配version.xml中的SupportModel字段
        self.SupportModel: Optional[str] = None
        self.SerialNumber = None
        self.AssetTag = None
        self.PCBVersion = None
        self.Temperature = None
        self.Power = None
        self.Voltage = None
        self.CpuHeating = None
        self.DiskHeating = None
        self.UsbHubHeating = None
        self.AiTemperature = None
        self.OSVersion = ""
        # 固件(MCU、NPU、Ascend-firmware)信息
        self.Firmware: List[Dict[str, Any]] = []
        self.KernelVersion = None
        self.Uptime = None
        self.Datetime = None
        self.DatetimeLocalOffset = None
        self.CpuUsage = None
        self.MemoryUsage = None
        self.Health = None
        self.HI3559_frequency = None
        self.DateTimeLocalOffset = None
        self.InactiveConfiguration = None
        self.Manufacturer = None
        self.CPUModel = None
        self.Count = None
        self.TotalSystemMemoryGiB = None
        # 保存从内存信息文件获取到的内容
        self._local_mem_info = ""
        # 临时列表保存 "cat /cpu/stat" 信息
        self._local_cpu_state_info_list = []
        # 上报给FD的详细版本信息字段
        self.ProcessorArchitecture: Optional[str] = None

    @staticmethod
    def check_request_data(request):
        # 初步校验请求参数
        if request is None or request == "":
            log_formatted_err(log_helper.ERR_SET_LABEL_FAILED)
            return False
        return True

    @staticmethod
    def elabel_check(elabel_info):
        # 匹配所有英文字母+数字+_-.
        pattern = r"^([\x20-\x7E]){1,255}"
        if not re.fullmatch(pattern, elabel_info):
            run_log.error("elabel info check failed")
            return False
        return True

    @staticmethod
    def get_model_from_web_file() -> str:
        """从白牌配置中获取设备型号"""
        if not FileCheck.check_path_is_exist_and_valid(CommonConstants.WEB_CONF):
            run_log.error("invalid web config file")
            raise OSError("invalid web config file")

        if os.path.getsize(CommonConstants.WEB_CONF) > CommonConstants.OM_READ_FILE_MAX_SIZE_BYTES:
            run_log.error("The web config file size exceeds the max limit.")
            raise OSError("The web config file size exceeds the max limit.")

        with open(CommonConstants.WEB_CONF) as cfg_file:
            model = json.load(cfg_file).get("model") or ""

        if not ProductNameChecker().check({"model": model}):
            run_log.error("invalid product name")
            raise ValueError("invalid product name")

        return model

    @staticmethod
    def format_time(mcu_time):
        date_time = str(mcu_time)
        # 校验传来的时间格式
        try:
            if ":" in date_time:
                time.strptime(date_time, "%Y-%m-%d %H:%M:%S")
            else:
                time.strptime(date_time, "%Y-%m-%d")
        except Exception:
            return [CommonMethods.ERROR, "format time failed"]

        return [CommonMethods.OK, date_time]

    @staticmethod
    def set_os_time(date_time):
        cmd = ["/usr/bin/date", "-s", date_time]
        ret = ExecCmd.exec_cmd_get_output(cmd, wait=10)
        if ret[0] != 0:
            run_log.error("set system time failed. %s", ret[1])
            return [CommonMethods.INTERNAL_ERROR, "Set system time failed."]

        return [CommonMethods.OK, "set system time success."]

    @staticmethod
    def save_time_zone_config(time_zone: str):
        if not time_zone:
            run_log.info("no need to save time zone")
            return

        try:
            set_time_zone_offset(time_zone)
        except Exception as err:
            run_log.error("save time zone config failed : %s", err)
            return

        run_log.info("save time zone config successfully")

    @staticmethod
    def get_capacity(value, dest_unit='GB'):
        """
        功能描述：根据指定的目标单位返回换算结果，默认返回以 GB 为单位的数值，单位兼容大小写
        参数：
            value  需要换算的数据: 1、整型，单位默认为 'KB'；2、字符串类型，不带单位则默认为 KB
            dest_unit  默认为 'GB'
            value, unit 参数中可识别的单位字符串有: K/M/G/T/P/KB/MB/GB/TB/PB（包括小写形式）
        返回值：数值，精度为保留两位小数(如果目标单位为'KB'，则返回整型)
        异常描述：NA
        修改记录：2016-12-13 创建
                2017-02-08 修改： 添加支持 Byte 数据处理
        """

        _size_map = {
            "B": 1024 ** -1,
            "K": 1, "KB": 1,
            "M": 1024, "MB": 1024,
            "G": 1024 ** 2, "GB": 1024 ** 2,
            "T": 1024 ** 3, "TB": 1024 ** 3,
            "P": 1024 ** 4, "PB": 1024 ** 4
        }
        if isinstance(value, int):
            src_unit = 'KB'
        elif isinstance(value, str):
            value = value.split()

            if len(value) == 1:

                if value[0].isdigit():
                    src_unit = 'KB'
                    value = int(value[0])

                elif value[0][:-1].isdigit():
                    src_unit = value[0][-1:].upper()
                    value = int(value[0][:-1])

                elif value[0][:-2].isdigit():
                    src_unit = value[0][-2:].upper()
                    value = int(value[0][:-2])
                else:
                    return -1

            elif len(value) == 2:
                if value[0].isdigit():
                    src_unit = value[1].upper()
                    value = int(value[0])
                else:
                    return -1
            else:
                return -1
        else:
            return -1

        # 兼容前后带有空格的情况
        dest_unit = dest_unit.strip().upper()
        if src_unit in _size_map and dest_unit in _size_map:
            if dest_unit == "KB":
                return value * (_size_map.get(src_unit)) / (_size_map.get(dest_unit))
            return round(float(value) * (_size_map.get(src_unit)) / (_size_map.get(dest_unit)), 2)
        else:
            return -1

    @staticmethod
    def euler_set_host_name(host_name):
        ret = ExecCmd.exec_cmd([cmd_constants.OS_CMD_HOSTNAME, "-b", host_name], wait=30)
        if ret != 0:
            return Result(False, err_msg="hostname {}, return {}".format(host_name, ret))

        if not FileCheck.check_is_link("/etc/hostname"):
            return Result(False, err_msg="check hostname file failed.")

        try:
            with os.fdopen(os.open("/etc/hostname", os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o640), "w") as file:
                file.write(host_name)
        except Exception as err:
            raise Exception("write hostname file failed:{}.".format(err)) from err

        return Result(True)

    @staticmethod
    def openos_set_host_name(host_name):
        ret = ExecCmd.exec_cmd([cmd_constants.OS_CMD_HOSTNAMECTL, "hostname", host_name], wait=30)
        if ret != 0:
            return Result(False, err_msg="hostname {}, return {}".format(host_name, ret))
        return Result(True)

    @staticmethod
    def set_date_time_local_offset(date_time_local_offset):
        # 修改时区
        cmd = [cmd_constants.OS_CMD_TIMEDATECTL, "set-timezone", date_time_local_offset]
        ret = ExecCmd.exec_cmd(cmd, wait=10)
        if ret != 0:
            return log_and_ret_formatted_err(log_helper.ERR_SET_TIME_TIMEZONES)

        # 根据环境变量 重新初始化时间相关设置。
        time.tzset()
        time.sleep(3)
        run_log.info("Set timezones success.")
        return [CommonMethods.OK, "ok"]

    @staticmethod
    def get_file_content(filepath):
        if not FileCheck.check_path_is_exist_and_valid(filepath):
            return ""
        try:
            with open(filepath, "r") as file:
                return file.read().strip('\n')
        except Exception as err:
            run_log.error("read %s failed", err)
            return ""

    @staticmethod
    def get_cpu_stat_info(info_list, logical_cpu_id_list=None):
        """
        功能描述：取样 CPU 状态信息
        参数：
        info_list "/proc/stat" 信息列表
        logical_cpu_id_list  逻辑CPU ID列表，"/proc/cpuinfo" 中的 "processor"；
        使用默认值或者空列表时，获取并返回总CPU状态信息
        返回值：正确返回表示 CPU 状态信息的列表[已使用值,总值]；
        否则，返回 空列表
        异常描述：NA
        修改记录：2016-12-15 创建
        """

        # 初始化返回值
        ret_list = []
        # 初始化
        cpu_user = 0
        cpu_nice = 0
        cpu_system = 0
        cpu_idle = 0
        cpu_iowait = 0
        cpu_irq = 0
        cpu_softirq = 0
        cpu_steal = 0
        # logical_cpu_id_list 为空列表或者 None 时，返回总CPU占用率百分比
        if not logical_cpu_id_list:
            # 判断信息是否正确，错误返回
            # 正确的值类似："cpu  484964 180 776813 2317731354 95771 0 567 0 0 0"
            # 'cpu' 与 第二个字符串之间有两个空格
            key = "cpu  "
            if not info_list[0].startswith(key):
                return ret_list
            tmp_list = info_list[0].split(' ')
            if len(tmp_list) < 10:
                return ret_list
            try:
                cpu_user = cpu_user + int(tmp_list[2])
                cpu_nice = cpu_nice + int(tmp_list[3])
                cpu_system = cpu_system + int(tmp_list[4])
                cpu_idle = cpu_idle + int(tmp_list[5])
                cpu_iowait = int(tmp_list[6])
                cpu_irq = int(tmp_list[7])
                cpu_softirq = int(tmp_list[8])
                cpu_steal = int(tmp_list[9])
            except Exception:
                return ret_list
        else:
            # 去掉 总CPU 信息
            info_list = info_list[1:]
            for number in logical_cpu_id_list:
                if number.isdigit():
                    index = int(number)
                else:
                    return ret_list
                # 判断信息是否正确，错误返回
                # 正确的值类似： "cpu0

                # 'cpu0' 与 第二个字符串之间只有一个空格
                key = "cpu%s " % (logical_cpu_id_list[number])
                if not info_list[index].startswith(key):
                    return ret_list
                tmp_list = info_list[index].split(' ')
                if len(tmp_list) < 9:
                    return ret_list
                try:
                    cpu_user = cpu_user + int(tmp_list[1])
                    cpu_nice = cpu_nice + int(tmp_list[2])
                    cpu_system = cpu_system + int(tmp_list[3])
                    cpu_idle = cpu_idle + int(tmp_list[4])
                    cpu_iowait = cpu_iowait + int(tmp_list[5])
                    cpu_irq = cpu_irq + int(tmp_list[6])
                    cpu_softirq = cpu_softirq + int(tmp_list[7])
                    cpu_steal = cpu_steal + int(tmp_list[8])
                except Exception:
                    return ret_list

        # cpu使用总量为 user + system + nice + idle + iowait + irq + softirq + steal
        cpu_total = cpu_user + cpu_system + cpu_nice + cpu_idle + cpu_iowait + cpu_irq + cpu_softirq + cpu_steal
        ret_list.append(cpu_idle)
        ret_list.append(cpu_total)
        return ret_list

    # 获取kernel版本
    @staticmethod
    def get_kernel_version():
        version = ""
        version_file = "/proc/version"
        res = FileCheck.check_path_is_exist_and_valid(version_file)
        if not res:
            run_log.error("%s path invalid : %s", version_file, res.error)
            return version

        # 校验文件大小
        filesize = os.path.getsize(version_file)
        if filesize > CommonConstants.OM_READ_FILE_MAX_SIZE_BYTES:
            run_log.error("File: %s is too large.", version_file)
            return version

        try:
            with open(version_file, "r") as file:
                version = file.read().split(" ")[2]
        except Exception as err:
            run_log.error("get kernel version failed. %s, %s", version_file, err)

        return version

    # 设置产品资产标签
    @staticmethod
    def set_sys_electronic_tags(request_data, label):
        try:
            DEVM.get_device("elabel0").set_attribute("asstag", label)
        except DeviceManagerError:
            return log_and_ret_formatted_err(log_helper.ERR_SET_LABEL_FAILED)
        return [CommonMethods.OK, "Set system electronic tags successful."]

    def get_all_info(self, info=None):
        if info == "ProcessorsSummary":
            self.get_h3559_info()
        elif info == "MemorySummary":
            self.get_dynamic_info()
        else:
            self.get_all_system_info()

    def patch_request(self, request_data):
        """
        功能描述：修改指定系统资源属性
        参数：request_data 客户端传来的修改信息
        返回值：无
        异常描述：NA
        """
        if SystemInfo.SYSTEM_LOCK.locked():
            run_log.warning("System modify is busy")
            return [CommonMethods.ERROR, "System modify is busy"]
        with SystemInfo.SYSTEM_LOCK:
            asset_tag = request_data.get("AssetTag")
            date_time = request_data.get("DateTime")
            date_time_local_offset = request_data.get("DateTimeLocalOffset")
            operator_check = CommonCheck.check_operator(request_data.get("_User"), request_data.get("_Xip"))
            if not operator_check:
                run_log.error("The operator is illegal, %s", operator_check.error)
                return [CommonMethods.ERROR, "The operator is illegal."]

            # 异常参数限制
            input_error_info = "Parameter is wrong"
            if not set(self.SYSTEM_RESOURCES_LIST).issuperset(request_data):
                return [CommonMethods.ERROR, input_error_info]

            # 设置用户定义电子标签
            ret = self.set_asset_tag(request_data, asset_tag)
            if ret[0] != CommonMethods.OK:
                return ret

            # 修改时区
            if date_time_local_offset is not None:
                ret = SystemInfo.check_request_data(date_time_local_offset)
                if ret is False or re.fullmatch(r"^[\+\-_a-z:A-Z0-9\/]{0,100}$", date_time_local_offset) is None or \
                        ".." in date_time_local_offset:
                    return log_and_ret_formatted_err(log_helper.ERR_SET_TIME_TIMEZONES,
                                                     "check DateTimeLocalOffset failed.")

                ret = self.set_date_time_local_offset(date_time_local_offset)
                if ret[0] != CommonMethods.OK:
                    operation_log_info(request_data, "Set time zone failed.")
                    return ret
                operation_log_info(request_data, "Set time zone successfully.")
                self.save_time_zone_config(date_time_local_offset)

            # 设置时间
            ret = self.set_time(request_data, date_time)
            if ret[0] != CommonMethods.OK:
                return ret

            # 修改主机名
            ret = self.set_host_name(request_data)
            if ret[0] != CommonMethods.OK:
                return ret

            # 更新系统信息
            self.get_all_system_info()
            return [CommonMethods.OK, ]

    def set_host_name(self, request_data):
        """
        功能描述：修改主机名
        参数：request_data 客户端传来的修改信息
        返回值：无
        异常描述：NA
        """
        host_name = request_data.get("HostName")
        if host_name is not None:
            err_code = log_helper.ERR_SET_HOSTNAME
            if not AppCommonMethod.hostname_check(host_name):
                run_log.error("The hostname is invalid.")
                operation_log_info(request_data, "Set hostname failed.")
                return log_and_ret_formatted_err(err_code, "hostname failed to pass name check.")

            try:
                if cmd_constants.OS_NAME == "EulerOS":
                    ret = self.euler_set_host_name(host_name)
                else:
                    ret = self.openos_set_host_name(host_name)
                if not ret:
                    operation_log_info(request_data, "Set hostname failed.")
                    return log_and_ret_formatted_err(err_code, ret.error)

            except Exception as err:
                operation_log_info(request_data, "Set hostname failed.")
                return log_and_ret_formatted_err(err_code, err)

            run_log.info("The hostname has been changed to [%s].", host_name)
            operation_log_info(request_data, "Set hostname successfully.")

        return [CommonMethods.OK, ]

    def get_time_zone(self):
        # 获取时区
        shell_str = "%s | grep 'Time zone'" % cmd_constants.OS_CMD_TIMEDATECTL
        ret = ExecCmd.exec_cmd_use_pipe_symbol(shell_str, wait=30)
        self.DateTimeLocalOffset = ret[1].split(":")[1].strip() if ret[0] == 0 else None

    def get_all_system_info(self):
        self.get_host_name()
        self.get_sys_uuid()
        # 产品序列号
        self.get_sys_product_num()
        # 用户定义电子标签
        self.get_sys_electronic_tags()

        if self.PCBVersion is None:
            self.get_pcb_id()

        self.get_model()

        # 获取各版本号
        self.get_os_version_info()
        self.get_kernel_number()

        # 更新固件信息
        self.get_firmware_version()
        self.get_processor_architecture()

        # 获取运行时间
        self.get_sys_uptime()
        # 获取时间
        self.get_sys_time()
        # 时区
        self.get_time_zone()
        # 获取CPU、内存占用率
        self.get_dynamic_info()
        self.get_minid_temperature()

        # 固件升级过程mcu相关信息获取跳过
        if Upgrade.is_running() is True:
            run_log.info('Firmware upgrade running,  passing mcu related check.')
        else:
            self.get_power()
            # mainboard0不存在则直接返回
            if not self.main_board:
                return

            self.get_sys_health_status()
            self.get_mcu_temperature()
            self.get_mcu_voltage()
            self.get_mcu_hot_cold_status()

    def get_firmware_version(self):
        self.Firmware = [f.get_module_msg() for f in self.firmware_info]

    # 获取pcbid
    def get_pcb_id(self):
        if not self.main_board:
            return

        try:
            self.PCBVersion = self.main_board.get_attribute("pcb_id")
        except DeviceManagerError:
            log_and_ret_formatted_err(log_helper.ERR_GET_BOARD_INFO)

    # 获取系统连续运行时间
    def get_sys_uptime(self):
        ret = FileCheck.check_path_is_exist_and_valid(SystemInfo.UPTIME_PATH)
        if not ret:
            run_log.error("%s path invalid : %s", SystemInfo.UPTIME_PATH, ret.error)
            return

        try:
            with open(SystemInfo.UPTIME_PATH, "r") as file:
                up_time = float(file.readline().split(" ")[0])
        except Exception as err:
            run_log.error("get sys run time failed! reason: %s", err)
            return

        # 将时间转换为日 时 分 秒
        days = int(up_time // 86400)
        hours = int(up_time // 3600 % 24)
        if hours < 10:
            hours = f"0{hours}"
        minutes = int(up_time % 3600 // 60)
        if minutes < 10:
            minutes = f"0{minutes}"
        seconds = int(up_time % 60)
        if seconds < 10:
            seconds = f"0{seconds}"
        self.Uptime = f"{hours}:{minutes}:{seconds} {days} days"

    # 获取时间
    def get_sys_time(self):
        self.Datetime = time.strftime("%a %b %d %H:%M:%S %Y", time.localtime())

    def get_model(self):
        """获取设备型号"""
        if self.SupportModel is not None and self.Model is not None:
            return
        self.SupportModel = SystemSetting().board_type

        # 不存在白牌标记则以npu-smi获取结果为准
        if not os.path.exists(OMUpgradeConstants.WHITEBOX_BACKUP_DIR):
            run_log.info("The system has not been white.")
            self.Model = self.SupportModel
            return

        try:
            self.Model = self.get_model_from_web_file()
        except Exception as err:
            run_log.error("Get model from config err, catch %s", err.__class__.__name__)
            self.Model = self.SupportModel

    def get_dynamic_info(self):
        """
        功能描述：获取系统动态属性信息
        参数：无
        返回值：无
        异常描述：NA
        修改记录：2016-12-06 创建
        """
        # 获取总CPU占用率百分占比
        cpu_usage_temp = self.get_cpu_usage(logical_cpu_id_list=None)
        if cpu_usage_temp != -1:
            self.CpuUsage = cpu_usage_temp
        self.get_memory_usage()

    def get_memory_usage(self):
        # 获取系统内存
        file_content = self.get_file_content(self._local__MenInfoFile)
        if not file_content:
            return

        # 获取系统内存信息
        mem_total = self.get_capacity(
            CommonMethods.get_value_by_key(file_content, "MemTotal", ":", last_match=True), "KB")
        mem_free = self.get_capacity(
            CommonMethods.get_value_by_key(file_content, "MemFree", ":", last_match=True), "KB")
        page_cache = self.get_capacity(
            CommonMethods.get_value_by_key(file_content, "Cached", ":", last_match=True), "KB")
        slab_reclaimable = self.get_capacity(
            CommonMethods.get_value_by_key(file_content, "SReclaimable", ":", last_match=True), "KB")
        main_buffers = self.get_capacity(
            CommonMethods.get_value_by_key(file_content, "Buffers", ":", last_match=True), "KB")

        # 如果有一个值为 None，不计算已使用情况
        if (mem_total, mem_free, page_cache, slab_reclaimable, main_buffers).count(-1):
            return
        mem_total = int(mem_total)
        mem_total_gb = self.get_capacity(mem_total, "GB")
        if mem_total_gb != -1:
            self.TotalSystemMemoryGiB = mem_total_gb
        cached = int(page_cache) + int(slab_reclaimable)
        main_buffers = int(main_buffers)

        # 根据linux free的源码，正确计算已使用memory的方法为：MemTotal - MemFree - (Cached + SReclaimable) - Buffers
        mem_used = mem_total - mem_free - cached - main_buffers
        if mem_total != 0 and 0 <= mem_used <= mem_total:
            mem_used = int(mem_used)
            self.MemoryUsage = round((mem_used / float(mem_total)) * 100, 2)

    def get_cpu_usage(self, logical_cpu_id_list=None):
        """
        功能描述：获取 CPU 占用率百分占比
        参数：logical_cpu_id_list  逻辑CPU ID列表，"/proc/cpuinfo" 中的 "processor"；
        使用默认值或者空列表时，获取并返回总CPU占用率百分比
        返回值：正确返回表示 CPU 占用率百分占比的数值；否则，返回 None
        异常描述：NA
        修改记录：2016-12-06 创建
        2016-12-10 修改： 传入cpu_id 为列表
        """
        ret1 = self.get_file_content(self._local__CPUStatFile)
        # 间隔 1S，再取值
        time.sleep(1)
        ret2 = self.get_file_content(self._local__CPUStatFile)
        if not ret1 or not ret2:
            run_log.error("read %s fail.", self._local__CPUStatFile)
            return -1

        self._local_cpu_state_info_list[:2] = ret1.split("\n"), ret2.split("\n")
        ret1 = self.get_cpu_stat_info(
            self._local_cpu_state_info_list[0], logical_cpu_id_list)
        ret2 = self.get_cpu_stat_info(
            self._local_cpu_state_info_list[1], logical_cpu_id_list)
        if not ret1 or not ret2:
            run_log.error("result is exception : '%s'", self._local__CPUStatFile)
            return -1
        # cpu利用率 = (1 - (idle2 - idle1)  / (total2 - total1)) * 100
        cpu_idle = ret2[0] - ret1[0]
        cpu_total = ret2[1] - ret1[1]
        if cpu_total != 0 and cpu_total >= cpu_idle:
            return round((1 - cpu_idle / float(cpu_total)) * 100, 2)
        return -1

    def get_os_version_info(self):
        """
        功能描述：获取操作系统版本信息
        参数：无
        返回值：无
        异常描述: NA
        修改记录: 2018-04-12 创建
        """
        if self.OSVersion:
            return

        version_file = '/etc/os-release'
        line_count = 0
        try:
            with open(version_file, 'r') as file:
                for line in file:
                    if line.startswith("PRETTY_NAME"):
                        self.OSVersion = line.split("=")[-1].strip("\n").strip('"')
                    line_count += 1
                    if line_count > 128:
                        break
        except Exception as err:
            run_log.error('Failed to get system version: %s' % err)

    def get_processor_architecture(self) -> NoReturn:
        """从version.xml中获取ProcessorArchitecture"""
        if self.ProcessorArchitecture is not None:
            return

        try:
            self.ProcessorArchitecture = VersionXmlManager().processor_architecture
        except Exception as err:
            run_log.error("Failed to get MindXOM version, because of %s", err)

    # 获取系统健康状态
    def get_sys_health_status(self):
        try:
            local_minid_health_status = DEVM.get_device("system0").get_attribute("health")
        except DeviceManagerError:
            log_formatted_err(log_helper.ERR_GET_HEALTH)
            return

        count = 0
        local_extend_health_status = 0
        try:
            for fmt_info in Extend.valid_fmt_info_generator():
                if fmt_info.status == STATUS_INACTIVE:
                    local_extend_health_status = 1
                    break
                count += 1
                if count > config.MAX_DEVICE_NUMBER:
                    break
        except DeviceManagerError:
            run_log.error("Failed to get peripheral information")
            return

        # 假如mindi数量为0，则健康状态做为正常状态处理
        if -2 == local_minid_health_status:
            local_minid_health_status = 0
        local_3559_health_status = 0
        fin_health_status = local_minid_health_status + local_extend_health_status + local_3559_health_status
        if fin_health_status == 0:
            self.Health = "OK"
            return

        if fin_health_status == 1:
            self.Health = "General warning"
        elif fin_health_status == 2:
            self.Health = "Important warning"
        elif fin_health_status == 3:
            self.Health = "Urgent warning"
        else:
            self.Health = "Unknown mistake"
        run_log.info("%s, state list: %s/%s/%s" %
                     (self.Health, local_minid_health_status,
                      local_extend_health_status, local_3559_health_status))

    # 获取内核kernel版本
    def get_kernel_number(self):
        if self.KernelVersion is not None:
            return

        local__kernelversion_info = self.get_kernel_version()
        if not local__kernelversion_info:
            log_and_ret_formatted_err(log_helper.ERR_GET_KERNEL_VERSION)
        else:
            self.KernelVersion = local__kernelversion_info.strip('\n')

    # 获取产品资产标签
    def get_sys_electronic_tags(self):
        try:
            self.AssetTag = DEVM.get_device("elabel0").get_attribute("asstag")
        except DeviceManagerError:
            log_and_ret_formatted_err(log_helper.ERR_GET_ELECTRONIC_TAGS)

    # 获取hostname
    def get_host_name(self):
        hostname = socket.gethostname()
        if not hostname or hostname == -1:
            log_formatted_err(log_helper.ERR_GET_HOSTNAME)
        else:
            self.HostName = hostname

    # 获取uuid
    def get_sys_uuid(self):
        try:
            self.UUID = DEVM.get_device("elabel0").get_attribute("board_sn")
        except DeviceManagerError:
            log_and_ret_formatted_err(log_helper.ERR_GET_UUID_TAGS)

    # 获取单板温度
    def get_mcu_temperature(self):
        try:
            self.Temperature = self.main_board.get_attribute("temperature")
        except DeviceManagerError:
            log_and_ret_formatted_err(log_helper.ERR_GET_TEMPERATURE)

    # 获取电压
    def get_mcu_voltage(self):
        try:
            self.Voltage = round(self.main_board.get_attribute("board_voltage") / 100.0, 2)
        except DeviceManagerError:
            log_and_ret_formatted_err(log_helper.ERR_GET_VOLTAGE)

    # 获取加热制冷状态
    def get_mcu_hot_cold_status(self):
        try:
            self.CpuHeating = SystemInfo.CpuHeatingState.get(self.main_board.get_attribute("hot_cold_cpu_state"))
        except DeviceManagerError:
            log_and_ret_formatted_err(log_helper.ERR_GET_MCU_HEATING_COOL)

        try:
            self.DiskHeating = SystemInfo.DiskHeatingState.get(self.main_board.get_attribute("hot_cold_disk_state"))
        except DeviceManagerError:
            log_and_ret_formatted_err(log_helper.ERR_GET_MCU_HEATING_COOL)

        try:
            self.UsbHubHeating = SystemInfo.UsbHubHeatingState.get(
                self.main_board.get_attribute("hot_cold_usb_hub_state")
            )
        except DeviceManagerError:
            log_and_ret_formatted_err(log_helper.ERR_GET_MCU_HEATING_COOL)

    # 获取功率
    def get_power(self):
        try:
            if SystemSetting().board_type in CommonConstants.A500_MODELS:
                # 新的驱动接口，功率驱动已经乘以系数0.1了
                self.Power = round(self.main_board.get_attribute("board_power_info"), 2)
            else:
                self.Power = round(float(SystemUtils.get_power_by_npu_smi()), 2)
        except (DeviceManagerError, ValueError):
            log_and_ret_formatted_err(log_helper.ERR_GET_POWER)

    # 获取AI温度
    def get_minid_temperature(self):
        # try to get object id of npu
        try:
            # choose the first one to get temperature
            npu_name = DEVM.get_device_list_by_module("npu")[0]
        except (DeviceManagerError, IndexError):
            log_and_ret_formatted_err(log_helper.ERR_GET_NPU_LIST)
            return

        try:
            self.AiTemperature = DEVM.get_device(npu_name).get_attribute("temperature")
        except DeviceManagerError:
            log_and_ret_formatted_err(log_helper.ERR_GET_MINIDI_TEMPERATURE)

    # 获取CPU的信息
    def get_h3559_info(self):
        try:
            # choose the first cpu object id
            cpu_name = DEVM.get_device_list_by_module("cpu")[0]
        except (DeviceManagerError, IndexError):
            log_and_ret_formatted_err(log_helper.ERR_GET_CPU_ERROR)
            return

        try:
            chip_info = DEVM.get_device(cpu_name).get_attribute("chip_info")
        except DeviceManagerError:
            log_and_ret_formatted_err(log_helper.ERR_GET_CPU_ERROR)
            return

        # CPU厂商
        self.Manufacturer = chip_info.get("vendor")
        # CPU型号
        self.CPUModel = chip_info.get("model")
        # CPU核数
        self.Count = chip_info.get("core_num")
        # CPU频数
        self.HI3559_frequency = chip_info.get("frequency")
        return

    def get_sys_product_num(self):
        try:
            if SystemSetting().board_type in CommonConstants.A500_MODELS:
                self.SerialNumber = DEVM.get_device("elabel0").get_attribute("product_sn")
            else:
                self.SerialNumber = SystemUtils.get_sn_by_npu_smi()
        except DeviceManagerError:
            log_and_ret_formatted_err(log_helper.ERR_GET_PRODUCT_NUM)

    # 设置电子标签
    def set_asset_tag(self, request_data, asset_tag):
        if asset_tag is not None:
            # 校验标签
            ret = SystemInfo.check_request_data(asset_tag)
            ret_elabel_check = SystemInfo.elabel_check(asset_tag)
            if ret is False or ret_elabel_check is False:
                return log_and_ret_formatted_err(log_helper.ERR_SET_LABEL_FAILED)
            ret = self.set_sys_electronic_tags(request_data, asset_tag)
            if ret[0] != CommonMethods.OK:
                operation_log_info(request_data, "Set electronic label failed.")
                return ret
            operation_log_info(request_data, "Set electronic label successfully.")

        return [CommonMethods.OK, ""]

    # 设置时间
    def set_time(self, request_data, mcu_time):
        if mcu_time is None:
            return [CommonMethods.OK, ""]

        ret_date_time = SystemInfo.check_request_data(mcu_time)
        if ret_date_time is False:
            return log_and_ret_formatted_err(log_helper.ERR_SET_TIME_FAILED)

        ret = self.format_time(mcu_time)
        if ret[0] != CommonMethods.OK:
            return [CommonMethods.INTERNAL_ERROR, "format time failed."]

        date_time = ret[1]
        ret = self.set_os_time(date_time)
        if ret[0] != CommonMethods.OK:
            operation_log_info(request_data, "Set OS time failed.")
            return [CommonMethods.INTERNAL_ERROR, "Set OS time failed."]
        operation_log_info(request_data, "Set OS time successfully.")

        return [CommonMethods.OK, ""]
