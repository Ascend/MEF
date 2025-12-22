# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import dataclasses
import os
import threading
import time
import zipfile
from collections import defaultdict
from typing import Any
from typing import Callable
from typing import Dict
from typing import List
from typing import NoReturn
from typing import Union

from bin.environ import Env
from bin.global_exclusive_control import GlobalExclusiveController
from common.constants.base_constants import CommonConstants, MefBusyConstants
from common.constants.base_constants import RecoverMiniOSConstants
from common.constants.upgrade_constants import OMUpgradeConstants
from common.constants.upgrade_constants import UpgradeConstants
from common.constants.upgrade_constants import UpgradeResult
from common.constants.upgrade_constants import UpgradeState
from common.file_utils import FileCheck
from common.file_utils import FileCopy as CopyFile
from common.file_utils import FileCreate
from common.file_utils import FileOperator
from common.file_utils import FilePermission
from common.file_utils import FileUtils
from common.log.logger import operate_log
from common.log.logger import run_log
from common.utils.app_common_method import AppCommonMethod
from common.utils.common_check import CommonCheck
from common.utils.exception_utils import OperateBaseError
from common.utils.exec_cmd import ExecCmd
from common.utils.singleton import Singleton
from common.utils.version_xml_file_manager import VercfgManager
from common.utils.version_xml_file_manager import VersionXmlManager
from common.verify_cms_file import verify_cms_file
from lib.Linux.mef.mef_info import MefInfo
from lib.Linux.upgrade.module_type import ModuleType
from lib.Linux.upgrade.schemas import ModuleState
from common.common_methods import CommonMethods


class TarGzPaths:
    def __init__(self, driver_tar_gz=None, om_tar_gz=None, mef_tar_gz=None, toolbox_tar_gz=None, cann_tar_gz=""):
        self.driver_tar_gz = driver_tar_gz
        self.om_tar_gz = om_tar_gz
        self.mef_tar_gz = mef_tar_gz
        self.toolbox_tar_gz = toolbox_tar_gz
        self.cann_tar_gz = cann_tar_gz


class UpgradeError(OperateBaseError):
    pass


class Upgrade(Singleton):
    # 起始时间
    upgrade_start_time = ""

    # 状态值为0表示空闲，1表示正在升级，2表示升级失败，3表示升级成功
    upgrade_state: int = UpgradeState.UPGRADE_NO_STATE

    # 用字符串表示状态 "Running" ，"New"，"Failed",Success
    upgrade_proc_msg: str = UpgradeState.UPGRADE_STATE_MSG.get(upgrade_state)

    # 保存当前正在升级的固件版本信息，每次上传升级包升级时更新
    cur_module: ModuleState = ModuleState()

    # 当前已进行升级的固件版本信息，保存已升级的固件名称、升级的版本、已经升级结果
    modules: Dict[str, ModuleState] = {m: ModuleState(name=m) for m in ModuleType.values()}

    # 保存升级的结果，默认状态为-1
    upgrade_result: int = UpgradeResult.ERR_NO_UPGRADE

    # 升级包所有需要升级的组件
    upgrade_components_list: List[str] = []

    username = ""
    request_ip = ""

    # 请求并发处理
    upgrade_mutex = GlobalExclusiveController

    def __init__(self):
        self.StartTime: str = ""
        self.TaskState: str = ""
        self.Messages: Dict[str, str] = {}
        self.Id: str = "1"
        self.Name: str = "Upgrade Task"
        self.Version: str = ""
        self.PercentComplete: int = 0
        self.Module: str = ""

    @staticmethod
    def get_drv_upgrade_process() -> int:
        if not os.path.exists(UpgradeConstants.PROGRESS_FILE):
            run_log.warning("drv process file not exist! ")
            return Upgrade.cur_module.process
        if os.path.getsize(UpgradeConstants.PROGRESS_FILE) > CommonConstants.OM_READ_FILE_MAX_SIZE_BYTES:
            run_log.error("upgrade process file failed, because it is too large")
            return Upgrade.cur_module.process
        try:
            with open(UpgradeConstants.PROGRESS_FILE) as file:
                process = int(file.readline().strip("\n"))
        except Exception as err:
            run_log.error("Read upgrade process file failed, %s", err)
            return Upgrade.cur_module.process
        run_log.info("get drv process: %s", process)
        return process

    @staticmethod
    def verify_input_param(request_dic: dict) -> str:
        if not isinstance(request_dic, dict):
            raise UpgradeError("The type of request data is not dict")

        # 校验用户名和IP是否合法
        ret = CommonCheck.check_operator(request_dic.get("_User"), request_dic.get("_Xip"))
        if not ret:
            raise UpgradeError(f"The User or Xip is illegal, {ret.error}")

        Upgrade.username = request_dic.get("_User")
        Upgrade.request_ip = request_dic.get("_Xip")
        package_name = request_dic.get("ImageURI")

        # 校验使用的协议是否在规定的协议范围内
        if request_dic.get("TransferProtocol") not in ("https",):
            raise UpgradeError("Request data protocol is invalid")

        if not package_name.endswith(".zip"):
            raise UpgradeError("Upgrade file type needs to be zip")

        # 校验升级包路径是否合法
        ret = FileCheck.check_path_is_exist_and_valid(package_name)
        if not ret:
            raise UpgradeError(f"The upgrade package  path check failed, {ret.error}")

        return package_name

    @staticmethod
    def remove_and_create_dir(dest_path: str, mode: int) -> NoReturn:
        try:
            FileUtils.delete_full_dir(dest_path)
        except Exception as err:
            raise UpgradeError(f"clear dir {dest_path} failed, {err}") from err

        ret = FileCreate.create_dir(dest_path, mode)
        if not ret:
            raise UpgradeError(f"Create upgrade dir {dest_path} failed, {ret.error}.")

    @staticmethod
    def cms_verify(file: str) -> NoReturn:
        file_cms = f"{file}.cms"
        file_crl = f"{file}.crl"

        lib_path = os.path.join(CommonConstants.OM_WORK_DIR_PATH, "lib", UpgradeConstants.VERIFY_LIB)
        if not verify_cms_file(lib_path, file_cms, file_crl, file):
            Upgrade.upgrade_result = UpgradeResult.ERR_UPGRADE_VERIFY_FAILED
            raise UpgradeError(f"cms verify failed")

    @staticmethod
    def exec_shell(shell_cmd) -> NoReturn:
        """
        执行升级脚本
        :param shell_cmd: 执行升级的cmd命令
        :return: NA
        """
        ret = FileUtils.check_script_file_valid(shell_cmd[0], "root", "root")
        if not ret:
            raise UpgradeError(f"check shell invalid, {ret.error}")

        ret = ExecCmd.exec_cmd(shell_cmd, UpgradeConstants.UPGRADE_FIRMWARE_TIMEOUT)
        if ret != 0:
            # 升级失败，记录升级错误码
            Upgrade.upgrade_result = ret
            raise UpgradeError(f"upgrade failed, error code is {ret}")

    @staticmethod
    def create_flag_file(path: str) -> NoReturn:
        try:
            with os.fdopen(os.open(path, os.O_CREAT, 0o400), "w"):
                pass
        except Exception as err:
            raise UpgradeError(f"create MindXOM flag failed, {err}") from err

    @staticmethod
    def _copy_tar_to_recover_mini_os_path(recover_os_files):
        if not os.path.exists(RecoverMiniOSConstants.FIRMWARE_PATH):
            os.mkdir(RecoverMiniOSConstants.FIRMWARE_PATH, mode=0o700)

        firmware_path = os.path.join(RecoverMiniOSConstants.FIRMWARE_PATH, RecoverMiniOSConstants.FIRMWARE_NAME)
        with zipfile.ZipFile(firmware_path, "w") as zip_file:
            for tar_file in recover_os_files:
                if not os.path.exists(tar_file):
                    continue
                zip_file.write(tar_file, os.path.basename(tar_file))
        run_log.info("finish package %s", RecoverMiniOSConstants.FIRMWARE_NAME)

    @classmethod
    def is_running(cls) -> bool:
        return Upgrade.upgrade_state == UpgradeState.UPGRADE_RUNNING_STATE

    @classmethod
    def get_modules(cls) -> Dict[str, ModuleState]:
        return cls.modules

    @classmethod
    def omsdk_upgraded(cls) -> bool:
        """判断OMSDK是否升级过，且成功的"""
        module = cls.modules.get(ModuleType.FIRMWARE.value)
        return module and module.state

    @classmethod
    def change_module_state(cls, **changes: Any):
        """设置当前升级固件信息，并同步更新modules信息"""
        cls.cur_module = dataclasses.replace(cls.cur_module, **changes)
        if cls.cur_module.name in cls.modules:
            cls.modules[cls.cur_module.name].update_front_msg(cls.cur_module)
        # 更新Systems缓存
        try:
            from lib.Linux.systems.systems import SystemInfo
            from bin.lib_adapter import LibAdapter
            sys_obj = SystemInfo()
            sys_obj.get_all_info()
            msg = AppCommonMethod.get_json_info(sys_obj)
            LibAdapter.set_resource(msg, "System", "all", True)
        except Exception as err:
            run_log.error("Failed to update system cache: %s", err)
        run_log.info("update system cache success")

    @classmethod
    def allow_effect(cls) -> bool:
        """根据升级结果，判断是否允许生效"，只要有升级成功的固件，就允许生效"""
        return any(module.state for module in cls.get_modules().values())

    def parse_module_msg(self, xml_path: str) -> str:
        """解析version.xml获取固件类型和版本，并返回固件类型"""
        self.cms_verify(xml_path)
        try:
            # 读取version.xml获取升级类型与版本信息
            version_xml = VersionXmlManager(xml_path)
        except Exception as err:
            Upgrade.upgrade_result = UpgradeResult.ERR_UPGRADE_VERIFY_FAILED
            raise UpgradeError(f"Read version.xml failed, because {err}") from err

        upgrade_module = version_xml.module
        upgrade_version = version_xml.version
        if upgrade_module not in ModuleType.values() or not upgrade_version:
            Upgrade.upgrade_result = UpgradeResult.ERR_UPGRADE_VERIFY_FAILED
            raise UpgradeError("version.xml is invalid.")
        # m.2不支持升级RC1的固件大包
        if all((
                upgrade_module == ModuleType.FIRMWARE.value,
                Env().start_from_m2,
                "RC1" in upgrade_version.upper(),
        )):
            Upgrade.upgrade_result = UpgradeResult.ERR_M2_UPGRADE_TO_RC1_FAILED
            raise UpgradeError("m.2 not support upgrade to rc1")
        Upgrade.change_module_state(name=upgrade_module, version=upgrade_version)
        return upgrade_module

    def get_all_info(self) -> NoReturn:
        # 下一次查询结果时，应为未升级状态
        if Upgrade.upgrade_state == UpgradeState.UPGRADE_NO_STATE:
            Upgrade.upgrade_proc_msg = UpgradeState.UPGRADE_STATE_MSG.get(Upgrade.upgrade_state)
            Upgrade.upgrade_start_time = ""
            Upgrade.upgrade_result = UpgradeResult.ERR_NO_UPGRADE
            Upgrade.cur_module = ModuleState()

        # 记录并返回本次用户进行升级的结果，并恢复为未升级状态
        elif Upgrade.upgrade_state in (UpgradeState.UPGRADE_FAILED, UpgradeState.UPGRADE_SUCCESS):
            Upgrade.upgrade_proc_msg = UpgradeState.UPGRADE_STATE_MSG.get(Upgrade.upgrade_state)
            Upgrade.upgrade_state = UpgradeState.UPGRADE_NO_STATE

        ret_msg = UpgradeResult.UPGRADE_ERROR_CODE_MAP.get(Upgrade.upgrade_result, "Unknown result.")
        self.Messages = {"upgradeState": ret_msg}
        self.PercentComplete = self.get_progress()
        self.Version = Upgrade.cur_module.version
        self.TaskState = Upgrade.upgrade_proc_msg
        self.StartTime = Upgrade.upgrade_start_time
        self.Module = Upgrade.cur_module.name

    def decompress_tar_gz_package(self, tar_path: str, dest_path: str, mode: int) -> NoReturn:
        # 清空并创建升级目录
        self.remove_and_create_dir(dest_path, mode)
        # 校验并解压tar包后删除
        ret = FileOperator.extra_tar_file(tar_path, dest_path, no_same_owner=True)
        if not ret:
            raise UpgradeError(f"Failed to decompress tar file, because {ret.error}.")

    def unzip_upgrade_package(self, zip_path: str) -> NoReturn:
        # 清空并创建升级目录
        self.remove_and_create_dir(UpgradeConstants.UPGRADE_HOME_PATH, 0o700)

        # 将下载的zip升级包拷贝到指定升级目录
        dest_zip_path = os.path.join(UpgradeConstants.UPGRADE_HOME_PATH, os.path.basename(zip_path))
        ret = CopyFile.copy_file(zip_path, dest_zip_path)
        if not ret:
            raise UpgradeError(f"copy zip file failed, {ret.error}")

        # 校验并解压zip包后删除
        ret = FileOperator.extra_zip_file(dest_zip_path, UpgradeConstants.UPGRADE_HOME_PATH, check_symbolic_link=True)
        if not ret:
            raise UpgradeError(f"Failed to decompress  zip file, because {ret.error}.")
        try:
            FileUtils.delete_file_or_link(dest_zip_path)
        except Exception as err:
            raise UpgradeError(f"delete zip file failed, {err}") from err

    def get_components_upgrade_list(self, vercfg_path: str) -> NoReturn:
        self.cms_verify(vercfg_path)
        # 读取vercfg.xml校验sha256，获取所有的升级包
        try:
            Upgrade.upgrade_components_list = VercfgManager(vercfg_path).verify_sha256()
        except Exception as err:
            Upgrade.upgrade_result = UpgradeResult.ERR_UPGRADE_VERIFY_FAILED
            raise UpgradeError(f"verify sha256 failed in vercfg.xml, because {err}") from err

    def exec_upgrade(self) -> NoReturn:
        # A200形态默认使用通用升级接口
        upgrade_handler_map: [str, Callable] = defaultdict(lambda: self.exec_sdk_upgrade)
        upgrade_handler_map[UpgradeConstants.FIRMWARE_TYPE] = self.exec_firmware_upgrade
        upgrade_handler_map[UpgradeConstants.MCU_TYPE] = self.exec_mcu_upgrade
        upgrade_handler_map[UpgradeConstants.NPU_TYPE] = self.exec_npu_upgrade

        try:
            # 记录操作日志
            operate_log.info("[%s@%s] Upgrade executing.", Upgrade.username, Upgrade.request_ip)

            # 需要校验并解析vercfg.xml，获取需要升级的组件
            vercfg_path = os.path.join(UpgradeConstants.UPGRADE_HOME_PATH, UpgradeConstants.VERCFG_XML_NAME)
            self.get_components_upgrade_list(vercfg_path)

            # 获取升级类型，只支持3种固件类型的升级
            version_xml_path = os.path.join(UpgradeConstants.UPGRADE_HOME_PATH, UpgradeConstants.VERSION_XML_NAME)
            upgrade_type = self.parse_module_msg(version_xml_path)
            run_log.info("Now, it is executing %s upgrade", upgrade_type)

            upgrade_handler_map[upgrade_type]()

            # 升级成功更新升级状态
            msg = f"upgrade {Upgrade.cur_module.name} success"
            operate_log.info("[%s@%s] %s.", Upgrade.username, Upgrade.request_ip, msg)
            Upgrade.upgrade_result = UpgradeResult.ERR_UPGRADE_SUCCEED
            Upgrade.upgrade_state = UpgradeState.UPGRADE_SUCCESS
            Upgrade.change_module_state(state=True)
            run_log.info("upgrade %s success, wait to be active", upgrade_type)

        # 升级失败更新升级状态
        except Exception as err:
            msg = f"upgrade {Upgrade.cur_module.name} failed"
            operate_log.info("[%s@%s] %s.", Upgrade.username, Upgrade.request_ip, msg)
            Upgrade.upgrade_state = UpgradeState.UPGRADE_FAILED
            Upgrade.change_module_state(state=False)
            run_log.error("upgrade failed because %s", err)
            FileUtils.delete_file_or_link(os.path.join(RecoverMiniOSConstants.FIRMWARE_PATH,
                                                       RecoverMiniOSConstants.FIRMWARE_NAME))
        finally:
            self.get_progress()
            # 无论升级成功还是失败，升级结束后都要清空临时的升级目录
            FileUtils.delete_full_dir(UpgradeConstants.UPGRADE_HOME_PATH)
            Upgrade.upgrade_mutex.release()

    def exec_firmware_upgrade(self) -> NoReturn:
        tar_gz_paths = TarGzPaths()
        for tar_gz in Upgrade.upgrade_components_list:
            if "om" in tar_gz:
                tar_gz_paths.om_tar_gz = tar_gz
            elif "mefedge" in tar_gz:
                tar_gz_paths.mef_tar_gz = tar_gz
            elif "os" in tar_gz:
                tar_gz_paths.driver_tar_gz = tar_gz
            elif "toolbox" in tar_gz:
                tar_gz_paths.toolbox_tar_gz = tar_gz
            elif "cann" in tar_gz:
                tar_gz_paths.cann_tar_gz = tar_gz

        recover_os_files = [tar_gz for tar_gz in Upgrade.upgrade_components_list if "os" not in tar_gz]
        # 签名校验后重新打包到p7分区
        for tar_gz in recover_os_files:
            self.cms_verify(tar_gz)

        recover_os_files += [
            os.path.join(UpgradeConstants.UPGRADE_HOME_PATH, UpgradeConstants.VERCFG_XML_NAME),
            os.path.join(UpgradeConstants.UPGRADE_HOME_PATH, UpgradeConstants.VERCFG_CMS_NAME),
            os.path.join(UpgradeConstants.UPGRADE_HOME_PATH, UpgradeConstants.VERCFG_CRL_NAME),
        ]
        if Env().start_from_emmc:
            self._copy_tar_to_recover_mini_os_path(recover_os_files)

        # 执行升级
        try:
            self._upgrade_firmware(tar_gz_paths)
        except Exception as err:
            # 升级失败，删除OM升级备区目录
            FileUtils.delete_full_dir(CommonConstants.OM_UPGRADE_DIR_PATH)
            self.clear_mef()
            raise UpgradeError(f"upgrade firmware failed, because {err}") from err

    def clear_mef(self) -> NoReturn:
        """升级进度大于MEF升级进度时，如果异常了应该尝试清理mef"""
        if self.cur_module.process < 70:
            return

        mef = MefInfo()
        if not mef.allow_upgrade:
            run_log.warning("mef not managed by om")
            return

        if not FileUtils.check_script_file_valid(mef.rm_sh, "root", "root"):
            run_log.error("mef recovery.sh invalid")
            return

        if ExecCmd.exec_cmd((mef.rm_sh,)) != 0:
            run_log.warning("clear mef failed")

    def exec_mcu_upgrade(self) -> NoReturn:
        mcu_file_path = ""
        for hpm_file in Upgrade.upgrade_components_list:
            if "mcu" in hpm_file:
                mcu_file_path = hpm_file
                break

        # 执行升级
        shell_cmd = (UpgradeConstants.UPGRADE_DRV_SH, "-u", mcu_file_path, UpgradeConstants.PROGRESS_FILE)
        try:
            self.exec_shell(shell_cmd)
        except Exception as err:
            Upgrade.upgrade_result = UpgradeResult.ERR_UPGRADE_MCU
            raise UpgradeError(f"upgrade mcu failed, because {err}") from err

    def exec_npu_upgrade(self) -> NoReturn:
        npu_file_path = ""
        for tar_gz in Upgrade.upgrade_components_list:
            if "npu" in tar_gz:
                npu_file_path = tar_gz
                break
        npu_file_cms = f"{npu_file_path}.cms"
        npu_file_crl = f"{npu_file_path}.crl"

        # 执行升级
        shell_cmd = (
            UpgradeConstants.UPGRADE_DRV_SH, "-u", npu_file_path, npu_file_cms,
            npu_file_crl, UpgradeConstants.PROGRESS_FILE
        )
        try:
            self.exec_shell(shell_cmd)
        except Exception as err:
            Upgrade.upgrade_result = UpgradeResult.ERR_UPGRADE_NPU
            raise UpgradeError(f"upgrade npu failed, because {err}") from err

    def exec_sdk_upgrade(self) -> NoReturn:
        for upgrade_tar_gz in Upgrade.upgrade_components_list:

            # 升级包必须是tar.gz格式
            if not upgrade_tar_gz.endswith("tar.gz"):
                continue

            # 对升级包进行签名校验
            self.cms_verify(upgrade_tar_gz)
            Upgrade.change_module_state(process=30)

            # 将tar.gz解压到/run/upgrade/sdk_upgrade目录下
            run_log.info("Start upgrade %s, now", self.cur_module.name)
            try:
                self.decompress_tar_gz_package(upgrade_tar_gz, UpgradeConstants.SDK_UPGRADE_UNPACK_PATH, 0o755)
            except Exception as err:
                Upgrade.upgrade_result = UpgradeResult.ERR_UPGRADE_FILE_DECOMPRESS_FAIL
                raise UpgradeError(f"unpack upgrade tar.gz failed, {err}") from err
            Upgrade.change_module_state(process=50)

            # 将升级脚本设为可执行权限
            upgrade_sh = os.path.join(UpgradeConstants.SDK_UPGRADE_UNPACK_PATH, UpgradeConstants.SDK_UPGRADE_SH)
            ret = FilePermission.set_path_permission(upgrade_sh, 0o500)
            if not ret:
                Upgrade.upgrade_result = UpgradeResult.ERR_SDK_UPGRADE_FAILED
                raise UpgradeError(f"set upgrade.sh mode failed, {ret.error}")
            Upgrade.change_module_state(process=60)

            # 执行升级
            try:
                self.exec_shell((upgrade_sh, upgrade_tar_gz, f"{upgrade_tar_gz}.cms", f"{upgrade_tar_gz}.crl"))
            except Exception as err:
                Upgrade.upgrade_result = UpgradeResult.ERR_SDK_UPGRADE_FAILED
                raise UpgradeError(f"upgrade {self.cur_module.name} failed") from err
            Upgrade.change_module_state(process=100)

    # 查询进度
    def get_progress(self) -> int:
        # 如果升级MCU、NPU需要从文件获取
        if Upgrade.cur_module.name in UpgradeConstants.UPGRADE_DRV_PROCESS_TYPE:
            Upgrade.change_module_state(process=self.get_drv_upgrade_process())
        run_log.info("get upgrade progress: %s", Upgrade.cur_module.process)
        return Upgrade.cur_module.process

    # 升级固件
    def post_request(self, request_data: dict) -> List[Union[int, str]]:
        if Upgrade.upgrade_mutex.locked():
            run_log.error("upgrade is in processing, UPGRADE_MUTEX is locked.")
            return [CommonMethods.ERROR, f"ERR.0{UpgradeResult.ERR_UPGRADE_CONFLICT}-locked, upgrade process busy"]
        Upgrade.upgrade_mutex.acquire(UpgradeConstants.UPGRADE_FIRMWARE_TIMEOUT)

        # 校验升级参数
        try:
            upgrade_zip_filepath = self.verify_input_param(request_data)
        except UpgradeError as err:
            run_log.error("verify input param failed, because %s", err.err_msg)
            Upgrade.upgrade_mutex.release()
            return [CommonMethods.ERROR, f"ERR.0{UpgradeResult.ERR_PARAM_INVALID}-{err.err_msg}"]

        run_log.info("check upgrade env success.")
        # 将解压zip包到升级目录
        try:
            self.unzip_upgrade_package(upgrade_zip_filepath)
        except Exception as err:
            run_log.error("unzip upgrade package failed, %s", err)
            Upgrade.upgrade_mutex.release()
            return [CommonMethods.ERROR, f"ERR.0{UpgradeResult.ERR_UPGRADE_FILE_DECOMPRESS_FAIL}-unzip failed."]

        # 初始化升级状态
        Upgrade.upgrade_state = UpgradeState.UPGRADE_RUNNING_STATE
        Upgrade.upgrade_result = UpgradeResult.ERR_UPGRADE_CONFLICT
        Upgrade.cur_module = ModuleState()
        Upgrade.upgrade_proc_msg = UpgradeState.UPGRADE_STATE_MSG.get(Upgrade.upgrade_state)
        Upgrade.upgrade_components_list = []
        Upgrade.upgrade_start_time = time.strftime('%Y-%m-%d %H:%M:%S', time.localtime(time.time()))

        # 执行升级
        run_log.info("The upgrade process starts.")
        threading.Thread(target=self.exec_upgrade).start()

        # 更新返回信息
        self.Messages = {"upgradeState": Upgrade.upgrade_proc_msg}
        self.PercentComplete = 0
        self.StartTime = Upgrade.upgrade_start_time
        self.Version = Upgrade.cur_module.version
        self.TaskState = Upgrade.upgrade_proc_msg
        self.Module = Upgrade.cur_module.name
        return [CommonMethods.OK, "Start to Upgrade"]

    def _upgrade_firmware(self, tar_gz_paths: TarGzPaths) -> NoReturn:
        """先升级OM，再升级驱动"""
        # 对OM包进行签名校验
        self.cms_verify(tar_gz_paths.om_tar_gz)

        # 将om.tar.gz解压到MindXUpgrade目录下
        run_log.info("Start upgrade MindXOM, now")
        try:
            self.decompress_tar_gz_package(tar_gz_paths.om_tar_gz, CommonConstants.OM_UPGRADE_DIR_PATH, 0o755)
        except Exception as err:
            Upgrade.upgrade_result = UpgradeResult.ERR_UPGRADE_FILE_DECOMPRESS_FAIL
            raise UpgradeError(f"untar om package failed {err}") from err
        Upgrade.change_module_state(process=30)

        # 将升级包中的脚本设为可执行权限
        om_sh = os.path.join(CommonConstants.OM_UPGRADE_DIR_PATH, "scripts", UpgradeConstants.UPGRADE_OM_SH)
        ret = FilePermission.set_path_permission(om_sh, 0o500)
        if not ret:
            Upgrade.upgrade_result = UpgradeResult.ERR_UPGRADE_OM
            raise UpgradeError(f"set upgrade_om.sh mode failed, {ret.error}")

        # 执行OM升级
        try:
            self.exec_shell((om_sh,))
        except Exception as err:
            Upgrade.upgrade_result = UpgradeResult.ERR_UPGRADE_OM
            raise UpgradeError(f"upgrade om failed, {err}") from err
        Upgrade.change_module_state(process=50)

        # 执行mef升级，如果未装mef或者mef未被om管理都不升级
        run_log.info("Start upgrade MEF, now")
        mef_info = MefInfo()
        if mef_info.allow_upgrade:
            mef_cmd = (
                mef_info.run_sh, "upgrade", "-file", tar_gz_paths.mef_tar_gz, "-cms",
                f"{tar_gz_paths.mef_tar_gz}.cms", "-crl", f"{tar_gz_paths.mef_tar_gz}.crl", "-delay"
            )
            try:
                self.exec_shell(mef_cmd)
            except Exception as err:
                if Upgrade.upgrade_result == MefBusyConstants.MEF_BUSY_INNER_CODE:
                    Upgrade.upgrade_result = UpgradeResult.ERR_MEF_IS_BUSY
                else:
                    Upgrade.upgrade_result = UpgradeResult.ERR_UPGRADE_MEF
                raise UpgradeError(f"exec upgrade mef failed: {err}") from err
        Upgrade.change_module_state(process=70)

        # 执行驱动升级
        run_log.info("Start upgrade OS, now")
        driver_cmd = (
            UpgradeConstants.UPGRADE_DRV_SH, "-u", tar_gz_paths.driver_tar_gz,
            f"{tar_gz_paths.driver_tar_gz}.cms", f"{tar_gz_paths.driver_tar_gz}.crl", UpgradeConstants.PROGRESS_FILE
        )
        try:
            self.exec_shell(driver_cmd)
        except Exception as err:
            Upgrade.upgrade_result = UpgradeResult.ERR_UPGRADE_ROOTFS
            raise UpgradeError(f"upgrade os failed, {err}") from err
        Upgrade.change_module_state(process=80)

        # 校验解压ascend-dmi, 从m.2启动时不升级toolbox与cann
        if Env().start_from_emmc:
            toolbox_unpack_path = os.path.join(UpgradeConstants.UPGRADE_HOME_PATH, "toolbox")
            cann_unpack_path = os.path.join(UpgradeConstants.UPGRADE_HOME_PATH, "cann")
            try:
                self._unpack_toolkit_package(tar_gz_paths.toolbox_tar_gz, toolbox_unpack_path)
                self._unpack_toolkit_package(tar_gz_paths.cann_tar_gz, cann_unpack_path)
            except Exception as err:
                Upgrade.upgrade_result = UpgradeResult.ERR_UPGRADE_TOOLBOX_FAILED
                raise UpgradeError(f"upgrade ascend_dmi failed, {err}") from err

        # 拷贝service相关文件到备区
        run_log.info("sync MindXOM files.")
        copy_cmd = (
            OMUpgradeConstants.OM_EFFECT_SHELL, UpgradeConstants.FIRMWARE_TYPE,
            OMUpgradeConstants.COPY_TO_BACK_AREA, CommonConstants.OM_UPGRADE_DIR_PATH
        )
        try:
            self.exec_shell(copy_cmd)
        except Exception as err:
            Upgrade.upgrade_result = UpgradeResult.ERR_SYNC_FILES_FAILED
            raise UpgradeError(f"sync MindXOM file failed, {err}") from err
        Upgrade.change_module_state(process=90)

        # 从M.2启动时，升级OS会切区但不会擦除对区，需要设置文件同步标记
        if Env().start_from_m2:
            try:
                self.create_flag_file(OMUpgradeConstants.SYNC_FLAG)
            except Exception as err:
                Upgrade.upgrade_result = UpgradeResult.ERR_CREATE_FILE_FAILED
                raise UpgradeError(f"create sync flag failed, {err}") from err
            run_log.info("Create MindXOM upgrade sync flag succeed")

        # 创建firmware升级成功标记
        try:
            self.create_flag_file(OMUpgradeConstants.UPGRADE_FINISHED_FLAG)
        except Exception as err:
            Upgrade.upgrade_result = UpgradeResult.ERR_CREATE_FILE_FAILED
            raise UpgradeError(f"create upgrade finish flag failed, {err}") from err
        run_log.info("Create MindXOM upgrade finish flag succeed")

        # 驱动升级占整体进度的50%
        Upgrade.change_module_state(process=100)

    def _unpack_toolkit_package(self, tar_gz_path: str, dest_path: str) -> NoReturn:
        # 为了兼容rc1，cann包可能不存在
        if not tar_gz_path:
            return

        # 签名校验ascend-dmi
        self.cms_verify(tar_gz_path)

        # 清空并创建升级目录
        self.remove_and_create_dir(dest_path, 0o700)
        ret = FileOperator.extra_tar_file(tar_gz_path, dest_path, no_same_owner=True)
        if not ret:
            raise UpgradeError(f"Failed to decompress tar file, because {ret.error}.")
