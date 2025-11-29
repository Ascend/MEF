# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from functools import cached_property
from pathlib import Path

from common.file_utils import FileCheck
from common.log.logger import run_log
from common.utils.singleton import Singleton
from lib.Linux.systems.disk import contants as const
from lib.Linux.systems.disk.device_loader import DeviceLoader
from lib.Linux.systems.disk.errors import DevError
from lib.Linux.systems.disk.schema import Part


class Env(Singleton):
    """用于获取系统环境相关参数"""

    @cached_property
    def sys_root(self) -> str:
        """
        获取系统启动root=后的内容：
            EMMC启动时：返回/dev/mmcblk0p2或/dev/mmcblk0p3
            M.2启动时：返回PARTUUID=对应id
        """
        ret = FileCheck.check_path_is_exist_and_valid(const.CMDLINE_PATH)
        if not ret:
            run_log.error("cmdline invalid. %s", ret.error)
            raise OSError("cmdline invalid.")

        with open(const.CMDLINE_PATH) as cmdline:
            for content in cmdline.readline().split():
                if content.lstrip().startswith("root="):
                    return content.lstrip().lstrip("root=")

        run_log.error("Get sys root failed")
        raise ValueError("Get sys root failed")

    @cached_property
    def cur_partition(self) -> Path:
        """当前启动区"""
        if "=" not in self.sys_root:
            return Path(self.sys_root)

        if "PARTUUID=" in self.sys_root:
            return Path(const.PARTUUID_DIR, self.sys_root.split("=")[-1]).resolve()

        run_log.error("parse cmdline failed.")
        raise DevError("parse cmdline failed.")

    @cached_property
    def boot_part(self) -> Part:
        """系统启动区"""
        # 加载启动区时，先加载一下磁盘配置，monitor启动时会触发系统区加载，并捕获异常
        DeviceLoader.load_dev_configs()
        try:
            return DeviceLoader().load_part(self.cur_partition.as_posix())
        except Exception as err:
            msg = str(err) if isinstance(err, DevError) else f"catch {err.__class__.__name__}"
            raise DevError(f"Get system boot partition failed, {msg}.") from err

    @cached_property
    def cur_partition_num(self) -> int:
        """当前分区号"""
        return int(self.boot_part.part_num)

    @cached_property
    def back_partition_num(self) -> int:
        """备区分区号，要么p2要么p3"""
        return 3 if self.cur_partition_num == 2 else 2

    @cached_property
    def start_from_emmc(self) -> bool:
        return self.boot_part.disk_name.startswith(const.EMMC_PREFIX)

    @cached_property
    def start_from_m2(self) -> bool:
        return not self.start_from_emmc

    @cached_property
    def back_partition(self) -> Path:
        """获取备区"""
        return self.get_partition_by_label(f"p{self.back_partition_num}")

    def get_partition_by_label(self, label: str) -> Path:
        """根据分区名获取，label以p开头，例如p1"""
        if self.boot_part.site_name.endswith(self.boot_part.part_name):
            return Path(const.DEV_ROOT, "".join((self.boot_part.disk_site_name, label)))

        return Path(const.DEV_ROOT, "".join((self.boot_part.disk_site_name, label[1:])))
