# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import time
import uuid

from common.kmc_lib.kmc import Kmc
from sqlalchemy import Column, Integer, String

from common.db.base_models import Base, SaveDefaultsMixin
from common.utils.date_utils import DateUtils
from net_manager.constants import NetManagerConstants
from net_manager.exception import KmcOperateException
from wsclient.connect_status import FdConnectStatus

kmc = Kmc(NetManagerConstants.REDFISH_KSF, NetManagerConstants.REDFISH_BAK_KSF, NetManagerConstants.REDFISH_ALG_CFG)


class CertManager(Base):
    __tablename__ = "cert_manager"

    id = Column(Integer, primary_key=True, comment="自增id")
    name = Column(String(64), default="", comment="FusionDirector根证书名字")
    source = Column(String(16), default="Web", comment="FusionDirector根证书来源，取值('Web', 'FusionDirector')")
    update_time = Column(String(32), default="", comment="FusionDirector根证书导入时间")
    cert_contents = Column(String(), default="", comment="FusionDirector根证书内容")
    crl_contents = Column(String(), default="", comment="FusionDirector吊销列表内容")

    @classmethod
    def from_dict(cls, data: dict):
        """将data转成CertManager对象"""
        return cls(
            name=data.get("CertName", NetManagerConstants.FD_CERT_NAME),
            source=data.get("Source", NetManagerConstants.WEB),
            update_time=data.get("UpdateTime", DateUtils.get_format_time(time.time())),
            cert_contents=data.get("CertContents", ""),
            crl_contents=data.get("CrlContents", ""),
        )

    def to_dict(self):
        return {
            "name": self.name,
            "source": self.source,
            "update_time": self.update_time,
            "cert_contents": self.cert_contents,
            "crl_contents": self.crl_contents,
        }


class NetManager(Base, SaveDefaultsMixin):
    __tablename__ = "net_manager"

    id = Column(Integer, primary_key=True, comment="自增id")
    net_mgmt_type = Column(String(16), default="Web", comment="网管模式，初始化为点对点Web管理模式")
    node_id = Column(String(64), default=str(uuid.uuid4()), comment="节点ID，初始化为UUID")
    server_name = Column(String(64), default="", comment="服务器名称，FusionDirector管理模式存在")
    ip = Column(String(16), default="", comment="对接IP地址，FusionDirector管理模式存在")
    port = Column(String(16), default="", comment="对接端口号，FusionDirector管理模式存在")
    cloud_user = Column(String(256), default="", comment="对接账号，FusionDirector管理模式存在")
    cloud_pwd = Column(String(256), comment="对接密码，FusionDirector管理模式存在，kmc加密保存")
    status = Column(String(16), default="", comment="对接状态，取值范围('', 'connecting', 'connected', 'ready')")

    @staticmethod
    def encrypt_cloud_pwd(cloud_pwd):
        try:
            return kmc.encrypt(cloud_pwd)
        except Exception as err:
            raise KmcOperateException("encrypt cloud pwd failed!") from err

    @classmethod
    def from_dict(cls, data: dict):
        """将data转成NetManager对象"""
        return cls(
            net_mgmt_type=data.get("ManagerType"),
            node_id=data.get("NodeId"),
            server_name=data.get("ServerName", ""),
            ip=data.get("NetIP"),
            port=data.get("Port"),
            cloud_user=data.get("NetAccount"),
            cloud_pwd=cls.encrypt_cloud_pwd(data.get("NetPassword")),
            status=data.get("Status", ""),
        )

    def decrypt_cloud_pwd(self):
        try:
            return kmc.decrypt(self.cloud_pwd)
        except Exception as err:
            raise KmcOperateException("decrypt cloud pwd failed!") from err

    def to_dict_for_query(self) -> dict:
        return {
            "NetManager": self.net_mgmt_type,
            "NetIP": self.ip,
            "Port": self.port,
            "NetAccount": self.cloud_user,
            "ServerName": self.server_name,
            "ConnectStatus": FdConnectStatus().get_cur_status(),
        }

    def to_dict_for_update(self) -> dict:
        return {
            "net_mgmt_type": self.net_mgmt_type,
            "node_id": self.node_id,
            "server_name": self.server_name,
            "ip": self.ip,
            "port": self.port,
            "cloud_user": self.cloud_user,
            "cloud_pwd": self.cloud_pwd,
            "status": self.status,
        }
