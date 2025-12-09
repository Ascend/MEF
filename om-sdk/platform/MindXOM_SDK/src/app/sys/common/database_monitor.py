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
import grp
import os
import pwd
from abc import ABC
from abc import abstractmethod

from common.file_utils import FileCheck
from common.utils.app_common_method import AppCommonMethod
from common.utils.result_base import Result


class DatabaseMonitorBase(ABC):
    """ 数据库状态监控基类 """

    DB_MODE: str = "600"
    DB_USER: str = pwd.getpwuid(os.getuid()).pw_name
    DB_GROUP: str = grp.getgrgid(os.getgid()).gr_name

    def __init__(self, db_path: str, backup_dir: str):
        self.db_path: str = db_path
        self.backup_dir: str = backup_dir

    def check_database_is_valid(self) -> Result:
        """
        数据库状态
        :return:
        """
        check_ret = FileCheck.check_path_is_exist_and_valid(self.db_path)
        if not check_ret:
            return Result(result=False, err_msg=f"{self.DB_USER} database status is invalid, {check_ret.error}.")

        check_ret = FileCheck.check_path_mode_owner_group(self.db_path, self.DB_MODE, self.DB_USER, self.DB_GROUP)
        if not check_ret:
            return Result(result=False, err_msg=f"{self.DB_USER} database status is invalid, {check_ret.error}.")

        return AppCommonMethod.check_database_available(self.db_path)

    @abstractmethod
    def monitor_database_status(self):
        pass
