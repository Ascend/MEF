# !/usr/bin/python
# -*- coding: UTF-8 -*-
# 版权所有 (c) 华为技术有限公司 2023-2023
from typing import Iterable

from bin.environ import Env
from common.checkers.param_checker import ExtendedDeviceChecker
from common.log.logger import run_log
from lib.Linux.systems.disk.device_loader import DeviceLoader
from lib.Linux.systems.disk.errors import DevError
from lib.Linux.systems.disk.schema import ExtendCls, FormatInfo
from common.common_methods import CommonMethods


class Extend:
    def __init__(self):
        self.Name = None
        self.DeviceClass = None
        self.DeviceName = None
        self.Manufacturer = None
        self.Model = None
        self.SerialNumber = None
        self.FirmwareVersion = None
        self.Location = None
        self.Status = None
        self.items = []

    @staticmethod
    def valid_fmt_info_generator() -> Iterable[FormatInfo]:
        for fmt_info in DeviceLoader().fmt_info_generator():
            if ExtendCls.invalid(fmt_info, Env().start_from_m2):
                continue
            yield fmt_info

    def get_all_info(self, ext_dev_name=None) -> list:
        if ext_dev_name is None:
            try:
                self.items = [fmt.name for fmt in self.valid_fmt_info_generator()]
            except Exception as err:
                msg = str(err) if isinstance(err, DevError) else f"catch {err.__class__.__name__}"
                run_log.error("get device info by %s failed, %s", ext_dev_name, msg)
                return [CommonMethods.ERROR, "get extend device list failed."]

            return [0, "get extend device list success."]

        ret = ExtendedDeviceChecker().check_dict({"extend_id": ext_dev_name})
        if not ret:
            run_log.error("ext_dev_name invalid: %s", ret.reason)
            return [CommonMethods.ERROR, "ext_dev_name invalid."]

        try:
            fmt_info = DeviceLoader().get_fmt_info(ext_dev_name, "name")
        except Exception as err:
            msg = str(err) if isinstance(err, DevError) else f"catch {err.__class__.__name__}"
            run_log.error("get device info by %s failed, %s", ext_dev_name, msg)
            return [CommonMethods.NOT_EXIST, f"{ext_dev_name} is fail"]

        if ExtendCls.invalid(fmt_info, Env().start_from_m2):
            run_log.warning("%s invalid.", ext_dev_name)
            return [CommonMethods.ERROR, f"{ext_dev_name} invalid."]

        self.Name = fmt_info.name
        self.DeviceClass = fmt_info.cls_name
        self.DeviceName = fmt_info.path
        self.Manufacturer = fmt_info.vendor
        self.Model = fmt_info.model
        self.SerialNumber = fmt_info.serial_number
        self.FirmwareVersion = fmt_info.fw_version
        self.Location = fmt_info.location
        self.Status = fmt_info.status_dict
        return [CommonMethods.OK, "get extend info success."]
