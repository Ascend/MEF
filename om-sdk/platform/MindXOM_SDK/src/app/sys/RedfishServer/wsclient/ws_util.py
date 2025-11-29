# coding: utf-8
#  Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from typing import Tuple

from websockets import datastructures
from websockets.datastructures import Headers
from websockets.legacy.client import WebSocketClientProtocol
from websockets.legacy.exceptions import InvalidStatusCode


class WsInvalidStatusCode(InvalidStatusCode):

    def __init__(self, status_code: int, headers: datastructures.Headers, body: bytes) -> None:
        super().__init__(status_code, headers)
        self.body = body


class WsClientProtocol(WebSocketClientProtocol):
    """自定义Websocket客户端协议消息处理类"""
    LIMIT_1M_SIZE = 1 * 1024 * 1024
    REDIRECT_STATUS_CODES = (301, 302, 303, 307, 308)

    async def read_http_response(self) -> Tuple[int, Headers]:
        status_code, headers = await super().read_http_response()

        if status_code in WsClientProtocol.REDIRECT_STATUS_CODES:
            return status_code, headers

        if status_code != 101:
            resp_body = await self.reader.read(WsClientProtocol.LIMIT_1M_SIZE)
            raise WsInvalidStatusCode(status_code, headers, resp_body)

        return status_code, headers
