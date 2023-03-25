#!/usr/bin/env python3
# -*- coding:utf-8 -*-

# Copyright(C) Huawei Technologies Co.,Ltd. 2023. All rights reserved.

import json
import glob
import argparse
import os
from xml.etree import ElementTree


def pretty_xml(element, indent, newline, level=0):
    if element is not None:
        if element.text is None or element.text.isspace():
            element.text = newline + indent * (level + 1)
        else:
            element.text = newline + indent * (level + 1) + \
                           element.text.strip() + newline + \
                           indent * (level + 1)
    temp = list(element)
    for sub_element in temp:
        if temp.index(sub_element) < (len(temp) - 1):
            sub_element.tail = newline + indent * (level + 1)
        else:
            sub_element.tail = newline + indent * level
        pretty_xml(sub_element, indent, newline, level=level + 1)


def merge_xml(src, det, language):
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
    result_file_name = "unit-tests.xml"
    if language == 'python':
        result_file_name = "final.xml"
    xml_element_tree.write(os.path.join(det, result_file_name), encoding='utf-8', xml_declaration=True)


def merger_json(src, det):
    json_flies = glob.glob(os.path.join(src, "*.json"))
    json_dump = {
        "Packages": []
    }
    for json_file in json_flies:
        with open(json_file, 'r') as f:
            data = f.read()
            json_data = json.loads(data)
            for pkg in json_data['Packages']:
                json_dump['Packages'].append(pkg)
    json_w = json.dumps(json_dump)
    with open(os.path.join(det, "gocov.json"), 'w') as fw:
        fw.write(json_w)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="merge llt result")
    parser.add_argument('--src', type=str, default='')
    parser.add_argument('--det', type=str, default='')
    parser.add_argument('--language', type=str, default='go')
    args = parser.parse_args()
    if args.src == '' or args.det == '':
        print("parameter invalid!!!")
        exit(1)
    merge_xml(args.src, args.det, args.language)
    if args.language == 'go':
        merger_json(args.src, args.det)
    print("merger llt result done!")
