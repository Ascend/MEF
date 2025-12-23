# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import glob
import hashlib
import os
import shutil
import subprocess
import time
from itertools import islice
from pathlib import Path

from bin.environ import Env
from bin.global_exclusive_control import GlobalExclusiveController
from common.constants.base_constants import CommonConstants
from common.constants.base_constants import RecoverMiniOSConstants
from common.constants.upgrade_constants import OMUpgradeConstants
from common.constants.upgrade_constants import UpgradeConstants
from common.file_utils import FileCheck
from common.file_utils import FileCopy
from common.file_utils import FileCreate
from common.file_utils import FileUtils
from common.init_cmd import cmd_constants
from common.log.logger import run_log
from common.utils.exec_cmd import ExecCmd
from common.utils.version_xml_file_manager import XmlManger
from common.verify_cms_file import verify_cms_file
from common.checkers import IPChecker
from lib.Linux.mef.mef_info import MefInfo
from common.common_methods import CommonMethods


class RecoverOSError(Exception):
    pass


class RecoverMiniOS:
    TAR_FILES = ("Ascend-mindx-toolbox_*.tar.gz", "Ascend-cann-nnrt_*.tar.gz",
                 "Ascend-om_*_linux-aarch64.tar.gz", "Ascend-mefedge_*_linux-aarch64.tar.gz")
    FILE_MAX_SIZE_BYTES = 300 * 1024 * 1024

    RECOVER_MINI_OS_LOCK = GlobalExclusiveController()
    LOCK_TIMEOUT = 180

    @staticmethod
    def mount_golden_dev():
        p1_mount_path = Path("/mnt/p1")
        if p1_mount_path.is_mount():
            run_log.info("P1 is already mounted.")
            return

        dev_p1 = "/dev/mmcblk0p1"
        cmd = (cmd_constants.OS_CMD_MOUNT, dev_p1, p1_mount_path.as_posix())
        if ExecCmd.exec_cmd(cmd) not in (0, 32):
            run_log.error("Exec cmd : %s failed", cmd)
            raise RecoverOSError("Mount P1 failed")
        run_log.info("Mount p1 success.")

    @staticmethod
    def umount_golden_dev():
        if not Path("/mnt/p1").is_mount():
            run_log.info("P1 is not mounted.")
            return

        dev_p1 = "/dev/mmcblk0p1"
        cmd = (cmd_constants.OS_CMD_UMOUNT, dev_p1)
        if ExecCmd.exec_cmd(cmd) != 0:
            run_log.error("Exec cmd : %s failed", cmd)
            raise RecoverOSError("Umount P1 failed")
        run_log.info("Umount p1 success.")

    @staticmethod
    def _verify_vercfg_validity():
        ver_cfg_path = os.path.join(RecoverMiniOSConstants.FIRMWARE_PATH, UpgradeConstants.VERCFG_XML_NAME)
        ver_cms_path = os.path.join(RecoverMiniOSConstants.FIRMWARE_PATH, UpgradeConstants.VERCFG_CMS_NAME)
        ver_crl_path = os.path.join(RecoverMiniOSConstants.FIRMWARE_PATH, UpgradeConstants.VERCFG_CRL_NAME)

        for file_path in (ver_cfg_path, ver_cms_path, ver_crl_path):
            ret = FileCheck.check_path_is_exist_and_valid(file_path)
            if not ret:
                raise RecoverOSError(f"{os.path.basename(file_path)} not exist.")

        # 签名校验动态库 libverify.so
        lib_path = os.path.join(CommonConstants.OM_WORK_DIR_PATH, "lib", UpgradeConstants.VERIFY_LIB)
        if not verify_cms_file(lib_path, ver_cms_path, ver_crl_path, ver_cfg_path):
            raise RecoverOSError("cms verify failed.")

    @staticmethod
    def _clean_files():
        reset_shell_path = os.path.join(RecoverMiniOSConstants.FIRMWARE_PATH, "reset_om.sh")
        whitebox_path = os.path.join(RecoverMiniOSConstants.PACKAGE_PATH, "whiteboxConfig")
        om_certs_path = "/home/package/om_certs"
        restore_log_path = os.path.join(RecoverMiniOSConstants.PACKAGE_PATH, "om_restore_msg.log")
        file_list = [reset_shell_path, whitebox_path, om_certs_path, restore_log_path]

        for file in file_list:
            FileCopy.remove_path(file)

    @staticmethod
    def _compute_sha256(file_path):
        """
        传入文件路径计算 sha256 值，在这个场景无需校验文件路径合法性、大小
        :param file_path: 需要计算sha256的文件路径
        :return: 文件的sha256值
        """
        sha256 = hashlib.sha256()
        with open(file_path, "rb") as file:
            sha256.update(file.read())
            return sha256.hexdigest()

    @staticmethod
    def _copy_pki_om_certs():
        """
        从golden拷贝从PKI系统签发的证书到p7分区
        """
        # 如果设备执行过白牌，不拷贝pki证书
        if os.path.isdir(OMUpgradeConstants.WHITEBOX_BACKUP_DIR):
            run_log.info("The device has been white-labeled, not copy the pki certificate.")
            return

        src_path = "/mnt/p1/om_certs"
        if not os.path.exists(src_path):
            raise RecoverOSError(f"path not exists in p1: {src_path}")

        dest_path = "/home/package/om_certs"
        if not os.path.exists(dest_path):
            ret = FileCreate.create_dir(dest_path, mode=0o700)
            if not ret:
                raise RecoverOSError(f"create {dest_path} failed: {ret.error}")

        cert_files = (
            "om_cert.keystore",
            "om_cert_backup.keystore",
            "server_kmc.cert",
            "server_kmc.priv",
            "server_kmc.psd",
            "om_alg.json",
        )
        for file in cert_files:
            src_file = os.path.join(src_path, file)
            dest_file = os.path.join(dest_path, file)
            ret = FileCopy.copy_file(src_file, dest_file, mode=0o600)
            if not ret:
                raise RecoverOSError(f"copy file {file} from p1 failed: {ret.error}")

    @staticmethod
    def _copy_reset_om():
        """
        拷贝 reset_om.sh 脚本到 P7分区
        """
        src_reset_shell = os.path.join(CommonConstants.OM_WORK_DIR_PATH, "scripts", "reset_om.sh")
        dest_reset_shell = os.path.join(RecoverMiniOSConstants.FIRMWARE_PATH, "reset_om.sh")
        if not os.path.exists(src_reset_shell):
            raise RecoverOSError("reset_om.sh does not exist.")

        try:
            FileCopy.copy_file(src_reset_shell, dest_reset_shell, 0o500)
        except Exception:
            raise RecoverOSError("copy reset_om.sh failed")

        run_log.info("copy reset_om.sh success")

    @staticmethod
    def _copy_reset_middleware():
        """
        调用MEF脚本拷贝 reset_middleware.sh
        """
        copy_shell = f"{MefInfo().install_path}/edge_installer/script/copy_reset_script.sh"
        if not os.path.exists(copy_shell):
            raise RecoverOSError(f"{os.path.basename(copy_shell)} does not exist.")

        cmd = ["bash", copy_shell]
        if ExecCmd.exec_cmd(cmd) != 0:
            run_log.error("Exec %s failed", os.path.basename(copy_shell))
            raise RecoverOSError(f"Exec {os.path.basename(copy_shell)} failed")
        run_log.info("Exec %s success", os.path.basename(copy_shell))

    @staticmethod
    def _copy_component_from_golden(component_tar_pattern):
        golden_path = "/mnt/p1"

        if not os.path.exists(RecoverMiniOSConstants.FIRMWARE_PATH):
            ret = FileCreate.create_dir(RecoverMiniOSConstants.FIRMWARE_PATH, mode=0o700)
            if not ret:
                raise RecoverOSError(f"create path {RecoverMiniOSConstants.FIRMWARE_PATH} failed: {ret.error}")

        for component_tar in islice(glob.iglob(os.path.join(golden_path, component_tar_pattern)),
                                    CommonConstants.MAX_ITER_LIMIT):
            dest_component_tar = os.path.join(RecoverMiniOSConstants.FIRMWARE_PATH, os.path.basename(component_tar))
            try:
                shutil.copyfile(component_tar, dest_component_tar)
                run_log.info("copy %s package to recover mini os path success.", component_tar_pattern)
                return
            except Exception:
                raise RecoverOSError(f"copy {os.path.basename(component_tar)} package to recover mini os path failed")

        raise RecoverOSError(f"golden path does not have {component_tar_pattern} package file.")

    @staticmethod
    def _check_whitebox_and_copy():
        """
        检查是否有白牌化。如果有白牌化，则拷贝白牌配置文件到P7分区。
        """
        if os.path.exists(OMUpgradeConstants.WHITEBOX_BACKUP_DIR):
            # 先清空，再将OM工作目录的白牌配置拷贝到P7 /home/package/firmware/whiteboxConfig 目录
            whitebox_path = os.path.join(RecoverMiniOSConstants.PACKAGE_PATH, "whiteboxConfig")
            try:
                FileUtils.delete_full_dir(whitebox_path)
                shutil.copytree(OMUpgradeConstants.WHITEBOX_BACKUP_DIR, whitebox_path)
            except Exception:
                run_log.error("Copy whitebox config to P7 dir failed.")
                raise RecoverOSError("Copy whitebox config to P7 dir failed.")

            run_log.info("Restore whitebox config to recover mini os path success.")

    @staticmethod
    def _restore_log(fd_ip):
        # 记录日志
        restore_log_path = os.path.join(RecoverMiniOSConstants.PACKAGE_PATH, "om_restore_msg.log")
        log_msg = f'[{time.strftime("%Y-%m-%d %H:%M:%S")}] [FD@{fd_ip}] Recover mini os success.\n'
        try:
            with os.fdopen(os.open(restore_log_path, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600), "w") as file:
                file.write(log_msg)
        except Exception as err:
            raise RecoverOSError("Write restore msg failed.")

    @staticmethod
    def _recover_system():
        """
        调用驱动接口
        """
        cmd_shell = f"{cmd_constants.OS_CMD_IES_TOOL} restore_minios"
        try:
            with subprocess.Popen(cmd_shell.split(), stdin=subprocess.PIPE, stdout=subprocess.PIPE, shell=False) as s:
                s.stdin.write(b"Y\n")
                s.stdin.flush()
                msg_str = ",".join(str(msg.strip()) for msg in s.stdout.readlines())
                s.communicate(timeout=600)
        except Exception as ex:
            run_log.error("Execute restore_minios has error: %s.", ex)
            raise RecoverOSError("Execute restore_minios failed.")

        if "The system will be automatically reset" in msg_str:
            # 开始执行恢复出厂
            run_log.info("Execute restore_minios success")
            return

        raise RecoverOSError("Execute ies_tool restore_minios failed.")

    def post_request(self, payload):
        if self.RECOVER_MINI_OS_LOCK.locked():
            run_log.error("Recover mini OS process is busy.")
            return [CommonMethods.ERROR, "Recover mini OS process is busy."]
        self.RECOVER_MINI_OS_LOCK.acquire(self.LOCK_TIMEOUT)

        if not isinstance(payload, dict):
            run_log.error("Recover mini os post request param [payload] is invalid")
            self.RECOVER_MINI_OS_LOCK.release()
            return [CommonMethods.ERROR, "Recover mini os post request param [payload] is invalid"]

        request_ip = payload.get("fd_ip")
        check_result = IPChecker("data").check({"data": request_ip})
        if not check_result.success:
            run_log.error("Recover mini os post request param [fd_ip] is invalid")
            self.RECOVER_MINI_OS_LOCK.release()
            return [CommonMethods.ERROR, "Recover mini os post request param [fd_ip] is invalid"]

        if Env().start_from_m2:
            run_log.info("Recover mini OS not support on M.2")
            self.RECOVER_MINI_OS_LOCK.release()
            return [CommonMethods.ERROR, "Recover mini OS not support on M.2"]

        # 当前已升级但未生效不能恢复最小系统
        if os.path.exists(OMUpgradeConstants.UPGRADE_FINISHED_FLAG):
            run_log.error("The current device is not in effect, not allow to recover mini OS.")
            self.RECOVER_MINI_OS_LOCK.release()
            return [CommonMethods.ERROR, "The current device is not in effect, not allow to recover mini OS."]

        # 没有升过级，从P1分区拷贝
        if self._has_no_tar_file():
            run_log.info("Recover mini os path does not have packages")
            try:
                self._copy_from_golden()
            except RecoverOSError as err:
                run_log.error("Copy from golden failed, because %s", err)
                self.RECOVER_MINI_OS_LOCK.release()
                return [CommonMethods.ERROR, "Copy from golden failed"]
        # 升过级，校验vercfg.xml签名和tar包完整性
        else:
            try:
                self._verify_vercfg_validity()
                self._verify_firmware_sha256()
            except RecoverOSError as err:
                run_log.error("Verify recover mini OS file failed, because %s", err)
                self.RECOVER_MINI_OS_LOCK.release()
                return [CommonMethods.ERROR, "Verify recover mini OS file failed"]

        try:
            self.mount_golden_dev()
            self._copy_pki_om_certs()
        except RecoverOSError as err:
            run_log.error("Recover mini OS failed, because %s", err)
            self._clean_files()
            self.umount_golden_dev()
            self.RECOVER_MINI_OS_LOCK.release()
            return [CommonMethods.ERROR, err]

        try:
            self.umount_golden_dev()
            self._copy_reset_om()
            self._copy_reset_middleware()
            self._check_whitebox_and_copy()
            self._restore_log(request_ip)
            self._recover_system()
        except RecoverOSError as err:
            run_log.error("Prepare for recover mini OS failed, because %s", err)
            self._clean_files()
            return [CommonMethods.ERROR, err]
        except Exception as err:
            run_log.error("Prepare for recover mini OS failed, because %s", err)
            self._clean_files()
            return [CommonMethods.INTERNAL_ERROR, "Internal error."]
        finally:
            self.RECOVER_MINI_OS_LOCK.release()
        return [CommonMethods.OK, "Prepare for recover mini os successfully."]

    def _copy_from_golden(self):
        """
        P1分区拷贝包到P7
        """
        for component in self.TAR_FILES:
            for tar in glob.iglob(os.path.join(RecoverMiniOSConstants.FIRMWARE_PATH, component)):
                FileCopy.remove_path(tar)

        try:
            self.mount_golden_dev()
            for component in self.TAR_FILES:
                self._copy_component_from_golden(component)
        except RecoverOSError as err:
            raise err
        finally:
            self.umount_golden_dev()

    def _has_no_tar_file(self):
        return not all(
            glob.glob(os.path.join(RecoverMiniOSConstants.FIRMWARE_PATH, component)) for component in self.TAR_FILES
        )

    def _verify_firmware_sha256(self):
        vercfg_path = os.path.join(RecoverMiniOSConstants.FIRMWARE_PATH, UpgradeConstants.VERCFG_XML_NAME)
        root = XmlManger(vercfg_path).xml_tree
        for module in root.iterfind("File"):
            tar_name = os.path.basename(module.find("FilePath").text)
            if not tar_name.endswith(".tar.gz") or "os" in tar_name:
                continue
            tar_path = os.path.join(RecoverMiniOSConstants.FIRMWARE_PATH, tar_name)
            if os.path.getsize(tar_path) > self.FILE_MAX_SIZE_BYTES:
                raise RecoverOSError(f"The size of {tar_name} is too large.")
            file_sha256 = module.find("SHAValue").text
            cal_sha256 = self._compute_sha256(tar_path)
            if file_sha256 != cal_sha256:
                raise RecoverOSError(f"{tar_name} sha256 verify failed")

