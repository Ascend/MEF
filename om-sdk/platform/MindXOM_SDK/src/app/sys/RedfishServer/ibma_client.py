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
import socket
import ssl
import struct

from common.constants.base_constants import CommonConstants
from common.file_utils import FileCheck
from common.log.logger import run_log
from common.kmc_lib.tlsconfig import TlsConfig
from common.kmc_lib.kmc import Kmc


def socket_set_so_linger(c_sock):
    l_onoff = 1
    l_linger = 60 * 2
    try:
        c_sock.setsockopt(socket.SOL_SOCKET, socket.SO_LINGER, struct.pack('ii', l_onoff, l_linger))
    except Exception:
        pass


class Client:
    """
    功能描述：socket 客户端
    接口：NA
    修改记录：2017-01-02 创建
    """
    dataLength = 1024
    maxDataLen = 1000 * 1024
    overStr = ";;over;"

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
            data = sock.recv(Client.dataLength).decode()
            if len(data) < Client.dataLength:
                if data.endswith(Client.overStr):
                    # 截取掉最后面的结束符
                    data = data[: len(data) - len(Client.overStr)]

                if total_length + len(data) > Client.maxDataLen:
                    raise Exception("Received socket message is too long.")

                tmp_ret.append(data)
                break
            else:
                if total_length + len(data) > Client.maxDataLen:
                    raise Exception("Received socket message is too long.")

                tmp_ret.append(data)
                total_length += len(data)

        return "".join(tmp_ret)

    @staticmethod
    def send_msg(msg, socket_path=CommonConstants.IBMA_SOCK_PATH):
        if not os.path.exists(os.path.realpath(socket_path)):
            run_log.error("Socket path is not exist")
            return [400, "Socket path is not exist."]

        cert_path = "/usr/local/mindx/MindXOM/software/RedfishServer/cert"
        cert_primary_ksf = f"{cert_path}/om_cert.keystore"
        cert_standby_ksf = f"{cert_path}/om_cert_backup.keystore"
        alg_config_file = f"{cert_path}/om_alg.json"
        kmc = Kmc(cert_primary_ksf, cert_standby_ksf, alg_config_file)
        ret = FileCheck.check_path_is_exist_and_valid(f"{cert_path}/client_kmc.psd")
        if not ret:
            run_log.error(f"{cert_path}/client_kmc.psd invalid: {ret.error}")
            return [400, f"{cert_path}/client_kmc.psd invalid: {ret.error}."]

        with open(f"{cert_path}/client_kmc.psd", "r") as psd:
            encrypted = psd.read()
        decrypted = kmc.decrypt(encrypted)
        # 生成SSL上下文
        cert_file = f"{cert_path}/client_kmc.cert"
        key_file = f"{cert_path}/client_kmc.priv"
        root_ca = f"{cert_path}/root_ca.cert"
        is_normal_value, context = TlsConfig.get_ssl_context(root_ca, cert_file, key_file, pwd=decrypted)
        if not is_normal_value:
            run_log.error(f"Send message failed. reason is get client ssl context error.")
            return [400, "Send message failed."]

        decrypted = None
        # Monitor与Redfish进程之间通过uds通信，Redfish客户端需要校验Monitor服务端证书
        context.verify_mode = ssl.CERT_REQUIRED

        # 与服务端建立socket连接
        try:
            sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
            ret = sock.connect(socket_path)
            if ret is not None:
                run_log.error("Connect server(%s) failed, error code: %s.", socket_path, ret)
                return [ret, "Connect server failed."]
            socket_set_so_linger(sock)

            with context.wrap_socket(sock) as ssl_sock:
                # 向服务端发送信息 发送数据请求
                if len(msg) % Client.dataLength == 0:
                    msg = msg + Client.overStr
                ssl_sock.sendall(msg.encode("utf-8"))
                # 接收服务端返回的信息
                try:
                    ret = Client.receive_socket_message(ssl_sock)
                except Exception as err:
                    run_log.error("Server response is invalid. Error msg: %s.", err)
                    return [400, "Server response is invalid."]
                try:
                    # 先 shutdown, 然后 close.
                    ssl_sock.shutdown(socket.SHUT_RDWR)
                except Exception:
                    pass

                return [0, ret]
        except Exception as err:
            run_log.error("Send message failed. %s ", err)
            return [400, "Send message failed."]
        finally:
            if sock is not None:
                try:
                    # 先 shutdown, 然后 close.
                    sock.shutdown(socket.SHUT_RDWR)
                except Exception:
                    pass
                try:
                    sock.close()
                except Exception:
                    run_log.error("Socket client close failed.")
