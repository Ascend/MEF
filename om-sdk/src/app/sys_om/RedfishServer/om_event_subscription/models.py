#!/usr/bin/python3
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

from datetime import datetime

from sqlalchemy import Column, Integer, String, DateTime

from common.kmc_lib.kmc import Kmc
from common.db.base_models import Base
from common.utils.common_check import CommonCheck
from om_event_subscription.constants import (EventTypes, Protocol, ESIGHT_KSF,
                                             ESIGHT_BAK_KSF, ESIGHT_ALG_CONFIG, AlarmTypes)
from om_event_subscription.errors import KmcDecryptError

kmc = Kmc(ESIGHT_KSF, ESIGHT_BAK_KSF, ESIGHT_ALG_CONFIG)


class Subscription(Base):
    __tablename__ = "subscriptions"

    id = Column(Integer, primary_key=True, nullable=False, comment="订阅ID，范围1~4")
    destination = Column(String(256), unique=True, comment="合法的https接收地址，订阅者地址")
    types = Column(String(256), default=EventTypes.ALERT.value, comment="监听事件类型，多个用','分开")
    protocol = Column(String(64), default=Protocol.REDFISH.value, comment="事件订阅使用的协议")
    credential = Column(String(), comment="上报给订阅者时的认证凭据，需加密保存")
    create_time = Column(DateTime(), default=datetime.utcnow, comment="创建订阅时间")
    update_time = Column(DateTime(), default=datetime.utcnow, onupdate=datetime.utcnow, comment="更新订阅时间")

    @classmethod
    def from_dict(cls, data: dict):
        """
        功能描述：将data转换为Subscription对象
        """
        return cls(
            id=data["id"],
            destination=kmc.encrypt(data["Destination"]),
            types=",".join(data["EventTypes"]),
            protocol=data["Protocol"],
            credential=kmc.encrypt(data["HttpHeaders"]["X-Auth-Token"])
        )

    def to_dict(self) -> dict:
        """
        功能描述：将实例转换成字典，以供序列化成接口返回数据使用
        """
        return {
            "Id": str(self.id),
            "Name": f"EventSubscription {self.id}",
            "@odata.id": f"/redfish/v1/EventService/Subscriptions/{self.id}",
            "Destination": self.get_decrypt_destination(),
            "EventTypes": self.types.split(","),
            "Protocol": self.protocol,
        }

    def get_decrypt_destination(self):
        try:
            return kmc.decrypt(self.destination)
        except Exception as err:
            raise KmcDecryptError("decrypt destination failed!") from err

    def get_decrypt_credential(self):
        try:
            return kmc.decrypt(self.credential)
        except Exception as err:
            raise KmcDecryptError("decrypt credential failed!") from err


class SubscriptionCert(Base):

    __tablename__ = "subscriptions_cert"

    id = Column(Integer, primary_key=True, comment="自增id")
    root_cert_id = Column(Integer, comment="根证书ID")
    type = Column(String(16), default="", comment="证书导入方式")
    cert_contents = Column(String(), default="", comment="根证书内容")
    crl_contents = Column(String(), default="", comment="吊销列表内容")

    @classmethod
    def from_dict(cls, data: dict):
        """将data转成SubscriptionCert对象"""
        return cls(
            root_cert_id=data.get("root_cert_id"),
            type=data.get("type"),
            cert_contents=data.get("cert_contents"),
            crl_contents=data.get("crl_contents"),
        )

    def to_dict(self):
        return {
            "root_cert_id": self.root_cert_id,
            "type": self.type,
            "cert_contents": self.cert_contents,
            "crl_contents": self.crl_contents,
        }


class ActiveAlarm(Base):

    __tablename__ = 'active_alarm'

    id = Column(Integer, primary_key=True, comment="自增id")
    alarm_id = Column(String(32), comment="告警id")
    alarm_instance = Column(String(32), comment="告警实体，哪个部件产生了告警")
    alarm_name = Column(String, comment="告警名称，与告警ID对应")
    type = Column(String(), default=AlarmTypes.ALERT.value, comment="告警类型")
    severity = Column(Integer, comment="告警级别 0：Critical；1：Major；2：Minor")
    create_time = Column(DateTime(), default=datetime.utcnow, comment="创建告警时间")

    def __eq__(self, other):
        return self.alarm_id == other.alarm_id and self.alarm_instance == other.alarm_instance

    def __hash__(self):
        return hash((self.alarm_id, self.alarm_instance))

    @classmethod
    def from_dict(cls, data: dict):
        default_list = ["AlarmId", "AlarmInstance", "AlarmName", "Timestamp", "PerceivedSeverity"]
        if not CommonCheck.check_sub_list(list(data.keys()), default_list):
            return cls()

        if not data.get("Timestamp") or not data.get("Timestamp", "").isdigit():
            raise Exception("alarm {} create_time is wrong".format(cls.alarm_id))

        if not data.get("PerceivedSeverity") or not data.get("PerceivedSeverity", "").isdigit():
            raise Exception("alarm {} PerceivedSeverity is wrong".format(cls.alarm_id))

        return cls(
            alarm_id=data.get("AlarmId"),
            alarm_instance=data.get("AlarmInstance"),
            alarm_name=data.get("AlarmName"),
            create_time=datetime.fromtimestamp(int(data.get("Timestamp"))),
            type=data.get("Severity") or AlarmTypes.ALERT.value,
            severity=data.get("PerceivedSeverity")
        )


class AlarmReportTask(Base):

    __tablename__ = "alarm_report_task"

    id = Column(Integer, primary_key=True, comment="自增id")
    subscriber_id = Column(Integer, comment="订阅者id")
    event_type = Column(String, comment="告警类型")
    event_id = Column(String, comment="告警事件流水id")
    event_name = Column(String, comment="告警事件名称")
    severity = Column(String, comment="告警事件级别")
    event_detail = Column(String(), default="", comment="告警描述")
    reason = Column(String(), default="", comment="告警原因")
    extra_column = Column(String(), default="", comment="扩展字段")
    event_timestamp = Column(DateTime(), default=datetime.utcnow, comment="事件产生时间")
    task_status = Column(Integer, comment="任务状态：状态为待新增（1）、已新增（2）、待消除（3）、已消除（4）")

    def to_dict(self):
        return {
            "id": self.id,
            "subscriber_id": self.subscriber_id,
            "event_type": self.event_type,
            "event_id": self.event_id,
            "event_timestamp": self.event_timestamp,
            "event_name": self.event_name,
            "severity": self.severity,
            "task_status": self.task_status,
            "event_detail": self.event_detail,
            "reason": self.reason,
            "extra_column": self.extra_column,
        }
