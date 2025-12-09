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
import os.path
from typing import Iterable

from sqlalchemy import Column, String, Integer

from common.constants.upgrade_constants import OMUpgradeConstants
from common.db.base_models import Base, SaveDefaultsMixin
from common.file_utils import FileUtils
from common.common_methods import CommonMethods


class NetConfig(Base, SaveDefaultsMixin):
    __tablename__ = "net_config"

    id = Column(Integer, primary_key=True)
    name = Column(String, comment="网络接口名称")
    tag = Column(String, comment="ip对应的用途")
    ipv4 = Column(String, comment="ipv4地址")

    @classmethod
    def default_instance_generator(cls) -> Iterable["NetConfig"]:
        # 如果有tag.ini与迁移数据库标记，则将数据迁移至数据库
        if not os.path.exists(CommonMethods.NET_TAG_INI) or not os.path.exists(OMUpgradeConstants.CONFIG_TO_DB_FLAG):
            return

        ret = CommonMethods.load_net_tag_ini()
        if not ret:
            raise Exception(f"load tag.ini failed, {ret.error}")

        tag_info = ret.data
        for eth_name in ("eth0", "eth1"):
            for tag, ipv4 in zip(tag_info.get(eth_name, []), tag_info.get(f"ip_{eth_name}", [])):
                yield cls(name=eth_name, tag=tag, ipv4=ipv4)

        FileUtils.delete_file_or_link(OMUpgradeConstants.CONFIG_TO_DB_FLAG)
