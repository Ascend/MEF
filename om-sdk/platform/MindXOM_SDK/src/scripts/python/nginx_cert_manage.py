# -*- coding: utf-8 -*-
#  Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
import os
import signal
import sys
from argparse import ArgumentParser
from enum import Enum
from typing import Dict, Callable, NoReturn

from common.constants.base_constants import CommonConstants
from common.file_utils import FileCheck, FileCopy, FileReader, FileUtils
from common.kmc_lib.kmc import Kmc
from common.kmc_lib.tlsconfig import TlsConfig
from common.log.logger import run_log
from common.utils.result_base import Result
from common.utils.scripts_utils import signal_handler
from logger import terminal_print


class NginxCertMgr:
    def __init__(self):
        self.nginx_cert_dir = CommonConstants.NGINX_KS_DIR
        self.nginx_pre_cert_dir = CommonConstants.WEB_PRE_DIR
        self.pre_cert = os.path.join(self.nginx_pre_cert_dir, "server_kmc.cert")
        self.pre_certs_priv = os.path.join(self.nginx_pre_cert_dir, "server_kmc.priv")
        self.pre_cert_psd = os.path.join(self.nginx_pre_cert_dir, "server_kmc.psd")
        self.pre_cert_primary_ksf = os.path.join(self.nginx_pre_cert_dir, "om_cert.keystore")
        self.pre_cert_standby_ksf = os.path.join(self.nginx_pre_cert_dir, "om_cert_backup.keystore")
        self.pre_alg_config = os.path.join(self.nginx_pre_cert_dir, "om_alg.json")
        self.kmc_inst = Kmc(self.pre_cert_primary_ksf, self.pre_cert_standby_ksf, self.pre_alg_config)
        self.cert_related_file = (
            self.pre_cert, self.pre_certs_priv, self.pre_cert_psd, self.pre_cert_primary_ksf,
            self.pre_cert_standby_ksf, self.pre_alg_config
        )
        self.file_count = 6
        self.file_max_size = 1 * 1024 * 1024
        self.expired_info = "Previous web cert has expired."

    def check_cert_pkey_match(self) -> Result:
        try:
            with open(self.pre_cert_psd, "r") as file:
                pwd = self.kmc_inst.decrypt(file.read())
        except Exception as error:
            return Result(False, err_msg=f"decrypt fialed. Error: {error}")

        if not TlsConfig.get_ssl_context(None, self.pre_cert, self.pre_certs_priv, pwd)[0]:
            return Result(False, err_msg=f"get context failed.")

        return Result(True)

    def check_cert_is_valid(self) -> Result:
        try:
            from lib.Linux.systems.security_service.security_service_clib import certificate_verification, \
                check_cert_expired
        except Exception as error:
            return Result(False, err_msg=f"Import security service failed. Error: {error.__class__.__name__}")

        for file_name in self.cert_related_file:
            ret = FileCheck.check_path_is_exist_and_valid(file_name)
            if not ret:
                return Result(False, err_msg=f"Related file is invalid. {ret.error}")

            if os.path.getsize(file_name) > self.file_max_size:
                return Result(False, err_msg=f"{file_name} oversize.")

        if not self.check_cert_pkey_match():
            return Result(False, err_msg="Check cert and pkey matching failed.")

        if not check_cert_expired(self.pre_cert):
            return Result(False, err_msg=self.expired_info)

        return Result(True)

    def copy_pre_cert_to_work_dir(self) -> NoReturn:
        for file in self.cert_related_file:
            FileCopy.copy_file(file, os.path.join(self.nginx_cert_dir, os.path.basename(file)),
                               0o600, "nobody", "nobody")

    def restore_pre_cert(self) -> Result:
        ret = self.check_cert_is_valid()
        if not ret:
            run_log.error("Restore previous web cert failed. %s", ret.error)
            return Result(False, err_msg="Restore previous web cert failed.")
        try:
            self.copy_pre_cert_to_work_dir()
        except Exception as error:
            run_log.error("Restore previous web cert failed. Error: %s", error.__class__.__name__)
            return Result(False, err_msg="Restore previous web cert failed.")

        run_log.info("Restore previous web cert successfully.")
        return Result(True, data="Restore previous web cert successfully.")

    def get_unused_cert(self) -> Result:
        ret = self.check_cert_is_valid()
        if not ret:
            if ret.error == self.expired_info:
                run_log.warning("Previous web cert exists but has expired.")
            return Result(False, err_msg=f"Previous web cert is invalid. {ret.error}")

        result = FileReader(self.pre_cert).read()
        if not result:
            run_log.error("Get previous web cert content failed. %s", result.error)
            return Result(False, err_msg=result.error)

        run_log.info("Get previous web cert content successfully.")
        return Result(True, data=f"cert name: {self.pre_cert}\ncert content: {result.data}")

    def delete_unused_cert(self):
        ret = self.get_unused_cert()
        if not ret:
            run_log.warning("Unused web cert is invalid and will be deleted if exist. %s", ret.error)
            return Result(False, err_msg=f"Delete unused web cert failed. {ret.error}")

        try:
            for file in self.cert_related_file:
                FileUtils.delete_file_or_link(file)
        except Exception as error:
            run_log.error("Delete previous web cert failed. Error: %s", error)
            return Result(False, err_msg="Delete unused web cert failed.")

        run_log.info("Delete previous web cert successfully.")
        return Result(True, data="Delete unused web cert successfully.")


class Action(Enum):
    GETUNUSED = "getunusedcert"
    RESTORE = "restorecert"
    DELETE = "deletecert"


class Component(Enum):
    WEB = "web"


def parse_args():
    parse = ArgumentParser()
    parse.add_argument("--action", type=str, choices={item.value for item in Action}, help="操作类型")
    parse.add_argument("--component", type=str, choices={item.value for item in Component}, help="操作组件")
    return parse.parse_args()


OPERATE: Dict[str, Callable[[], Result]] = {
    Action.GETUNUSED.value: NginxCertMgr().get_unused_cert,
    Action.RESTORE.value: NginxCertMgr().restore_pre_cert,
    Action.DELETE.value: NginxCertMgr().delete_unused_cert,
}

if __name__ == '__main__':
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)
    try:
        args = parse_args()
        res = OPERATE.get(args.action, lambda: Result(False, err_msg="Action not found"))()
        terminal_print.info(res.data if res else res.error)
        sys.exit(0 if res else 1)
    except Exception as err:
        terminal_print.error("Cert manage failed. Error: %s", err)
        sys.exit(1)
