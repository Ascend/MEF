# Copyright (c) Huawei Technologies Co., Ltd. 2023-2025. All rights reserved.
import datetime
import json
from typing import NoReturn, Optional, List, Type, Iterable

from sqlalchemy import func
from sqlalchemy.orm import Query, scoped_session

from cert_manager.parse_tools import CertChainParser, CrlChainParser
from common.checkers import ExistsChecker
from common.db.base_models import Base
from common.log.logger import run_log
from common.utils.result_base import Result
from net_manager.checkers.contents_checker import CertInfoValidator
from net_manager.checkers.external_params_checker import CertNameChecker
from net_manager.checkers.table_data_checker import NetManagerCfgChecker
from net_manager.constants import NetManagerConstants
from net_manager.exception import DbOperateException, DataCheckException, InvalidDataException, NetManagerException
from net_manager.models import NetManager, CertManager, CertInfo, FdPreCert, PreCertInfo
from redfish_db.session import session_maker, simple_session_maker


class NetTableManagerBase:
    LIMIT_NUM = NetManagerConstants.CERT_FROM_FD_LIMIT_NUM
    model: Base
    checkers_key: str
    checkers_class: Type[ExistsChecker]

    def get_all(self) -> List["model"]:
        """
        获取所有数据对象
        :return: List["model"]: 数据对象列表，[]: 未找到数据对象
        """
        with session_maker() as session:
            obj_list = []
            for obj in session.query(self.model).limit(self.LIMIT_NUM + 1).all():
                session.expunge(obj)
                try:
                    self.table_data_checker(obj)
                except NetManagerException as err:
                    run_log.warning("Check cert [%s] is invalid, reason is %s", obj.name, err.err_msg)
                    continue
                obj_list.append(obj)

            return obj_list

    def table_data_checker(self, check_data: Base):
        """表中数据校验"""
        try:
            check_ret = self.checkers_class(self.checkers_key).check({self.checkers_key: check_data})
        except Exception as err:
            raise DataCheckException(f"Data check failed, {err}") from err

        if not check_ret.success:
            raise DataCheckException(f"Data check failed, {check_ret.reason}", err_code=check_ret.err_code)


class NetCfgManager(NetTableManagerBase):
    """网管配置表管理类."""
    model = NetManager
    checkers_key = "net_cfg"
    checkers_class = NetManagerCfgChecker

    def get_net_cfg_info(self) -> NetManager:
        try:
            net_cfg_info: NetManager = self._get_data_to_first()
        except Exception as err:
            raise DbOperateException(f"Get net config info failed.") from err

        self.table_data_checker(net_cfg_info)
        return net_cfg_info

    def update_net_cfg_info(self, set_data_map: dict) -> NoReturn:
        try:
            self._update_data_to_first(set_data_map)
        except Exception as err:
            raise DbOperateException(f"Update net config failed.") from err

    def _update_data_to_first(self, data) -> NoReturn:
        # net_manager表有且只有一条数据，每次更新直接更新第一条数据即可
        with session_maker() as session:
            if session.query(func.count(self.model.id)).scalar() != 1:
                raise InvalidDataException("Net config data is invalid.")

            session.query(self.model).update(data)

    def _get_data_to_first(self) -> Optional["model"]:
        # net_manager表有且只有一条数据，获取表第一条数据.
        with session_maker() as session:
            if session.query(func.count(self.model.id)).scalar() != 1:
                raise InvalidDataException("Net config data is invalid.")

            obj = session.query(self.model).first()
            session.expunge(obj)
            return obj


class CertMgr:
    """用于CertManager、CertInfo、FdPreCert数据表的增删查改"""
    cert = CertManager
    cert_info = CertInfo
    pre_cert = FdPreCert
    pre_cert_info = PreCertInfo
    session_maker = simple_session_maker
    net_mgr = NetCfgManager()
    cert_validator = CertInfoValidator()

    UNUNSED_KEY = "unused cert"
    UNUNSED_VAL = "unused cert not exist."
    PRE_KEY = "previous cert"

    FD_CRL_NAME = "FD.crl"

    def get_expire_certs(self, threshold_days: int) -> Iterable[CertInfo]:
        """即将过期与已过期证书"""
        with self.session_maker() as session:
            yield from session.query(self.cert_info).filter(
                self.cert_info.end_time < datetime.datetime.utcnow() + datetime.timedelta(days=threshold_days)
            ).limit(NetManagerConstants.CERT_FORM_FD_AND_WEB_LIMIT_NUM)

    def is_crl_expired(self, threshold_days: int) -> bool:
        """即将过期与已过期的CRL"""
        with self.session_maker() as session:
            cert = session.query(self.cert).first()
            if not cert.crl_contents:
                return False
            return CrlChainParser(cert.crl_contents).verify_crl_chain_overdue(
                datetime.datetime.utcnow() + datetime.timedelta(days=threshold_days))

    def del_by_name(self, name: str, session: scoped_session = None) -> int:
        with self.session_maker(session) as session:
            session.query(self.cert).filter_by(name=name).delete()
            return session.query(self.cert_info).filter_by(name=name).delete()

    def get_used_cert_name(self, session: scoped_session = None) -> str:
        with self.session_maker(session) as session:
            cert_obj: CertInfo = session.query(self.cert_info).filter_by(in_use=True).first()
            if not cert_obj:
                run_log.warning("Using cert info not exist.")
                return ""
            return cert_obj.get_cert_name()

    def get_used_cert(self, session: scoped_session = None):
        using_cert_name = self.get_used_cert_name()
        if not using_cert_name:
            run_log.warning("No FD cert is using.")
            return None

        with self.session_maker(session) as session:
            obj = session.query(self.cert).filter_by(name=using_cert_name).first()
            session.expunge(obj)
            return obj

    def get_cert_info_by_name(self, name: str) -> CertInfo:
        with self.session_maker() as session:
            obj = session.query(self.cert_info).filter_by(name=name).first()
            if not obj:
                run_log.warning("Cert info not exist.")
                raise InvalidDataException("Cert info not exist.")
            session.expunge(obj)
            return obj

    def del_cert_info_by_name(self, name: str) -> int:
        with self.session_maker() as session:
            return session.query(self.cert_info).filter_by(name=name).delete()

    def backup_previous_cert(self):
        # 只要成功与fd建连过，必定存在使用过的证书，当有使用过的证书时，备份使用过的证书，没有使用过的证书，说明还未对接过fd，备份web导入的前一份证书
        run_log.info("Backup previous FD cert start.")
        used_cert = self.get_used_cert()
        if used_cert:
            with self.session_maker() as session:
                # 备份cert
                session.query(self.pre_cert).delete()
                session.add(self.pre_cert.to_obj(used_cert.to_dict()))
                # 备份cert_info
                pre_cert_info = self.pre_cert_info.to_obj(self.get_cert_info_by_name(used_cert.get_name()).
                                                          to_cert_info_dict())
                session.query(self.pre_cert_info).delete()
                session.add(pre_cert_info)
        else:
            if not self._number_of_certs(source=NetManagerConstants.WEB):
                run_log.warning("Backup previous FD cert ignored, because web source FD cert not exist.")
                return

            with self.session_maker() as session:
                web_cert: CertManager = session.query(self.cert).filter_by(source=NetManagerConstants.WEB).first()
                # 备份cert
                session.query(self.pre_cert).delete()
                session.add(self.pre_cert.to_obj(web_cert.to_dict()))
                # 备份cert_info
                pre_cert_info = self.pre_cert_info.to_obj(self.get_cert_info_by_name(web_cert.get_name()).
                                                          to_cert_info_dict())
                session.query(self.pre_cert_info).delete()
                session.add(pre_cert_info)

        run_log.info("Backup previous FD cert successfully.")

    def check_fd_connection(self) -> bool:
        try:
            net_cfg = self.net_mgr.get_net_cfg_info()
        except Exception as err:
            run_log.warning("Get net manager info from db failed. Error: %s", err)
            return False

        return net_cfg.net_mgmt_type == NetManagerConstants.FUSION_DIRECTOR and net_cfg.status == "ready"

    def check_pre_cert_is_invalid(self, cert_obj: FdPreCert):
        try:
            cert_chain_parser = CertChainParser(cert_obj.cert_contents)
            for cert_schema in cert_chain_parser.cert_schema_generator():
                self.cert_validator.check_cert_info(cert_schema)
            cert_chain_parser.verify_cert_chain()
        except Exception as err:
            run_log.error("Previous cert is invalid.")
            raise InvalidDataException("Cert contents checkers: invalid cert.") from err

    def get_fd_pre_cert(self) -> FdPreCert:
        with self.session_maker() as session:
            fd_pre_cert = session.query(self.pre_cert).first()
            if not fd_pre_cert:
                run_log.warning("FD previous cert not exist.")
                raise InvalidDataException("FD previous cert not exist.")
            session.expunge(fd_pre_cert)
            return fd_pre_cert

    def get_fd_pre_cert_info(self) -> PreCertInfo:
        with self.session_maker() as session:
            pre_cert_info = session.query(self.pre_cert_info).first()
            if not pre_cert_info:
                run_log.warning("FD previous cert info not exist.")
                raise InvalidDataException("FD previous cert info not exist.")
            session.expunge(pre_cert_info)
            return pre_cert_info

    def get_unused_cert_info(self):
        with self.session_maker() as session:
            for cert_info in session.query(self.cert_info).filter_by(in_use=False)\
                    .limit(NetManagerConstants.CERT_FORM_FD_AND_WEB_LIMIT_NUM):
                yield cert_info.get_cert_info()

    def get_fd_pre_cert_name(self) -> str:
        return self.get_fd_pre_cert().get_name()

    def check_pre_cert_exists(self) -> bool:
        with self.session_maker() as session:
            if not session.query(self.cert).filter_by(name=self.get_fd_pre_cert_name()).first():
                return False

            return True

    def restore_fd_pre_cert(self):
        run_log.info("Restore FD previous cert start.")
        pre_cert: CertManager = self.get_fd_pre_cert()
        self.check_pre_cert_is_invalid(pre_cert)
        pre_cert_info: PreCertInfo = self.get_fd_pre_cert_info()
        with self.session_maker() as session:
            # 未连接fd时，清空证书并恢复前一份证书
            if not self.check_fd_connection():
                run_log.info("FD is not connected, overwrite the cert manager table.")
                session.query(self.cert).delete()
                session.add(self.cert.to_obj(pre_cert.to_dict()))
                session.query(self.cert_info).delete()
                session.add(self.cert_info.to_obj(pre_cert_info.to_dict()))

            # fd已连接，且前一份证书在FD证书列表中，不恢复
            elif self.check_pre_cert_exists():
                run_log.warning("Previous FD cert already exists in db, no need to restore.")
                raise InvalidDataException("Previous FD cert already exists in db, no need to restore.")

            # fd已连接，不存在同名证书，删除来源与前一份证书相同的最新导入的证书，恢复前一份证书
            else:
                run_log.info("FD is connected, restore previous cert to cert manager table.")
                record_to_delete: CertManager = session.query(self.cert).filter_by(source=pre_cert.get_source()). \
                    order_by(self.cert.update_time.desc()).first()
                if record_to_delete:
                    session.delete(record_to_delete)
                    session.query(self.cert_info).filter_by(name=record_to_delete.get_name()).delete()
                    run_log.info("Deleted the newly imported cert with the same source as the previous cert.")
                session.add(self.cert.to_obj(pre_cert.to_dict()))
                session.add(self.cert_info.to_obj(pre_cert_info.to_dict()))

    def get_fd_unused_cert(self) -> str:
        res = {self.UNUNSED_KEY: self.UNUNSED_VAL}
        unused_cert_info = list(self.get_unused_cert_info())
        if unused_cert_info:
            run_log.info("Get unused fd-ccae cert from cert manager succeed.")
            res[self.UNUNSED_KEY] = unused_cert_info

        try:
            pre_cert_info: PreCertInfo = self.get_fd_pre_cert_info()
        except Exception:
            run_log.error("Get unused fd-ccae cert from previous cert failed.")
            return json.dumps(res)

        res[self.PRE_KEY] = pre_cert_info.get_cert_info()
        run_log.info("Get unused cert successfully.")
        return json.dumps(res)

    def del_pre_cert_by_name(self, name: str) -> int:
        with self.session_maker() as session:
            session.query(self.pre_cert).filter_by(name=name).delete()
            return session.query(self.pre_cert_info).filter_by(name=name).delete()

    def check_cert_exist_by_name(self, name):
        with self.session_maker() as session:
            cert_info = session.query(self.cert_info).filter_by(name=name).first()
            if not cert_info:
                run_log.warning("%s cert not exist in cert info.", name)
            pre_cert_info = session.query(self.pre_cert_info).filter_by(name=name).first()
            if not pre_cert_info:
                run_log.warning("%s cert not exist in previous cert info.", name)

            return True if cert_info or pre_cert_info else False

    def del_fd_unused_cert_by_name(self, name: str):
        if not CertNameChecker().check({"cert_name": name}):
            run_log.error("Cert name is invalid.")
            raise InvalidDataException("Cert name is invalid.")
        # 同名证书不存在
        if not self.check_cert_exist_by_name(name):
            run_log.error("Previous cert with name [%s] not exist.", name)
            return Result(False, err_msg=f"Unused cert with name {name} not exist.")
        # 同名证书存在，但在使用，尝试删除同名的前一份证书
        if name == self.get_used_cert_name():
            run_log.warning("The cert with name [%s] is in use, can not delete.", name)

            if self.get_fd_pre_cert_info().get_name() == name and self.del_pre_cert_by_name(name):
                run_log.info("Delete previous cert with name [%s] successfully.", name)
                return Result(True)

            return Result(False, err_msg=f"Delete unused cert with name {name} failed.")
        # 同名证书存在，没有使用，优先删除同名当前证书，当前证书不存在再删除前一份证书
        else:
            if not self.del_by_name(name) and not self.del_pre_cert_by_name(name):
                run_log.error("Delete unused cert with name [%s] failed.", name)
                return Result(False, err_msg=f"Delete unused cert with name {name} failed.")

            run_log.info("Delete cert with name [%s] successfully.", name)
            return Result(True)

    def _certs_within_the_validity_period(self, session) -> Query:
        """有效期内的证书"""
        time_now = datetime.datetime.utcnow()
        return session.query(self.cert, self.cert_info).join(
            self.cert_info, self.cert.name == self.cert_info.name
        ).filter(
            self.cert_info.start_time < time_now, self.cert_info.end_time > time_now
        )

    def _number_of_certs(self, source: Optional[str] = None) -> int:
        with self.session_maker() as session:
            certs = session.query(self.cert)
            certs = certs.filter_by(source=source) if source else certs
            return certs.count()
