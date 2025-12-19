# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import datetime
import os
from itertools import islice
from pathlib import Path

from common.constants.base_constants import CommonConstants
from common.file_utils import FileCheck
from common.init_cmd import cmd_constants
from common.log.logger import run_log
from common.utils.exec_cmd import ExecCmd
from lib.Linux.EdgeSystem.hdd_info_mgr import clear_or_save_hdd_info, get_hdd_serial_number, get_hdd_info_count


class Event:
    """
    功能描述：事件上报
    接口：NA
    """
    RESTART_FLAG = "/run/restart_flag"
    RESTARTING_FLAG = "/run/restarting_flag"
    # 硬盘的SCSI地址, 0:0:0:0: mdisk0, 1:0:0:0: hdisk0, 2:0:0:0: hdisk1
    SCSI_ADDR = ("0:0:0:0", "1:0:0:0", "2:0:0:0")

    def __init__(self):
        self.event_time = ""
        self.result = []

    @staticmethod
    def clear_or_write_hdd_config(hdd_sn=None):
        try:
            clear_or_save_hdd_info(serial_number=hdd_sn)
        except Exception as err:
            run_log.error("clear or save hdd info err, catch %s", err.__class__.__name__)

    def on_start(self):
        """进程启动时，需要将hdd的serial_number先存一下"""
        try:
            hdd_count = get_hdd_info_count()
        except Exception as err:
            run_log.error("Get hdd count err, catch %s", err.__class__.__name__)
            return

        # 已存在则忽略
        if hdd_count:
            run_log.info("Hdd info already existed.")
            return

        # 将序列号数据存入数据库, 格式: "SNXXXXX,NULL,NULL"
        self.clear_or_write_hdd_config(self.get_hdd_dev_sn())

    def get_all_info(self, event_type="all"):
        if event_type not in ("all"):
            run_log.error("invalid event type parameter")
            return

        self.event_time = self.get_event_generation_time()
        if not self.event_time:
            return

        if not self.get_hdd_removal_event():
            # hdd 未被拔出，则要检查是否被替换
            self.get_hdd_replacement_event()

        self.rename_restarting_flag()

    def get_event_generation_time(self):
        ret = FileCheck.check_path_is_exist_and_valid(self.RESTARTING_FLAG)
        if not ret:
            run_log.error("%s invalid: %s", self.RESTARTING_FLAG, ret.error)
            return ''

        try:
            with open(self.RESTARTING_FLAG) as restart_fp:
                sys_start_time = restart_fp.readline().strip().split("=")[-1]
        except Exception as err:
            run_log.error("read RESTARTING_FLAG failed %s", err)
            return ''

        try:
            utc_time = datetime.datetime.utcfromtimestamp(int(sys_start_time))
        except Exception as err:
            run_log.error("utcfromtimestamp failed: %s", err)
            return ''

        tzinfo = datetime.timezone.utc
        event_time = utc_time.replace(microsecond=0, tzinfo=tzinfo).isoformat()
        run_log.info("utcfromtimestamp success")
        return event_time

    def rename_restarting_flag(self):
        payload_publish = {
            "alarm": [
                {
                    "type": "event",
                    "alarmId": "0x01000000",
                    "alarmName": "system restart",
                    "resource": "system",
                    "perceivedSeverity": "OK",
                    "timestamp": self.event_time,
                    "notificationType": "",
                    "detailedInformation": "System restart",
                    "suggestion": "",
                    "reason": "",
                    "impact": ""
                }
            ]
        }
        check_ret = FileCheck.check_input_path_valid(self.RESTART_FLAG)
        if not check_ret:
            run_log.error("check RESTART_FLAG failed:%s", check_ret.error)
            return

        try:
            os.rename(self.RESTARTING_FLAG, self.RESTART_FLAG)
        except Exception as ex:
            run_log.error('rename restarting flag failed %s', ex)
            return
        self.result.append(payload_publish)

    def get_hdd_removal_event(self) -> bool:
        check_tec_cmd = f"{cmd_constants.OS_CMD_NPU_SMI} info -t temp -i 0 | grep TEC_TEM"
        result, msg = ExecCmd.exec_cmd_use_pipe_symbol(check_tec_cmd, 10)
        if result != 0:
            run_log.error("Get HDD TEC_TEM info failed. %s", msg)
            return False

        tec_tem = msg.split(":")[1].strip() if msg else ""
        if tec_tem not in ("NA", ""):
            run_log.info("Get HDD temperature is [%s], check whether the hard disk is replaced.", tec_tem)
            return False

        # 获取加热模块温度数值失败，硬盘被拔出，上报硬盘被拔出事件
        payload_publish = {
            "alarm": [
                {
                    "type": "event",
                    "alarmId": "0x01010001",
                    "alarmName": "hard disk removal",
                    "resource": "HARD_DISK0",
                    "perceivedSeverity": "OK",
                    "timestamp": self.event_time,
                    "notificationType": "",
                    "detailedInformation": "Hard disk removal",
                    "suggestion": "",
                    "reason": "",
                    "impact": ""
                }
            ]
        }
        # 清空配置文件
        self.clear_or_write_hdd_config()
        self.result.append(payload_publish)
        run_log.info("Check the hard disk removal, report hard disk removal event.")
        return True

    def get_hdd_replacement_event(self):
        # 获取当前硬盘序列号信息
        hdd_sn_cur = self.get_hdd_dev_sn()
        try:
            # 获取数据库硬盘序列号信息
            hdd_sn_his = get_hdd_serial_number()
        except Exception as err:
            run_log.error("get HDD sn to error. %s", err.__class__.__name__)
            return

        # 获取数据库sn为空并且当前sn不为空，则新插入了硬盘，用户自己行为，不上报事件，并写入新的硬盘序列号到配置文件
        if hdd_sn_his == "NULL,NULL,NULL" and hdd_sn_cur != "NULL,NULL,NULL":
            run_log.info("Insert a new hard disk and do not detect events.")
            # 写入新的硬盘序列号到hdd配置中
            self.clear_or_write_hdd_config(hdd_sn_cur)
            return

        # 当前序列号和数据库的一致则认为硬盘未被替换
        if hdd_sn_his == hdd_sn_cur:
            run_log.info("Hard disk change event not detected.")
            return

        payload_publish = {
            "alarm": [
                {
                    "type": "event",
                    "alarmId": "0x01010000",
                    "alarmName": "hard disk replacement",
                    "resource": "HARD_DISK0",
                    "perceivedSeverity": "OK",
                    "timestamp": self.event_time,
                    "notificationType": "",
                    "detailedInformation": "Hard disk replacement",
                    "suggestion": "",
                    "reason": "",
                    "impact": ""
                }
            ]
        }
        self.result.append(payload_publish)
        run_log.info("report HD replacement event to FD success.")
        # 写入新的硬盘序列号到hdd配置中
        self.clear_or_write_hdd_config(hdd_sn_cur)

    def get_hdd_dev_sn(self) -> str:
        """ 获取HDD设备序列号 """
        sn_list = ["NULL"] * len(self.SCSI_ADDR)
        for path in islice(Path("/sys/block").glob("*"), CommonConstants.MAX_ITER_LIMIT):
            if "sata" not in path.resolve().as_posix():
                continue

            for index, scsi in enumerate(self.SCSI_ADDR):
                if scsi not in path.resolve().parts:
                    continue

                cmd_str = f"{cmd_constants.OS_CMD_SMARTCTL} -a /dev/{path.name} | grep Serial\ Number"
                result, msg = ExecCmd.exec_cmd_use_pipe_symbol(cmd_str, 10)
                if result != 0 or not msg:
                    run_log.warning("get '%s' serial number empty, HDD not presence.", scsi)
                    continue

                sn_list[index] = msg.partition(":")[-1].strip()

        return ",".join(sn_list)
