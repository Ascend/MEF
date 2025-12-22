# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import subprocess
from contextlib import contextmanager
from functools import partial
from pathlib import Path
from typing import Optional, Iterable, Tuple, Dict, Callable

from lib.Linux.systems.disk.errors import DevError
from lib.Linux.systems.disk.schema import DevAction, SubSystem, UdevMonitorInfo, Disk


def property_iter(dev_reader: Iterable[bytes]) -> Iterable[Tuple[str, str]]:
    """
    生成单个设备对应的udev的属性
    :param dev_reader: 迭代器，一个一个地产出udevadm monitor输出行
    :return: 属性（property, value）迭代器
    """
    for index, item in enumerate(dev_reader):
        # 如果读到b""， 表示udevadm monitor已time out,但proc进程还在运行会不断将b""写入到proc.stdout
        if item == b"":
            raise DevError("monitor timeout")
        # 非UDEV开头表示并非一个device
        if index == 0 and not item.startswith(b"UDEV"):
            break
        # 第一行类似UDEV [时间戳] action DEVPATH，非属性
        if index == 0:
            continue
        yield item.decode().strip("\n").split("=")


def dev_property_dict(proc: subprocess.Popen) -> Dict[str, str]:
    """
    生成单个设备的udev属性字典
    :param proc: timeout 超时时间 udevadm monitor -up命令对应的子进程
    :return: 单个设备的属性字典
    """
    # monitor 输出以b"\n"行分割出一个个的device信息
    one_dev_reader = iter(proc.stdout.readline, b"\n")
    return dict(property_iter(one_dev_reader))


def terminate_monitor(proc: subprocess.Popen, terminate: Callable[[UdevMonitorInfo], bool]):
    """
    monitor终止函数
    :param proc: timeout 超时时间 udevadm monitor -up命令对应的子进程
    :param terminate: 终止判断函数
    :return: 一个device的属性字典
    """
    for dev in iter(partial(dev_property_dict, proc), None):
        if not dev:
            continue
        if terminate(UdevMonitorInfo.from_dict(dev)):
            break


@contextmanager
def udev_monitor(
        terminate: Callable[[UdevMonitorInfo], bool],
        timeout: int = 3,
        subsystem: Optional[str] = SubSystem.DISK
):
    """
    udevadm monitor上下文管理器，用于监测设备变化
    :param terminate: 终止监测的判断函数，当该函数返回True时，即表示监测到了对应的变化
    :param timeout: 监测进程的运行时间，如果在该时间内没有terminate，则会抛出DevError
    :param subsystem: 用于过滤udev事件，block/disk表示只监控磁盘，block/partition表示只监控分区
    :return: NA
    """
    # -u: 只监测udev事件；-p：将udev事件属性输出到stdout
    udev_cmd = "timeout", str(timeout), "udevadm", "monitor", "-up"
    if subsystem:
        udev_cmd += ("-s", subsystem)
    with subprocess.Popen(udev_cmd, stdout=subprocess.PIPE, stdin=subprocess.PIPE, shell=False) as proc:
        try:
            yield
            # 此处try多行为contextmanager语法，当被监控的with语块执行完成后才会执行yield后面的逻辑
            terminate_monitor(proc, terminate)
        finally:
            proc.kill()


def terminate_by_gpt(udev: UdevMonitorInfo, disk: Disk) -> bool:
    """转换gpt成功时的终止监控函数"""
    return all((udev.path == disk.path, udev.action == DevAction.CHANGE, udev.pt_type == "gpt"))


def terminate_while_add_part(udev: UdevMonitorInfo, disk: Disk) -> bool:
    """创建分区成功时的终止监控函数"""
    return all((udev.action == DevAction.ADD, Path(udev.dev_path).parent.name == disk.name))


def terminate_by_path(udev: UdevMonitorInfo, path: str, action=DevAction.CHANGE) -> bool:
    return all((udev.path == path, udev.action == action))
