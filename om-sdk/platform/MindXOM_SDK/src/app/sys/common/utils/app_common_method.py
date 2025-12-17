# coding: UTF-8
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import base64
import fcntl
import os
import re
import shutil
import sqlite3

from common.checkers import IPV4Checker
from common.constants import error_codes
from common.constants.base_constants import CommonConstants
from common.exception.biz_exception import Exceptions
from common.file_utils import FileCheck
from common.utils.exec_cmd import ExecCmd
from common.utils.result_base import Result


class AppCommonMethod(object):
    OK = 200
    ERROR = 400
    NOT_EXIST = 404
    INTERNAL_ERROR = 500
    # 忽略的错误码，出现此错误，说明程序不支持
    IGNORE_ERROR = 800
    # 内部服务器错误
    INTERNAL_ERR_INFO = "Internal server error"
    # 任务失败给FD的通用提示信息
    COM_ERR_REASON_TO_FD = "ERR.600, Operate failed."

    FORBIDDEN_DOMAINS = (
        "localhost", "localhost.localdomain", "localhost4", "localhost4.localdomain4",
        "localhost6", "localhost6.localdomain6"
    )

    rootDir = None
    pDir = None

    @staticmethod
    def check_status_is_ok(ret_dict):
        return isinstance(ret_dict, dict) and ret_dict.get("status") == AppCommonMethod.OK

    @staticmethod
    def get_etc_host_content_lines():
        etc_hosts_file = "/etc/hosts"
        ret = FileCheck.check_path_is_exist_and_valid(etc_hosts_file)
        if not ret:
            return Result(False, err_msg=f"etc hosts file invalid: {ret.error}")

        if os.path.getsize(etc_hosts_file) > CommonConstants.OM_READ_FILE_MAX_SIZE_BYTES:
            return Result(False, err_msg="etc hosts file is too large")

        try:
            with open(etc_hosts_file) as file:
                fcntl.flock(file.fileno(), fcntl.LOCK_EX)
                return Result(True, data=file.readlines())
        except Exception as err:
            return Result(False, err_msg=f"get etc host content failed, find exception: {err}")

    @staticmethod
    def get_json_info(class_obj):
        """
        功能描述：通过类对象获取节点JSON
        参数：class_obj 类对象
        返回值：JSON文本
        异常描述：NA
        """
        ret = {}
        for key, value in vars(class_obj).items():
            if key.find("__local__") == 0 \
                    or key.find("_local_") == 0 or key == "items":
                # 私有变量不进行解析展示
                continue

            if key.find("OEM_Huawei_") > -1:
                key = key[11:]

            if isinstance(value, dict):
                ret[key] = value
            elif isinstance(value, list):
                ret[key] = value

            elif str(value).find("class") > -1:
                ret.update(AppCommonMethod.get_json_info(value))
            else:
                ret[key] = value

        return ret

    @staticmethod
    def get_project_absolute_path():
        """
        功能描述：获取工程的绝对路径
        参数：无
        返回值：工程的绝对路径
        异常描述：NAF
        """
        if AppCommonMethod.rootDir is None:
            working_dir = os.path.dirname(os.path.realpath(__file__))
            p_dir = os.path.dirname(working_dir)
            AppCommonMethod.rootDir = os.path.dirname(p_dir)

        return AppCommonMethod.rootDir

    @staticmethod
    def get_json_error_by_array(result, default_status=OK):
        """
        功能描述：获取Json形式的状态返回值
        参数：status 状态码
        message 错误信息
        返回值：无
        异常描述：NA
        """
        ret = dict()
        ret["status"] = default_status

        if result is None:
            ret["message"] = ""
        elif isinstance(result, dict):
            return result
        elif isinstance(result, list):
            if len(result) > 1:
                if isinstance(result[0], int):
                    if result[0] != 0 and result[0] != AppCommonMethod.OK:
                        if result[0] == AppCommonMethod.NOT_EXIST:
                            ret["status"] = AppCommonMethod.NOT_EXIST
                        else:
                            ret["status"] = AppCommonMethod.ERROR
                    ret["message"] = result[1]
                else:
                    ret["message"] = result
            else:
                ret["message"] = result
        else:
            ret["message"] = result

        return ret

    @staticmethod
    def force_remove_file(file_path):
        if not os.path.exists(file_path):
            return Result(result=True)

        if os.path.islink(file_path):
            os.remove(file_path)
            return Result(result=True)

        res = FileCheck.check_input_path_valid(file_path)
        if not res:
            return res

        if os.path.isdir(file_path):
            shutil.rmtree(file_path)
        else:
            os.remove(file_path)
        return Result(result=True)

    @staticmethod
    def replace_kv_list(adict, copy):
        """
        批量替换key和value

        @param adict JSON模板
        @param copy 替换的键值对列表
        @return 无
        """
        if isinstance(copy, list):
            for key in copy:
                AppCommonMethod.replace_kv(adict, key, copy[key])
        else:
            for key in copy.keys():
                if isinstance(copy[key], dict) and key not in adict:
                    AppCommonMethod.replace_kv_list(adict, copy[key])
                else:
                    AppCommonMethod.replace_kv(adict, key, copy[key])

    @staticmethod
    def replace_kv(adict, k, v):
        """
        替换JSON指定的Key和Value,支持复杂嵌套类型的替换

        @param adict JSON模板
        @param k 替换的key值
        @param v 替换内容
        @return 无
        """
        if isinstance(adict, list):
            for key in adict:
                if key == k:
                    adict[k] = v
                    return
                elif isinstance(key, dict):
                    AppCommonMethod.replace_kv(key, k, v)
                elif isinstance(key, list):
                    AppCommonMethod.replace_kv(key, k, v)
                else:
                    continue  # 其他类型不处理
        else:
            for key in adict.keys():
                if key == k:
                    adict[k] = v
                    return
                elif isinstance(adict[key], dict):
                    AppCommonMethod.replace_kv(adict[key], k, v)
                elif isinstance(adict[key], list):
                    AppCommonMethod.replace_kv(adict[key], k, v)
                else:
                    continue  # 其他类型不处理

    @staticmethod
    def check_ipv4_format(ipv4):
        """检查IP地址是否为IPV4地址"""
        # 非法地址：
        # 1、系统回环地址：127.0.0.0 - 127.255.255.255
        # 2、保留地址：0.0.0.0
        # 3、限制广播地址：255.255.255.255
        return all((IPV4Checker("ip").check({"ip": ipv4}), not ipv4.startswith("127."), ipv4 != "255.255.255.255"))

    @staticmethod
    def check_network_path(path) -> Result:
        """
        校验network相关路径。
        """
        if not path or not isinstance(path, str):
            return Result(result=False, err_msg="path is null or not string")

        if len(path) > 1024:
            return Result(result=False, err_msg="path length > 1024")

        pattern_name = re.compile(r"[^0-9a-zA-Z_./:-]")
        match_name = pattern_name.findall(path)
        if match_name or ".." in path:
            return Result(result=False, err_msg="there are illegal characters in path")

        ret = FileCheck.check_is_link(path)
        if not ret:
            return ret

        return Result(result=True)

    @staticmethod
    def hostname_check(hostname):
        pattern = r"^(?!-)[A-Za-z0-9\-]{1,63}(?<!-)$"
        return re.fullmatch(pattern, hostname)

    @staticmethod
    def check_input_parm(parm):
        """
        校验参数字符串是否存在特殊字符以及防止有更改路径的文件名

        @param parm 参数字符串
        @return True,False
        """
        if not parm:
            return True
        if not isinstance(parm, str):
            return False
        reg_str = "^[a-zA-Z0-9_.-]{0,255}$"
        pattern_str = re.compile(reg_str)
        if pattern_str.fullmatch(parm) is None or ".." in parm:
            return False
        else:
            return True

    @staticmethod
    def get_key_value_pair(org_str, split_char):
        """
        解析":"分割的键值对

        @param org_str 字符串
        @param split_char 分隔符
        @return [k, v]
        """
        if not org_str or not split_char:
            return ["", ""]
        idx = org_str.find(split_char)
        if idx == -1:
            return ["", ""]
        res = [org_str[:idx], org_str[(idx + 1):]]
        return res

    @staticmethod
    def get_fd_ip_from_etc_hosts(net_manager_domain):
        fd_ip = ""
        with open("/etc/hosts") as file:
            for line in file:
                if net_manager_domain in line and len(line.split()) == 2 and line.split()[1] == net_manager_domain:
                    fd_ip = line.split()[0]
                    return fd_ip
        return fd_ip

    @staticmethod
    def partition_id_check(partition_id):
        if len(partition_id) < 1 or len(partition_id) > 128:
            return False

        return bool(re.fullmatch(r"^[a-z0-9A-Z_]+$", partition_id))

    @staticmethod
    def convert_err_code_fd_format(error_code):
        # 小站上报的错误码格式与FD的定义规则匹配，在[100,299]范围内的需要前导0
        return f"{error_code:04d}" if 100 <= error_code <= 299 else f"{error_code}"

    @staticmethod
    def make_authentication_string(str_user, str_password):
        """
        功能描述：根据用户名和密返回认证字符串
        """
        if str_user and str_password:
            comstr = str_user + ':' + str_password
            base64string = base64.b64encode(comstr.encode()).decode('ascii')
            return "Basic {}".format(base64string)
        raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_PARAMETER_INVALID)

    @staticmethod
    def check_database_available(db_path: str) -> Result:
        if not os.path.exists(db_path) or not os.path.getsize(db_path):
            return Result(result=False, err_msg=f"database file {db_path} is not exists or is empty file.")

        try:
            with sqlite3.connect(database=db_path, check_same_thread=False) as conn:
                cur = conn.cursor()
                cur.execute("select name from sqlite_master where type='table' order by name")
        except Exception as err:
            return Result(result=False, err_msg=f"database file {db_path} is unavailable, {err}")

        return Result(result=True, err_msg=f"database file {db_path} available")

    @staticmethod
    def check_service_status_is_active(service_name: str) -> Result:
        """
        检查服务状态是否为active，查看服务状态普通权限即可
        :param service_name: 服务名称
        :return: Result: True--active, False--inactive
        """
        mef_service_status_cmd = ("systemctl", "is-active", service_name)
        status, out = ExecCmd.exec_cmd_get_output(mef_service_status_cmd)
        if status != 0:
            return Result(False, err_msg=f"exec cmd to get service {service_name} status failed: {status}")

        if not out or out.strip() != "active":
            return Result(False, err_msg=f"service {service_name} status is {out.strip()}")

        return Result(True)
