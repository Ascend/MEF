#!/usr/bin/python
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import gc
import os
import shlex
import threading
import time
import urllib.request
from http.client import HTTPResponse
from typing import Optional, Tuple, Union, NoReturn

from common.file_utils import FileCheck
from common.file_utils import FileUtils
from common.init_cmd import cmd_constants
from common.log.logger import run_log
from common.utils.app_common_method import AppCommonMethod
from common.utils.exec_cmd import ExecCmd
from common.utils.result_base import Result
from net_manager.manager.fd_cert_manager import FdCertManager
from net_manager.manager.net_cfg_manager import NetCfgManager

N_HTTP_SUCCESS = 0  # 成功
N_HTTP_URLISNONE = 102400  # URL为空
N_HTTP_USERPASSERROR = 102401  # 用户名密错误
N_HTTP_UNKNOWN = 102499  # 其他失败
N_HTTP_SOCKET_GETADDINFOERROR = 11001  # 获得连接失败

# 文件类
N_HTTP_MAKEDIRERROR = 102500  # Make download dir error.
N_HTTP_FILEOPENERROR = 102501  # Open file for write error.
N_HTTP_FILESIZEERROR = 102502  # File size less than 0.
N_HTTP_FILEMAKEERROR = 102503  # Make file error.
# Make file less than Online file error.
N_HTTP_FILELESSTHANONLINEERROR = 102504
# Make file biger than Online file error.
N_HTTP_FILEBIGERTHANONLINEERROR = 102505
N_HTTP_FILECOPYERROR = 102506  # 拷贝文件失败
N_HTTP_FILEGETCONTENTERROR = 102507  # 读取数据失败
N_HTTP_FILEWRITEERROR = 102508  # 写入文件异常
N_HTTP_USERCANCEL = 102600  # 用户取消

# 最大文件尺寸
MAX_FILE_SIZE = 512 * 1024 * 1024


def https_download_file(str_url, str_user, str_p_word, download_file, check_code=None):
    """
    :param download_file: 下载的文件
    :param check_code: 校验值
    :return: list
    """
    obj_downloader = UrlDownloader(str_url, str_user, str_p_word)
    ret = obj_downloader.download_file(download_file)
    if not ret[0]:
        run_log.error("Download file failed ret: %s errInfo:%s %d", ret[0], ret[1], ret[2])
        return [-1, "download file failed"]

    if check_code:
        valid_result = check_file_valid(download_file, check_code)
        if not valid_result[0]:
            error_msg = "Download file failed: {}.".format(valid_result[1])
            run_log.error(f'{error_msg}')
            return [-1, error_msg]

    run_log.info("Download file Success.")
    return [0, ""]


def check_file_valid(file_path, check_code):
    if not os.path.exists(file_path):
        return [False, "check file valid failed, file not exist"]

    cmd = "{} {} {} | sha256sum --check".format(cmd_constants.OS_CMD_ECHO, shlex.quote(check_code),
                                                shlex.quote(file_path))
    try:
        ret = ExecCmd.exec_cmd_use_pipe_symbol(cmd, wait=30)
    except Exception as err:
        run_log.error("Exec cmd failed: %s", err)
        return [False, "Exec cmd failed."]

    if ret[0] != 0:
        try:
            os.remove(file_path)
        except Exception as err:
            run_log.error("remove file failed: %s", err)

        return [False, "check code verify failed"]

    return [True, ""]


class UrlConnect(object):
    """
    功能描述：连接管理配置类
    """

    def __init__(self, str_url, str_user, str_p_word):
        self._str_url = str_url
        self._str_user = str_user
        self._str_pwd = str_p_word
        self._obj_response: Optional[HTTPResponse] = None

    @staticmethod
    def get_context() -> Result:
        net_manager = NetCfgManager().get_net_cfg_info()
        try:
            ctx = FdCertManager(net_manager.ip, net_manager.port).get_client_ssl_context()
        except Exception:
            return Result(False, err_msg="No available certificates found.")

        return Result(result=True, data=ctx)

    @staticmethod
    def _make_range_string(n_range_start, n_range_end):
        """
        功能描述：生成Range字符串
        """
        if not n_range_end and not n_range_start:
            str_range = "bytes=0-"
        elif not n_range_end:
            str_range = "bytes=%d-" % n_range_start
        else:
            str_range = "bytes=%d-%d" % (n_range_start, n_range_end)

        return str_range

    def connect(self, n_range_start=None, n_range_end=None) -> Tuple[bool, Union[str, HTTPResponse], int]:
        """
        功能描述：打开url 进行连接
        参数：nRangeStart 起始位置 nRangeEnd 结束位置
        """
        if self._str_url == "":
            return False, "URL is None.", N_HTTP_URLISNONE

        req = urllib.request.Request(self._str_url)
        # 处理下载开始位置
        str_range = self._make_range_string(n_range_start, n_range_end)
        req.add_header("Range", str_range)
        str_authentication = AppCommonMethod.make_authentication_string(self._str_user, self._str_pwd)
        req.add_header("Authorization", str_authentication)
        run_log.info("UrlConnect._OpenURL Open URL")
        ret = self.get_context()
        if not ret:
            run_log.error(ret.error)
            return False, ret.error, N_HTTP_UNKNOWN

        try:
            obj_response: HTTPResponse = urllib.request.urlopen(req, context=ret.data, timeout=30)
        except Exception:
            run_log.error("UrlConnect connect meet exception.")
            return False, "Connect meet exception", N_HTTP_UNKNOWN
        finally:
            gc.collect()

        # 连接成功对返回值进行处理
        try:
            dict_info = obj_response.info()
        except Exception:
            run_log.error("UrlConnect get response failed.")
            return False, "Connect meet exception", N_HTTP_UNKNOWN

        # 用户名密错误等
        if "download_error" in dict_info:
            run_log.error("UrlConnect._OpenURL Open URL error")
            obj_response.close()
            if isinstance(dict_info["download_error"], str) and \
                    dict_info["download_error"].startswith("please reinput username"):
                return False, "Username or Pword error.", N_HTTP_USERPASSERROR
            else:
                return False, "Open URL download_error", N_HTTP_UNKNOWN

        # 设置连接对象
        self._obj_response = obj_response
        return True, obj_response, N_HTTP_SUCCESS

    def is_connected(self):
        """
        功能描述：返回是否连接
        """
        if self._obj_response is None:
            return False

        if self._obj_response.fp is None:
            return False

        b_connect = (not self._obj_response.fp.closed)
        return b_connect

    def read(self, n_block_size):
        """
        功能描述：读文件
        参数：nBlockSize
        """
        return self._obj_response.read(n_block_size)

    def close(self):
        """
        功能描述：关闭该连接
        """
        try:
            self._obj_response.close()
        except Exception:
            return False
        finally:
            gc.collect()
        return True


class UrlDownloader(object):
    """
    功能描述：下载器接口类
    """

    def __init__(self, str_url, str_user, str_p_word, n_retry=0):
        self._n_retry = n_retry
        # 默认下载每块大小
        self.n_block_size = 1024 * 3
        # 对URL进行适当处理
        self._str_url = str_url.replace(" ", "%20")
        self._str_user = str_user
        self._str_pwd = str_p_word
        self._n_sum_size = -1
        self._n_download_size = 0
        self._obj_url_connect: Optional[UrlConnect] = None
        self._lock_progress = threading.Lock()
        self._lock_sum_size = threading.Lock()
        # 保证只启动一个下载的锁
        self._lock_download = threading.Lock()
        self._tuple_result: Tuple[bool, str, int] = (False, "", 0)

    def download_file(self, str_download_file):
        """
        功能描述：下载到指定文件中
        参数：strDownloadFile 下载的文件路径
        """
        # 保证不同时对一个文件进行下载
        self._lock_download.acquire()

        try:
            st_info = os.statvfs(os.path.dirname(str_download_file))
            rest_space = st_info.f_bavail * st_info.f_frsize
            file_size = self.get_file_size()
            rest_space = min(rest_space, MAX_FILE_SIZE)
            if file_size < 0:
                return False, "File size invalid.", N_HTTP_FILESIZEERROR
            elif file_size > rest_space:
                return False, "File size oversize.", N_HTTP_UNKNOWN

            if self.get_local_file_size(str_download_file) == self.get_file_size():
                return True, "file already down Success.", N_HTTP_SUCCESS

            n_range_start = self.get_down_file_seek(str_download_file)
            if self.get_local_file_size(str_download_file) < self.get_file_size():
                # 控制文件异常需要删除文件重新下载
                if n_range_start > self.get_local_file_size(str_download_file):
                    run_log.error("The seek file is illegal, download again.")
                    AppCommonMethod.force_remove_file(str_download_file)
                    n_range_start = 0
            else:
                # 本地已下载文件比预期下载文件大，删除重新下载
                run_log.error("The local file is illegal, download again.")
                n_range_start = 0
                for file in (str_download_file, f"{str_download_file}.seek"):
                    AppCommonMethod.force_remove_file(file)

            tuple_result = self._obj_url_connect.connect(n_range_start=n_range_start)
            if tuple_result[0] is False:
                return tuple_result
            return self.__make_local_file(str_download_file)
        except Exception:
            run_log.error("Download file meet exception.")
            return False, "Download meet exception", N_HTTP_UNKNOWN
        finally:
            self.dis_connect()
            self._lock_download.release()

    def dis_connect(self):
        """
        功能描述：断开连接
        """
        if self._obj_url_connect is None:
            return

        if self._obj_url_connect.is_connected():
            self._obj_url_connect.close()

    def get_file_name(self):
        """
        功能描述：获得文件名称
        """
        str_file_name = ""
        try:
            if "fileName=" in self._str_url:
                str_file_name = self._str_url.split("fileName=")[-1]
            else:
                str_file_name = self._str_url.split("/")[-1]
        except Exception:
            run_log.error("GetFileName meet Exception")
        return str_file_name

    def get_file_size(self):
        """
        功能描述：内部使用获得文件总大小
        """
        # 如果初始化过文件大小
        if self._n_sum_size >= 0:
            return self._n_sum_size

        try:
            with self._lock_sum_size:
                # 连接到Support获得文件大小
                self._connect_if_need(self._n_retry)
                if self._tuple_result[0] is False:
                    return self._n_sum_size

                if "content-length" in self._tuple_result[1].headers:
                    self._n_sum_size = int(self._tuple_result[1].headers["Content-Length"])
                self.dis_connect()
                return self._n_sum_size
        except Exception:
            run_log.error("Get File Size meet exception")
            return self._n_sum_size

    def get_down_file_seek(self, download_file):
        """
        功能描述：支持HTTPS断点续传，根据控制文件查询本地已下载文件大小
        :param download_file: 文件下载路径
        :return: 已下载文件的字节数
        """
        file_size = 0
        seek_file = download_file + ".seek"
        try:
            if FileCheck.check_path_is_exist_and_valid(seek_file):
                with open(seek_file) as file:
                    file_size = int(file.read().strip().split("/")[0])

            return file_size
        except Exception:
            run_log.error("Get DownFile Size meet exception")
            return 0

    def get_local_file_size(self, download_file):
        """
        功能描述：查询本地已下载文件的大小
        """
        file_size = 0
        if os.path.exists(download_file):
            file_size = os.path.getsize(download_file)

        return file_size

    def get_download_size(self):
        """
        功能描述：获得已经下载的大小
        """
        self._lock_progress.acquire()
        try:
            return self._n_download_size
        finally:
            self._lock_progress.release()

    def get_download_progress(self):
        """
        功能描述: 获取当前进度信息
        返回值: (int 总大小, int 已经下载大小)
        """
        return self._n_sum_size, self.get_download_size()

    def _set_download_size(self, n_download_size):
        """
        功能描述：设置下载文件大小
        参数：int nDownloadSize
        """
        self._lock_progress.acquire()
        try:
            self._n_download_size = n_download_size
        finally:
            self._lock_progress.release()

    def _connect_if_need(self, n_retry=None):
        """
        功能描述：需要的时候重新连接
        """
        if n_retry is None:
            n_retry = self._n_retry

        if self._obj_url_connect and self._obj_url_connect.is_connected():
            return True, "Success", N_HTTP_SUCCESS

        self._obj_url_connect = UrlConnect(self._str_url, self._str_user, self._str_pwd)
        n_connect = 0
        while n_connect <= n_retry:
            self._tuple_result = self._obj_url_connect.connect()
            # 打开成功或者用户名密错误直接返回,在加上Support代理不正确的情况
            if self._tuple_result[0] or self._tuple_result[2] in (N_HTTP_USERPASSERROR, N_HTTP_SOCKET_GETADDINFOERROR):
                return self._tuple_result

            # 如果服务器暂时不可用
            if self._tuple_result[2] == 503:
                # 强制回收
                gc.collect()
                # 20秒后重试
                time.sleep(20)
                n_connect -= 1

            # 失败后重新连接
            self._obj_url_connect.close()
            n_connect += 1

        return self._tuple_result

    def _make_dir_tree(self, str_local_path):
        """
        功能描述: 建立下载文件夹
        """
        str_local_dir = os.path.dirname(str_local_path)
        if not os.path.exists(str_local_dir):
            # 如果目录结构不存在则创建
            try:
                os.makedirs(str_local_dir, 0o750)
            except Exception:
                return False
        return True

    def __make_local_file(self, str_local_path):
        """
        功能描述: 根据打开的URL生成文件
        参数：strLocalPath 本地文件路径
        """

        # 如果没有文件夹则建立文件，如果失败直接返回
        if self._make_dir_tree(str_local_path) is False:
            return False, "Make download dir error.", N_HTTP_MAKEDIRERROR

        n_read = self.get_down_file_seek(str_local_path)
        total_size = self.get_file_size()
        if total_size > MAX_FILE_SIZE or total_size <= 0:
            return False, f"Download File size less than 0 or bigger than {MAX_FILE_SIZE}.", N_HTTP_FILESIZEERROR

        seek_file_path = f"{str_local_path}.seek"
        try:
            if os.path.exists(seek_file_path):
                os.remove(seek_file_path)
        except Exception:
            run_log.error("Open file for write meet exception")
            self.dis_connect()
            return False, "Open file for write error.", N_HTTP_FILEOPENERROR

        # 打开转存文件
        try:
            if os.path.exists(str_local_path):
                res = FileCheck.check_is_link(str_local_path)
                if not res:
                    raise ValueError(f"{res.error}")
                with open(str_local_path, "rb+") as file_target, \
                        os.fdopen(os.open(seek_file_path, os.O_WRONLY | os.O_CREAT, 0o600), "w") as seek_fp:
                    file_target.seek(n_read)
                    res = self.__write_seek(file_target, seek_fp, n_read)
            else:
                n_read = 0
                with os.fdopen(os.open(str_local_path, os.O_WRONLY | os.O_CREAT, 0o600), "wb") as file_target, \
                        os.fdopen(os.open(seek_file_path, os.O_WRONLY | os.O_CREAT, 0o600), "w") as seek_fp:
                    res = self.__write_seek(file_target, seek_fp, n_read)

        except Exception:
            run_log.error("Open file for write meet exception")
            return False, "Open file for write error.", N_HTTP_FILEOPENERROR
        finally:
            self.dis_connect()

        if not res[0]:
            return res

        # 判断是否成功
        if self.get_download_size() < self.get_file_size():
            self.__clear_download_file(str_local_path)
            return (False, f"retrieval incomplete: got only {self.get_download_size()} out of {total_size} bytes, "
                           f"please download again.",
                    N_HTTP_FILELESSTHANONLINEERROR)

        elif self.get_download_size() > self.get_file_size():
            self.__clear_download_file(str_local_path)
            self.__clear_download_file(str_local_path + ".seek")
            return (False, f"download file is larger than expected: got {self.get_download_size()} "
                           f"bigger than {total_size} bytes, please download again.",
                    N_HTTP_FILELESSTHANONLINEERROR)

        self.__clear_download_file(str_local_path + ".seek")
        return True, "Success", N_HTTP_SUCCESS

    def __clear_download_file(self, str_download_path) -> NoReturn:
        """
        功能描述: 清除下载内容
        """
        del_func = FileUtils.delete_full_dir if os.path.isdir(str_download_path) else FileUtils.delete_file_or_link
        try:
            del_func(str_download_path)
        except Exception as err:
            run_log.error("clear Upgrade %s failed, %s", str_download_path, err)

    def __write_seek(self, file_target, seek_fp, n_read):
        total_size = self.get_file_size()
        try:
            while True:
                b_block = self._obj_url_connect.read(self.n_block_size)
                # 读到空数据表示读完成
                if not b_block:
                    break

                # 下载的文件大于预期，下载失败
                if n_read + len(b_block) > total_size:
                    n_read += len(b_block)
                    self._set_download_size(n_read)
                    break

                file_target.write(b_block)
                file_target.flush()
                n_read += len(b_block)
                seek_fp.seek(0)
                seek_fp.write("{}/{}".format(n_read, total_size))
                seek_fp.flush()

                # 记录下载大小
                self._set_download_size(n_read)

                log_downloading_process(n_read, total_size, self.n_block_size)
            return [True, "Write file success"]
        except Exception:
            return [False, "Write file meet exception.", N_HTTP_FILEWRITEERROR]


def log_downloading_process(seek: int, total: int, block: int):
    need_log = False
    if block != 0:
        # 大约下载25M，打印一条进度日志
        need_log = not int(seek / block) % 8000

    if seek == block or total == seek or need_log:
        percentage = "{:.2%}".format(0)
        if total != 0:
            percentage = "{:.2%}".format(seek / total)
        run_log.info("downloading file %s", percentage)
