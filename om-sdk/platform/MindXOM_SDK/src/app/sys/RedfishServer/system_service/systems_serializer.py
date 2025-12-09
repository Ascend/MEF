# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
from ibma_redfish_serializer import Serializer


class ExtendedDevicesCollectionResourceSerializer(Serializer):
    """
    功能描述：外部设备资源集合
    """
    service = Serializer.root.systems.extended_devices_resource_collection


class ExtendedDevicesMemberResourceSerializer(Serializer):
    """
    功能描述：外部设备资源
    """
    service = Serializer.root.systems.extended_devices_resource_member


class SystemServiceResourceSerializer(Serializer):
    """
    功能描述：系统资源集合
    """
    service = Serializer.root.systems.system_resource_collection


class ProcessorResourceSerializer(Serializer):
    """
    功能描述: 处理器资源集合
    """
    service = Serializer.root.systems.processor_resource_collection


class CPUProcessorResourceSerializer(Serializer):
    """
    功能描述：CPU资源
    """
    service = Serializer.root.systems.cpu_processor_resource


class AiProcessorResourceSerializer(Serializer):
    """
    功能描述：AI处理器资源
    """
    service = Serializer.root.systems.ai_processor_resource


class NtpResourceSerializer(Serializer):
    """
    功能描述：NTP资源
    """
    service = Serializer.root.systems.ntp_service_resource


class LogServiceResourceSerializer(Serializer):
    """
    功能描述：日志服务资源集合
    """
    service = Serializer.root.systems.log_service_resource_collection


class LogCollectProgressResourceSerializer(Serializer):
    """
    功能描述：日志下载进度
    """
    service = Serializer.root.systems.log_collect_progress


class MemoryResourceSerializer(Serializer):
    """
    功能描述：内存资源
    """
    service = Serializer.root.systems.memory_resource


class NfsManageServiceResourceSerializer(Serializer):
    """
    功能描述：Nfs Manage资源
    """
    service = Serializer.root.systems.nfs_manage_resource


class SimpleStorageCollectionResourceSerializer(Serializer):
    """
    功能描述：简单存储资源集合
    """
    service = Serializer.root.systems.simple_storage_resource_collection


class LteResourceSerializer(Serializer):
    """
    功能描述：LTE资源集合
    """
    service = Serializer.root.systems.lte_resource


class LteStatusInfoResourceSerializer(Serializer):
    """
    功能描述：LTE状态信息资源
    """
    service = Serializer.root.systems.lte_statusinfo_resource


class LteConfigInfoResourceSerializer(Serializer):
    """
    功能描述：LTE配置信息资源
    """
    service = Serializer.root.systems.lte_configinfo_resource


class AccessControlSerializer(Serializer):
    """
    功能描述：近端访问能力信息资源
    """
    service = Serializer.root.systems.access_control_resource


class SimpleStorageResourceSerializer(Serializer):
    """
    功能描述：简单存储资源
    """
    service = Serializer.root.systems.simple_storage_resource


class PartitionCollectionResourceSerializer(Serializer):
    """
    功能描述：磁盘分区资源集合
    """
    service = Serializer.root.systems.partition_resource_collection


class PartitionResourceSerializer(Serializer):
    """
    功能描述：磁盘分区资源
    """
    service = Serializer.root.systems.partition_resource


class EthernetCollectionResourceSerializer(Serializer):
    """
    功能描述：以太网集合资源
    """
    service = Serializer.root.systems.ethernet_resource_collection


class EthernetGetMembersResourceSerializer(Serializer):
    """
    功能描述：以太网资源 GET操作
    """
    service = Serializer.root.systems.ethernet_getmembers_resource


class EthernetPatchMembersResourceSerializer(Serializer):
    """
    功能描述： 以太网资源 PATCH操作
    """
    service = Serializer.root.systems.ethernet_patchmembers_resource


class SystemTimeResourceSerializer(Serializer):
    """
    功能描述：系统时间资源
    """
    service = Serializer.root.systems.system_time_resource


class EthernetEthIpListResourceSerializer(Serializer):
    """
    功能描述：网口与IP列表资源
    """
    service = Serializer.root.systems.ethernet_ethiplist_resource


class AlarmResourceSerializer(Serializer):
    """
    功能描述：Alarm资源集合
    """
    service = Serializer.root.systems.alarm_resource


class AlarmInfoResourceSerializer(Serializer):
    """
    功能描述：Alarm状态信息资源
    """
    service = Serializer.root.systems.alarm_info_resource


class AlarmShieldResourceSerializer(Serializer):
    """
    功能描述：Alarm屏蔽规则
    """
    service = Serializer.root.systems.alarm_shield_resource


class SecurityServiceResourceSerializer(Serializer):
    """
    功能描述：安全服务顶层资源
    """
    service = Serializer.root.systems.security_service_resource


class HttpsCertResourceSerializer(Serializer):
    """
    功能描述：证书信息资源
    """
    service = Serializer.root.systems.httpscert_resource


class HttpsCertAlarmTimeResourceSerializer(Serializer):
    """
    功能描述：证书有效期提醒时间资源
    """
    service = Serializer.root.systems.httpscert_alarm_time_resource


class ModuleCollectionResourceSerializer(Serializer):
    """ 功能描述：外部设备资源集合"""
    service = Serializer.root.systems.module_collection_resource


class ModuleInfoResourceSerializer(Serializer):
    """ 功能描述： 外部设备模组信息 """
    service = Serializer.root.systems.module_info_resource


class DeviceInfoResourceSerializer(Serializer):
    """ 功能描述： 外部设备信息 """
    service = Serializer.root.systems.device_info_resource


class SecurityLoadResourceSerializer(Serializer):
    """
    功能描述：登录规则资源
    """
    service = Serializer.root.systems.security_load_resource
