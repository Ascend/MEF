import threading
from collections import namedtuple
from unittest.mock import mock_open

import pytest
from pytest_mock import MockFixture

from common.checkers.param_checker import EthernetInterfaceChecker
from common.init_cmd import cmd_constants
from common.utils.exec_cmd import ExecCmd
from common.checkers import CheckResult, LocalIpChecker
from common.utils.result_base import Result
from lib.Linux.systems.nic.config_web_ip import NginxConfig
from lib.Linux.systems.nic.nic import Nic, NicOperateError, UbuntuNetworkHandler, EulerNetworkHandler, \
    OpenEulerNetHandler, CommNetworkHandler
from lib.Linux.systems.nic.nic_mgr import NetConfigMgr
from common.common_methods import CommonMethods
from common.yaml.yaml_methods import YamlMethod
from ut_utils.mock_utils import mock_read_data

GetMaskForNum = namedtuple("GetMaskForNum", "expect, mask")
GetEthName = namedtuple("GetEthName", "expect, name")
GetFileContent = namedtuple("GetFileContent", "expect filepath, exist, content")
ModifyNetworkConfig = namedtuple("GetFileContent", "os_name, request_data, eth_name, exception")
CheckWebTag = namedtuple("CheckWebTag", "tag_list, eth_li, query_web, exception")
GetIpCmdJsonInfo = namedtuple("GetIpCmdJsonInfo", "expect, exec_ok, json_load, exception")
GetGatewayByDev = namedtuple("GetGatewayByDev", "expect, gateway")
GetIpAddrShowInfo = namedtuple("GetIpAddrShowInfo", "expect, ip_info, catch")
GetEthXInfo = namedtuple("GetEthXInfo", "expect, content")
CheckRequestData = namedtuple("CheckRequestData", "expect, check, get_eth, ip_info, check_web")
GetAllInfo = namedtuple("GetAllInfo", "expect, gmac, eth_li")
GetDynamicInfo = namedtuple("GetDynamicInfo", "name, content, link_state, eth_info, ipv4_info, ipv6_info")
GetIpv4AddrInfo = namedtuple("GetIpv4AddrInfo", "expect, gw, ip_info, addr_mode, tag")
GetIpv6AddrInfo = namedtuple("GetIpv4AddrInfo", "expect, gw, ip_info, addr_mode")
CheckEthLinkState = namedtuple("CheckEthLinkState", "expect, link_info")
GetIpv4AddrMode = namedtuple("GetIpv4AddrMode", "expect, dhcp")
GetIpv6AddrMode = namedtuple("GetIpv6AddrMode", "expect, info_str, dhcp")
GetLinkStatus = namedtuple("GetLinkStatus", "expect, link_state, exec_cmd")
GetNameServer = namedtuple("GetNameServer", "expect, content")
GetNetTrafficInfo = namedtuple("GetNetTrafficInfo", "recv, error, drop, send, traffic_info")
PatchRequest = namedtuple("PatchRequest", "expect, lock, check_rq, modify")
RestartNginx = namedtuple("RestartNginx", "expect, exec_cmd")
ConvertMaskNum = namedtuple("ConvertMaskNum", "expect, mask")
CheckIpConflict = namedtuple("CheckIpConflict", "exception, exec_cmd")
UpdateOldTagIni = namedtuple("UpdateOldTagIni", "exception, tag_list, ip_list, load_ini, write_ini")
GetCurVlanCollection = namedtuple("GetCurVlanCollection", "expect, content")
GetArpingCmd = namedtuple("GetArpingCmd", "expect, addr")
GetCurrGateWay = namedtuple("GetCurrGateWay", "expect, exec_cmd, json, exception")
CheckIpConnection = namedtuple("CheckIpConnection", "exec_cmd, exception")
NetworkCheck = namedtuple("NetworkCheck", "gateway, ip_conn, ip_conflict")
DealIpv4Network = namedtuple("DealIpv4Network", "exception")
DealWithIpv4 = namedtuple("DealWithIpv4", "expect")
EulerBackNetcfgFile = namedtuple("EulerBackNetcfgFile", "expect")
EulerModifyNetwork = namedtuple("EulerModifyNetwork", "exception")
EulerEffectNetwork = namedtuple("EulerEffectNetwork", "exception, exec_cmd")
EulerRollbackIpv4 = namedtuple("EulerRollbackIpv4", "exec_cmd")
EulerWriteIpv4Netcfg = namedtuple("EulerRollbackIpv4", "exception")
EulerWriteIpv4File = namedtuple("EulerWriteIpv4File", "exception")
UbuntuGetArpingCmd = namedtuple("UbuntuGetArpingCmd", "expect")
UbuntuModifyNetplanData = namedtuple("UbuntuModifyNetplanData", "data")
UbuntuModifyNetwork = namedtuple("UbuntuModifyNetwork", "exception, load_failed, dumps_failed")
UbuntuEffectNetwork = namedtuple("UbuntuEffectNetwork", "exception, exec_cmd")
UbuntuRollbackIpv4 = namedtuple("EulerRollbackIpv4", "exec_cmd")
OpenEulerRollbackIpv4 = namedtuple("OpenEulerRollbackIpv4", "exec_cmd")
OpenEulerModifyNetwork = namedtuple("OpenEulerModifyNetwork", "exec_cmd")
OpenEulerEffectNetwork = namedtuple("OpenEulerEffectNetwork", "exception")
OpenEulerRestartInterface = namedtuple("RestartInterface", "exception, exec_cmd")
OpenEulerBackNetcfgFile = namedtuple("OpenEulerBackNetcfgFile", "expect")
OpenEulerModifyIpv4Netcfg = namedtuple("OpenEulerModifyIpv4Netcfg", "exception, exec_cmd")


class TestNic:
    use_cases = {
        "test_get_mask_for_num": {
            "get_mask": ("255.255.255.0", 24),
        },
        "test_get_eth_name": {
            "get_eth": ("eth1", "GMAC1"),
        },
        "test_get_file_content": {
            "not_exist": ("", "1.txt", False, ""),
            "read_file": ("a", "1.txt", True, "a"),
        },
        "test_modify_network_config": {
            "not_support": ("test_os", {}, "eth1", NicOperateError),
            "deal_ipv4_euler": ("EulerOS", {}, "eth1", None),
            "deal_ipv4_open_euler": ("openEuler", {}, "eth1", None),
            "deal_ipv4_ubuntu": ("Ubuntu", {}, "eth1", None)
        },
        "test_check_web_tag": {
            "web_in_tag_li": (["web"], ["eth1"], [1], None),
            "web_in_db": (["tag"], ["eth1"], [1], None),
            "raise": (["tag"], ["eth1"], [], NicOperateError),
        },
        "test_get_ip_cmd_json_info": {
            "exec_failed": (None, [1, ""], {"a", 1}, NicOperateError),
            "ok": ({"a", 1}, [0, "{'a', 1}"], {"a", 1}, None)
        },
        "get_gateway_by_dev": {
            "get_gw": ("1.1.1.1", [{"gateway": "1.1.1.1", "dev": "eth1"}]),
            "no_gw": ("", [{}])
        },
        "test_get_ip_addr_show_info": {
            "get_ip": ({"ip": "xxx"}, [{"addr_info": [{"ip": "xxx"}]}], False)
        },
        "test_get_eth_x_info": {
            "normal": (["eth0", "eth1", "eth1.2"], "xxx\neth0: xxx\neth1: xxx\neth1.2: xxx\n")
        },
        "test_check_request_data": {
            "normal": (None, True, (["GMAC0", "GMAC1"], ["eth0", "eth1"]), [{"Address": "1.1.1.1"}], None),
            "check_err": (NicOperateError, CheckResult.make_failed("failed"), (["GMAC0", "GMAC1"], ["eth0", "eth1"]),
                          [{"Address": "1.1.1.1"}], None),
            "not_in": (NicOperateError, True, (["GMAC2", "GMAC3"], ["eth0", "eth1"]), [{"Address": "1.1.1.1"}], None),
            "duplicate_ip": (NicOperateError, True, (["GMAC0", "GMAC1"], ["eth0", "eth1"]),
                             [{"Address": "2.2.2.2"}], None),
            "check_web": (NicOperateError, True, (["GMAC0", "GMAC1"], ["eth0", "eth1"]),
                          [{"Address": "1.1.1.1"}], NicOperateError),
        },
        "test_get_all_info": {
            "items": ([CommonMethods.OK, ""], None, ([], [])),
            "not_in": ([CommonMethods.NOT_EXIST, "GMAC1 not in net_name list"], "GMAC1", (["GMAC2", "GMAC3"], [])),
            "normal": ([CommonMethods.OK, ""], "GMAC1", (["GMAC1", "GMAC2"], [])),
        },
        "test_get_dynamic_info": {
            "normal": ("GMAC1", "22:ff:ff:aa", ["LinkUp", "1000Mbps"], ([], ["eth0", "eth1"]),
                       [{"ipv4": "1"}], [{"ipv6": "1"}])
        },
        "test_get_ipv4_addresses_info": {
            "normal": ({"Address": "192.168.3.101", "SubnetMask": "255.255.255.0", "AddressOrigin": "Static",
                        "Gateway": "192.168.3.1", "Tag": "web", "VlanId": 1}, "192.168.3.1",
                       [{"local": "192.168.3.101", "prefixlen": 24, }], "Static", "web")
        },
        "test_get_ipv6_addresses_info": {
            "normal": ({"Address": "11:ff:ff:ff", "PrefixLength": 36, "AddressOrigin": "Static",
                        "AddressState": "Preferred", "Gateway": "11:ff:ff:ff"}, "11:ff:ff:ff",
                       [{"local": "11:ff:ff:ff", "prefixlen": 36, }], "Static")
        },
        "test_check_eth_link_state": {
            "UP": (True, [{"flags": "Link UP"}]),
            "DOWN": (False, [{"flags": "Link DOWN"}]),
        },
        "test_get_ipv4_addr_mode": {
            "dhcp": ("DHCP", True),
            "static": ("Static", False)
        },
        "test_get_ipv6_addr_mode": {
            "LinkLocal": ("LinkLocal", "link xxx", False),
            "SLAAC": ("SLAAC", "dynamic xxx", False),
            "DHCPv6": ("DHCPv6", "xxx xxx", True),
            "static": ("Static", "xxx xxx", False)
        },
        "test_get_link_status": {
            "LinkDown": ("LinkDown", False, [0, "yes"]),
            "LinkUp": ("LinkUp", True, [0, "yes"]),
            "NoLink": ("NoLink", True, [1, "no"]),
        },
        "test_get_name_server": {
            "no_dns": ([], "xxxxx"),
            "dns": (["1.1.1.1"], "nameserver 1.1.1.1")
        },
        "test_get_net_traffic_info": {
            "ok": ("1", "1", "1", "1", [{"stats64": {"rx": {"packets": "1", "errors": "1", "dropped": "1"},
                                                     "tx": {"packets": "1"}}}]),
            "error": ("", "", "", "", [{"stats64": {}}])
        },
        "test_patch_request": {
            "locked": ([CommonMethods.ERROR, "The operation is busy."], True, None, None),
            "modify_err": ([CommonMethods.ERROR, "deal with EthernetInterface config failed"],
                           False, None, NicOperateError),
            "check_err": ([CommonMethods.ERROR, "deal with EthernetInterface config failed"],
                          False, NicOperateError, None),
            "ok": ([CommonMethods.OK, ""], False, None, None)
        }

    }

    @staticmethod
    def test_get_mask_for_num(mocker: MockFixture, model: GetMaskForNum):
        assert Nic()._get_mask_for_num(model.mask) == model.expect

    @staticmethod
    def test_get_eth_name(mocker: MockFixture, model: GetEthName):
        assert Nic()._get_eth_name(model.name) == model.expect

    @staticmethod
    def test_get_file_content(mocker: MockFixture, model: GetFileContent):
        mocker.patch("os.path.exists", return_value=model.exist)
        mock_read_data(mocker, read_data=model.content)
        assert Nic()._get_file_content(model.filepath) == model.expect

    @staticmethod
    def test_modify_network_config(mocker: MockFixture, model: ModifyNetworkConfig):
        mocker.patch.object(UbuntuNetworkHandler, "deal_ipv4_network")
        mocker.patch.object(EulerNetworkHandler, "deal_ipv4_network")
        mocker.patch.object(OpenEulerNetHandler, "deal_ipv4_network")
        old_os_name = cmd_constants.OS_NAME
        cmd_constants.OS_NAME = model.os_name
        if model.exception:
            with pytest.raises(model.exception):
                Nic()._modify_network_config(model.request_data, model.eth_name)
        else:
            Nic()._modify_network_config(model.request_data, model.eth_name)
        cmd_constants.OS_NAME = old_os_name

    @staticmethod
    def test_check_web_tag(mocker: MockFixture, model: CheckWebTag):
        mocker.patch.object(NetConfigMgr, "query_info_with_condition").return_value = model.query_web
        if model.exception:
            with pytest.raises(model.exception):
                Nic()._check_web_tag(model.tag_list, model.eth_li)
            return

        Nic()._check_web_tag(model.tag_list, model.eth_li)

    @staticmethod
    def test_get_ip_cmd_json_info(mocker: MockFixture, model: GetIpCmdJsonInfo):
        mocker.patch.object(ExecCmd, "exec_cmd_use_pipe_symbol").return_value = model.exec_ok
        mocker.patch("json.loads").return_value = model.json_load
        if model.exception:
            with pytest.raises(model.exception):
                Nic()._get_ip_cmd_json_info("ip cmd")
            return
        assert model.expect == Nic()._get_ip_cmd_json_info("ip cmd")

    @staticmethod
    def get_gateway_by_dev(mocker: MockFixture, model: GetGatewayByDev):
        mocker.patch.object(Nic, "_get_gateway_by_dev").return_value = model.gateway
        assert model.expect == Nic()._get_gateway_by_dev("ip cmd", "eth1")

    @staticmethod
    def test_get_ip_addr_show_info(mocker: MockFixture, model: GetIpAddrShowInfo):
        mocker.patch.object(Nic, "_get_ip_cmd_json_info").return_value = model.ip_info
        for ip_info in Nic()._get_ip_addr_show_info("ip cmd"):
            assert model.expect == ip_info

    @staticmethod
    def test_get_eth_x_info(mocker: MockFixture, model: GetEthXInfo):
        mocker.patch.object(Nic, "_get_file_content").return_value = model.content
        assert Nic()._get_eth_x_info()[1] == model.expect

    @staticmethod
    def test_check_request_data(mocker: MockFixture, model: CheckRequestData):
        mocker.patch.object(EthernetInterfaceChecker, "check").return_value = model.check
        mocker.patch.object(Nic, "_get_eth_x_info").return_value = model.get_eth
        mocker.patch.object(Nic, "_get_ipv4_addresses_info").return_value = model.ip_info
        mocker.patch.object(Nic, "_check_web_tag").side_effect = model.check_web
        if not model.expect:
            Nic()._check_request_data({"IPv4Addresses": [{"Address": "2.2.2.2"}]}, "GMAC1")
            return
        with pytest.raises(NicOperateError):
            Nic()._check_request_data({"IPv4Addresses": [{"Address": "2.2.2.2"}]}, "GMAC1")

    @staticmethod
    def test_get_all_info(mocker: MockFixture, model: GetAllInfo):
        mocker.patch.object(Nic, "_get_eth_x_info").return_value = model.eth_li
        assert Nic().get_all_info(model.gmac) == model.expect

    @staticmethod
    def test_get_dynamic_info(mocker: MockFixture, model: GetDynamicInfo):
        mocker.patch.object(Nic, "_get_file_content").side_effect = model.content
        mocker.patch.object(Nic, "_get_link_status").return_value = model.link_state
        mocker.patch.object(Nic, "_get_eth_x_info").return_value = model.eth_info
        mocker.patch.object(Nic, "_get_ipv4_addresses_info").return_value = model.ipv4_info
        mocker.patch.object(Nic, "_get_ipv6_addresses_info").return_value = model.ipv6_info
        mocker.patch.object(Nic, "_get_net_traffic_info")
        mocker.patch.object(Nic, "_get_name_server")
        Nic().get_all_info()

    @staticmethod
    def test_get_ipv4_addresses_info(mocker: MockFixture, model: GetIpv4AddrInfo):
        mocker.patch.object(Nic, "_get_gateway_by_dev").return_value = model.gw
        mocker.patch.object(Nic, "_get_ip_addr_show_info").return_value = model.ip_info
        mocker.patch.object(Nic, "_get_ipv4_addr_mode").return_value = model.addr_mode
        mocker.patch.object(NetConfigMgr, "query_tag_from_ip").return_value = model.tag
        for ipv4 in Nic()._get_ipv4_addresses_info("eth1.1"):
            assert model.expect == ipv4

    @staticmethod
    def test_get_ipv6_addresses_info(mocker: MockFixture, model: GetIpv6AddrInfo):
        mocker.patch.object(Nic, "_get_gateway_by_dev").return_value = model.gw
        mocker.patch.object(Nic, "_get_ip_addr_show_info").return_value = model.ip_info
        mocker.patch.object(Nic, "_get_ipv6_addr_mode").return_value = model.addr_mode
        for ipv6 in Nic()._get_ipv6_addresses_info("eth1.1"):
            assert model.expect == ipv6

    @staticmethod
    def test_check_eth_link_state(mocker: MockFixture, model: CheckEthLinkState):
        mocker.patch.object(Nic, "_get_ip_cmd_json_info").return_value = model.link_info
        assert Nic()._check_eth_link_state("ip cmd") == model.expect

    @staticmethod
    def test_get_ipv4_addr_mode(mocker: MockFixture, model: GetIpv4AddrMode):
        assert Nic()._get_ipv4_addr_mode(model.dhcp) == model.expect

    @staticmethod
    def test_get_ipv6_addr_mode(mocker: MockFixture, model: GetIpv6AddrMode):
        assert Nic()._get_ipv6_addr_mode(model.info_str, model.dhcp) == model.expect

    @staticmethod
    def test_get_link_status(mocker: MockFixture, model: GetLinkStatus):
        mocker.patch.object(Nic, "_check_eth_link_state").return_value = model.link_state
        mocker.patch.object(ExecCmd, "exec_cmd_use_pipe_symbol").return_value = model.exec_cmd
        assert Nic()._get_link_status() == model.expect

    @staticmethod
    def test_get_name_server(mocker: MockFixture, model: GetNameServer):
        mocker.patch.object(Nic, "_get_file_content").return_value = model.content
        nic = Nic()
        nic._get_name_server()
        assert nic.NameServers == model.expect

    @staticmethod
    def test_get_net_traffic_info(mocker: MockFixture, model: GetNetTrafficInfo):
        mocker.patch.object(Nic, "_get_ip_cmd_json_info").return_value = model.traffic_info
        nic = Nic()
        nic._get_net_traffic_info("ip cmd")
        assert nic.RecvPackages == model.recv and nic.ErrorPackages == model.error and nic.DropPackages == \
               model.error and nic.SendPackages == model.send

    @staticmethod
    def test_patch_request(mocker: MockFixture, model: PatchRequest):
        mocker.patch.object(Nic, "nic_lock").locked.return_value = model.lock
        mocker.patch.object(Nic, "_check_request_data").side_effect = model.check_rq
        mocker.patch.object(Nic, "_modify_network_config").side_effect = model.modify
        mocker.patch.object(Nic, "_get_dynamic_info")
        assert Nic().patch_request({"ip": {}}, "GMAC1") == model.expect


class TestHandler(CommNetworkHandler):
    def rollback_ipv4(self):
        pass

    def effect_network(self):
        pass

    def modify_network(self):
        pass


class TestCommNetworkHandler:
    use_cases = {
        "test_restart_nginx": {
            "ok": (None, 1),
            "error": (None, 0)
        },
        "test_convert_mask_num": {
            "normal": (24, "255.255.255.0")
        },
        "test_check_ip_conflict": {
            "ok": (None, 1),
            "error": (NicOperateError, 0)
        },
        "test_update_old_tag_ini": {
            "load_err": (NicOperateError, ["test"], ["1.1.1.1"], Result(False), Result(False)),
            "write_err": (NicOperateError, ["test"], ["1.1.1.1"],
                          Result(True, data={"eth1": ["web"], "ip_eth1": ["1.1.1.1"]}), Result(False)),
            "ok": (None, ["test"], ["1.1.1.1"],
                   Result(True, data={"eth1": ["web"], "ip_eth1": ["1.1.1.1"]}), Result(True))
        },
        "test_get_cur_vlan_collection": {
            "vlan": ("eth1.2", "xxx\neth0: xxx\neth1: xxx\neth1.2: xxx\n")
        },
        "test_get_arping_cmd": {
            "normal": (
                (cmd_constants.OS_CMD_ARPING, "-f", "-c", "1", "-I", "eth1", "-s", "1.1.1.1", "1.1.1.1"),
                "1.1.1.1"
            )
        },
        "test_get_current_gateway": {
            "exec_failed": (None, [1, ""], NicOperateError, True),
            "ok": ("1.1.1.1", [0, "[{'gateway': '1.1.1.1'}]"], [[{"gateway": "1.1.1.1"}]], False),
            "json_failed": ("1.1.1.1", [0, "[{'gateway': '1.1.1.1'}]"], NicOperateError, True),
        },
        "test_check_ip_connection": {
            "ok": (0, False),
            "err": (1, True)
        },
        "test_network_check": {
            "normal": ("1.1.1.1", None, None)
        },
        "test_deal_ipv4_network": {
            "ok": (False,),
            "error": (True,)
        },
        "test_deal_with_ipv4": {
            "normal": (None,),
        }
    }

    @staticmethod
    def test_restart_nginx(mocker: MockFixture, model: RestartNginx):
        mocker.patch.object(ExecCmd, "exec_cmd", return_value=model.exec_cmd)
        assert TestHandler._restart_nginx() == model.expect

    @staticmethod
    def test_convert_mask_num(mocker: MockFixture, model: ConvertMaskNum):
        assert TestHandler._convert_mask_num(model.mask) == model.expect

    @staticmethod
    def test_check_ip_conflict(mocker: MockFixture, model: CheckIpConflict):
        mocker.patch.object(ExecCmd, "exec_cmd").return_value = model.exec_cmd
        if model.exception:
            with pytest.raises(model.exception):
                TestHandler._check_ip_conflict(("cmd",))
            return
        TestHandler._check_ip_conflict(("cmd",))

    @staticmethod
    def test_update_old_tag_ini(mocker: MockFixture, model: UpdateOldTagIni):
        mocker.patch.object(CommonMethods, "load_net_tag_ini").return_value = model.load_ini
        mocker.patch.object(CommonMethods, "write_net_tag_ini").return_value = model.write_ini
        if model.exception:
            with pytest.raises(model.exception):
                TestHandler({}, "eth1")._update_old_tag_ini(model.tag_list, model.ip_list)
            return
        TestHandler({}, "eth1")._update_old_tag_ini(model.tag_list, model.ip_list)

    @staticmethod
    def test_get_cur_vlan_collection(mocker: MockFixture, model: GetCurVlanCollection):
        mocker.patch("builtins.open", mock_open(read_data=model.content))
        for vlan in TestHandler({}, "eth1")._get_cur_vlan_collection():
            assert vlan == model.expect

    @staticmethod
    def test_get_arping_cmd(mocker: MockFixture, model: GetArpingCmd):
        mocker.patch.object(LocalIpChecker, "get_ip_address").return_value = model.addr
        assert TestHandler({}, "eth1")._get_arping_cmd(model.addr) == model.expect

    @staticmethod
    def test_get_current_gateway(mocker: MockFixture, model: GetCurrGateWay):
        mocker.patch.object(ExecCmd, "exec_cmd_use_pipe_symbol").return_value = model.exec_cmd
        mocker.patch("json.loads").side_effect = model.json
        if model.exception:
            with pytest.raises(NicOperateError):
                TestHandler({}, "eth1")._get_current_gateway()
        else:
            assert TestHandler({}, "eth1")._get_current_gateway() == model.expect

    @staticmethod
    def test_check_ip_connection(mocker: MockFixture, model: CheckIpConnection):
        mocker.patch.object(ExecCmd, "exec_cmd").return_value = model.exec_cmd
        if model.exception:
            with pytest.raises(NicOperateError):
                TestHandler({}, "eth1")._check_ip_connection(("ip", "cmd"))
        else:
            TestHandler({}, "eth1")._check_ip_connection(("ip", "cmd"))

    @staticmethod
    def test_network_check(mocker: MockFixture, model: NetworkCheck):
        mocker.patch.object(CommNetworkHandler, "_get_current_gateway").return_value = model.gateway
        mocker.patch.object(LocalIpChecker, "get_local_ips").side_effect = ["1.1.1.1"]
        mocker.patch.object(CommNetworkHandler, "_check_ip_connection")
        mocker.patch.object(CommNetworkHandler, "_check_ip_conflict")
        mocker.patch.object(LocalIpChecker, "get_ip_address")
        TestHandler({"IPv4Addresses": [{
            "ConnectTest": True,
            "RemoteTestIp": "1.1.1.1",
            "Address": "1.1.1.1"
        }]}, "eth1")._network_check()

    @staticmethod
    def test_deal_ipv4_network(mocker: MockFixture, model: DealIpv4Network):
        mocker.patch.object(CommNetworkHandler, "rollback_ipv4")
        if model.exception:
            mocker.patch.object(CommNetworkHandler, "_deal_with_ipv4").side_effect = NicOperateError
            with pytest.raises(NicOperateError):
                TestHandler({}, "eth1").deal_ipv4_network()
        else:
            mocker.patch.object(CommNetworkHandler, "_deal_with_ipv4")
            TestHandler({}, "eth1").deal_ipv4_network()

    @staticmethod
    def test_deal_with_ipv4(mocker: MockFixture, model: DealWithIpv4):
        mocker.patch.object(CommNetworkHandler, "modify_network")
        mocker.patch.object(CommNetworkHandler, "effect_network")
        mocker.patch.object(CommNetworkHandler, "_network_check")
        mocker.patch.object(NginxConfig, "set_nginx_config")
        mocker.patch.object(CommNetworkHandler, "_update_old_tag_ini")
        mocker.patch.object(CommNetworkHandler, "_update_database_and_tag_ini")
        mocker.patch.object(threading, "Timer")
        TestHandler({}, "eth1")._deal_with_ipv4()


class TestEulerNetworkHandler:
    use_cases = {
        "test_backup_netcfg_file": {
            "normal": (None,)
        },
        "test_modify_network": {
            "normal": (None,),
        },
        "test_effect_network": {
            "normal": (False, [0, 0, 0]),
            "vlan_failed": (True, [1, 0, 0]),
            "effect_failed": (True, [0, 0, 1]),
        },
        "test_rollback_ipv4": {
            "normal": (0,)
        },
        "test_write_ipv4_netcfg": {
            "normal": (None,)
        },
        "test__write_ipv4_file": {
            "normal": (None,)
        }
    }

    @staticmethod
    def test_backup_netcfg_file(mocker, model: EulerBackNetcfgFile):
        mocker.patch("shutil.copyfile")
        EulerNetworkHandler({}, "eth1")._backup_netcfg_file()

    @staticmethod
    def test_modify_network(mocker: MockFixture, model: EulerModifyNetwork):
        mocker.patch.object(EulerNetworkHandler, "_backup_netcfg_file")
        mocker.patch.object(EulerNetworkHandler, "_write_ipv4_netcfg")
        EulerNetworkHandler({}, "eth1").modify_network()

    @staticmethod
    def test_effect_network(mocker: MockFixture, model: EulerEffectNetwork):
        mocker.patch.object(EulerNetworkHandler, "_get_cur_vlan_collection").return_value = ["eth1.1"]
        mocker.patch.object(ExecCmd, "exec_cmd").side_effect = model.exec_cmd
        if model.exception:
            with pytest.raises(NicOperateError):
                EulerNetworkHandler({}, "eth1").effect_network()
        else:
            EulerNetworkHandler({}, "eth1").effect_network()

    @staticmethod
    def test_rollback_ipv4(mocker: MockFixture, model: EulerRollbackIpv4):
        mocker.patch.object(EulerNetworkHandler, "_get_cur_vlan_collection").return_value = ["eth1.1"]
        mocker.patch("os.remove")
        mocker.patch("shutil.copyfile")
        mocker.patch.object(ExecCmd, "exec_cmd").return_value = model.exec_cmd
        EulerNetworkHandler({}, "eth1").rollback_ipv4()

    @staticmethod
    def test_write_ipv4_netcfg(mocker: MockFixture, model: EulerWriteIpv4Netcfg):
        mocker.patch.object(EulerNetworkHandler, "_write_ipv4_file").return_value = ["eth1.1"]
        mocker.patch("os.remove")
        EulerNetworkHandler({"IPv4Addresses": [{
            "ConnectTest": True,
            "RemoteTestIp": "1.1.1.1",
            "Address": "1.1.1.1",
            "SubnetMask": "255.255.255.0",
            "VlanId": 1,
        }]}, "eth1")._write_ipv4_netcfg()

    @staticmethod
    def test__write_ipv4_file(mocker: MockFixture, model: EulerWriteIpv4Netcfg):
        mocker.patch("os.remove")
        mocker.patch("os.open")
        mocker.patch("os.fdopen")
        EulerNetworkHandler({}, "eth1")._write_ipv4_file({
            "SubnetMask": "255.255.255.0",
            "Gateway": "1.1.1.1",
            "Address": "1.1.1.1"
        }, "eth1")


class TestUbuntuNetworkHandler:
    use_cases = {
        "test_get_arping_cmd": {
            "normal": ((cmd_constants.OS_CMD_ARPING, "-C", "1", "-c", "1", "-i", "eth1", "-S", "1.1.1.1", "1.1.1.1"),)
        },
        "test_modify_netplan_data": {
            "normal": (None,)
        },
        "test_modify_network": {
            "ok": (False, [{"network": {"ethernets": {}}}], None),
            "load_failed": (True, NicOperateError, None),
            "write_failed": (True, [{"network": {"ethernets": {}}}], NicOperateError),
        },
        "test_effect_network": {
            "normal": (False, [0, 0]),
            "vlan_failed": (True, [1, 0]),
            "effect_failed": (True, [0, 1]),
        },
        "test_rollback_ipv4": {
            "normal": (0,)
        },
    }

    @staticmethod
    def test_get_arping_cmd(mocker: MockFixture, model: UbuntuGetArpingCmd):
        mocker.patch.object(LocalIpChecker, "get_ip_address").return_value = "1.1.1.1"
        assert UbuntuNetworkHandler({}, "eth1")._get_arping_cmd("1.1.1.1") == model.expect

    @staticmethod
    def test_modify_netplan_data(mocker: MockFixture, model: UbuntuModifyNetplanData):
        UbuntuNetworkHandler({"IPv4Addresses": [{
            "SubnetMask": "255.255.255.0",
            "Address": "1.1.1.1",
            "VlanId": 1,
            "Gateway": "1.1.1.1"
        }]}, "eth1")._modify_netplan_data({"network": {"ethernets": {}}})

    @staticmethod
    def test_modify_network(mocker: MockFixture, model: UbuntuModifyNetwork):
        mocker.patch("shutil.copyfile")
        mocker.patch.object(YamlMethod, "load_yaml_info").side_effect = model.load_failed
        mocker.patch.object(YamlMethod, "dumps_yaml_file").side_effect = model.dumps_failed
        mocker.patch.object(UbuntuNetworkHandler, "_modify_netplan_data")
        if model.exception:
            with pytest.raises(NicOperateError):
                UbuntuNetworkHandler({}, "eth1").modify_network()
        else:
            UbuntuNetworkHandler({}, "eth1").modify_network()

    @staticmethod
    def test_effect_network(mocker: MockFixture, model: UbuntuEffectNetwork):
        mocker.patch.object(UbuntuNetworkHandler, "_get_cur_vlan_collection").return_value = ["eth1.1"]
        mocker.patch.object(ExecCmd, "exec_cmd").side_effect = model.exec_cmd
        if model.exception:
            with pytest.raises(NicOperateError):
                UbuntuNetworkHandler({}, "eth1").effect_network()
        else:
            UbuntuNetworkHandler({}, "eth1").effect_network()

    @staticmethod
    def test_rollback_ipv4(mocker: MockFixture, model: UbuntuRollbackIpv4):
        mocker.patch.object(EulerNetworkHandler, "_get_cur_vlan_collection").return_value = ["eth1.1"]
        mocker.patch("os.remove")
        mocker.patch("shutil.copyfile")
        mocker.patch.object(ExecCmd, "exec_cmd").return_value = model.exec_cmd
        EulerNetworkHandler({}, "eth1").rollback_ipv4()


class TestOpenEulerNetworkHandler:
    use_cases = {
        "test_modify_network": {
            "ok": (False,),
        },
        "test_effect_network": {
            "normal": (False,),
            "failed": (True,),
        },
        "test_rollback_ipv4": {
            "normal": (0,)
        },
        "test_backup_netcfg_file": {
            "normal": (None,)
        },
        "test_restart_interface": {
            "ok": (False, [0, 0]),
            "error": (True, [0, 1])
        },
        "test_write_ipv4_netcfg": {
            "normal": (None,)
        },
        "test_modify_ipv4_netcfg": {
            "ok": (False, [0, 0, 0]),
            "delete_vlan_failed": (True, [1, 0, 0]),
            "modify_ip_failed": (True, [0, 1, 0]),
            "modify_vlan_failed": (True, [0, 1, 1]),
        }
    }

    @staticmethod
    def test_modify_network(mocker: MockFixture, model: OpenEulerModifyNetwork):
        mocker.patch.object(OpenEulerNetHandler, "_backup_netcfg_file")
        mocker.patch.object(OpenEulerNetHandler, "_modify_ipv4_netcfg")
        OpenEulerNetHandler({}, "eth1").modify_network()

    @staticmethod
    def test_effect_network(mocker: MockFixture, model: OpenEulerEffectNetwork):
        if model.exception:
            mocker.patch.object(OpenEulerNetHandler, "_restart_interface").side_effect = NicOperateError
            with pytest.raises(NicOperateError):
                OpenEulerNetHandler({}, "eth1").effect_network()
        else:
            mocker.patch.object(OpenEulerNetHandler, "_restart_interface")
            OpenEulerNetHandler({}, "eth1").effect_network()

    @staticmethod
    def test_rollback_ipv4(mocker: MockFixture, model: OpenEulerRollbackIpv4):
        mocker.patch.object(OpenEulerNetHandler, "_get_cur_vlan_collection").return_value = ["eth1.1"]
        mocker.patch.object(OpenEulerNetHandler, "_restart_interface")
        mocker.patch("shutil.copyfile")
        mocker.patch.object(ExecCmd, "exec_cmd").return_value = model.exec_cmd
        OpenEulerNetHandler({}, "eth1").rollback_ipv4()

    @staticmethod
    def test_restart_interface(mocker: MockFixture, model: OpenEulerRestartInterface):
        mocker.patch.object(ExecCmd, "exec_cmd").side_effect = model.exec_cmd
        if model.exception:
            with pytest.raises(NicOperateError):
                OpenEulerNetHandler({}, "eth1")._restart_interface(["eth1"], True)
        else:
            OpenEulerNetHandler({}, "eth1")._restart_interface(["eth1"], True)

    @staticmethod
    def test_backup_netcfg_file(mocker, model: OpenEulerBackNetcfgFile):
        mocker.patch("shutil.copyfile")
        OpenEulerNetHandler({}, "eth1")._backup_netcfg_file()

    @staticmethod
    def test_modify_ipv4_netcfg(mocker, model: OpenEulerModifyIpv4Netcfg):
        mocker.patch.object(OpenEulerNetHandler, "_get_cur_vlan_collection").return_value = ["eth1.1"]
        mocker.patch.object(ExecCmd, "exec_cmd").side_effect = model.exec_cmd
        if model.exception:
            with pytest.raises(NicOperateError):
                OpenEulerNetHandler({"IPv4Addresses": [{
                    "SubnetMask": "255.255.255.0",
                    "Address": "1.1.1.1",
                    "VlanId": 1,
                    "Gateway": "1.1.1.1"
                }]}, "eth1")._modify_ipv4_netcfg()
        else:
            OpenEulerNetHandler({"IPv4Addresses": [{
                "SubnetMask": "255.255.255.0",
                "Address": "1.1.1.1",
                "VlanId": 1,
                "Gateway": "1.1.1.1"
            }]}, "eth1")._modify_ipv4_netcfg()
