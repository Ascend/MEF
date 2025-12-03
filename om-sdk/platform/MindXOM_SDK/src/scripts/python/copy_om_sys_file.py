# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import argparse
import os.path
import shutil
import sys
from collections import namedtuple
from typing import NoReturn, Type, Iterable

from bin.environ import Env
from common.constants.base_constants import CommonConstants
from common.constants.upgrade_constants import OMUpgradeConstants, UpgradeConstants
from common.file_utils import FileCopy, FileUtils
from common.file_utils import FileCreate
from common.log.logger import run_log
from common.utils.exception_utils import OperateBaseError
from common.utils.exec_cmd import ExecCmd
from common.utils.system_utils import SystemUtils

EffectParam = namedtuple("EffectParam", ["dest_path", "copy_type", "om_src_path"])


class EffectResultCode(object):
    SUCCESS_OPERATION: int = 0
    FAILED_OPERATION: int = 1


class CopyFileError(OperateBaseError):
    pass


class ParseParamError(OperateBaseError):
    pass


class CopySysFileOperator:

    @staticmethod
    def check_dest_path(dest_path: str) -> str:
        if dest_path not in OMUpgradeConstants.OM_SYS_PATH:
            raise ParseParamError("Illegal parameter <Dest path>.")
        return dest_path

    @staticmethod
    def check_copy_type(copy_type: str) -> str:
        if copy_type not in OMUpgradeConstants.OM_COPY_TYPE:
            raise ParseParamError("Illegal parameter <Copy type>.")
        return copy_type

    @staticmethod
    def check_om_src_path(om_src_path: str) -> str:
        if om_src_path not in OMUpgradeConstants.OM_COPY_PATH:
            raise ParseParamError("Illegal parameter <Src path>.")
        return om_src_path

    @staticmethod
    def copy_files(files: Iterable[str], src_dir: str, dest_dir: str, mode: int) -> NoReturn:
        for file in files:
            src_file_path = os.path.join(src_dir, file)
            dest_file_path = os.path.join(dest_dir, file)
            FileUtils.delete_file_or_link(dest_file_path)
            ret = FileCopy.copy_file(src_file_path, dest_file_path, mode=mode)
            if not ret:
                raise CopyFileError(f"copy {file} failed, {ret.error}")

    @staticmethod
    def copy_mef_sys(dest_path: str):
        # 非firmware升级场景且非A500场景，无需拷贝
        if SystemUtils.get_model_by_npu_smi() not in CommonConstants.A500_MODELS:
            run_log.info("not a500, no need to copy mef service")
            return

        # 局部导入，保证环境上相关准备已存在
        run_log.info("copy mef service")
        from lib.Linux.mef.mef_info import MefInfo
        mef_info = MefInfo()
        if not mef_info.allow_upgrade:
            run_log.warning("mef not managed by om")
            return

        ret = mef_info.check_script_file_valid(mef_info.cp_sys_sh)
        if not ret:
            run_log.error("mef copy service sh invalid: %s, path: %s", ret.error, mef_info.cp_sys_sh)
            raise CopyFileError("mef copy service sh invalid")

        if ExecCmd.exec_cmd((mef_info.cp_sys_sh, dest_path)) != 0:
            run_log.error("copy mef service failed")
            raise CopyFileError("copy mef service failed")
        run_log.info("copy mef service success")

    def execute(self) -> int:
        try:
            args = self._parse_param()
            self.copy_sys_file_to_system_dir(args.dest_path, args.copy_type, args.om_src_path)
        except Exception as err:
            run_log.error("effect MindXOM failed, %s", err)
            return EffectResultCode.FAILED_OPERATION

        run_log.info("effect MindXOM success")
        return EffectResultCode.SUCCESS_OPERATION

    def copy_sys_file_to_system_dir(self, dest_path: str, copy_type: str, om_scr_path: str) -> NoReturn:
        # 拷贝service服务
        dest_usr_lib_systemd_system = os.path.join(dest_path, "usr", "lib", "systemd", "system")
        scr_service_path = os.path.join(om_scr_path, "software", "service_main")
        self._copy_service(scr_service_path, dest_usr_lib_systemd_system)
        run_log.info("copy service finish.")

        # 设置软连接
        target_dir_path = os.path.join(dest_path, "etc", "systemd", "system", "multi-user.target.wants")
        src_dir_path = os.path.join(OMUpgradeConstants.SYS_MAIN_PART, "usr", "lib", "systemd", "system")
        self._set_soft_link(src_dir_path, target_dir_path)
        run_log.info("set soft link finish.")

        if copy_type == OMUpgradeConstants.COPY_TO_BACK_AREA:
            # 拷贝mef服务文件
            self.copy_mef_sys(dest_path)

            # 拷贝g2.crl至备区
            src_g2_crl_path = os.path.join("/", "etc", "hwsipcrl")
            dest_g2_clr_path = os.path.join(dest_path, "etc", "hwsipcrl")
            self._copy_g2_crl(src_g2_crl_path, dest_g2_clr_path)
            run_log.info("copy g2 crl finish.")

            # 拷贝ascend-dmi的软件包到备区
            ascend_dmi_path = os.path.join(UpgradeConstants.UPGRADE_HOME_PATH, "toolbox")
            dest_ascend_dmi_path = os.path.join(dest_path, "usr", "local", "Ascend", "toolbox")
            self._copy_toolkit(ascend_dmi_path, dest_ascend_dmi_path)
            run_log.info("copy ascend dmi finish")

            # 拷贝cann的软件包到备区
            cann_nnrt_path = os.path.join(UpgradeConstants.UPGRADE_HOME_PATH, "cann")
            dest_cann_nnrt_path = os.path.join(dest_path, "usr", "local", "Ascend", "cann")
            self._copy_toolkit(cann_nnrt_path, dest_cann_nnrt_path)
            run_log.info("copy cann finish")

        # 保证文件落盘
        os.sync()

    def _copy_g2_crl(self, src_dir: str, dest_dir: str) -> NoReturn:
        if os.path.exists(os.path.join(src_dir, UpgradeConstants.ASCENDSIP_G2_CRL)):
            if not os.path.exists(dest_dir):
                FileCreate.create_dir(dest_dir, 0o750)
            self.copy_files((UpgradeConstants.ASCENDSIP_G2_CRL,), src_dir, dest_dir, 0o644)

    def _copy_toolkit(self, src_path, dest_path) -> NoReturn:
        # 从m.2启动，升级场景不需要拷贝toolbox与cann，升级场景下/run/upgrade无对应目录
        if Env().start_from_m2 and not os.path.exists(src_path):
            return

        # 清空并创建升级目录
        try:
            FileUtils.delete_full_dir(dest_path)
        except Exception as err:
            raise CopyFileError(f"clear dir {dest_path} failed") from err

        ascend_dir = os.path.dirname(dest_path)
        if not os.path.exists(ascend_dir):
            FileCreate.create_dir(ascend_dir, 0o755)

        # 拷贝文件至备区
        try:
            shutil.move(src_path, ascend_dir)
        except Exception as err:
            raise CopyFileError(f"Failed to copy ascend dmi files, because {err}") from err

    def _set_soft_link(self, src_dir_path, target_dir_path: str) -> bool:
        files = (
            OMUpgradeConstants.OM_INIT_SERVICE,
            OMUpgradeConstants.IBMA_EDGE_START_SERVICE,
            OMUpgradeConstants.START_NGINX_SERVICE,
            OMUpgradeConstants.PLATFORM_APP_SERVICE,
            OMUpgradeConstants.SYNC_SYS_FILES_SERVICE
        )
        try:
            for serv in files:
                scr_file_path = os.path.join(src_dir_path, serv)
                target_file_path = os.path.join(target_dir_path, serv)
                if os.path.islink(target_file_path):
                    os.remove(target_file_path)
                os.symlink(scr_file_path, target_file_path)
        except Exception as err:
            raise CopyFileError(f"create soft link failed, {err}") from err

        return True

    def _copy_service(self, src_dir: str, dest_dir: str) -> NoReturn:
        files = (
            OMUpgradeConstants.OM_INIT_SERVICE,
            OMUpgradeConstants.IBMA_EDGE_START_SERVICE,
            OMUpgradeConstants.START_NGINX_SERVICE,
            OMUpgradeConstants.PLATFORM_APP_SERVICE,
            OMUpgradeConstants.SYNC_SYS_FILES_SERVICE
        )
        self.copy_files(files, src_dir, dest_dir, 0o644)

    def _parse_param(self) -> Type[EffectParam]:
        """解析校验参数"""
        parse = argparse.ArgumentParser()
        parse.add_argument("dest_path", type=self.check_dest_path, help="拷贝的目的路径")
        parse.add_argument("copy_type", type=self.check_copy_type, help="拷贝方式，包括安装、拷贝至备区、拷贝至主区")
        parse.add_argument("om_src_path", type=self.check_om_src_path, help="拷贝的源路径")
        return parse.parse_args(namespace=EffectParam)


if __name__ == "__main__":
    sys.exit(CopySysFileOperator().execute())
