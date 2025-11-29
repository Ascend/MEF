# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import sys

from common.log.logger import run_log
from common.utils.exec_cmd import ExecCmd
from lib.Linux.systems.time_zone_mgr import get_time_zone_offset


class TimeZone:
    CMD_TIME_OUT = 120

    @staticmethod
    def read_time_zone_cfg() -> str:
        try:
            return get_time_zone_offset()
        except Exception as err:
            run_log.error("read time zone meet unknown err: %s", err)
            return ""

    def set_time_zone(self) -> bool:
        time_zone = self.read_time_zone_cfg()
        if not time_zone:
            run_log.info("time_zone not change!")
            return True

        cmd = ("timedatectl", "set-timezone", time_zone)
        ret = ExecCmd.exec_cmd_get_output(cmd, wait=self.CMD_TIME_OUT)
        if ret[0] != 0:
            run_log.error("Set timezones %s failed", time_zone)
            return False

        run_log.info("Set timezones %s success", time_zone)
        return True


if __name__ == '__main__':
    if not TimeZone().set_time_zone():
        sys.exit(1)
    sys.exit(0)
