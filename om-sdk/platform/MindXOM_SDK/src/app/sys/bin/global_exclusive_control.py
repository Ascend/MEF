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
import threading
from typing import NoReturn


class GlobalExclusiveController:
    """Monitor进程涉及系统重启等高风险的操作时，进行互斥操作控制"""
    high_exclusive_lock = threading.Lock()

    @classmethod
    def locked(cls) -> bool:
        """只要条件有任何一个满足，则认为当前存在互斥关系"""
        return cls.high_exclusive_lock.locked()

    @classmethod
    def acquire(cls, release_time: int = 600) -> NoReturn:
        cls.high_exclusive_lock.acquire()
        threading.Timer(function=cls.release, interval=release_time)

    @classmethod
    def release(cls) -> NoReturn:
        cls.high_exclusive_lock.release()
