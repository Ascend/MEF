# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import sys

from common.constants.base_constants import CommonConstants
from common.file_utils import FileUtils
from common.utils import ability_policy
from common.utils.ability_policy import OmAbility, AbilityConfig
from common.utils.exec_cmd import ExecCmd
from common.utils.system_utils import SystemUtils
from lib.Linux.mef.mef_info import MefInfo
from logger import upgrade_log


def effect_mef() -> int:
    if SystemUtils.get_model_by_npu_smi() not in CommonConstants.A500_MODELS:
        return 1

    try:
        ability_policy.init(AbilityConfig.CONFIG_FILE)
    except Exception as err:
        upgrade_log.error("init ability_policy.json failed: %s", err)
        return 1

    if not ability_policy.is_allow(OmAbility.MEF_CONFIG):
        upgrade_log.warning("mef ability is disable")
        return 1

    upgrade_log.info("effect mef")
    mef_info = MefInfo()
    if not FileUtils.check_script_file_valid(mef_info.run_sh, user="root", group="root"):
        upgrade_log.error("mef run.sh invalid.")
        return 1

    if ExecCmd.exec_cmd((mef_info.run_sh, "effect")) != 0:
        upgrade_log.error("effect mef failed")
        return 1

    upgrade_log.info("effect mef finish.")
    return 0


if __name__ == '__main__':
    sys.exit(effect_mef())
