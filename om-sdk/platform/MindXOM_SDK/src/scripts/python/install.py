# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
import os.path
import shutil
import signal
import sys
from functools import partial
from typing import Iterable, Callable
from typing import NoReturn, List

from common.constants.base_constants import CommonConstants
from common.constants.base_constants import MigrateOperate
from common.constants.upgrade_constants import OMUpgradeConstants
from common.file_utils import FileCheck
from common.file_utils import FileCopy
from common.file_utils import FileCreate
from common.file_utils import FilePermission
from common.file_utils import FilePermission as Chown
from common.file_utils import FileUtils
from common.utils.exception_utils import OperateBaseError
from common.utils.exec_cmd import ExecCmd
from common.utils.scripts_utils import signal_handler
from common.utils.system_utils import SystemUtils
from copy_om_sys_file import CopySysFileOperator
from logger import upgrade_log, terminal_print
from upgrade_om import OMUpgradeProcessor
from utils import OperationRetCode


class InstallError(OperateBaseError):
    pass


class OMInstaller(OMUpgradeProcessor):
    SERVICES = ("om-init", "platform-app", "start-nginx", "ibma-edge-start")
    OM_FILES = ("bin", "config", "lib", "scripts", "software", "tools", "version.xml", "uninstall.sh")

    OS_CMD_FILE_LIST = ("os_cmd_euleros2.0.conf", "os_cmd_ubuntu22.04.conf", "os_cmd_openEuler_22.03.conf")
    ROOT_DIR_PATH = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

    @staticmethod
    def get_os_name_and_version_id() -> List[str]:
        os_info = ["EulerOS", "2.0"]
        with open("/etc/os-release") as file:
            for line in file:
                segment = line.strip().split('=')
                if len(segment) != 2 or segment[0] not in ("NAME", "VERSION_ID"):
                    continue
                if segment[0] == "NAME":
                    os_info[0] = segment[1].replace('"', '')
                else:
                    os_info[1] = segment[1].replace('"', '')
        return os_info

    @staticmethod
    def judge_os_cmd_file(os_info: list, os_cmd_file_path: str) -> bool:
        ret = FileCheck.check_input_path_valid(os_cmd_file_path)
        if not ret:
            raise InstallError(f"check os_cmd_file failed, {ret.error}")
        try:
            with open(os_cmd_file_path) as file:
                line = file.readline()
                name = line.strip().split('=')[1].replace('"', '')
                line = file.readline()
                version_id = line.strip().split('=')[1].replace('"', '')
        except Exception as err:
            raise InstallError(f"Read cmd config failed, {err}") from err

        return os_info[0] == name and os_info[1] == version_id

    @staticmethod
    def create_flag_file(path: str) -> NoReturn:
        try:
            with os.fdopen(os.open(path, os.O_CREAT, 0o400), "w"):
                pass
        except Exception as err:
            raise InstallError(f"create install flag failed, {err}") from err

    @staticmethod
    def init_database() -> NoReturn:
        """安装初始化数据库"""
        # 先将redfish的日志目录初始化
        log_dir = "/var/plog/redfish"
        if not os.path.exists(log_dir):
            ret = FileCreate.create_dir(log_dir, 0o700)
            if not ret:
                raise InstallError(f"create {log_dir} failed, because {ret.error}")
            ret = Chown.set_path_owner_group(log_dir, CommonConstants.MINDXOM_USER)
            if not ret:
                raise InstallError(f"chown failed, {ret.error}")

        migrate_cmd = (CommonConstants.MIGRATE_SH, MigrateOperate.INSTALL.value)
        ret = ExecCmd.exec_cmd_get_output(migrate_cmd)
        if ret[0] != 0:
            raise InstallError(f"init database failed, {ret[1]}")

    @staticmethod
    def copy_om_config_file() -> NoReturn:
        # 创建Nginx config目录
        ret = FileCreate.create_dir(OMUpgradeConstants.NGINX_CONFIG_DIR, 0o700)
        if not ret:
            raise InstallError(f"clear nginx config dir failed, {ret.error}")
        ret = FilePermission.set_path_permission(OMUpgradeConstants.CONFIG_HOME_PATH, 0o755)
        if not ret:
            raise InstallError(f"set MindXOMUpgrade dir permission failed, {ret.error}")

        # 创建ies目录
        ret = FileCreate.create_dir(OMUpgradeConstants.IES_HOME_PATH, 0o700)
        if not ret:
            raise InstallError(f"clear ies dir failed, {ret.error}")

        # 拷贝access_control.ini，如果存在就不拷贝
        src_access_control_path = os.path.join(
            CommonConstants.OM_UPGRADE_DIR_PATH,
            "software", "ibma", "lib", "Linux", "config",
            OMUpgradeConstants.ACCESS_CONTROL_INI
        )
        dest_access_control_path = os.path.join(OMUpgradeConstants.IES_HOME_PATH, OMUpgradeConstants.ACCESS_CONTROL_INI)
        ret = FileCopy.copy_file(src_access_control_path, dest_access_control_path)
        if not ret:
            raise InstallError(f"copy access_control.ini failed, {ret.error}")

        # 拷贝mindx_om_env.conf
        om_env_path = os.path.join(CommonConstants.OM_UPGRADE_DIR_PATH, "config", OMUpgradeConstants.MINDX_OM_ENV_CONF)
        ret = FileCopy.copy_file(
            om_env_path,
            os.path.join(OMUpgradeConstants.CONFIG_HOME_PATH, OMUpgradeConstants.MINDX_OM_ENV_CONF)
        )
        if not ret:
            raise InstallError(f"copy om_env.conf failed, {ret.error}")

    @staticmethod
    def set_nginx_ip() -> NoReturn:
        # 局部导入，保证安装场景os_cmd.conf已经存在
        from lib.Linux.systems.nic import config_web_ip
        nginx_conf_path = os.path.join(CommonConstants.OM_UPGRADE_DIR_PATH, "config", OMUpgradeConstants.NGINX_CONF)
        if config_web_ip.main(["config_web_ip.py", nginx_conf_path]) != 0:
            raise InstallError("config web ip failed")

    def effect_om(self) -> NoReturn:
        # 拷贝OM日志配置、service、脚本到系统目录
        try:
            CopySysFileOperator().copy_sys_file_to_system_dir(OMUpgradeConstants.SYS_MAIN_PART,
                                                              OMUpgradeConstants.COPY_INSTALL,
                                                              CommonConstants.OM_UPGRADE_DIR_PATH)
        except Exception as err:
            raise InstallError(f"copy MindXOM files to system dir failed, {err}") from err

        # 重新加载服务
        systemctl_cmd = shutil.which("systemctl")
        reload_cmd = (systemctl_cmd, "daemon-reload")
        if ExecCmd.exec_cmd(reload_cmd) != 0:
            raise InstallError("systemctl reload failed")

        # 重启所有OM服务
        for serv in self.SERVICES:
            cmd = (systemctl_cmd, "start", serv)
            if ExecCmd.exec_cmd(cmd) != 0:
                raise InstallError(f"systemctl start {serv} failed")

    def check_install_env(self) -> NoReturn:
        # 校验当前安装包所在的操作路径
        ret = FileCheck.check_input_path_valid(self.ROOT_DIR_PATH)
        if not ret:
            raise InstallError(f"check file path {self.ROOT_DIR_PATH} failed, {ret.error}")

        # 检查安装目录权限、是否为root属组
        ret = FileCheck.check_path_is_root(self.ROOT_DIR_PATH)
        if not ret:
            raise InstallError(f"install path {self.ROOT_DIR_PATH} is invalid, please check, {ret.error}")

        # 检查文件系统目录是否完整，没有则创建
        sys_dirs = (
            (CommonConstants.OM_HOME_PATH, 0o755),
            (CommonConstants.OM_LOG_HOME_PATH, 0o755),
            (CommonConstants.OM_LOG_IBMA_PATH, 0o750),
            (CommonConstants.LOCAL_CRACKLIB_DIR, 0o755),
            (CommonConstants.LOCAL_SCRIPT_DIR, 0o550),
            (CommonConstants.HOME_DATA_PATH, 0o755)
        )

        for (dir_path, mode) in sys_dirs:
            if not os.path.exists(dir_path):
                os.mkdir(dir_path, mode=mode)
            if os.path.islink(dir_path):
                raise InstallError(f"path {dir_path} is soft link !")
            os.chmod(dir_path, mode=mode)

        # 删除Upgrade目录
        try:
            FileUtils.delete_full_dir(CommonConstants.OM_UPGRADE_DIR_PATH)
        except Exception as err:
            raise InstallError(f"clear Upgrade dir failed, {err}") from err

        # 检查OM安装目录磁盘剩余空间是否充足
        if SystemUtils.get_available_size(CommonConstants.OM_HOME_PATH) < OMUpgradeConstants.OM_SOFTWARE_SPACE_BYTES:
            raise InstallError(f"Install space {CommonConstants.OM_HOME_PATH} is insufficient, please clean it first")

        # 检查OM配置目录磁盘剩余空间是否充足
        if SystemUtils.get_available_size(CommonConstants.HOME_DATA_PATH) < OMUpgradeConstants.OM_CONFIG_SPACE_BYTES:
            raise InstallError(f"Config space {CommonConstants.HOME_DATA_PATH} is insufficient, please clean first")

        # 检查OM家目录权限、是否为root属组
        ret = FileCheck.check_path_is_root(CommonConstants.OM_HOME_PATH)
        if not ret:
            raise InstallError(f"install path {CommonConstants.OM_HOME_PATH} is invalid, please check, {ret.error}")

        # 检查日志目录权限、是否为root属组
        ret = FileCheck.check_path_is_root(CommonConstants.OM_LOG_HOME_PATH)
        if not ret:
            raise InstallError(f"install path {CommonConstants.OM_LOG_HOME_PATH} is invalid, please check, {ret.error}")

    def copy_files_to_upgrade_dir(self) -> NoReturn:
        # 创建OM备区目录，指定为755权限，升级目录需要非root进程操作
        ret = FileCreate.create_dir(CommonConstants.OM_UPGRADE_DIR_PATH, 0o755)
        if not ret:
            raise InstallError(f"creat MindXOMUpgrade dir and set mode failed, {ret.error}")

        # 只拷贝指定的目录和文件
        try:
            for file in self.OM_FILES:
                src_path = os.path.join(self.ROOT_DIR_PATH, file)
                dest_path = os.path.join(CommonConstants.OM_UPGRADE_DIR_PATH, file)
                if os.path.isdir(src_path):
                    shutil.copytree(src_path, dest_path, symlinks=True)
                else:
                    shutil.copyfile(src_path, dest_path)
        except Exception as err:
            raise InstallError(f"copy file to upgrade dir failed, {err}") from err

    def replace_os_cmd_conf(self) -> NoReturn:
        om_config_dir = os.path.join(self.ROOT_DIR_PATH, "config")
        if not os.path.exists(OMUpgradeConstants.CONFIG_HOME_PATH):
            ret = FileCreate.create_dir(OMUpgradeConstants.CONFIG_HOME_PATH, mode=0o755)
            if not ret:
                raise InstallError(f"replace os_cmd_conf failed, {ret.error}")

        os_cmd_conf_path = os.path.join(OMUpgradeConstants.CONFIG_HOME_PATH, OMUpgradeConstants.OS_CMD_CONF)
        os_info = self.get_os_name_and_version_id()
        for item in self.OS_CMD_FILE_LIST:
            file_path = os.path.join(om_config_dir, item)
            if self.judge_os_cmd_file(os_info, file_path):
                ret = FileCopy.copy_file(file_path, os_cmd_conf_path)
                if not ret:
                    raise InstallError(f"replace os_cmd_conf failed, {ret.error}")

                ret = FilePermission.set_path_permission(os_cmd_conf_path, 0o644)
                if not ret:
                    raise InstallError(f"set permission failed, {ret.error}")
                return

        raise InstallError("replace os_cmd_conf failed, not match!")

    def tasks(self) -> Iterable[Callable]:
        # 安装环境检查
        yield self.check_install_env
        terminal_print.info("check install environment success")

        # 拷贝os_cmd.conf
        yield self.replace_os_cmd_conf

        # 设置可用的命令
        yield self.set_valid_commands

        # 将OM所有文件拷贝到Upgrade升级目录下
        yield self.copy_files_to_upgrade_dir

        # 拷贝OM的配置文件
        yield self.copy_om_config_file
        terminal_print.info("prepare service file success")

        # OM白牌化
        yield self.whitebox_process

        # 更改文件属性
        yield self.change_upgrade_files_permission

        # 拷贝加密配置文件
        yield self.copy_encrypt_file

        # 更新证书
        yield self.update_cert

        # 初始化数据库
        yield self.init_database

        # 设置nginx监听IP
        yield self.set_nginx_ip

        # 设置OM升级成功标记
        yield partial(self.create_flag_file, OMUpgradeConstants.UPGRADE_FINISHED_FLAG)
        terminal_print.info("executing install success")

        # 生效OM服务
        yield self.effect_om
        terminal_print.info("start service success")

    def install(self) -> int:
        # 执行安装
        try:
            for task in self.tasks():
                task()
                upgrade_log.info("exec %s success", task.func.__name__ if isinstance(task, partial) else task.__name__)
        except Exception as err:
            upgrade_log.error("Install MindXOM failed, because err: %s.", err)
            return OperationRetCode.FAILED_OPERATION

        upgrade_log.info("Install MindXOM success.")
        return OperationRetCode.SUCCESS_OPERATION


if __name__ == "__main__":
    # 注册退出信号的中断处理函数
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)
    try:
        sys.exit(OMInstaller().install())
    except Exception as error:
        upgrade_log.error("Install MindXOM failed, because unknown error: %s", error)
        terminal_print.error("Install MindXOM failed, please check [/var/plog/upgrade.log] to see the specific reason.")
        sys.exit(OperationRetCode.FAILED_OPERATION)
