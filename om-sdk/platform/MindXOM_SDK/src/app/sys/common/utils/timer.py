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
import threading
from typing import Callable

from common.log.logger import run_log


class RepeatingTimer(threading.Timer):
    def run(self):
        while not self.finished.is_set():
            self.function(*self.args, **self.kwargs)
            self.finished.wait(self.interval)


class DynamicRepeatingTimer(threading.Timer):
    """等待时间动态调整的循环定时器
    每个循环周期最多重试5次，每次等待时间依次递增：interval * 1, interval * 2, ... , interval * MAX_RETRY_COUNT_PER_CYCLE
    """
    MAX_RETRY_COUNT_PER_CYCLE = 5

    def __init__(self, name: str, interval: float, function: Callable, args=None, kwargs=None) -> None:
        super().__init__(interval, function, args, kwargs)
        self.retry_cnt = 0
        self.name = name

    def run(self):
        need_log = True
        while not self.finished.is_set():
            if need_log:
                run_log.info("The repeating timer %s begin.", self.name)

            try:
                need_log: bool = self.function(*self.args, **self.kwargs)
            except Exception as err:
                run_log.error("The %s time run failed, caught exception: %s.", self.name, err)

            self.retry_cnt += 1
            if self.retry_cnt > self.MAX_RETRY_COUNT_PER_CYCLE:
                self.retry_cnt = 1

            wait_time = self.interval * self.retry_cnt

            if need_log:
                run_log.info("The repeating timer %s will wait for %s seconds.", self.name, wait_time)

            self.finished.wait(wait_time)

            if need_log:
                run_log.info("The repeating timer %s wait for %s seconds end.", self.name, wait_time)
