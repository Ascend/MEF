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
import ctypes
import os
import shutil
import signal
import sys

from common.constants.base_constants import CertInfo
from common.file_utils import FileCheck, FilePermission
from common.file_utils import FileCopy
from common.file_utils import FileCreate
from common.log.logger import run_log
from common.utils.scripts_utils import signal_handler
from common.kmc_lib.kmc import Kmc
from common.utils.result_base import Result
from lib.Linux.systems.security_service.custom_cert_info import CustomCertInfo
from lib.Linux.systems.security_service.security_service_clib import cert_manage_finalize
from lib.Linux.systems.security_service.security_service_clib import cert_manage_init
from lib.Linux.systems.security_service.security_service_clib import create_cert_sign_request
from common.common_methods import CommonMethods


class CustomCrtMgr:

    def __init__(self) -> None:
        self._file_items_to_remove = tuple()

    def clear(self):
        all_success = True
        for file in self._file_items_to_remove:
            if not os.path.exists(file):
                continue

            ret = FileCopy.remove_path(file)
            if not ret:
                run_log.error("remove path %s failed: %s", file, ret.error)
                all_success = False
                continue

            run_log.info("remove %s successfully", file)

        if not all_success:
            return Result(result=False, err_msg="not all successfully")

        return Result(result=True)


class ExportCrtMgr(CustomCrtMgr):

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
            self._server_psd, self._server_pri, self._server_csr
        )
        self._file_items_to_remove = self._necessary_files + (self._server_cert,)
        self._kmc = Kmc(self._kmc_primary_ksf, self._kmc_standby_ksf, self._alg_json)

    @staticmethod
    def _copy_om_alg_json():
        src_om_alg = os.path.join(CustomCertInfo.CERT_WORK_DIR, CustomCertInfo.OM_ALG_JSON)
        ret = FileCheck.check_path_is_exist_and_valid(src_om_alg)
        if not ret:
            return Result(result=False, err_msg=f"{src_om_alg} invalid: {ret.error}")

        if os.path.getsize(src_om_alg) > CustomCertInfo.FILE_MAX_SIZE:
            return Result(result=False, err_msg=f"{src_om_alg} invalid: too large")

        dir_path = CustomCertInfo.EXPORT_TMP_DIR
        if not os.path.exists(dir_path):
            ret = FileCreate.create_dir(dir_path, CustomCertInfo.DEFAULT_DIR_MODE)
            if not ret:
                return Result(result=False, err_msg=f"create dir {dir_path} failed: {ret.error}")

        dst_om_alg = os.path.join(dir_path, CustomCertInfo.OM_ALG_JSON)
        ret = FileCheck.check_input_path_valid(dst_om_alg)
        if not ret:
            return Result(result=False, err_msg=f"{dst_om_alg} invalid: {ret.error}")

        if os.path.exists(dst_om_alg):
            ret = FileCopy.remove_path(dst_om_alg)
            if not ret:
                return Result(result=False, err_msg=f"remove exist file failed: {ret.error}")
            os.mknod(dst_om_alg, mode=CustomCertInfo.DEFAULT_FILE_MODE)

        shutil.copyfile(src_om_alg, dst_om_alg)
        FilePermission.set_path_permission(dst_om_alg, CustomCertInfo.DEFAULT_FILE_MODE)
        return Result(result=True)

    def create_csr_file(self):
        """
        创建证书签名请求文件
        步骤1.生成公私钥对、生成CSR、生成私钥的随机口令
        步骤2.加密私钥，加密口令
        步骤3.保存csr、私钥、口令
        """
        clear_ret = self.clear()
        if not clear_ret:
            run_log.error("remove old files failed: %s", clear_ret.error)
            return Result(result=False, err_msg=clear_ret.error)

        return self._create_csr()

    def _create_csr(self):
        if not self._kmc.available():
            run_log.error("kmc is unavailable")
            return Result(False, err_msg="kmc is unavailable")

        cert_path = CustomCertInfo.EXPORT_TMP_DIR
        ret = FileCopy.remove_path(cert_path)
        if not ret:
            run_log.error("remove path %s failed: %s", cert_path, ret.error)
            return ret

        ret = FileCreate.create_dir(cert_path, CustomCertInfo.DEFAULT_DIR_MODE)
        if not ret:
            return Result(result=False, err_msg=f"create dir {cert_path} failed: {ret.error}")

        primary_ksf_p = ctypes.create_string_buffer(self._kmc_primary_ksf.encode("utf-8"))
        standby_ksf_p = ctypes.create_string_buffer(self._kmc_standby_ksf.encode("utf-8"))
        alg_json_p = ctypes.create_string_buffer(self._alg_json.encode("utf-8"))
        ret = cert_manage_init(primary_ksf_p, standby_ksf_p, alg_json_p)
        if ret != 0:
            run_log.error("init kmc failed, ret=%s", ret)
            return Result(False, err_msg=f"init kmc failed, ret={ret}")

        cert_path_p = ctypes.create_string_buffer(cert_path.encode("utf-8"))
        domain_name_str = CertInfo.gen_domain_name_with_uuid()
        domain_name = ctypes.create_string_buffer(domain_name_str.encode(encoding="utf-8"))
        ret = create_cert_sign_request(cert_path_p, domain_name)
        if ret != 0:
            run_log.error("create cert sign request file failed, ret=%s", ret)
            return Result(False, err_msg=f"create cert sign request file failed, ret={ret}")

        cert_manage_finalize()

        ret = self._copy_om_alg_json()
        if not ret:
            return Result(result=False, err_msg=f"copy om alg json file failed: {ret.error}")

        ret = FileCreate.create_dir(CustomCertInfo.IMPORT_TMP_DIR, CustomCertInfo.WORK_DIR_MODE)
        if not ret:
            return Result(result=False, err_msg=f"create {CustomCertInfo.IMPORT_TMP_DIR} directory failed: {ret.error}")

        csr_file = os.path.join(CustomCertInfo.EXPORT_TMP_DIR, "x509Req.pem")
        ret = FileCheck.check_path_is_exist_and_valid(csr_file)
        if not ret:
            return Result(result=False, err_msg=f"csr {csr_file} file invalid: {ret.error}")

        shutil.move(csr_file, self._server_csr)
        FilePermission.set_path_permission(self._server_csr, CustomCertInfo.DEFAULT_FILE_MODE)
        return Result(True)


def main():
    mgr_inst = ExportCrtMgr()
    msg = "create cert sign req"
    try:
        ret = mgr_inst.create_csr_file()
    except Exception:
        run_log.error("%s failed, caught exception", msg)
        mgr_inst.clear()
        return 1
    # 没有异常
    if not ret:
        run_log.error("%s failed: %s", msg, ret.error)
        mgr_inst.clear()
        return 1

    run_log.info("%s successfully", msg)
    return 0


if __name__ == "__main__":
    # 注册退出信号的中断处理函数
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)
    sys.exit(main())
