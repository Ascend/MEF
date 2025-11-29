# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

import os
import sys

from common.file_utils import FileCheck
from common.log.logger import run_log
from lib.Linux.systems.disk.device_loader import DeviceLoader


def load_storage_configs():
    DeviceLoader.load_dev_configs()
    if not DeviceLoader.cfg:
        return 0

    ret = FileCheck.check_input_path_valid("/run/formated_hw.info")
    if not ret:
        run_log.error("/run/format_hw.info is invalid, ret: %s", ret.error)
        return 1
    try:
        with os.fdopen(os.open("/run/formated_hw.info", os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600), "w") as file:
            file.writelines(f"{dev_cfg.to_string()}\n" for dev_cfg in DeviceLoader.cfg.dev_configs
                            if dev_cfg.cls_name in ("u-disk", "eMMC", "DISK"))

            file.write(f"usb_hub_id={DeviceLoader.cfg.usb_hub_id}\n")
    except Exception as err:
        run_log.error(err)
        return 1

    return 0


if __name__ == "__main__":
    try:
        sys.exit(load_storage_configs())
    except Exception as error:
        run_log.error("load storage config failed, reason: %s", error)
        sys.exit(1)
