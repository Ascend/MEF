# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
from pathlib import Path
from typing import NamedTuple, Tuple

from pytest_mock import MockerFixture

from lib.Linux.systems.disk.device_loader import DeviceLoader, CommonConstants, const
from lib.Linux.systems.disk.schema import DevConfigs

ETH_INFO = f"""any: any
version: 5.10.0+
any: any
"""


class A500GetFmtInfoModel(NamedTuple):
    expect: Tuple
    value: str
    field: str


class GetEthVersionModel(NamedTuple):
    expect: str
    info: str


class TestDeviceLoader:
    use_cases = {
        "test_a500_get_fmt_info": {
            "get_fmt_info_by_path": A500GetFmtInfoModel(expect=("disk0", "/dev/sda"), value="/dev/sda", field="path"),
            "get_fmt_info_by_loc": A500GetFmtInfoModel(expect=("disk0", "/dev/sda"), value="PCIE3-1", field="location"),
        },
        "test_get_eth_version": {
            "normal": GetEthVersionModel(expect=" 5.10.0+", info=ETH_INFO),
            "null": GetEthVersionModel(expect="", info="a:b"),
        }
    }

    @staticmethod
    def test_load_200_dev_config(mocker: MockerFixture):
        """测试包中的200配置文件能否正常加载，若不能，则修改的配置或者处理代码可能有问题"""
        mocker.patch("lib.Linux.systems.disk.device_loader.SystemUtils").return_value.is_a500 = False
        from lib import Linux
        cfg_path = Path(Linux.__file__).parent.joinpath("config", "device_config.json")
        CommonConstants.DEVICE_CFG_PATH = cfg_path.as_posix()
        DeviceLoader.load_dev_configs()
        assert isinstance(DeviceLoader.cfg, DevConfigs)
        DeviceLoader.cfg = None

    @staticmethod
    def test_a500_fmt_info_generator():
        const.FORMAT_INFO = Path(__file__).parent.joinpath("format_hw.info").as_posix()
        # 执行加载配置，能测验FormatInfo数据结构及其初始化逻辑
        assert len(list(DeviceLoader()._a500_fmt_info_generator())) == 5

    @staticmethod
    def test_a500_get_fmt_info(mocker: MockerFixture, model: A500GetFmtInfoModel):
        mocker.patch("lib.Linux.systems.disk.device_loader.SystemUtils").return_value.is_a500 = True
        fmt = DeviceLoader().get_fmt_info(model.value, model.field)
        assert (fmt.name, fmt.path) == model.expect

    @staticmethod
    def test_get_eth_version(mocker: MockerFixture, model: GetEthVersionModel):
        """测试解析逻辑是否正常"""
        mocker.patch.object(DeviceLoader, "exec_cmd", return_value=(0, model.info))
        assert DeviceLoader()._get_eth_version("eth0") == model.expect

    @staticmethod
    def test_load_disk(mocker: MockerFixture):
        """根据用例数据测验加载逻辑是否异常"""
        cur_dir = Path(__file__).parent
        lsblk_info = cur_dir.joinpath("lsblk.json").read_text()
        udev_sda = cur_dir.joinpath("sda_udev_info.txt").read_text()
        udev_sda1 = cur_dir.joinpath("sda1_udev_info.txt").read_text()
        udev_sda2 = cur_dir.joinpath("sda2_udev_info.txt").read_text()
        effect = ((0, lsblk_info), (0, udev_sda), (0, udev_sda1), (0, udev_sda2))
        mocker.patch.object(DeviceLoader, "exec_cmd", side_effect=effect)
        disk = DeviceLoader().load_disk("sda")
        assert disk.offset_for_new_part(1 * const.KILO_BINARY_BYTE_RATE ** 2, disk.children[0]) == (50470912, 51519488)
