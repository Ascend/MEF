# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import itertools
import os
import re
import socket
import sys
from functools import partial
from pathlib import Path
from typing import List, NoReturn, Tuple, Dict, Iterator

from common.init_cmd import cmd_constants
from common.log.logger import run_log
from common.utils.exception_utils import OperateBaseError
from common.utils.exec_cmd import ExecCmd
from lib.Linux.systems.nic.models import NetConfig
from lib.Linux.systems.nic.nic_mgr import NetConfigMgr
from common.common_methods import CommonMethods
from common.yaml.yaml_methods import YamlMethod


class OperationRetCode(object):
    SUCCESS_OPERATION = 0
    FAILED_OPERATION = 1


class WebIpError(OperateBaseError):
    pass


class NginxConfig:
    """nginx配置类"""
    SYS_ETH_PATH = "/etc/sysconfig/network-scripts"
    UBUNTU_NET_CFG = "/etc/netplan/01-netcfg.yaml"
    ETH_LIST = ["eth0", "eth1"]

    @staticmethod
    def get_nginx_listen_ipv4() -> Tuple[str, str]:
        """
        从数据库中获取用途为web的ipv4地址与网卡
        :return: 网卡名与ipv4地址
        """
        for net_cfg in NetConfigMgr.query_info_with_condition(tag="web"):
            # 非DHCP方式返回静态IP
            if net_cfg.ipv4 != "dhcp":
                return net_cfg.name, net_cfg.ipv4

            # DHCP方式，动态获取IP
            cmd = f"{cmd_constants.OS_CMD_IP} addr show {net_cfg.name} | grep -w 'inet' | awk '{{print $2}}'"
            ret, ip_mask = ExecCmd.exec_cmd_use_pipe_symbol(cmd)
            if ret != 0:
                raise WebIpError(f"get <{net_cfg.name}> dhcp ip failed")

            return net_cfg.name, ip_mask.split("/")[0]

        raise WebIpError("get nic cfg from database failed")

    @staticmethod
    def _replace_listen_ip_string(web_ip_list: List[str], line: str) -> str:
        """
        将监听ip加入到符合正则要求的字符串中
        :param web_ip_list: 待监听的ip地址列表
        :param line: 待修改的字符串
        :return: 加入ip后的字符串
        """
        pattern = re.compile(r"listen.*443 ssl;")
        all_ip_config = ""
        for ip_address in web_ip_list:
            new_str = f"listen {ip_address}:443 ssl;"
            new_line = re.sub(pattern, new_str, line)
            all_ip_config += new_line
        return all_ip_config

    @staticmethod
    def _get_nginx_listen_ipv6(eth_name: str) -> str:
        """
        根据已监听的ipv4网卡获取该网卡的ipv6地址
        :param eth_name: 网卡名
        :return: ipv6地址
        """
        cmd = f"{cmd_constants.OS_CMD_IP} -6 addr show {eth_name} | {cmd_constants.OS_CMD_GREP} inet6"
        ret = ExecCmd.exec_cmd_use_pipe_symbol(cmd)
        if ret[0] != 0:
            run_log.warning("Get [%s] ipv6 addr info failed.", eth_name)
            return ""

        ret = re.findall(r"\binet6\b\s+([^\s]+)", ret[1])
        if not ret:
            run_log.warning("Get [%s] ipv6 addr info failed.", eth_name)
            return ""

        ipv6_addr = ret[0].split("/")[0].strip()
        # 检查ipv6地址是否合法
        try:
            socket.inet_pton(socket.AF_INET6, ipv6_addr)
        except Exception as err:
            run_log.warning("Check ipv6 %s failed: %s, not a valid ip", ipv6_addr, err)
            return ""

        # ipv6地址拼接网卡名
        eth_name = eth_name.split(":")[0]
        return f"[{ipv6_addr}%{eth_name}]"

    @staticmethod
    def _ubuntu_get_vlan_addr(data: Dict) -> List[str]:
        """
        获取vlan网卡的ip地址
        :param data: 网卡配置文件数据
        :return: ipv4地址列表
        """
        # 查询Vlan接口的ip地址
        ipv4_address = []
        if "vlans" not in data["network"]:
            return ipv4_address

        for vlan_name in data["network"]["vlans"]:
            vlan_addr = [addr.split("/")[0] for addr in data["network"]["vlans"][vlan_name]["addresses"]]
            ipv4_address.extend(addr.split("/")[0] for addr in vlan_addr)

        return ipv4_address

    @staticmethod
    def _ubuntu_get_normal_addr(data: Dict, eth_name: str) -> List[str]:
        """
        获取普通网卡的ip地址
        :param data: 网卡配置文件数据
        :param eth_name: 网卡名
        :return: ipv4地址列表
        """
        try:
            eth_cfg = data["network"]["ethernets"][eth_name]
        except Exception as err:
            run_log.error("key error: %s", err)
            return []

        if eth_cfg.get("dhcp4"):
            return ["dhcp"]

        return [addr.split("/")[0] for addr in eth_cfg.get("addresses", [])]

    @staticmethod
    def _update_old_tag_ini(eth_name: str, ipv4_addr: str) -> NoReturn:
        """
        更新tag.ini文件内容，写入到配置文件目录中
        :param eth_name:
        :param ipv4_addr:
        """
        tag_info = {"eth0": ["default", "default"], "eth1": ["default", "default"]}
        tag_info.update({eth_name: ["web"], f"ip_{eth_name}": [ipv4_addr]})
        ret = CommonMethods.write_net_tag_ini(str(tag_info))
        if not ret:
            raise WebIpError(f"Write tag.ini failed, {ret.error}")

    @staticmethod
    def _get_ipv4_addr(eth_sys_file: Path) -> Iterator[str]:
        """
        从系统网卡配置文件中获取ipv4地址
        :param eth_sys_file: 网卡配置文件路径
        :return: ip地址
        """
        with open(eth_sys_file, "r") as file:
            content = file.readlines()

        for line in content:
            if "dhcp" in line:
                yield "dhcp"
                return

        for line in content:
            if not line.startswith("IPADDR"):
                continue

            ip_address = line.split("=")[1].strip().strip('"')
            try:
                socket.inet_pton(socket.AF_INET, ip_address)
            except Exception as err:
                run_log.warning("Check ipv4 %s failed: %s, not a valid ip", ip_address, err)
                continue
            yield ip_address

    def set_nginx_config(self, nginx_file: str, operate_type: str) -> NoReturn:
        """
        获取nginx监听ip，写入到nginx配置文件中
        :param nginx_file: nginx配置文件路径
        """
        # 更新数据库
        self.update_database(operate_type)

        # 从数据库中获取监听IP
        all_listen_ip = []
        eth_name, ipv4_addr = self.get_nginx_listen_ipv4()
        all_listen_ip.append(ipv4_addr)

        ipv6_addr = self._get_nginx_listen_ipv6(eth_name)
        if ipv6_addr:
            all_listen_ip.append(ipv6_addr)

        self._write_nginx_config(all_listen_ip, nginx_file)

    def update_database(self, operate_type) -> NoReturn:
        """
        检测网卡的连接状态是否正常，如果多个网卡都处于连接状态，
        按照eth0->eth1的顺序，获取该网卡所有的ip地址，并将其中一个ipv4地址写入到数据库，
        最后更新tag.ini配置文件，兼容老版本
        """
        eth_li = self._get_link_state()
        for eth_name in eth_li:
            ipv4_li = self._get_ipv4_info_from_file(eth_name)
            if "dhcp" in ipv4_li:
                NetConfigMgr.delete_specific_eth_config(name=eth_name)
                NetConfigMgr.save_net_config([NetConfig(name=eth_name, ipv4="dhcp", tag="web")])
                self._update_old_tag_ini(eth_name, "dhcp")
                return

            # 如果数据库中的web用途ip已无效，就刷ip
            for net_cfg in NetConfigMgr.query_info_with_condition(name=eth_name):
                if net_cfg.ipv4 in ipv4_li and net_cfg.tag == "web":
                    return
                if net_cfg.ipv4 not in ipv4_li:
                    NetConfigMgr.delete_specific_eth_config(name=eth_name, ipv4=net_cfg.ipv4)

        eth_name = eth_li[0]
        ipv4_addr = self._get_ipv4_info_from_file(eth_name)[0]
        NetConfigMgr.delete_specific_eth_config(name=eth_name, ipv4=ipv4_addr)
        NetConfigMgr.save_net_config([NetConfig(name=eth_name, ipv4=ipv4_addr, tag="web")])
        self._update_old_tag_ini(eth_name, ipv4_addr)

    def _ubuntu_get_ipv4_addr(self, eth_name: str) -> List[str]:
        """
        获取ubuntu系统的指定网卡的所有ip地址
        :param eth_name: 网卡名
        :return: ip地址列表
        """
        try:
            data = YamlMethod.load_yaml_info(self.UBUNTU_NET_CFG)
        except Exception as err:
            run_log.error("load net config yaml failed %s", err)
            return []

        return self._ubuntu_get_normal_addr(data, eth_name) + self._ubuntu_get_vlan_addr(data)

    def _get_link_state(self):
        """
        获取连接状态为“link ok”的网卡列表
        :return: 网卡列表
        """
        link_eth_list = []
        for eth in self.ETH_LIST:
            cmd = f"{cmd_constants.OS_CMD_ETHTOOL} {eth} | {cmd_constants.OS_CMD_GREP} 'Link detected'"
            ret = ExecCmd.exec_cmd_use_pipe_symbol(cmd)
            if ret[0] != 0:
                run_log.warning("Get %s link state failed.", eth)
                continue
            if "yes" in ret[1]:
                link_eth_list.append(eth)

        if not link_eth_list:
            raise WebIpError("no nic in connected")

        return link_eth_list

    def _open_euler_get_all_ipv4_addr(self, eth_name: str) -> List[str]:
        """
        获取OpenEuler系统指定网卡的所有ip地址
        :param eth_name: 网卡名
        :return: ipv4地址列表
        """
        return self._euler_get_all_ipv4_addr(eth_name)

    def _get_ipv4_info_from_file(self, eth_name: str) -> List[str]:
        """
        获取当前系统的指定网卡的所有ip地址
        :param eth_name: 网卡名
        :return: ipv4地址列表
        """
        ipv4_info_handler_dic = {
            "EulerOS": partial(self._euler_get_all_ipv4_addr, eth_name),
            "Ubuntu": partial(self._ubuntu_get_ipv4_addr, eth_name),
            "openEuler": partial(self._open_euler_get_all_ipv4_addr, eth_name)
        }
        if ipv4_info_handler_dic.get(cmd_constants.OS_NAME) is None:
            raise WebIpError(f"OS [{cmd_constants.OS_NAME}] is not support")

        ipv4_li = ipv4_info_handler_dic.get(cmd_constants.OS_NAME)()
        if not ipv4_li:
            raise WebIpError("get ipv4 addr failed")

        return ipv4_li

    def _euler_get_all_ipv4_addr(self, eth_name: str) -> List[str]:
        """
        获取Euler系统的指定网卡的所有ip地址
        :param eth_name: 网卡名
        :return: ipv4地址列表
        """
        iterator = (self._get_ipv4_addr(path) for path in sorted(Path(self.SYS_ETH_PATH).glob(f"ifcfg-{eth_name}*")))
        return list(itertools.chain.from_iterable(iterator))

    def _write_nginx_config(self, web_ip_list: List[str], nginx_config: str) -> NoReturn:
        """
        将ip地址写入到nginx配置文件中
        :param web_ip_list: ip地址列表
        :param nginx_config: nginx配置文件路径
        """
        ip_add_flag = False
        file_data = ""
        with open(nginx_config, "r") as file:
            for line in file:
                if "listen" in line and not ip_add_flag:
                    file_data = "".join([file_data, self._replace_listen_ip_string(web_ip_list, line)])
                    ip_add_flag = True
                if "listen" in line:
                    continue
                file_data = "".join([file_data, line])

        with os.fdopen(os.open(nginx_config, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600), "w") as file:
            file.write(file_data)


def para_check(argv: List[str]) -> NoReturn:
    """
    校验参数是否合法
    :param argv: 命令行参数
    """
    if len(argv) != 3:
        raise WebIpError("Para num error.")

    if not os.path.exists(argv[1]):
        raise WebIpError("Nginx config file not exist.")

    if argv[2] not in ("Start", "Install"):
        raise WebIpError("Operate type error.")
    return argv[1], argv[2]


def main(argv: List[str]) -> int:
    """
    配置nginx监听ip的入口函数
    :param argv: 命令行参数
    :return: 成功 or 失败的返回码
    """
    try:
        nginx_file, operate_type = para_check(argv)
        NginxConfig().set_nginx_config(nginx_file, operate_type)
    except Exception as err:
        run_log.error("config web ip failed: %s", err)
        return OperationRetCode.FAILED_OPERATION

    return OperationRetCode.SUCCESS_OPERATION


if __name__ == '__main__':
    sys.exit(main(sys.argv))
