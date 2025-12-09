# coding: utf-8
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import threading
from enum import Enum

from common.utils.singleton import Singleton


class NetManagerStatus(Enum):
    NOT_CONFIGURED = "not_configured"
    CONNECTING = "connecting"
    CONNECTED = "connected"
    ERROR_CONFIGURED = "error_configured"


class FdConnectStatus(Singleton):
    """Fd对接状态"""

    lock = threading.Lock()

    def __init__(self) -> None:
        super().__init__()
        self.cur_connect_status = NetManagerStatus.NOT_CONFIGURED

    def get_cur_status(self):
        with self.lock:
            return self.cur_connect_status.value

    def set_cur_status(self, status: NetManagerStatus):
        with self.lock:
            self.cur_connect_status = status

    def trans_to_connecting(self):
        self.set_cur_status(NetManagerStatus.CONNECTING)

    def trans_to_connected(self):
        self.set_cur_status(NetManagerStatus.CONNECTED)

    def trans_to_not_configured(self):
        self.set_cur_status(NetManagerStatus.NOT_CONFIGURED)

    def trans_to_err_configured(self):
        self.set_cur_status(NetManagerStatus.ERROR_CONFIGURED)
