# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
from collections import namedtuple
from unittest.mock import mock_open, patch

import pytest
from pytest_mock import MockerFixture

from common.file_utils import FileCheck
from common.yaml.yaml_exception import YamlException
from common.yaml.yaml_node import YamlNodeType, YamlNode
from common.utils.result_base import Result
from common.yaml.yaml_methods import YamlMethod

ReadYamlExceptionCase = namedtuple("ReadYamlExceptionCase", "check_file, size, read_lines")
GenerateYamlNodeCase = namedtuple("GenerateYamlNodeCase", "value, type, line")
DealNotConvert2NodeCase = namedtuple("DealNotConvert2NodeCase", "value, node, line")
DealNotConvert2NodeExceptionCase = namedtuple("DealNotConvert2NodeCase", "node, line")
ConvertYamlToTreeCase = namedtuple("ConvertYamlToTreeCase", "root_name, child1_name, child2_name, read_yaml")


class TestYamlMethod:
    use_cases = {
        "test_read_yaml_exception": {
            "check file failed": (Result(False, "err"), None, None),
            "check size failed": (True, 2 * 1024 * 1024, None),
            "check lines failed": (True, 2 * 1024, ["a"] * 400),
        },
        "test_generate_yaml_node": {
            "value null": (None, YamlNodeType.LIST, "  - a:"),
            "start with [": (["1"], YamlNodeType.LIST, "  a: [1]"),
            "start with \"": ("1", YamlNodeType.NORMAL, "a: \"1\""),
            "value is false": (False, YamlNodeType.NORMAL, "a: false"),
            "value is true": (True, YamlNodeType.NORMAL, "a: true"),
        },
        "test_deal_not_convert2_node": {
            "type null and start with -": (["a"], YamlNode("test", None, 0), "  - a"),
            "type null and start with [": (["1"], YamlNode("test", None, 0), "  [1]"),
            "type null": ("1", YamlNode("test", None, 0), "  1"),
            "type normal and value exist": ("1 2", YamlNode("test", "1", 0, node_type=YamlNodeType.NORMAL), "  2"),
            "type normal": ("1", YamlNode("test", None, 0, node_type=YamlNodeType.NORMAL), "1"),
            "type list": (["1", "2"], YamlNode("test", ["1"], 0, node_type=YamlNodeType.LIST), "  - 2"),
        },
        "test_deal_not_convert2_node_exception": {
            "node null": (None, "  - a"),
            "type list and line error": (YamlNode("test", ["1"], 0, node_type=YamlNodeType.LIST), "  1"),
            "type dict and line error": (YamlNode("test", ["1"], 0, node_type=YamlNodeType.DICT), "  1"),
        },
        "test_convert_yaml_to_tree": {
            "normal": ("network", "ethernets", "renderer",
                       [
                           "network:", "  ethernets:", "    eth0:",
                           "      addresses:", "      - 51.38.68.152/19",
                           "      dhcp4: false", "      gateway4: 51.38.64.1",
                           "      nameservers:", "        addresses:",
                           "        - 255.255.224.0", "      routes:",
                           "      - to: default", "        via: 51.38.64.1",
                           "    usb0:", "      addresses:",
                           "      - 192.168.1.2/24", "      dhcp4: false",
                           "  renderer: networkd", "  version: 2"
                       ],),
        },
    }

    @staticmethod
    def test_read_yaml(mocker: MockerFixture):
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid", return_value=True)
        mocker.patch("os.path.getsize", return_value=1)
        with patch("builtins.open", new_callable=mock_open, read_data="test") as mock_file:
            mock_file.return_value.__enter__.return_value.readlines.return_value = ["test"]
            ret = YamlMethod._read_yaml("/test")
            assert ret == ["test"]

    @staticmethod
    def test_read_yaml_exception(mocker: MockerFixture, model: ReadYamlExceptionCase):
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid", return_value=model.check_file)
        mocker.patch("os.path.getsize", return_value=model.size)
        with patch("builtins.open", new_callable=mock_open, read_data="test") as mock_file:
            mock_file.return_value.__enter__.return_value.readlines.return_value = model.read_lines
            with pytest.raises(YamlException):
                YamlMethod._read_yaml("/test")

    @staticmethod
    def test_generate_yaml_node(model: GenerateYamlNodeCase):
        node = YamlMethod._generate_yaml_node(model.line)
        assert node.node_type == model.type
        assert node.value == model.value

    @staticmethod
    def test_can_convert2_node_or_not():
        assert YamlMethod._can_convert2_node_or_not("a: b")

    @staticmethod
    def test_deal_not_convert2_node(model: DealNotConvert2NodeCase):
        YamlMethod._deal_not_convert2_node(model.node, model.line)
        assert model.node.value == model.value

    @staticmethod
    def test_deal_not_convert2_node_exception(model: DealNotConvert2NodeExceptionCase):
        with pytest.raises(YamlException):
            YamlMethod._deal_not_convert2_node(model.node, "test")

    @staticmethod
    def test_convert_yaml_to_tree(mocker: MockerFixture, model: ConvertYamlToTreeCase):
        mocker.patch.object(YamlMethod, "_read_yaml", return_value=model.read_yaml)
        ret = YamlMethod._convert_yaml_to_tree("/test.yaml")
        assert ret[0].name == model.root_name
        assert ret[0].next_nodes[0].name == model.child1_name
        assert ret[0].next_nodes[1].name == model.child2_name

    @staticmethod
    def test_convert2_dict_by_tree():
        node_network = YamlNode("network", None, 0)
        node_ethernets = YamlNode("ethernets", None, 2)
        node_eth0 = YamlNode("eth0", None, 4)
        node_address = YamlNode("addresses", ["51.38.68.152"], 6, node_type=YamlNodeType.LIST)
        node_dhcp4 = YamlNode("dhcp4", False, 6, node_type=YamlNodeType.NORMAL)
        node_route = YamlNode("routes", None, 6, node_type=YamlNodeType.LIST)
        node_to = YamlNode("to", "default", 8, list_flag=True, node_type=YamlNodeType.NORMAL)
        node_via = YamlNode("via", "51.38.64.1", 8, node_type=YamlNodeType.NORMAL)
        node_network.next_nodes.append(node_ethernets)
        node_ethernets.next_nodes.append(node_eth0)
        node_eth0.next_nodes.extend([node_address, node_dhcp4, node_route])
        node_route.next_nodes.extend([node_to, node_via])
        res = {}

        YamlMethod._convert2_dict_by_tree([node_network], res)
        eth0 = res.get("network").get("ethernets").get("eth0")
        assert eth0.get("addresses")[0] == "51.38.68.152"
        assert not eth0.get("dhcp4")
        assert eth0.get("routes")[0].get("to") == "default"

    @staticmethod
    def test_load_yaml_info(mocker: MockerFixture):
        node_network = YamlNode("network", "123", 0, node_type=YamlNodeType.NORMAL)
        mocker.patch.object(YamlMethod, "_convert_yaml_to_tree", return_value=[node_network])
        res = YamlMethod.load_yaml_info("/test.yaml")
        assert res.get("network") == "123"

    @staticmethod
    def test_convert2_yaml_info():
        source = {
            "network": {
                "ethernets": {
                    "eth0": {
                        "addresses": ["51.38.68.152"],
                        "dhcp4": False,
                        "routes": [{
                            "to": "default",
                            "via": "51.38.64.1"
                        }]
                    }
                }
            }
        }
        res = []
        YamlMethod._convert2_yaml_info(source, res)
        assert res[0] == "network:\n"
        assert res[1] == "  ethernets:\n"
        assert res[2] == "    eth0:\n"
