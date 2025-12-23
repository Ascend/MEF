# coding: utf-8
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import json
from dataclasses import dataclass

from common.schema import BaseModel
from net_manager.schemas import HeaderData
from net_manager.schemas import RouteData


@dataclass
class MefMsgData(BaseModel):
    """FD下发消息"""
    header: HeaderData
    route: RouteData
    content: dict

    def to_ws_msg_str(self):
        """websocket消息字符串格式"""
        data = {
            "header": self.header.to_dict(),
            "route": self.route.to_dict(),
            "content": self.content if isinstance(self.content, str) else json.dumps(self.content)
        }
        return json.dumps(data)
