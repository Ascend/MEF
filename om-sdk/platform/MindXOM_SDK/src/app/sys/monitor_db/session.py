# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from common.backup_restore_service.restore import Restore
from common.constants.base_constants import CommonConstants
from common.db.database import DataBase

# 先尝试恢复
Restore(CommonConstants.MONITOR_BACKUP_DIR, CommonConstants.MONITOR_EDGE_DB_FILE_PATH).entry()
# 该单例将会在第一次导入session时被创建，如果数据库存文件不合法，可能初化话失败导致进程无法拉起；此时数据库表已创建
database = DataBase(CommonConstants.MONITOR_EDGE_DB_FILE_PATH)
session_maker = database.session_maker
simple_session_maker = database.simple_session_maker
