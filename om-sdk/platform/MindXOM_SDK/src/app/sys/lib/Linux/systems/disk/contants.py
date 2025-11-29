# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

CMDLINE_PATH = "/proc/cmdline"
EMMC_PREFIX = "mmcblk"
EMMC0 = f"{EMMC_PREFIX}0"
SYS_ROOT = "/sys"
BLOCK_DIR = "/sys/block"
NET_ROOT = "/sys/class/net"
DEV_ROOT = "/dev"
PARTUUID_DIR = "/dev/disk/by-partuuid"
PLATFORM_ROOT = "/sys/devices/platform"
FORMAT_INFO = "/run/formated_hw_1.info"

MAX_PARTITION_NUM = 16
A200_PRIMARY = 1, 2, 3, 4, 5
A500_PRIMARY = 1, 2, 3, 4, 5, 6, 7, 99

STORAGE_NAME = "sata", "usb", "sdhci"
STORAGE_ID = "1", "2", "3"
STORAGE_NAME_ID_MAP = dict(zip(STORAGE_NAME, STORAGE_ID))
STORAGE_ID_NAME_MAP = dict(zip(STORAGE_ID, STORAGE_NAME))

STATE_GOOD = {"State": "Enabled", "Health": "OK"}
STATE_BAD = {"State": "Disabled", "Health": "Critical"}
STATUS_ACTIVE = "active"
STATUS_INACTIVE = "inactive"
# 只有sata会查询alarm
LOCATION_ALARM_MAP = {"PCIE3-0": "M.2", "PCIE3-1": "HARD_DISK0", "PCIE3-2": "HARD_DISK1"}
EMMC_DRV_LIB = "/usr/lib64/libdrvemmc.so"

KILO_BINARY_BYTE_RATE = 1024
FREE_RESERVED_SIZE = 100 * KILO_BINARY_BYTE_RATE ** 2
# eMMC的p1-p7，p99分区的总大小，为与FD显示一致，需加上存储预留100M空间
RESERVED_PARTITION_SIZE = 22832742400 + FREE_RESERVED_SIZE
