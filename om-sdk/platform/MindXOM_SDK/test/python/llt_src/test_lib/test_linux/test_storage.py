# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from pathlib import Path
from typing import NamedTuple, Iterable, Tuple, Dict, List, Optional

import pytest
from pytest_mock import MockerFixture

from lib.Linux.systems.disk import contants as const
from lib.Linux.systems.disk.errors import DevError
from lib.Linux.systems.disk.storage import Storage

EMMC = "mmcblk0", "82000000.sdhci0/mmc_host/mmc0/mmc0:0001/block/mmcblk0"
SD_CARD = "mmcblk1", "82000000.sdhci1/mmc_host/mmc0/mmc0:0001/block/mmcblk1"
BOOT0 = "mmcblk0boot0", "82000000.sdhci0/mmc_host/mmc0/mmc0:0001/block/mmcblk0/mmcblk0boot0"
BOOT1 = "mmcblk0boot1", "82000000.sdhci0/mmc_host/mmc0/mmc0:0001/block/mmcblk0/mmcblk0boot1"
M2 = "sda", "a6000000.sata/ata1/host0/target0:0:0/0:0:0:0/block/sda"
USB = "sdb", "a5180000.hiusbc3/xhci-hcd.2.auto/usb5/5-1/5-1:1.0/host5/target5:0:0/5:0:0:0/block/sdb"
BLOCKS = EMMC, SD_CARD, BOOT1, BOOT0, M2, USB


class GetStorageIdModel(NamedTuple):
    expect: str
    platform: str


class SysBlocksModel(NamedTuple):
    expect: Iterable[Tuple[str, str]]
    is_a500: bool
    start_from_m2: bool


class GetStoragesModel(NamedTuple):
    expect: Iterable[str]
    is_a500: bool
    start_from_m2: bool
    platforms: Iterable[str]


class SetStorageInfoModel(NamedTuple):
    expect: Tuple[Optional[str], Dict]
    storage_id: str
    storages: Dict[str, List[Path]]
    storage_info: Dict


class TestStorage:
    use_cases = {
        "test_get_storage_id": {
            "platform-sata": GetStorageIdModel(expect="1", platform="*sata*"),
            "platform-usb": GetStorageIdModel(expect="2", platform="*usb*"),
            "platform-sdhci": GetStorageIdModel(expect="3", platform="*sdhci*"),
        },
        "test_sys_blocks": {
            "a500_emmc": SysBlocksModel(expect=(EMMC, SD_CARD, M2, USB), is_a500=True, start_from_m2=False),
            "a500_m.2": SysBlocksModel(expect=(SD_CARD, M2, USB), is_a500=True, start_from_m2=True),
            "a200": SysBlocksModel(expect=(EMMC, SD_CARD, M2, USB), is_a500=False, start_from_m2=False),
        },
        "test_get_storages": {
            "a500_emmc": GetStoragesModel(expect=["1", "2", "3"], is_a500=True, start_from_m2=False, platforms=[]),
            "a500_m.2": GetStoragesModel(expect=["1", "2"], is_a500=True, start_from_m2=True, platforms=[]),
            "a200": GetStoragesModel(expect=["1", "2", "3"], is_a500=False, start_from_m2=False,
                                     platforms=["a6000000.sata", "a5180000.hiusbc3", "82000000.sdhci0"]),
        },
        "test_set_storage_info": {
            "invalid_storage_id": SetStorageInfoModel(expect=(None, const.STATE_GOOD), storage_id="4", storages={},
                                                      storage_info={}),
            "status_bad": SetStorageInfoModel(expect=("sata", const.STATE_BAD), storage_id="1",
                                              storages={"1": [Path(const.PLATFORM_ROOT, M2[1])]},
                                              storage_info={"Status": const.STATE_BAD})
        }
    }

    @staticmethod
    def test_get_storage_id(model: GetStorageIdModel):
        assert model.expect == Storage._get_storage_id(model.platform)

    @staticmethod
    def test_get_storage_id_failed():
        with pytest.raises(DevError):
            Storage._get_storage_id("")

    @staticmethod
    def test_sys_blocks(mocker: MockerFixture, model: SysBlocksModel):
        mocker.patch.object(Path, "glob", return_value=(Path(const.BLOCK_DIR, dev[0]) for dev in BLOCKS))
        mocker.patch("lib.Linux.systems.disk.storage.SystemUtils").return_value.is_a500 = model.is_a500
        mocker.patch("lib.Linux.systems.disk.storage.Env").return_value.start_from_m2 = model.start_from_m2
        blocks = Storage._sys_blocks()
        assert isinstance(blocks, Iterable)
        assert sorted(blocks) == sorted(Path(const.BLOCK_DIR, dev[0]) for dev in model.expect)

    @staticmethod
    def test_get_storages(mocker: MockerFixture, model: GetStoragesModel):
        mocker.patch.object(Path, "glob",
                            return_value=(Path(const.PLATFORM_ROOT, dev[1]) for dev in (EMMC, BOOT1, BOOT0, M2, USB)))
        sys_utils = mocker.patch("lib.Linux.systems.disk.storage.SystemUtils")
        sys_utils.return_value.is_a500 = model.is_a500
        mocker.patch("lib.Linux.systems.disk.storage.Env").return_value.start_from_m2 = model.start_from_m2
        mocker.patch("lib.Linux.systems.disk.storage.DeviceLoader").cfg.platforms = model.platforms
        assert sorted(set(Storage().get_storages())) == model.expect

    @staticmethod
    def test_set_storage_info(mocker: MockerFixture, model: SetStorageInfoModel):
        mocker.patch("lib.Linux.systems.disk.storage.DeviceLoader")
        mocker.patch.object(Storage, "_get_storage_info", return_value=model.storage_info)
        storage = Storage()
        storage._set_storage_info(model.storage_id, model.storages)
        assert (storage.Name, storage.Status) == model.expect
