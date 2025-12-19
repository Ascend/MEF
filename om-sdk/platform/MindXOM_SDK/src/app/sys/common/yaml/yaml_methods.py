# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import os
from typing import List, Optional

from common.constants.base_constants import CommonConstants
from common.file_utils import FileCheck
from common.log.logger import run_log
from common.yaml.yaml_exception import YamlException
from common.yaml.yaml_node import YamlNode, YamlNodeType


class YamlMethod:
    """
    功能：1. 读取层级简单的yaml，不支持读取列表里嵌套列表、"[" 和 "]"不在同一行的情况
         2. 将字典写入yaml，不支持列表里嵌套列表且嵌套的列表里存放的是字典等复杂场景
    """

    MAX_LINES = 300
    MAX_LEVEL = 5

    @staticmethod
    def load_yaml_info(yaml_path: str):
        """
        功能描述：根据yaml文件路径，读取yaml内容，并将其转化为字典
        参数：yaml_path 文件路径
        返回值：yaml文件转换后的字典
        """
        res = {}
        yaml_tree = YamlMethod._convert_yaml_to_tree(yaml_path)
        YamlMethod._convert2_dict_by_tree(yaml_tree, res)
        return res

    @staticmethod
    def dumps_yaml_file(source: dict, file_path: str, mode=0o600):
        """
        功能描述：将字典信息转化为yaml信息
        参数：source 待转换的字典
        参数：file_path yaml文件路径
        参数：mode yaml文件权限, 默认600
        返回值：NA
        """
        ret = FileCheck.check_input_path_valid(file_path)
        if not ret:
            run_log.error("yaml path is invalid, err: %s", ret.error)
            raise YamlException(f"yaml path is invalid.")

        res = []
        YamlMethod._convert2_yaml_info(source, res)
        file_info = "".join(res).rstrip("\n")

        with os.fdopen(os.open(file_path, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, mode), "w") as file:
            file.write(file_info)

    @staticmethod
    def _read_yaml(file_path: str) -> List[str]:
        """
        功能描述：读取yaml文件信息
        参数：file_path 文件路径
        返回值：lines 文件里的每行信息列表
        """
        ret = FileCheck.check_path_is_exist_and_valid(file_path)
        if not ret:
            run_log.error("Yaml path is invalid, err: %s.", ret.error)
            raise YamlException("Yaml path is invalid.")

        if os.path.getsize(file_path) > CommonConstants.OM_READ_FILE_MAX_SIZE_BYTES:
            run_log.error("Yaml %s content is too large", file_path)
            raise YamlException("Yaml content is too large")

        with open(file_path, "r") as yaml_file:
            lines = yaml_file.readlines()
            if len(lines) > YamlMethod.MAX_LINES:
                run_log.error("Yaml %s lines is too much.", file_path)
                raise YamlException("Yaml lines is too much!")

            return [line.rstrip("\n") for line in lines]

    @staticmethod
    def _generate_yaml_node(line: str) -> YamlNode:
        """
        功能描述：根据yaml文件的某一行信息生成yaml节点
        参数：line yaml文件的某一行
        返回值：yaml节点
        """
        start_with_list_flag = False
        without_space_line = line.lstrip()
        key_value = without_space_line.split(":")
        # yaml文件可能存在缩进一样，但也算子节点的情况，所以需要去掉'- '的长度后计算缩进
        # a:
        # - b: c
        if without_space_line.startswith("- "):
            space_len = len(line) - len(without_space_line) + 2
        else:
            space_len = len(line) - len(without_space_line)
        key = key_value[0]
        value = key_value[1].strip()
        if key.startswith("- "):
            key = key[2:]
            start_with_list_flag = True

        if not value and start_with_list_flag:
            return YamlNode(key, None, space_len, start_with_list_flag, YamlNodeType.LIST)

        if not value:
            return YamlNode(key, None, space_len, start_with_list_flag, None)

        if value.startswith("[") and value.endswith("]"):
            value = value[1: -1]
            res = []
            if value:
                res = value.split(",")
            return YamlNode(key, res, space_len, start_with_list_flag, YamlNodeType.LIST)
        contain_apostrophe = value.startswith("'") and value.endswith("'")
        contain_double_quotes = value.startswith('"') and value.endswith('"')
        if contain_apostrophe or contain_double_quotes:
            value = value[1: -1]
            return YamlNode(key, value, space_len, start_with_list_flag, YamlNodeType.NORMAL)

        if value.lower() in ("false", "no"):
            value = False
        elif value.lower() in ("true", "yes"):
            value = True
        return YamlNode(key, value, space_len, start_with_list_flag, YamlNodeType.NORMAL)

    @staticmethod
    def _can_convert2_node_or_not(line: str) -> bool:
        """
        功能： 判断某行能否转为节点
        参数： line yaml文件里的某行信息
        返回值： bool 是否能转换为节点
        """
        return (":" in line) and not ("'" in line and line.index(": ") > line.index("'"))

    @staticmethod
    def _deal_not_convert2_node(compare_node: YamlNode, line: str):
        """
        功能描述：当某一行不能转换为节点时的处理场景，比如-a -c这种
        参数：compare_node 比较节点
        参数： line yaml文件里的某一行
        返回值：NA
        """
        if not compare_node:
            run_log.error("Yaml content is invalid, pre_node is null!")
            raise YamlException("Yaml content is invalid, pre_node is null!")

        # compare_node还不确定是什么类型时的处理逻辑，比如生成compare_node节点时，lines："a:"
        if not compare_node.node_type:
            # 以'- '开头说明是list，这时将compare_node节点类型设置为LIST
            if line.lstrip().startswith("- "):
                compare_node.node_type = YamlNodeType.LIST
                compare_node.value = [line.strip()[2:]]
            # 以'['开头 ']'结尾说明是list，比如 a:\n [1,2,3]这时将compare_node节点类型设置为LIST
            elif line.lstrip().startswith("[") and line.rstrip().endswith("]"):
                compare_node.node_type = YamlNodeType.LIST
                list_value = line.lstrip()[1: -1]
                compare_node.value = [] if not list_value else list_value.split(",")
            # 否则节点类型是normal，如果是dict类型的话，line一定能转化为节点
            else:
                compare_node.node_type = YamlNodeType.NORMAL
                compare_node.value = line.strip()

        # 节点类型为NORMAL时处理逻辑，针对下述情况, 读取到c时，a节点类型确认为NORMAL,此时会把c拼接到b后面
        #  a:
        #    b
        #    c
        elif compare_node.node_type == YamlNodeType.NORMAL:
            if compare_node.value:
                compare_node.value = f"{compare_node.value} {line.lstrip()}"
            else:
                compare_node.value = line.lstrip()

        # 节点类型为LIST时处理场景，针对下述情况, 如果节点类型已经是LIST 那么line一定得是'- '开头，不考虑跨行的[]情况，比如
        # a:
        #   [1, 2, 3
        #   4]
        elif compare_node.node_type == YamlNodeType.LIST:
            if not line.lstrip().startswith("- "):
                run_log.error("Line need start with '- ' when compare_node type is list.")
                raise YamlException("Line need start with '- ' when compare_node type is list.")
            line = line.strip()[2:]
            compare_node.value.append(line)
        # 如果不能转为节点，且比较节点的类型为dict，说明yaml格式错误
        else:
            run_log.error("when compare_node is dict, the line must can convert to node!")
            raise YamlException("when compare_node is dict, the line must can convert to node!")

    @staticmethod
    def _convert_yaml_to_tree(yaml_path: str):
        """
        功能描述：根据yaml文件路径，读取yaml内容，并将其转化为以YamlNode组成的一棵树
        参数：yaml_path 文件路径
        返回值：根节点列表
        """
        # 用作比较
        compare_node: Optional[YamlNode] = None
        # 根节点列表
        root_nodes = []
        for line in YamlMethod._read_yaml(yaml_path):
            # line为空或者是注释内容则跳过
            if not line.strip() or line.strip().startswith("#"):
                continue

            # 不能转化为节点的处理方式
            if not YamlMethod._can_convert2_node_or_not(line):
                YamlMethod._deal_not_convert2_node(compare_node, line)
                continue

            # 将line转换为节点
            cur_node = YamlMethod._generate_yaml_node(line)

            # 如果是根节点，添加到返回结果后将比较节点设置为cur_node
            if cur_node.is_root_node():
                root_nodes.append(cur_node)
                compare_node = cur_node
                continue

            # 寻找当前节点的父节点，建立父节点和子节点的关系
            # a:
            #   b: 1  compare_node
            #   c: 2  cur_node
            fa_node = cur_node.get_father_node(compare_node)
            if not fa_node:
                run_log.error("Not found father node")
                raise YamlException("Not found father node!")

            if cur_node.list_flag:
                fa_node_type_is_wrong = fa_node.node_type and fa_node.node_type != YamlNodeType.LIST
                if fa_node_type_is_wrong and cur_node.node_type != YamlNodeType.LIST:
                    raise YamlException("Yaml content is invalid, fa_node type is wrong!")
                if not fa_node.node_type:
                    fa_node.node_type = YamlNodeType.LIST

            fa_node.next_nodes.append(cur_node)
            cur_node.pre_node = fa_node
            compare_node = cur_node
        return root_nodes

    @staticmethod
    def _convert2_dict_by_tree(yaml_nodes: List[YamlNode], res: dict, level: int = 0):
        """
        功能描述：将YamlNode组成的一棵树转化成字典
        实现逻辑：递归遍历树, 根据节点类型，转化为对应的value，为了防止无限制遍历，导致堆栈溢出，增加了层级限制<=5
        参数：yaml_nodes YamlNode组成的一棵树, res 转换后的字典
        返回值：NA
        """
        if level > YamlMethod.MAX_LEVEL:
            run_log.error("Yaml content is too completed, more than 5 level")
            raise YamlException("Yaml content is too completed, more than 5 level")

        for node in yaml_nodes:
            # 针对只有key，没有value的场景, 比如下面的a
            # a:
            # b: 2
            if not node.node_type:
                # 针对只有key，没有value的场景, 比如下面的a
                # a:
                # b: 2
                if not node.value and not node.next_nodes:
                    node.node_type = YamlNodeType.NORMAL
                    res[node.name] = ""
                    continue

                # 针对下面的a
                # a:
                #   b: 2
                if not node.value and node.next_nodes:
                    res[node.name] = {}
                    YamlMethod._convert2_dict_by_tree(node.next_nodes, res[node.name], level + 1)
                    continue

            if node.node_type == YamlNodeType.NORMAL:
                res[node.name] = node.value
                continue

            elif node.node_type == YamlNodeType.LIST:
                # 针对下面的a
                # a:
                # -2
                if not node.next_nodes:
                    res[node.name] = node.value
                    continue

                # 如果当前节点的第一个子节点不是以'- '开头，而后续有以'- '开头的节点，说明yaml文件内容不符合规范，比如
                # a:
                #   b: 1
                #   - c: 2
                first_next_node = node.next_nodes[0]
                list_flag_nodes = [node for node in node.next_nodes if node.list_flag]
                if not first_next_node.list_flag and list_flag_nodes:
                    raise Exception("First node not has list_flag but others has list_flag!")

                # 列表里存放字典
                res[node.name] = []
                temp_dict = None

                # 如果当前节点的第一个子节点以'- '开头, 有两种情况，分别是
                # a:   场景一        a:    场景二
                #   - b: 1            - b: 1
                #   c: 2              - c: 2
                # 场景一转换后 a是一个list，里面存放一个字典，key分别是b和c 场景2 a是一个list 里面会存放两个字典
                for next_node in node.next_nodes:
                    if next_node.list_flag:
                        temp_dict = {}
                        res[node.name].append(temp_dict)
                    if next_node.next_nodes:
                        temp_dict[next_node.name] = {}
                        YamlMethod._convert2_dict_by_tree(next_node.next_nodes, temp_dict[next_node.name], level + 2)
                    else:
                        temp_dict[next_node.name] = next_node.value
            else:
                # 节点类型为DICT
                res[node.name] = {}
                YamlMethod._convert2_dict_by_tree(node.next_nodes, res[node.name], level + 1)

    @staticmethod
    def _convert2_yaml_info(source: dict, res: list, level: int = 0):
        """
        功能描述：将字典信息转化为yaml信息
        实现逻辑：递归遍历字典，根据value的不同类型，将字典信息转换为一行一行的yaml信息
        参数：source 待转换的字典
        参数：res 存放转换后的yaml每行信息列表
        参数：level 层级
        返回值：yaml信息列表，存放yaml每一行
        """
        space = "  " * level
        if len(res) > YamlMethod.MAX_LINES or level > YamlMethod.MAX_LEVEL:
            run_log.error("Source is too large!")
            raise YamlException("source is too large")

        for key, value in source.items():
            if isinstance(value, dict):
                res.append(f"{space}{key}:\n")
                YamlMethod._convert2_yaml_info(value, res, level=level + 1)
            elif isinstance(value, list):
                if not value:
                    # 添加 - name: []
                    res.append(f"{space}{key}: []\n")
                    continue

                if not isinstance(value[0], dict):
                    res.append(f"{space}{key}:\n")
                    for item_value in value:
                        res.append(f"{space}- {item_value}\n")
                    continue

                res.append(f"{space}{key}:\n")

                for item in value:
                    # 获取第一个键值对, key特殊处理 增加"- "， 针对a:[{b: 1, c: 2}]
                    for index, item_key in enumerate(item):
                        head = f"{space}  {item_key}"
                        if index == 0:
                            head = f"{space}- {item_key}"
                        if not isinstance(item.get(item_key), dict):
                            res.append(f"{head}: {item.get(item_key)}\n")
                        else:
                            res.append(f"{head}:\n")
                            YamlMethod._convert2_yaml_info(item.get(item_key), res, level=level + 2)
            else:
                res.append(f"{space}{key}: {value}\n")
            if level == 0:
                res.append("\n")
