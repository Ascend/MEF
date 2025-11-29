# coding: utf-8
#  Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import os
from abc import ABC
from abc import abstractmethod
from enum import Enum
from typing import Dict
from typing import Iterable
from typing import List
from typing import Optional
from typing import Tuple
from typing import Type

from common.constants.base_constants import CommonConstants
from common.exception.biz_exception import BizException
from common.file_utils import FileReader
from common.file_utils import FileUtils
from common.init_cmd import cmd_constants
from common.log.logger import run_log
from common.utils.exec_cmd import ExecCmd
from common.checkers import IPV4Checker
from common.yaml.yaml_methods import YamlMethod


class RouteCfgError(Exception):
    """路由配置异常"""
    pass


class ActionType(Enum):
    CONFIG = "config"
    ROLLBACK = "rollback"


class BaseCfgCmd(ABC):
    """配置命令基类"""

    def __init__(self, route_table_id: str):
        self.route_table_id = route_table_id

    def config(self) -> None:
        """
        先检查是否需要执行，再执行配置命令。
        """
        if not self._check_need_config():
            run_log.info("check config already exist, no need config.")
            return

        self._exec_config_cmd()

    def rollback(self) -> None:
        """
        先检查是否需要执行，再执行配置回滚命令。
        """
        if not self._check_need_rollback():
            run_log.info("check config not exist, no need rollback.")
            return

        self._exec_rollback_cmd()

    @abstractmethod
    def _check_cfg_exist(self) -> bool:
        pass

    @abstractmethod
    def _exec_config_cmd(self) -> bool:
        pass

    @abstractmethod
    def _exec_rollback_cmd(self) -> bool:
        pass

    def _check_need_config(self) -> bool:
        return not self._check_cfg_exist()

    def _check_need_rollback(self) -> bool:
        return self._check_cfg_exist()

    def _exec_cmd(self, cmd: Tuple[str, ...], action: ActionType = ActionType.CONFIG) -> None:
        cmd_str = " ".join(cmd)
        ret, msg = ExecCmd.exec_cmd_get_output(cmd)
        if ret != 0:
            run_log.error("exec %s cmd failed: cmd=%s, ret=%s, msg=%s", action, cmd_str, ret, msg)
            raise RouteCfgError(f"{self.__class__.__name__} {action.value} failed")

        run_log.info("exec %s cmd success: cmd=%s", action.value, cmd_str)


class AddRouteCfg(BaseCfgCmd):
    """添加路由配置命令"""

    def __init__(self, route_table_id: str, dest_ip: str, eth_gw_ip: str):
        super().__init__(route_table_id)
        self.dest_ip = dest_ip
        self.eth_gw_ip = eth_gw_ip
        self.dest_net = f"{dest_ip}/32"

    def _check_cfg_exist(self) -> bool:
        check_route_cmd = (
            "ip", "route", "show", "to", self.dest_net, "table", self.route_table_id, "via",
            self.eth_gw_ip
        )
        ret, msg = ExecCmd.exec_cmd_get_output(check_route_cmd)
        run_log.info("check cmd result: cmd=%s, ret=%s, msg=%s", " ".join(check_route_cmd), ret, msg)
        return ret == 0 and self.dest_ip in msg

    def _exec_config_cmd(self) -> None:
        add_route_cmd = "ip", "route", "add", "to", self.dest_net, "table", self.route_table_id, "via", self.eth_gw_ip
        self._exec_cmd(add_route_cmd)

    def _exec_rollback_cmd(self) -> None:
        del_route_cmd = "ip", "route", "del", "to", self.dest_net, "table", self.route_table_id, "via", self.eth_gw_ip
        self._exec_cmd(del_route_cmd, action=ActionType.ROLLBACK)


class AddRuleCfg(BaseCfgCmd):
    """添加策略规则命令"""

    def __init__(self, route_table_id: str, dest_ip: str):
        super().__init__(route_table_id)
        self.dest_ip = dest_ip
        self.dest_net = f"{dest_ip}/32"

    def _check_cfg_exist(self) -> bool:
        check_rule_cmd = "ip", "rule", "show", "to", self.dest_net, "table", self.route_table_id
        ret, msg = ExecCmd.exec_cmd_get_output(check_rule_cmd)
        run_log.info("check cmd result: cmd=%s, ret=%s, msg=%s", " ".join(check_rule_cmd), ret, msg)
        return ret == 0 and self.dest_ip in msg

    def _exec_config_cmd(self) -> None:
        add_rule_cmd = "ip", "rule", "add", "to", self.dest_net, "table", self.route_table_id
        self._exec_cmd(add_rule_cmd)

    def _exec_rollback_cmd(self) -> None:
        del_rule_cmd = "ip", "rule", "del", "to", self.dest_net, "table", self.route_table_id
        self._exec_cmd(del_rule_cmd, action=ActionType.ROLLBACK)


class PolicyScriptCmd(BaseCfgCmd, ABC):
    """依赖驱动提供的策略脚本的配置命令类"""
    # 驱动提供的路由持久化脚本，方便实现路由持久化数据的管理，主要使用配置持久化功能，其他路由配置直接透传给ip命令处理
    POLICY_BASED_ROUTING_SH = "policy_based_routing.sh"

    def __init__(self, route_table_id: str):
        super().__init__(route_table_id)
        self._check_routing_sh_valid()

    def _check_routing_sh_valid(self) -> None:
        get_sh_path_cmd = "which", self.POLICY_BASED_ROUTING_SH
        ret, msg = ExecCmd.exec_cmd_get_output(get_sh_path_cmd)
        if ret != 0:
            run_log.error("exec cmd failed: cmd=%s, ret=%s, msg=%s", " ".join(get_sh_path_cmd), ret, msg)
            raise RouteCfgError(f"get {self.POLICY_BASED_ROUTING_SH} path failed")

        ret = FileUtils.check_script_file_valid(sh_file=msg.strip(), user="root", group="root")
        if not ret:
            run_log.error("check %s file path invalid: %s", self.POLICY_BASED_ROUTING_SH, ret.error)
            raise RouteCfgError(f"{self.POLICY_BASED_ROUTING_SH} invalid")


class SaveRouteCfg(PolicyScriptCmd):
    """保存路由表内容命令"""

    def _check_cfg_exist(self) -> bool:
        check_save_cmd = f"{self.POLICY_BASED_ROUTING_SH} ip route save list | grep -w {self.route_table_id}"
        ret, msg = ExecCmd.exec_cmd_use_pipe_symbol(check_save_cmd)
        return ret == 0 and self.route_table_id in msg

    def _exec_config_cmd(self) -> None:
        save_route_cmd = (
            self.POLICY_BASED_ROUTING_SH, f"ip route save table {self.route_table_id}", "export",
            self.route_table_id
        )
        self._exec_cmd(save_route_cmd)

    def _exec_rollback_cmd(self) -> None:
        cancel_saved_route_cmd = self.POLICY_BASED_ROUTING_SH, "ip", "route", "save", "cancel", self.route_table_id
        self._exec_cmd(cancel_saved_route_cmd, action=ActionType.ROLLBACK)


class SaveRuleCfg(PolicyScriptCmd):
    """保存策略规则命令"""

    def _check_cfg_exist(self) -> bool:
        check_save_cmd = f"{self.POLICY_BASED_ROUTING_SH} ip rule save list | grep -w {self.route_table_id}"
        ret, msg = ExecCmd.exec_cmd_use_pipe_symbol(check_save_cmd)
        return ret == 0 and self.route_table_id in msg

    def _exec_config_cmd(self) -> None:
        save_rule_cmd = self.POLICY_BASED_ROUTING_SH, "ip rule save", "export", self.route_table_id
        self._exec_cmd(save_rule_cmd)

    def _exec_rollback_cmd(self) -> None:
        cancel_saved_rule_cmd = self.POLICY_BASED_ROUTING_SH, "ip", "rule", "save", "cancel", self.route_table_id
        self._exec_cmd(cancel_saved_rule_cmd, action=ActionType.ROLLBACK)


class PersistRouteCfg(PolicyScriptCmd):
    """持久化路由表内容命令"""

    def _check_cfg_exist(self) -> bool:
        check_persist_cmd = f"{self.POLICY_BASED_ROUTING_SH} ip route persist list | grep -w {self.route_table_id}"
        ret, msg = ExecCmd.exec_cmd_use_pipe_symbol(check_persist_cmd)
        return ret == 0 and self.route_table_id in msg

    def _exec_config_cmd(self) -> None:
        persist_route_cmd = self.POLICY_BASED_ROUTING_SH, "ip", "route", "persist", self.route_table_id
        self._exec_cmd(persist_route_cmd)

    def _exec_rollback_cmd(self) -> None:
        cancel_persist_route_cmd = self.POLICY_BASED_ROUTING_SH, "ip", "route", "persist", "cancel", self.route_table_id
        self._exec_cmd(cancel_persist_route_cmd, action=ActionType.ROLLBACK)


class PersistRuleCfg(PolicyScriptCmd):
    """持久化策略规则命令"""

    def _check_cfg_exist(self) -> bool:
        check_persist_cmd = f"{self.POLICY_BASED_ROUTING_SH} ip rule persist list | grep -w {self.route_table_id}"
        ret, msg = ExecCmd.exec_cmd_use_pipe_symbol(check_persist_cmd)
        return ret == 0 and self.route_table_id in msg

    def _exec_config_cmd(self) -> None:
        persist_rule_cmd = self.POLICY_BASED_ROUTING_SH, "ip", "rule", "persist", self.route_table_id
        self._exec_cmd(persist_rule_cmd)

    def _exec_rollback_cmd(self) -> None:
        cancel_persist_rule_cmd = self.POLICY_BASED_ROUTING_SH, "ip", "rule", "persist", "cancel", self.route_table_id
        self._exec_cmd(cancel_persist_rule_cmd, action=ActionType.ROLLBACK)


class DeleteDefRouteCfg(BaseCfgCmd):
    """删除默认路由配置"""

    def __init__(self, route_table_id: str, eth_gw_ip: str):
        super().__init__(route_table_id)
        self.eth_gw_ip = eth_gw_ip

    def _check_cfg_exist(self) -> bool:
        check_default_route_cmd = "ip", "route", "show", "default", "via", self.eth_gw_ip
        ret, msg = ExecCmd.exec_cmd_get_output(check_default_route_cmd)
        # 如果未查询到默认路由，则认为已经满足删除的条件，不需要再执行删除
        return ret != 0 or "default dev" not in msg

    def _exec_config_cmd(self) -> None:
        del_default_route_cmd = "ip", "route", "delete", "default"
        self._exec_cmd(del_default_route_cmd)

    def _exec_rollback_cmd(self) -> None:
        add_default_route_cmd = "ip", "route", "add", "default", "via", self.eth_gw_ip
        self._exec_cmd(add_default_route_cmd, action=ActionType.ROLLBACK)


class SaveGatewayAndRouteTableID(BaseCfgCmd):
    """保存有线网络的默认网关、策略路由的路由表ID"""
    # OM侧保存数据的配置文件
    CFG_INI_FILE = CommonConstants.IBMA_EDGE_SERVICE_PATH
    POLICY_SECTION = "policy_route"
    ROUTE_OPTION = "route_table_id"
    GATEWAY_OPTION = "eth_gw_ip"
    ETH_NAME_OPTION = "eth_name"

    def __init__(self, route_table_id: str, eth_gw_ip: str, eth_name: str):
        super().__init__(route_table_id)
        self.eth_gw_ip = eth_gw_ip
        self.eth_name = eth_name

    def _check_cfg_exist(self) -> bool:
        # 配置文件没有策略路由配置部分
        if not FileUtils.check_section_exist(self.CFG_INI_FILE, self.POLICY_SECTION):
            return False

        return all(self._check_file_item_exists())

    def _check_need_rollback(self) -> bool:
        if not FileUtils.check_section_exist(self.CFG_INI_FILE, self.POLICY_SECTION):
            return False

        return any(self._check_file_item_exists())

    def _check_file_item_exists(self) -> Tuple[bool, ...]:
        option_dict = FileUtils.get_option_list(self.CFG_INI_FILE, self.POLICY_SECTION)
        route_id_exist = option_dict.get(self.ROUTE_OPTION) == self.route_table_id
        eth_gw_ip_exist = option_dict.get(self.GATEWAY_OPTION) == self.eth_gw_ip
        eth_name_exist = option_dict.get(self.ETH_NAME_OPTION) == self.eth_name
        return route_id_exist, eth_gw_ip_exist, eth_name_exist

    def _exec_config_cmd(self) -> None:
        option_dict = {
            self.ROUTE_OPTION: self.route_table_id,
            self.GATEWAY_OPTION: self.eth_gw_ip,
            self.ETH_NAME_OPTION: self.eth_name,
        }
        exist = FileUtils.check_section_exist(self.CFG_INI_FILE, self.POLICY_SECTION)
        if not exist:
            FileUtils.add_one_section(self.CFG_INI_FILE, self.POLICY_SECTION, option_dict)
            run_log.info("add one section %s success: %s", self.POLICY_SECTION, option_dict)
            return

        for option, value in option_dict.items():
            FileUtils.modify_one_option(self.CFG_INI_FILE, self.POLICY_SECTION, option, value)
            run_log.info("modify section %s success: %s = %s", self.POLICY_SECTION, option, value)
        return

    def _exec_rollback_cmd(self) -> None:
        cfg = FileUtils.get_config_parser(self.CFG_INI_FILE)
        cfg.remove_section(self.POLICY_SECTION)
        FileUtils.write_file_with_lock(self.CFG_INI_FILE, cfg.write)
        run_log.info("remove section %s success", self.POLICY_SECTION)


class DeleteGateway(BaseCfgCmd, ABC):
    """删除系统中的默认网关配置"""

    def __init__(self, route_table_id: str, eth_gw_ip: str, eth_name: str):
        super().__init__(route_table_id)
        self.eth_gw_ip = eth_gw_ip
        self.eth_name = eth_name

    @staticmethod
    def _check_cfg_path_valid(network_cfg: str) -> bool:
        ret = FileUtils.check_script_file_valid(network_cfg, user="root", group="root")
        if not ret:
            run_log.error("%s path invalid: %s", network_cfg, ret.error)
            return False

        if os.path.getsize(network_cfg) > CommonConstants.OM_READ_FILE_MAX_SIZE_BYTES:
            run_log.error("%s path invalid: too large", network_cfg)
            return False

        return True

    def _get_gateway_by_normal(self, network_cfg: str) -> str:
        if not self._check_cfg_path_valid(network_cfg):
            return ""

        ret = FileReader(network_cfg).readlines()
        if not ret:
            run_log.error("read data from cfg file failed: %s", ret.error)
            return ""

        for line in ret.data:
            if line.strip().startswith("GATEWAY="):
                return line.strip().split("=")[1].strip("\"")
        return ""

    def _set_gateway_by_normal(self, network_cfg: str, gateway_ip: str) -> bool:
        if not self._check_cfg_path_valid(network_cfg):
            return False

        ret = FileReader(network_cfg).readlines()
        if not ret:
            run_log.error("read data from cfg file failed: %s", ret.error)
            return False

        data = []
        found = False
        modify = False
        modified_line = f"GATEWAY={gateway_ip}{os.linesep}"
        for line in ret.data:
            if line.strip().startswith("GATEWAY="):
                data.append(modified_line)
                found = True
                modify = line != modified_line
                continue
            data.append(line)

        if not found:
            data.append(modified_line)
            modify = True

        if not modify:
            return True

        try:
            with os.fdopen(os.open(network_cfg, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o644), "w") as file:
                file.writelines(data)
        except Exception as error:
            run_log.error("modify config file %s set GATEWAY to %s success, caught exception: %s", network_cfg,
                          gateway_ip or "empty", error)
            return False

        # 修改成功记录日志
        run_log.info("modify config file %s set GATEWAY to %s success", network_cfg, gateway_ip or "empty")
        return True


class DeleteEulerGateway(DeleteGateway):
    """删除EulerOS系统中的默认网关配置"""

    def _check_cfg_exist(self) -> bool:
        return not self._get_gateway()

    def _exec_config_cmd(self) -> None:
        # gateway为空，表示删除
        if not self._set_gateway("", self.eth_name):
            run_log.error("set gateway to empty for EulerOS failed")
            raise RouteCfgError(f"{self.__class__.__name__} config failed")

    def _exec_rollback_cmd(self) -> None:
        if not self._set_gateway(self.eth_gw_ip, self.eth_name):
            run_log.error("set gateway to %s for EulerOS failed", self.eth_gw_ip)
            raise RouteCfgError(f"{self.__class__.__name__} rollback failed")

    def _get_gateway(self) -> str:
        network_cfg_files = f"/etc/sysconfig/network-scripts/ifcfg-{self.eth_name}", "/etc/sysconfig/network"
        for cfg_file in network_cfg_files:
            gateway = self._get_gateway_by_normal(cfg_file)
            if gateway:
                return gateway
        return ""

    def _set_gateway(self, gateway_ip: str, eth_name: str) -> bool:
        # euleros存在两个地方配置GATEWAY,与Nic模块保持一致,只配置ifcfg的配置文件
        network_cfg_file = "/etc/sysconfig/network"
        # 不管是清空，还是配置，都将/etc/sysconfig/network中的配置项置空
        if not self._set_gateway_by_normal(network_cfg_file, ""):
            run_log.error("modify %s failed when setting GATEWAY to empty", network_cfg_file)
            return False

        network_cfg_file = f"/etc/sysconfig/network-scripts/ifcfg-{eth_name}"
        if not self._set_gateway_by_normal(network_cfg_file, gateway_ip):
            run_log.error("modify %s failed when setting GATEWAY to %s", network_cfg_file, gateway_ip or "empty")
            return False

        return True


class DeleteOpenEulerGateway(DeleteGateway):
    """删除openEuler系统中的默认网关配置"""

    def _check_cfg_exist(self) -> bool:
        return not self._get_gateway()

    def _exec_config_cmd(self) -> None:
        # gateway为空，表示删除
        if not self._set_gateway("", self.eth_name):
            run_log.error("set gateway to empty for openEuler failed")
            raise RouteCfgError(f"{self.__class__.__name__} config failed")

    def _exec_rollback_cmd(self) -> None:
        if not self._set_gateway(self.eth_gw_ip, self.eth_name):
            run_log.error("set gateway to %s for openEuler failed", self.eth_gw_ip)
            raise RouteCfgError(f"{self.__class__.__name__} rollback failed")

    def _get_gateway(self) -> str:
        network_cfg = f"/etc/sysconfig/network-scripts/ifcfg-{self.eth_name}"
        return self._get_gateway_by_normal(network_cfg)

    def _set_gateway(self, gateway_ip: str, eth_name: str) -> bool:
        network_cfg = f"/etc/sysconfig/network-scripts/ifcfg-{eth_name}"
        return self._set_gateway_by_normal(network_cfg, gateway_ip)


class DeleteUbuntuGateway(DeleteGateway):
    """删除Ubuntu系统中的默认网关配置"""
    NETCFG_YAML = "/etc/netplan/01-netcfg.yaml"
    GATEWAY4_NAME = "gateway4"
    ROUTES_NAME = "routes"
    DEFAULT_NAME = "default"

    def _check_cfg_exist(self) -> bool:
        return not self._get_gateway()

    def _exec_config_cmd(self) -> None:
        # gateway为空，表示删除
        if not self._set_gateway("", self.eth_name):
            run_log.error("set gateway to empty for Ubuntu failed")
            raise RouteCfgError(f"{self.__class__.__name__} config failed")

    def _exec_rollback_cmd(self) -> None:
        if not self._set_gateway(self.eth_gw_ip, self.eth_name):
            run_log.error("set gateway to %s for Ubuntu failed", self.eth_gw_ip)
            raise RouteCfgError(f"{self.__class__.__name__} rollback failed")

    def _get_gateway(self) -> str:
        try:
            data = YamlMethod.load_yaml_info(self.NETCFG_YAML)
        except Exception as error:
            run_log.error("load yaml config failed, caught exception: %s", error)
            return ""

        try:
            eth_data = data["network"]["ethernets"][self.eth_name]
        except Exception as error:
            run_log.error("get config data for %s failed, caught exception: %s", self.eth_name, error)
            return ""

        if not isinstance(eth_data, dict) or not eth_data:
            return ""

        # 默认先找gateway4配置项，找到则返回；否则继续从routes中找
        gateway = eth_data.get(self.GATEWAY4_NAME)
        if gateway:
            return gateway

        for route in eth_data.get(self.ROUTES_NAME, []):
            if route.get("to") == self.DEFAULT_NAME:
                return route.get("via", "")

        return ""

    def _set_gateway(self, gateway_ip: str, eth_name: str) -> bool:
        # 加载yaml
        try:
            data = YamlMethod.load_yaml_info(self.NETCFG_YAML)
        except Exception as error:
            run_log.error("load yaml config failed, caught exception: %s", error)
            return False

        # 修改配置数据
        modify = False
        try:
            if eth_name in data["network"]["ethernets"]:
                eth_data = data["network"]["ethernets"][eth_name]
                # 先删除默认网关的配置，保留非默认路由
                routes_cfg = [cfg for cfg in eth_data.get(self.ROUTES_NAME, []) if cfg.get("to") != self.DEFAULT_NAME]
                if self.GATEWAY4_NAME in eth_data:
                    # 如果存在旧的配置项则删除，统一采用routes新配置项方式
                    eth_data.pop(self.GATEWAY4_NAME)
                    modify = True

                if gateway_ip:
                    # 有值修改场景，则追加默认网关路由配置项
                    routes_cfg.append({"to": self.DEFAULT_NAME, "via": gateway_ip})
                    modify = True
                elif self.ROUTES_NAME in eth_data:
                    # 无值删除场景，先整个routes项删除，再根据routes_cfg是否为空进行赋值
                    eth_data.pop(self.ROUTES_NAME)
                    modify = True

                if routes_cfg:
                    eth_data[self.ROUTES_NAME] = routes_cfg
                    modify = True
        except Exception as error:
            run_log.error("modify config failed, caught exception: %s", error)
            return False

        if not modify:
            return True

        # 重写yaml文件
        try:
            YamlMethod.dumps_yaml_file(data, self.NETCFG_YAML, 0o644)
        except Exception as error:
            run_log.error("write net config yaml failed, caught exception: %s", error)
            return False

        return True


DELETE_GATEWAY_IMPLS: Dict[str, Type[DeleteGateway]] = {
    "EulerOS": DeleteEulerGateway,
    "openEuler": DeleteOpenEulerGateway,
    "Ubuntu": DeleteUbuntuGateway,
}


def delete_gateway_factory(route_table_id: str, eth_gw_ip: str, eth_name: str) -> DeleteGateway:
    os_name = cmd_constants.OS_NAME
    cls = DELETE_GATEWAY_IMPLS.get(os_name)
    if not cls:
        run_log.error("unsupported os name: %s", os_name)
        raise RouteCfgError(f"unsupported os name: {os_name}")

    return cls(route_table_id, eth_gw_ip, eth_name)


class PolicyBasedRouting:
    """基于策略的路由配置"""
    # 路由表ID: 全局取值范围[0,255],其中0、254、255系统默认分配被占用
    MIN_TABLE_ID = 1
    MAX_TABLE_ID = 253
    MAX_ID_STR_LEN = 3

    def __init__(self, eth_name: str, eth_gw_ip: str, route_table_id: Optional[str] = None):
        if not route_table_id:
            route_table_id = self._choose_available_table_id()

        if not self._check_table_id_valid(route_table_id):
            raise RouteCfgError("invalid route table id")

        if eth_gw_ip and not IPV4Checker("ip").check({"ip": eth_gw_ip}):
            raise RouteCfgError("invalid ethernet gateway ip")

        self.route_table_id = route_table_id
        self.eth_name = eth_name
        self.eth_gw_ip = eth_gw_ip

    @staticmethod
    def _check_route_table_available(table_id) -> bool:
        """
        检查table_id路由表是否可用，包括不存在、存在且为空
        :param table_id: 路由表ID
        :return: True-可用；False-不可用
        """
        chk_table_exists_cmd = rf"ip route show table {table_id}"
        ret, out = ExecCmd.exec_cmd_use_pipe_symbol(chk_table_exists_cmd)
        not_exist = ret != 0 and "table does not exist" in out
        exist_but_empty = ret == 0 and not out.strip()
        return not_exist or exist_but_empty

    @classmethod
    def _choose_available_table_id(cls) -> str:
        for table_id in range(cls.MIN_TABLE_ID, cls.MAX_TABLE_ID):
            if not cls._check_route_table_available(table_id):
                continue
            run_log.info("choose available route table id: %s", table_id)
            return str(table_id)

        raise RouteCfgError("can not find available route table id")

    def config_policy_route(self, dest_ips: List[str]) -> bool:
        """
        配置小站通过有线网卡默认网关eth_gw_ip到dest_ips地址的策略路由
        :param dest_ips: 目的地址IPv4字符串列表
        :return: True-配置成功；False-配置失败
        """
        if not self.eth_gw_ip:
            run_log.warning("ether net gateway is empty, no need config")
            return True

        return self._config_route_cmds(dest_ips, self.eth_gw_ip)

    def cancel_policy_route(self, dest_ips: List[str]) -> None:
        """
        无线网络关闭时取消策略路由的相关配置，恢复有线网络默认路由配置
        :param dest_ips: 目的地址IPv4字符串列表
        :return: None
        """
        try:
            for cmd in self._cancel_task_generator(dest_ips, self.eth_gw_ip):
                try:
                    cmd.rollback()
                except (RouteCfgError, BizException) as error:
                    run_log.error("exec rollback cmd failed: %s, because: %s", cmd.__class__.__name__, error)
                    continue

                run_log.info("exec rollback cmd success: %s", cmd.__class__.__name__)
        except (RouteCfgError, BizException) as error:
            run_log.error("cancel policy route failed, caught exception: %s", error)

    def _config_route_cmds(self, dest_ips: List[str], eth_gw_ip: str) -> bool:
        # 记录已经成功执行的命令
        executed_cmds: List[BaseCfgCmd] = []
        need_rollback = False
        try:
            for cmd in self._config_task_generator(dest_ips, eth_gw_ip):
                executed_cmds.append(cmd)
                cmd.config()
                run_log.info("exec config cmd success: %s", cmd.__class__.__name__)
        except (RouteCfgError, BizException) as error:
            need_rollback = True
            run_log.error("exec config cmd failed, caught exception: %s", error)

        if not need_rollback:
            run_log.info("config route cmds all done success")
            return True

        for cmd in reversed(executed_cmds):
            try:
                cmd.rollback()
            except (RouteCfgError, BizException) as error:
                # 倒序执行撤销命令失败，记录告警日志，不影响其他的回滚命令
                run_log.warning("exec rollback cmd failed: %s, caught exception: %s", cmd.__class__.__name__, error)
                continue

            run_log.warning("exec rollback cmd success: %s", cmd.__class__.__name__)

        return False

    def _config_task_generator(self, dest_ips: List[str], eth_gw_ip: str) -> Iterable[BaseCfgCmd]:
        for dest_ip in dest_ips:
            # 配置一条到dest_ip的路由规则到该路由表
            yield AddRouteCfg(self.route_table_id, dest_ip, eth_gw_ip)
            # 新增该路由表的策略路由规则
            yield AddRuleCfg(self.route_table_id, dest_ip)
        # 路由表内容保存至文件
        yield SaveRouteCfg(self.route_table_id)
        # 策略路由规则信息保存至文件
        yield SaveRuleCfg(self.route_table_id)
        # 路由表持久化注册
        yield PersistRouteCfg(self.route_table_id)
        # 策略路由持久化
        yield PersistRuleCfg(self.route_table_id)
        # 保存默认网关IP地址与路由表ID
        yield SaveGatewayAndRouteTableID(self.route_table_id, self.eth_gw_ip, self.eth_name)
        # 删除ethernet的默认路由
        yield DeleteDefRouteCfg(self.route_table_id, self.eth_gw_ip)
        # 删除默认网关配置文件
        yield delete_gateway_factory(self.route_table_id, self.eth_gw_ip, self.eth_name)

    def _cancel_task_generator(self, dest_ips: List[str], eth_gw_ip: str) -> Iterable[BaseCfgCmd]:
        # 恢复默认网关配置文件
        yield delete_gateway_factory(self.route_table_id, self.eth_gw_ip, self.eth_name)
        # 恢复ethernet的默认路由
        yield DeleteDefRouteCfg(self.route_table_id, self.eth_gw_ip)
        # 删除默认网关IP地址与路由表ID
        yield SaveGatewayAndRouteTableID(self.route_table_id, self.eth_gw_ip, self.eth_name)
        # 取消策略路由持久化
        yield PersistRuleCfg(self.route_table_id)
        # 取消路由表持久化注册
        yield PersistRouteCfg(self.route_table_id)
        # 删除策略路由规则信息文件
        yield SaveRuleCfg(self.route_table_id)
        # 删除路由表内容保存文件
        yield SaveRouteCfg(self.route_table_id)
        for dest_ip in dest_ips:
            # 删除该路由表的策略路由规则
            yield AddRuleCfg(self.route_table_id, dest_ip)
            # 删除配置到dest_ip的路由规则到该路由表
            yield AddRouteCfg(self.route_table_id, dest_ip, eth_gw_ip)

    def _check_table_id_valid(self, table_id: str) -> bool:
        return table_id and len(table_id) <= self.MAX_ID_STR_LEN \
               and self.MIN_TABLE_ID <= int(table_id) <= self.MAX_TABLE_ID
