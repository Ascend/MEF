#!/usr/bin/python
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.

import json
import os

from common.constants.config_constants import ConfigPathConstants
from common.file_utils import FileCheck
from common.log.logger import run_log


class Capability:
    def __init__(self):
        self.product_capability = Capability.get_product_capability()

    @staticmethod
    def get_product_capability():
        default_capability = Capability.get_default_capability()
        return Capability.combine_product_capability(default_capability)

    @staticmethod
    def combine_product_capability(default_ability):
        if not default_ability or not isinstance(default_ability, dict):
            return {}
        if not default_ability.get("product_capability", []):
            return default_ability

        default_ability["product_capability"] = list(set(default_ability["product_capability"]))
        return default_ability

    @staticmethod
    def get_default_capability():
        return Capability.get_capability_from_file(ConfigPathConstants.DEFAULT_CAPABILITY_FILE)

    @staticmethod
    def get_capability_from_file(file_path):
        capability_data = {}
        if not os.path.exists(file_path):
            return capability_data

        res = FileCheck.check_path_is_exist_and_valid(file_path)
        if not res:
            run_log.error(f"get capability from file failed, {file_path} invalid : {res.error}")
            return capability_data

        try:
            with open(file_path, "r") as content:
                capability_data = json.load(content)
                run_log.info("Get capability from {} sucess".format(os.path.basename(file_path)))
        except (IOError, OSError) as err:
            run_log.error("open capability file:{} exception:{}".format(os.path.basename(file_path), err))
        except Exception:
            run_log.error(f"json load capability:{file_path} exception")

        return capability_data


CAPABILITY = Capability()
