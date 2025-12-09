#!/usr/bin/python
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.

"""
功能：基础类定义，包含一些公共属性
修改记录：2016-11-12 创建
"""
import ast
import json
import logging
import os
import re

from common.constants.base_constants import CommonConstants
from common.file_utils import FileCheck
from common.log.logger import run_log
from common.utils.app_common_method import AppCommonMethod
from common.utils.result_base import Result


class CommonMethods(object):
    """
    功能描述：公共函数定义类
    接口：NA
    修改记录：2016-11-12 创建
    """

    OK = 200
    PARTIAL_OK = 206
    ERROR = 400
    NOT_EXIST = 404
    INTERNAL_ERROR = 500
    # 忽略的错误码，出现此错误，说明程序不支持
    IGNORE_ERROR = 800

    NET_TAG_INI = "/home/data/ies/tag.ini"
    NFS_MAX_CFG_NUM = 32

    rootDir = None
    pDir = None

    @staticmethod
    def object_to_json(status, message):
        """
        功能描述：获取Json形式的状态返回值
        参数：status 状态码
        message 错误信息
        返回值：无
        异常描述：NA
        修改记录：2016-11-12 创建
        """

        ret = {}
        ret["status"] = status
        ret["message"] = message

        return ret

    @staticmethod
    def get_value_by_key(value, key1, key2=None, last_match=False, is_find=False):
        """
        功能描述：通过关键键值获取对应的值
        参数：value  原始字符串
        key1  关键键值
        key2  通过此字符串进行分割，获取第一个值；为 None 时，只根据key1进行获取
        last_match  指定是否逆向匹配
        is_find  指定是否采用 find 模式；
        True：返回关键键值之后的所有内容（去掉前后空格）
        False：使用正则表达式匹配返回一行信息
        返回值：获取到的字符串值，如果没有找到，则返回 ""
        异常描述：NA
        修改记录：2016-11-14 创建
        2016-11-24 增加  is_find 参数
        2017-02-08 修改 value 为 None,直接返回 None
        """

        if not value or len(value) > 2048:
            return ""

        if is_find:
            n_pos = value.find(key1)
            if n_pos == -1:
                return ""

            n_pos = n_pos + len(key1)
            value = value[n_pos:].strip()
        else:
            pattern = "^.*" + key1 + "(.*)$"
            ret = re.findall(pattern, value, re.M)
            if ret is None or not ret:
                return ""

            value = ret[0].strip()
        if key2 is None:
            return value.strip()

        index = value.find(key2)
        if index < 0:
            # 没有找到匹配的字符串，直接返回
            return value.strip()

        if last_match:
            index = value.rfind(key2)
            if index < 0:
                # 没有找到匹配的字符串，直接返回
                return value.strip()
            else:
                return value[index + 1:].strip()
        else:
            index = value.find(key2)
            if index < 0:
                # 没有找到匹配的字符串，直接返回
                return value.strip()
            else:
                return value[0: index].strip()

    @staticmethod
    def check_json_data(json_data):
        """
        功能描述：检查数据是否合法
        参数：jsonData json数据
        返回值：是否合法
        异常描述：NA
        修改记录：2016-12-23 创建
        """
        if json_data is None:
            return [0, json_data]

        if isinstance(json_data, dict):
            return [0, json_data]

        try:
            # 通过转换判断字符串是否为 json
            n_dict = json.loads(json_data)
        except Exception:
            return [1, "Request data is not json."]

        if not isinstance(n_dict, dict):
            return [1, "Request data is not json."]

        try:
            # 通过加载以及钩子函数检查json文本是否存在重复的键值
            json.loads(json_data, object_pairs_hook=CommonMethods.check_duplicate_attributes)
        except Exception as e:
            # 参数出现重复
            return [1, str(e)]

        return [0, n_dict]

    @staticmethod
    def check_duplicate_attributes(lst):
        """
        功能描述：钩子函数，检查 json文本中是否存在重复的 key
        参数：jsonData json数据
        返回值：是否合法
        异常描述：键重复时，抛出异常
        修改记录：2016-12-23 创建
        """
        tmp_dict = {}
        for key, val in lst:
            if key in tmp_dict:
                # 参数出现重复
                raise Exception("Duplicate attributes: %s." % key)
            else:
                tmp_dict[key] = val

        return tmp_dict

    @staticmethod
    def get_config_value(section_name, key, file="config/iBMA.ini"):
        """
        功能描述：从配置文件中获取指定的 key值
        参数：sectionName分段定义
             key 配置的键值
             file 配置文件相对于工程目录的路径
        返回值：值
        异常描述：NA
        修改记录：2017-01-02创建
                2017-12-12增加file参数,默认为iBMA.ini
        """

        import configparser

        value = None
        conf_parser = configparser.ConfigParser()
        file = os.path.join(AppCommonMethod.get_project_absolute_path(), file)
        if not FileCheck.check_path_is_exist_and_valid(file):
            logging.error("file path check failed!")
            return value
        try:
            conf_parser.read(file, encoding="utf-8")
            if conf_parser.has_section(section_name) and conf_parser.has_option(section_name, key):
                value = conf_parser.get(section_name, key)

        except IOError as ioerr:
            logging.error("IOERROR (%s)", ioerr.errno)
        except ValueError:
            logging.error("Could not convert data to an integer.")
        except Exception as e:
            logging.error("unexpected error %s", e)

        return value

    @staticmethod
    def load_net_tag_ini() -> Result:
        ret = FileCheck.check_path_is_exist_and_valid(CommonMethods.NET_TAG_INI)
        if not ret:
            return ret

        ret = FileCheck.check_path_is_root(CommonMethods.NET_TAG_INI)
        if not ret:
            return ret

        if os.path.getsize(CommonMethods.NET_TAG_INI) > CommonConstants.OM_READ_FILE_MAX_SIZE_BYTES:
            return Result(False, err_msg="file size is too large")

        try:
            with open(CommonMethods.NET_TAG_INI, "r") as fd:
                tag_ini_content = ast.literal_eval(fd.read())

            return Result(True, data=tag_ini_content)
        except Exception as err:
            return Result(False, err_msg=str(err))

    @staticmethod
    def write_net_tag_ini(tag_ini_data: str) -> Result:
        ret = FileCheck.check_input_path_valid(CommonMethods.NET_TAG_INI)
        if not ret:
            return ret

        try:
            with os.fdopen(os.open(CommonMethods.NET_TAG_INI, os.O_WRONLY | os.O_TRUNC | os.O_CREAT, 0o640), "w") as fd:
                fd.write(tag_ini_data)

            return Result(True)
        except Exception as err:
            return Result(False, err_msg=str(err))

    @staticmethod
    def _is_need_change_owner(path):
        if os.path.islink(path):
            if os.stat(path, follow_symlinks=False).st_uid != 0 and os.stat(path, follow_symlinks=False).st_gid != 0:
                run_log.warning("The path is soft link and the owner is not root, not need chown")
                return False
        return True
