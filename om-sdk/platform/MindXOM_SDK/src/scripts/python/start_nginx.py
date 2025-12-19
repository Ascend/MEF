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
import os
import sys
import threading
import time

from common.constants.base_constants import CommonConstants
from common.log.logger import run_log
from common.utils.exec_cmd import ExecCmd
from common.kmc_lib.kmc import Kmc


class ResultCode(object):
    SUCCESS_OPERATION: int = 0
    FAILED_OPERATION: int = 1


class StartNginxOperator:
    def __init__(self):
        self.nginx_path = os.path.join(CommonConstants.OM_WORK_DIR_PATH, "software/nginx/sbin/nginx")
        self.nginx_conf_path = os.path.join(CommonConstants.OM_WORK_DIR_PATH, "software/nginx/conf/nginx.conf")
        self.nginx_prefix_path = os.path.join(CommonConstants.OM_WORK_DIR_PATH, "software/nginx")
        self.fifo_file_name = "/run/nginx/nginxfifo"
        self.cert_primary_ksf = "/home/data/config/default/om_cert.keystore"
        self.cert_standby_ksf = "/home/data/config/default/om_cert_backup.keystore"
        self.alg_config_file = "/home/data/config/default/om_alg.json"
        self.psd_file = "/home/data/config/default/server_kmc.psd"
        self.kmc_inst = Kmc(self.cert_primary_ksf, self.cert_standby_ksf, self.alg_config_file)

    def execute(self) -> int:
        try:
            return self._execute()
        except Exception as err:
            run_log.error("Start nginx failed, caught exception: %s", err)
            return ResultCode.FAILED_OPERATION
        finally:
            # 清理fifo文件
            if os.path.exists(self.fifo_file_name):
                os.unlink(self.fifo_file_name)
                run_log.info("Remove fifo file %s success.", self.fifo_file_name)

    def start_nginx(self):
        start_nginx_cmd = [self.nginx_path, "-c", self.nginx_conf_path, "-p", self.nginx_prefix_path]
        ret = ExecCmd.exec_cmd_get_output(start_nginx_cmd, wait=5)
        if ret[0] != 0:
            run_log.error("Exec start nginx cmd failed: %s.", ret[1])
            return

        run_log.info("Exec start nginx cmd success.")

    def _execute(self):
        if os.path.exists(self.fifo_file_name):
            os.unlink(self.fifo_file_name)
            run_log.info("Remove old fifo file %s success.", self.fifo_file_name)

        # 默认600权限，前提py脚本nobody运行
        os.mkfifo(self.fifo_file_name, mode=0o600)

        if not os.path.exists(self.fifo_file_name):
            run_log.error("Create fifo file %s failed.", self.fifo_file_name)
            return ResultCode.FAILED_OPERATION

        run_log.info("Create fifo file %s success.", self.fifo_file_name)

        threading.Thread(target=self.start_nginx).start()
        time.sleep(1)

        # 如果没有其他进程如nginx读文件，file.write可能会一致阻塞，可能会卡死，需要先判断nginx进程是否存在，没有则直接return
        nginx_cmd = f"ps -ef | grep -w {self.nginx_path} | grep -v grep"
        nginx_cmd += " | awk '{print $2}'"
        ret = ExecCmd.exec_cmd_use_pipe_symbol(nginx_cmd, wait=5)
        if ret[0] != 0 or not ret[1]:
            run_log.error("Nginx master process not exist, abort.")
            return ResultCode.FAILED_OPERATION

        with open(self.psd_file, "r") as file:
            psd_data = self.kmc_inst.decrypt(file.read())

        # os.O_WRONLY 只读方式打开 | os.O_CREAT若文件不存在将创建 | os.O_TRUNC 针对FIFO文件会被忽略
        with os.fdopen(os.open(self.fifo_file_name, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600), "w") as file:
            file.write(psd_data)

        time.sleep(3)
        # 查看nginx的worker进程是否启动
        master_cmd = f"ps -ef | grep -w {self.nginx_path} | grep 'nginx: master process' | grep -v grep"
        master_cmd += " | awk '{print $2}'"
        ret = ExecCmd.exec_cmd_use_pipe_symbol(master_cmd, wait=5)
        if ret[0] != 0 or not ret[1].strip():
            run_log.error("Start nginx failed, no nginx master process.")
            return ResultCode.FAILED_OPERATION

        run_log.info("Start nginx successful, master process id %s.", ret[1].strip())
        return ResultCode.SUCCESS_OPERATION


if __name__ == "__main__":
    sys.exit(StartNginxOperator().execute())
