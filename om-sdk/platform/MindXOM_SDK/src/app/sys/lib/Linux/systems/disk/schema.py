# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import os
from dataclasses import dataclass
from enum import Enum
from itertools import islice
from pathlib import Path
from typing import List, Tuple, Union

from common.constants.base_constants import CommonConstants
from common.log.logger import run_log
from common.schema import BaseModel, field
from lib.Linux.systems.Alarm.alarm import Alarm
from lib.Linux.systems.disk import contants as const
from lib.Linux.systems.disk.errors import SpaceNotEnough, DevError, DevNotFound


@dataclass
class DevCfg(BaseModel):
    cls_name: str = field(comment="类名")
    name: str = field(comment="设备名")
    platform: str = field(comment="/sys/devices/platform下的名字")
    location: str = field(comment="设备位置表示")
    device: str = field(default="", comment="设备device标识")

    def fit_device_path(self, device: Path) -> bool:
        dev_path = device.resolve()
        if not dev_path.is_relative_to(const.PLATFORM_ROOT):
            return False

        if dev_path.relative_to(const.PLATFORM_ROOT).parts[0] != self.platform:
            return False

        # 没配device，platform匹配成功
        if not self.device:
            return True

        return self.device == dev_path.joinpath("device").resolve().name

    def get_dev_path(self, is_net: bool = False) -> str:
        dev_root = const.NET_ROOT if is_net else const.BLOCK_DIR
        for dev_path in islice(Path(dev_root).glob("*"), CommonConstants.MAX_ITER_LIMIT):
            if self.fit_device_path(dev_path):
                return dev_path.resolve().as_posix() if is_net else Path(const.DEV_ROOT, dev_path.name).as_posix()

        raise DevNotFound("not found dev path by config.")

    def to_string(self):
        """将设备配置转换成string后需要写入到/run/format_hw.info中，供告警使用"""
        return ",".join(f"{_field}={value}" for _field, value in self.to_dict().items())


@dataclass
class DevConfigs(BaseModel):
    """A200设备配置"""
    dev_configs: List[DevCfg] = field(comment="磁盘配置列表")
    primary_partitions: List[int] = field(default=const.A200_PRIMARY, comment="主分区号")
    platforms: List[str] = field(init=False, default=None, comment="配置的所有platform")
    usb_hub_id: str = field(default=None, comment="USB HUB的ID")

    def __post_init__(self):
        self.valid_primary_partitions()
        self.platforms = [cfg.platform for cfg in self.dev_configs]

    def valid_primary_partitions(self):
        if any((
                not isinstance(self.primary_partitions, list),
                not self.primary_partitions,
                len(self.primary_partitions) > const.MAX_PARTITION_NUM,
                any(not isinstance(num, int) for num in self.primary_partitions),
        )):
            raise DevError("primary partitions invalid.")


@dataclass
class FormatInfo(BaseModel):
    name: str = field(comment="设备名")
    cls_name: str = field(alias="class", comment="设备类别")
    path: str = field(alias="devname", comment="设备路径, 例如/dev/sda")
    serial_number: str = field(alias="serialnumber", default="N/A", comment="序列号")
    vendor: str = field(default="N/A", comment="驱动转换后的厂商名")
    model: str = field(default="N/A", comment="型号")
    location: str = field(default="N/A", comment="设备位置")
    fw_version: str = field(default="N/A", comment="版本号")
    status: str = field(default=const.STATUS_ACTIVE, comment="状态")
    status_dict: dict = field(init=False, comment="接口返回的状态字典")

    def __post_init__(self):
        self.status_dict = const.STATE_GOOD if self.status == const.STATUS_ACTIVE else const.STATE_BAD
        if const.LOCATION_ALARM_MAP.get(self.location) in Alarm.getallalarmdev():
            self.status_dict = const.STATE_BAD


@dataclass
class UdevMonitorInfo(BaseModel):
    """执行udevadm monitor时使用的数据结构"""
    path: str = field(alias="DEVNAME", comment="设备路径, 例如/dev/sda")
    dev_path: str = field(default="", alias="DEVPATH", comment="设备路径，devices下的")
    action: str = field(default="", alias="ACTION", comment="monitor状态，DevAction中的一种")
    pt_type: str = field(default="", alias="ID_PART_TABLE_TYPE", comment="磁盘类型")


@dataclass
class UdevFmtInfo(BaseModel):
    """生成A200的FormatInfo时，需要解析udevadm info的数据结构"""
    id_name: str = field(default="", alias="ID_NAME", comment="用于解析model")
    id_model: str = field(default="", alias="ID_MODEL", comment="用于解析model")
    serial_number: str = field(default="", alias="ID_SERIAL", comment="序列号")
    version: str = field(default="1", alias="ID_REVISION", comment="版本号")
    vendor: str = field(default="unknown", alias="ID_VENDOR", comment="厂商")


@dataclass
class Block(BaseModel):
    """磁盘与分区基类"""
    # lsblk中获取的字段
    name: str = field(comment="设备名")
    path: str = field(comment="设备路径, 例如/dev/sda")
    size: str = field(comment="设备大小，bytes")

    # udevadm info中获取的字段
    dev_path: str = field(default="", alias="DEVPATH", comment="设备路径，devices下的")
    links: str = field(default="", alias="DEVLINKS", comment="设备链接")

    # 转换出来的字段
    site_name: str = field(init=False, default="", comment="磁盘或分区显示名")
    storage_name: str = field(init=False, default="", comment="磁盘或分区对应的存储显示名")

    def __post_init__(self):
        self.site_name = get_site_name(self.path, self.links)
        self.set_storage_name()

    def set_storage_name(self):
        dev_path = Path(const.SYS_ROOT, *self.dev_path.split(os.sep))
        platform = dev_path.relative_to(const.PLATFORM_ROOT).parts[0]
        for name in const.STORAGE_NAME:
            if name in platform:
                self.storage_name = name
                break


@dataclass
class Part(Block):
    # lsblk中获取的字段
    fs_avail: str = field(default="", alias="fsavail", comment="文件系统可用容量，bytes")
    fs_type: str = field(default="unknown", alias="fstype", comment="文件系统类型")
    sector_size: str = field(default="512", alias="log-sec", comment="单个sector的大小")
    mount_point: str = field(default="", alias="mountpoint", comment="挂载点")
    disk_name: str = field(default="", alias="pkname", comment="分区对应的磁盘名")

    # udevadm info中获取的字段
    part_num: str = field(default="", alias="PARTN", comment="分区号，例如1")
    offset: str = field(default="", alias="ID_PART_ENTRY_OFFSET", comment="分区起始位,sector")

    # 转换出来的字段
    part_name: str = field(init=False, default="", comment="分区名，例如p1")
    disk_site_name: str = field(init=False, default="", comment="磁盘显示名")
    disk_path: str = field(init=False, comment="磁盘路径")

    def __post_init__(self):
        super().__post_init__()
        # udevadm info中的分区名可能为p{part_num}也可能为primary，需要转换成统一的p{part_num}格式
        self.part_name = f"p{self.part_num}"
        self.disk_path = Path(const.DEV_ROOT, self.disk_name).as_posix()
        # site_name由驱动脚本做的软连接，要么转换前后一致，要么不一致，不一致的情况下均会以part_name结尾
        self.disk_site_name = self.disk_name
        if self.site_name.endswith(self.part_name):
            self.disk_site_name = self.site_name.rpartition(self.part_name)[0]

    def free(self) -> int:
        return int(self.fs_avail or self.size)

    def offset_kib(self) -> Tuple[int, int]:
        start = int(self.sector_size) * int(self.offset)
        return int(start / const.KILO_BINARY_BYTE_RATE), int((start + int(self.size)) / const.KILO_BINARY_BYTE_RATE)


@dataclass
class Disk(Block):
    # lsblk中获取的字段
    pt_type: str = field(default="", alias="pttype", comment="磁盘类型")
    children: List[Part] = field(default=None, comment="分区")

    def free(self) -> int:
        # 磁盘容量为总容量 - 分区占用量 - 预留量
        return int(self.size) - sum(int(child.size) for child in self.children) - const.FREE_RESERVED_SIZE

    def capacity(self) -> int:
        if self.name == const.EMMC0:
            return int(self.size) - const.RESERVED_PARTITION_SIZE
        return int(self.size)

    def offset_for_new_part(self, size_kib: Union[int, float], boot_part: Part) -> Tuple[float, float]:
        """新增分区的偏移量计算"""
        # 预留10MiB
        reserved = 10 * const.KILO_BINARY_BYTE_RATE
        end = reserved + size_kib
        # 无分区时, 尾部预留10MiB不分配
        if not self.children and end <= int(self.size) / const.KILO_BINARY_BYTE_RATE - reserved:
            return end - size_kib, end

        children = sorted(self.children, key=lambda part: int(part.offset))
        # 分配在头部，如果是启动盘不允许分配到头部
        if self.name != boot_part.disk_name and end <= children[0].offset_kib()[0]:
            run_log.info("try malloc disk in start")
            return end - size_kib, end

        # 分配在尾部，预留10MiB不分配
        end = children[-1].offset_kib()[-1] + size_kib
        if end <= int(self.size) / const.KILO_BINARY_BYTE_RATE - reserved:
            run_log.info("try malloc disk in end")
            return end - size_kib, end

        # 尝试分配到中部
        for index, child in enumerate(children[:-1]):
            end = child.offset_kib()[-1] + size_kib
            if end <= children[index + 1].offset_kib()[0]:
                run_log.info("try malloc disk in middle")
                return end - size_kib, end

        raise SpaceNotEnough("do not have sufficient space")


class DevAction:
    ADD = "add"
    CHANGE = "change"
    RM = "remove"


class SubSystem:
    """支持udev事件子系统过滤的类型"""
    DISK = "block/disk"
    PART = "block/partition"


class ExtendCls(Enum):
    EMMC = "eMMC"
    USB = "u-disk"
    ETH = "ethnet"
    DISK = "DISK"
    WIRELESS = "Wireless_Module"

    @classmethod
    def invalid(cls, fmt_info: FormatInfo, is_m2: bool) -> bool:
        return any((
            fmt_info.cls_name not in tuple(item.value for item in cls),
            is_m2 and fmt_info.name == cls.EMMC.value,  # M.2启动不能展示eMMC
        ))


def get_site_name(dev_path: str, udev_links: str) -> str:
    """根据udevadm查出的links转换成对应dev_path(磁盘或分区)的显示名"""
    lnk_list = [lnk for lnk in udev_links.strip().split(" ") if Path(const.DEV_ROOT, Path(lnk).name).exists()]
    return Path(lnk_list[0]).name if lnk_list else Path(dev_path).name
