# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import configparser
import fcntl
import grp
import os
import pwd
import re
import shutil
import stat
import tarfile
import threading
import zipfile
from pathlib import Path
from typing import List, Dict, Any, AnyStr, Union, NoReturn

from common.constants import error_codes
from common.constants.base_constants import CommonConstants
from common.exception.biz_exception import Exceptions
from common.utils.exception_utils import OperateBaseError
from common.utils.exec_cmd import ExecCmd
from common.utils.result_base import Result

_lock = threading.Lock()


class CompressedFileCheckUtils:
    MAX_EXTRACT_FILE_COUNT = 6000
    MAX_EXTRACT_FILE_SIZE = 1 * 1024 * 1024 * 1024

    @staticmethod
    def get_zip_file_info(file_path, check_symbolic_link=False):
        total_size = 0
        file_names = []
        ret = CompressedFileCheckUtils._check_number_of_zip_file(file_path)
        if not ret:
            return Result(result=False, data=[0, [""]], err_msg=ret.error)

        with zipfile.ZipFile(file_path, "r") as file_list:
            for file in file_list.infolist():
                ret = CompressedFileCheckUtils._check_zip_file_symbolic_link(file, check_symbolic_link)
                if not ret:
                    return Result(result=False, data=[0, [""]], err_msg="zip inner file have symbolic link")

                ret = CompressedFileCheckUtils._check_zip_file_suid_and_sgid(file)
                if not ret:
                    return Result(result=False, data=[0, [""]], err_msg=ret.error)

                total_size += file.file_size
                file_names.append(file.filename)

        return Result(result=True, data=[total_size, file_names])

    @staticmethod
    def get_tar_file_info(file_path, file_mode, check_symbolic_link=False):
        total_size = 0
        circle_count = 0
        file_names = []
        with tarfile.open(file_path, file_mode) as file_list:
            for file in file_list:
                circle_count += 1
                if circle_count > CompressedFileCheckUtils.MAX_EXTRACT_FILE_COUNT:
                    return Result(result=False, data=[0, [""]], err_msg="too many tar extract file count")

                ret = CompressedFileCheckUtils._check_tar_file_symbolic_link(file, check_symbolic_link)
                if not ret:
                    return Result(result=False, data=[0, [""]], err_msg="tar inner file have symbolic link")

                ret = CompressedFileCheckUtils._check_tar_file_suid_and_sgid(file)
                if not ret:
                    return Result(result=False, data=[0, [""]], err_msg=ret.error)

                total_size += file.size
                file_names.append(file.name)

        return Result(result=True, data=[total_size, file_names])

    @staticmethod
    def check_compressed_file_valid(
            file_path,
            check_symbolic_link=False,
            extract_path="/run",
            max_extract_file_cnt=MAX_EXTRACT_FILE_COUNT,
    ):
        try:

            if file_path.endswith((".tar.gz", ".tar")):
                file_mode = "r:gz" if file_path.endswith(".gz") else "r:tar"
                ret = CompressedFileCheckUtils.get_tar_file_info(file_path, file_mode, check_symbolic_link)
                if not ret:
                    error_msg = f"check compressed file failed: {ret.error}"
                    return False, error_msg

                file_size, file_names = ret.data[0], ret.data[1]
                return CompressedFileCheckUtils._check_compressed_file_attribute(
                    file_size, file_names, extract_path, max_extract_file_cnt)
            elif file_path.endswith(".zip"):
                ret = CompressedFileCheckUtils.get_zip_file_info(file_path, check_symbolic_link)
                if not ret:
                    error_msg = f"check compressed file failed: {ret.error}"
                    return False, error_msg

                file_size, file_names = ret.data[0], ret.data[1]
                return CompressedFileCheckUtils._check_compressed_file_attribute(
                    file_size, file_names, extract_path, max_extract_file_cnt)
            else:
                return False, "unsupported compressed file format"
        except Exception as e:
            return False, f"check compressed file failed: {e}"

    @staticmethod
    def check_image_inner_file_name(image_file):
        try:
            if image_file.endswith((".tar.gz", ".tar")):
                file_mode = "r:gz" if image_file.endswith(".gz") else "r:tar"
                ret = CompressedFileCheckUtils.get_tar_file_info(image_file, file_mode)
                if not ret:
                    error_msg = f"check image inner file name compressed file failed: {ret.error}"
                    return False, error_msg
                file_names = ret.data[1]
            elif image_file.endswith(".zip"):
                ret = CompressedFileCheckUtils.get_zip_file_info(image_file)
                if not ret:
                    error_msg = f"check image inner file name compressed file failed: {ret.error}"
                    return False, error_msg
                file_names = ret.data[1]
            else:
                return False, "unsupported image file format"

            white_list = (".crl", ".tar.gz", ".cms")
            for file_name in file_names:
                if not file_name.endswith(white_list):
                    return False, "invalid image inner file name"

            return True, ""
        except Exception as e:
            return False, f"check image inner file name failed: {e}"

    @staticmethod
    def _check_number_of_zip_file(file_path):
        count_cmd = f"unzip -l {file_path} | wc -l"
        ret_code, msg = ExecCmd.exec_cmd_use_pipe_symbol(count_cmd)
        if ret_code != 0:
            error_msg = f"Execute count zip package file cmd failed: {ret_code}"
            return Result(result=False, err_msg=error_msg)

        file_num = int(msg)
        if file_num > CompressedFileCheckUtils.MAX_EXTRACT_FILE_COUNT:
            error_msg = f"Execute count zip package file cmd failed: too many zip extract file count"
            return Result(result=False, err_msg=error_msg)

        return Result(True)

    @staticmethod
    def _check_avail_space_enough(file_size, extract_path):
        st = os.statvfs(extract_path)
        avail_space = st.f_bsize * st.f_bavail
        if file_size > avail_space:
            return False

        return True

    @staticmethod
    def _check_tar_file_symbolic_link(file_info, check_symbolic_link):
        if check_symbolic_link:
            if file_info.issym():
                return False, "inner file have symbolic link"

        return True, ""

    @staticmethod
    def _check_zip_file_symbolic_link(file_info, check_symbolic_link):
        # external_attr表示zip中该文件的外部属性，包括目录，符号链接，文件权限，形如：lrwxrwxrwx，drwxr-xr-x，-rw-rw-r--
        # 0o120000为符号链接的权限模式前缀，加上文件权限就是0o120777，使用os.lstat(符号链接路径).st_mode查看符号链接权限模式
        # 0o40000为目录的权限模式前缀，加上文件权限就是0o40755，使用os.stat(目录路径).st_mode查看目录权限模式
        # 0o100000为普通文件的权限模式前缀，加上文件权限就是0o100664，使用os.stat(普通文件路径).st_mode查看普通文件权限模式
        # external_attr=0，然后分文件和目录处理不同
        # external_attr |= (权限模式) << 16，然后目录为了兼容ms-dos会再来一下：external_attr |= 0x10
        # 所以判断zip文件中的文件是否为符号链接，只需要external_attr > 0o120000 << 16即可，前提是文件类型是ZIP_STORED
        if check_symbolic_link:
            if file_info.compress_type == zipfile.ZIP_STORED and file_info.external_attr > 0o120000 << 16:
                return False, "inner file have symbolic link"
        return True, ""

    @staticmethod
    def _check_tar_file_suid_and_sgid(file_info):
        if file_info.mode & 0o4000 == 0o4000:
            return Result(result=False, err_msg="tar inner file have SUID file")

        if file_info.mode & 0o2000 == 0o2000:
            return Result(result=False, err_msg="tar inner file have SGID file")

        return Result(True)

    @staticmethod
    def _check_zip_file_suid_and_sgid(file_info):
        tmp_attr = file_info.external_attr >> 16
        if tmp_attr & 0o4000 == 0o4000:
            return Result(result=False, err_msg="zip inner file have SUID file")

        if tmp_attr & 0o2000 == 0o2000:
            return Result(result=False, err_msg="zip inner file have SGID file")

        return Result(True)

    @staticmethod
    def _check_package_inner_file_name(inner_file_name):
        check_str_list = ["../", "..\\", ".\\", "./", "~/"]
        for string in check_str_list:
            if string in inner_file_name:
                return False
        return True

    @staticmethod
    def _check_compressed_file_attribute(file_size, file_names, extract_path, max_extract_file_cnt):
        if len(file_names) > max_extract_file_cnt:
            error_msg = "too many extract file count"
            return False, error_msg

        if file_size > CompressedFileCheckUtils.MAX_EXTRACT_FILE_SIZE:
            error_msg = "too much extract file size"
            return False, error_msg

        if not CompressedFileCheckUtils._check_avail_space_enough(file_size, extract_path):
            return False, "check compressed file failed, space not enough."

        for file_name in file_names:
            if not CompressedFileCheckUtils._check_package_inner_file_name(file_name):
                err_msg = "check compressed file failed, inner file path has special string"
                return False, err_msg

            if os.path.isabs(file_name):
                err_msg = "check compressed file failed, inner file can not be abspath"
                return False, err_msg

        return True, ""


class FileOperateError(OperateBaseError):
    pass


class FileCreate:
    @staticmethod
    def create_file(file_path, mode) -> Result:
        try:
            ret = FileCheck.check_input_path_valid(file_path)
            if not ret:
                return Result(result=False, err_msg=f"create file failed, the file path not valid: {ret.error}")

            if not os.path.exists(file_path):
                with os.fdopen(os.open(file_path, os.O_CREAT, mode), "w"):
                    pass

            return FilePermission.set_path_permission(file_path, mode)
        except Exception as e:
            return Result(result=False, err_msg="create file failed: {}".format(e))

    @staticmethod
    def create_dir(dir_path, mode):
        try:
            if not FileCheck.check_system_key_path(dir_path):
                error_msg = f"Create dir failed: input dir path is system key dir path"
                return Result(result=False, err_msg=error_msg)

            ret = FileCheck.check_input_path_valid(dir_path)
            if not ret:
                return Result(result=False,
                              err_msg=f"Create dir failed, the file path not valid: {ret.error}")

            os.makedirs(dir_path, mode=mode, exist_ok=True)
            return FilePermission.set_path_permission(dir_path, mode)
        except Exception as e:
            return Result(result=False, err_msg="create dir failed: {}".format(e))


class FileOperator:
    @staticmethod
    def extra_tar_file(file, path, check_symbolic_link=False, no_same_owner=False):
        """
        解压tar.gz压缩文件

        @param file tar.gz文件
        @param path 解压目标路径
        @param check_symbolic_link 是否检查符号链接
        @param no_same_owner 解压文件的属主与解压进程的属主保持一致
        """
        try:
            ret = FileCheck.check_input_path_valid(file)
            if not ret:
                return Result(result=False, err_msg=f"check compressed file path failed:{ret.error}")

            ret, error_msg = CompressedFileCheckUtils.check_compressed_file_valid(file, check_symbolic_link, path)
            if not ret:
                return Result(result=False, err_msg=error_msg)

            with tarfile.open(file) as tar:
                if no_same_owner:
                    tar.extractall(path, members=FileOperator._to_cur_effective_user(tar))
                else:
                    tar.extractall(path)

            return Result(result=True)
        except Exception as e:
            err_msg = f"extra tar file error {e}"
            return Result(result=False, err_msg=err_msg)

    @staticmethod
    def extra_zip_file(file, path, check_symbolic_link=False):
        try:
            ret = FileCheck.check_input_path_valid(file)
            if not ret:
                return Result(result=False, err_msg=f"check compressed file path failed:{ret.error}")

            ret, error_msg = CompressedFileCheckUtils.check_compressed_file_valid(file, check_symbolic_link, path)
            if not ret:
                return Result(result=False, err_msg=error_msg)

            with zipfile.ZipFile(file) as zip_file:
                zip_file.extractall(path)

            return Result(result=True)
        except Exception as e:
            err_msg = f"extra zip file error {e}"
            return Result(result=False, err_msg=err_msg)

    @staticmethod
    def _to_cur_effective_user(tar_members):
        cur_e_uid = os.geteuid()
        cur_e_uname = pwd.getpwuid(cur_e_uid).pw_name
        cur_e_gid = os.getegid()
        cur_e_gname = grp.getgrgid(cur_e_gid).gr_name
        for tar_item in tar_members:
            if tar_item.uid != cur_e_uid or tar_item.gid != cur_e_gid:
                tar_item.uid, tar_item.gid = cur_e_uid, cur_e_gid
                tar_item.uname, tar_item.gname = cur_e_uname, cur_e_gname
            yield tar_item


class FileUtils(object):
    @staticmethod
    def get_config_parser(file_path: AnyStr) -> configparser.ConfigParser:
        """

        @param file_path:
        @return:
        """
        if not os.path.exists(file_path):
            raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_FILE_PATH_NOT_EXIST, file_path)
        config_parser = configparser.ConfigParser()
        with open(file_path, "r", encoding="utf-8") as fr:
            for index, _ in enumerate(fr):
                if index > CommonConstants.MAX_ITER_LIMIT:
                    raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_FILE_LINES)

        config_parser.read(file_path)
        return config_parser

    @staticmethod
    def get_section_list(file_path: str) -> List[str]:
        config_parser = FileUtils.get_config_parser(file_path)
        return config_parser.sections()

    @staticmethod
    def get_option_list(file_path: AnyStr, section: AnyStr) -> Dict[AnyStr, Any]:
        option_dict = {}
        if not FileCheck.check_path_is_exist_and_valid(file_path):
            raise Exceptions.biz_exception(error_codes.FileErrorCodes.ERROR_FILE_PATH_CHECK, file_path, section)
        config_parser = FileUtils.get_config_parser(file_path)
        if not config_parser.has_section(section):
            raise Exceptions.biz_exception(error_codes.FileErrorCodes.ERROR_FILE_NO_SECTION, file_path, section)
        option_list = config_parser.options(section)
        for option in option_list:
            value = config_parser.get(section, option)
            option_dict[option] = value
        return option_dict

    @staticmethod
    def get_option(file_path: str, section: str, option: str) -> AnyStr:
        if not FileCheck.check_path_is_exist_and_valid(file_path):
            raise Exceptions.biz_exception(error_codes.FileErrorCodes.ERROR_FILE_PATH_CHECK, file_path, section)
        config_parser = FileUtils.get_config_parser(file_path)
        if not config_parser.has_option(section, option):
            raise Exceptions.biz_exception(error_codes.FileErrorCodes.ERROR_FILE_NO_OPTION, file_path, section)
        return config_parser.get(section, option)

    @staticmethod
    def check_section_exist(file_path: str, section) -> bool:
        config_parser = FileUtils.get_config_parser(file_path)
        return config_parser.has_section(section)

    @staticmethod
    def add_one_section(file_path, section, option_list: dict):
        """
        向文件中添加一个section
        :param file_path: 要操作的文件
        :param section:要操作的 section
        :param option_list:要操作的 option集合,存放的是option的{key,value}
        """
        # 判断要添加section的文件是否已经存在要添加的section
        if FileUtils.check_section_exist(file_path, section):
            raise Exceptions.biz_exception(error_codes.FileErrorCodes.ERROR_FILE_SECTION_EXIST, file_path, section)
        config_parser = configparser.ConfigParser()
        config_parser.add_section(section)
        for key, value in option_list.items():
            config_parser.set(section, key, value)
        FileUtils.operate_file(config_parser, file_path, "a+")

    @staticmethod
    def modify_one_option(file_path: AnyStr, section: AnyStr, option_key: AnyStr, option_value: AnyStr):
        """
        对文件已经存在的section里的某个option进行修改
        :param file_path: 要操作的文件
        :param section:要操作的 section
        :param option_key:要操作的 option key
        :param option_value:要操作的 option value
        """
        # 判断要添加section的文件是否已经存在要添加的section
        if not FileCheck.check_path_is_exist_and_valid(file_path):
            raise Exceptions.biz_exception(error_codes.FileErrorCodes.ERROR_FILE_PATH_CHECK, file_path, section)
        if not FileUtils.check_section_exist(file_path, section):
            raise Exceptions.biz_exception(error_codes.FileErrorCodes.ERROR_FILE_NO_SECTION, file_path, section)
        config_parser = FileUtils.get_config_parser(file_path)
        config_parser.set(section, option_key, option_value)
        FileUtils.operate_file(config_parser, file_path, "w+")

    @staticmethod
    def operate_file(config_parser, file_path, mode):
        if not FileCheck.check_is_link(file_path):
            raise Exceptions.biz_exception(error_codes.FileErrorCodes.ERROR_FILE_PATH_CHECK, file_path)
        with open(file_path, mode) as file:
            config_parser.write(file)

    @staticmethod
    def read_file(file_path: AnyStr, mode: AnyStr) -> List[AnyStr]:
        """
        读取指定文件的所有内容
        @param file_path: 文件所在的路径
        @param mode:
        @return:
        """
        line_list = []
        if not FileCheck.check_path_is_exist_and_valid(file_path):
            return line_list
        with open(file_path, mode) as read_file:
            for line in read_file:
                line_list.append(line.rstrip())
        return line_list

    @staticmethod
    def delete_dir_content(dir_path) -> Result:
        """删除指定目录的内容，但不包括目录本身"""
        if not os.path.exists(dir_path):
            return Result(result=True)
        ret = FileCheck.check_path_is_exist_and_valid(dir_path)
        if not ret:
            return Result(result=False, err_msg=f"check path failed, {ret.error}")
        try:
            file_list = os.listdir(dir_path)
            for file in file_list:
                file_path = os.path.join(dir_path, file)
                if os.path.isfile(file_path) or os.path.islink(file_path) or Path(file_path).is_socket():
                    os.remove(file_path)
                elif os.path.isdir(file_path):
                    shutil.rmtree(file_path)
        except Exception as err:
            return Result(result=False, err_msg=f"delete dir content failed, {err}")

        return Result(result=True)

    @staticmethod
    def delete_file_or_link(filename: Union[str, Path]):
        """删除指定的文件"""
        if not os.path.exists(filename):
            return
        if os.path.isfile(filename) or os.path.islink(filename) or Path(filename).is_socket():
            os.remove(filename)

    @staticmethod
    def write_file_with_lock(file_path, deal_func, *args, **kwargs):
        ret = FileCheck.check_input_path_valid(file_path)
        if not ret:
            return Result(result=False, err_msg=f"write file failed: check file path failed: {ret.error}")

        # 文件不存在则创建
        if not os.path.exists(file_path):
            flags = os.O_WRONLY | os.O_CREAT | os.O_TRUNC
            modes = stat.S_IWUSR | stat.S_IRUSR
            with os.fdopen(os.open(file_path, flags, modes), "w") as file:
                fcntl.flock(file.fileno(), fcntl.LOCK_EX)
                return deal_func(file, *args, **kwargs)

        # 文件存在则以"r+"的方式打开文件，不能够采用"w"的方式读写，因为还没有锁文件就已经对文件清空，竟态场景会读到空文件
        with open(file_path, "r+") as file:
            fcntl.flock(file.fileno(), fcntl.LOCK_EX)
            file.truncate(0)  # 锁住文件后，再对文件清零
            return deal_func(file, *args, **kwargs)

    @staticmethod
    def delete_full_dir(dir_path) -> NoReturn:
        """删除指定目录的内容，包括目录本身"""
        if not os.path.exists(dir_path):
            return
        ret = FileUtils.delete_dir_content(dir_path)
        if not ret:
            raise FileOperateError(f"Delete {os.path.basename(dir_path)} content failed, {ret.error}")
        # 删除最外层dir_path目录本身
        os.rmdir(dir_path)

    @staticmethod
    def check_script_file_valid(sh_file, user=None, group=None):
        if not sh_file:
            return Result(result=False, err_msg="null path")

        ret = FileCheck.check_path_is_exist_and_valid(sh_file)
        if not ret:
            return Result(result=False, err_msg=f"{sh_file} path invalid: {ret.error}")

        if stat.S_IWOTH & os.stat(sh_file).st_mode == stat.S_IWOTH:
            return Result(result=False, err_msg=f"The perm of the path is invalid")

        ret = FileCheck.check_path_mode_owner_group(sh_file, user=user, group=group)
        if not ret:
            return Result(result=False, err_msg=f"{sh_file} path owner or group is invalid")

        return Result(True)


class FileCheck:
    @staticmethod
    def is_exists(path):
        if not os.path.exists(path):
            return Result(False, err_msg=f"{path} not exists.")
        return Result(True)

    @staticmethod
    def check_is_link(path):
        """
        校验路径是否为软连接（路径可以不存在，只要是合法的路径字符串即可）。

        注意：
            即使路径不存在，但若合法，依然返回正确。
        """
        try:
            real_path = os.path.realpath(path)
            if real_path != os.path.normpath(path):
                return Result(result=False, err_msg="path is link")
            return Result(result=True)
        except Exception as e:
            return Result(result=False, err_msg=f"check link error: {e}")

    @staticmethod
    def check_is_link_exception(path):
        """如果是软连接则抛异常"""
        ret = FileCheck.check_is_link(path)
        if not ret:
            raise FileOperateError(ret.error)

    @staticmethod
    def check_input_path_valid(path, check_real_path=True) -> Result:
        if not path:
            return Result(result=False, err_msg="Input path is empty")

        if not isinstance(path, str):
            return Result(result=False, err_msg="Input path is not str type")

        if len(path) > 1024:
            return Result(result=False, err_msg="Input path length > 1024")

        ret = FileCheck._check_normal_file_path(path)
        if not ret:
            return ret

        if check_real_path and not FileCheck.check_is_link(path):
            return Result(result=False, err_msg="check path and realpath failed")

        return Result(result=True)

    @staticmethod
    def check_path_is_exist_and_valid(path, check_real_path=True) -> Result:
        """
        校验路径是否存在以及是否合法。
        """
        if not isinstance(path, str) or not os.path.exists(path):
            return Result(result=False, err_msg="path is not exists")

        return FileCheck.check_input_path_valid(path, check_real_path)

    @staticmethod
    def check_path_mode_owner_group(path, mode=None, user=None, group=None) -> Result:
        """
        校验路径权限，属主和属组（会先校验路径存在性及合法性）。
        """
        if mode and oct(os.stat(path).st_mode)[-3:] != mode:
            return Result(result=False, err_msg="path mode not right")

        invalid_owner_set = (user and not group) or (not user and group)
        if invalid_owner_set:
            return Result(result=False, err_msg="user and group must be specified at the same time")

        if user and group:
            if pwd.getpwuid(os.stat(path).st_uid).pw_name != user:
                return Result(result=False, err_msg="path user not right")
            if grp.getgrgid(os.stat(path).st_gid).gr_name != group:
                return Result(result=False, err_msg="path group not right")

        return Result(result=True)

    @staticmethod
    def check_system_key_path(file_path: str):
        system_key_path = ("/usr/bin", "/usr/sbin", "/etc", "/usr/lib", "/usr/lib64")
        check_path = os.path.realpath(file_path)
        return check_path not in system_key_path

    @staticmethod
    def check_xml_file_valid(xml_path):
        res = FileCheck.check_path_is_exist_and_valid(xml_path)
        if not res:
            return res

        return FileCheck._check_xml_is_safe(xml_path)

    @staticmethod
    def check_path_is_root(input_path, skip_no_exists=False) -> Result:
        """
        检查input_path的每一层路径是否都是root属主
        :param input_path: 输入文件路径
        :param skip_no_exists: 是否跳过不存在的文件路径，比如挂载时检查的路径必定时不存在的，需要跳过检查。默认值False，不跳过
        :return: Result检查结果
        """
        file_path = os.path.normpath(input_path)
        if not file_path.startswith(os.sep) or not os.path.abspath(file_path):
            return Result(result=False, err_msg="The path not a abspath path")

        temp_path = file_path
        while temp_path != os.sep:
            if skip_no_exists and not os.path.exists(temp_path):
                temp_path = os.path.dirname(temp_path)
                continue

            stat_info = os.stat(temp_path)
            dir_uid = stat_info.st_uid
            dir_gid = stat_info.st_gid
            if dir_uid != 0 or dir_gid != 0:
                return Result(result=False, err_msg="the path include no root path")

            path_mode = stat_info.st_mode
            if path_mode & stat.S_ISUID == stat.S_ISUID:
                return Result(result=False, err_msg="the path include uid path")

            if path_mode & stat.S_ISGID == stat.S_ISGID:
                return Result(result=False, err_msg="the path include gid path")

            other_mode = path_mode & stat.S_IRWXO
            if other_mode > 5:
                return Result(result=False, err_msg="the path mode not valid")

            temp_path = os.path.dirname(temp_path)

        return Result(result=True)

    @staticmethod
    def _check_xml_is_safe(xml_path):
        xml_dtd_key = "<!DOCTYPE"
        xml_max_size = 512 * 1024
        if os.path.getsize(xml_path) > xml_max_size:
            return Result(result=False, err_msg="This xml file is too large.")
        try:
            with open(xml_path, "r") as fw:
                content = fw.read()
                if xml_dtd_key in content:
                    return Result(result=False, err_msg="This is not a valid xml, it exist dtd content")
                return Result(True)
        except Exception as ex:
            return Result(result=False, err_msg=f"Read xml file failed, find exception: {ex}")

    @staticmethod
    def _check_normal_file_path(path):
        pattern_name = re.compile(r"[^0-9a-zA-Z_./-]")
        match_name = pattern_name.findall(path)
        if match_name or ".." in path:
            return Result(result=False, err_msg="there are illegal characters in path")
        return Result(result=True)


class FilePermission:
    @staticmethod
    def set_path_owner_group(path, user, recursive=False) -> Result:
        """
        设置路径的属主和数组。

        参数：
            recursive：如果路径是目录，可选择是否递归设置。
            ...

        说明：
            遵循函数功能单一设计原则，这里不做路径校验。用户使用前应调用FileCheck类进行路径校验。
        """
        try:
            user_uid = pwd.getpwnam(user).pw_uid
            user_gid = pwd.getpwnam(user).pw_gid
            if not recursive or os.path.isfile(path):
                os.lchown(path, user_uid, user_gid)
                return Result(result=True)

            for root, dirs, files in os.walk(path):
                os.lchown(root, user_uid, user_gid)
                for temp_dir in dirs:
                    os.lchown(os.path.join(root, temp_dir), user_uid, user_gid)
                for temp_file in files:
                    os.lchown(os.path.join(root, temp_file), user_uid, user_gid)

            return Result(result=True)
        except Exception as e:
            return Result(result=False, err_msg=f"set path owner and group error: {e}")

    @staticmethod
    def set_path_permission(path, mode, recursive=False, ignore_file=True, check_real_path=True):
        try:
            if not os.path.exists(path):
                return Result(result=False, err_msg=f"set path permission failed: path [{path}] not exist")

            ret = FileCheck.check_input_path_valid(path, check_real_path)
            if not ret:
                return Result(result=False,
                              err_msg=f"set path permission failed, the input path not valid: {ret.error}")

            os.chmod(path, mode)
            if not recursive or os.path.isfile(path):
                return Result(result=True)

            for root, dirs, files in os.walk(path):
                FilePermission._set_walk_dirs_mode(root, dirs, files, mode, ignore_file)

            return Result(result=True)
        except Exception as e:
            return Result(result=False, err_msg="set dir mode failed:{}".format(e))

    @staticmethod
    def _set_walk_dirs_mode(root, dirs, files, mode, ignore_file):
        if not ignore_file:
            for temp_file in files:
                os.chmod(os.path.join(root, temp_file), mode)

        for temp_dir in dirs:
            os.chmod(os.path.join(root, temp_dir), mode)

        return Result(result=True)


class FileConfusion:
    """文件内容混淆"""
    RANDOM_SIZE = 100
    BLOCK_SIZE = 100 * 1024
    MAX_SIZE = 5 * 1024 * 1024

    def __init__(self, filepath):
        self.filepath = filepath

    @staticmethod
    def confusion_path(path, force_flag=False):
        try:
            if not os.path.exists(path):
                return Result(result=True)

            if os.path.isdir(path):
                return FileConfusion._confusion_dir(path, force_flag)

            if FileConfusion._check_need_confusion(path, force_flag):
                return FileConfusion(path).start()

            return Result(result=True)
        except Exception as error:
            return Result(result=False, err_msg=f"confusion {path} error: {error}")

    @staticmethod
    def _check_need_confusion(path, force_flag=False):
        if force_flag:
            return True

        ext_tuple = (
            '.cert', '.crt', '.pem', '.cer', '.key', ".priv", '.psd', '.p12', '.pfx', '.ks', '.keystore', '.123', '.crl'
        )
        return path.endswith(ext_tuple)

    @staticmethod
    def _confusion_dir(path, force_flag=False):
        for parent, _, filenames in os.walk(path):
            for filename in filenames:
                if FileConfusion._check_need_confusion(filename, force_flag):
                    FileConfusion(os.path.join(parent, filename)).start()

        return Result(result=True)

    def start(self):
        try:
            file_size = os.stat(self.filepath).st_size
            file_size = min(file_size, self.MAX_SIZE)

            self._safe_overwrite_file(self.filepath, 0x00, file_size)
            self._safe_overwrite_file(self.filepath, 0xFF, file_size)
            self._safe_overwrite_file_by_random(self.filepath, file_size)
            return Result(result=True)
        except Exception as error:
            return Result(result=False, err_msg=f"confusion {self.filepath} error: {error}")

    def _safe_overwrite_file(self, path, data, size):
        data = bytearray([data] * self.BLOCK_SIZE)
        flags = os.O_WRONLY | os.O_CREAT | os.O_TRUNC
        modes = stat.S_IWUSR | stat.S_IRUSR
        with os.fdopen(os.open(path, flags, modes), "wb") as file:
            total_size = 0
            while total_size < size:
                write_size = min(size - total_size, self.BLOCK_SIZE)
                file.write(data[0:write_size])
                total_size += write_size

    def _safe_overwrite_file_by_random(self, path, size):
        with open('/dev/random', 'rb') as rand_file:
            random_bytes = bytearray(rand_file.read(self.RANDOM_SIZE))
        random_bytes = random_bytes * 1024

        flags = os.O_WRONLY | os.O_CREAT | os.O_TRUNC
        modes = stat.S_IWUSR | stat.S_IRUSR
        with os.fdopen(os.open(path, flags, modes), "wb") as file:
            total_size = 0
            while total_size < size:
                write_size = min(size - total_size, len(random_bytes))
                file.write(random_bytes[0:write_size])
                total_size += write_size


class FileCopy:

    @staticmethod
    def copy_file(src, dst, mode=None, user=None, group=None) -> Result:
        """
        拷贝文件，并设置文件的权限，属主和数组。

        参数：
            src：源文件。
            dst：目标文件。
            ...
        """
        if not os.path.isfile(src):
            return Result(result=False, err_msg=f"src not file")

        ret = FileCheck.check_path_is_exist_and_valid(src)
        if not ret:
            return ret

        ret = FileCheck.check_path_is_exist_and_valid(os.path.dirname(dst))
        if not ret:
            return ret

        ret = FileCheck.check_input_path_valid(dst)
        if not ret:
            return ret

        try:
            shutil.copyfile(src, dst)
        except Exception as error:
            return Result(result=False, err_msg=f"copy file error: {error}")

        if mode:
            ret = FilePermission.set_path_permission(dst, mode)
            if not ret:
                return ret

        invalid_owner_set = (user and not group) or (not user and group)
        if invalid_owner_set:
            return Result(result=False, err_msg="user and group must be specified at the same time")

        if user and group:
            ret = FilePermission.set_path_owner_group(dst, user)
            return ret

        return Result(result=True)

    @staticmethod
    def remove_path(path, check_real_path=True):
        try:
            if not os.path.exists(path):
                return Result(result=True)

            if not FileCheck.check_system_key_path(path):
                return Result(result=False, err_msg="path is system key path, can not remove")

            ret = FileCheck.check_input_path_valid(path, check_real_path)
            if not ret:
                return Result(result=False, err_msg=f"remove path failed, the path not valid: {ret.error}")

            if os.path.isdir(path):
                FileConfusion.confusion_path(path)
                shutil.rmtree(path)
                return Result(result=True)

            FileConfusion.confusion_path(path)
            os.remove(path)
            return Result(result=True)
        except Exception as error:
            return Result(result=False, err_msg=f"remove path failed: {error}")


class FileReader:
    def __init__(self, path, mode="r", encoding="utf-8"):
        self.path = path
        self.encoding = encoding
        self.mode = mode

    def read(self) -> Result:
        if self.mode == "rb":
            self.encoding = None
        try:
            with open(self.path, self.mode, encoding=self.encoding) as fd:
                return Result(True, data=fd.read())

        except Exception as error:
            err_msg = f"read {self.path} error {error}"
            return Result(result=False, err_msg=err_msg)

    def readline(self) -> Result:
        if self.mode == "rb":
            self.encoding = None
        try:
            with open(self.path, self.mode, encoding=self.encoding) as fd:
                return Result(True, data=fd.readline())
        except Exception as error:
            err_msg = f"read {self.path} error {error}"
            return Result(result=False, err_msg=err_msg)

    def readlines(self) -> Result:
        if self.mode == "rb":
            self.encoding = None
        try:
            with open(self.path, self.mode, encoding=self.encoding) as fd:
                data = fd.readlines()
                if len(data) > CommonConstants.MAX_ITER_LIMIT:
                    err_msg = f"File [{self.path}] has too many lines."
                    return Result(result=False, err_msg=err_msg)
                return Result(True, data=data)
        except Exception as error:
            err_msg = f"read {self.path} error {error}"
            return Result(result=False, err_msg=err_msg)


class FileWriter:
    def __init__(self, path, permission=0o700, mode="w", encoding="utf-8"):
        self.path = path
        self.encoding = encoding
        self.permission = permission
        self.mode = mode

    def write(self, data):
        ret = FileCheck.check_path_is_exist_and_valid(self.path)
        if not ret:
            return ret
        try:
            with open(self.path, mode=self.mode) as fd:
                fd.write(data)
        except Exception as error:
            return Result(result=False, err_msg=f"write to file failed: {error}")
        return Result(True)

    def append(self, data_list: list):
        ret = FileCheck.check_path_is_exist_and_valid(self.path)
        if not ret:
            return ret
        try:
            with open(self.path, mode=self.mode) as fd:
                fd.writelines(data_list)
        except Exception as error:
            return Result(result=False, err_msg=f"write to file failed: {error}")
        return Result(True)


class FileAttribute:

    @staticmethod
    def set_path_immutable_attr(path: str, set_flag: bool = False, recursive: bool = True) -> Result:
        if not isinstance(set_flag, bool):
            return Result(result=False, err_msg="Value must be bool.")

        if not FileCheck.check_path_is_exist_and_valid(path):
            Result(False, err_msg="File not exist or check failed")

        chattr_path = shutil.which("chattr")
        ret = FileCheck.check_path_is_root(chattr_path)
        if not ret:
            return ret

        operate = "+i" if set_flag else "-i"

        cmd = f"{chattr_path} {operate} {path}"
        if recursive:
            cmd = f"{chattr_path} {operate} -R {path}"
        exec_ret = ExecCmd.exec_cmd_get_output(cmd.split(" "))
        # 执行命令异常才返回False
        return Result(result=(exec_ret[0] != -1000), err_msg=exec_ret[1])
