# coding: utf-8
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
import os.path

from pytest_mock import MockerFixture

from devm.devm_configs import DevmConfigMgr
from ut_utils.mock_utils import mock_cdll

with mock_cdll():
    from devm.device_mgr import Devm


class TestDevm:
    cur_path = os.path.dirname(os.path.realpath(__file__))
    product_file = os.path.join(cur_path, "product_specification.json")
    module_files = [os.path.join(cur_path, "module_abc.json")]
    product_data = {
        "name": "Atlas 500",
        "modules": {
            "module_abc": {
                "devices": ["device_abc_0"]
            },
        }
    }
    modules_data = [
        {
            "id": 1,
            "name": "tlv_demo_module",
            "category": "internal",
            "driver": "/path/to/libmodule_abc.so",
            "dynamic": False,
            "attributes": {
                "INT ATTR NAME": {"id": 2, "type": "int", "accessMode": "ReadWrite"},
                "FLOAT ATTR NAME": {"id": 3, "type": "float", "accessMode": "ReadWrite"},
                "BOOL ATTR NAME": {"id": 1, "type": "bool", "accessMode": "ReadWrite"},
                "LONG LONG ATTR NAME": {"id": 4, "type": "long long", "accessMode": "ReadWrite"},
                "STRING ATTR NAME": {"id": 5, "type": "string", "accessMode": "ReadWrite"},
                "JSON ATTR NAME": {
                    "id": 6,
                    "type": "json",
                    "accessMode": "ReadWrite",
                    "subAttributes": {
                        "present": {
                            "id": 1,
                            "accessMode": "ReadWrite",
                            "description": "present desc",
                            "type": "bool"
                        },
                        "id": {
                            "id": 2,
                            "accessMode": "ReadWrite",
                            "description": "id desc",
                            "type": "int"
                        },
                        "rate": {
                            "id": 3,
                            "accessMode": "ReadWrite",
                            "description": "rate desc",
                            "type": "float"
                        },
                        "memory": {
                            "id": 4,
                            "accessMode": "ReadWrite",
                            "description": "memory desc",
                            "type": "long long"
                        },
                        "name": {
                            "id": 5,
                            "accessMode": "ReadWrite",
                            "description": "name desc",
                            "type": "string"
                        }
                    }
                }
            }
        },
    ]

    use_cases = {

    }

    def setup_class(self):
        with os.fdopen(os.open(self.product_file, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600), "w") as tmp_file:
            tmp_file.write(json.dumps(self.product_data, indent=4))

        for i, module_file in enumerate(self.module_files):
            with os.fdopen(os.open(module_file, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600), "w") as tmp_file:
                tmp_file.write(json.dumps(self.modules_data[i], indent=4))

    def teardown_class(self):
        if os.path.exists(self.product_file):
            os.remove(self.product_file)

        for module_file in self.module_files:
            if os.path.exists(module_file):
                os.remove(module_file)

    def test_get_product_config(self, mocker: MockerFixture):
        mocker.patch.object(DevmConfigMgr, "check_config", return_value=True)
        assert Devm().get_product_config(self.cur_path) == self.product_data

    def test_get_module_config(self, mocker: MockerFixture):
        mocker.patch.object(DevmConfigMgr, "check_config", return_value=True)
        res = list(Devm().get_module_configs(self.cur_path))
        assert res == self.modules_data
