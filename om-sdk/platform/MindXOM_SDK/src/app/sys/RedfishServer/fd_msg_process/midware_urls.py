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
import base64
import copy
import datetime
import glob
import json
import os
import shutil
import socket
import threading
import time
from typing import Dict
from typing import List

import urllib3

from common.checkers.fd_param_checker import ComputerSystemResetChecker
from common.checkers.fd_param_checker import FdCertDeleteChecker
from common.checkers.fd_param_checker import FdCertImportChecker
from common.checkers.fd_param_checker import FirmwareEffectiveChecker
from common.checkers.fd_param_checker import HostnameChecker
from common.checkers.fd_param_checker import ImportCrlChecker
from common.checkers.fd_param_checker import InfoCollectChecker
from common.checkers.fd_param_checker import ReAlarmChecker
from common.checkers.fd_param_checker import SysAssetTagChecker
from common.checkers.fd_param_checker import SysConfigChecker
from common.checkers.fd_param_checker import SysConfigEffectChecker
from common.checkers.fd_param_checker import UserInfoChecker
from common.checkers.param_checker import ExtendedDeviceChecker
from common.constants.base_constants import CommonConstants
from common.constants.base_constants import FlagPathConstant
from common.constants.base_constants import LoggerConstants
from common.constants.base_constants import UserManagerConstants
from common.constants.config_constants import ConfigPathConstants
from common.constants.product_constants import LOG_MODULES_MAP
from common.constants.product_constants import SERVICE_ROOT
from common.file_utils import FileCheck, FileCreate, FileUtils
from common.init_cmd import cmd_constants
from common.kmc_lib.kmc import Kmc
from common.log.logger import operate_log
from common.log.logger import run_log
from common.utils.app_common_method import AppCommonMethod
from common.utils.url_downloader import UrlConnect
from common.utils.version_xml_file_manager import VersionXmlManager
from fd_msg_process.capability import CAPABILITY
from fd_msg_process.common_redfish import CommonRedfish
from fd_msg_process.config import SysInfoTaskStatus
from fd_msg_process.config import Topic
from fd_msg_process.fd_common_methods import FdCommonMethod
from fd_msg_process.fd_common_methods import MidwareErrCode
from fd_msg_process.fd_common_methods import publish_ws_msg
from fd_msg_process.midware_proc import MidwareProc
from fd_msg_process.midware_route import MidwareRoute
from fd_msg_process.schemas import Firmware
from ibma_redfish_globals import RedfishGlobals
from lib_restful_adapter import LibRESTfulAdapter
from net_manager.common_methods import docker_root_is_mounted
from net_manager.constants import NetManagerConstants
from net_manager.exception import NetManagerException
from net_manager.manager.fd_cert_manager import FdCertManager
from net_manager.manager.fd_cfg_manager import FdCfgManager
from net_manager.manager.fd_cfg_manager import FdMsgData
from net_manager.manager.import_manager import ImportCert
from net_manager.manager.import_manager import ImportCrl
from net_manager.manager.net_cfg_manager import NetCfgManager
from net_manager.models import NetManager
from net_manager.schemas import PayloadPublish
from system_service.systems_serializer import AccessControlSerializer
from system_service.systems_serializer import AlarmInfoResourceSerializer
from system_service.systems_serializer import LteConfigInfoResourceSerializer
from system_service.systems_serializer import LteStatusInfoResourceSerializer
from user_manager.user_manager import SessionManager
from user_manager.user_manager import UserManager
from wsclient.fd_connect_check import FdConnectCheck

midWare = MidwareRoute()

REDFISH_PATH = "/home/data/config/redfish"
REDFISH_KSF = os.path.join(REDFISH_PATH, "redfish_encrypt.keystore")
REDFISH_BAK_KSF = os.path.join(REDFISH_PATH, "redfish_encrypt_backup.keystore")
REDFISH_ALG_CFG = os.path.join(REDFISH_PATH, "om_alg.json")
kmc = Kmc(REDFISH_KSF, REDFISH_BAK_KSF, REDFISH_ALG_CFG)


class MidwareUris:
    HOST_NAME_LOCK = threading.Lock()
    sys_lock = threading.Lock()
    alarm_lock = threading.Lock()

    profile_lock = threading.Lock()
    net_manager_lock = threading.Lock()
    resp_json_sys_info = json.loads(SERVICE_ROOT.systems.sysinfo_resource.get_resource())
    resp_json_sys_status = json.loads(SERVICE_ROOT.systems.sysstatus_resource.get_resource())
    import_cert_lock = threading.Lock()
    import_crl_lock = threading.Lock()
    cert_delete_lock = threading.Lock()
    cert_query_lock = threading.Lock()
    alarm_lists = {}
    alarm_info_dict = {}
    alarm_sn = 1
    max_cycle_num = 512
    simple_storages_info = []
    simple_storages_status = []

    @staticmethod
    def set_operation_log(input_info):
        if input_info:
            fd_ip = FdCfgManager.get_cur_fd_ip()
            if not fd_ip:
                run_log.error("get fd ip failed.")
                fd_ip = "AtlasEdge"
            operation_subject = "FD@{}".format(fd_ip)
            operate_log.info("[%s] %s", operation_subject, input_info)

    @staticmethod
    def check_external_param(checker_class, param_data, payload_publish, err_info):
        check_ret = checker_class().check(param_data)
        if not check_ret.success:
            run_log.error(err_info % check_ret.reason)
            if "restartable" in payload_publish.keys():
                payload_publish["restartable"] = "false"
            if "result" in payload_publish.keys():
                payload_publish["result"] = "failed"
            payload_publish["reason"] = MidwareErrCode.midware_generate_err_msg(
                MidwareErrCode.midware_input_parameter_invalid, check_ret.reason)
            return [-1, payload_publish]
        else:
            return [0, ""]

    @staticmethod
    def check_json_request(ret_dict):
        err_code = CommonRedfish.inputParamErrCode
        ret = {'status': err_code, 'message': "MalformedJSON"}
        try:
            # 通过转换判断字符串是否为 json
            request_data_dict = json.loads(ret_dict)
        except Exception as err:
            run_log.error("ERR.0%s,Request data is not json.", err)
            return [1, ret]
        if not isinstance(request_data_dict, dict) or request_data_dict == {}:
            return [1, ret]
        return [0, request_data_dict]

    @staticmethod
    def check_capacity_bytes_valid(partition_cfg: List[Dict]) -> bool:
        """
        FD下发分区配置时，限制分区大小只能为0（表示删除分区），或者为512MB的整数倍，与Redfish接口保持校验规则保持一致
        :param partition_cfg: 分区配置列表信息
        :return: True 符合要求；False 不符合要求
        """
        if not partition_cfg:
            return True

        permit_unit = 512 * 1024 * 1024
        for item in partition_cfg:
            capacity_bytes = item.get("capacity_bytes")
            if capacity_bytes == 0:
                continue

            if capacity_bytes % permit_unit:
                return False

        return True

    @staticmethod
    def get_log_collect_publish_template(module_name):
        if module_name not in ("all", ):
            run_log.error("module_name invalid")

        return {"type": "all", "module": module_name, "percentage": "0%",
                "result": "processing", "reason": ""}

    @staticmethod
    def https_upload(payload, local_path):
        if os.path.getsize(local_path) > LoggerConstants.COLLECT_LOG_UPPER_LIMIT:
            run_log.info("The log size is up to limit")
            return [-1, "The log size is up to limit."]

        https_server = payload.get("https_server")
        url = https_server.get("url").split(" ")[1]
        user_name = https_server.get("user_name")
        password = https_server.get("password")

        cert_ret = UrlConnect.get_context()
        if not cert_ret:
            run_log.error(f"{cert_ret.error}")
            return [-1, "get client ssl context error."]

        ret = FileCheck.check_path_is_exist_and_valid(local_path)
        if not ret:
            return [-1, f"{local_path} path invalid : {ret.error}"]

        with open(local_path, "rb") as file:
            file_data = file.read()

        files = {"tiFile": (local_path.split("/")[-1], file_data)}
        str_authentication = AppCommonMethod.make_authentication_string(user_name, password)
        request_header = {"Authorization": str_authentication}

        http = urllib3.PoolManager(ssl_context=cert_ret.data, assert_hostname=False, timeout=20)
        resp = http.request("POST", url, headers=request_header, fields=files, preload_content=False)
        # 连接成功对返回值进行处理
        resp_status = resp.status
        resp.release_conn()
        if resp_status != 200:
            err_msg = "upload log to fd failed"
            run_log.error("Upload failed status: {} msg: {}".format(resp_status, err_msg))
            return [-1, "upload log to fd failed"]
        run_log.info("UrlUploader upload file finished.")
        return [0, "upload log to fd success"]

    @staticmethod
    def get_alarm_health_status():
        critical_flag = False
        major_flag = False
        minor_flag = False
        for alarm_id_exist in MidwareUris.alarm_lists:
            for alarm in MidwareUris.alarm_lists[alarm_id_exist]:
                if alarm["perceivedSeverity"] == "CRITICAL":
                    critical_flag = True
                elif alarm["perceivedSeverity"] == "MAJOR":
                    major_flag = True
                elif alarm["perceivedSeverity"] == "MINOR":
                    minor_flag = True

        if critical_flag:
            return "Critical"
        elif major_flag or minor_flag:
            return "Warning"
        else:
            return "OK"

    @staticmethod
    def get_product_capability():
        cap = CAPABILITY.product_capability
        # 2.0.4版本与设备OM解耦，能力项内容各自拆分上报，且上报FD的名字改变
        not_support_caps = [
            "ief_enable",
            "pod_config",
            "resource_files_config",
            "pod_restart",
            "container_app_task_stop",
            "npu_sharing"
        ]
        ret_dict = LibRESTfulAdapter.lib_restful_interface("LteInfo", "GET", None, False)
        if not LibRESTfulAdapter.check_status_is_ok(ret_dict):
            not_support_caps.append("lte_config")
        else:
            if isinstance(ret_dict["message"], dict) and not ret_dict["message"].get("lte_enable"):
                not_support_caps.append("lte_config")

        ret_dict = LibRESTfulAdapter.lib_restful_interface("System", "GET", None, False)
        if not LibRESTfulAdapter.check_status_is_ok(ret_dict):
            not_support_caps.append("min_recovery")
        else:
            is_not_a500 = isinstance(ret_dict["message"], dict) and \
                          ret_dict["message"].get("SupportModel") != CommonConstants.ATLAS_500_A2
            if is_not_a500 or cmd_constants.OS_NAME != "EulerOS":
                not_support_caps.append("min_recovery")

        return [v for v in cap.get("product_capability", []) if v not in not_support_caps]

    @staticmethod
    def health_string_to_bool(health):
        return health == 'OK'

    @staticmethod
    def xb_to_b(string):
        try:
            if not isinstance(string, str):
                return int(string)
        except Exception as err:
            run_log.error("xb_to_b caught exception: %s", err)
            return 0

        try:
            if 'MB' in string:
                num_int = float(string.strip('MB')) * 1024 * 1024
            elif 'GB' in string:
                num_int = float(string.strip('GB')) * 1024 * 1024 * 1024
            elif 'KB' in string:
                num_int = float(string.strip('KB')) * 1024
            else:
                num_int = float(string.strip('B'))
        except Exception as err:
            run_log.error("xb_to_b caught exception: %s", err)
            return 0

        return int(num_int)

    @staticmethod
    def b_to_gb(string):
        if isinstance(string, str):
            return string

        try:
            value = round(float(string) / (1024 * 1024 * 1024), 2)
        except Exception as err:
            run_log.error("b_to_gb caught exception : %s", err)
            return "unknown"

        return str(value) + 'GB'

    @staticmethod
    def get_edge_system_info():
        # 获取 EdgeSystem信息
        with MidwareUris.sys_lock:
            try:
                system_resource = SERVICE_ROOT.systems.system_resource_collection.get_resource()
                resp_json = json.loads(system_resource)
                err_info = "Get edgesystem info failed."

                ret_dict = LibRESTfulAdapter.lib_restful_interface("System", "GET", None, False)
                ret = CommonRedfish.check_status_is_ok(ret_dict)
                if not ret:
                    run_log.error("Get edge system info failed.")
                    return

                res = json.loads(CommonRedfish.update_json_of_list(resp_json, ret_dict, err_info))
                MidwareUris._update_product_info_from_json_data(res)
                MidwareUris.resp_json_sys_info["product_capability_om"] = MidwareUris.get_product_capability()
                MidwareUris.resp_json_sys_info["system"]["host_name"] = socket.gethostname()
                MidwareUris._update_net_manager_params()
                profile_name = CommonRedfish.get_inactive_profile()
                MidwareUris.resp_json_sys_info["system"]["inactive_configuration"] = profile_name
                # 模板已校验数据，故字典值一定能取到
                MidwareUris.resp_json_sys_info["firmware_list"] = Firmware.firmware_list(res["Oem"]["Firmware"])
                MidwareUris.resp_json_sys_status["system"]["temperature"] = res["Oem"]["Temperature"]
                MidwareUris.resp_json_sys_status["system"]["power"] = "{}W".format(res["Oem"]["Power"])
                MidwareUris.resp_json_sys_status["system"]["voltage"] = "{}V".format(res["Oem"]["Voltage"])
                MidwareUris.resp_json_sys_status["system"]["cpu_heating"] = res["Oem"]["CpuHeating"]
                MidwareUris.resp_json_sys_status["system"]["disk_heating"] = res["Oem"]["DiskHeating"]
                MidwareUris.resp_json_sys_status["system"]["usb_hub_heating"] = res["Oem"]["UsbHubHeating"]
                MidwareUris.resp_json_sys_status["system"]["up_time"] = res["Oem"]["Uptime"]
                MidwareUris.resp_json_sys_status["system"]["date_time"] = res["Oem"]["Datetime"]
                MidwareUris.resp_json_sys_status["system"]["time_zone"] = res["Oem"]["DateTimeLocalOffset"]
                MidwareUris.resp_json_sys_status["system"]["cpu_usage"] = "{}%".format(res["Oem"]["CpuUsage"])
                MidwareUris.resp_json_sys_status["system"]["memory_usage"] = "{}%".format(res["Oem"]["MemoryUsage"])
                MidwareUris.resp_json_sys_status["system"]["health_status"] = MidwareUris.get_alarm_health_status()
            except Exception as err:
                run_log.error("get_edge_system_info caught exception : %s", err)

    @staticmethod
    def get_cpu_summary_info():
        # 获取cpu信息
        with MidwareUris.sys_lock:
            try:
                cpu_resource = SERVICE_ROOT.systems.cpu_processor_resource.get_resource()
                resp_json = json.loads(cpu_resource)
                err_info = "Get cpu summary info failed."
                info = "ProcessorsSummary"
                ret_dict = LibRESTfulAdapter.lib_restful_interface("System", "GET", None, False, info)
                ret = CommonRedfish.check_status_is_ok(ret_dict)
                if not ret:
                    run_log.error("Get cpu summary info failed.")
                    return
                res = json.loads(CommonRedfish.update_json_of_list(resp_json, ret_dict, err_info))
                MidwareUris.resp_json_sys_info["system"]["cpu_count"] = 1
                MidwareUris.resp_json_sys_info["system"]["cpu_model"] = res["Oem"]["CPUModel"]
            except Exception as err:
                run_log.error("get_cpu_summary_info caught exception : %s", err)

    @staticmethod
    def get_memory_summary_info():
        # 获取内存信息
        with MidwareUris.sys_lock:
            try:
                memory_resource = SERVICE_ROOT.systems.memory_resource.get_resource()
                resp_json = json.loads(memory_resource)
                err_info = "Get memory summary failed."
                info = "MemorySummary"
                ret_dict = LibRESTfulAdapter.lib_restful_interface("System", "GET", None, False, info)
                ret = CommonRedfish.check_status_is_ok(ret_dict)
                if ret:
                    res = json.loads(CommonRedfish.update_json_of_list(resp_json, ret_dict, err_info))
                    MidwareUris.resp_json_sys_info["system"]["memory_size"] = \
                        "{}GB".format(res["Oem"]["TotalSystemMemoryGiB"])
                else:
                    run_log.error("get memory summary info failed")
            except Exception as err:
                run_log.error("get_memory_summary_info caught exception : %s", err)

    @staticmethod
    def get_every_extend_info(extend_id):
        # 获取外设信息
        try:
            ret = ExtendedDeviceChecker().check_dict({"extend_id": extend_id})
            if not ret:
                run_log.error("extend_id invalid: %s", ret.reason)
                return

            err_info = "Get extend devices id:%s failed." % extend_id
            extend_device_resource = SERVICE_ROOT.systems.extended_devices_resource_member.get_resource()
            resp_json = json.loads(extend_device_resource)
            ret_dict = LibRESTfulAdapter.lib_restful_interface("Extend", "GET", None, False, extend_id)
            ret = CommonRedfish.check_status_is_ok(ret_dict)
            if not ret:
                run_log.error("Get every extend info failed")
                return

            res = json.loads(CommonRedfish.update_json_of_list(resp_json, ret_dict, err_info))
            extend = {}
            extend["name"] = res["Name"]
            extend["device_type"] = res["DeviceClass"]
            extend["device_name"] = res["DeviceName"]
            extend["manufacturer"] = res["Manufacturer"]
            if extend["device_type"] == "NPU":
                extend["model"] = "Atlas 200"
            else:
                extend["model"] = res["Model"]

            extend["serial_number"] = res["SerialNumber"]
            extend["firmware_version"] = res["FirmwareVersion"]
            extend["location"] = res["Location"]

            MidwareUris.resp_json_sys_info["extended_devices"].append(extend)

            extend_status = {}
            extend_status["status"] = {}
            extend_status["name"] = res["Name"]
            extend_status["status"]["state"] = res["Status"]["State"]
            extend_status["status"]["health"] = MidwareUris.health_string_to_bool(res["Status"]["Health"])

            MidwareUris.resp_json_sys_status["extended_devices"].append(extend_status)
        except Exception as err:
            run_log.error("get_every_extend_info caught exception : %s", err)

    @staticmethod
    def get_extend_info():
        # 获取外设信息
        with MidwareUris.sys_lock:
            try:
                MidwareUris.resp_json_sys_info["extended_devices"] = []
                MidwareUris.resp_json_sys_status["extended_devices"] = []

                extend_device_collection_resource = SERVICE_ROOT.systems. \
                    extended_devices_resource_collection.get_resource()
                resp_json = json.loads(extend_device_collection_resource)
                ret_dict = LibRESTfulAdapter.lib_restful_interface("Extend", "GET", None, True)
                err_info = "Query edge system extend devices resource failed."
                ret = CommonRedfish.check_status_is_ok(ret_dict)
                if not ret:
                    run_log.error("Get extend info failed.")
                    return
                res = json.loads(CommonRedfish.update_json_of_mem_count_and_odata_id(resp_json, ret_dict, err_info))
                for k in res["Members"]:
                    if k:
                        MidwareUris.get_every_extend_info(k['@odata.id'].split("/")[-1])
            except Exception as err:
                run_log.error("get_extend_info caught exception : %s", err)

    @staticmethod
    def get_ntp_info():
        # 获取ntp信息
        with MidwareUris.sys_lock:
            try:
                ntp_resource = SERVICE_ROOT.systems.ntp_service_resource.get_resource()
                resp_json = json.loads(ntp_resource)
                err_info = "Get ntp service failed."
                ret_dict = LibRESTfulAdapter.lib_restful_interface("NTPService", "GET", None, False)
                ret = CommonRedfish.check_status_is_ok(ret_dict)
                if not ret:
                    run_log.error("Get ntp info failed.")
                    return

                res = json.loads(CommonRedfish.update_json_of_list(resp_json, ret_dict, err_info))
                MidwareUris.resp_json_sys_info["ntp_server"]["service_enabled"] = res["ClientEnabled"]
                if res["ClientEnabled"]:
                    MidwareUris.resp_json_sys_info["ntp_server"]["preferred_server"] = res["NTPRemoteServers"]
                    MidwareUris.resp_json_sys_info["ntp_server"]["alternate_server"] = res["NTPRemoteServersbak"]
                else:
                    MidwareUris.resp_json_sys_info["ntp_server"]["preferred_server"] = ""
                    MidwareUris.resp_json_sys_info["ntp_server"]["alternate_server"] = ""

                MidwareUris.resp_json_sys_info["ntp_server"]["sync_net_manager"] = False

                if os.path.exists(FlagPathConstant.NTP_SYNC_NET_MANAGER):
                    MidwareUris.resp_json_sys_info["ntp_server"]["sync_net_manager"] = True
                    MidwareUris.resp_json_sys_info["ntp_server"]["preferred_server"] = ""
            except Exception as err:
                run_log.error("get_ntp_info caught exception : %s", err)

    @staticmethod
    def get_ai_temp_info():
        # 获取miniD温度信息
        try:
            ret_dict = LibRESTfulAdapter.lib_restful_interface("System", "GET", None, False)
            ret = CommonRedfish.check_status_is_ok(ret_dict)
            if ret:
                return ret_dict["message"]["AiTemperature"]
            # 获取失败按照协议，返回0
            run_log.error("Get AI processor temperature failed.")
            return 0
        except Exception as err:
            run_log.error("get_ai_temp_info caught exception : %s", err)
            return 0

    @staticmethod
    def get_extend_location_info(extend_name):
        for member in MidwareUris.resp_json_sys_info["extended_devices"]:
            if extend_name == member["name"]:
                return member["location"]
        return ""

    @staticmethod
    def get_npu_info():
        # 获取NPU信息
        with MidwareUris.sys_lock:
            try:
                MidwareUris.resp_json_sys_info["ai_processors"] = []
                MidwareUris.resp_json_sys_status["ai_processors"] = []
                npu_resource = SERVICE_ROOT.systems.ai_processor_resource.get_resource()
                resp_json = json.loads(npu_resource)
                err_info = "Get AI processor failed!"
                ret_dict = LibRESTfulAdapter.lib_restful_interface("AiProcessor", "GET", None, False)
                ret = CommonRedfish.check_status_is_ok(ret_dict)
                if not ret:
                    run_log.error("Get NPU info failed.")
                    return

                res = json.loads(CommonRedfish.update_json_of_list(resp_json, ret_dict, err_info))
                if int(res["Oem"]["Count"]) == 0:
                    MidwareUris.resp_json_sys_info["ai_processors"] = []
                    MidwareUris.resp_json_sys_status["ai_processors"] = []
                    return

                ai_processor = {
                    "id": 1,
                    "manufacturer": res["Manufacturer"],
                    "model": "Atlas 200",
                    "calc_ability": res["Oem"]["Capability"]["Calc"],
                    "ddr_capacity": "{}GB".format(res["Oem"]["Capability"]["Ddr"]),
                    "location": MidwareUris.get_extend_location_info("Atlas 200")
                }
                MidwareUris.resp_json_sys_info["ai_processors"].append(ai_processor)

                ai_processor_status = {
                    "id": 1,
                    "health": MidwareUris.health_string_to_bool(res["Status"]["Health"]),
                    "temperature": MidwareUris.get_ai_temp_info(),
                    "occupancy_rate": {
                        "ai_core": "{}%".format(res["Oem"]["OccupancyRate"]["AiCore"]),
                        "ai_cpu": "{}%".format(res["Oem"]["OccupancyRate"]["AiCpu"]),
                        "ctrl_cpu": "{}%".format(res["Oem"]["OccupancyRate"]["CtrlCpu"]),
                        "ddr_cap": "{}%".format(res["Oem"]["OccupancyRate"]["DdrUsage"]),
                        "ddr_bw": "{}%".format(res["Oem"]["OccupancyRate"]["DdrBandWidth"])
                    }
                }
                MidwareUris.resp_json_sys_status["ai_processors"].append(ai_processor_status)
            except Exception as err:
                run_log.error("get_npu_info caught exception : %s", err)

    @staticmethod
    def get_every_simple_storage_info(storage_id):
        try:
            err_info = "Get simple storage id:%s failed." % storage_id
            simple_storage_resource = SERVICE_ROOT.systems.simple_storage_resource.get_resource()
            resp_json = json.loads(simple_storage_resource)
            ret_dict = LibRESTfulAdapter.lib_restful_interface("Storage", "GET", None, False, storage_id)
            ret = CommonRedfish.check_status_is_ok(ret_dict)
            if not ret:
                run_log.error("Get every simple storage info failed.")
                return

            res = json.loads(CommonRedfish.update_json_of_list(resp_json, ret_dict, err_info))
            storage = {
                "name": storage_id,
                "type": res["Name"],
                "description": res["Description"],
                "devices": [],
            }

            devices = []
            for device in res["Devices"]:
                device_dict = {
                    "name": device["Name"],
                    "manufacturer": device["Manufacturer"],
                    "model": device["Model"],
                    "capacity_bytes": int(MidwareUris.xb_to_b(device["CapacityBytes"])),
                    # 上报给FD的消息中预留100M空间，继承原有功能，修改需要与FD对齐；
                    "reserved_bytes": 1024 * 1024 * 100,
                    "partition_style": device["PartitionStyle"],
                    "location": device["Location"]
                }

                if device_dict["name"] != "/dev/mmcblk0":
                    devices.append(device_dict)
                    continue

                MidwareUris.resp_json_sys_info["system"]["storage_size"] = MidwareUris.b_to_gb(device["CapacityBytes"])
                for partion in MidwareUris.resp_json_sys_info["partitions"]:
                    if partion["name"] in CommonRedfish.PRIMARY_EMMC_LIST:
                        device_dict["reserved_bytes"] = device_dict["reserved_bytes"] + int(partion["capacity_bytes"])
                devices.append(device_dict)

            storage["devices"] = devices
            MidwareUris.resp_json_sys_info["simple_storages"].append(storage)

            storage_status = {}
            storage_status["name"] = storage_id
            devices_status = []
            for device in res["Devices"]:
                device_dict = {}
                device_dict["name"] = device["Name"]
                device_dict["left_bytes"] = MidwareUris.xb_to_b(device["LeftBytes"])
                device_dict["health"] = MidwareUris.health_string_to_bool(device["Status"]["Health"])
                devices_status.append(device_dict)
            storage_status["devices"] = devices_status
            MidwareUris.resp_json_sys_status["simple_storages"].append(storage_status)
        except Exception as err:
            run_log.error("get_every_simple_storages_info caught exception : %s", err)

    @staticmethod
    def get_simple_storages_info():
        # 查询简单存储
        with MidwareUris.sys_lock:
            try:
                MidwareUris.resp_json_sys_info["simple_storages"] = []
                MidwareUris.resp_json_sys_status["simple_storages"] = []
                simple_storage_collection_resource = SERVICE_ROOT.systems. \
                    simple_storage_resource_collection.get_resource()
                resp_json = json.loads(simple_storage_collection_resource)
                err_info = "Get simple storages info failed."
                ret_dict = LibRESTfulAdapter.lib_restful_interface("Storage", "GET", None, True)
                ret = CommonRedfish.check_status_is_ok(ret_dict)
                if not ret:
                    run_log.error("Get simple storage info failed, use old one.")
                    # 如果查询失败使用上次的查询结果
                    MidwareUris.resp_json_sys_info["simple_storages"] = MidwareUris.simple_storages_info
                    MidwareUris.resp_json_sys_status["simple_storages"] = MidwareUris.simple_storages_status
                    return

                res = json.loads(CommonRedfish.update_json_of_mem_count_and_odata_id(resp_json, ret_dict, err_info))
                for k in res["Members"]:
                    if k:
                        time.sleep(1)
                        MidwareUris.get_every_simple_storage_info(k['@odata.id'].split("/")[-1])

                if MidwareUris.resp_json_sys_info["simple_storages"]:
                    MidwareUris.simple_storages_info = MidwareUris.resp_json_sys_info["simple_storages"]

                if MidwareUris.resp_json_sys_status["simple_storages"]:
                    MidwareUris.simple_storages_status = MidwareUris.resp_json_sys_status["simple_storages"]
            except Exception as err:
                run_log.error("get_simple_storages_info Caught exception : %s", err)

    @staticmethod
    def get_every_partition_info(partition_id):
        # 查询特定分区信息
        try:
            err_info = "Get Partition id:%s failed." % partition_id
            partition_resource = SERVICE_ROOT.systems.partition_resource.get_resource()
            resp_json = json.loads(partition_resource)
            ret_dict = LibRESTfulAdapter.lib_restful_interface("Partition", "GET", None, False, partition_id)
            ret = CommonRedfish.check_status_is_ok(ret_dict)
            if not ret:
                run_log.error("Get every partition info failed. ret: %s" % ret)
                return
            res = json.loads(CommonRedfish.update_json_of_list(resp_json, ret_dict, err_info))
            partition = {}
            if not res["Primary"]:
                partition["logic_name"] = partition_id
            else:
                partition["logic_name"] = None

            partition["name"] = res["Name"]
            partition["capacity_bytes"] = int(MidwareUris.xb_to_b(res["CapacityBytes"]))
            partition["storage_name"] = res["Links"][0]["Device"].split("/")[-1]
            partition["storage_location"] = res["Links"][0]["Location"]
            partition["storage_device"] = res["Links"][0]["DeviceName"]
            partition["file_system"] = res["FileSystem"]
            partition["mount_path"] = res["MountPath"]
            partition["system_partition_flag"] = res["Primary"]

            MidwareUris.resp_json_sys_info["partitions"].append(partition)

            partition_status = {}
            if not res["Primary"]:
                partition_status["logic_name"] = partition_id
            else:
                partition_status["logic_name"] = None

            partition_status["name"] = res["Name"]
            partition_status["free_bytes"] = int(MidwareUris.xb_to_b(res["FreeBytes"]))
            partition_status["health"] = MidwareUris.health_string_to_bool(res["Status"]["Health"])
            MidwareUris.resp_json_sys_status["partitions"].append(partition_status)
        except Exception as err:
            run_log.error("get_every_partition_info caught exception : %s", err)

    @staticmethod
    def get_partitions_info():
        # 查询分区信息
        with MidwareUris.sys_lock:
            try:
                MidwareUris.resp_json_sys_info["partitions"] = []
                MidwareUris.resp_json_sys_status["partitions"] = []
                err_info = "Get partitions failed."
                partition_collection_resource = SERVICE_ROOT.systems.partition_resource_collection.get_resource()
                resp_json = json.loads(partition_collection_resource)
                ret_dict = LibRESTfulAdapter.lib_restful_interface("Partition", "GET", None, True)
                ret = CommonRedfish.check_status_is_ok(ret_dict)
                if not ret:
                    run_log.error("Get partitions failed.")
                    return
                res = json.loads(CommonRedfish.update_json_of_mem_count_and_odata_id(resp_json, ret_dict, err_info))
                for k in res["Members"]:
                    if k:
                        time.sleep(1)
                        MidwareUris.get_every_partition_info(k['@odata.id'].split("/")[-1])
            except Exception as err:
                run_log.error("get_partitions_info caught exception : %s", err)

    @staticmethod
    def get_every_eth_info(eth_id):
        # 查询指定以太网信息
        try:
            err_info = "Get ethernet interface id:%s failed." % eth_id
            ethernet_resource = SERVICE_ROOT.systems.ethernet_getmembers_resource.get_resource()
            resp_json = json.loads(ethernet_resource)
            ret_dict = LibRESTfulAdapter.lib_restful_interface("Nic", "GET", None, False, eth_id)
            ret = CommonRedfish.check_status_is_ok(ret_dict)
            if not ret:
                run_log.error("Get ethernet interface info failed.")
                return

            res = json.loads(CommonRedfish.update_json_of_list(resp_json, ret_dict, err_info))
            nic = {}
            ipv4 = []
            nic["id"] = eth_id
            nic["name"] = res["Name"]
            nic["description"] = res["Description"]
            nic["permanent_mac"] = res["PermanentMACAddress"]
            nic["mac"] = res["MACAddress"]
            nic["interface_enabled"] = res["InterfaceEnabled"]
            nic["ipv4_addresses"] = []
            nic["location"] = MidwareUris.get_extend_location_info(nic["name"])

            for ip_address_dict in res["IPv4Addresses"]:
                ipv4_dict = {}
                ipv4_dict["address"] = ip_address_dict["Address"]
                ipv4_dict["subnet_mask"] = ip_address_dict["SubnetMask"]
                ipv4_dict["gateway"] = ip_address_dict["Gateway"]
                ipv4_dict["address_origin"] = ip_address_dict["AddressOrigin"]
                ipv4_dict["tag"] = ip_address_dict.get("Tag")
                ipv4.append(ipv4_dict)

            nic["ipv4_addresses"] = ipv4
            nic["name_servers"] = res["NameServers"]
            nic["adapter_type"] = res["Oem"]["AdapterType"]

            if res["Oem"]["LteDataSwitchOn"]:
                nic["lte_data_switch"] = 'enable'
            else:
                nic["lte_data_switch"] = 'disable'

            MidwareUris.resp_json_sys_info["ethernet_interfaces"].append(nic)

            nic_status = {"statistics": {}, "id": eth_id}
            if res["LinkStatus"] == 'LinkUp':
                nic_status["link_status"] = 'LinkUp'
            else:
                nic_status["link_status"] = 'LinkDown'

            nic_status["work_mode"] = res["Oem"]["WorkMode"]
            nic_status["statistics"]["send_packages"] = int(res["Oem"]["Statistic"]["SendPackages"])
            nic_status["statistics"]["recv_packages"] = int(res["Oem"]["Statistic"]["RecvPackages"])
            nic_status["statistics"]["error_packages"] = int(res["Oem"]["Statistic"]["ErrorPackages"])
            nic_status["statistics"]["drop_packages"] = int(res["Oem"]["Statistic"]["DropPackages"])
            MidwareUris.resp_json_sys_status["eth_statistics"].append(nic_status)
        except Exception as err:
            run_log.error("get_every_eth_info caught exception : %s", err)

    @staticmethod
    def get_eth_info():
        # 查询以太网信息
        with MidwareUris.sys_lock:
            try:
                MidwareUris.resp_json_sys_info["ethernet_interfaces"] = []
                MidwareUris.resp_json_sys_status["eth_statistics"] = []

                ethernet_collection_resource = SERVICE_ROOT.systems.ethernet_resource_collection.get_resource()
                resp_json = json.loads(ethernet_collection_resource)
                err_info = "Get ethernet interfaces resource failed."
                ret_dict = LibRESTfulAdapter.lib_restful_interface("Nic", "GET", None, True)
                ret = CommonRedfish.check_status_is_ok(ret_dict)
                if not ret:
                    run_log.error("Get ethernet interfaces resource failed.")
                    return

                res = json.loads(CommonRedfish.update_json_of_mem_count_and_odata_id(resp_json, ret_dict, err_info))
                for k in res["Members"]:
                    if k:
                        eth_id = k['@odata.id'].split("/")[-1]
                        MidwareUris.get_every_eth_info(eth_id)
            except Exception as err:
                run_log.error("get_eth_info caught exception : %s", err)

    @staticmethod
    def get_dns_and_host_map_info():
        try:
            ret_dict = LibRESTfulAdapter.lib_restful_interface("DomainCfg", "GET", None, False)
            ret = CommonRedfish.check_status_is_ok(ret_dict)
            if ret:
                fd_server_name = FdCfgManager.get_cur_fd_server_name()
                with MidwareUris.sys_lock:
                    MidwareUris.resp_json_sys_info["static_host_list"] = \
                        [
                            item
                            for item in ret_dict["message"]["static_host_list"]
                            if item.get("name") != fd_server_name
                        ]
                    MidwareUris.resp_json_sys_info["name_server"] = ret_dict["message"]["name_server"]
            else:
                run_log.error("Get domainCfg failed: {}".format(ret_dict))
        except Exception as err:
            run_log.error("get_dns_and_host_map_info caught exception : %s", err)

    @staticmethod
    def get_passwd_validity_info():
        try:
            recv_conf_dict = {"oper_type": UserManagerConstants.OPER_TYPE_GET_ACCOUNT_EXPIRATION_DAY}
            ret_dict = UserManager().get_all_info(recv_conf_dict)
            if CommonRedfish.check_status_is_ok(ret_dict):
                with MidwareUris.sys_lock:
                    MidwareUris.resp_json_sys_info["security_policy"] = {
                        "password_validity": str(ret_dict["message"]["PasswordExpirationDays"])
                    }
            else:
                run_log.error("Get passwd validity info failed: {}".format(ret_dict))
        except Exception:
            run_log.error("get_passwd_validity_info caught exception.")

    @staticmethod
    def get_session_timeout():
        try:
            recv_conf_dict = {"oper_type": UserManagerConstants.OPER_TYPE_GET_SESSION_TIMEOUT}
            ret_dict = SessionManager().get_all_info(recv_conf_dict)
            # 判断返回的字典是否正确
            if CommonRedfish.check_status_is_ok(ret_dict):
                with MidwareUris.sys_lock:
                    MidwareUris.resp_json_sys_info["security_policy"]["session_timeout"] = \
                        str(ret_dict["message"]["SessionTimeout"])
            else:
                run_log.error("Get session timeout info failed: %s", ret_dict.get("message"))
        except Exception as err:
            run_log.error("Get session timeout info caught exception(%s).", err)

    @staticmethod
    def get_cert_alarm_time():
        try:
            ret_dict = LibRESTfulAdapter.lib_restful_interface("CertAlarmTime", "GET", None, False)
            # 判断返回的字典是否正确
            if CommonRedfish.check_status_is_ok(ret_dict):
                with MidwareUris.sys_lock:
                    MidwareUris.resp_json_sys_info["security_policy"]["cert_alarm_time"] = \
                        str(ret_dict["message"]["CertAlarmTime"])
            else:
                run_log.error("Get cert alarm time info failed: %s", ret_dict.get("message"))
        except Exception as err:
            run_log.error("Get cert alarm time info caught exception(%s).", err)

    @staticmethod
    def get_security_load_config():
        try:
            ret_dict = LibRESTfulAdapter.lib_restful_interface("SecurityLoad", "GET", None, False)
            # 判断返回的字典是否正确
            if CommonRedfish.check_status_is_ok(ret_dict):
                with MidwareUris.sys_lock:
                    MidwareUris.resp_json_sys_info["security_policy"]["security_load"] = \
                        ret_dict["message"]["load_cfg"]
            else:
                run_log.error("Get security load config failed: %s", ret_dict.get("message"))
        except Exception as err:
            run_log.error("Get security load config caught exception(%s).", err)

    @staticmethod
    def get_accounts_info():
        try:
            recv_conf_dict = {'oper_type': UserManagerConstants.OPER_TYPE_GET_USER_LIST}
            ret_dict = UserManager().get_all_info(recv_conf_dict)
            if CommonRedfish.check_status_is_ok(ret_dict):
                with MidwareUris.sys_lock:
                    MidwareUris.resp_json_sys_info["accounts"] = ret_dict["message"]["result"]
            else:
                run_log.error("Get accounts info failed: {}".format(ret_dict))
        except Exception:
            run_log.error("get_accounts_info caught exception.")

    @staticmethod
    def get_lte_info():
        try:
            MidwareUris.resp_json_sys_status["lte_info"] = []

            err_info = "Get LTE info failed."
            lte_info_json = json.loads(LteStatusInfoResourceSerializer().service.get_resource())
            ret_dict = LibRESTfulAdapter.lib_restful_interface("LteInfo", "GET", None, False)
            ret = CommonRedfish.check_status_is_ok(ret_dict)
            if not ret or not isinstance(ret_dict["message"], dict):
                run_log.error(err_info)
                return

            if not ret_dict["message"].get("lte_enable"):
                return

            res = json.loads(CommonRedfish.update_json_of_list(lte_info_json, ret_dict, err_info))

            lte_item = dict()
            lte_item["default_gateway"] = res["default_gateway"]
            lte_item["lte_enable"] = res["lte_enable"]
            lte_item["sim_exist"] = res["sim_exist"]
            lte_item["state_lte"] = res["state_lte"]
            lte_item["state_data"] = res["state_data"]
            lte_item["network_signal_level"] = res["network_signal_level"]
            lte_item["network_type"] = res["network_type"]
            lte_item["ip_addr"] = res["ip_addr"]
            lte_item["apn_info"] = []
            lte_item["apn_info"].append(MidwareUris.get_apn_info())
            MidwareUris.resp_json_sys_status["lte_info"].append(lte_item)
        except Exception:
            run_log.error("get_lte_info caught exception.")

    @staticmethod
    def get_apn_info():
        try:
            apn_info_json = json.loads(LteConfigInfoResourceSerializer().service.get_resource())
            ret_dict = LibRESTfulAdapter.lib_restful_interface("LteConfigInfo", "GET", None, False)
            ret = CommonRedfish.check_status_is_ok(ret_dict)
            err_info = "Get APN info failed."
            if not ret:
                run_log.error(err_info)
                return
            res = json.loads(CommonRedfish.update_json_of_list(apn_info_json, ret_dict, err_info))
            apn_item = dict()
            apn_item["apn_name"] = res["apn_name"]
            apn_item["apn_user"] = res["apn_user"]
            apn_item["auth_type"] = res["auth_type"]
            apn_item["mode_type"] = res["mode_type"]
            return apn_item
        except Exception:
            run_log.error("get_apn_info caught exception.")

    @staticmethod
    def get_access_control_info():
        try:
            ret_json = json.loads(AccessControlSerializer().service.get_resource())
            ret_dict = LibRESTfulAdapter.lib_restful_interface("AccessControl", "GET", None, False)

            ret = CommonRedfish.check_status_is_ok(ret_dict)
            if not ret:
                err_info = "Get access control info failed."
                run_log.error(err_info)
                return

            success_msg = "Get access control info successfully."
            res = json.loads(CommonRedfish.update_json_of_list(ret_json, ret_dict, success_msg))
            MidwareUris.resp_json_sys_info["security_policy"]["web_access"] = res["web_access"]
            MidwareUris.resp_json_sys_info["security_policy"]["ssh_access"] = res["ssh_access"]
        except Exception:
            run_log.error("get access control info caught exception.")

    @staticmethod
    def alarm_delete(payload_publish, res):
        """
        从内存变量alarm_lists中删除已经恢复的告警信息，并通报信息给FD
        """
        alarm_need_delete_id = set()
        for alarm_id_exists in MidwareUris.alarm_lists:
            alarm_need_delete_info = []
            for i, item in enumerate(MidwareUris.alarm_lists[alarm_id_exists]):
                is_find = MidwareUris.find_alarm(alarm_id_exists, item, res)
                if is_find:
                    item["notificationType"] = "clear"
                    item["sn"] = str(MidwareUris.alarm_sn)
                    item["timestamp"] = datetime.datetime.utcnow().replace(
                        microsecond=0, tzinfo=datetime.timezone.utc).isoformat()
                    MidwareUris.alarm_sn = MidwareUris.alarm_sn + 1
                    payload_publish["alarm"].append(copy.deepcopy(item))
                    alarm_need_delete_info.append(i)
            if len(alarm_need_delete_info) == len(MidwareUris.alarm_lists[alarm_id_exists]):
                alarm_need_delete_id.add(alarm_id_exists)
                continue
            for index in sorted(alarm_need_delete_info, reverse=True):
                del MidwareUris.alarm_lists[alarm_id_exists][index]

        for alarm_id in alarm_need_delete_id:
            MidwareUris.alarm_lists.pop(alarm_id)

    @staticmethod
    def find_alarm(alarm_id_exist, alarm_exist, alar_messages):
        for alarm_event in alar_messages:
            if alarm_id_exist == alarm_event["AlarmId"] and alarm_exist["resource"] == alarm_event["AlarmInstance"]:
                return False
        return True

    @staticmethod
    def add_alarm_lists(alarm_id, alarm_messages):
        if alarm_id not in MidwareUris.alarm_lists:
            MidwareUris.alarm_lists[alarm_id] = []
            MidwareUris.alarm_lists.get(alarm_id).append(copy.deepcopy(alarm_messages))
            return

        alarm_infos = []
        for alarm_info in MidwareUris.alarm_lists.get(alarm_id):
            alarm_infos.append(alarm_info["resource"])
        if alarm_messages["resource"] not in alarm_infos:
            MidwareUris.alarm_lists.get(alarm_id).append(copy.deepcopy(alarm_messages))

    @staticmethod
    def get_ntp_config():
        try:
            ntp_service_reource = SERVICE_ROOT.systems.ntp_service_resource.get_resource()
            resp_json = json.loads(ntp_service_reource)
            err_info = "Get NTPService failed."
            ret_dict = LibRESTfulAdapter.lib_restful_interface("NTPService", "GET", None, False)

            ret = CommonRedfish.check_status_is_ok(ret_dict)
            if ret:
                succ_info = "Get NTPService Successfully."
                run_log.info(succ_info)
                res = json.loads(CommonRedfish.update_json_of_list(resp_json, ret_dict, succ_info))
                return [0, res]

            resp_info = MidwareErrCode.midware_generate_err_msg(
                MidwareErrCode.midware_config_ntp_common_err, err_info)
            run_log.error("%s", resp_info)
            return [-1, resp_info]
        except Exception as err:
            run_log.error("Get NTPService error: %s", err)
            error_resp = MidwareErrCode.midware_generate_err_msg(
                MidwareErrCode.midware_config_ntp_common_err, "Get NTPService error.")
            return [-1, error_resp]

    @staticmethod
    def set_ntp_config(ntp_server_ip):
        ntp_cfg = {
            "ClientEnabled": True,
            "ServerEnabled": False,
            "Target": "Client",
            "NTPLocalServers": "",
            "NTPRemoteServers": ntp_server_ip,
            "NTPRemoteServersbak": "",
        }
        try:
            ret_dict = LibRESTfulAdapter.lib_restful_interface("NTPService", "PATCH", ntp_cfg, False)
            ret = CommonRedfish.check_status_is_ok(ret_dict)
            if ret:
                resp_info = f"NTP sync success, ip: {ntp_server_ip}"
                run_log.info("%s", resp_info)
                return [0, resp_info]

            err_info = "NTP sync failed"
            resp_info = MidwareErrCode.midware_generate_err_msg(
                MidwareErrCode.midware_config_ntp_common_err, err_info)
            run_log.error("%s", resp_info)
            return [-1, resp_info]
        except Exception as err:
            run_log.error("Sync NTP service error:%s", err)
            err_resp = MidwareErrCode.midware_generate_err_msg(
                MidwareErrCode.midware_config_ntp_common_err, "Sync NTP service error.")
            return [-1, err_resp]

    # NTP同步网管时间处理
    @staticmethod
    def snyc_ntp_config():
        try:
            if not os.path.exists(FlagPathConstant.NTP_SYNC_NET_MANAGER):
                resp_info = "Ntp is not config to sync net_manager"
                run_log.info("%s", resp_info)
                return [0, resp_info]

            result = MidwareUris.get_ntp_config()
            if result[0] != 0:
                return [-1, result[1]]

            ntp_cfg = result[1]
            fd_ip = FdCfgManager.get_cur_fd_ip()
            if ntp_cfg["NTPRemoteServers"] == fd_ip:
                resp_info = "NTP sync already OK"
                run_log.info("%s", resp_info)
                return [0, resp_info]

            return MidwareUris.set_ntp_config(fd_ip)
        except Exception as err:
            err_info = f"Sync NTP config failed: {err}"
            resp_info = MidwareErrCode.midware_generate_err_msg(
                MidwareErrCode.midware_config_ntp_common_err, err_info)
            run_log.error("%s", resp_info)
            return [-1, resp_info]

    @staticmethod
    def mid_ware_add_route():

        # 系统静态信息定时任务入口
        @midWare.route(r'espmanager/SysInfoProc')
        def system_info_proc(extend_system_info_funcs: List = None):
            time.sleep(0.5)
            MidwareUris.get_edge_system_info()
            time.sleep(0.5)
            MidwareUris.get_cpu_summary_info()
            time.sleep(0.5)
            MidwareUris.get_memory_summary_info()
            time.sleep(0.5)
            MidwareUris.get_extend_info()
            time.sleep(0.5)
            MidwareUris.get_ntp_info()
            time.sleep(0.5)
            MidwareUris.get_npu_info()
            time.sleep(1)
            MidwareUris.get_partitions_info()
            time.sleep(1)
            MidwareUris.get_simple_storages_info()
            time.sleep(1)
            MidwareUris.get_eth_info()
            time.sleep(0.5)
            MidwareUris.get_dns_and_host_map_info()
            time.sleep(0.5)
            MidwareUris.get_passwd_validity_info()
            time.sleep(0.5)
            MidwareUris.get_session_timeout()
            time.sleep(0.5)
            MidwareUris.get_cert_alarm_time()
            time.sleep(0.5)
            MidwareUris.get_security_load_config()
            time.sleep(0.5)
            MidwareUris.get_accounts_info()
            time.sleep(0.5)
            MidwareUris.get_lte_info()
            time.sleep(0.5)
            MidwareUris.get_access_control_info()
            time.sleep(0.5)
            if extend_system_info_funcs:
                for extend_system_info_func in extend_system_info_funcs:
                    ret = extend_system_info_func()
                    if isinstance(ret, dict):
                        MidwareUris.resp_json_sys_info.update(ret)
            SysInfoTaskStatus.set()

        # 系统静态消息查询入口
        @midWare.route(r'espmanager/SysInfo')
        def system_info():
            with MidwareUris.sys_lock:
                resp_system_info = copy.deepcopy(MidwareUris.resp_json_sys_info)
            return [0, resp_system_info]

        # 系统状态定时任务入口
        @midWare.route(r'espmanager/SysStatusProc')
        def system_status_proc():
            time.sleep(0.5)
            MidwareUris.get_edge_system_info()

        # 系统状态消息查询入口
        @midWare.route(r'espmanager/SysStatus')
        def system_status():
            with MidwareUris.sys_lock:
                resp_system_status = copy.deepcopy(MidwareUris.resp_json_sys_status)
            return [0, resp_system_status]

        @midWare.route(Topic.SUB_RESET_ALARM)
        @FdCommonMethod.fd_operational_log("Reset alarm")
        def re_alarm(payload):
            run_log.info(f"enter re_alarm [{Topic.SUB_RESET_ALARM}]")
            if not ReAlarmChecker().check(payload).success:
                run_log.error("SUB_RESET_ALARM param error.")
                return [-1, "Rearm failed"]

            # 未到连接就绪状态，不上报事件
            connect_ready = FdConnectCheck.check_fd_connect_status_ready()
            if connect_ready:
                # 连接FD就绪之后，等待120s，上报事件
                threading.Thread(target=MidwareProc.handling_msg_from_fd).start()

            with MidwareUris.alarm_lock:
                MidwareUris.alarm_lists = {}
                MidwareUris.alarm_sn = 1
                return [0, "Rearm successful"]

        # 硬件告警/事件查询入口
        @midWare.route(r'espmanager/Alarm')
        def alarm(payload):
            payload_publish = {"alarm": []}
            alarm_messages = {}
            alarm_severity_map = {"0": "CRITICAL", "1": "MAJOR", "2": "MINOR"}
            try:
                alarm_info_resource = json.loads(AlarmInfoResourceSerializer().service.get_resource())
                err_info = "Get Alarm service resource failed."
                # 获取OM告警
                om_ret_dict = LibRESTfulAdapter.lib_restful_interface("Alarm", "GET", None, False)
                if not CommonRedfish.check_status_is_ok(om_ret_dict):
                    run_log.warning("Get om alarm info failed.")
                    om_alarms = []
                else:
                    om_alarms = json.loads(
                        CommonRedfish.update_json_of_list(alarm_info_resource, om_ret_dict, err_info)
                    )["AlarMessages"]

                # 获取MEF告警
                from wsclient.ws_client_mef import WsClientMef
                mef_alarms = WsClientMef().get_mef_alarm_info()
                with MidwareUris.alarm_lock:
                    # 处理OM告警
                    for alarm_event in om_alarms:
                        alarm_info = MidwareUris._get_alarm_info(alarm_event["AlarmId"])
                        alarm_messages["type"] = "alarm"
                        alarm_messages["notificationType"] = "alarm"
                        alarm_messages["alarmId"] = "0x" + alarm_event["AlarmId"]
                        alarm_messages["alarmName"] = alarm_info.get("name", alarm_event["AlarmName"])
                        alarm_messages["resource"] = alarm_event["AlarmInstance"]
                        alarm_messages["perceivedSeverity"] = \
                            alarm_severity_map.get(alarm_event["PerceivedSeverity"], "unknown")
                        alarm_messages["timestamp"] = datetime.datetime. \
                            utcfromtimestamp(int(alarm_event["Timestamp"])). \
                            replace(microsecond=0, tzinfo=datetime.timezone.utc).isoformat()
                        alarm_messages["detailedInformation"] = \
                            alarm_info.get("detailedInformation", alarm_event["AlarmName"])
                        alarm_messages["sn"] = str(MidwareUris.alarm_sn)
                        MidwareUris.alarm_sn = MidwareUris.alarm_sn + 1
                        alarm_messages["suggestion"] = alarm_info.get("dealsuggestion", "unknown")
                        alarm_messages["reason"] = alarm_info.get("reason", "unknown")
                        alarm_messages["impact"] = alarm_info.get("impact", "unknown")
                        payload_publish["alarm"].append(copy.deepcopy(alarm_messages))
                        MidwareUris.add_alarm_lists(alarm_event["AlarmId"], alarm_messages)

                    MidwareUris.alarm_delete(payload_publish, om_alarms)

                    # 处理MEF告警
                    for alarm_event in mef_alarms:
                        alarm_event["sn"] = str(MidwareUris.alarm_sn)
                        MidwareUris.alarm_sn = MidwareUris.alarm_sn + 1
                        payload_publish["alarm"].append(alarm_event)

                return [0, payload_publish]
            except Exception as err:
                run_log.error("Get Alarm service resource error: {}".format(err))
                return [-1, "Get Alarm service resource error."]

        # 硬件告警/事件查询入口
        @midWare.route(r'espmanager/Event')
        def event(event_type="all"):
            try:
                ret_dict = LibRESTfulAdapter.lib_restful_interface("Event", "GET", None, False, event_type)
            except Exception as err:
                run_log.error("Get Event service resource error: %s", err)
                return [-1, "Get Event service resource error."]
            ret = CommonRedfish.check_status_is_ok(ret_dict)
            if not ret:
                err_info = "Get Event service resource failed."
                run_log.error(err_info)
                return [-1, err_info]
            return [0, ret_dict.get("message", {}).get("result")]

        # 对接FD，一机一密落盘
        @midWare.route(r'espmanager/netmanager')
        @FdCommonMethod.fd_operational_log("Config net manager info")
        def config_netmanager(msg_obj: FdMsgData):
            payload_failed = {
                "topic": "netmanager", "percentage": "0%", "result": "failed",
                "reason": AppCommonMethod.COM_ERR_REASON_TO_FD
            }

            if MidwareUris.net_manager_lock.locked():
                run_log.warning("Netmanager modify is busy")
                payload_failed["reason"] = f"ERR.{MidwareErrCode.midware_resource_busy}, Netmanager modify is busy."
                return [-1, payload_failed]

            with MidwareUris.net_manager_lock:
                try:
                    ret = CommonRedfish.check_net_manager_info(msg_obj)
                except Exception as err:
                    run_log.error("check net manager info failed, caught exception: %s", err)
                    payload_failed["reason"] = f"ERR.{MidwareErrCode.midware_input_parameter_invalid}" \
                                               f", Invalid parameters."
                    return [-1, payload_failed]

                if ret[0] != 0:
                    payload_failed["reason"] = f"ERR.{MidwareErrCode.midware_input_parameter_invalid}" \
                                               f", Invalid parameters."
                    return [-1, payload_failed]

                cloud_password_encrypt = NetManager.encrypt_cloud_pwd(msg_obj.content.get("password"))
                # 一机一密下发之后，状态status改为connected状态
                status = "connected"
                update_data = {
                    "cloud_user": msg_obj.content.get("account"),
                    "cloud_pwd": cloud_password_encrypt,
                }
                address = msg_obj.content.get("address")
                if address:
                    # 如果是切换网管IP，status状态更新为初始态
                    update_data.update({"ip": address})
                    if FdConnectCheck.is_switch_fd(address):
                        if not docker_root_is_mounted():
                            run_log.error("Switch FusionDirector manage failed: docker root dir is not mounted.")
                            return [-1, payload_failed]
                        status = ""

                update_data.update({"status": status})
                try:
                    NetCfgManager().update_net_cfg_info(update_data)
                except Exception as err:
                    run_log.error("update net manager info failed: %s", err)
                    return [-1, payload_failed]

                payload_success = {"topic": "netmanager", "percentage": "100%", "result": "success", "reason": ""}
                return [0, payload_success]

        # FD证书管理，导入多证书
        @midWare.route(r'espmanager/cert_update')
        @FdCommonMethod.fd_operational_log("Update cert")
        def fd_cert_import_handle(payload):
            payload_publish = PayloadPublish(topic="cert_update", reason="import fd cert failed")
            if MidwareUris.import_cert_lock.locked():
                run_log.error("Importing FD cert is busy.")
                payload_publish.reason = f"ERR.{MidwareErrCode.midware_resource_busy}, Importing FD cert is busy."
                return [-1, payload_publish.to_dict()]

            try:
                with MidwareUris.import_cert_lock:
                    run_log.info("start import cert handle")
                    param_check_error = "input param check failed reason is %s"
                    check_ret = MidwareUris.check_external_param(
                        FdCertImportChecker, payload, payload_publish.to_dict(), param_check_error)
                    if check_ret[0] != 0:
                        return check_ret

                    if FdCertManager().is_cert_pool_full():
                        err_msg = "The certificate pool is full. please delete unused cert firstly."
                        run_log.error(err_msg)
                        payload_publish.reason = f"ERR.{MidwareErrCode.midware_import_cert_error}, {err_msg}"
                        return [-1, payload_publish.to_dict()]

                    if FdCertManager().cert_is_in_using(payload.get("cert_name")):
                        run_log.error("cert is using, cannot update the certificate with the same name.")
                        payload_publish.reason = f"ERR.{MidwareErrCode.midware_import_cert_error}, " \
                                                 f"cert is using, cannot update the certificate with the same name."
                        return [-1, payload_publish.to_dict()]

                    ImportCert(NetManagerConstants.FUSION_DIRECTOR, payload.get("cert_name")).import_deal(
                        base64.b64decode(payload.get("content")).decode()
                    )
                    # 导入证书时查询一次FD证书即将过期告警
                    FdCfgManager().check_cert_status()
                    run_log.info("fd cert import handle success")
                    payload_publish.percentage = "100%"
                    payload_publish.result = "success"
                    payload_publish.reason = "import fd cert success"
                    return [0, payload_publish.to_dict()]
            except NetManagerException as err:
                run_log.error("net manager exception: %s", err.err_msg)
                payload_publish.reason = err.err_msg
                payload_publish.reason = f"ERR.{MidwareErrCode.midware_import_cert_error}, {err.err_msg}"
                return [-1, payload_publish.to_dict()]
            except Exception as err:
                run_log.error("fd cert import handle failed, find exception: %s", err)
                payload_publish.reason = f"ERR.{MidwareErrCode.midware_import_cert_error}, fd cert import handle failed"
                return [-1, payload_publish.to_dict()]

        # FD证书管理，导入吊销列表
        @midWare.route(r'espmanager/crl_update')
        @FdCommonMethod.fd_operational_log("Update crl")
        def fd_crl_import_handle(payload):
            payload_publish = PayloadPublish(topic="crl_update", reason="import fd crl failed")
            payload_publish.reason = f"ERR.{MidwareErrCode.midware_common_err}, Importing FD crl is failed."
            if MidwareUris.import_crl_lock.locked():
                run_log.error("Importing FD crl is busy.")
                payload_publish.reason = f"ERR.{MidwareErrCode.midware_resource_busy}, Importing FD crl is busy."
                return [-1, payload_publish.to_dict()]

            try:
                with MidwareUris.import_crl_lock:
                    run_log.info("start import crl handle")
                    param_check_error = "input param check failed reason is %s"
                    check_ret = MidwareUris.check_external_param(
                        ImportCrlChecker, payload, payload_publish.to_dict(), param_check_error)
                    if check_ret[0] != 0:
                        return check_ret

                    ImportCrl().import_deal(base64.b64decode(payload.get("content")).decode())
                    run_log.info("fd crl import handle success")
                    payload_publish.percentage = "100%"
                    payload_publish.result = "success"
                    payload_publish.reason = "import fd crl success"
                    return [0, payload_publish.to_dict()]
            except NetManagerException as err:
                run_log.error("net manager exception: %s", err.err_msg)
                return [-1, payload_publish.to_dict()]
            except Exception as err:
                run_log.error("fd crl import handle failed, find exception: %s", err)
                return [-1, payload_publish.to_dict()]

        # FD证书管理，删除证书
        @midWare.route(r'espmanager/cert_delete')
        @FdCommonMethod.fd_operational_log("Delete cert")
        def fd_cert_delete_handle(payload):
            payload_publish = PayloadPublish(topic="cert_delete", reason="delete cert failed")
            payload_publish.reason = f"ERR.{MidwareErrCode.midware_common_err}, Delete cert failed."
            if MidwareUris.cert_delete_lock.locked():
                run_log.error("cert delete is busy.")
                payload_publish.reason = f"ERR.{MidwareErrCode.midware_resource_busy}, Delete cert is busy."
                return [-1, payload_publish.to_dict()]

            try:
                with MidwareUris.cert_delete_lock:
                    run_log.info("start delete cert handle")
                    param_check_error = "input param check failed reason is %s"
                    check_ret = MidwareUris.check_external_param(
                        FdCertDeleteChecker, payload, payload_publish.to_dict(), param_check_error)
                    if check_ret[0] != 0:
                        return check_ret

                    cert_mgr = FdCertManager()
                    if cert_mgr.cert_is_in_using(payload.get("cert_name")):
                        run_log.error("cert is using, cannot be delete")
                        payload_publish.reason = "cert is using, cannot be delete"
                        return [-1, payload_publish.to_dict()]

                    result = cert_mgr.del_by_name(payload.get("cert_name"))
                    if result == 0:
                        run_log.error("not found %s cert", payload.get("cert_name"))
                        return [-1, payload_publish.to_dict()]

                    # 删除证书时查询一次FD证书即将过期告警
                    FdCfgManager().check_cert_status()
                    run_log.info("delete cert handle success")
                    payload_publish.percentage = "100%"
                    payload_publish.result = "success"
                    payload_publish.reason = "delete cert success"
                    return [0, payload_publish.to_dict()]
            except NetManagerException as err:
                run_log.error("net manager exception: %s", err.err_msg)
                return [-1, payload_publish.to_dict()]
            except Exception as err:
                run_log.error("delete cert handle failed, find exception: %s", err)
                return [-1, payload_publish.to_dict()]

        # FD证书管理，查询证书
        @midWare.route(r'espmanager/cert_query')
        @FdCommonMethod.fd_operational_log("Query cert info")
        def fd_cert_query_handle(payload):
            payload_publish = {
                "root_certificate": {
                    "cert_is_full": False,
                    "cert_lists": [],
                }
            }
            if MidwareUris.cert_query_lock.locked():
                run_log.error("cert query is busy.")
                return [-1, payload_publish]

            try:
                with MidwareUris.cert_query_lock:
                    run_log.info("start query cert handle")
                    cert_lists = list(FdCertManager().cert_info_for_fd_generator())
                    if not cert_lists:
                        run_log.error("cert content is empty, please check and upload CA certificate")
                        return [-1, payload_publish]

                    run_log.info("query cert handle success")
                    payload_publish["root_certificate"]["cert_is_full"] = FdCertManager().is_cert_pool_full()
                    payload_publish["root_certificate"]["cert_lists"] = cert_lists
                    return [0, payload_publish]
            except NetManagerException as err:
                run_log.error("net manager exception: %s", err.err_msg)
                return [-1, payload_publish]
            except Exception as err:
                run_log.error("query cert handle failed, find exception: %s", err)
                return [-1, payload_publish]

        # 系统复位入口
        @midWare.route(r'espmanager/ComputerSystemReset')
        @FdCommonMethod.fd_operational_log("Restart system")
        def computer_system_reset(payload):
            payload_publish = {
                "restartable": "false",
                "reason": "resource is busy"
            }
            try:
                check_ret = MidwareUris.check_external_param(
                    ComputerSystemResetChecker, payload, payload_publish, "Restart system check param failed. %s")
                if check_ret[0] != 0:
                    return check_ret

                reset_type = payload.get("restart_method", None)

                if reset_type is None:
                    run_log.error("Input parameter is invalid.")
                    payload_publish["restartable"] = "false"
                    payload_publish["reason"] = MidwareErrCode.midware_generate_err_msg(
                        MidwareErrCode.midware_input_parameter_invalid, "Input parameter invalid!")
                    return [-1, payload_publish]

                recv_conf_dict = {}
                if reset_type == "Graceful":
                    recv_conf_dict["ResetType"] = "GracefulRestart"
                else:
                    run_log.error("Parameter reset_type error.")
                    payload_publish["restartable"] = "false"
                    payload_publish["reason"] = MidwareErrCode.midware_generate_err_msg(
                        MidwareErrCode.midware_input_parameter_invalid, "Parameter reset_type error.")
                    return [-1, payload_publish]

                if RedfishGlobals.high_risk_exclusive_lock.locked():
                    payload_publish["reason"] = MidwareErrCode.midware_generate_err_msg(
                        MidwareErrCode.midware_resource_busy, "Resource is busy")
                    return [-1, payload_publish]

                with RedfishGlobals.high_risk_exclusive_lock:
                    ret_dict = LibRESTfulAdapter.lib_restful_interface("ExclusiveStatus", "GET", None, False)
                    if not LibRESTfulAdapter.check_status_is_ok(ret_dict):
                        return [-1, payload_publish]

                    if not isinstance(ret_dict.get("message"), dict) or not \
                            isinstance(ret_dict.get("message").get("system_busy"), bool) or \
                            ret_dict.get("message").get("system_busy"):
                        run_log.error("System is busy, operate failed.")
                        return [-1, payload_publish]

                    # 注意：此接口内部用定时器延迟重启，所以总能收到返回值
                    payload_publish["restartable"] = "true"
                    payload_publish["reason"] = "System is restartable"
                    resp_msg = FdMsgData.gen_ws_msg_obj(payload_publish, "websocket/restart_result")
                    publish_ws_msg(resp_msg)
                    run_log.info("System restart now.")
                    ret_dict = LibRESTfulAdapter.lib_restful_interface("Actions", "POST", recv_conf_dict, False)
                    if CommonRedfish.check_status_is_ok(ret_dict):
                        run_log.info("Restart system [%s] success.", reset_type)
                        return [0, payload_publish]

                    run_log.error("Restart system [%s] failed.", reset_type)
                    payload_publish["restartable"] = "false"
                    payload_publish["reason"] = MidwareErrCode.midware_generate_err_msg(
                        MidwareErrCode.midware_common_err, ret_dict['message'])
                    return [-1, payload_publish]
            except Exception as err:
                run_log.error("Restart system failed, %s", err)
                payload_publish["reason"] = MidwareErrCode.midware_generate_err_msg(
                    MidwareErrCode.midware_common_err, "System restart error.")
                return [-1, payload_publish]

        # 信息收集入口
        @midWare.route(r'espmanager/InfoCollect')
        @FdCommonMethod.fd_operational_log("Collect info")
        def info_collect(payload):
            def publish_info_collect_progress():
                if collect_done == 1:
                    run_log.info("collect_done: %s", collect_done)
                    return

                if collect_done == 0:
                    ret_dict = LibRESTfulAdapter.lib_restful_interface("LoggerCollect", "GET", None, False)

                    if not LibRESTfulAdapter.check_status_is_ok(ret_dict):
                        return
                    info_collect_progress = ret_dict["message"]["PercentComplete"]
                    payload_publish["percentage"] = f"{info_collect_progress}%"
                    run_log.info("collect_done: %s, progress: %s", collect_done, payload_publish["percentage"])
                    resp_msg_obj = FdMsgData.gen_ws_msg_obj(payload_publish, "websocket/info_collect_process")
                    publish_ws_msg(resp_msg_obj)
                    threading.Timer(5, publish_info_collect_progress).start()

            def collect_system_log():
                if os.path.exists(LoggerConstants.FD_LOG_COLLECT_PATH):
                    os.remove(LoggerConstants.FD_LOG_COLLECT_PATH)

                if not FileCreate.create_dir(LoggerConstants.REDFISH_LOG_COLLECT_DIR, 0o1700):
                    run_log.error("Collect log failed because create collect_log_dir failed.")
                    return False

                request_data_dict = {"name": LOG_MODULES_MAP.get(payload.get("module"))}
                ret_dict = LibRESTfulAdapter.lib_restful_interface("LoggerCollect", "POST", request_data_dict, False)
                if not LibRESTfulAdapter.check_status_is_ok(ret_dict):
                    run_log.error("Collect log failed, om core operate failed")
                    return False

                # 将文件拷贝到MindXOM用户目录下，提供下载，并通知monitor进程删除临时日志文件
                shutil.copyfile(LoggerConstants.MONITOR_TMP_COLLECT_LOG, LoggerConstants.FD_LOG_COLLECT_PATH)
                LibRESTfulAdapter.lib_restful_interface("LoggerCollect", "DELETE", None)
                return True

            payload_publish = MidwareUris.get_log_collect_publish_template(payload.get("module", ""))
            try:
                collect_done = 0
                error_msg = "Info collect failed. %s"
                ret = MidwareUris.check_external_param(InfoCollectChecker, payload, payload_publish, error_msg)
                if ret[0] != 0:
                    return ret

                # 查询锁状态，如果锁已经占用，直接返回失败
                if CommonRedfish.SYS_CRITIC_LOCK.locked():
                    run_log.error("Log collection failed. System resource is busy.")
                    payload_publish["result"] = "failed"
                    payload_publish["reason"] = MidwareErrCode.midware_generate_err_msg(
                        MidwareErrCode.midware_resource_busy, "System is busy. Please try again later.")
                    return [-1, payload_publish]

                # 申请锁
                with CommonRedfish.SYS_CRITIC_LOCK:
                    run_log.info("Parameters check ok. Start to collect.")
                    # 定时查询日志收集进度
                    publish_info_collect_progress()
                    # 执行日志收集
                    ret = collect_system_log()
                    collect_done = 1
                    payload_publish["result"] = "failed"
                    if not ret:
                        run_log.error("Collect log failed.")
                        payload_publish["reason"] = MidwareErrCode.midware_generate_err_msg(
                            MidwareErrCode.midware_info_collect_exec_err,
                            "Failed to execute the collection script. Please try again later.")
                        return [-1, payload_publish]
                    ret = MidwareUris.https_upload(payload, LoggerConstants.FD_LOG_COLLECT_PATH)
                    if ret[0] != 0:
                        run_log.error("Log upload failed, %s", ret[1])
                        payload_publish["reason"] = MidwareErrCode.midware_generate_err_msg(
                            MidwareErrCode.midware_info_collect_upload_err,
                            "Log upload failed. Please check parameters or network.")
                        return [-1, payload_publish]

                    payload_publish["result"] = "success"
                    payload_publish["percentage"] = "100%"
                    return [0, payload_publish]
            except Exception:
                run_log.error("Collect log failed, caught exception.")
                payload_publish["result"] = "failed"
                payload_publish["reason"] = MidwareErrCode.midware_generate_err_msg(
                    MidwareErrCode.midware_info_collect_exec_err,
                    "Failed to execute the collection script. Please try again later.")
                return [-1, payload_publish]
            finally:
                FileUtils.delete_file_or_link(LoggerConstants.FD_LOG_COLLECT_PATH)

        # 固件生效入口
        @midWare.route(r'espmanager/UpdateService/FirmwareEffective')
        @FdCommonMethod.fd_operational_log("Firmware effective")
        def firmware_effective(payload):
            run_log.info("Firmware effective start.")

            payload_publish = {
                "restartable": "false",
                "reason": "resource is busy, can't effective firmware now, pls try later"
            }

            check_ret = MidwareUris.check_external_param(
                FirmwareEffectiveChecker, payload, payload_publish, "Firmware effective start failed. %s")
            if check_ret[0] != 0:
                run_log.error(f'check external param failed')
                return check_ret

            # 校验数据是否为json
            active_type = payload.get("active", None)
            if active_type is None:
                err_info = "Request data is not json."
                run_log.error(f'{err_info}')
                payload_publish["restartable"] = "false"
                payload_publish["reason"] = MidwareErrCode.midware_generate_err_msg(
                    MidwareErrCode.midware_input_not_json, err_info)
                return [-1, payload_publish]

            try:
                # 获取资源模板
                if RedfishGlobals.high_risk_exclusive_lock.locked():
                    payload_publish["reason"] = MidwareErrCode.midware_generate_err_msg(
                        MidwareErrCode.midware_resource_busy,
                        "Resource is busy, can't effective firmware now, please try later.")
                    return [-1, payload_publish]

                with RedfishGlobals.high_risk_exclusive_lock:
                    recv_conf_dict = {}

                    if active_type == "inband":
                        recv_conf_dict["ResetType"] = "GracefulRestart"
                        payload_publish["restartable"] = "true"
                        payload_publish["reason"] = "effective firmware now"
                        resp_msg_obj = FdMsgData.gen_ws_msg_obj(payload_publish, "websocket/restart_result")
                        publish_ws_msg(resp_msg_obj)
                        time.sleep(5)

                        # 注意：此接口内部用定时器延迟重启，所以总能收到返回值
                        run_log.info("System to restart gracefully to effective firmware!")
                        ret_dict = LibRESTfulAdapter.lib_restful_interface("Upgrade_effect", "POST", recv_conf_dict)
                        run_log.info(f'{ret_dict}')
                        ret = CommonRedfish.check_status_is_ok(ret_dict)
                        if ret:
                            MidwareUris.set_operation_log("Restart system (GracefulRestart) successfully.")
                            return [0, payload_publish]

                        run_log.error("Force system restart to effective firmware failed.")
                        payload_publish["restartable"] = "false"
                        payload_publish["reason"] = MidwareErrCode.midware_generate_err_msg(
                            MidwareErrCode.midware_common_err, "effect failed")
                        return [-1, payload_publish]

                    payload_publish["restartable"] = "false"
                    payload_publish["reason"] = MidwareErrCode.midware_generate_err_msg(
                        MidwareErrCode.midware_input_parameter_invalid,
                        "Only suport inband mode, can't effective firmware")
                    return [-1, payload_publish]

            except Exception as err:
                run_log.error("exception caught in firmware_effective: %s", err)
                payload_publish["restartable"] = "false"
                payload_publish["reason"] = MidwareErrCode.midware_generate_err_msg(
                    MidwareErrCode.midware_common_err, str(err))
                return [-1, payload_publish]

        # 配置导入入口
        @midWare.route(r'espmanager/SysConfig')
        @FdCommonMethod.fd_operational_log("Import System config")
        def profile(payload):
            def check_product_name(name):
                with MidwareUris.sys_lock:
                    product_name = MidwareUris.resp_json_sys_info["product_info"]["product_name"]

                return name == product_name

            def check_profile_name(name):
                # 防止出现非法字符
                if "/" in name:
                    return [-1, "Invalid character is not permmited."]

                # 文件长度范围为1~32字节
                if not len(name) in range(1, 33):
                    return [-1, "Input profile_name length should be between 1 and 32 bytes."]

                return [0, "OK"]

            with MidwareUris.profile_lock:
                try:
                    check_ret = SysConfigChecker().check(payload)
                    err_info = "System config check failed"
                    if any(info in check_ret.reason for info in ("apn_user", "apn_passwd", "apn_info")):
                        err_info = "apn_info config check failed"
                    if not check_ret.success:
                        run_log.error("Import System config failed. %s", check_ret.reason)
                        return [-1, f"FAILED: {err_info}"]

                    product_info = payload.get("product", None)
                    profile_name = payload.get("profile_name", None)
                    config_info = payload.get("config", None)
                    if product_info is None or profile_name is None or config_info is None:
                        return [-1, "FAILED: Request data is not json."]

                    if not check_product_name(product_info):
                        err_info = "Product info mismatch."
                        run_log.error("%s", err_info)
                        return [-1, f"FAILED: {err_info}"]

                    ret = check_profile_name(profile_name)
                    if ret[0] != 0:
                        err_info = ret[1]
                        run_log.error("%s", err_info)
                        return [-1, f"FAILED: {err_info}"]

                    ret = MidwareUris.check_capacity_bytes_valid(config_info.get("partitions"))
                    if not ret:
                        run_log.error("The partition capacity bytes must be an integer multiple of 512MB")
                        return [-1, "FAILED: the partition capacity bytes must be an integer multiple of 512MB"]

                    config_path = ConfigPathConstants.SYS_CONFIG_PATH
                    # 生成新的*.prf文件之前先删除/home/data/config/redfish/*.prf文件
                    for index, prf_file in enumerate(glob.iglob(os.path.join(config_path, "*.prf"))):
                        if index > MidwareUris.max_cycle_num:
                            run_log.error("Check more than %s prf files.", MidwareUris.max_cycle_num)
                            return [-1, "FAILED: System config failed"]

                        os.remove(prf_file)

                    profile_name = os.path.join(config_path, (profile_name + ".prf"))
                    file_info = {"product": product_info, "profile_name": profile_name, "config": config_info}

                    res = FileCheck.check_is_link(profile_name)
                    if not res:
                        raise ValueError(res.error)
                    # 对整个文件内容加密，使用的时候再解密
                    file_content = kmc.encrypt(json.dumps(file_info))
                    with os.fdopen(os.open(profile_name, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600), "w+") as file:
                        file.write(file_content)
                        return [0, "SUCCESS"]
                except Exception as err:
                    run_log.error("caught an exception in import profile: %s", err)
                    return [-1, "FAILED: caught an exception in import profile"]

        # 配置资产标签入口
        @midWare.route(r'espmanager/SysAssetTag')
        @FdCommonMethod.fd_operational_log("Set electronic label")
        def tag(payload):

            payload_publish = {
                "topic": "tag",
                "percentage": "0%",
                "result": "failed",
                "reason": "",
            }
            try:
                param_check_error = "Set system tag failed. %s"
                check_ret = MidwareUris.check_external_param(
                    SysAssetTagChecker, payload, payload_publish, param_check_error)
                if check_ret[0] != 0:
                    return check_ret

                request_dict = {}
                request_dict["AssetTag"] = payload.get("asset_tag", None)
                request_dict["_User"] = "FD"
                request_dict["_Xip"] = FdCfgManager.get_cur_fd_ip()
                if not request_dict.get("AssetTag") or len(request_dict.get("AssetTag")) > 255:
                    run_log.error("Modify edge system tag parametr check failed.")
                    payload_publish["reason"] = MidwareErrCode.midware_generate_err_msg(
                        MidwareErrCode.midware_input_parameter_invalid, "Input parameter is invalid.")
                    return [-1, payload_publish]

                ret_dict = LibRESTfulAdapter.lib_restful_interface("System", "PATCH", request_dict, False)
                ret = CommonRedfish.check_status_is_ok(ret_dict)
                if ret:
                    run_log.info("Modify edge system tag successfully.")
                    payload_publish["percentage"] = "100%"
                    payload_publish["result"] = "success"
                    return [0, payload_publish]
                run_log.error("Modify edge system tag failed.")
                payload_publish["reason"] = MidwareErrCode.midware_generate_err_msg(
                    MidwareErrCode.midware_assert_tag_common_err, "Modify edge system tag failed.")
                return [-1, payload_publish]

            except Exception as err:
                run_log.error(str(err))
                payload_publish["result"] = "failed"
                payload_publish["reason"] = MidwareErrCode.midware_generate_err_msg(
                    MidwareErrCode.midware_assert_tag_common_err, str(err))
                return [-1, payload_publish]

        # 配置系统主机名入口
        @midWare.route(r'espmanager/Hostname')
        @FdCommonMethod.fd_operational_log("Set hostname")
        def config_hostname(payload):
            payload_publish = {
                "topic": "config_hostname",
                "percentage": "0%",
                "result": "failed",
                "reason": ""
            }
            try:
                if MidwareUris.HOST_NAME_LOCK.locked():
                    run_log.warning("Set hostname is busy")
                    payload_publish["reason"] = f"ERR.{MidwareErrCode.midware_resource_busy}, Set hostname is busy."
                    return [-1, payload_publish]
                with MidwareUris.HOST_NAME_LOCK:
                    param_check_error = "Invalid hostname. %s"
                    check_ret = MidwareUris.check_external_param(
                        HostnameChecker, payload, payload_publish, param_check_error)
                    if check_ret[0] != 0:
                        return check_ret

                    hostname = payload.get("hostname", None)
                    if hostname is None or not AppCommonMethod.hostname_check(hostname):
                        run_log.error("Invalid hostname.")
                        payload_publish["reason"] = f"ERR.{MidwareErrCode.midware_input_parameter_invalid}" \
                                                    f", Invalid hostname."
                        payload_publish["result"] = "failed"
                        return [-1, payload_publish]

                    ret = CommonRedfish.modify_host_name(hostname)
                    if not ret:
                        run_log.error("Config hostname failed: %", ret.error)
                        payload_publish["result"] = "failed"
                        payload_publish["reason"] = f"ERR.{MidwareErrCode.midware_common_err}, Config hostname failed"
                        return [-1, payload_publish]

                    payload_publish["result"] = "success"
                    payload_publish["percentage"] = "100%"
                    run_log.info("Config hostname success: hostname is %s" % (hostname))
                    return [0, payload_publish]

            except Exception as err:
                run_log.error("config hostname exception: %s", err)
                payload_publish["result"] = "failed"
                payload_publish["reason"] = f"ERR.{MidwareErrCode.midware_common_err}, Config hostname failed"
                return [-1, payload_publish]

        # 配置生效
        @midWare.route(r'espmanager/SysConfigEffect')
        @FdCommonMethod.fd_operational_log("Config effective")
        def profile_effect(payload):
            payload_lock = threading.Lock()
            cfg_to_effective = {}
            percentage = 0
            fail_reason = ""

            payload_publish = {
                "profile_name": "NA",
                "percentage": "0%",
                "result": "in progress",
                "reason": AppCommonMethod.COM_ERR_REASON_TO_FD
            }

            route = {}
            route["operation"] = "post"
            route["resource"] = "espmanager/SysConfigProgress"
            error_mesg = ""

            def publish_sys_config_effect_progress():
                while True:
                    payload_report = {}
                    with payload_lock:
                        payload_publish["percentage"] = "%s" % percentage + "%"
                        payload_report = copy.deepcopy(payload_publish)
                    run_log.info("SysConfigEffect progress: %s" % (payload_report.get("percentage")))
                    resp_msg_obj = FdMsgData.gen_ws_msg_obj(payload_report, "websocket/profile_effect")
                    publish_ws_msg(resp_msg_obj)

                    if payload_report.get("result") != "in progress":
                        break
                    time.sleep(5)

            def profile_effect_lte(lte_info_payload):
                def set_status_info(lte_info_payload):
                    try:
                        if lte_info_payload and \
                                "state_lte" not in lte_info_payload[0] and "state_data" not in lte_info_payload[0]:
                            run_log.info("Effect wireless status skip: wireless info is empty, no need to config.")
                            return True
                    except Exception:
                        return False

                    error_message = "Effect wireless status info failed."
                    try:
                        req_dict = {}
                        if lte_info_payload:
                            req_dict["state_lte"] = lte_info_payload[0]["state_lte"]
                            req_dict["state_data"] = lte_info_payload[0]["state_data"]

                        req_dict["fd_server_ip"] = FdCfgManager.get_cur_fd_ip()
                        ret_dict = LibRESTfulAdapter.lib_restful_interface("LteInfo", "PATCH", req_dict, False)
                        if CommonRedfish.check_status_is_ok(ret_dict):
                            error_message = "Effect wireless status info successfully."
                            return True
                        return False
                    except Exception as err:
                        run_log.error("Effect wireless status info error: %s", err)
                        return False
                    finally:
                        MidwareUris.set_operation_log(error_message)

                def set_config_info(lte_info_payload):
                    try:
                        if lte_info_payload \
                                and "apn_info" in lte_info_payload[0] and not lte_info_payload[0]["apn_info"]:
                            run_log.info("Effect LTE config info skip: apn_info is empty, no need to config.")
                            return True
                    except Exception:
                        run_log.error("Check lte_info_payload failed.")
                        return False

                    error_message = "Effect APN config info failed."
                    try:
                        req_dict = {}
                        if lte_info_payload and "apn_info" in lte_info_payload[0]:
                            req_dict["apn_name"] = lte_info_payload[0]["apn_info"][0]["apn_name"]
                            req_dict["apn_user"] = lte_info_payload[0]["apn_info"][0]["apn_user"]
                            req_dict["apn_passwd"] = lte_info_payload[0]["apn_info"][0]["apn_passwd"]
                            req_dict["auth_type"] = lte_info_payload[0]["apn_info"][0]["auth_type"]

                        ret_dict = LibRESTfulAdapter.lib_restful_interface("LteConfigInfo", "PATCH", req_dict, False)
                        if CommonRedfish.check_status_is_ok(ret_dict):
                            error_message = "Effect APN config info successfully."
                            return True
                        return False
                    except Exception:
                        return False
                    finally:
                        MidwareUris.set_operation_log(error_message)

                input_opera_log = "Effect wireless failed."
                try:
                    if not set_config_info(lte_info_payload):
                        return False

                    if not set_status_info(lte_info_payload):
                        return False

                    input_opera_log = "Effect wireless successfully."
                    return True
                except Exception as err:
                    run_log.error("Caught an exception in profile_effect_lte")
                    return False
                finally:
                    MidwareUris.set_operation_log(input_opera_log)

            def profile_effect_ntp(payload):
                input_opera_log = "Modify NTP service failed."
                try:
                    run_log.info(payload["config"]["ntp_server"])
                    ntp_cfg = {}
                    ntp_cfg["Target"] = "Client"
                    ntp_cfg["NTPLocalServers"] = ""
                    ntp_cfg["ClientEnabled"] = payload["config"]["ntp_server"]["service_enabled"]

                    if "sync_net_manager" in payload["config"]["ntp_server"] \
                            and payload["config"]["ntp_server"]["sync_net_manager"]:

                        ip_info = FdCfgManager.get_cur_fd_ip()
                        if ip_info:
                            ntp_cfg["NTPRemoteServers"] = ip_info

                        res = FileCheck.check_is_link(FlagPathConstant.NTP_SYNC_NET_MANAGER)
                        if not res:
                            raise ValueError(res.error)
                        with os.fdopen(os.open(FlagPathConstant.NTP_SYNC_NET_MANAGER,
                                               os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600), "w+"):
                            pass

                    else:
                        ntp_cfg["NTPRemoteServers"] = payload["config"]["ntp_server"]["preferred_server"]
                        ntp_cfg["NTPRemoteServersbak"] = payload["config"]["ntp_server"]["alternate_server"]
                        if os.path.exists(FlagPathConstant.NTP_SYNC_NET_MANAGER):
                            os.remove(FlagPathConstant.NTP_SYNC_NET_MANAGER)

                    ret_dict = LibRESTfulAdapter.lib_restful_interface("NTPService", "PATCH", ntp_cfg, False)
                    ret = CommonRedfish.check_status_is_ok(ret_dict)
                    if ret:
                        input_opera_log = "Modify NTP service successfully."
                        run_log.info(input_opera_log)
                        return True

                    run_log.error("Modify NTP service failed.")
                    return False
                except Exception as err:
                    run_log.error("caught an exception in profile_effect_ntp: %s", err)
                    return False
                finally:
                    MidwareUris.set_operation_log(input_opera_log)

            def profile_effect_partion(payload):
                nonlocal error_mesg
                a500_emmc = "/dev/mmcblk0"
                emmc_p8 = "mmcblk0p8"

                def get_dev_name(dev_location):
                    nonlocal error_mesg

                    ret = CommonRedfish.get_dev_site_name_by_location(dev_location)
                    if not ret:
                        if not error_mesg:
                            error_mesg = "Cannot find any devices from loacation:%s" % (dev_location)
                            error_mesg = MidwareErrCode.midware_generate_err_msg(
                                MidwareErrCode.midware_partition_dev_not_fount, error_mesg)
                        return False

                    return ret.data

                def del_partion(dev_name):
                    nonlocal error_mesg
                    try:
                        if not dev_name:
                            if not error_mesg:
                                error_mesg = MidwareErrCode.midware_generate_err_msg(
                                    MidwareErrCode.midware_common_err, "Internel error")
                            run_log.error("Internel error.")
                            return False

                        request_data_dict = {"_User": "FD", "_Xip": FdCfgManager.get_cur_fd_ip()}

                        # 获取partion列表
                        ret_dict = LibRESTfulAdapter.lib_restful_interface("Partition", "GET", None, True)
                        run_log.info("Partion list is %s", ret_dict)
                        ret = CommonRedfish.check_status_is_ok(ret_dict)
                        if not ret:
                            if not error_mesg:
                                error_mesg = MidwareErrCode.midware_generate_err_msg(
                                    MidwareErrCode.midware_partition_list_failed, "Get partition list fail")
                            run_log.error("Get partition list failed.")
                            return False

                        for partion in ret_dict["message"]:
                            time.sleep(2)
                            # 目标不匹配，略过
                            if dev_name.split("/")[-1] not in partion:
                                continue

                            # A500 emmc不支持删除分区
                            if partion.startswith(a500_emmc.split("/")[-1]):
                                continue

                            run_log.info("Start to delete partion.")
                            ret_dict = LibRESTfulAdapter.lib_restful_interface(
                                "Partition", "DELETE", request_data_dict, False, partion)

                            # 删除成功，继续删除下一个partition
                            ret = CommonRedfish.check_status_is_ok(ret_dict)
                            if ret:
                                message = "Delete partition [%s] successfully" % partion
                                run_log.info(message)
                                continue

                            # 删除失败，error_mesg不为空时，直接返回失败
                            if error_mesg:
                                run_log.error("Delete partition [%s] fail. Please make sure the partition "
                                              "is not using by program. %s", partion, ret_dict["message"])
                                return False

                            # 删除失败，error_mesg为空时，给error_mesg正确赋值，再返回失败
                            error_mesg = f"Delete partition [{partion}] fail. Please make sure the partition is not " \
                                         f"using by program or not damaged. {ret_dict['message']}"

                            if "apparently in use by the system" in ret_dict["message"]:
                                error_mesg = MidwareErrCode.midware_generate_err_msg(
                                    MidwareErrCode.midware_partition_damaged, error_mesg)
                                run_log.error(f'{error_mesg}')
                            else:
                                error_mesg = MidwareErrCode.midware_generate_err_msg(
                                    MidwareErrCode.midware_partition_is_using, error_mesg)
                            return False

                        # 全部成功删除，返回成功
                        return True
                    except Exception as err:
                        run_log.error("del partion error: %s", err)
                        return False

                def create_partion(dev_name):
                    nonlocal error_mesg
                    try:
                        if not dev_name:
                            if not error_mesg:
                                error_mesg = MidwareErrCode.midware_generate_err_msg(
                                    MidwareErrCode.midware_common_err, "Internel error")
                            return ""

                        # A500 emmc不支持创建分区
                        if dev_name.startswith(a500_emmc):
                            return emmc_p8

                        time.sleep(2)
                        run_log.info("create partition: %s", dev_name)
                        capacity = str(float(partion["capacity_bytes"]) / 1024.0 / 1024.0 / 1024.0)
                        request_dict = {
                            "Number": 1,
                            "Links": [{
                                "Device": {
                                    "@odata.id": dev_name
                                }
                            }],
                            "CapacityBytes": capacity,
                            "FileSystem": "ext4"
                        }
                        request_dict["_User"] = "FD"
                        request_dict["_Xip"] = FdCfgManager.get_cur_fd_ip()
                        ret_dict = LibRESTfulAdapter.lib_restful_interface("Partition", "POST", request_dict, False)
                        ret = CommonRedfish.check_status_is_ok(ret_dict)
                        if ret:
                            return ret_dict["message"]["DeviceName"]

                        if not error_mesg:
                            error_mesg = "Create partition [%s] failed. %s" % (dev_name, ret_dict["message"])
                            run_log.error(f'{error_mesg}')
                            if "out of range" in ret_dict["message"] or "Insufficient disk" in ret_dict["message"]:
                                error_mesg = MidwareErrCode.midware_generate_err_msg(
                                    MidwareErrCode.midware_partition_out_range, error_mesg)
                            elif "apparently in use by the system" in ret_dict["message"]:
                                error_mesg = MidwareErrCode.midware_generate_err_msg(
                                    MidwareErrCode.midware_partition_damaged, error_mesg)
                            else:
                                error_mesg = MidwareErrCode.midware_generate_err_msg(
                                    MidwareErrCode.midware_common_err, error_mesg)
                            run_log.error(f'{error_mesg}')

                        run_log.error(f'{ret_dict["message"]}')
                        return ""
                    except Exception as err:
                        run_log.error("create partion error: %s", err)
                        return ""

                def mount_partion(partion, path):
                    nonlocal error_mesg
                    if not partion or not path:
                        if not error_mesg:
                            error_mesg = "Mount parameter error"
                            error_mesg = MidwareErrCode.midware_generate_err_msg(
                                MidwareErrCode.midware_common_err, error_mesg)
                        run_log.error("Mount parameter error.")
                        return False

                    time.sleep(2)
                    try:
                        request_dict = {"_User": "FD", "_Xip": FdCfgManager.get_cur_fd_ip()}
                        # emmc p8分区如果挂载了，先解挂
                        if partion == emmc_p8:
                            run_log.info("Start to unmount partiton")
                            ret_dict = LibRESTfulAdapter.lib_restful_interface(
                                "Partition", "GET", None, False, partion)

                            if ret_dict.get("message", {}).get("MountPath"):
                                ret_dict = LibRESTfulAdapter.lib_restful_interface(
                                    "Partition", "PATCH", request_dict, False, partion)
                                if not CommonRedfish.check_status_is_ok(ret_dict):
                                    error_mesg = f"Umount partition {partion} failed"
                                    run_log.info("Umount partition %s failed", partion)
                                    return False

                        request_dict["MountPath"] = path
                        ret_dict = LibRESTfulAdapter.lib_restful_interface(
                            "Partition", "PATCH", request_dict, False, partion)
                        ret = CommonRedfish.check_status_is_ok(ret_dict)
                        run_log.info(f'{ret_dict}')
                        if ret:
                            message = "Mount partition [%s] successfully" % partion
                            run_log.info(message)
                            return True

                        if not error_mesg:
                            error_mesg = "Mount partition [%s] to path [%s] failed. %s" % \
                                         (partion, path, ret_dict['message'])
                            error_mesg = MidwareErrCode.midware_generate_err_msg(
                                MidwareErrCode.midware_partition_path_not_empty, error_mesg)
                        run_log.error(ret_dict['message'])
                        return False
                    except Exception as err:
                        if not error_mesg:
                            error_mesg = str(err)
                            error_mesg = MidwareErrCode.midware_generate_err_msg(
                                MidwareErrCode.midware_common_err, error_mesg)
                        run_log.error("%s", err)
                        return False

                def get_partion_num(location_info, partitions):
                    num = 0
                    for partion in partitions:
                        if location_info == partion["device"]["device_location"]:
                            num = num + 1
                    return num

                input_opera_log = "Effect partition config failed."
                try:
                    run_log.info(f'{payload["config"]["partitions"]}')
                    partitions = payload["config"]["partitions"]
                    mount_paths = []
                    # 删除所有分区
                    locations = []

                    for partion in partitions:
                        if not isinstance(partion["capacity_bytes"], int):
                            run_log.info("capacity_bytes is not int type")
                            error_mesg = "capacity_bytes is not int type"
                            error_mesg = MidwareErrCode.midware_generate_err_msg(
                                MidwareErrCode.midware_input_parameter_invalid, error_mesg)
                            return False

                        if float(partion["capacity_bytes"]) > 0:
                            file_system = partion.get("file_system", None)
                            if file_system and file_system != "ext4":
                                run_log.info("File_system %s is not support." % (file_system))
                                error_mesg = "The File_system %s is not support." % (file_system)
                                error_mesg = MidwareErrCode.midware_generate_err_msg(
                                    MidwareErrCode.midware_partition_fs_not_support, error_mesg)
                                return False
                        locations.append(partion["device"]["device_location"])

                    locations = list(set(locations))
                    for location in locations:
                        dev_name = get_dev_name(location)
                        if not del_partion(dev_name):
                            input_opera_log = "Delete partition failed."
                            return False

                    time.sleep(2)
                    # 重新创建并挂载分区
                    for partion in partitions:
                        capacity = float(partion["capacity_bytes"])
                        if capacity <= 0:
                            continue

                        dev_name = get_dev_name(partion["device"]["device_location"])
                        pation_num = get_partion_num(partion["device"]["device_location"], partitions)
                        if pation_num > 16:
                            error_mesg = "The number of partions is more than 16"
                            error_mesg = MidwareErrCode.midware_generate_err_msg(
                                MidwareErrCode.midware_partition_out_range, error_mesg)
                            return False

                        partion_name = create_partion(dev_name)
                        run_log.info(f"{partion_name}")
                        if not partion_name:
                            input_opera_log = "Create partition failed."
                            return False

                        time.sleep(2)
                        run_log.info(f'{partion["mount_path"]}')
                        mount_paths.append(partion["mount_path"])
                        if partion["mount_path"] and not mount_partion(partion_name, partion["mount_path"]):
                            input_opera_log = "Mount partition failed."
                            return False

                    input_opera_log = "Effect partition config successfully."
                    return True

                except Exception as err:
                    run_log.error("caught an exception in profile_effect_partition: %s", err)
                    return False
                finally:
                    MidwareUris.set_operation_log(input_opera_log)

            def profile_effect_domain(payload):
                message = "Effect domain configuration failed."
                try:
                    req_dict = {"net_manager_domain": FdCfgManager.get_cur_fd_server_name()}
                    if "static_host_list" in payload["config"]:
                        req_dict["static_host_list"] = payload["config"]["static_host_list"]
                        domain_names = (item.get("name") for item in req_dict["static_host_list"])
                        if FdCommonMethod.contains_forbidden_domain(domain_names):
                            run_log.error("DomainCfg failed: invalid domain format.")
                            return False

                    if "name_server" in payload["config"]:
                        req_dict["name_server"] = payload["config"]["name_server"]

                    ret_dict = LibRESTfulAdapter.lib_restful_interface("DomainCfg", "POST", req_dict, False)
                    ret = CommonRedfish.check_status_is_ok(ret_dict)
                    if ret:
                        message = "Effect domain configuration successfully."
                        return True

                    run_log.error("DomainCfg failed: %s", ret_dict.get("message"))
                    return False
                except Exception as err:
                    run_log.error("Effect domain configuration error: %s", err)
                    return False
                finally:
                    MidwareUris.set_operation_log(f'{message}')

            def profile_effect_security_policy(security_policy_payload):
                return _effect_passwd_validity(security_policy_payload) \
                       and _effect_access_control(security_policy_payload) \
                       and _effect_session_timeout(security_policy_payload) \
                       and _effect_cert_alarm_time(security_policy_payload) \
                       and _effect_security_load_rule(security_policy_payload)

            def wait_for_profile_a_minute(profile_name_path):
                for _ in range(30):
                    if os.path.exists(profile_name_path):
                        break
                    run_log.info("Profile file {} is not exist".format(profile_name_path))
                    time.sleep(2)

            def get_cfg_to_effective(profile_name_path):
                with MidwareUris.profile_lock:
                    res = FileCheck.check_path_is_exist_and_valid(profile_name_path)
                    if not res:
                        run_log.error(f"{profile_name_path} path invalid : {res.error}")
                    else:
                        with open(profile_name_path, 'r') as file:
                            config = json.loads(kmc.decrypt(file.read()))
                    os.remove(profile_name_path)
                    return config

            def _effect_passwd_validity(security_policy_payload):
                if "password_validity" not in security_policy_payload:
                    return True

                message = "Modify account info failed."
                try:
                    req_dict = {
                        "PasswordExpirationDays": security_policy_payload["password_validity"],
                        "oper_type": UserManagerConstants.OPER_TYPE_FD_MODIFY_ACCOUNT_EXPIRATION_DAY
                    }
                    ret_dict = UserManager().patch_request(req_dict)
                    if CommonRedfish.check_status_is_ok(ret_dict):
                        # 立即更新缓存信息
                        MidwareUris.get_passwd_validity_info()
                        FdCommonMethod.sec_cfg_check_done = True
                        message = "Modify account info successfully."
                        return True

                    run_log.error("passwd validity config failed: {}".format(ret_dict.get("message")))
                    return False
                except Exception as err:
                    input_error = "Modify account info validity error: {}".format(err)
                    run_log.error(input_error)
                    return False
                finally:
                    MidwareUris.set_operation_log(message)

            def _effect_session_timeout(security_policy_payload):
                if "session_timeout" not in security_policy_payload:
                    return True

                message = "Modify session timeout config failed."
                try:
                    req_dict = {
                        "SessionTimeout": security_policy_payload["session_timeout"],
                        "oper_type": UserManagerConstants.OPER_TYPE_FD_MODIFY_SESSION_TIMEOUT
                    }
                    ret_dict = SessionManager().patch_request(req_dict)
                    if CommonRedfish.check_status_is_ok(ret_dict):
                        # 立即更新缓存信息
                        MidwareUris.get_session_timeout()
                        FdCommonMethod.sec_cfg_check_done = True
                        message = "Modify session timeout config successfully."
                        return True

                    run_log.error("session timeout config failed: %s", ret_dict.get("message"))
                    return False
                except Exception as err:
                    run_log.error("Modify cert alarm time config validity error: %s", err)
                    return False
                finally:
                    MidwareUris.set_operation_log(message)

            def _effect_cert_alarm_time(security_policy_payload):
                if "cert_alarm_time" not in security_policy_payload:
                    return True

                message = "Modify cert alarm time failed."
                try:
                    req_dict = {"CertAlarmTime": security_policy_payload["cert_alarm_time"]}
                    ret_dict = LibRESTfulAdapter.lib_restful_interface("CertAlarmTime", "PATCH", req_dict, False)
                    if LibRESTfulAdapter.check_status_is_ok(ret_dict):
                        # 立即更新缓存信息
                        MidwareUris.get_cert_alarm_time()
                        FdCommonMethod.sec_cfg_check_done = True
                        message = "Modify cert alarm time successfully."
                        # 更新证书告警时间之后执行一次证书状态检查
                        FdCfgManager().check_cert_status()
                        return True

                    run_log.error("cert alarm time config failed: %s", ret_dict.get("message"))
                    return False
                except Exception as err:
                    run_log.error("Modify cert alarm time config validity error: %s", err)
                    return False
                finally:
                    MidwareUris.set_operation_log(message)

            def _effect_security_load_rule(security_policy_payload):
                if "security_load" not in security_policy_payload:
                    return True

                message = "Modify security load rule failed."
                try:
                    req_dict = {"load_cfg": security_policy_payload["security_load"]}
                    ret_dict = LibRESTfulAdapter.lib_restful_interface("SecurityLoad", "PATCH", req_dict, False)
                    if LibRESTfulAdapter.check_status_is_ok(ret_dict):
                        # 立即更新缓存信息
                        MidwareUris.get_security_load_config()
                        FdCommonMethod.sec_cfg_check_done = True
                        message = "Modify security load rule successfully."
                        return True

                    run_log.error("security load rule config failed: %s", ret_dict.get("message"))
                    return False
                except Exception as err:
                    run_log.error("Modify security load rule validity error: %s", err)
                    return False
                finally:
                    MidwareUris.set_operation_log(message)

            def _effect_access_control(security_policy_payload):
                if "web_access" not in security_policy_payload and "ssh_access" not in security_policy_payload:
                    return True

                req_dict = {
                    "_User": "FD",
                    "_Xip": FdCfgManager.get_cur_fd_ip()
                }
                if "web_access" in security_policy_payload:
                    req_dict["web_access"] = security_policy_payload["web_access"]

                if "ssh_access" in security_policy_payload:
                    req_dict["ssh_access"] = security_policy_payload["ssh_access"]

                msg = "Effect access control failed"
                try:
                    ret_dict = LibRESTfulAdapter.lib_restful_interface("AccessControl", "PATCH", req_dict, False)
                    if LibRESTfulAdapter.check_status_is_ok(ret_dict):
                        msg = "Effect access control successfully"
                        run_log.info("Config access control successfully.")
                        return True

                    run_log.error("Config access control failed.")
                    return False
                except Exception:
                    run_log.error("Config access control has exception.")
                    return False
                finally:
                    MidwareUris.set_operation_log(msg)

            if CommonRedfish.SYS_CRITIC_LOCK.locked():
                err_info = "Resouce is busy; Can't do profile effective Now; Please try later."
                payload_publish["result"] = "failed"
                payload_publish["reason"] = MidwareErrCode.midware_generate_err_msg(
                    MidwareErrCode.midware_resource_busy, err_info)
                return [-1, payload_publish]

            with CommonRedfish.SYS_CRITIC_LOCK:
                try:
                    param_check_error = "Invalid profile_name. %s"
                    check_ret = MidwareUris.check_external_param(
                        SysConfigEffectChecker, payload, payload_publish, param_check_error)
                    if check_ret[0] != 0:
                        return check_ret

                    profile_name = payload.get("profile_name", None)
                    if profile_name is None:
                        err_info = "Profile_name is none."
                        run_log.error(f'{err_info}')
                        payload_publish["result"] = "failed"
                        payload_publish["reason"] = MidwareErrCode.midware_generate_err_msg(
                            MidwareErrCode.midware_input_parameter_invalid, err_info)
                        return [-1, payload_publish]

                    if not len(profile_name) in range(1, 33):
                        err_info = "Input profile_name length should be between 1 and 32 bytes."
                        run_log.error(f'{err_info}')
                        payload_publish["result"] = "failed"
                        payload_publish["reason"] = MidwareErrCode.midware_generate_err_msg(
                            MidwareErrCode.midware_input_parameter_invalid, err_info)
                        return [-1, payload_publish]

                    profile_name_path = os.path.join(ConfigPathConstants.SYS_CONFIG_PATH, (profile_name + ".prf"))
                    payload_publish["profile_name"] = profile_name_path

                    wait_for_profile_a_minute(profile_name_path)

                    if not os.path.exists(profile_name_path):
                        run_log.error("Profile file {} is not exist".format(profile_name_path))
                        payload_publish["result"] = "failed"
                        payload_publish["reason"] = MidwareErrCode.midware_generate_err_msg(
                            MidwareErrCode.midware_profile_not_exist, "Profile file is not exist.")
                        return [-1, payload_publish]

                    cfg_to_effective = get_cfg_to_effective(profile_name_path)

                    threading.Thread(target=publish_sys_config_effect_progress).start()
                    # lte
                    if "lte_info" in cfg_to_effective["config"]:
                        run_log.info("Start to config lte.")
                        lte_ret = profile_effect_lte(cfg_to_effective["config"]["lte_info"])
                        if lte_ret:
                            percentage += 15
                        else:
                            fail_reason = fail_reason + MidwareErrCode.midware_generate_err_msg(
                                MidwareErrCode.midware_config_ntp_common_err, "LTE failed.")
                    else:
                        percentage += 15

                    # ntp
                    if "ntp_server" in cfg_to_effective["config"]:
                        run_log.info("Start to config ntp.")
                        ntp_ret = profile_effect_ntp(cfg_to_effective)
                        if ntp_ret:
                            percentage += 45
                        else:
                            fail_reason = fail_reason + MidwareErrCode.midware_generate_err_msg(
                                MidwareErrCode.midware_config_ntp_common_err, "NTP failed.")
                    else:
                        percentage += 45

                    # partion
                    if "partitions" in cfg_to_effective["config"] and cfg_to_effective["config"]["partitions"]:
                        partition_ret = profile_effect_partion(cfg_to_effective)
                        if partition_ret:
                            percentage += 15
                        else:
                            fail_reason = fail_reason + error_mesg
                    else:
                        percentage += 15

                    # dns
                    if "static_host_list" in cfg_to_effective["config"] or "name_server" in cfg_to_effective["config"]:
                        ret = profile_effect_domain(cfg_to_effective)
                        if ret:
                            percentage += 15
                        else:
                            fail_reason = fail_reason + MidwareErrCode.midware_generate_err_msg(
                                MidwareErrCode.midware_config_dns_common_err, "Dns/Domian config failed.")
                    else:
                        percentage += 15

                    # security_policy
                    if "security_policy" in cfg_to_effective["config"]:
                        ret = profile_effect_security_policy(cfg_to_effective["config"]["security_policy"])
                        if ret:
                            percentage += 10
                        else:
                            fail_reason = fail_reason + MidwareErrCode.midware_generate_err_msg(
                                MidwareErrCode.midware_config_passwd_validity_err,
                                "security policy config failed.")
                    else:
                        percentage += 10

                    payload_lock.acquire()
                    if percentage == 0:
                        payload_publish["result"] = "failed"
                        ret = -1
                    elif percentage == 100:
                        payload_publish["result"] = "success"
                        ret = 0
                    else:
                        payload_publish["result"] = "partial failed"
                        ret = -1

                    payload_publish["reason"] = fail_reason
                    payload_lock.release()

                    payload_publish["percentage"] = "%s" % (percentage) + "%"
                    return [ret, payload_publish]

                except Exception as err:
                    run_log.error("caught an exception in effective profile : %s", err)
                    payload_publish["result"] = "failed"
                    payload_publish["reason"] = MidwareErrCode.midware_generate_err_msg(
                        MidwareErrCode.midware_common_err, str(err))
                    return [-1, payload_publish]

        # 修改用户密码
        @midWare.route(r'espmanager/passthrough/account_modify')
        @FdCommonMethod.fd_operational_log("Modify account")
        def account_passwd_modify(payload):
            opeartion_message = "Modify account info failed."
            payload_publish = {
                "topic": "passthrough/account_modify",
                "percentage": "0%",
                "result": "failed",
                "reason": ""
            }
            check_ret = MidwareUris.check_external_param(
                UserInfoChecker, payload, payload_publish, "Parameter is invalid: %s.")
            if check_ret[0] != 0:
                run_log.error("%s %s", opeartion_message, "Parameter is invalid.")
                return check_ret

            user_name = payload.get("account")
            password = payload.get("new_password")
            recv_conf_dict = {
                'oper_type': UserManagerConstants.OPER_TYPE_MODIFY_PASSWORD,
                'UserName': user_name,
                'Password': password
            }
            try:
                ret_dict = UserManager().patch_request(recv_conf_dict)
            except Exception as err:
                run_log.error("account password modify exception: %s", err)
                payload_publish["reason"] = f"ERR.{MidwareErrCode.midware_common_err}, Config account password failed"
                payload_publish["result"] = "failed"
                return [-1, payload_publish]

            ret_status_is_ok = CommonRedfish.check_status_is_ok(ret_dict)
            if ret_status_is_ok:
                payload_publish["result"] = "success"
                payload_publish["percentage"] = "100%"
                run_log.info("Config account(%s) passwd success.", user_name)
                return [0, payload_publish]

            err_reason = ret_dict.get("message")[1]
            run_log.error("Config account(%s) info failed, %s", user_name, err_reason)
            payload_publish["reason"] = f"ERR.{MidwareErrCode.midware_common_err}, Config account password failed"
            payload_publish["result"] = "failed"
            return [-1, payload_publish]

    @staticmethod
    def _update_net_manager_params():
        try:
            net_manger_info: NetManager = NetCfgManager().get_net_cfg_info()
        except NetManagerException as err:
            run_log.warning("Get net manger info from db failed. Reason: %s", err)
            return

        data = {"net_manager_account": net_manger_info.cloud_user, "net_manager_address": net_manger_info.ip}
        if net_manger_info.port != 443:
            # 只有端口号存在，且不为443的时候，表示为用户设置的端口号，这种情况才上报域名
            data.update({"net_manager_domain": f"{net_manger_info.server_name}:{net_manger_info.port}"})

        MidwareUris.resp_json_sys_info["system"].update(data)

    @staticmethod
    def _update_product_info_from_json_data(res):
        # 未生效固件版本在post接口获取.
        version_xml = VersionXmlManager(CommonConstants.OM_VERSION_XML_FILE)
        MidwareUris.resp_json_sys_info["product_info"]["manufacturer"] = version_xml.vendor
        software_version = version_xml.firmware_version or version_xml.version
        MidwareUris.resp_json_sys_info["product_info"]["software_version"] = software_version
        # 显示名字需要systems.py从前端的配置文件中获取，上传给FD用作型号显示
        MidwareUris.resp_json_sys_info["product_info"]["product_name"] = res["Model"]
        # support_model值要能与version.xml中的SupportModel一致
        MidwareUris.resp_json_sys_info["product_info"]["support_model"] = res["SupportModel"]
        MidwareUris.resp_json_sys_info["product_info"]["serial_number"] = res["SerialNumber"]

        pcb_table = ['Ver.A', 'Ver.B', 'Ver.C', 'Ver.D']
        pcb_str = "unknown"
        if isinstance(res["Oem"]["PCBVersion"], int) and (len(pcb_table) + 1) > res["Oem"]["PCBVersion"] > 0:
            pcb_str = pcb_table[res["Oem"]["PCBVersion"] - 1]

        MidwareUris.resp_json_sys_info["product_info"]["pcb_version"] = pcb_str
        MidwareUris.resp_json_sys_info["product_info"]["os_version"] = res["Oem"]["OSVersion"]
        MidwareUris.resp_json_sys_info["product_info"]["kernel_version"] = res["Oem"]["KernelVersion"]
        MidwareUris.resp_json_sys_info["product_info"]["asset_tag"] = res["AssetTag"]
        MidwareUris.resp_json_sys_info["product_info"]["cpuArchitecture"] = res["Oem"]["ProcessorArchitecture"]

    @staticmethod
    def _get_alarm_info(alarm_id):
        if not MidwareUris.alarm_info_dict:
            ret = FileCheck.check_path_is_exist_and_valid(ConfigPathConstants.ALARM_INFO_EN_JSON)
            if not ret:
                run_log.error(f"alarm_info_en.json check result is {ret.error}")
                return {}

            with open(ConfigPathConstants.ALARM_INFO_EN_JSON) as file:
                MidwareUris.alarm_info_dict = json.loads(file.read())

        alarm_info = {}
        try:
            for item in MidwareUris.alarm_info_dict["EventSuggestion"]:
                if alarm_id == item["id"]:
                    alarm_info = item
                    break
            return alarm_info
        except Exception as err:
            run_log.error("Exception : %s", err)
        return alarm_info
