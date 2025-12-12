# coding: utf-8
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import glob
import json
import os
import threading
import time
from typing import Dict
from typing import Iterable
from typing import List
from typing import Optional

from common.constants.base_constants import CommonConstants
from common.log.logger import run_log
from common.utils.exec_cmd import ExecCmd
from common.utils.singleton import Singleton
from common.utils.timer import RepeatingTimer
from devm import config
from devm.constants import DeviceAttrTypeEnum
from devm.constants import ModuleCategoryEnum
from devm.device_attrs import DeviceAttr
from devm.device_attrs import attr_factory
from devm.device_driver import DeviceDriver
from devm.devm_configs import DevmConfigMgr
from devm.exception import AttributeNotExistError, DriverError
from devm.exception import DeviceManagerError
from devm.exception import DeviceNotExistError
from devm.exception import MethodNotAllowedError
from devm.exception import TooManyModuleFilesError


class Devm(Singleton):

    UPDATE_MODULES_LOCK = threading.Lock()

    def __init__(self):
        self.product_name = None
        self.modules = {}
        self.refresh_task: Optional[threading.Thread] = None
        self.monitor_5g_task: Optional[threading.Thread] = None
        self.device_5g = ""
        self.last_module_load_state = {}

    @staticmethod
    def check_cfg_file_valid(cfg_file):
        mgr = DevmConfigMgr(CommonConstants.OM_WORK_DIR_PATH)

        try:
            ret = mgr.check_config(cfg_file)
        except Exception as err:
            run_log.error("check config file failed, caught exception: %s", err)
            return False

        if not ret:
            run_log.error("%s file info not same with db", cfg_file)

        return ret

    @staticmethod
    def get_5g_device():
        cmd = 'lsusb | grep "2c7c:0900"'
        ret = ExecCmd.exec_cmd_use_pipe_symbol(cmd, wait=10)
        if ret[0] != 0:
            return ""
        return ret[1]

    def init(self, config_dir):
        """
        从配置文件目录中加载产品和模组配置

        :param config_dir:
        :return:
        """
        product_config = self.get_product_config(config_dir)
        self.product_name = product_config["name"]
        self.modules = {}

        product_modules = product_config["modules"]
        for _, module_config in enumerate(self.get_module_configs(config_dir)):
            module_name = module_config["name"]
            if module_name not in product_modules:
                run_log.error("%s not in product_modules, ignored", module_name)
                continue

            device_names = product_modules[module_name]["devices"]
            try:
                module = Module.load_from_json(module_config, device_names)
            except Exception as err:
                run_log.error("init module %s failed: %s", module_name, err)
                continue

            self.add_module(module_name, module)

        # 初始化各个动态模组加载状态，用于控制同一模组加载失败日志打印，防止失败之后一直刷屏
        self.last_module_load_state = {module.name: True for module in self.modules.values() if module.is_dynamic}
        self.device_5g = self.get_5g_device()
        self.start_refresh_task()
        self.start_monitor_lte_task()

    def get_product_config(self, config_dir) -> Dict:
        """
        获取product.json并解析

        :param config_dir: 配置文件目录
        :return:
        """
        product_file = os.path.join(config_dir, "product_specification.json")
        ret = self.check_cfg_file_valid(product_file)
        if not ret:
            run_log.error("%s product file invalid", product_file)
            raise DeviceManagerError("product file invalid")

        with open(product_file, "r") as file_read:
            return json.loads(file_read.read())

    def get_module_configs(self, config_dir) -> Iterable[Dict]:
        """
        获取所有模组的配置并解析

        :param config_dir: 配置文件目录
        :return: 迭代器
        """
        module_files = glob.iglob(os.path.join(config_dir, "module_*.json"))
        for index, module_file in enumerate(module_files):
            if index > config.MAX_MODULE_NUM:
                run_log.error("too many module files, count must be less than or equal to %s", config.MAX_MODULE_NUM)
                raise TooManyModuleFilesError("too many module files")

            ret = self.check_cfg_file_valid(module_file)
            if not ret:
                run_log.error("%s module file invalid", module_file)
                raise DeviceManagerError("module file invalid")

            with open(module_file, "r") as file_read:
                try:
                    yield json.load(file_read)
                except Exception as err:
                    run_log.warning("%s file content invalid json format: %s", module_file, err)
                    continue

    def get_module_count(self):
        return len(self.modules)

    def add_module(self, name, module: "Module"):
        with Devm.UPDATE_MODULES_LOCK:
            self.modules[name] = module
            run_log.info("added module %s successfully", name)

    def del_module(self, module_name):
        with Devm.UPDATE_MODULES_LOCK:
            module = self.get_module(module_name)
            if not module:
                run_log.info("module %s not exist", module_name)
                return
            module.driver.unload()
            del self.modules[module_name]
            run_log.info("delete module %s successfully", module_name)

    def get_module(self, module_name) -> "Module":
        return self.modules.get(module_name)

    def get_device(self, device_name) -> "Device":
        for module in self.modules.values():
            device = module.get_device(device_name)
            if device:
                return device

        raise DeviceNotExistError(f"device {device_name} not exists")

    def count_devices_by_module(self, module_name) -> int:
        if module_name not in self.modules:
            return 0

        return self.modules[module_name].get_device_count()

    def get_device_list_by_category(self, category_name):
        """
        获取指定模组类别的所有设备名称列表
        :param category_name: 类别
        :return: 设备名称字符串列表
        """
        try:
            category = ModuleCategoryEnum(category_name)
        except ValueError as err:
            run_log.error("category name invalid")
            raise DeviceManagerError("category name invalid") from err

        device_names = []
        for module in self.modules.values():
            if module.category != category:
                continue

            device_names.extend(module.get_all_devices_names())

        return device_names

    def get_module_list_by_category(self, category_name):
        """
        获取指定模组类别的所有模组列表
        :param category_name: 类别
        :return: 模组名称字符串列表
        """
        try:
            category = ModuleCategoryEnum(category_name)
        except ValueError as err:
            run_log.error("category name invalid")
            raise DeviceManagerError("category name invalid") from err

        return [name for name, module in self.modules.items() if module.category == category]

    def get_device_list_by_module(self, module_name):
        """
        获取指定模组名的所有设备名称列表
        :param module_name: 模组名
        :return:
        """
        if module_name not in self.modules:
            return []

        return self.modules[module_name].get_all_devices_names()

    def get_module_info(self, module_name: str) -> Dict:
        if module_name not in self.modules:
            return {}

        module = self.modules[module_name]
        module_info = {
            "name": module_name,
            "attributes": module.get_all_attr_info(),
            "category": module.category.value,
            "devices": module.get_all_devices_names(),
        }
        web_location = module.get_module_web_location()
        if web_location:
            module_info["UILocation"] = web_location
        return module_info

    def has_device(self, dev_name):
        for module in self.modules.values():
            if dev_name in module.get_all_devices_names():
                return True

        return False

    def refresh_module(self):
        """动态模组设备检测任务"""
        for module_name, module in self.modules.items():
            if not module.is_dynamic:
                continue

            try:
                module.refresh_dynamic_devices()
            except Exception as err:
                if self.last_module_load_state.get(module.name):
                    run_log.error("dynamic module [%s] refresh failed, caught exception: %s", module_name, err)
                    self.last_module_load_state[module.name] = False
            else:
                self.last_module_load_state[module.name] = True

    def get_module_config_by_module_name(self, config_dir: str, module_name: str):
        module_file = os.path.join(config_dir, f"module_{module_name}.json")
        ret = self.check_cfg_file_valid(module_file)
        if not ret:
            run_log.error("%s module file invalid", module_file)
            return {}

        with open(module_file, "r") as file_read:
            try:
                return json.load(file_read)
            except Exception as err:
                run_log.warning("%s file content invalid json format: %s", module_file, err)
                return {}

    def monitor_5g_whether_restart(self):
        while True:
            try:
                time.sleep(config.FIVE_G_MODULES_POLLING_INTERVAL)
                # 检测5g模组设备号是否发生变化
                cur_device_5g = self.get_5g_device()
                if not cur_device_5g or cur_device_5g == self.device_5g:
                    self.device_5g = cur_device_5g
                    continue

                # 发现device值变化后，等驱动初始化完成后，再重新初始化lte模块
                usb_real_path = os.path.realpath("/dev/LTE_TTY")
                if not os.path.exists(usb_real_path):
                    run_log.warning("link dev/LTE_TTY target file not exist")
                    continue

                self.del_module("lte")
                product_config = self.get_product_config(DEVM_CONFIG_DIR)
                product_modules = product_config["modules"]
                if "lte" not in product_modules:
                    continue

                module_config = self.get_module_config_by_module_name(DEVM_CONFIG_DIR, "lte")
                if not module_config:
                    continue

                device_names = product_modules["lte"]["devices"]
                try:
                    module = Module.load_from_json(module_config, device_names)
                except Exception as err:
                    run_log.error("init module %s failed: %s", "lte", err)
                    continue

                self.add_module("lte", module)
                self.device_5g = cur_device_5g
            except Exception as err:
                run_log.error("init lte module failed, %s", err)

    def start_refresh_task(self):
        """启动动态模组设备检测任务"""
        if self.refresh_task:
            run_log.warning("refresh task already exists")
            return

        self.refresh_task = RepeatingTimer(config.DYNAMIC_MODULES_POLLING_INTERVAL, self.refresh_module)
        self.refresh_task.setDaemon(True)
        self.refresh_task.start()
        run_log.info("refresh task created")

    def start_monitor_lte_task(self):
        """启动检测5g模块是否重启过任务"""
        if self.monitor_5g_task:
            run_log.warning("monitor lte task already exists")
            return

        self.monitor_5g_task = threading.Thread(target=self.monitor_5g_whether_restart, daemon=True)
        self.monitor_5g_task.start()
        run_log.info("monitor lte task created")


class Module:
    """模组类"""

    def __init__(self, id_, name, category, driver, dynamic, attrs: Dict, web_location, device_names=None):
        self.id: int = id_
        self.name: str = name
        self.category: ModuleCategoryEnum = ModuleCategoryEnum(category)
        self.driver: DeviceDriver = DeviceDriver(driver)
        self.is_dynamic: bool = dynamic
        self.web_location: str = web_location if self.category == ModuleCategoryEnum.ADDITION else "Manage"

        self.attrs = {}
        for attr_name, attr in attrs.items():
            attr_type = DeviceAttrTypeEnum(attr["type"])
            self.attrs[attr_name] = attr_factory(attr_type, attr_name, attr)

        self._devices: Dict[str, Device] = {}
        device_names = device_names or []
        for device_name in device_names:
            try:
                self.add_device(device_name)
            except Exception as err:
                run_log.error(f"add device {device_name} failed: {err}")

    @classmethod
    def load_from_json(cls, module_json, device_names=None) -> "Module":
        id_ = module_json["id"]
        name = module_json["name"]
        category = module_json["category"]
        driver = module_json["driver"]
        dynamic = module_json["dynamic"]
        attrs = module_json["attributes"]
        web_location = module_json.get("webLocation")
        return cls(id_, name, category, driver, dynamic, attrs, web_location, device_names)

    def open_device(self, device_name) -> int:
        return self.driver.open(device_name)

    def close_device(self, fd):
        return self.driver.close(fd)

    def add_device(self, name):
        self._devices[name] = Device(name, self)
        run_log.info("added device %s successfully", name)

    def get_device(self, name) -> "Device":
        return self._devices.get(name)

    def del_device(self, name):
        dev = self._devices[name]
        dev.close()
        self._devices.pop(name)
        run_log.info("deleted device %s successfully", name)

    def get_attr(self, name) -> DeviceAttr:
        attr = self.attrs.get(name)
        if not attr:
            run_log.error("attr %s not exist", name)
            raise AttributeNotExistError(f"attr {name} not exist")

        return attr

    def get_device_count(self) -> int:
        return len(self._devices)

    def get_all_devices_names(self) -> List[str]:
        return list(self._devices.keys())

    def get_all_attr_info(self) -> Dict:
        attr_info = {}
        for attr_name, attr in self.attrs.items():
            attr_info[attr_name] = attr.to_json()
        return attr_info

    def get_all_readable_attrs(self) -> Iterable[str]:
        for attr_name, attr in self.attrs.items():
            if attr.access_mode.is_readable:
                yield attr_name

    def get_module_web_location(self):
        return self.web_location

    def refresh_dynamic_devices(self):
        if not self.is_dynamic:
            raise DriverError(f"module [{self.name}] is not dynamic, can not be refreshed")

        try:
            real_devices = self.driver.get_dynamic_devices(self.id)
        except Exception as err:
            raise DriverError(f"refresh dynamic devices failed: {err}") from err

        add_devs = real_devices - self._devices.keys()
        del_devs = self._devices.keys() - real_devices
        for name in del_devs:
            try:
                self.del_device(name)
            except DeviceManagerError as err:
                run_log.error("delete device [%s] failed, caught exception: %s", name, err)
                continue

        for name in add_devs:
            try:
                self.add_device(name)
            except DeviceManagerError as err:
                run_log.error("add device [%s] failed, caught exception: %s", name, err)
                continue


class Device:
    """设备类"""

    def __init__(self, name, module):
        self.name: str = name
        self.module: Module = module
        self.fd: int = self.open()

    def open(self):
        return self.module.open_device(self.name)

    def close(self):
        self.module.close_device(self.fd)

    def get_attr_by_module(self, name: str) -> Optional[DeviceAttr]:
        return self.module.get_attr(name)

    def get_attribute(self, name):
        attr = self.module.get_attr(name)
        if not attr.access_mode.is_readable:
            raise MethodNotAllowedError(f"attr name: {name} is not readable")

        tlv_req = attr.dump_tlv()
        tlv_resp = self.module.driver.get_attr(self.fd, tlv_req)
        return attr.load_tlv(tlv_resp)[2]

    def get_all_attributes(self):
        attrs_info = {}
        for attr_name in self.module.get_all_readable_attrs():
            try:
                attrs_info[attr_name] = self.get_attribute(attr_name)
            except DeviceManagerError as err:
                run_log.error("get attribute %s failed, caught exception: %s", attr_name, err)
        return attrs_info

    def set_attribute(self, name, val):
        attr = self.module.get_attr(name)
        if not attr.access_mode.is_writeable:
            raise MethodNotAllowedError(f"attr name: {name} is not writeable")

        self.set_attr_by_module(attr, val)

    def set_attr_by_module(self, attr_instance: DeviceAttr, val: int):
        self.module.driver.set_attr(self.fd, attr_instance.dump_tlv(val))


DEVM: Devm = Devm()
DEVM_CONFIG_DIR = "/usr/local/mindx/MindXOM/software/ibma/config/devm_configs/"
try:
    DEVM.init(DEVM_CONFIG_DIR)
except Exception as error:
    run_log.error("devm init failed, caught exception: %s", error)
