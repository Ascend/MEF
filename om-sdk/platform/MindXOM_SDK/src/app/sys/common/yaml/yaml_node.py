# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from enum import Enum
from typing import Optional, Any, List

from common.log.logger import run_log
from common.yaml.yaml_exception import YamlException


class YamlNodeType(Enum):
    NORMAL = "normal"
    LIST = "list"
    DICT = "dict"


class YamlNode:

    def __init__(self, name: str, value: Any, space_len: int, list_flag: bool = False, node_type: YamlNodeType = None):
        self.name = name
        self.value = value
        self.list_flag = list_flag
        self.space_len = space_len
        # 类型有list normal dict
        self.node_type = node_type
        self.pre_node: Optional[YamlNode] = None
        self.next_nodes: List[YamlNode] = []

    def is_root_node(self) -> bool:
        """
        功能： 判断是否为根节点
        参数： self
        返回值： bool 是否为根节点
        """
        return self.space_len == 0

    def is_child_of(self, standard_node: "YamlNode"):
        """
        功能： 通过每行前面的空格数量判断当前节点是否为前一个节点的子节点
        参数： self
        参数： standard_node 比较节点
        返回值： bool 是否为比较节点的子节点
        """
        return self.space_len > standard_node.space_len

    def get_father_node(self, standard_node: "YamlNode"):
        """
        功能： 获取父节点
        参数： self
        参数： standard_node 比较节点
        返回值：父节点
        """
        if not standard_node:
            raise YamlException("standard_node is null!")

        temp_node = standard_node
        while True:
            if temp_node.space_len < self.space_len:
                return temp_node
            if not temp_node.pre_node:
                run_log.error("Every node must has father node except it's root node!")
                raise YamlException("Every node must has father node except it's root node!")
            temp_node = temp_node.pre_node
