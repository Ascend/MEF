# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import pytest

from common.yaml.yaml_exception import YamlException
from common.yaml.yaml_node import YamlNode


class TestYamlNode:

    @staticmethod
    def test_is_root_node():
        node = YamlNode("test", None, 0)
        assert node.is_root_node()

    @staticmethod
    def test_is_child_of():
        node = YamlNode("test", None, 0)
        node2 = YamlNode("test2", None, 2)
        assert node2.is_child_of(node)

    @staticmethod
    def test_get_father_node():
        node = YamlNode("test", None, 0)
        node2 = YamlNode("test2", None, 2)
        node3 = YamlNode("test3", None, 2)
        node2.pre_node = node
        node3.get_father_node(node2)

    @staticmethod
    def test_get_father_node_null():
        node = YamlNode("test", None, 0)
        with pytest.raises(YamlException):
            node.get_father_node(None)

    @staticmethod
    def test_get_father_node_not_found():
        node = YamlNode("test", None, 0)
        node2 = YamlNode("test2", None, 2)
        with pytest.raises(YamlException):
            node.get_father_node(node2)
