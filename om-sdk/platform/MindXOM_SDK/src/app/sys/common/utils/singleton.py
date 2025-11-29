# !/usr/bin/python
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import threading


class _Singleton(type):
    """A metaclass that creates a Singleton base class when called."""

    _instances = {}
    _lock = threading.RLock()

    def __call__(cls, *args, **kwargs):
        with cls._lock:
            if cls not in cls._instances:
                cls._instances[cls] = super(_Singleton, cls).__call__(*args, **kwargs)
        return cls._instances.get(cls)


class Singleton(metaclass=_Singleton):
    pass
