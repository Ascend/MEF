# coding: utf-8
#  Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
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
