# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import sys

from common.constants.base_constants import CommonConstants
from common.log.logger import run_log
from common.kmc_lib.kmc_updater import MultiKmcUpdater, KmcUpdater
from redfish_kmc.kmc_adapter import NetMgrKmcAdapter, ProfileKmcAdapter, ClientPsdAdapter, FlaskPsdAdapter


def register_key_update_tasks():
    MultiKmcUpdater.register_updaters(
        KmcUpdater(CommonConstants.FLASK_KS_DIR, task_name="flask"),
        KmcUpdater(CommonConstants.UDS_CLIENT_KS_DIR, task_name="client"),
        KmcUpdater(CommonConstants.REDFISH_KS_DIR, ks_name="redfish_encrypt", task_name="redfish"),
    )

    MultiKmcUpdater.extend_adapters("redfish", NetMgrKmcAdapter, ProfileKmcAdapter)
    MultiKmcUpdater.extend_adapters("client", ClientPsdAdapter)
    MultiKmcUpdater.extend_adapters("flask", FlaskPsdAdapter)

    # 扩展adapter与updater
    try:
        from extend_interfaces import extend_updater_and_adapters
        extend_updater_and_adapters(MultiKmcUpdater)
    except ImportError as err:
        run_log.warning("Failed to import extension, ignore. %s", err)
    except Exception as err:
        run_log.error("Register extend updater failed, catch %s", type(err))


def start_keystore_update_tasks():
    run_log.info("start keystore update tasks.")
    register_key_update_tasks()

    # 启动定时更新keystore任务
    MultiKmcUpdater.start_auto_update_key_tasks()
    run_log.info("start keystore update tasks finish.")


def manual_update_key():
    register_key_update_tasks()
    MultiKmcUpdater.manual_update_key()


if __name__ == '__main__':
    try:
        manual_update_key()
    except Exception as error:
        run_log.error("manual update key failed, catch %s", error.__class__.__name__)
        sys.exit(1)
    sys.exit(0)
