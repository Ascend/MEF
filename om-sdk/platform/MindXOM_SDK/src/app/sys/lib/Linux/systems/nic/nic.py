#!/usr/bin/python
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import copy
import json
import os
import re
import shlex
import shutil
import threading
import time
from abc import ABC, abstractmethod
from collections import defaultdict
from functools import partial
from itertools import islice
from pathlib import Path
from typing import Dict, NoReturn, Iterator, Tuple, List

from common.checkers import LocalIpChecker
from common.checkers.param_checker import EthernetInterfaceChecker
from common.constants.base_constants import CommonConstants
from common.init_cmd import cmd_constants
from common.log.logger import run_log
from common.utils.exception_utils import OperateBaseError
from common.utils.exec_cmd import ExecCmd
from lib.Linux.systems.nic.models import NetConfig
from lib.Linux.systems.nic.nic_mgr import NetConfigMgr
from common.common_methods import CommonMethods
from common.yaml.yaml_methods import YamlMethod
from monitor_db.session import session_maker


class NicOperateError(OperateBaseError):
    pass


class Nic:
    """处理网络信息的资源类"""
    # 资源文件
    PHYSIC_CLASS_PATH = "/sys/class/net"
    PHYSIC_INTERFACE_PATH = "/proc/net/dev"
    ETC_RESOLV_PATH = "/etc/resolv.conf"
    SPEED_FILE = "speed"
    ADDRESS_FILE = "address"

    # 网口连接的三种状态
    NIC_NO_LINK = "NoLink"
    NIC_LINK_UP = "LinkUp"
    NIC_LINK_DOWN = "LinkDown"

    # 网口通用信息相关命令
    CMD_IPV4_ADDR_SHOW = "ip -j -4 addr show"
    CMD_IPV6_ADDR_SHOW = "ip -j -6 addr show"
    CMD_IP_LINK_SHOW = "ip -s -j link show"
    CMD_IP_ROUTE = "ip -j route"
    CMD_IP_6_ROUTE = "ip -6 -j route"
    CMD_ETH_LINK = "ethtool {} | grep 'Link detected'"

    REGEX_STRING_MAX_LEN = 128

    nic_lock = threading.Lock()

    def __init__(self):
        # 网口名称
        self.Name = None
        # MAC 地址
        self.MACAddress = None
        # 网口速率，数值类型
        self.WorkMode = None
        # link状态：{"LinkUp", "LinkDown", "NoLink"}
        self.LinkStatus = None
        # 网口是否使能
        self.InterfaceEnabled = None
        # IPv4 地址列表
        self.IPv4Addresses = []
        # IPv6 地址列表
        self.IPv6Addresses = []
        # DNS
        self.NameServers = []
        # 网口类型
        self.AdapterType = None
        # 发包
        self.SendPackages = None
        # 收包
        self.RecvPackages = None
        # 错包
        self.ErrorPackages = None
        # 丢包
        self.DropPackages = None
        # 以太网接口集合
        self.items = None

    @staticmethod
    def _get_mask_for_num(mask_num: int) -> str:
        """
        功能描述：子网掩码转换，把 IP 地址后的数字解析成子网掩码
        :param mask_num: 子网掩码长度
        :return: 子网掩码
        """
        binary_mask = "1" * mask_num + "0" * (32 - mask_num)
        return ".".join(str(int(binary_mask[i:i + 8], 2)) for i in range(0, 32, 8))

    @staticmethod
    def _get_eth_name(name: str) -> str:
        """
        将GMAC替换为eth
        :param name: GMAC名称
        :return: 网卡名
        """
        return name.replace("GMAC", "eth")

    @staticmethod
    def _get_file_content(filepath: str) -> str:
        """
        读取指定路径的文件内容
        :param filepath: 文件路径
        :return: 文件内容
        """
        if not os.path.exists(filepath):
            return ""
        try:
            with open(filepath, "r") as file:
                return file.read().strip()
        except Exception as err:
            run_log.error("read %s failed, %s", filepath, err)
            return ""

    @staticmethod
    def _modify_network_config(request_data: Dict, eth_name: str) -> NoReturn:
        """
        根据请求参数，修改网络配置
        :param request_data: 请求参数
        :param eth_name: 网卡名
        """
        network_handler_dic = {
            "Ubuntu": partial(UbuntuNetworkHandler, request_data, eth_name),
            "EulerOS": partial(EulerNetworkHandler, request_data, eth_name),
            "openEuler": partial(OpenEulerNetHandler, request_data, eth_name)
        }
        if network_handler_dic.get(cmd_constants.OS_NAME) is None:
            raise NicOperateError(f"OS [{cmd_constants.OS_NAME}] is not support")

        network_handler_dic.get(cmd_constants.OS_NAME)().deal_ipv4_network()

    @staticmethod
    def _check_web_tag(tag_list, eth_li) -> NoReturn:
        """
        检查是否存在web用途的ip
        :param tag_list: 所有ip的tag列表
        :param eth_li: 网卡名列表
        """
        if "web" in tag_list:
            return

        for eth_name in eth_li:
            if list(NetConfigMgr.query_info_with_condition(name=eth_name, tag="web")):
                return

        raise NicOperateError("Find no web ip address")

    @staticmethod
    def _get_ip_cmd_json_info(cmd: str) -> Dict:
        """
        获取指定命令的输出结果，并转换为字典形式
        :param cmd: 执行的命令
        :return: dict形式的执行结果
        """
        ret = ExecCmd.exec_cmd_use_pipe_symbol(cmd)
        if ret[0] != 0:
            raise NicOperateError(f"bash <{cmd}> failed")

        return json.loads(ret[1])

    @staticmethod
    def _get_ipv4_addr_mode(dynamic: bool) -> str:
        """
        取得 IPv4 地址的模式
        :param dynamic: 是否为动态ip
        :return: DHCP 或者 Static
        """
        # 验证地址获取方式是否为 DHCP
        return "Static" if not dynamic else "DHCP"

    @staticmethod
    def _get_ipv6_addr_mode(scope: str, dynamic: bool) -> str:
        """
        取得 IPv6 地址模式
        :param scope: 地址获取方式的字符串
        :param dynamic: 是否为动态ip
        :return: LinkLocal、SLAAC、DHCPv6、Static其中一种
        """
        # 验证地址获取方式是否为 LinkLocal
        if "link" in scope:
            return "LinkLocal"

        # 验证地址获取方式是否为 SLAAC
        if "dynamic" in scope:
            return "SLAAC"

        # 验证地址获取方式是否为 DHCPv6
        return "Static" if not dynamic else "DHCPv6"

    def patch_request(self, request_data: Dict, gmac_name: str) -> List:
        """
        PATCH 请求的接口：根据请求参数修改网络配置
        :param request_data: 请求参数
        :param gmac_name: 网卡名称
        :return:
        """
        if self.nic_lock.locked():
            run_log.error("Net modify is busy")
            return [CommonMethods.ERROR, "The operation is busy."]

        with self.nic_lock:
            try:
                self._check_request_data(request_data, gmac_name)
                run_log.info("check request data %s success, start modify network", request_data)

                self._modify_network_config(request_data, self._get_eth_name(gmac_name))

                run_log.info("modify network success refresh information")
                self._get_dynamic_info(gmac_name)
            except Exception as err:
                run_log.error("deal with ethernet interface failed：%s", err)
                return [CommonMethods.ERROR, "deal with EthernetInterface config failed"]
        return [CommonMethods.OK, ""]

    def get_all_info(self, gmac_name: str = None) -> List:
        """
        GET 请求的接口
            1. 如果不传参数则返回网络资源集合
            2. 如果传参数则返回指定网卡的信息
        :param gmac_name: 网卡名
        :return:
        """
        gmacs, _ = self._get_eth_x_info()
        if gmac_name is None:
            self.items = gmacs
            return [CommonMethods.OK, ""]

        if gmac_name not in gmacs:
            run_log.error(r"%s not in gmacs list %s", gmac_name, gmacs)
            return [CommonMethods.NOT_EXIST, f"{gmac_name} not in net_name list"]

        try:
            self._get_dynamic_info(gmac_name)
        except Exception as err:
            run_log.error("get %s info failed, %s", gmac_name, err)
            return [CommonMethods.ERROR, f"get {gmac_name} info failed"]

        return [CommonMethods.OK, ""]

    def _get_gateway_by_dev(self, ip_route_cmd: str, dev: str) -> str:
        """
        获取当前网卡的网关ip地址
        :param ip_route_cmd: 获取网关的命令
        :param dev: 网卡名称
        :return: 网关ip
        """
        try:
            gw_info = self._get_ip_cmd_json_info(ip_route_cmd)
        except Exception as err:
            run_log.error("get gateway failed, %s", err)
            return ""

        for gw in gw_info:
            if "gateway" in gw and gw.get("dev") == dev:
                return gw.get("gateway")

        return ""

    def _get_ip_addr_show_info(self, ip_addr_cmd: str) -> Iterator[Dict]:
        """
        获取指定网卡的ip信息
        :param ip_addr_cmd: ip addr show命令
        :return: 执行结果的字典形式的迭代器
        """
        try:
            ip_info = self._get_ip_cmd_json_info(ip_addr_cmd)
        except Exception as err:
            run_log.error("get ip addr info failed, %s", err)
            return

        for ip in ip_info:
            yield from ip.get("addr_info")

    def _get_eth_x_info(self) -> Tuple[List[str], List[str]]:
        """
        获取所有网口资源列表
        :return: 当前设备上真实的网卡列表 + 当前设备上的网卡列表
        """
        eth_list = []
        eth_list_pure = set()
        for line in self._get_file_content(self.PHYSIC_INTERFACE_PATH).split("\n"):
            face = line.split(":", 1)[0].strip()
            if not face.startswith("eth"):
                continue
            eth_list.append(face)  # 如果配置vlan eth0 eth0.1 eth1
            eth_list_pure.add(face.split(".")[0])  # eth0 eth1

        gmacs = [i.replace("eth", "GMAC") for i in eth_list_pure]  # GMAC0 GMAC1
        return gmacs, eth_list

    def _check_request_data(self, request_data: Dict, gmac_name: str) -> NoReturn:
        """
        校验请求参数是否合法
        :param request_data: 请求参数
        :param gmac_name: 网卡名
        """
        # 检查参数格式
        check_ret = EthernetInterfaceChecker().check(request_data)
        if not check_ret:
            raise NicOperateError(f"check parameter failed, {check_ret.reason}")

        gmac_li, eth_li = self._get_eth_x_info()
        if gmac_name not in gmac_li:
            raise NicOperateError(f"System not exits {gmac_name} resource.")

        # 获取其他网卡上的所有ip
        other_eth_ip_li = []
        other_eth_li = []
        for eth in eth_li:
            if eth.startswith(self._get_eth_name(gmac_name)):
                continue
            other_eth_li.append(eth)
            for ipv4_dict in self._get_ipv4_addresses_info(eth):
                other_eth_ip_li.append(ipv4_dict.get("Address"))

        # 获取ip、网关、tag
        addr_set = set()
        gateway_set = set()
        tag_li = []
        ipv4_dic_li = request_data.get("IPv4Addresses")
        for ipv4_dict in ipv4_dic_li:
            addr_set.add(ipv4_dict.get("Address"))
            tag_li.append(ipv4_dict.get("Tag", ""))
            if ipv4_dict.get("Gateway"):
                gateway_set.add(ipv4_dict.get("Gateway"))

        # 检查网关是否唯一
        if len(gateway_set) > 1:
            raise NicOperateError("ensure that the default gateway is globally unique.")

        # 检查IP重复
        if len(addr_set) != len(ipv4_dic_li):
            raise NicOperateError("ip repeated")

        # 校验ip是否在其他网口也配置过
        if addr_set.intersection(other_eth_ip_li):
            raise NicOperateError("ip has existed")

        # 所有网卡中必须有一个web用途的网卡
        self._check_web_tag(tag_li, other_eth_li)

    def _get_dynamic_info(self, gmac_name: str) -> NoReturn:
        """
        通过网口名称，获取网卡的动态信息
        :param gmac_name: 网卡名称
        """
        self.Name = self._get_eth_name(gmac_name)
        self.AdapterType = "GMAC"

        # 获取 MAC 地址
        self.MACAddress = self._get_file_content(os.path.join(self.PHYSIC_CLASS_PATH, self.Name, self.ADDRESS_FILE))

        # 获取接口连接状态
        # 如果接口连接状态为 LinkUp， 则获取网口速率
        self.LinkStatus = self._get_link_status()
        if self.LinkStatus == self.NIC_LINK_UP:
            speed_content = self._get_file_content(os.path.join(self.PHYSIC_CLASS_PATH, self.Name, self.SPEED_FILE))
            self.WorkMode = f"{speed_content}Mb/s"

        # 获取 IPv4、IPv6 地址信息
        _, eth_li = self._get_eth_x_info()
        for eth_name in eth_li:
            if not eth_name.startswith(self.Name):
                continue

            # vlan网卡与普通网卡的配置
            self.IPv4Addresses.extend(self._get_ipv4_addresses_info(eth_name))
            self.IPv6Addresses.extend(self._get_ipv6_addresses_info(eth_name))

        # 获取 网卡流量
        self._get_net_traffic_info(f"{self.CMD_IP_LINK_SHOW} {self.Name}")
        self._get_name_server()

    def _get_ipv4_addresses_info(self, eth_name: str) -> Iterator[Dict[str, str]]:
        """
        获取网卡的ipv4地址信息
        :param eth_name: 网卡名称
        :return: ipv4地址信息的迭代器
        """
        # 获取网关
        gateway_v4 = self._get_gateway_by_dev(self.CMD_IP_ROUTE, eth_name)
        # 提取 IPv4地址信息与掩码
        for addr in self._get_ip_addr_show_info(f"{self.CMD_IPV4_ADDR_SHOW} {eth_name}"):
            ipv4_address = {
                "Address": addr.get("local"),
                "SubnetMask": self._get_mask_for_num(addr.get("prefixlen")),
                "AddressOrigin": self._get_ipv4_addr_mode(addr.get("dynamic")),
                "Gateway": gateway_v4,
                "Tag": NetConfigMgr.query_tag_from_ip(addr.get("local"), addr.get("dynamic")),
                "VlanId": int(eth_name.split(".")[1]) if "." in eth_name else None
            }
            yield ipv4_address

    def _get_ipv6_addresses_info(self, eth_name: str) -> Iterator[Dict[str, str]]:
        """
        获取网卡的ipv6地址信息
        :param eth_name: 网卡名称
        :return: ipv6地址信息的迭代器
        """
        # 获取网关
        gateway_v6 = self._get_gateway_by_dev(self.CMD_IP_6_ROUTE, eth_name)
        # 提取IPv6地址信息、掩码，ipv6的scope，ipv6的地址状态
        for addr in self._get_ip_addr_show_info(f"{self.CMD_IPV6_ADDR_SHOW} {eth_name}"):
            ipv6_address = {
                "Address": addr.get("local"),
                "PrefixLength": addr.get("prefixlen"),
                "AddressOrigin": self._get_ipv6_addr_mode(addr.get("scope"), addr.get("dynamic")),
                "AddressState": "Preferred",
                "Gateway": gateway_v6
            }
            yield ipv6_address

    def _check_eth_link_state(self, link_cmd: str) -> bool:
        """
        检查网卡的链接状态
        :param link_cmd: 获取连接状态的命令
        :return: True or False
        """
        try:
            link_info = self._get_ip_cmd_json_info(link_cmd)
        except Exception as err:
            run_log.error("get link state failed, %s", err)
            return False

        for link in link_info:
            if "UP" in link.get("flags"):
                return True

        return False

    def _get_link_status(self) -> str:
        """
        获取连接状，三种连接状态处理逻辑：
            1、LinkUp： 软连接 True，硬连接 True
            2、NoLink： 软连接 True，硬连接 False
            3、LinkDown：软连接 False
        """
        # 获取软连接状态
        self.InterfaceEnabled = False
        if not self._check_eth_link_state(f"{self.CMD_IP_LINK_SHOW} {shlex.quote(self.Name)}"):
            return self.NIC_LINK_DOWN

        # 获取硬连接状态
        self.InterfaceEnabled = True
        ret = ExecCmd.exec_cmd_use_pipe_symbol(self.CMD_ETH_LINK.format(self.Name))
        if ret[0] == 0 and "yes" in ret[1]:
            return self.NIC_LINK_UP

        return self.NIC_NO_LINK

    def _get_name_server(self) -> NoReturn:
        """获取DNS地址"""
        for info in self._get_file_content(self.ETC_RESOLV_PATH).split("\n"):
            if len(info) > self.REGEX_STRING_MAX_LEN:
                continue
            name_server = re.search(r"\s*nameserver\s+([\d.]+)", info)
            if name_server and name_server.group(1) not in ("0.0.0.0", "255.255.255.255"):
                self.NameServers.append(name_server.group(1))

    def _get_net_traffic_info(self, traffic_cmd: str) -> NoReturn:
        """
        获取网口IO流量信息
        :param traffic_cmd: 获取流量信息的命令
        """
        try:
            traffic_info = self._get_ip_cmd_json_info(traffic_cmd)
        except Exception as err:
            run_log.error("get traffic info failed, %s", err)
            return

        for traffic in traffic_info:
            self.RecvPackages = traffic.get("stats64", {}).get("rx", {}).get("packets", "")
            self.ErrorPackages = traffic.get("stats64", {}).get("rx", {}).get("errors", "")
            self.DropPackages = traffic.get("stats64", {}).get("rx", {}).get("dropped", "")
            self.SendPackages = traffic.get("stats64", {}).get("tx", {}).get("packets", "")
            return


class CommNetworkHandler(ABC):
    """修改网络配置的公共处理类"""
    IP_ROUTE_CMD = "ip -j route"
    PHYSIC_INTERFACE_PATH = "/proc/net/dev"
    NETWORK_CHECK_TIMES = 5

    def __init__(self, request_data, eth_name):
        """
        构造函数
        :param request_data: 处理参数
        :param eth_name: 网卡名称
        """
        self.eth_name = eth_name
        self.ip_info_items = request_data.get("IPv4Addresses")

    @staticmethod
    def _restart_nginx() -> NoReturn:
        """重启Nginx服务"""
        if ExecCmd.exec_cmd((cmd_constants.OS_CMD_SYSTEMCTL, "restart", "start-nginx")) != 0:
            run_log.error("systemctl restart start-nginx failed")

    @staticmethod
    def _convert_mask_num(subnet_mask: str) -> int:
        """
        把子网掩码解析成数字表示
        :param subnet_mask: 子网掩码
        :return: 子网掩码数字表示
        """
        return "".join(bin(int(x))[2:] for x in subnet_mask.split(".")).count("1")

    @staticmethod
    def _check_ip_conflict(cmd: Tuple) -> NoReturn:
        """
        检查ip是否存在冲突
        :param cmd: 执行ip冲突检测的命令
        """
        # 如果命令执行失败，那么说明ip不存在冲突
        if ExecCmd.exec_cmd(cmd, wait=5) != 0:
            return
        # 如果命令执行成功，那么说明ip存在冲突，直接报错
        raise NicOperateError(f"check ip conflict failed, bash <{cmd}>")

    @abstractmethod
    def rollback_ipv4(self) -> NoReturn:
        """网络配置失败后，进行回滚的接口"""
        pass

    @abstractmethod
    def effect_network(self) -> NoReturn:
        """生效网络配置的接口"""
        pass

    @abstractmethod
    def modify_network(self) -> NoReturn:
        """修改网络配置的接口"""
        pass

    def deal_ipv4_network(self) -> NoReturn:
        """处理ipv4网络配置的接口"""
        try:
            self._deal_with_ipv4()
        except Exception as err:
            run_log.error("deal ipv4 network failed: %s, begin to rollback", err)
            self.rollback_ipv4()
            raise NicOperateError(f"deal ipv4 network failed: {err}") from err

    def _deal_with_ipv4(self) -> NoReturn:
        """处理ipv4网络配置"""
        # 修改网络配置
        self.modify_network()

        # 生效网络配置
        self.effect_network()

        # ip冲突、连通性检测
        self._network_check()

        # 更新数据库与tag.ini文件
        self._update_database_and_tag_ini()

        # 重启nginx
        threading.Timer(10, self._restart_nginx).start()

    def _update_old_tag_ini(self, new_tag_list: List[str], new_ip_list: List[str]) -> NoReturn:
        """更新tag.ini兼容老版本"""
        ret = CommonMethods.load_net_tag_ini()
        if not ret:
            raise NicOperateError(f"load tag.ini failed, {ret.error}")

        # 写入新的tag内容到tag.ini
        tag_info = ret.data
        tag_info[self.eth_name] = new_tag_list
        tag_info[f"ip_{self.eth_name}"] = new_ip_list

        ret = CommonMethods.write_net_tag_ini(str(tag_info))
        if not ret:
            raise NicOperateError(f"write tag.ini failed, {ret.error}")

    def _get_cur_vlan_collection(self) -> Iterator[str]:
        """
        获取当前所有配置vlan的网卡名
        :return: vlan网卡名
        """
        with open(self.PHYSIC_INTERFACE_PATH) as file:
            lines = file.readlines()

        for line in lines:
            if self.eth_name not in line:
                continue
            interface = line.split(":", 1)[0].strip()
            if "." in interface:
                yield interface

    def _get_arping_cmd(self, addr_ipv4: str) -> Tuple:
        """
        获取执行ip冲突检测的命令
        :param addr_ipv4: ipv4地址
        :return: ip冲突检测的命令
        """
        local_ip = LocalIpChecker().get_ip_address(self.eth_name.encode())
        arping_cmd = (cmd_constants.OS_CMD_ARPING, "-f", "-c", "1", "-I", self.eth_name, "-s", local_ip, addr_ipv4)
        return arping_cmd

    def _get_current_gateway(self) -> str:
        """
        获取当前的网关地址
        :return: 网关地址
        """
        ret = ExecCmd.exec_cmd_use_pipe_symbol(self.IP_ROUTE_CMD, wait=10)
        if ret[0] != 0:
            raise NicOperateError(f"get gateway failed, bash <{self.IP_ROUTE_CMD}>")

        try:
            for gw in json.loads(ret[1]):
                return gw.get("gateway", "")
        except Exception as err:
            raise NicOperateError(f"jsonfy gateway info failed, {err}") from err

    def _check_ip_connection(self, cmd: Tuple) -> NoReturn:
        """
        IP连通性测试
        :param cmd: ip连通性测试的命令
        """
        for _ in range(self.NETWORK_CHECK_TIMES):
            # 如果命令执行成功，那么说明网络正常
            if ExecCmd.exec_cmd(cmd, wait=5) == 0:
                return
            # 如果命令执行失败，重新检测
            time.sleep(0.2)

        raise NicOperateError(f"check ip connection failed, bash <{cmd}>")

    def _network_check(self) -> NoReturn:
        """网络状况检测"""
        gateway = self._get_current_gateway()
        # 网关连通性测试
        can_connect_to_gateway = False
        for local_ip in LocalIpChecker().get_local_ips():
            gateway_ping_cmd = (cmd_constants.OS_CMD_PING, "-I", local_ip, "-c", "1", gateway, "-w", "2")
            try:
                self._check_ip_connection(gateway_ping_cmd)
            except NicOperateError as error:
                run_log.warning("check gateway connection failed, err: %s", error)
                continue

            can_connect_to_gateway = True
            break

        if not can_connect_to_gateway:
            raise NicOperateError(f"check gateway connection failed, cannot connect to gateway: {gateway}")

        for ip_addr_item in self.ip_info_items:
            connect_test = ip_addr_item.get("ConnectTest")
            remote_test_ip = ip_addr_item.get("RemoteTestIp")
            addr_ipv4 = ip_addr_item.get("Address")
            # IP冲突检测
            self._check_ip_conflict(self._get_arping_cmd(addr_ipv4))

            # 远程ip连通性测试
            if connect_test and remote_test_ip:
                remote_ip_ping_cmd = (cmd_constants.OS_CMD_PING, "-I", addr_ipv4, "-c", "1", remote_test_ip, "-w", "2")
                self._check_ip_connection(remote_ip_ping_cmd)

    def _update_database_and_tag_ini(self) -> NoReturn:
        """生效网络配置后，更新数据库"""
        # 先删除旧的网卡ip配置，添加新的网卡ip配置
        eth_cfgs = [
            NetConfig(name=self.eth_name, ipv4=info.get("Address"), tag=info.get("Tag"))
            for info in self.ip_info_items
        ]
        new_tag_list = [info.get("Tag") for info in self.ip_info_items]
        new_ip_list = [info.get("Address") for info in self.ip_info_items]
        with session_maker() as session:
            # 显示开启事务，保证同个session下的多个db的操作都能一起回滚
            with session.begin():
                session.query(NetConfig).filter_by(name=self.eth_name).delete()
                session.bulk_save_objects(eth_cfgs)
                self._update_old_tag_ini(new_tag_list, new_ip_list)


class EulerNetworkHandler(CommNetworkHandler):
    """修改网络配置的Euler系统处理类"""
    ETH_CONFIG_PATH = "/etc/sysconfig/network-scripts"
    NETWORK_FILE = "/etc/sysconfig/network"
    NETWORK_FILE_BAK = "/etc/sysconfig/network.bak"

    def __init__(self, request_data, eth_name):
        """
        构造函数
        :param request_data: 处理参数
        :param eth_name: 网卡名称
        """
        super().__init__(request_data, eth_name)
        self.back_ifcfg_eth_files = []
        self.modify_ifcfg_eth_files = []

    def modify_network(self) -> NoReturn:
        """修改网络配置"""
        # 先备份网卡配置文件
        self._backup_netcfg_file()

        # 修改网卡配置文件
        self._write_ipv4_netcfg()

    def effect_network(self) -> NoReturn:
        """生效网络配置"""
        # 删除当前的vlan链接
        for vlan in self._get_cur_vlan_collection():
            if ExecCmd.exec_cmd((cmd_constants.OS_CMD_IP, "link", "delete", vlan)) != 0:
                raise NicOperateError(f"delete vlan link <{vlan}> failed")

        # 清空当前网卡ip
        if ExecCmd.exec_cmd((cmd_constants.OS_CMD_IP, "addr", "flush", self.eth_name)) != 0:
            raise NicOperateError("flush ip failed")

        # 生效网络配置
        if ExecCmd.exec_cmd((cmd_constants.OS_CMD_SYSTEMCTL, "restart", "network")) != 0:
            raise NicOperateError("restart network failed")
        time.sleep(2)

    def rollback_ipv4(self) -> NoReturn:
        """回滚网络配置"""
        # 回滚网关
        shutil.copyfile(self.NETWORK_FILE_BAK, self.NETWORK_FILE)

        # 网卡文件
        for path in self.modify_ifcfg_eth_files:
            os.remove(path)

        # 删除当前的vlan链接
        for vlan in self._get_cur_vlan_collection():
            ExecCmd.exec_cmd((cmd_constants.OS_CMD_IP, "link", "delete", vlan))

        for path in self.back_ifcfg_eth_files:
            shutil.copyfile(path, Path(self.ETH_CONFIG_PATH).joinpath(path.name.replace("back_", "")))

        # 生效网络配置
        ExecCmd.exec_cmd((cmd_constants.OS_CMD_SYSTEMCTL, "restart", "network"))
        time.sleep(2)

    def _backup_netcfg_file(self) -> NoReturn:
        """备份网卡配置文件"""
        # 网卡文件
        for path in islice(Path(self.ETH_CONFIG_PATH).glob(f"ifcfg-{self.eth_name}*"), CommonConstants.MAX_ITER_LIMIT):
            back_file = Path(self.ETH_CONFIG_PATH).joinpath(f"back_{path.name}")
            shutil.copyfile(path, back_file)
            self.back_ifcfg_eth_files.append(back_file)

        shutil.copyfile(self.NETWORK_FILE, self.NETWORK_FILE_BAK)

    def _write_ipv4_netcfg(self) -> NoReturn:
        """修改网卡配置文件"""
        eth_id = 0
        gateway = ""
        for ip_dict in self.ip_info_items:
            gateway = ip_dict.get("Gateway", "") or gateway
            # 获取网卡名称
            if ip_dict.get("VlanId") is not None:
                dev_name = f"{self.eth_name}.{ip_dict.get('VlanId')}"
            else:
                dev_name = self.eth_name if eth_id == 0 else f"{self.eth_name}:{eth_id}"
                eth_id += 1

            self._write_ipv4_file(ip_dict, dev_name)
            self.modify_ifcfg_eth_files.append(Path(self.ETH_CONFIG_PATH, f"ifcfg-{dev_name}"))

        if gateway:
            with os.fdopen(os.open(self.NETWORK_FILE, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o644), "w") as f_in:
                f_in.write(f"GATEWAY={gateway}\n")

        # 清除该网卡的其余配置文件
        for path in islice(Path(self.ETH_CONFIG_PATH).glob(f"ifcfg-{self.eth_name}*"), CommonConstants.MAX_ITER_LIMIT):
            if path not in self.modify_ifcfg_eth_files:
                os.remove(path)

    def _write_ipv4_file(self, ip_dict: Dict, dev_name: str):
        """
        修改网卡配置文件
        :param ip_dict: ip信息的字典
        :param dev_name: 网卡名称
        :return:
        """
        ip_address = ip_dict.get("Address")
        ip_info_str = 'DEVICE="{0}"\nBOOTPROTO="static"\nONBOOT="yes"\n' \
                      'IPADDR="{1}"\nNETMASK="{2}"\n'.format(dev_name, ip_address, ip_dict["SubnetMask"])
        if "." in dev_name:
            ip_info_str = f'{ip_info_str}\nVLAN="yes"'

        ifcfg_eth_file = os.path.join(self.ETH_CONFIG_PATH, f"ifcfg-{dev_name}")
        with os.fdopen(os.open(ifcfg_eth_file, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o644), "w") as file:
            file.write(ip_info_str)


class UbuntuNetworkHandler(CommNetworkHandler):
    """修改网络配置的Ubuntu系统处理类"""

    # 用于保存本次修改的vlan接口，回滚时使用
    BAK_NETCFG_YAML = "/etc/netplan/01-netcfg.yaml.back"
    NETCFG_YAML = "/etc/netplan/01-netcfg.yaml"
    NETPLAN_CMD = shutil.which("netplan")

    def modify_network(self) -> NoReturn:
        """修改网络配置"""
        # 拷贝文件，用于恢复和备份
        shutil.copyfile(self.NETCFG_YAML, self.BAK_NETCFG_YAML)

        # 加载yaml
        try:
            data = YamlMethod.load_yaml_info(self.NETCFG_YAML)
        except Exception as err:
            raise NicOperateError(f"load {self.NETCFG_YAML} failed: {err}") from err

        # 修改修改配置数据
        self._modify_netplan_data(data)

        # 重写yaml文件
        try:
            YamlMethod.dumps_yaml_file(data, self.NETCFG_YAML, 0o644)
        except Exception as err:
            raise NicOperateError(f"write net config yaml failed: {err}") from err

    def effect_network(self) -> NoReturn:
        """生效网络配置"""
        # 删除当前的vlan链接
        for vlan in self._get_cur_vlan_collection():
            if ExecCmd.exec_cmd((cmd_constants.OS_CMD_IP, "link", "delete", vlan)) != 0:
                raise NicOperateError(f"delete vlan link <{vlan}> failed")

        # 生效网络配置
        if ExecCmd.exec_cmd((self.NETPLAN_CMD, "apply")) != 0:
            raise NicOperateError("netplan apply failed")
        time.sleep(2)

    def rollback_ipv4(self) -> NoReturn:
        """回滚网络配置"""
        # 删除当前的vlan链接
        for vlan in self._get_cur_vlan_collection():
            ExecCmd.exec_cmd((cmd_constants.OS_CMD_IP, "link", "delete", vlan))

        # 回退配置
        shutil.copyfile(self.BAK_NETCFG_YAML, self.NETCFG_YAML)

        # 生效网络配置
        ExecCmd.exec_cmd((self.NETPLAN_CMD, "apply"))
        time.sleep(2)

    def _get_arping_cmd(self, addr_ipv4: str) -> Tuple:
        """
        获取ip冲突检测的命令
        :param addr_ipv4: ip地址
        :return:
        """
        local_ip = LocalIpChecker().get_ip_address(self.eth_name.encode())
        arping_cmd = (
            cmd_constants.OS_CMD_ARPING, "-C", "1", "-c", "1", "-i", self.eth_name, "-S", local_ip, addr_ipv4
        )
        return arping_cmd

    def _modify_netplan_data(self, data: Dict) -> NoReturn:
        """
        修改网卡配置文件
        :param data: 网卡配置文件01-netcfg.yaml的内容
        :return:
        """
        # 清空原有配置，如果配置文件中不存在，添加默认值
        if self.eth_name not in data["network"]["ethernets"]:
            data["network"]["ethernets"][self.eth_name] = {
                "dhcp4": "no",
                "addresses": [],
            }
        data["network"]["ethernets"][self.eth_name]["addresses"].clear()

        # 清除vlan配置
        if "vlans" in data["network"]:
            for vlan in copy.deepcopy(data["network"]["vlans"]):
                if vlan.startswith(self.eth_name):
                    data["network"]["vlans"].pop(vlan)

            # 如果vlan已经为空，则删除
            if not data["network"]["vlans"]:
                data["network"].pop("vlans")

        vlan_map = defaultdict(list)
        for ip_dict in self.ip_info_items:
            new_ip_addr = ip_dict.get("Address")
            new_mask_num = self._convert_mask_num(ip_dict.get("SubnetMask"))
            new_vlan_id = ip_dict.get("VlanId")
            new_gateway = ip_dict.get("Gateway")

            # 下发了Vlan配置则配置vlan接口
            if new_vlan_id is not None:
                if "vlans" not in data["network"]:
                    data["network"]["vlans"] = {}

                vlan_map[new_vlan_id].append(f"{new_ip_addr}/{new_mask_num}")
                data["network"]["vlans"][f"{self.eth_name}.{new_vlan_id}"] = {
                    "id": new_vlan_id,
                    "link": self.eth_name,
                    "dhcp4": "no",
                    "addresses": vlan_map[new_vlan_id]
                }
                continue

            # 修改网关，网管唯一
            if new_gateway:
                data["network"]["ethernets"][self.eth_name]["routes"] = [{"to": "default", "via": new_gateway}]
                # 只保留routes新配置项，旧的gateway4删除
                if "gateway4" in data["network"]["ethernets"][self.eth_name]:
                    data["network"]["ethernets"][self.eth_name].pop("gateway4")
            # IP和掩码
            data["network"]["ethernets"][self.eth_name]["addresses"].append(f"{new_ip_addr}/{new_mask_num}")


class OpenEulerNetHandler(CommNetworkHandler):
    """修改网络配置的OpenEuler系统处理类"""
    ETH_CONFIG_PATH = "/etc/sysconfig/network-scripts"
    NMCLI_CMD = shutil.which("nmcli")

    def __init__(self, request_data, eth_name):
        """
        构造函数
        :param request_data: 处理参数
        :param eth_name: 网卡名称
        """
        super().__init__(request_data, eth_name)
        self.back_ifcfg_eth_files = []
        self.modify_eth = []

    def rollback_ipv4(self) -> NoReturn:
        """回滚网络配置"""
        # 删除本次新配置的Vlan
        for vlan in self._get_cur_vlan_collection():
            ExecCmd.exec_cmd((self.NMCLI_CMD, "con", "delete", vlan))

        # 回滚系统网络配置文件
        for path in self.back_ifcfg_eth_files:
            shutil.copyfile(path, Path(self.ETH_CONFIG_PATH).joinpath(path.name.replace("back_", "")))

        # 执行网卡重启
        self._restart_interface([path.name.split("-")[1] for path in self.back_ifcfg_eth_files])

    def modify_network(self) -> NoReturn:
        """修改网络配置"""
        # 先备份网卡配置文件
        self._backup_netcfg_file()

        # 修改网卡配置文件
        self._modify_ipv4_netcfg()

    def effect_network(self) -> NoReturn:
        """生效网络配置"""
        self._restart_interface(self.modify_eth, raise_exception=True)

    def _backup_netcfg_file(self) -> NoReturn:
        """备份网卡文件"""
        # 网卡文件
        for path in islice(Path(self.ETH_CONFIG_PATH).glob(f"ifcfg-{self.eth_name}*"), CommonConstants.MAX_ITER_LIMIT):
            back_file = Path(self.ETH_CONFIG_PATH).joinpath(f"back_{path.name}")
            shutil.copyfile(path, back_file)
            self.back_ifcfg_eth_files.append(back_file)

    def _modify_ipv4_netcfg(self) -> NoReturn:
        """修改ipv4网络配置"""
        eth_cfg = []
        vlan_cfg = {}
        new_gateway = ""
        for req_item in self.ip_info_items:
            ipv4 = req_item.get("Address")
            new_gateway = req_item.get("Gateway") or new_gateway
            mask_num = self._convert_mask_num(req_item.get("SubnetMask"))
            vlan_id = req_item.get("VlanId")
            if vlan_id is None:
                eth_cfg.append(f"{ipv4}/{mask_num}")
                continue
            vlan_cfg[vlan_id] = f"{ipv4}/{mask_num}"

        # 配置接口前先删掉已存在的配置
        for vlan in self._get_cur_vlan_collection():
            if ExecCmd.exec_cmd((self.NMCLI_CMD, "con", "delete", vlan)) != 0:
                raise NicOperateError(f"delete vlan link <{vlan}> failed")

        if os.path.exists(os.path.join(self.ETH_CONFIG_PATH, f"ifcfg-{self.eth_name}")):
            if ExecCmd.exec_cmd((self.NMCLI_CMD, "con", "delete", self.eth_name)) != 0:
                raise NicOperateError(f"delete link <{self.eth_name}> failed")

        # 添加普通网卡接口配置
        if eth_cfg:
            nmcli_cmd = (
                self.NMCLI_CMD, "con", "add", "con-name", self.eth_name,
                "ifname", self.eth_name, "type", "ethernet", "ip4", ",".join(eth_cfg)
            )
            if new_gateway:
                nmcli_cmd += ("gw4", new_gateway)
            if ExecCmd.exec_cmd(nmcli_cmd, wait=10) != 0:
                raise NicOperateError(f"bash <{nmcli_cmd}> failed")
            self.modify_eth.append(self.eth_name)

        # 添加vlan网卡接口配置
        for v_id, ip in vlan_cfg.items():
            eth_vlan = f"{self.eth_name}.{v_id}"
            nmcli_cmd = (
                self.NMCLI_CMD, "con", "add", "type", "vlan", "con-name", eth_vlan,
                "dev", self.eth_name, "id", f"{v_id}", "ip4", ip
            )
            if ExecCmd.exec_cmd(nmcli_cmd, wait=10) != 0:
                raise NicOperateError(f"bash <{nmcli_cmd}> failed")
            self.modify_eth.append(eth_vlan)

    def _restart_interface(self, eth_li, raise_exception=False) -> NoReturn:
        """
        重启指定网卡接口
        :param eth_li: 网卡接口列表
        :param raise_exception: 是否抛出异常
        """
        for eth in eth_li:
            nmcli_cmds = (
                (self.NMCLI_CMD, "con", "reload", f"ifcfg-{eth}"),  # # 重载网卡配置文件
                (self.NMCLI_CMD, "con", "up", eth)  # 网卡up
            )
            for cmd in nmcli_cmds:
                if ExecCmd.exec_cmd(cmd, wait=10) != 0 and raise_exception:
                    raise NicOperateError(f"bash <{cmd}> failed")
                time.sleep(2)
