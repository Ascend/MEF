# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import uuid
from typing import NoReturn, Iterable

from sqlalchemy import Column, Integer, String
from sqlalchemy import DateTime, Boolean
from sqlalchemy.orm import scoped_session

from cert_manager.parse_tools import CertChainParser
from cert_manager.schemas import CertInfoSchema
from common.constants.base_constants import CommonConstants
from common.db.base_models import Base, SaveDefaultsMixin
from common.kmc_lib.kmc import Kmc
from common.log.logger import run_log
from common.utils.date_utils import DateUtils
from net_manager.checkers.contents_checker import CertInfoValidator
from net_manager.constants import NetManagerConstants
from net_manager.exception import KmcOperateException
from wsclient.connect_status import FdConnectStatus

# FD 显示时间格式
GREENWICH_MEAN_TIME = "%b %d %Y GMT"

kmc = Kmc(NetManagerConstants.REDFISH_KSF, NetManagerConstants.REDFISH_BAK_KSF, NetManagerConstants.REDFISH_ALG_CFG)


class CertManager(Base):
    __tablename__ = "cert_manager"

    id = Column(Integer, primary_key=True, comment="自增id")
    name = Column(String(64), default=NetManagerConstants.FD_CERT_NAME, comment="FusionDirector根证书名字")
    source = Column(String(16), default=NetManagerConstants.WEB, comment="FD根证书来源，取值('Web', 'FusionDirector')")
    update_time = Column(String(32), default=DateUtils.default_time, comment="FusionDirector根证书导入时间")
    cert_contents = Column(String(), default="", comment="FusionDirector根证书内容")
    crl_contents = Column(String(), default="", comment="FusionDirector吊销列表内容")

    @classmethod
    def to_obj(cls, data: dict):
        return cls(
            name=data.get("name"),
            source=data.get("source"),
            update_time=data.get("update_time"),
            cert_contents=data.get("cert_contents"),
            crl_contents=data.get("crl_contents"),
        )

    def to_dict(self):
        return {
            "name": self.name,
            "source": self.source,
            "update_time": self.update_time,
            "cert_contents": self.cert_contents,
            "crl_contents": self.crl_contents,
        }

    def get_name(self):
        return self.name


class CertInfo(Base, SaveDefaultsMixin):
    __tablename__ = "cert_info"

    id = Column(Integer, primary_key=True, comment="自增id")
    name = Column(String(64), default="", unique=True, comment="FusionDirector根证书名字，关联cert_manager")
    subject = Column(String(256), comment="持有者")
    issuer = Column(String(256), comment="颁发者")
    serial_num = Column(String(128), comment="序列号")
    signature_algorithm = Column(String(128), comment="签名算法")
    signature_len = Column(Integer, comment="签名长度")
    cert_version = Column(Integer, comment="证书版本号")
    pubkey_type = Column(Integer, comment="公钥类型")
    fingerprint = Column(String(256), comment="指纹")
    is_ca = Column(Integer, comment="是否是ca")
    chain_num = Column(Integer, comment="级链数")
    start_time = Column(DateTime, comment="开始时间")
    key_cert_sign = Column(Integer, comment="用途")
    end_time = Column(DateTime, comment="过期时间")
    in_use = Column(Boolean, default=False, comment="证书是否正被FD使用")

    @classmethod
    def to_obj(cls, data: dict):
        # 默认为未使用
        return cls(
            name=data.get("name"),
            subject=data.get("subject"),
            issuer=data.get("issuer"),
            serial_num=data.get("serial_num"),
            signature_algorithm=data.get("signature_algorithm"),
            signature_len=data.get("signature_len"),
            cert_version=data.get("cert_version"),
            pubkey_type=data.get("pubkey_type"),
            fingerprint=data.get("fingerprint"),
            is_ca=data.get("is_ca"),
            chain_num=data.get("chain_num"),
            start_time=data.get("start_time"),
            key_cert_sign=data.get("key_cert_sign"),
            end_time=data.get("end_time"),
        )

    @classmethod
    def default_cert_info_generator(cls, session: scoped_session) -> Iterable["CertInfo"]:
        for cert in session.query(CertManager):
            try:
                cert_info = cls(name=cert.name)
                cert_info.update_by_schema(
                    CertChainParser(
                        cert.cert_contents, CommonConstants.OM_UPGRADE_DIR_PATH
                    ).get_root_ca_schema(CertInfoValidator().check_cert_info)
                )
            except Exception as err:
                # 解析失败应当删除对应的证书内容，避免升级前传入的不安全证书到新版本
                session.query(CertManager).filter_by(name=cert.name).delete()
                run_log.warning("parse %s failed, because %s. ignored", cert.name, err)
                continue
            yield cert_info

    @classmethod
    def save_defaults(cls, session: scoped_session) -> NoReturn:
        session.bulk_save_objects(cls.default_cert_info_generator(session))

    def update_by_schema(self, schema: CertInfoSchema):
        for column in self.__table__.columns:
            if hasattr(schema, column.key):
                setattr(self, column.key, getattr(schema, column.key))

    def to_cert_info_dict(self):
        """返给cert_info表用的结构"""
        return {
            "name": self.name,
            "subject": self.subject,
            "issuer": self.issuer,
            "serial_num": self.serial_num,
            "signature_algorithm": self.signature_algorithm,
            "signature_len": self.signature_len,
            "cert_version": self.cert_version,
            "pubkey_type": self.pubkey_type,
            "fingerprint": self.fingerprint,
            "is_ca": self.is_ca,
            "chain_num": self.chain_num,
            "start_time": self.start_time,
            "key_cert_sign": self.key_cert_sign,
            "end_time": self.end_time,
            "in_use": self.in_use,
        }

    def get_cert_info(self) -> dict:
        """用于查询未使用证书返回"""
        return {
            "name": self.name,
            "subject": self.subject,
            "issuer": self.issuer,
            "serial number": self.serial_num,
            "signature algorithm": self.signature_algorithm,
            "signature length": self.signature_len,
            "pubkey type": self.pubkey_type,
            "fingerprint": self.fingerprint,
        }

    def to_dict(self) -> dict:
        """返给web前端用的结构"""
        return {
            "SerialNum": self.serial_num,
            "Subject": self.subject,
            "Issuer": self.issuer,
            "Fingerprint": self.fingerprint,
            "Date": f"{self.start_time}--{self.end_time}",
        }

    def content_to_dict(self, crl_exists: bool) -> dict:
        """返给FD"""
        return {
            "cert_type": "FDRootCert",
            "cert_name": self.name,
            "issuer": self.issuer,
            "subject": self.subject,
            "valid_not_before": self.start_time.strftime(GREENWICH_MEAN_TIME),
            "valid_not_after": self.end_time.strftime(GREENWICH_MEAN_TIME),
            "serial_number": self.serial_num,
            "is_import_crl": crl_exists,
            "signature_algorithm": self.signature_algorithm,
            "fingerprint": self.fingerprint,
            "key_usage": "Signing, CRL Sign",
            "public_key_length_bits2": str(self.signature_len),
        }

    def get_cert_name(self) -> str:
        return self.name


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


class FdPreCert(Base):
    __tablename__ = "fd_previous_cert"

    id = Column(Integer, primary_key=True, comment="自增id")
    name = Column(String(64), default="", comment="FD/CCAE根证书名字")
    source = Column(String(16), default="", comment="证书来源")
    update_time = Column(String(32), default=DateUtils.default_time, comment="FD/CCAE根证书备份时间")
    cert_contents = Column(String(), default="", comment="FD/CCAE根证书内容")
    crl_contents = Column(String(), default="", comment="FD/CCAE吊销列表内容")

    @classmethod
    def to_obj(cls, data: dict):
        return cls(
            name=data.get("name"),
            source=data.get("source"),
            update_time=data.get("update_time"),
            cert_contents=data.get("cert_contents"),
            crl_contents=data.get("crl_contents"),
        )

    def to_dict(self):
        return {
            "name": self.name,
            "source": self.source,
            "update_time": self.update_time,
            "cert_contents": self.cert_contents,
            "crl_contents": self.crl_contents,
        }

    def get_name(self):
        return self.name

    def get_source(self):
        return self.source


class PreCertInfo(Base):
    __tablename__ = "pre_cert_info"

    id = Column(Integer, primary_key=True, comment="自增id")
    name = Column(String(64), default="", unique=True, comment="FusionDirector根证书名字，关联cert_manager")
    subject = Column(String(256), comment="持有者")
    issuer = Column(String(256), comment="颁发者")
    serial_num = Column(String(128), comment="序列号")
    signature_algorithm = Column(String(128), comment="签名算法")
    signature_len = Column(Integer, comment="签名长度")
    cert_version = Column(Integer, comment="证书版本号")
    pubkey_type = Column(Integer, comment="公钥类型")
    fingerprint = Column(String(256), comment="指纹")
    is_ca = Column(Integer, comment="是否是ca")
    chain_num = Column(Integer, comment="级链数")
    start_time = Column(DateTime, comment="开始时间")
    key_cert_sign = Column(Integer, comment="用途")
    end_time = Column(DateTime, comment="过期时间")
    in_use = Column(Boolean, default=False, comment="证书是否正被FD使用")

    @classmethod
    def to_obj(cls, data: dict):
        return cls(
            name=data.get("name"),
            subject=data.get("subject"),
            issuer=data.get("issuer"),
            serial_num=data.get("serial_num"),
            signature_algorithm=data.get("signature_algorithm"),
            signature_len=data.get("signature_len"),
            cert_version=data.get("cert_version"),
            pubkey_type=data.get("pubkey_type"),
            fingerprint=data.get("fingerprint"),
            is_ca=data.get("is_ca"),
            chain_num=data.get("chain_num"),
            start_time=data.get("start_time"),
            key_cert_sign=data.get("key_cert_sign"),
            end_time=data.get("end_time"),
        )

    def to_dict(self) -> dict:
        return {
            "name": self.name,
            "subject": self.subject,
            "issuer": self.issuer,
            "serial_num": self.serial_num,
            "signature_algorithm": self.signature_algorithm,
            "signature_len": self.signature_len,
            "cert_version": self.cert_version,
            "pubkey_type": self.pubkey_type,
            "fingerprint": self.fingerprint,
            "is_ca": self.is_ca,
            "chain_num": self.chain_num,
            "start_time": self.start_time,
            "key_cert_sign": self.key_cert_sign,
            "end_time": self.end_time,
            "in_use": self.in_use,
        }

    def get_cert_info(self) -> dict:
        """用于查询未使用证书返回"""
        return {
            "name": self.name,
            "subject": self.subject,
            "issuer": self.issuer,
            "serial num": self.serial_num,
            "signature algorithm": self.signature_algorithm,
            "signature length": self.signature_len,
            "pubkey type": self.pubkey_type,
            "fingerprint": self.fingerprint,
        }

    def get_name(self):
        return self.name
