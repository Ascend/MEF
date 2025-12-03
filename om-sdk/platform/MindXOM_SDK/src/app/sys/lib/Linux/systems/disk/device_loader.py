# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import json
import os
import shlex
from functools import partial
from pathlib import Path
from typing import Dict, Iterable, Optional, List

from common.constants.base_constants import CommonConstants
from common.file_utils import FileCheck
from common.log.logger import run_log
from common.utils.exec_cmd import ExecCmd
from common.utils.singleton import Singleton
from common.utils.system_utils import SystemUtils
from lib.Linux.systems.disk import contants as const
from lib.Linux.systems.disk.errors import DevError, DevNotFound
from lib.Linux.systems.disk.schema import (DevConfigs, FormatInfo, ExtendCls, DevCfg, UdevFmtInfo, Disk, Part,
                                           get_site_name)


class DeviceLoader(Singleton):
    """用于加载设备信息与配置，采用单例目的在于减少实例"""
    WAIT = 3
    exec_cmd = partial(ExecCmd.exec_cmd_get_output)
    cfg: Optional[DevConfigs] = None

    @staticmethod
    def _dev_cfg_generator(conf: dict) -> Iterable[DevCfg]:
        extend_info: Dict[str, List[dict]] = conf.get("extend_info")
        if not isinstance(extend_info, dict):
            raise DevError("device info invalid.")

        if extend_info and len(extend_info) > CommonConstants.MAX_ITER_LIMIT:
            raise DevError("device class is too large")

        for cls_name, dev_configs in extend_info.items():
            if not isinstance(dev_configs, List):
                raise DevError("disk config invalid.")

            if dev_configs and len(dev_configs) > CommonConstants.MAX_ITER_LIMIT:
                raise DevError("device info is too large")

            for dev_cfg in dev_configs:
                dev_cfg.update(cls_name=cls_name)
                yield DevCfg.from_dict(dev_cfg)

    @staticmethod
    def _a500_fmt_info_generator() -> Iterable[FormatInfo]:
        """a500从formated_hw_1.info读取设备信息"""
        ret = FileCheck.check_path_is_exist_and_valid(const.FORMAT_INFO)
        if not ret:
            raise DevError(f"check device config failed, {ret.error}.")
        if os.path.getsize(const.FORMAT_INFO) > CommonConstants.OM_READ_FILE_MAX_SIZE_BYTES:
            raise DevError("the size of formated_hw_1.info is too large.")
        with open(const.FORMAT_INFO) as fmt_file:
            for line in fmt_file.read().splitlines():
                items = line.split(";")
                if not items:
                    continue
                yield FormatInfo.from_dict(dict(item.partition(":")[::2] for item in items))

    @classmethod
    def load_dev_configs(cls):
        """系统启动时加载磁盘配置"""
        if SystemUtils().is_a500:
            return

        if os.path.getsize(CommonConstants.DEVICE_CFG_PATH) > CommonConstants.OM_READ_FILE_MAX_SIZE_BYTES:
            raise DevError("the size of device config is too large.")
        ret = FileCheck.check_path_is_exist_and_valid(CommonConstants.DEVICE_CFG_PATH)
        if not ret:
            raise DevError(f"check device config failed, {ret.error}")
        cfg = json.loads(Path(CommonConstants.DEVICE_CFG_PATH).read_text())
        cls.cfg = DevConfigs(dev_configs=list(cls._dev_cfg_generator(cfg)),
                             primary_partitions=cfg.get("primary_partitions"),
                             usb_hub_id=cfg.get("usb_hub_id"))

    def fmt_info_generator(self) -> Iterable[FormatInfo]:
        yield from self._a500_fmt_info_generator() if SystemUtils().is_a500 else self._a200_fmt_info_generator()

    def get_fmt_info(self, value: str, field: str = "path") -> FormatInfo:
        """查找具体的设备信息"""
        for fmt in self.fmt_info_generator():
            if getattr(fmt, field, "") == value:
                return fmt
        raise DevNotFound(f"unable to find format information of {value}")

    def get_disk_path_by_location(self, loc: str) -> str:
        dev_path = self.get_fmt_info(loc, "location").path
        return Path(const.DEV_ROOT, get_site_name(dev_path, self._load_dev_links(dev_path))).as_posix()

    def load_disk(self, dev_path_or_name: str) -> Disk:
        normal_dev_name = Path(const.DEV_ROOT, Path(dev_path_or_name).name).as_posix()
        block_info = self._load_block_info(normal_dev_name)
        disk = Disk.from_dict(dict(**block_info, **self._load_udev_info(normal_dev_name)))
        children = block_info.get("children") or []
        # block采用小写字母，udev采用大写，生成字典时不会字段冲突
        disk.children = [Part.from_dict(dict(**block, **self._load_udev_info(block.get("path")))) for block in children]
        return disk

    def load_part(self, dev_path_or_name: str) -> Part:
        normal_dev_name = Path(const.DEV_ROOT, Path(dev_path_or_name).name).as_posix()
        return Part.from_dict(dict(**self._load_block_info(normal_dev_name), **self._load_udev_info(normal_dev_name)))

    def _get_eth_version(self, eth: str) -> str:
        """a200 获取eth版本号"""
        drv_info = "ethtool", "-i", shlex.quote(eth)
        status, info = self.exec_cmd(drv_info)
        if status != 0:
            raise DevError(f"exec ethtool failed, {info}")
        for item in info.split("\n"):
            name, value = item.split(":")
            if name == "version":
                return value
        return ""

    def _load_block_info(self, dev: str) -> Dict:
        """执行lsblk获取block的基本信息"""
        block_info = "lsblk", "-bJO", shlex.quote(dev)
        status, info = self.exec_cmd(block_info)
        if status != 0:
            raise DevError("exec lsblk failed.")
        return json.loads(info)["blockdevices"][0]

    def _load_dev_links(self, dev_path: str) -> str:
        """加载分区或磁盘的链接"""
        udev_info = "udevadm", "info", "-w", "-q", "symlink", shlex.quote(dev_path)
        status, info = self.exec_cmd(udev_info, wait=self.WAIT)
        if status != 0:
            raise DevError("load dev links failed.")
        return info

    def _load_udev_info(self, dev: str) -> Dict:
        """查询设备udev信息"""
        # -w 的作用是等待设备初始化完成，该参数必须要，否则执行了partprobe后快速执行udevadm时，设备可能未初始化
        udev_info = "udevadm", "info", "-w", "-q", "property", shlex.quote(dev)
        status, info = self.exec_cmd(udev_info, wait=self.WAIT)
        if status != 0:
            raise DevError("exec udevadm failed.")
        return dict(line.split("=") for line in info.split("\n") if line)

    def _a200_fmt_info_generator(self) -> Iterable[FormatInfo]:
        """a200根据配置查询其他命令转换出设备信息"""
        for cfg in self.cfg.dev_configs:
            is_net = cfg.cls_name == ExtendCls.ETH.value
            try:
                dev_path = cfg.get_dev_path(is_net)
            except DevNotFound:
                # 有配置不一定存在对应的设备，找不到则忽略
                continue
            if not is_net:
                udev = UdevFmtInfo.from_dict(self._load_udev_info(dev_path))
                yield FormatInfo(
                    name=cfg.name, cls_name=cfg.cls_name, serial_number=udev.serial_number, path=dev_path,
                    vendor=udev.vendor, location=cfg.location, fw_version=udev.version,
                    model=udev.id_model or udev.id_name or "N/A"
                )
                continue

            status = const.STATUS_ACTIVE
            try:
                version = self._get_eth_version(Path(dev_path).name)
            except Exception as err:
                msg = str(err) if isinstance(err, DevError) else f"catch {err.__class__.__name__}"
                run_log.error("get eth version failed, %s", msg)
                version = "N/A"
                status = const.STATUS_INACTIVE
            yield FormatInfo(
                name=cfg.name, cls_name=cfg.cls_name, location=cfg.location, path=dev_path, fw_version=version,
                status=status
            )
