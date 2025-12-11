#!/usr/bin/env python
# -*- coding:utf-8 -*-
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
功    能：Socket 客户端
修改记录：2017-01-02 创建
"""
import json
import os
import queue
import socket
import ssl
import struct
import threading
import time
from concurrent.futures import ThreadPoolExecutor

from common.file_utils import FileCheck, FilePermission
from common.log.logger import run_log
from common.kmc_lib.tlsconfig import TlsConfig
from common.kmc_lib.kmc import Kmc
from common.common_methods import CommonMethods


class Server:
    """
    功能描述：socket 服务端
    接口：NA
    修改记录：2017-01-02 创建
    """
    stopSocket = False
    socketServerThread = None
    host = "localhost"
    maxLink = 100
    dataLength = 1024
    maxDataLen = 1000 * 1024
    overStr = ";;over;"

    requestQueue = queue.Queue(maxsize=1024)
    maxRequestWorkers = 4

    @staticmethod
    def socket_set_so_linger(c_sock):
        l_onoff = 1
        l_linger = 60 * 2
        try:
            c_sock.setsockopt(socket.SOL_SOCKET, socket.SO_LINGER, struct.pack('ii', l_onoff, l_linger))
        except Exception as err:
            pass

    @staticmethod
    def start(socket_path, function):
        if os.path.exists(socket_path):
            os.unlink(socket_path)

        # 加载服务器所用证书和私钥
        cert_path = "/usr/local/mindx/MindXOM/software/ibma/cert/"
        cert = f"{cert_path}/server_kmc.cert"
        key = f"{cert_path}/server_kmc.priv"
        root_ca = f"{cert_path}/root_ca.cert"
        cert_primary_ksf = f"{cert_path}/om_cert.keystore"
        cert_standby_ksf = f"{cert_path}/om_cert_backup.keystore"
        alg_config_file = f"{cert_path}/om_alg.json"
        kmc = Kmc(cert_primary_ksf, cert_standby_ksf, alg_config_file)
        ret = FileCheck.check_path_is_exist_and_valid(f"{cert_path}/server_kmc.psd")
        if not ret:
            run_log.error(f"{cert_path}/server_kmc.psd invalid: {ret.error}")
            return
        with open(f"{cert_path}/server_kmc.psd", "r") as psd:
            encrypted = psd.read()
        decrypted = kmc.decrypt(encrypted)
        # 生成SSL上下文
        is_normal_value, context = TlsConfig.get_ssl_context(root_ca, cert, key, pwd=decrypted)
        if not is_normal_value:
            err_info = "get ssl context error."
            run_log.error(f'{err_info}')
            raise Exception(err_info)

        context.options &= ~ssl.OP_NO_TLSv1_2
        decrypted = None
        # Monitor与Redfish进程之间通过uds通信，Redfish客户端需要校验Monitor服务端证书
        context.verify_mode = ssl.CERT_REQUIRED
        # 监听端口
        with socket.socket(socket.AF_UNIX, socket.SOCK_STREAM) as sock:
            Server.socket_set_so_linger(sock)

            try:
                sock.setsockopt(socket.SOL_SOCKET, socket.SO_EXCLUSIVEADDRUSE, 1)
            except AttributeError:
                sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)

            sock.bind(socket_path)
            FilePermission.set_path_owner_group(socket_path, "MindXOM")
            sock.listen(Server.maxLink)

            Server._wait_client(context, sock, function)

    @staticmethod
    def receive_socket_message(sock):
        """
        功能描述：接收数据
        参数：sock Socket请求
        返回值：接收到的数据数据
        异常描述：最大长度超过限制后抛出异常
        修改记录：2023-07-15 限制消息最大长度
        """
        tmp_ret = []
        total_length = 0

        while True:
            data = sock.recv(Server.dataLength).decode()
            if len(data) < Server.dataLength:
                if data.endswith(Server.overStr):
                    # 截取掉最后面的结束符
                    data = data[: len(data) - len(Server.overStr)]

                if total_length + len(data) > Server.maxDataLen:
                    raise Exception("Received socket message is too long.")

                tmp_ret.append(data)
                break
            else:
                if total_length + len(data) > Server.maxDataLen:
                    raise Exception("Received socket message is too long.")

                tmp_ret.append(data)
                total_length += len(data)

        return "".join(tmp_ret)

    @staticmethod
    def deal_request(socket_request, function):
        """
        功能描述：Socket 服务端的回调函数，用于接收并处理数据
        参数：socket_request Socket请求
            function 回调函数
        返回值：调用底层接口获取数据并返回
        异常描述：NA
        修改记录：2017-01-02 创建
                2017-02-13 增加发送数据的判断，增加结束符
                2017-05-20 LibAdapter提供给Redfish的数据长度最大限制为38K.
                2017-06-26 socket close 之前先调用 shutdown.
        """
        try:
            data = Server.receive_socket_message(socket_request)

            # 调用回调函数，处理数据并返回结果
            ret = function(data)

            # 需要将结果转换为字符串
            if ret is None:
                ret = "None"
            elif isinstance(ret, dict):
                ret = json.dumps(ret)
            else:
                ret = str(ret)

            total_len = len(ret)

            # 将得到的结果返回上去
            if total_len % Server.dataLength == 0:
                ret = ret + Server.overStr
                total_len += len(Server.overStr)

            if total_len >= Server.maxDataLen:
                # 如果回复过长, 记录日志, 不发送数据.
                run_log.error("Response is too long(%d), do not send anything data." % total_len)
                return
            socket_request.sendall(ret.encode("utf-8"))
        except Exception as err:
            run_log.error("Deal request failed. Error msg: %s.", err)
            ret = CommonMethods.object_to_json(CommonMethods.ERROR, "Request data is invalid.")
            socket_request.sendall(str(ret).encode("utf-8"))
        finally:
            try:
                # 先 shutdown, 然后 close.
                socket_request.shutdown(socket.SHUT_RDWR)
            except Exception:
                pass
            try:
                socket_request.close()
            except Exception:
                run_log.error("Deal request socket close failed.")

    @staticmethod
    def worker(worker_id):
        while True:
            try:
                socket_request, function = Server.requestQueue.get()
                Server.deal_request(socket_request, function)
            except Exception:
                run_log.error("process request failed.")

    @staticmethod
    def init_server(socket_path, function):
        """
        功能描述：初始化函数
        参数：port 端口号
        返回值：NA
        异常描述：NA
        修改记录：2017-01-02 创建
        """
        Server.stopSocket = False

        if socket_path is None or not isinstance(socket_path, str):
            return [1, "socketPath is invalid."]

        if function is None:
            return [1, "Function is invalid."]

        try:
            Server.socketServerThread = threading.Thread(target=Server.start, args=(socket_path, function))
            Server.socketServerThread.start()
        except Exception as err:
            err_info = "Socket server start failed."
            run_log.error(f'{err_info} reason is {err}')
            return [1, err_info]

        # 等待 2 秒钟，然后检查是否还在运行
        time.sleep(2)

        if Server.socketServerThread.is_alive():
            return [0, "Socket server start successfully."]

        return [1, "Socket server start failed."]

    @staticmethod
    def _wait_client(context, sock, function):
        with context.wrap_socket(sock, server_side=True) as ssl_sock:
            executor = ThreadPoolExecutor(max_workers=Server.maxRequestWorkers)
            for i in range(Server.maxRequestWorkers):
                executor.submit(Server.worker, i)

            while not Server.stopSocket:
                # 接收客户端连接 接收到一个消息
                try:
                    client_socket, addr = ssl_sock.accept()
                except Exception as err:
                    run_log.warning("Socket accept failed: %r." % err)
                    continue

                Server.socket_set_so_linger(sock)

                # 交给worker线程完成数据的处理
                Server.requestQueue.put((client_socket, function))
