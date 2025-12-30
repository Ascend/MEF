# -*- coding:utf-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MEF is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import json
import glob
import argparse
import os
from xml.etree import ElementTree


def pretty_xml(element, indent, newline, level=0):
    if element is None:
        return

    if element.text is None or element.text.isspace():
        element.text = newline + indent * (level + 1)
    else:
        stripped_text = element.text.strip()
        element.text = newline + indent * (level + 1) + stripped_text + newline + indent * (level + 1)

    children = list(element)
    num_children = len(children)

    for i, child in enumerate(children):
        if i < num_children - 1:
            child.tail = newline + indent * (level + 1)
        else:
            child.tail = newline + indent * level
        pretty_xml(child, indent, newline, level + 1)


def merge_xml(src, det):
    xml_files = glob.glob(os.path.join(src, "*.xml"))
    xml_element_tree = None
    xml_element_root = None
    for xml_file in xml_files:
        tree = ElementTree.parse(xml_file)
        root = tree.getroot()
        if xml_element_tree is None:
            xml_element_tree = tree
            xml_element_root = root
        else:
            data = root.findall('testsuite')
            for d in data:
                xml_element_root.append(d)
    pretty_xml(xml_element_root, '\t', '\n')
    xml_element_tree.write(os.path.join(det, "unit-tests.xml"),
                           encoding='utf-8', xml_declaration=True)


def merger_json(src, det):
    json_flies = glob.glob(os.path.join(src, "*.json"))
    json_dump = {
        "Packages": []
    }
    for json_file in json_flies:
        with open(json_file, 'r') as f:
            data = f.read()
            json_data = json.loads(data)
            if json_data['Packages'] is None:
                continue
            for pkg in json_data['Packages']:
                json_dump['Packages'].append(pkg)
    json_w = json.dumps(json_dump)
    with open(os.path.join(det, "gocov.json"), 'w') as fw:
        fw.write(json_w)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="merge llt result")
    parser.add_argument('--src', type=str, default='')
    parser.add_argument('--det', type=str, default='')
    args = parser.parse_args()
    if args.src == '' or args.det == '':
        print("parameter invalid!!!")
        exit(1)
    merge_xml(args.src, args.det)
    merger_json(args.src, args.det)
    print("merger llt result done!")
