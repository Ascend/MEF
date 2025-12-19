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
import configparser
import os.path
from dataclasses import dataclass

from common.constants.base_constants import CommonConstants
from common.constants.base_constants import MefNetStatus
from common.file_utils import FileUtils
from common.init_cmd import cmd_constants
from common.log.logger import run_log
from common.schema import field
from common.utils import ability_policy
from common.utils.ability_policy import OmAbility, AbilityConfig
from common.utils.exec_cmd import ExecCmd
from common.utils.system_utils import SystemUtils


@dataclass
class MefInfo:
    install_path: str = field(init=False, comment="MEF安装路径")
    run_sh: str = field(default="", comment="MEF的run.sh路径")
    exchange_root_ca_sh: str = field(default="", comment="MEF的exchange_root_ca.sh路径")
    import_ca_path: str = field(default="", comment="MEF需要导入的root_ca证书的路径")
    export_ca_path: str = field(default="", comment="MEF需要导出给OM的root_ca证书的路径")
    service: str = field(default="mef-edge-main", comment="MEF服务")
    status: int = field(default=MefNetStatus.UNKNOWN.value, comment="MEF网管状态")
    cp_sys_sh: str = field(default="", comment="MEF生效拷贝服务文件的脚本路径")
    rm_sh: str = field(default="", comment="MEF删除升级目录的脚本路径")
    ca_data: str = field(default="", comment="MEF的root_ca证书内容")

    def __post_init__(self):
        self.install_path = self.get_install_path()
        if self.install_path:
            self.run_sh = os.path.join(self.install_path, "run.sh")
            self.cp_sys_sh = os.path.join(self.install_path, "edge_installer", "script", "prepare_back_ptn_svc.sh")
            self.rm_sh = os.path.join(self.install_path, "edge_installer", "script", "recovery.sh")
            self.exchange_root_ca_sh = os.path.join(self.install_path, "edge_installer/script/exchange_root_ca.sh")

        self.import_ca_path = os.path.join(CommonConstants.OM_WORK_DIR_PATH, "software/ibma/cert/root_ca.cert")
        self.export_ca_path = os.path.join(CommonConstants.OM_WORK_DIR_PATH, "software/ibma/cert/root_ca_mef.cert")
        self.status = self.get_net_status()
        self.ca_data = self.get_ca_data()

    @property
    def allow_upgrade(self) -> bool:
        """是否允许升级"""
        ability_policy.init(AbilityConfig.CONFIG_FILE)
        if not ability_policy.is_allow(OmAbility.MEF_CONFIG):
            run_log.warning("mef ability is disable")
            return False

        return self.status == MefNetStatus.FD_OM.value

    @staticmethod
    def check_script_file_valid(file_path):
        """校验MEF的run.sh脚本合法性"""
        return FileUtils.check_script_file_valid(file_path, "root", "root")

    def get_install_path(self) -> str:
        if SystemUtils.get_model_by_npu_smi() in CommonConstants.A500_MODELS:
            return os.path.realpath(CommonConstants.A500_MEF_INSTALL_PATH)

        status, out = ExecCmd.exec_cmd_get_output((cmd_constants.OS_CMD_SYSTEMCTL, "cat", self.service))
        if status != 0:
            run_log.error("cat mef service failed.")
            return ""

        cfg_parser = configparser.RawConfigParser()
        try:
            cfg_parser.read_string(out)
            return cfg_parser.get("Unit", "MefEdgeSoftwareDir")
        except Exception as err:
            run_log.error("get mef install dir failed, catch %s", err.__class__.__name__)
            return ""

    def get_net_status(self) -> int:
        """获取MEF网管状态"""
        if not self.run_sh:
            return MefNetStatus.UNKNOWN.value

        ret = self.check_script_file_valid(self.run_sh)
        if not ret:
            run_log.warning("mef run script file invalid: %s", ret.error)
            return MefNetStatus.UNKNOWN.value

        status, _ = ExecCmd.exec_cmd_get_output((self.run_sh, "getnetconfig"))
        return {
            0: MefNetStatus.FD_OM.value,
            1: MefNetStatus.UNKNOWN.value,
            2: MefNetStatus.MEF.value,
        }.get(status, MefNetStatus.UNKNOWN.value)

    def get_ca_data(self) -> str:
        """获取MEF的根CA证书内容"""
        if not self.export_ca_path or not os.path.exists(self.export_ca_path):
            return ""

        ret = FileUtils.check_script_file_valid(self.export_ca_path, "root", "root")
        if not ret:
            run_log.error("ca path %s invalid: %s", ret.error)
            return ""

        if os.path.getsize(self.export_ca_path) > CommonConstants.MAX_CERT_LIMIT:
            run_log.error("ca content is too large")
            return ""

        try:
            with open(self.export_ca_path) as file:
                return file.read(CommonConstants.MAX_CERT_LIMIT)
        except Exception as error:
            run_log.error("read mef root ca cert file failed: %s", error)
            return ""
