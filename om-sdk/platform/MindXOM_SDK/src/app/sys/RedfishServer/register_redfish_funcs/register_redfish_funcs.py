#!/usr/bin/python3
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from common.log.logger import run_log


def start_om_extend_funcs():
    try:
        from extend_interfaces import register_extend_func
        register_extend_func()
    except ImportError as err:
        run_log.warning("Failed to import extension, ignore. %s", err)
    except Exception as err:
        run_log.error("Register extend func failed, catch %s", err.__class__.__name__)
