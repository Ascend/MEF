# coding: utf-8
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.

from dataclasses import dataclass, field
from typing import Optional

from common.schema import BaseModel


@dataclass
class SystemInfo(BaseModel):
    asset_tag: Optional[str] = field(metadata={"alias": "AssetTag"})
    serial_number: str = field(metadata={"alias": "SerialNumber"}, default="")
    product_name: str = field(metadata={"alias": "Model"}, default="Atlas 500")


@dataclass
class WebNetInfo(BaseModel):
    cloud_pwd: str = field(metadata={"alias": "NetPassword"}, default="")
    net_mgmt_type: str = field(metadata={"alias": "ManagerType"}, default="Web")
    server_name: str = field(metadata={"alias": "ServerName"}, default="")
    ip: str = field(metadata={"alias": "NetIP"}, default="")
    port: str = field(metadata={"alias": "Port"}, default="")
    cloud_user: str = field(metadata={"alias": "NetAccount"}, default="")
    status: str = field(metadata={"alias": "Status"}, default="")


@dataclass
class HeaderData(BaseModel):
    """FD下发消息的header字段"""
    msg_id: str
    parent_msg_id: str
    sync: bool
    timestamp: int


@dataclass
class RouteData(BaseModel):
    """FD下发消息的route字段"""
    group: str
    operation: str
    resource: str
    source: str


@dataclass
class PayloadPublish(BaseModel):
    """上报FD消息字段"""
    topic: str
    percentage: str = field(default="0%")
    result: str = field(default="failed")
    reason: str = field(default="")
