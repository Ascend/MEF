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
from common.db.database import DataBase
from common.log.logger import run_log
from net_manager.models import NetManager, CertManager
from user_manager.models import EdgeConfig, HisPwd, Session, User, LastLoginInfo


def register_models():
    DataBase.register_models(User, Session, HisPwd, EdgeConfig, CertManager, NetManager, LastLoginInfo)
    try:
        from extend_interfaces import register_extend_models
        register_extend_models(DataBase)
    except ImportError as err:
        run_log.warning("Failed to import extension, ignore. %s", err)
    except Exception as err:
        run_log.error("Register extend models failed, catch %s", err.__class__.__name__)
