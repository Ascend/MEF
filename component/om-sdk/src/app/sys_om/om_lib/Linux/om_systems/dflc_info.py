# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import datetime
import threading

from bin.lib_adapter import LibAdapter
from common.log.logger import run_log
from common.utils.app_common_method import AppCommonMethod
from devm.device_mgr import DEVM
from devm.exception import DeviceManagerError
from common.common_methods import CommonMethods


class DflcInfo:
    DFLC_LOCK = threading.Lock()

    def __init__(self):
        self.ProductName = ""
        self.SerialNumber = ""
        self.StartPoint = ""
        self.LifeSpan = 0
        self.ManufactureDate = ""
        self.Version = ""
        self.ItemId = ""

    @staticmethod
    def get_format_time(time):
        try:
            time = datetime.datetime.strptime(time.strip(), "%Y-%m-%d")
        except Exception as e:
            run_log.error("invalid time {} reason {}".format(time, e))
            return ""
        finally:
            pass

        return datetime.datetime.strftime(time, "%Y-%m-%d")

    @staticmethod
    def post_request(payload):
        if DflcInfo.DFLC_LOCK.locked():
            run_log.error("Config life_span is busy")
            return [CommonMethods.ERROR, "Config life_span is busy"]

        with DflcInfo.DFLC_LOCK:
            start_point = payload["StartPoint"]
            life_span = payload["LifeSpan"]

            if not isinstance(life_span, int) or life_span > 255 or life_span < 0:
                return [CommonMethods.ERROR, "invalid life_span"]
            if not isinstance(start_point, str) or len(start_point) > 32:
                return [CommonMethods.ERROR, "invalid start_point"]

            if start_point:
                start_point = DflcInfo.get_format_time(start_point)
                if not start_point:
                    return [CommonMethods.ERROR, "invalid start_point"]

            try:
                DEVM.get_device("system0").set_attribute("service_start_time", start_point)
            except DeviceManagerError as err:
                run_log.error("Set service_start_time of device system0 failed, caught exception: %s", err)
                return [CommonMethods.ERROR, "config start_point failed"]

            try:
                DEVM.get_device("system0").set_attribute("service_life", life_span)
            except DeviceManagerError as err:
                run_log.error("Set service_life of device system0 failed, caught exception: %s", err)
                return [CommonMethods.ERROR, "config life_span failed"]

            return [CommonMethods.OK, ""]

    def get_all_info(self):
        self.get_item_id()
        self.get_edge_system_info()
        self.get_manufacture_date()
        self.get_start_point()
        self.get_life_span()

    def get_start_point(self):
        try:
            time = DEVM.get_device("system0").get_attribute("service_start_time")
        except DeviceManagerError:
            return

        if time == "NA":
            try:
                time = DEVM.get_device("system0").get_attribute("first_power_on_time")
            except DeviceManagerError:
                return

            if time == "NA":
                time = ""

        time = time[2:-1] if time.startswith("-e") else time
        self.StartPoint = DflcInfo.get_format_time(time) if time else ""

    def get_life_span(self):
        try:
            self.LifeSpan = DEVM.get_device("system0").get_attribute("service_life")
        except DeviceManagerError:
            return

    def get_manufacture_date(self):
        try:
            manufacture_date = DEVM.get_device("elabel0").get_attribute("mfg_date")
        except DeviceManagerError:
            return

        split_list = manufacture_date.split()
        if not split_list:
            return

        self.ManufactureDate = DflcInfo.get_format_time(split_list[0])

    # 获取item id
    def get_item_id(self):
        try:
            self.ItemId = DEVM.get_device("elabel0").get_attribute("product_sn")
        except DeviceManagerError:
            return

    def get_edge_system_info(self):
        ret_dict = LibAdapter.lib_restful_interface("System", "GET", None, False)
        ret = AppCommonMethod.check_status_is_ok(ret_dict)
        if not ret:
            run_log.error("Get system info failed.")
            return

        ret = ret_dict.get("message")
        self.ProductName = ret.get("Model")
        self.SerialNumber = ret.get("SerialNumber")
