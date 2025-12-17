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
功    能：Redfish Server Systems资源定义
"""
import os.path

from common.ResourceDefV1.resource import RfResource


class RfSystemsCollection(RfResource):
    """
    功能描述：创建Systems对象集合, 导入配置模板
    接口：NA
    """
    SYSTEM_COLLECTION_RESOURCE_DIR = os.path.normpath("redfish/v1/Systems")
    PROCESSOR_COLLECTION_RESOURCE_DIR = os.path.normpath("redfish/v1/Systems/Processors")
    CPU_PROCESSOR_RESOURCE_DIR = os.path.normpath("redfish/v1/Systems/Processors/CPU")
    AI_PROCESSOR_RESOURCE_DIR = os.path.normpath("redfish/v1/Systems/Processors/AiProcessor")
    NTP_SERVICE_RESOURCE_DIR = os.path.normpath("redfish/v1/Systems/NTPService")
    LOG_SERVICE_COLLECTION_RESOURCE_DIR = os.path.normpath("redfish/v1/Systems/LogServices")
    LOG_COLLECTION_PROGRESS_RESOURCE_DIR = os.path.normpath("redfish/v1/Systems/LogServices/progress")
    MEMORY_RESOURCE_DIR = os.path.normpath("redfish/v1/Systems/Memory")
    LTE_RESOURCE_DIR = os.path.normpath("redfish/v1/Systems/LTE")
    LTE_STATUSINFO_RESOURCE_DIR = os.path.normpath("redfish/v1/Systems/LTE/StatusInfo")
    LTE_CONFIGINFO_RESOURCE_DIR = os.path.normpath("redfish/v1/Systems/LTE/ConfigInfo")
    ACCESS_CONTROL_DIR = os.path.normpath("redfish/v1/Systems/AccessControl")
    NFS_MANAGE_RESOURCE_DIR = os.path.normpath("redfish/v1/Systems/NfsManage")
    SIMPLE_STORAGE_COLLECTION_RESOURCE_DIR = os.path.normpath("redfish/v1/Systems/SimpleStorage")
    SIMPLE_STORAGE_RESOURCE_DIR = os.path.normpath("redfish/v1/Systems/SimpleStorage/Members")
    PARTITION_COLLECTION_RESOURCE_DIR = os.path.normpath("redfish/v1/Systems/Partition")
    PARTITION_RESOURCE_DIR = os.path.normpath("redfish/v1/Systems/Partition/Members")
    ETHERNET_COLLECTION_RESOURCE_DIR = os.path.normpath("redfish/v1/Systems/EthernetInterfaces")
    ETHERNET_GETMEMBERS_RESOURCE_DIR = os.path.normpath("redfish/v1/Systems/EthernetInterfaces/GetMembers")
    ETHERNET_PATCHMEMBERS_RESOURCE_DIR = os.path.normpath("redfish/v1/Systems/EthernetInterfaces/PatchMembers")
    SYSTEM_TIME_RESOURCE_DIR = os.path.normpath("redfish/v1/Systems/Time")
    ETHERNET_ETHIPLIST_RESOURCE_DIR = os.path.normpath("redfish/v1/Systems/EthernetEthiplist")
    ALARM_RESOURCE_DIR = os.path.normpath("redfish/v1/Systems/Alarm")
    ALARM_INFO_RESOURCE_DIR = os.path.normpath("redfish/v1/Systems/Alarm/AlarmInfo")
    ALARM_SHIELD_RESOURCE_DIR = os.path.normpath("redfish/v1/Systems/Alarm/AlarmShield")
    SECURITY_SERVICE_RESOURCE_DIR = os.path.normpath("redfish/v1/Systems/SecurityService")
    HTTPSCERT_RESOURCE_DIR = os.path.normpath("redfish/v1/Systems/SecurityService/HttpsCert")
    HTTPSCERT_ALARM_TIME_RESOURCE_DIR = os.path.normpath("redfish/v1/Systems/SecurityService/HttpsCertAlarmTime")
    SYSINFO_RESOURCE_DIR = os.path.normpath("redfish/v1/Systems/Sysinfo")
    SYSSTATUS_RESOURCE_DIR = os.path.normpath("redfish/v1/Systems/Sysstatus")
    EXTENDED_DEVICES_COLLECTION_RESOURCE_DIR = os.path.normpath("redfish/v1/Systems/ExtendedDevices")
    EXTENDED_DEVICES_MEMBER_RESOURCE_DIR = os.path.normpath("redfish/v1/Systems/ExtendedDevices/Members")
    GET_COLLECTION_RESOURCE_DIR = os.path.normpath("redfish/v1/Systems")
    GET_MODULE_COLLECTION_RESOURCE_DIR = os.path.normpath("redfish/v1/Systems/Modules")
    GET_MODULE_INFO_RESOURCE_DIR = os.path.normpath("redfish/v1/Systems/Modules/Info")
    GET_DEVICE_INFO_RESOURCE_DIR = os.path.normpath("redfish/v1/Systems/Modules/Info/1")
    SECURITY_LOAD_RESOURCE_DIR = os.path.normpath("redfish/v1/Systems/SecurityService/SecurityLoad")

    system_resource_collection: RfResource
    lte_resource: RfResource
    lte_statusinfo_resource: RfResource
    lte_configinfo_resource: RfResource
    access_control_resource: RfResource
    processor_resource_collection: RfResource
    cpu_processor_resource: RfResource
    ai_processor_resource: RfResource
    ntp_service_resource: RfResource
    log_service_resource_collection: RfResource
    log_collect_progress: RfResource
    memory_resource: RfResource
    security_service_resource: RfResource
    httpscert_resource: RfResource
    httpscert_alarm_time_resource: RfResource
    nfs_manage_resource: RfResource
    simple_storage_resource_collection: RfResource
    simple_storage_resource: RfResource
    partition_resource_collection: RfResource
    partition_resource: RfResource
    ethernet_resource_collection: RfResource
    ethernet_getmembers_resource: RfResource
    ethernet_patchmembers_resource: RfResource
    system_time_resource: RfResource
    ethernet_ethiplist_resource: RfResource
    alarm_resource: RfResource
    alarm_info_resource: RfResource
    alarm_shield_resource: RfResource
    sysinfo_resource: RfResource
    sysstatus_resource: RfResource
    extended_devices_resource_collection: RfResource
    extended_devices_resource_member: RfResource
    get_resource_collection: RfResource
    module_collection_resource: RfResource
    module_info_resource: RfResource
    device_info_resource: RfResource
    security_load_resource: RfResource

    def create_sub_objects(self, base_path, rel_path):
        self.create_basic_sub_objects(base_path, rel_path)
        self.create_extend_sub_objects(base_path, rel_path)

    def create_basic_sub_objects(self, base_path, rel_path):
        self.system_resource_collection = RfResource(base_path, self.SYSTEM_COLLECTION_RESOURCE_DIR)
        self.processor_resource_collection = RfResource(base_path, self.PROCESSOR_COLLECTION_RESOURCE_DIR)
        self.cpu_processor_resource = RfResource(base_path, self.CPU_PROCESSOR_RESOURCE_DIR)
        self.ai_processor_resource = RfResource(base_path, self.AI_PROCESSOR_RESOURCE_DIR)
        self.ntp_service_resource = RfResource(base_path, self.NTP_SERVICE_RESOURCE_DIR)
        self.log_service_resource_collection = RfResource(base_path, self.LOG_SERVICE_COLLECTION_RESOURCE_DIR)
        self.log_collect_progress = RfResource(base_path, self.LOG_COLLECTION_PROGRESS_RESOURCE_DIR)
        self.memory_resource = RfResource(base_path, self.MEMORY_RESOURCE_DIR)
        self.lte_resource = RfResource(base_path, self.LTE_RESOURCE_DIR)
        self.lte_statusinfo_resource = RfResource(base_path, self.LTE_STATUSINFO_RESOURCE_DIR)
        self.lte_configinfo_resource = RfResource(base_path, self.LTE_CONFIGINFO_RESOURCE_DIR)
        self.access_control_resource = RfResource(base_path, self.ACCESS_CONTROL_DIR)
        self.nfs_manage_resource = RfResource(base_path, self.NFS_MANAGE_RESOURCE_DIR)
        self.simple_storage_resource_collection = RfResource(base_path, self.SIMPLE_STORAGE_COLLECTION_RESOURCE_DIR)
        self.simple_storage_resource = RfResource(base_path, self.SIMPLE_STORAGE_RESOURCE_DIR)
        self.partition_resource_collection = RfResource(base_path, self.PARTITION_COLLECTION_RESOURCE_DIR)
        self.partition_resource = RfResource(base_path, self.PARTITION_RESOURCE_DIR)
        self.ethernet_resource_collection = RfResource(base_path, self.ETHERNET_COLLECTION_RESOURCE_DIR)
        self.ethernet_getmembers_resource = RfResource(base_path, self.ETHERNET_GETMEMBERS_RESOURCE_DIR)
        self.ethernet_patchmembers_resource = RfResource(base_path, self.ETHERNET_PATCHMEMBERS_RESOURCE_DIR)
        self.system_time_resource = RfResource(base_path, self.SYSTEM_TIME_RESOURCE_DIR)
        self.ethernet_ethiplist_resource = RfResource(base_path, self.ETHERNET_ETHIPLIST_RESOURCE_DIR)
        self.alarm_resource = RfResource(base_path, self.ALARM_RESOURCE_DIR)
        self.alarm_info_resource = RfResource(base_path, self.ALARM_INFO_RESOURCE_DIR)
        self.alarm_shield_resource = RfResource(base_path, self.ALARM_SHIELD_RESOURCE_DIR)
        self.security_service_resource = RfResource(base_path, self.SECURITY_SERVICE_RESOURCE_DIR)
        self.httpscert_resource = RfResource(base_path, self.HTTPSCERT_RESOURCE_DIR)
        self.httpscert_alarm_time_resource = RfResource(base_path, self.HTTPSCERT_ALARM_TIME_RESOURCE_DIR)
        self.sysinfo_resource = RfResource(base_path, self.SYSINFO_RESOURCE_DIR)
        self.sysstatus_resource = RfResource(base_path, self.SYSSTATUS_RESOURCE_DIR)
        self.extended_devices_resource_collection = RfResource(base_path, self.EXTENDED_DEVICES_COLLECTION_RESOURCE_DIR)
        self.extended_devices_resource_member = RfResource(base_path, self.EXTENDED_DEVICES_MEMBER_RESOURCE_DIR)
        self.get_resource_collection = RfResource(base_path, self.GET_COLLECTION_RESOURCE_DIR)
        self.module_collection_resource = RfResource(base_path, self.GET_MODULE_COLLECTION_RESOURCE_DIR)
        self.module_info_resource = RfResource(base_path, self.GET_MODULE_INFO_RESOURCE_DIR)
        self.device_info_resource = RfResource(base_path, self.GET_DEVICE_INFO_RESOURCE_DIR)
        self.security_load_resource = RfResource(base_path, self.SECURITY_LOAD_RESOURCE_DIR)

    def create_extend_sub_objects(self, base_path, rel_path):
        pass
