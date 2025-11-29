# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import os.path
from functools import partial
from typing import Iterable, Type, Dict, List

from common.kmc_lib.kmc import Kmc
from common.log.logger import run_log
from common.utils.timer import RepeatingTimer
from common.kmc_lib.update_kmc_key import Adapter, KeyUpdateExecutor


class KmcUpdater:
    AUTO_ADVANCE_UPDATE_RK_DAYS = 15
    AUTO_ADVANCE_UPDATE_MK_DAYS = 15
    MANUAL_ADVANCE_UPDATE_RK_DAYS = 3650
    MANUAL_ADVANCE_UPDATE_MK_DAYS = 180
    UPDATE_KEY_INTERVAL = 3600 * 24  # 更新间隔时间1天

    def __init__(self, ks_dir: str, task_name: str, ks_name="om_cert", alg="om_alg.json"):
        self.kmc = Kmc(
            os.path.join(ks_dir, f"{ks_name}.keystore"),
            os.path.join(ks_dir, f"{ks_name}_backup.keystore"),
            os.path.join(ks_dir, alg),
        )
        self.task_name = task_name
        self.adapters: List[Adapter] = []
        self.executor = partial(KeyUpdateExecutor, task_name=task_name, kmc=self.kmc, logger=run_log)

    def extend_adapters(self, adapters: Iterable[Type[Adapter]]):
        self.adapters.extend(adapter(self.kmc) for adapter in adapters)

    def interval_update(self):
        executor = self.executor(
            advance_update_rk_days=self.AUTO_ADVANCE_UPDATE_RK_DAYS,
            advance_update_mk_days=self.AUTO_ADVANCE_UPDATE_MK_DAYS,
            adapters=self.adapters,
        )
        RepeatingTimer(self.UPDATE_KEY_INTERVAL, executor.execute_update_key).start()

    def manual_update(self):
        executor = self.executor(
            advance_update_rk_days=self.MANUAL_ADVANCE_UPDATE_RK_DAYS,
            advance_update_mk_days=self.MANUAL_ADVANCE_UPDATE_MK_DAYS,
            adapters=self.adapters,
        )
        executor.execute_update_key()


class MultiKmcUpdater:
    """用户维护多个kmc的密钥更新, 便于支持其他仓进行扩展"""
    updaters: Dict[str, KmcUpdater] = {}

    @classmethod
    def register_updaters(cls, *updaters: KmcUpdater):
        for updater in updaters:
            cls.updaters[updater.task_name] = updater

    @classmethod
    def extend_adapters(cls, name: str, *adapters: Type[Adapter]):
        """支持添加扩展的适配器"""
        if name in cls.updaters:
            cls.updaters[name].extend_adapters(adapters)

    @classmethod
    def start_auto_update_key_tasks(cls):
        for updater in cls.updaters.values():
            updater.interval_update()

    @classmethod
    def manual_update_key(cls):
        for updater in cls.updaters.values():
            updater.manual_update()
