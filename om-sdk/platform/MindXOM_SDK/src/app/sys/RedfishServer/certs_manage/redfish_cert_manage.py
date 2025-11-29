# -*- coding: utf-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
import json
import signal
import sys
from argparse import ArgumentParser
from enum import Enum
from typing import Dict, Callable

from common.log.logger import run_log, terminal_print
from common.utils.result_base import Result
from common.utils.scripts_utils import signal_handler
from net_manager.manager.net_cfg_manager import CertMgr


class Action(Enum):
    GETUNUSED = "getunusedcert"
    RESTORE = "restorecert"
    DELETE = "deletecert"


class Component(Enum):
    FD = "fd-ccae"
    HTTPS = "https"


def extend_cert_nameger(component: str, action: str) -> Result:
    try:
        from extend_interfaces import extend_certs_manager_funcs
        return extend_certs_manager_funcs(action)
    except ImportError as error:
        run_log.warning("Current device not support %s type. Error: %s", component, error)
        return Result(False, err_msg=f"Import extend cert manager func failed. Error: {error}")
    except Exception as error:
        run_log.error("Call extend cert manage func failed. %s %s failed. Error: %s", action, component, error)
        return Result(False, err_msg=f"Call extend cert manage func failed. {action} {component} failed.")


def restore_cert(component: str, _) -> Result:
    if component == Component.FD.value:
        try:
            CertMgr().restore_fd_pre_cert()
        except Exception:
            run_log.error("Restore fd-ccae previous cert failed.")
            return Result(False, err_msg="Restore fd-ccae previous cert failed.")

        run_log.info("Restore fd-ccae previous cert successfully.")
        return Result(True, data="Restore fd-ccae previous cert successfully.")
    elif component == Component.HTTPS.value:
        return extend_cert_nameger(component, Action.RESTORE.value)


def get_unused_cert(component: str, _) -> Result:
    if component == Component.FD.value:
        try:
            cert_info = CertMgr().get_fd_unused_cert()
            res_dict = json.loads(cert_info)
        except Exception as error:
            run_log.error("Get fd-ccae unused cert failed. %s", error)
            return Result(False, err_msg="Get fd-ccae unused cert failed.")

        if not res_dict.get(CertMgr.PRE_KEY) and res_dict.get(CertMgr.UNUNSED_KEY) == CertMgr.UNUNSED_VAL:
            run_log.error("Get fd-ccae unused cert failed. Unused cert not exist.")
            return Result(False, err_msg="Get fd-ccae unused cert failed. Unused cert not exist.")

        run_log.info("Get fd-ccae unused cert successfully.")
        return Result(True, data=cert_info)
    elif component == Component.HTTPS.value:
        return extend_cert_nameger(component, Action.GETUNUSED.value)


def delete_unused_cert(component: str, name: str) -> Result:
    if component == Component.FD.value:
        try:
            ret = CertMgr().del_fd_unused_cert_by_name(name)
        except Exception:
            run_log.error("Delete fd-ccae unused cert failed.")
            return Result(False, err_msg="Delete fd-ccae unused cert failed.")
        if not ret:
            run_log.error("Delete fd-ccae unused cert failed.")
            return Result(False, err_msg="Delete fd-ccae unused cert failed.")

        run_log.info("Delete fd-ccae unused cert successfully.")
        return Result(True, data="Delete fd-ccae unused cert successfully.")
    elif component == Component.HTTPS.value:
        return extend_cert_nameger(component, Action.DELETE.value)


OPERATE: Dict[str, Callable[[str, str], Result]] = {
    Action.GETUNUSED.value: get_unused_cert,
    Action.RESTORE.value: restore_cert,
    Action.DELETE.value: delete_unused_cert,
}


def parse_args():
    parse = ArgumentParser()
    parse.add_argument("--action", type=str, choices={item.value for item in Action}, help="操作类型")
    parse.add_argument("--component", type=str, choices={item.value for item in Component}, help="操作组件")
    parse.add_argument("--name", type=str, default="", help="操作名称，可选，默认为空")
    return parse.parse_args()


if __name__ == '__main__':
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)
    try:
        args = parse_args()
        res = OPERATE.get(args.action, lambda: Result(False, err_msg="Action not found"))(args.component, args.name)
        terminal_print.info(res.data if res else res.error)
        sys.exit(0 if res else 1)
    except Exception as err:
        terminal_print.error("Cert manage failed. Error: %s", err)
        sys.exit(1)
