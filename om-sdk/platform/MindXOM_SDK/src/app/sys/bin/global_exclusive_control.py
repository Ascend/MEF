# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
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
