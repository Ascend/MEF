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
import sys

from common.constants.base_constants import CommonConstants
from common.log.logger import run_log
from common.kmc_lib.kmc_updater import KmcUpdater, MultiKmcUpdater
from monitor_kmc.kmc_adapter import NginxPsdAdapter, UdsPsdAdapter, DevmConfigAdapter


def register_key_update_tasks():
    MultiKmcUpdater.register_updaters(
        KmcUpdater(CommonConstants.NGINX_KS_DIR, task_name="nginx"),
        KmcUpdater(CommonConstants.UDS_KS_DIR, task_name="uds"),
    )

    MultiKmcUpdater.extend_adapters("nginx", NginxPsdAdapter)
    MultiKmcUpdater.extend_adapters("uds", DevmConfigAdapter, UdsPsdAdapter)

    # 扩展adapter与updater
    try:
        from bin.extend_interfaces import extend_updater_and_adapters
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
