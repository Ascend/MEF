# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from common.db.database import DataBase
from common.log.logger import run_log
from net_manager.models import NetManager, CertManager, CertInfo, FdPreCert, PreCertInfo
from user_manager.models import EdgeConfig, HisPwd, Session, User, LastLoginInfo


def register_models():
    DataBase.register_models(
        User, Session, HisPwd, EdgeConfig, CertManager, NetManager, LastLoginInfo, CertInfo, FdPreCert, PreCertInfo
    )
    try:
        from extend_interfaces import register_extend_models
        register_extend_models(DataBase)
    except ImportError as err:
        run_log.warning("Failed to import extension, ignore. %s", err)
    except Exception as err:
        run_log.error("Register extend models failed, catch %s", err.__class__.__name__)
