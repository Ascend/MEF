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
from logging import Logger
from typing import Iterable, Dict, Tuple, Any, NoReturn, List

from common.kmc_lib.kmc import KeyType, Kmc
from common.kmc_lib.kmc import KmcError

MK_UP_NUMS = 100
LATEST_KEY_CREATE_TIME_INTERVAL = 60


class Adapter:
    """用于适配不同类型数据的加、解密过程"""

    # 解密后的数据容器
    decrypted: Dict

    def __init__(self, kmc: Kmc):
        self.kmc = kmc

    def decrypted_generator(self) -> Iterable[Tuple[Any, Any]]:
        pass

    def decrypt(self) -> NoReturn:
        self.decrypted = dict(self.decrypted_generator())

    def encrypt(self) -> NoReturn:
        pass


class KeyUpdateExecutor:
    """更新密钥工具类"""

    def __init__(self, advance_update_rk_days, advance_update_mk_days, task_name: str, adapters: List[Adapter],
                 kmc: Kmc, logger: Logger):
        self._advance_update_rk_days = advance_update_rk_days
        self._advance_update_mk_days = advance_update_mk_days
        self.task_name: str = task_name
        self.kmc = kmc
        self.log = logger
        self.adapters: List[Adapter] = adapters

    def judge_key_is_expire(self, key_type):
        return self.kmc.get_interval_since_key_expire(key_type) > 0

    def instant_update_key_process(self, key_type):
        create_time_before = self.kmc.get_key_latest_create_time(key_type)
        self.log.info(create_time_before)

        if key_type == KeyType.ROOT_KEY.value:
            self.kmc.instant_update_rk()

        if key_type == KeyType.MASTER_KEY.value:
            self.kmc.instant_update_mk()

        create_time_after = self.kmc.get_key_latest_create_time(key_type)
        if create_time_before != create_time_after:
            self.log.info("Instantly update the %s success.", key_type)
        else:
            self.log.warning("It's not time to instantly update the %s.", key_type)

    def update_key_process(self, key_type):
        create_time_interval = self.kmc.get_interval_since_key_create(key_type)
        if create_time_interval >= LATEST_KEY_CREATE_TIME_INTERVAL:
            create_time_before = self.kmc.get_key_latest_create_time(key_type)

            if key_type == KeyType.ROOT_KEY.value:
                self.kmc.update_rk(self._advance_update_rk_days)

            if key_type == KeyType.MASTER_KEY.value:
                self.kmc.update_mk(self._advance_update_mk_days)

            create_time_after = self.kmc.get_key_latest_create_time(key_type)
            if create_time_before != create_time_after:
                self.log.info("Update the [%s] %s success.", self.task_name, key_type)
            else:
                self.log.warning("It's not time to update the %s.", key_type)
        else:
            self.log.warning("Updating the [%s] %s too quickly! Please wait!", self.task_name, key_type)

    def update_instance_key_while_expire(self, key_type: str):
        if self.judge_key_is_expire(key_type):
            self.instant_update_key_process(key_type)

    def update_root_key(self):
        self.log.info("Update [%s] root key.", self.task_name)
        self.update_instance_key_while_expire(KeyType.ROOT_KEY.value)
        self.update_key_process(KeyType.ROOT_KEY.value)

    def decrypt(self) -> NoReturn:
        for adapter in self.adapters:
            try:
                adapter.decrypt()
            except Exception as err:
                adapter.decrypted.clear()
                self.log.error("%s decrypt failed, catch %s", adapter.__class__.__name__, err.__class__.__name__)
            self.log.info("%s decrypt finish.", adapter.__class__.__name__)

    def encrypt(self) -> NoReturn:
        for adapter in self.adapters:
            try:
                adapter.encrypt()
            except Exception as err:
                self.log.error("%s encrypt failed, catch %s", adapter.__class__.__name__, err.__class__.__name__)
            adapter.decrypted.clear()
            self.log.info("%s encrypt finish.", adapter.__class__.__name__)

    def reset_master_key(self):
        key_type = KeyType.MASTER_KEY.value
        self.update_instance_key_while_expire(key_type)

        # 先将self.kmc加密过的所有数据解密
        self.log.info("Start decrypt...")
        self.decrypt()
        self.log.info("Load and decrypt ciphered text success.")

        # 去掉旧的主密钥
        self.kmc.check_and_remove_oldest_mk()
        self.log.info("Remove the oldest master key success.")
        self.update_key_process(key_type)

        # 使用更新过的kmc再次加密保存
        self.log.info("Start encrypt...")
        self.encrypt()
        self.log.info("Upload and encrypt ciphered text success.")

    def update_master_key(self):
        key_type = KeyType.MASTER_KEY.value
        if self.kmc.get_mk_count() < MK_UP_NUMS:
            self.log.info("Update [%s] mask key.", self.task_name)
            self.update_instance_key_while_expire(key_type)
            self.update_key_process(key_type)
            return

        self.log.warning("The number of [%s] master keys are up to limits", self.task_name)
        self.reset_master_key()

    def execute_update_key(self):
        if not self.kmc.available():
            self.log.error("kmc not available.")
            return

        self.log.info("[%s] process for checking and updating the root/master key start!", self.task_name)
        try:
            self.update_root_key()
        except KmcError as err:
            self.log.error("Update root key failed, err: %s", err)

        try:
            self.update_master_key()
        except KmcError as err:
            self.log.error("Update master key failed, err: %s", err)

        self.log.info("[%s] process for checking and updating the root/master key end!", self.task_name)
