# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import ctypes
from collections import defaultdict
from itertools import islice
from pathlib import Path
from typing import Iterable, Dict, Tuple, List

from bin.environ import Env
from common.constants.base_constants import CommonConstants
from common.log.logger import run_log
from common.utils.system_utils import SystemUtils
from lib.Linux.systems.disk import contants as const
from lib.Linux.systems.disk.device_loader import DeviceLoader
from lib.Linux.systems.disk.errors import DevError
from lib.Linux.systems.disk.schema import Disk
from common.common_methods import CommonMethods


class Storage:
    def __init__(self):
        self.Name = None
        self.Description = None
        self.Status = None
        self.Devices = []
        self.items = []

    @staticmethod
    def _get_life_time() -> int:
        if not Path(const.EMMC_DRV_LIB).exists():
            raise DevError("drive lib not exists.")

        lib_op = ctypes.CDLL(const.EMMC_DRV_LIB)
        life_time_clib = lib_op.get_emmc_extcsd_info
        life_time_clib.restype = ctypes.c_int
        life_time_clib.argtypes = [ctypes.POINTER(ctypes.c_int)]
        life_time = ctypes.c_int(255)
        ret = life_time_clib(ctypes.pointer(life_time))
        if ret == 0:
            return life_time.value
        raise DevError("get life time failed.")

    @staticmethod
    def _get_storage_id(platform) -> str:
        for name, storage_id in const.STORAGE_NAME_ID_MAP.items():
            if name in platform:
                return storage_id
        raise DevError(f"platform {platform} not supported.")

    @staticmethod
    def _sys_blocks() -> Iterable[Path]:
        for path in islice(Path(const.BLOCK_DIR).glob("*"), CommonConstants.MAX_ITER_LIMIT):
            # 以mmcblk0boot开头的并非实际的磁盘，需过滤掉
            if path.name.startswith(f"{const.EMMC0}boot"):
                continue
            # A500从M.2启动时，不能呈现eMMC启动盘
            if all((SystemUtils().is_a500, Env().start_from_m2, path.name == const.EMMC0)):
                continue
            yield path

    def get_storages(self) -> Dict[str, List[Path]]:
        storages = defaultdict(list)
        for storage_id, block_dir in self._storage_generator():
            storages[storage_id].append(block_dir)
        return storages

    def get_all_info(self, storage_id=None) -> list:
        storages = self.get_storages()
        self.items = sorted(storages)
        if storage_id is None:
            return [CommonMethods.OK, ]

        if storage_id not in self.items:
            return [CommonMethods.NOT_EXIST, f"{storage_id} not in storages"]

        self._set_storage_info(storage_id, storages)
        return [CommonMethods.OK, ]

    def _storage_generator(self) -> Iterable[Tuple[str, Path]]:
        for block_dir in self._sys_blocks():
            dev_path = block_dir.resolve()
            if not dev_path.is_relative_to(const.PLATFORM_ROOT):
                continue
            platform = dev_path.relative_to(const.PLATFORM_ROOT).parts[0]
            if not SystemUtils().is_a500 and platform not in DeviceLoader.cfg.platforms:
                continue
            try:
                yield self._get_storage_id(platform), block_dir
            except DevError as err:
                run_log.warning(err)

    def _set_storage_info(self, storage_id: str, storages: Dict[str, List[Path]]):
        self.Devices.clear()
        self.Status = const.STATE_GOOD
        self.Name = const.STORAGE_ID_NAME_MAP.get(storage_id)
        for storage in storages.get(storage_id, []):
            try:
                storage_info = self._get_storage_info(DeviceLoader().load_disk(storage.as_posix()))
            except Exception as err:
                msg = str(err) if isinstance(err, DevError) else f"catch {err.__class__.__name__}"
                run_log.warning("load block %s failed, %s, ignore.", storage.name, msg)
                continue
            self.Description = f"System {self.Name} Flash"
            self.Devices.append(storage_info)
            if storage_info.get("Status") != const.STATE_GOOD:
                self.Status = const.STATE_BAD

    def _get_storage_info(self, disk: Disk) -> Dict:
        """返回给前端的device字典"""
        life_time = None
        if disk.name.startswith(const.EMMC_PREFIX):
            try:
                life_time = self._get_life_time()
            except Exception as err:
                run_log.warning("get life time failed, catch %s", err)
        fmt_info = DeviceLoader().get_fmt_info(disk.path)
        return {
            "Name": disk.path.replace(disk.name, disk.site_name),
            "Manufacturer": fmt_info.vendor,
            "Model": fmt_info.model,
            "CapacityBytes": disk.capacity(),
            "LeftBytes": disk.free(),
            "PartitionStyle": {"gpt": "GPT", "dos": "MBR"}.get(disk.pt_type, ""),
            "Location": fmt_info.location,
            "DeviceLifeTimeUsed": life_time,
            "Status": fmt_info.status_dict,
        }
