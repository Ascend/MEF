#!/usr/bin/python3
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.

from typing import NoReturn, List, Optional, Dict, Iterable

from sqlalchemy import func

from common.db.base_models import Base
from common.log.logger import run_log
from om_event_subscription.constants import MIN_SUBSCRIPTION_ID, MAX_SUBSCRIPTION_ID
from om_event_subscription.models import Subscription, SubscriptionCert
from redfish_db.session import session_maker


class BaseManager:
    """
    库redfish_edge.db中表对应对象的基类，
    定义了根据id列表删除、批量添加、查询所有的函数
    """
    model: Base

    @staticmethod
    def add_objs(objs: Iterable["model"]):
        with session_maker() as session:
            session.bulk_save_objects(objs)

    def get_all(self) -> List["model"]:
        with session_maker() as session:
            objs = session.query(self.model).all()
            for obj in objs:
                session.expunge(obj)
            return objs

    def delete_by_ids(self, ids: List[int]) -> int:
        with session_maker() as session:
            return session.query(self.model).filter(self.model.id.in_(ids)).delete()

    def get_first(self) -> Optional["model"]:
        with session_maker() as session:
            obj = session.query(self.model).order_by(self.model.id).first()
            if obj:
                session.expunge(obj)
            return obj


class SubscriptionsMgr(BaseManager):
    model = Subscription

    @staticmethod
    def add_subscription(subscription: Subscription) -> NoReturn:
        with session_maker() as session:
            session.add(subscription)

    def subscription_detail(self, _id: int) -> Dict:
        """查询订阅事件详情"""
        with session_maker() as session:
            subscription = session.query(self.model).filter_by(id=_id).first()
            if not subscription:
                return {}

            session.expunge(subscription)
        return subscription.to_dict()

    def delete_subscription(self, _id: int):
        """删除指定ID的订阅事件记录"""
        with session_maker() as session:
            return session.query(self.model).filter(self.model.id == _id).delete()

    def get_first_subscription(self) -> List[int]:
        with session_maker() as session:
            subscription = session.query(self.model).order_by(self.model.id).first()
            return [subscription.id] if subscription else []

    def check_destination_existed(self, destination: str) -> bool:
        """
        功能描述：判断 destination 是否已存在于 Subscription 表中
        """
        subscriptions = self.get_valid_subscriptions()
        if not subscriptions:
            return False
        for subscription in subscriptions:
            if subscription.get_decrypt_destination() == destination:
                return True
        return False

    def get_subscription_num(self) -> int:
        """
        功能描述：获取已订阅的数量
        """
        with session_maker() as session:
            return session.query(func.count(self.model.id)).scalar()

    def get_available_min_id(self) -> int:
        """
        功能描述：获取最小可用 id，即最小的未被订阅的资源 id。
        """
        ids = set(self.get_first_subscription())
        min_id = 0
        try:
            min_id = min(set(range(MIN_SUBSCRIPTION_ID, MAX_SUBSCRIPTION_ID + 1)) - ids)
        except Exception as err:
            run_log.error("Get min id failed. Error is %s", err)

        return min_id

    def get_valid_subscriptions(self) -> List[Subscription]:
        """
        功能描述：获取有效的订阅者，目前只支持一个
        """
        subscription = self.get_first()
        return [subscription] if subscription else []


class SubscriptionsCertMgr(BaseManager):
    model = SubscriptionCert

    def get_obj_by_id(self, index_id: int) -> Optional["model"]:
        """
        根据id获取数据对象
        :return: "model": 数据对象
        """
        with session_maker() as session:
            obj = session.query(self.model).filter_by(id=index_id).first()
            if obj:
                session.expunge(obj)
            return obj

    def overwrite_subs_cert(self, obj: SubscriptionCert) -> NoReturn:
        with session_maker() as session:
            session.query(self.model).delete()
            session.add(obj)

    def delete_by_cert_id(self, cert_id: int) -> int:
        with session_maker() as session:
            return session.query(self.model).filter(self.model.root_cert_id == cert_id).delete()

    def update_crt_with_crl(self, crt: str, crl: Dict) -> NoReturn:
        with session_maker() as session:
            session.query(self.model).filter_by(cert_contents=crt).update(crl)
