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
import os
import shutil
import signal
import sys

from common.file_utils import FileCheck
from common.file_utils import FileCreate
from common.file_utils import FilePermission
from common.log.logger import run_log
from common.utils.scripts_utils import signal_handler
from common.kmc_lib.kmc import Kmc
from common.utils.result_base import Result
from lib.Linux.systems.security_service.create_custom_csr import CustomCrtMgr
from lib.Linux.systems.security_service.custom_cert_info import CustomCertInfo
from lib.Linux.systems.security_service.security_service import SecurityService
from lib.Linux.systems.security_service.security_service_clib import ERR_SIGNATURE_LEN_NOT_ENOUGH
from lib.Linux.systems.security_service.security_service_clib import certificate_verification
from common.common_methods import CommonMethods


class ImportCrtMgr(CustomCrtMgr):

    def __init__(self) -> None:
        super().__init__()
        self._server_cert = os.path.join(CustomCertInfo.IMPORT_TMP_DIR, CustomCertInfo.SERVER_KMC_CERT)
        self._server_csr = os.path.join(CustomCertInfo.IMPORT_TMP_DIR, CustomCertInfo.SERVER_KMC_CSR)
        self._server_pri = os.path.join(CustomCertInfo.EXPORT_TMP_DIR, CustomCertInfo.SERVER_KMC_PRI)
        self._server_psd = os.path.join(CustomCertInfo.EXPORT_TMP_DIR, CustomCertInfo.SERVER_KMC_PSD)
        self._kmc_primary_ksf = os.path.join(CustomCertInfo.EXPORT_TMP_DIR, CustomCertInfo.OM_CERT_KEYSTORE)
        self._kmc_standby_ksf = os.path.join(CustomCertInfo.EXPORT_TMP_DIR, CustomCertInfo.OM_CERT_BACKUP_KEYSTORE)
        self._alg_json = os.path.join(CustomCertInfo.EXPORT_TMP_DIR, CustomCertInfo.OM_ALG_JSON)
        self._necessary_files = (
            self._kmc_primary_ksf, self._kmc_standby_ksf, self._alg_json,
            self._server_psd, self._server_pri, self._server_cert
        )
        # 导入失败时删除证书
        self._file_items_to_remove = (self._server_cert,)
        # 导入成功时证书与其他临时文件都要删除
        self._file_items_to_remove_when_success = (
            self._server_csr, self._server_pri, self._server_psd, self._kmc_primary_ksf, self._kmc_standby_ksf,
            self._alg_json
        )
        self._kmc = Kmc(self._kmc_primary_ksf, self._kmc_standby_ksf, self._alg_json)

    @staticmethod
    def _check_file_items_info(file_items):
        if not file_items:
            return Result(True)

        try:
            for src, dst in file_items:
                ret = FileCheck.check_path_is_exist_and_valid(src)
                if not ret:
                    return Result(result=False, err_msg=f"{src} path invalid: {ret.error}")

                if not os.path.isfile(src):
                    return Result(result=False, err_msg=f"{src} path invalid: is not file")

                if os.path.getsize(src) > CustomCertInfo.FILE_MAX_SIZE:
                    return Result(result=False, err_msg=f"{src} path invalid: {ret.error}")

                ret = FileCheck.check_input_path_valid(dst)
                if not ret:
                    return Result(result=False, err_msg=f"{dst} path invalid: {ret.error}")
        except Exception as ex:
            return Result(result=False, err_msg=f"check file failed: {ex}")

        return Result(result=True)

    def import_server_cert(self):
        ret = self._verify_server_cert()
        if not ret:
            run_log.error("verify server cert failed: %s", ret.error)
            return Result(result=False, err_msg=f"verify server cert failed: {ret.error}")

        dest_dirs = (
            (CustomCertInfo.GOLDEN_MOUNT_DIR, CustomCertInfo.BACKUP_FILE_MODE, CustomCertInfo.BACKUP_DIR_OWNER,
             CustomCertInfo.BACKUP_DIR_GROUP),
            (CustomCertInfo.PACKAGE_MOUNT_DIR, CustomCertInfo.BACKUP_FILE_MODE, CustomCertInfo.BACKUP_DIR_OWNER,
             CustomCertInfo.BACKUP_DIR_GROUP),
        )

        # 恢复出厂和恢复最小系统的分区证书来源
        for path, file_mode, owner, group in dest_dirs:
            dest_path = os.path.join(path, "om_certs")
            ret = self._save_files(dest_path, file_mode, owner, group)
            if not ret:
                run_log.error("save cert files to %s failed: %s", dest_path, ret.error)
                return Result(result=False, err_msg=f"save cert files to {dest_path} failed: {ret.error}")

        # 导入成功需要删除所有临时文件
        self._file_items_to_remove += self._file_items_to_remove_when_success
        run_log.info("import server cert success")
        return Result(result=True)

    def _check_cert_info(self):
        """
        校验证书的基本信息: notAfter/notBefore/Signature/Algorithm/Private key match
        """
        try:
            ret = SecurityService.check_cert_pkey_match(self._kmc, self._server_cert, self._server_pri,
                                                        self._server_psd)
        except Exception:
            return Result(result=False, err_msg="check cert pkey match failed")

        if not ret:
            run_log.error("Import custom certificate failed: the certificate does not match the private key.")
            return Result(result=False, err_msg="check cert pkey match failed")

        cert_check_result = certificate_verification(self._server_cert)
        if cert_check_result not in (0, ERR_SIGNATURE_LEN_NOT_ENOUGH,):
            run_log.error("Import custom certificate type error.")
            return cert_check_result

        run_log.info("signature algorithm and pubkey info and signature version is ok.")
        return Result(result=True)

    def _verify_server_cert(self):
        for file in self._necessary_files:
            ret = FileCheck.check_path_is_exist_and_valid(file)
            if not ret:
                run_log.error("%s path invalid: %s", file, ret.error)
                return Result(result=False, err_msg=f"{file} invalid: {ret.error}")

            if os.path.getsize(file) > CustomCertInfo.FILE_MAX_SIZE:
                run_log.error("%s path invalid: size is too large", file)
                return Result(result=False, err_msg=f"{file} invalid: size is too large")

        ret = self._check_cert_info()
        if not ret:
            run_log.error("server cert invalid: %s", ret.error)
            return Result(result=False, err_msg=f"server cert invalid: {ret.error}")

        run_log.info("verify server cert success")
        return Result(True)

    def _save_files(self, dest_dir, file_mode, owner, group):
        if not os.path.exists(dest_dir):
            FileCreate.create_dir(dest_dir, CustomCertInfo.WORK_DIR_MODE)

        file_items = (
            (self._server_psd, os.path.join(dest_dir, CustomCertInfo.SERVER_KMC_PSD)),
            (self._server_pri, os.path.join(dest_dir, CustomCertInfo.SERVER_KMC_PRI)),
            (self._server_cert, os.path.join(dest_dir, CustomCertInfo.SERVER_KMC_CERT)),
            (self._kmc_primary_ksf, os.path.join(dest_dir, CustomCertInfo.OM_CERT_KEYSTORE)),
            (self._kmc_standby_ksf, os.path.join(dest_dir, CustomCertInfo.OM_CERT_BACKUP_KEYSTORE)),
            (self._alg_json, os.path.join(dest_dir, CustomCertInfo.OM_ALG_JSON)),
        )
        ret = self._check_file_items_info(file_items)
        if not ret:
            run_log.error("check files failed: %s", ret.error)
            return Result(result=False, err_msg=f"check files failed: {ret.error}")

        for src, dst in file_items:
            shutil.copyfile(src, dst)
            FilePermission.set_path_permission(dst, file_mode)
            FilePermission.set_path_owner_group(dst, owner)

        return Result(result=True)


def main():
    mgr_inst = ImportCrtMgr()
    msg = "Import server cert"
    try:
        ret = mgr_inst.import_server_cert()
    except Exception:
        run_log.error("%s failed, caught exception", msg)
        mgr_inst.clear()
        return 1
    # 没有异常
    if not ret:
        run_log.error("%s failed: %s", msg, ret.error)
        mgr_inst.clear()
        return 1

    # 导入成功之后也删除该证书文件
    mgr_inst.clear()
    run_log.info("%s successfully", msg)
    return 0


if __name__ == "__main__":
    # 注册退出信号的中断处理函数
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)
    sys.exit(main())
