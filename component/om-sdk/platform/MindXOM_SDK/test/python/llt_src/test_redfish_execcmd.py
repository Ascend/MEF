# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import os
import unittest

from common.log.logger import run_log
from common.utils.exec_cmd import ExecCmd


class TestRedfishExecCmd(unittest.TestCase):

    def test_redfish_globals_exec_cmd(self):
        """
        exec_cmd 接口测试
        """
        om_file_a = "/etc/test_om_exe.txt"
        om_file_b = "/etc/test_om_exe_back.txt"
        ExecCmd.exec_cmd(["touch", om_file_a], wait=2)
        self.assertTrue(os.path.exists(om_file_a))

        ExecCmd.exec_cmd(["mv", om_file_a, om_file_b], wait=2)
        self.assertTrue(os.path.exists(om_file_b))

        ExecCmd.exec_cmd(["rm", om_file_b], wait=2)
        self.assertFalse(os.path.exists(om_file_b))


if __name__ == '__main__':
    run_log.info("==test_redfish_globals_exec_cmd test start==")
    unittest.main()
    run_log.info("==test_redfish_globals_exec_cmd test end==")
