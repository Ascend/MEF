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
        ExecCmd.exec_cmd(["touch", om_file_a], 2)
        self.assertTrue(os.path.exists(om_file_a))

        ExecCmd.exec_cmd(["mv", om_file_a, om_file_b], 2)
        self.assertTrue(os.path.exists(om_file_b))

        ExecCmd.exec_cmd(["rm", om_file_b], 2)
        self.assertFalse(os.path.exists(om_file_b))


if __name__ == '__main__':
    run_log.info("==test_redfish_globals_exec_cmd test start==")
    unittest.main()
    run_log.info("==test_redfish_globals_exec_cmd test end==")
