# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import os.path
import pwd
import shutil
import signal
import sys
from functools import partial
from itertools import islice
from pathlib import Path
from typing import Generator
from typing import NoReturn

import create_server_certs
from common.constants.base_constants import CommonConstants
from common.constants.upgrade_constants import OMUpgradeConstants
from common.constants.upgrade_constants import UpgradeConstants
from common.constants.upgrade_constants import UpgradeResult
from common.file_utils import FileConfusion
from common.file_utils import FileCopy
from common.file_utils import FileCreate
from common.file_utils import FilePermission as Chmod
from common.file_utils import FilePermission as Chown
from common.file_utils import FileUtils
from common.utils.exception_utils import OperateBaseError
from common.utils.exec_cmd import ExecCmd
from common.utils.scripts_utils import signal_handler
from common.utils.version_xml_file_manager import VersionXmlManager
from logger import upgrade_log


class UpgradeResultCode(object):
    SUCCESS_OPERATION: int = UpgradeResult.ERR_UPGRADE_SUCCEED
    FAILED_OPERATION: int = UpgradeResult.ERR_UPGRADE_OM


class UpgradeOMError(OperateBaseError):
    pass


class OMUpgradeProcessor:
    """OM升级处理类"""

    MAX_LOOP_NUM = 3000

    chmod_cmd: str = ""
    mkdir_cmd: str = ""
    sh_cmd: str = ""

    @staticmethod
    def create_config_to_db_flag():
        xml_path = os.path.join(CommonConstants.OM_WORK_DIR_PATH, UpgradeConstants.VERSION_XML_NAME)
        cur_om_version = VersionXmlManager(xml_path).version
        if "RC1" in cur_om_version or "RC2" in cur_om_version:
            with os.fdopen(os.open(OMUpgradeConstants.CONFIG_TO_DB_FLAG, os.O_CREAT, 0o400), "w"):
                pass
            upgrade_log.info("Create config file to database flag succeed")

    @staticmethod
    def set_valid_commands() -> NoReturn:
        OMUpgradeProcessor.chmod_cmd = shutil.which("chmod")
        OMUpgradeProcessor.mkdir_cmd = shutil.which("mkdir")
        OMUpgradeProcessor.sh_cmd = shutil.which("bash")

    @staticmethod
    def copy_encrypt_file() -> NoReturn:
        # 保证Nginx配置目录存在，不存在就创建
        if not os.path.exists(OMUpgradeConstants.NGINX_CONFIG_DIR):
            ret = FileCreate.create_dir(OMUpgradeConstants.NGINX_CONFIG_DIR, 0o700)
            if not ret:
                raise UpgradeOMError(f"create dir failed, {ret.error}")

        ret = Chown.set_path_owner_group(OMUpgradeConstants.NGINX_CONFIG_DIR, user=CommonConstants.NGINX_USER)
        if not ret:
            raise UpgradeOMError(f"chown {OMUpgradeConstants.NGINX_CONFIG_DIR} failed, because {ret.error}")

        # 如果存在用户配置文件om_alg.json，则保留，不使用默认的
        if not os.path.exists(os.path.join(OMUpgradeConstants.NGINX_CONFIG_DIR, OMUpgradeConstants.OM_ALG_JSON)):
            upgrade_log.info("om_alg.json not exist, recover from upgrade package")
            omg_json_path = os.path.join(CommonConstants.OM_UPGRADE_DIR_PATH, "software", "ibma", "config",
                                         "nginx_default", OMUpgradeConstants.OM_ALG_JSON)
            ret = FileCopy.copy_file(
                omg_json_path,
                os.path.join(OMUpgradeConstants.NGINX_CONFIG_DIR, OMUpgradeConstants.OM_ALG_JSON),
                user=CommonConstants.NGINX_USER,
                group=CommonConstants.NGINX_USER
            )
            if not ret:
                raise UpgradeOMError(f"copy {OMUpgradeConstants.OM_ALG_JSON} failed, because {ret.error}")

    @staticmethod
    def whitebox_process() -> NoReturn:
        # 如果有白牌标记，进行白牌化，没有则退出
        if not os.path.exists(OMUpgradeConstants.WHITEBOX_BACKUP_DIR):
            return

        # 先清空，再将OM工作目录的白牌配置拷贝到升级目录
        nginx_wb_dir = os.path.join(CommonConstants.OM_UPGRADE_DIR_PATH, OMUpgradeConstants.OM_WHITEBOX_CONFIG_PATH)
        try:
            FileUtils.delete_full_dir(nginx_wb_dir)
            shutil.copytree(OMUpgradeConstants.WHITEBOX_BACKUP_DIR, nginx_wb_dir)
        except Exception as err:
            raise UpgradeOMError(f"copy whitebox config to upgrade dir failed, {err}") from err

    @staticmethod
    def _create_uds_cert() -> NoReturn:
        # 重新生成内部uds通信证书
        cert_dir = os.path.join(CommonConstants.OM_UPGRADE_DIR_PATH, "software", "ibma", "cert")
        if not os.path.exists(cert_dir):
            ret = FileCreate.create_dir(cert_dir, 0o700)
            if not ret:
                raise UpgradeOMError(f"create dir {cert_dir}, {ret.error}")

        omg_json_path = os.path.join(
            CommonConstants.OM_UPGRADE_DIR_PATH,
            "software", "ibma", "config", "nginx_default",
            OMUpgradeConstants.OM_ALG_JSON
        )
        ret = FileCopy.copy_file(omg_json_path, os.path.join(cert_dir, OMUpgradeConstants.OM_ALG_JSON))
        if not ret:
            raise UpgradeOMError(f"copy {OMUpgradeConstants.OM_ALG_JSON}, {ret.error}")

        root = pwd.getpwnam("root")
        uid = root.pw_uid
        gid = root.pw_gid
        if not create_server_certs.create_server_certs(uid, gid, cert_dir, "force", 1):
            raise UpgradeOMError("create uds cert failed")

        # redfish 作为client需要root_ca.cert
        redfish_root_path = os.path.join(CommonConstants.OM_UPGRADE_DIR_PATH, "software", "RedfishServer", "cert")
        if not os.path.exists(redfish_root_path):
            cmd = [
                "su", "-", CommonConstants.MINDXOM_USER, "-s", OMUpgradeProcessor.sh_cmd, "-c",
                f"{OMUpgradeProcessor.mkdir_cmd} {redfish_root_path}"
            ]
            ret = ExecCmd.exec_cmd_get_output(cmd, wait=OMUpgradeConstants.EXEC_CMD_TIMEOUT)
            if ret[0] != 0:
                raise UpgradeOMError(f"su {CommonConstants.MINDXOM_USER} mkdir {redfish_root_path} failed")

            cmd = [
                "su", "-", CommonConstants.MINDXOM_USER, "-s", OMUpgradeProcessor.sh_cmd, "-c",
                f"{OMUpgradeProcessor.chmod_cmd} 700 {redfish_root_path}"
            ]
            ret = ExecCmd.exec_cmd_get_output(cmd, wait=OMUpgradeConstants.EXEC_CMD_TIMEOUT)
            if ret[0] != 0:
                raise UpgradeOMError(f"su {CommonConstants.MINDXOM_USER} chmod {redfish_root_path} failed")

        for file in [OMUpgradeConstants.OM_ALG_JSON, OMUpgradeConstants.OM_CERT_KEYSTORE,
                     OMUpgradeConstants.OM_CERT_BACKUP_KEYSTORE, OMUpgradeConstants.ROOT_CA_CERT,
                     OMUpgradeConstants.CLIENT_KMC_CERT, OMUpgradeConstants.CLIENT_KMC_PRIV,
                     OMUpgradeConstants.CLIENT_KMC_PSD]:

            ret = FileCopy.copy_file(os.path.join(cert_dir, file), os.path.join(redfish_root_path, file),
                                     mode=0o600, user=CommonConstants.MINDXOM_USER, group=CommonConstants.MINDXOM_USER)
            if not ret:
                raise UpgradeOMError(f"copy {file} failed, {ret.error}")

        # 清除不再使用的文件
        useless_files = (
            OMUpgradeConstants.CLIENT_KMC_CERT,
            OMUpgradeConstants.CLIENT_KMC_PRIV,
            OMUpgradeConstants.CLIENT_KMC_PSD,
            OMUpgradeConstants.ROOT_CA_PRIV,
            OMUpgradeConstants.ROOT_CA_PSD,
        )
        for file in useless_files:
            file_path = os.path.join(cert_dir, file)
            if not os.path.islink(file_path):
                FileConfusion.confusion_path(file_path)

            FileUtils.delete_file_or_link(file_path)

    @staticmethod
    def _copy_cert(scr_dir: str, dest_dir: str, user: str) -> NoReturn:
        if not os.path.exists(dest_dir):
            ret = FileCreate.create_dir(dest_dir, 0o700)
            if not ret:
                raise UpgradeOMError(f"create {dest_dir} failed, because {ret.error}")

        ret = Chown.set_path_owner_group(dest_dir, user)
        if not ret:
            raise UpgradeOMError(f"chown failed, {ret.error}")

        cert_list = (
            OMUpgradeConstants.OM_ALG_JSON,
            OMUpgradeConstants.OM_CERT_KEYSTORE,
            OMUpgradeConstants.OM_CERT_BACKUP_KEYSTORE,
            OMUpgradeConstants.SERVER_KMC_CERT,
            OMUpgradeConstants.SERVER_KMC_PSD,
            OMUpgradeConstants.SERVER_KMC_PRIV
        )
        for cert in cert_list:
            ret = FileCopy.copy_file(os.path.join(scr_dir, cert), os.path.join(dest_dir, cert),
                                     mode=0o600, user=user, group=user)
            if not ret:
                raise UpgradeOMError(f"copy {cert}, failed, {ret.error}")

    @staticmethod
    def _create_nginx_cert() -> NoReturn:
        # 重新生成nginx证书（只更新自己生成的nginx证书，若用户自己导入了证书就不更新）
        if not os.path.exists(OMUpgradeConstants.IMPORT_CERT_FLAG):
            nobody = pwd.getpwnam("nobody")
            uid = nobody.pw_uid
            gid = nobody.pw_gid
            if not create_server_certs.create_server_certs(uid, gid, OMUpgradeConstants.NGINX_CONFIG_DIR, "normal"):
                raise UpgradeOMError("create nginx cert failed")
        else:
            upgrade_log.info("User imports a certificate, no new certificate is generated.")

    def change_upgrade_files_permission(self) -> NoReturn:
        upgrade_dir = Path(CommonConstants.OM_UPGRADE_DIR_PATH)

        # 将整体OM升级备区目录修改为root用户，755权限
        ret = Chmod.set_path_permission(upgrade_dir.as_posix(), mode=0o755)
        if not ret:
            raise UpgradeOMError(f"chmod {upgrade_dir.name} failed, {ret.error}")
        ret = Chown.set_path_owner_group(upgrade_dir.as_posix(), CommonConstants.MONITOR_USER, recursive=True)
        if not ret:
            raise UpgradeOMError(f"chown {upgrade_dir.name} failed, {ret.error}")

        # 修改各个子目录整体权限
        args_li = (
            (("scripts",), 0o550, False, True),
            (("lib",), 0o550, True, False),
            (("bin",), 0o550, True, False),
            (("software",), 0o755, False, True),
            (("software", "ens"), 0o550, True, False),
            (("software", "sec_agent"), 0o750, False, True),
            (("config",), 0o600, True, False),
            (("tools",), 0o500, True, False),
            (("version.xml",), 0o644, False, True),
            (("software", "service_main"), 0o700, False, True)
        )
        for p_li, mode, recursive, ignore_file in args_li:
            ret = Chmod.set_path_permission(path=upgrade_dir.joinpath(*p_li).as_posix(), mode=mode,
                                            recursive=recursive, ignore_file=ignore_file)
            if not ret:
                raise UpgradeOMError(f"chmod {upgrade_dir.name} failed, {ret.error}")

        # ibma
        self._change_software_mode_recursive(upgrade_dir.joinpath("software", "ibma"), 0o750)

        # nginx
        self._change_software_mode_recursive(upgrade_dir.joinpath("software", "nginx"), 0o700)

        # ibma/config
        ret = Chmod.set_path_permission(
            path=upgrade_dir.joinpath("software", "ibma", "config").as_posix(),
            mode=0o600, recursive=True, ignore_file=False
        )
        if not ret:
            raise UpgradeOMError(f"chmod failed, {ret.error}")
        self._change_software_mode_recursive(upgrade_dir.joinpath("software", "ibma", "config"), 0o700)

        # RedfishServer
        self._change_software_mode_recursive(upgrade_dir.joinpath("software", "RedfishServer"), 0o700)

        # 修改所有文件的权限
        suffix_li = (
            ("*.sh", 0o540),
            ("*.py", 0o540),
            ("*.ini", 0o440),
            ("*.json", 0o440),
            ("*.so*", 0o540),
            ("*.service", 0o400),
            ("*.dat", 0o600),
            ("*.conf", 0o600),
            ("*.html", 0o400),
            ("*.js", 0o400),
            ("*.svg", 0o400),
            ("*.css", 0o400),
            ("*.ttf", 0o400),
            ("*.png", 0o400),
            ("*.ico", 0o400),
            ("*.gif", 0o400),
            ("*.txt", 0o400),
        )
        tmp_count = 0
        for suffix, mode in suffix_li:
            for filepath in upgrade_dir.rglob(suffix):
                tmp_count += 1
                if tmp_count > self.MAX_LOOP_NUM:
                    raise UpgradeOMError(f"The loop exceeds the maximum number of files")

                # 部分动态库文件为软连接 libcrypto.so -> libcrypto.so.1.1
                if filepath.is_symlink():
                    continue

                ret = Chmod.set_path_permission(path=filepath.as_posix(), mode=mode)
                if not ret:
                    raise UpgradeOMError(f"chmod {filepath.name} failed, {ret.error}")

        # 将seceye-agent、ensd、nginx设为可执行权限
        seceye_agent_path = upgrade_dir.joinpath("software", "sec_agent", "seceye-agent").as_posix()
        ensd_path = upgrade_dir.joinpath("software", "ens", "bin", "ensd").as_posix()
        nginx_path = upgrade_dir.joinpath("software", "nginx", "sbin", "nginx").as_posix()
        for path in (seceye_agent_path, ensd_path, nginx_path):
            ret = Chmod.set_path_permission(path, mode=0o500)
            if not ret:
                raise UpgradeOMError(f"chmod failed, {ret.error}")

        # 删除nginx默认的html文件
        html_files = upgrade_dir.joinpath("software", "nginx", "html").glob("*.html")
        for file in islice(html_files, CommonConstants.MAX_ITER_LIMIT):
            FileUtils.delete_file_or_link(file)

        # 修改属主
        ret = Chown.set_path_owner_group(
            upgrade_dir.joinpath("software", "RedfishServer").as_posix(),
            CommonConstants.MINDXOM_USER, recursive=True
        )
        if not ret:
            raise UpgradeOMError(f"chmod failed, {ret.error}")

        ret = Chown.set_path_owner_group(
            upgrade_dir.joinpath("software", "nginx").as_posix(),
            CommonConstants.NGINX_USER, recursive=True
        )
        if not ret:
            raise UpgradeOMError(f"change owner of nginx failed, {ret.error}")

    def update_cert(self) -> NoReturn:
        # 生成Nginx证书
        self._create_nginx_cert()

        # 拷贝：nginx--redfish通信证书
        # 来源：拷贝/home/data/config/default下的证书 ---> /home/data/config/redfish
        # 属主：MindXOM
        self._copy_cert(
            OMUpgradeConstants.NGINX_CONFIG_DIR,
            os.path.join(OMUpgradeConstants.CONFIG_HOME_PATH, "redfish"),
            CommonConstants.MINDXOM_USER
        )

        # 生成uds通信证书
        self._create_uds_cert()

    def tasks(self) -> Generator:
        # 设置可用的命令
        yield self.set_valid_commands

        # OM白牌化
        yield self.whitebox_process

        # 更改文件属性
        yield self.change_upgrade_files_permission

        # 拷贝加密配置文件
        yield self.copy_encrypt_file

        # 更新证书
        yield self.update_cert

        # 配置文件是否写入数据库
        yield self.create_config_to_db_flag

    def upgrade(self) -> int:
        # 执行升级
        try:
            for task in self.tasks():
                task()
                upgrade_log.info("exec %s success", task.func.__name__ if isinstance(task, partial) else task.__name__)
        except Exception as err:
            upgrade_log.error("upgrade MindXOM failed, because: %s", err)
            return UpgradeResultCode.FAILED_OPERATION

        return UpgradeResultCode.SUCCESS_OPERATION

    def _change_software_mode_recursive(self, path_obj: Path, mode: int) -> NoReturn:
        # glob("**") 返回当前目录及其下所有子目录中的【所有文件夹】
        tmp_count = 0
        for filepath in path_obj.glob("**"):
            tmp_count += 1
            if tmp_count > self.MAX_LOOP_NUM:
                raise UpgradeOMError(f"The loop exceeds the maximum number of files")

            ret = Chmod.set_path_permission(path=filepath.as_posix(), mode=mode)
            if not ret:
                raise UpgradeOMError(f"chmod {filepath.name} failed, {ret.error}")


if __name__ == "__main__":
    # 注册退出信号的中断处理函数
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)
    sys.exit(OMUpgradeProcessor().upgrade())
