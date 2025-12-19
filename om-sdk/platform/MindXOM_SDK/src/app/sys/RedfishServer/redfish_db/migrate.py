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
import signal
import sys

from common.backup_restore_service.restore import Restore
from common.constants.base_constants import CommonConstants, MigrateOperate
from common.db.migrate import Migrate, input_args
from common.log.logger import run_log
from common.utils.scripts_utils import signal_handler
from redfish_db.init_structure import INIT_COLUMNS
from redfish_db.register import register_models


def main(argv=None):
    args = input_args(argv)
    register_models()
    if args.operate == MigrateOperate.INSTALL.value:
        return Migrate.execute_on_install(CommonConstants.REDFISH_EDGE_DB_FILE_PATH, INIT_COLUMNS)

    # 先尝试恢复数据库，数据库正常的情况下不会被恢复
    Restore(CommonConstants.REDFISH_BACKUP_DIR, CommonConstants.REDFISH_EDGE_DB_FILE_PATH).entry()
    return Migrate.execute(CommonConstants.REDFISH_EDGE_DB_FILE_PATH, INIT_COLUMNS)


if __name__ == '__main__':
    # 注册退出信号的中断处理函数
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)
    try:
        sys.exit(main())
    except Exception as err:
        run_log.error("unknown error, %s", err.__class__.__name__)
        sys.exit(1)
