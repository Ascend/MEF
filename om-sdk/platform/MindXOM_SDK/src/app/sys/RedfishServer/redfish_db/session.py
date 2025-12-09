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
from common.backup_restore_service.restore import Restore
from common.constants.base_constants import CommonConstants
from common.db.database import DataBase

# 先尝试恢复数据库，数据库正常的情况下不会被恢复
Restore(CommonConstants.REDFISH_BACKUP_DIR, CommonConstants.REDFISH_EDGE_DB_FILE_PATH).entry()
# 该单例将会在第一次导入session时被创建，如果数据库存文件不合法，可能初化话失败导致进程无法拉起；此时数据库表已创建
database = DataBase(db_path=CommonConstants.REDFISH_EDGE_DB_FILE_PATH)
session_maker = database.session_maker
simple_session_maker = database.simple_session_maker
