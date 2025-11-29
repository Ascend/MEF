# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from typing import List, Iterable

from lib.Linux.systems.nic.models import NetConfig
from monitor_db.session import session_maker


class NetConfigMgr:

    @staticmethod
    def query_info_with_condition(**condition) -> Iterable[NetConfig]:
        with session_maker() as session:
            yield from session.query(NetConfig).filter_by(**condition).order_by(NetConfig.name, NetConfig.ipv4).all()

    @staticmethod
    def query_tag_from_ip(ipv4, dynamic) -> str:
        with session_maker() as session:
            if dynamic:
                ipv4 = "dhcp"
            net_cfg = session.query(NetConfig).filter_by(ipv4=ipv4).first()
            return net_cfg.tag if net_cfg else ""

    @staticmethod
    def delete_specific_eth_config(**condition) -> int:
        with session_maker() as session:
            return session.query(NetConfig).filter_by(**condition).delete()

    @staticmethod
    def save_net_config(eth_cfgs: List[NetConfig]):
        with session_maker() as session:
            session.bulk_save_objects(eth_cfgs)
