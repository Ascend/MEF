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
import os
import shlex
import threading
from typing import Iterable
from typing import NoReturn

from common.constants.base_constants import CommonConstants
from common.constants.upload_constants import UploadConstants
from common.file_utils import FileCheck, FileUtils
from common.file_utils import FileCopy
from common.file_utils import FilePermission
from common.init_cmd import cmd_constants
from common.log.logger import run_log
from common.utils.exec_cmd import ExecCmd
from lib.Linux.systems.security_service.puny_dict_sign_mgr import get_puny_dict_sign
from lib.Linux.systems.security_service.puny_dict_sign_mgr import set_puny_dict_sign
from common.common_methods import CommonMethods


class PunyDict:
    """
    功能描述：弱字典配置
    修改记录：2023-05-24 重构
    """

    PUNY_DICT_ERROR_MESSAGE = {
        "ERR.000": "ERR.000,Null",
        "ERR.001": "ERR.001,File does not exist.",
        "ERR.002": "ERR.002,File size is invalid.",
        "ERR.003": "ERR.003,File cannot be read or written.",
        "ERR.004": "ERR.004,File type error.",
        "ERR.005": "ERR.005,Import configuration puny dict failed.",
        "ERR.006": "ERR.006,File dir does not exist.",
        "ERR.007": "ERR.007,Export puny dict failed.",
        "ERR.008": "ERR.008,File not imported.",
        "ERR.009": "ERR.009,Operation type error.",
        "ERR.010": "ERR.010,File extension is not supported.",
        "ERR.011": "ERR.011,File name error.",
        "ERR.012": "ERR.012,Import file cannot empty.",
        "ERR.013": "ERR.013,request parameter is error.",
        "ERR.014": "ERR.014,PunyDict modify is busy.",
        "ERR.015": "ERR.015,Delete puny dict failed.",
    }
    OPERATION_TYPE_IMPORT = CommonConstants.TYPE_IMPORT
    OPERATION_TYPE_EXPORT = CommonConstants.TYPE_EXPORT
    OPERATION_TYPE_DELETE = CommonConstants.TYPE_DELETE
    PUNY_DICT_LOCK = threading.Lock()
    # 导入弱口令文件最大限制20MB
    FILE_SIZE_LIMIT = 20 * 1024 * 1024
    # 导入弱口令文件最大行数限制210万行
    FILE_LINE_LIMIT = 21 * 20 * 5000
    IMPORT_FILE = os.path.join(UploadConstants.CONF_UPLOAD_DIR, UploadConstants.WEAK_DICT_NAME)
    TMP_DELETE_FILE = "/run/delete_puny_dict.conf"
    TMP_IMPORT_FILE = "/run/temp_puny_dict.conf"
    TMP_EXPORT_FILE = "/run/export_puny_dict.conf"
    ENCODE_TYPE = "ISO-8859-1"
    MAX_LEN = 29

    def __init__(self):
        # 导出文件名，用于redfish进程处理
        self.export_file = ""

    @staticmethod
    def update_cracklib(import_file: str) -> NoReturn:
        """
        更新系统弱字典配置并设置相关文件权限
        :param import_file: 导入弱口令文件
        :return: NoReturn
        """
        # 更新系统弱字典配置
        cmd = f"{cmd_constants.OS_CMD_CRACKLIB_FORMAT} {shlex.quote(import_file)} | " \
              f"{cmd_constants.OS_CMD_CRACKLIB_PACKER} {CommonConstants.OS_PW_DICT_DIR}"
        cmd_ret = ExecCmd.exec_cmd_use_pipe_symbol(cmd, wait=15)
        if cmd_ret[0] != 0:
            raise OSError(f"Update weak dictionary file filed, reason is {cmd_ret[1]}")

        # 设置系统弱字典配置文件权限
        pw_dict_list = (
            "cracklib.magic", "cracklib-small.hwm", "cracklib-small.pwd",
            "cracklib-small.pwi", "pw_dict.hwm", "pw_dict.pwd", "pw_dict.pwi"
        )
        for pw_dict_file in pw_dict_list:
            file_path = os.path.join(CommonConstants.OS_CRACKLIB_DIR, pw_dict_file)
            # 系统默认权限为0o644，需要保持一致，避免出现other用户无法使用弱字典校验功能
            if not FilePermission.set_path_permission(file_path, 0o644):
                continue

    def delete_weak_dict(self):
        """
        删除弱口令
        """
        # 1、检查文件路径有效性
        check_ret = FileCheck.check_input_path_valid(self.TMP_DELETE_FILE)
        if not check_ret:
            run_log.error("Delete weak dictionary failed, %s", check_ret.error)
            return [CommonMethods.ERROR, self.PUNY_DICT_ERROR_MESSAGE.get("ERR.015")]

        # 2、写"abc"到弱字典文件
        try:
            with os.fdopen(os.open(self.TMP_DELETE_FILE, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600), "w") as file:
                file.write("abc")
        except Exception as err:
            run_log.error("Delete weak dictionary failed, %s", err)
            FileCopy.remove_path(self.TMP_DELETE_FILE)
            return [CommonMethods.ERROR, self.PUNY_DICT_ERROR_MESSAGE.get("ERR.003")]

        # 3、更新系统弱字典配置
        try:
            self.update_cracklib(self.TMP_DELETE_FILE)
        except Exception as err:
            run_log.error(err)
            return [CommonMethods.ERROR, self.PUNY_DICT_ERROR_MESSAGE.get("ERR.015")]
        finally:
            FileCopy.remove_path(self.TMP_DELETE_FILE)

        # 4、更新操作类型到数据库
        try:
            set_puny_dict_sign(self.OPERATION_TYPE_DELETE)
        except Exception as err:
            run_log.error("Delete weak dictionary failed, %s", err)
            return [CommonMethods.ERROR, self.PUNY_DICT_ERROR_MESSAGE.get("ERR.015")]

        return [CommonMethods.OK, "Delete weak dictionary success."]

    def export_weak_dict(self):
        """导出弱口令"""
        # 查询弱字典中是否有数据，若已经执行过删除弱口令操作则导出失败
        try:
            operation = get_puny_dict_sign()
        except Exception as err:
            run_log.error("Get operation type by database failed, catch %s", err.__class__.__name__)
            return [CommonMethods.ERROR, self.PUNY_DICT_ERROR_MESSAGE.get("ERR.007")]

        if operation == self.OPERATION_TYPE_DELETE:
            run_log.error("Export weak dictionary failed, the weak dictionary is deleted.")
            return [CommonMethods.ERROR, self.PUNY_DICT_ERROR_MESSAGE.get("ERR.008")]

        # 导出系统弱口令
        cracklib_unpacker = (cmd_constants.OS_CMD_CRACKLIB_UNPACKER, CommonConstants.OS_PW_DICT_DIR)
        try:
            with os.fdopen(os.open(self.TMP_EXPORT_FILE, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o644), "w+") as file:
                ret = ExecCmd.exec_cmd_get_output(cracklib_unpacker, stand_out=file, wait=30)
                if ret[0] != 0:
                    run_log.error("exec cracklib-unpack failed")
                    return [CommonMethods.ERROR, self.PUNY_DICT_ERROR_MESSAGE.get("ERR.007")]

            # 由于umask原因，创建文件时指定权限无效，需要重新修改权限
            os.chmod(self.TMP_EXPORT_FILE, 0o644)

        except Exception as err:
            run_log.error("export puny dict cmd failed, catch %s", err)
            FileCopy.remove_path(self.TMP_EXPORT_FILE)
            return [CommonMethods.ERROR, self.PUNY_DICT_ERROR_MESSAGE.get("ERR.007")]

        self.export_file = self.TMP_EXPORT_FILE
        return [CommonMethods.OK, "Export weak dictionary success."]

    def import_weak_dict(self):
        """
        导入弱口令
        """
        # 1、处理导入弱口令文件
        try:
            self.process_import_file()
        except Exception as err:
            run_log.error("Process import file failed, %s", err)
            FileCopy.remove_path(self.TMP_IMPORT_FILE)
            return [CommonMethods.ERROR, self.PUNY_DICT_ERROR_MESSAGE.get("ERR.005")]

        # 2、更新系统弱字典配置
        try:
            self.update_cracklib(self.TMP_IMPORT_FILE)
        except Exception as err:
            run_log.error(err)
            return [CommonMethods.ERROR, self.PUNY_DICT_ERROR_MESSAGE.get("ERR.005")]
        finally:
            FileCopy.remove_path(self.TMP_IMPORT_FILE)

        # 3、更新操作类型到数据库
        try:
            set_puny_dict_sign(self.OPERATION_TYPE_IMPORT)
        except Exception as err:
            run_log.error("Import weak dictionary failed, %s", err)
            return [CommonMethods.ERROR, self.PUNY_DICT_ERROR_MESSAGE.get("ERR.005")]

        return [CommonMethods.OK, "Import weak dictionary success."]

    def process_import_file(self) -> NoReturn:
        """
        处理导入弱口令文件，并检查文件的合法性
        :return:
        """
        # 1、检查导入弱口令文件有效性
        check_ret = FileCheck.check_path_is_exist_and_valid(self.IMPORT_FILE)
        if not check_ret:
            raise ValueError(f"Check import file is invalid, {check_ret.error}")

        # 2、导入文件大小不能大于20MB
        if os.path.getsize(self.IMPORT_FILE) > self.FILE_SIZE_LIMIT:
            raise ValueError("Check import file size more than 20 MB.")

        # 3、处理文件之后并写入新的内容
        contents = "\n".join(self.read_import_file(self.IMPORT_FILE))
        if not contents:
            raise ValueError("Check import file contents is null.")

        open_mode = os.O_WRONLY | os.O_CREAT | os.O_TRUNC
        with os.fdopen(os.open(self.TMP_IMPORT_FILE, open_mode, 0o600), "w", encoding=self.ENCODE_TYPE) as f_out:
            f_out.write(contents)

    def read_import_file(self, import_file: str) -> Iterable[str]:
        """
        处理导入文件，最大文件行数限制210万行，若单行超过29个字符时，直接丢弃
        EulerOS 2.0和openEuler 22.03超过29个字符时系统直接丢弃，Ubuntu 22.04超过30个字符时，系统会截取前30个字符保存
        :param import_file: 导入弱口令文件
        :return: Iterable[处理完毕之后行数据]
        """
        with open(import_file, encoding=self.ENCODE_TYPE) as f_in:
            for line_num, data in enumerate(f_in):
                if line_num > self.FILE_LINE_LIMIT:
                    run_log.error("Read file stop, maximum number of lines in a file.")
                    raise StopIteration

                if len(data) > self.MAX_LEN or not data.strip():
                    continue
                yield data.strip()

    def post_request(self, request_dict: dict):
        """
        PunyDict接口post请求处理函数
        :param request_dict: post请求数据
        :return: [返回码, 错误信息]
        """
        if self.PUNY_DICT_LOCK.locked():
            run_log.warning("PunyDict modify is busy.")
            return [CommonMethods.ERROR, self.PUNY_DICT_ERROR_MESSAGE.get("ERR.014")]

        with self.PUNY_DICT_LOCK:
            # 操作类型校验
            oper_type = request_dict.get("OperationType")
            if oper_type not in (self.OPERATION_TYPE_IMPORT, self.OPERATION_TYPE_EXPORT, self.OPERATION_TYPE_DELETE):
                run_log.error("Check operation type is invalid.")
                return [CommonMethods.ERROR, self.PUNY_DICT_ERROR_MESSAGE.get("ERR.009")]

            # 根据操作类型执行对应操作
            return {
                self.OPERATION_TYPE_IMPORT: self.import_weak_dict,
                self.OPERATION_TYPE_EXPORT: self.export_weak_dict,
                self.OPERATION_TYPE_DELETE: self.delete_weak_dict,
            }.get(oper_type)()

    def delete_request(self, *args):
        """当redfish拷贝好文件后，发送delete请求，删除临时日志文件"""
        with self.PUNY_DICT_LOCK:
            try:
                FileUtils.delete_file_or_link(self.TMP_EXPORT_FILE)
            except Exception as err:
                run_log.error("delete %s failed: %s", self.TMP_EXPORT_FILE, err)
