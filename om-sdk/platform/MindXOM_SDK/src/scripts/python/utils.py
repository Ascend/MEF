# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import os
import pwd
from typing import Tuple

from common.utils.exec_cmd import ExecCmd


class OperationRetCode(object):
    SUCCESS_OPERATION = 0
    FAILED_OPERATION = 1


def get_login_user() -> Tuple[str, str]:
    """获取当前登录的操作用户和ip"""
    op_user = pwd.getpwuid(os.getuid()).pw_name
    op_ip = "Local"
    ret_code, output = ExecCmd.exec_cmd_get_output(("who", "am", "i"), wait=60)
    output_li = output.split()
    if ret_code == 0 and len(output_li) == 5:
        op_ip = output_li[4]
    return op_user, op_ip


