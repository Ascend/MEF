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
from common.db.database import DataBase
from common.log.logger import run_log
from devm.devm_configs import DevmConfig
from lib.Linux.EdgeSystem.models import HddInfo
from lib.Linux.systems.disk.models import MountWhitelistPath
from lib.Linux.systems.models import TimeZoneConfig
from lib.Linux.systems.nfs.models import NfsCfg
from lib.Linux.systems.nic.models import NetConfig
from lib.Linux.systems.security_service.models import PunyDictSign, LoginRules


def register_models():
    DataBase.register_models(
        DevmConfig, TimeZoneConfig, PunyDictSign,
        MountWhitelistPath, HddInfo, NfsCfg, LoginRules, NetConfig
    )

    try:
        from bin.extend_interfaces import register_extend_models
        register_extend_models(DataBase)
    except ImportError as err:
        run_log.warning("Failed to import extension, ignore. %s", err)
    except Exception as err:
        run_log.error("Register extend models failed, catch %s", err.__class__.__name__)
