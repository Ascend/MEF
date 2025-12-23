# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
from typing import NoReturn, Optional, List, Dict, Any, Type

from sqlalchemy import func

from common.constants.error_codes import SecurityServiceErrorCodes
from common.db.base_models import Base
from common.log.logger import run_log
from common.checkers import ExistsChecker
from net_manager.checkers.table_data_checker import NetManagerCfgChecker, CertManagerChecker
from net_manager.constants import NetManagerConstants
from net_manager.exception import DbOperateException, DataCheckException, InvalidDataException, NetManagerException
from net_manager.models import NetManager, CertManager
from redfish_db.session import session_maker


class NetTableManagerBase:
    LIMIT_NUM = NetManagerConstants.CERT_FROM_FD_LIMIT_NUM
    FD = NetManagerConstants.FUSION_DIRECTOR
    model: Base
    checkers_key: str
    checkers_class: Type[ExistsChecker]

    @staticmethod
    def add_objs(*objs: "model") -> NoReturn:
        """
        添加数据对象
        :param objs: 数据对象
        :return:
        """
        if not objs:
            return

        with session_maker() as session:
            session.bulk_save_objects(objs)

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

    def get_obj_by_id(self, index_id: int) -> Optional["model"]:
        """
        根据id获取数据对象
        :return: "model": 数据对象
        """
        with session_maker() as session:
            obj = session.query(self.model).filter_by(id=index_id).first()
            session.expunge(obj)
            return obj

    def table_data_checker(self, check_data: Base):
        """表中数据校验"""
        try:
            check_ret = self.checkers_class(self.checkers_key).check({self.checkers_key: check_data})
        except Exception as err:
            raise DataCheckException(f"Data check failed, {err}") from err

        if not check_ret.success:
            raise DataCheckException(f"Data check failed, {check_ret.reason}", err_code=check_ret.err_code)


class NetCertManager(NetTableManagerBase):
    model = CertManager
    checkers_key = "net_crt"
    checkers_class = CertManagerChecker

    def delete_obj_by_name(self, name: str) -> int:
        """
        根据证书名字删除表中对应行数据
        :param name: 证书名字
        :return: 1: 找到数据且删除成功；0：未找到数据
        """
        with session_maker() as session:
            return session.query(self.model).filter_by(name=name).delete()

    def delete_obj_by_source(self, source: str) -> int:
        """
        根据证书来源删除表中对应行数据
        :param source: 证书名字
        :return: 1: 找到数据且删除成功；0：未找到数据
        """
        with session_maker() as session:
            return session.query(self.model).filter_by(source=source).delete()

    def get_obj_by_name(self, name: str) -> List["model"]:
        """
        根据证书名字获取数据对象列表
        :param name: 证书名字
        :return: List["model"]: 符合条件的对象列表，[]: 未找到符合条件对象
        """
        with session_maker() as session:
            obj_list = []
            for obj in session.query(self.model).filter_by(name=name).limit(self.LIMIT_NUM).all():
                session.expunge(obj)
                try:
                    self.table_data_checker(obj)
                except NetManagerException as err:
                    run_log.warning("Check cert [%s] is invalid, reason is %s", obj.name, err.err_msg)
                    continue
                obj_list.append(obj)

            return obj_list

    def get_obj_by_source(self, source: str) -> List["model"]:
        """
        根据证书来源获取所有数据对象列表
        :param source: 证书来源
        :return: List["model"]: 符合条件的对象列表，[]: 未找到符合条件对象
        """
        with session_maker() as session:
            obj_list = []
            for obj in session.query(self.model).filter_by(source=source).limit(self.LIMIT_NUM).all():
                session.expunge(obj)
                try:
                    self.table_data_checker(obj)
                except NetManagerException as err:
                    run_log.warning("Check cert [%s] is invalid, reason is %s", obj.name, err.err_msg)
                    continue
                obj_list.append(obj)

            return obj_list

    def update_obj_by_id(self, index_id: int, columns_map: Dict[str, Any]) -> NoReturn:
        """
        根据id名字更新列数据
        :param index_id: 自增id
        :param columns_map: 需要更新的列字典数据
        :return:
        """
        with session_maker() as session:
            session.query(self.model).filter_by(id=index_id).update(columns_map)

    def get_all_contain_expired(self) -> List["model"]:
        """
        获取所有正常的和过期的证书对象
        :return: List["model"]: 数据对象列表，[]: 未找到数据对象
        """
        with session_maker() as session:
            obj_list = []
            for obj in session.query(self.model).limit(NetManagerConstants.CERT_FORM_FD_AND_WEB_LIMIT_NUM).all():
                session.expunge(obj)
                if not self.table_data_expired_checker(obj):
                    continue
                obj_list.append(obj)

            return obj_list

    def table_data_expired_checker(self, obj) -> bool:
        """
        校验表中数据，但是不校验过期证书；判断证书是否是正常或已过期的
        """
        try:
            self.table_data_checker(obj)
        except NetManagerException as err:
            run_log.warning("Check cert [%s] is invalid, reason is %s", obj.name, err.err_msg)
            if err.err_code != SecurityServiceErrorCodes.ERROR_CERTIFICATE_HAS_EXPIRED.code:
                return False
        return True


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
