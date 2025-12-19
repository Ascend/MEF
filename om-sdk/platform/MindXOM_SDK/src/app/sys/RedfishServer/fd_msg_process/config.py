# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import threading


class Topic:
    # subscribe topics
    SUB_RESET_ALARM = r"$hw/edge/v1/hardware/operate/rearm"
    # 配置主机名
    SUB_CONFIG_HOST_NAME = "$hw/edge/v1/hardware/operate/config_hostname"
    # 复位主机系统
    SUB_COMPUTER_SYSTEM_RESET = "$hw/edge/v1/hardware/operate/restart"
    # 配置导入
    SUB_OPERATE_PROFILE = r"$hw/edge/v1/hardware/operate/profile"
    # 自定义电子标签
    SUB_HARDWARE_OPERATE_TAG = "$hw/edge/v1/hardware/operate/tag"
    # 重置设备账号密码
    SUB_PASS_THROUGH_ACCOUNT_MODIFY = "$hw/edge/v1/hardware/operate/passthrough/account_modify"
    # 配置生效
    SUB_OPERATE_PROFILE_EFFECT = r"$hw/edge/v1/hardware/operate/profile_effect"
    # 信息收集
    SUB_OPERATE_INFO_COLLECT = r"$hw/edge/v1/hardware/operate/info_collect"
    # 固件升级/主机软件安装
    SUB_OPERATE_INSTALL = r"$hw/edge/v1/hardware/operate/install"
    # 固件生效
    SUB_OPERATE_FIRMWARE_EFFECTIVE = r"$hw/edge/v1/hardware/operate/firmware_effective"
    # 配置网管信息
    SUB_CONFIG_NET_MANAGER = "$hw/edge/v1/hardware/operate/netmanager"
    # FD证书查询
    SUB_CERT_QUERY = "$hw/edge/v1/hardware/operate/cert_query"
    # FD导入证书
    SUB_CERT_UPDATE = "$hw/edge/v1/hardware/operate/cert_update"
    # FD导入吊销列表
    SUB_CRL_UPDATE = "$hw/edge/v1/hardware/operate/crl_update"
    # FD删除证书
    SUB_CERT_DELETE = "$hw/edge/v1/hardware/operate/cert_delete"
    # publish topics
    # 硬件告警/事件
    PUB_REPORT_ALARM_TO_FD = r"$hw/edge/v1/hub/report/alarm"
    # 系统状态信息
    PUB_REPORT_SYS_STATUS_TO_FD = r"$hw/edge/v1/hub/report/sys_status"
    # 通过EdgeCore向FD发布配置结果消息
    PUB_CONFIG_RESULT_TO_FD = r"$hw/edge/v1/hub/report/config_result"
    # 固件升级过程中上报可以重启，上报系统可复位状态
    PUB_REPORT_RESTART_RESULT = r"$hw/edge/v1/hub/report/restart_result"
    # 上报系统配置进度
    PUB_REPORT_SYS_CONFIG_PROGRESS = r"$hw/edge/v1/hub/report/profile_effect"
    # 上报信息收集进度
    PUB_REPORT_INFO_COLLECT_PROGRESS = r"$hw/edge/v1/hub/report/info_collect_process"
    # 固件升级进度上报
    PUB_REPORT_UPGRADE_PROGRESS = r"$hw/edge/v1/hub/report/upgrade_progress"


class Config:
    mqtt_max_msg_payload_size = 256 * 1024  # 每条接收到的mqtt的消息payload最大长度


class SysInfoTaskStatus:
    """SysInfo信息收集任务状态"""
    LOCK = threading.RLock()
    # 自从进程启动之后系统信息收集是否完成一次收集
    done: bool = False

    @staticmethod
    def get():
        with SysInfoTaskStatus.LOCK:
            return SysInfoTaskStatus.done

    @staticmethod
    def set():
        with SysInfoTaskStatus.LOCK:
            SysInfoTaskStatus.done = True
