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
import time
from threading import Lock

from common.log.logger import run_log


class TokenBucket:
    """令牌桶算法类。

    属性：
        bucket_capacity（int）：令牌桶的总容量。
        generate_rate（int）：令牌生成速率。
        current_token_amount（int）：当前令牌桶中的令牌数。
        last_consume_time （int）：上次消费时间。
        token_bucket_lock（obj：threading.Lock）：令牌桶线程锁。

    """
    def __init__(self, capacity: int = 20, rate: int = 10):
        """容量是存储桶中的令牌总数。速率是桶将被重新填充的速率（以令牌/秒为单位）。

        注：
            初始化令牌桶容量和令牌生成速率。

        参数说明：
            capacity（int）：令牌桶的总容量，默认值为20。
            rate（int）：令牌桶令牌生成速率，默认值为10/s。

        """

        self.bucket_capacity = capacity
        self.generate_rate = rate
        # 当前令牌数量初始化为桶容量，保证初始化之后能立即消费
        self.current_token_amount = capacity
        self.last_consume_time = int(time.perf_counter())
        self.token_bucket_lock = Lock()

    def consume_token(self, consume_token_amount: int) -> bool:
        """从令牌桶中消耗令牌。如果有足够的令牌，则返回True，否则返回False。"""
        if self.token_bucket_lock.locked():
            run_log.warning("Token bucket lock is locked. Please wait for it to release.")
            return False

        with self.token_bucket_lock:
            # 新发放的令牌数量 = (本次消费时间 - 上一次消费时间) * 令牌发放速率
            cur_consume_time = int(time.perf_counter())
            token_increment = (cur_consume_time - self.last_consume_time) * self.generate_rate
            # 当前令牌数量不能超过桶的容量，取‘桶容量与新发放的令牌数量+当令牌数量’中的最小值
            self.current_token_amount = min(self.bucket_capacity, token_increment + self.current_token_amount)
            # 消费令牌数 > 当前令牌数，表示没有足够的令牌，则不能发送数据
            if consume_token_amount > self.current_token_amount:
                run_log.warning("Consume tokens amount be more current tokens amount. "
                                "Please wait for it to be released.")
                # 上一次消费时间 = 当前时间
                self.last_consume_time = cur_consume_time
                return False

            # 当前令牌数量大于或等于需要消费的令牌数量，当前令牌数 = 当前令牌数 - 消费令牌数
            self.current_token_amount -= consume_token_amount
            # 上一次消费时间 = 当前时间
            self.last_consume_time = cur_consume_time
            return True
