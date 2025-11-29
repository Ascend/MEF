from collections import namedtuple
from pathlib import Path

import pytest
from pytest_mock import MockFixture

from common.init_cmd import cmd_constants
from common.utils.exec_cmd import ExecCmd
from common.utils.result_base import Result
from lib.Linux.systems.nic.config_web_ip import NginxConfig, WebIpError, para_check, main
from lib.Linux.systems.nic.models import NetConfig
from lib.Linux.systems.nic.nic_mgr import NetConfigMgr
from common.common_methods import CommonMethods
from common.yaml.yaml_methods import YamlMethod
from ut_utils.mock_utils import mock_read_data

GetNginxListenIpv4 = namedtuple("GetNginxListenIpv4", "expect, exception, query, exec_cmd")
ReplaceListenIpString = namedtuple("ReplaceListenIpString", "expect, ip_li, line")
GetNginxListenIpv6 = namedtuple("GetNginxListenIpv6", "expect, exec_cmd, check_ip")
UbuntuGetVlanAddr = namedtuple("UbuntuGetVlanAddr", "expect, data")
UbuntuGetNormalAddr = namedtuple("UbuntuGetNormalAddr", "expect, data")
UpdateOldTagIni = namedtuple("UpdateOldTagIni", "exception, write")
GetIpv4Addr = namedtuple("GetIpv4Addr", "expect, content")
UbuntuGetIpv4Addr = namedtuple("UbuntuGetIpv4Addr", "expect, yaml, addr")
GetLinkState = namedtuple("GetLinkState", "expect, exec_cmd, exception")
OpenEulerGetAllIpv4Addr = namedtuple("OpenEulerGetAllIpv4Addr", "expect, ip")
GetIpv4InfoFromFile = namedtuple("GetIpv4InfoFromFile", "expect, os_name, exception")
EulerGetAllIpv4Addr = namedtuple("EulerGetAllIpv4Addr", "expect")
UpdateDatabase = namedtuple("UpdateDatabase", "link, addr, operate")
SetNginxConfig = namedtuple("SetNginxConfig", "listen_ip, ipv6")
WriteNginxConfig = namedtuple("WriteNginxConfig", "content")
ParaCheck = namedtuple("ParaCheck", "exception, argv, path")
Main = namedtuple("Main", "expect")


class TestNic:
    use_cases = {
        "test_get_nginx_listen_ipv4": {
            "static": (("eth1", "1.1.1.1"), False, [NetConfig(name="eth1", tag="web", ipv4="1.1.1.1")], [0, ""]),
            "dhcp": (("eth1", "dhcp"), False, [NetConfig(name="eth1", tag="web", ipv4="dhcp")], [0, "dhcp/24"]),
            "dhcp_error": ((), True, [NetConfig(name="eth1", tag="web", ipv4="dhcp")], [1, ""]),
            "no_ip": ((), True, [], [1, ""]),
        },
        "test_replace_listen_ip_string": {
            "ok": ("listen 1.1.1.1:443 ssl;", ["1.1.1.1"], "listen :443 ssl;"),
            "not_line": ("xxxx :443 ssl;", ["1.1.1.1"], "xxxx :443 ssl;"),
            "not_ip": ("", [], "xxxx :443 ssl;"),
        },
        "test_get_nginx_listen_ipv6": {
            "exec_failed": ("", [1, ""], True),
            "check_failed": ("", [1, ""], False),
            "re_failed": ("", [0, "inet ff:ff:ff:ff"], False),
            "ok": ("[ff:ff:ff:ff%eth1]", [0, "inet6 ff:ff:ff:ff"], True)
        },
        "test_ubuntu_get_vlan_addr": {
            "no_vlan": ([], {"network": {}}),
            "vlan": (["1.1.1.1"], {"network": {"vlans": {"eth1.2": {"addresses": ["1.1.1.1/24"]}}}})
        },
        "test_ubuntu_get_normal_addr": {
            "no_ip": ([], {"network": {"ethernets": {}}}),
            "dhcp": (["dhcp"], {"network": {"ethernets": {"eth1": {"dhcp4": True}}}}),
            "normal": (["1.1.1.1"], {"network": {"ethernets": {"eth1": {"addresses": ["1.1.1.1/24"]}}}}),
        },
        "test_update_old_tag_ini": {
            "ok": (False, Result(True)),
            "error": (True, Result(False))
        },
        "test__get_ipv4_addr": {
            "dhcp": ("dhcp", 'xxx\nBOOTPROTO="dhcp"\n'),
            "static": ("192.168.3.101", 'xxx\nIPADDR="192.168.3.101"\n')
        },
        "test_ubuntu_get_ipv4_addr": {
            "yaml_failed": ([], True, []),
            "ok": (["1.1.1.1", "1.1.1.1"], False, ["1.1.1.1"])
        },
        "test_get_link_state": {
            "all_ok": (["eth0", "eth1"], [[0, "Link yes"], [0, "Link yes"]], False),
            "error": ([], [[1, "no"], [1, "no"]], True),
            "one_ok": (["eth1"], [[1, "no"], [0, "yes"]], False),
        },
        "test_open_euler_get_all_ipv4_addr": {
            "ok": (["1.1.1.1"], ["1.1.1.1"])
        },
        "test_get_ipv4_info_from_file": {
            "not_support": ("", "test_os", WebIpError),
            "euler": ("1.1.1.1", "EulerOS", None),
            "open_euler": ("1.1.1.1", "openEuler", None),
            "ubuntu": ("1.1.1.1", "Ubuntu", None)
        },
        "test_euler_get_all_ipv4_addr": {
            "normal": (["1.1.1.1", "1.1.1.1"], )
        },
        "test_update_database": {
            "Install": (["eth1"], ["1.1.1.1"], "Install"),
            "Start": (["eth1"], ["1.1.1.1"], "Start"),
            "Start_dhcp": (["eth1"], ["dhcp"], "Start"),
        },
        "test_set_nginx_config": {
            "normal": (("eth1", "1.1.1.1"), "ff:ff:ff:ff")
        },
        "test_write_nginx_config": {
            "normal": ("listen :443 ssl;\n", )
        },
        "test_para_check": {
            "normal": (False, ("config_web_ip.py", "nginx.conf"), True),
            "not_2": (True, ("config_web_ip.py", "nginx.conf", "Start"), True),
            "not_exist": (True, ("config_web_ip.py", "nginx.conf"), False),
        },
        "test_main": {
            "normal": (0, )
        }
    }

    @staticmethod
    def test_get_nginx_listen_ipv4(mocker: MockFixture, model: GetNginxListenIpv4):
        mocker.patch.object(NetConfigMgr, "query_info_with_condition").return_value = model.query
        mocker.patch.object(ExecCmd, "exec_cmd_use_pipe_symbol").return_value = model.exec_cmd
        if model.exception:
            with pytest.raises(WebIpError):
                assert model.expect == NginxConfig().get_nginx_listen_ipv4()
        else:
            assert model.expect == NginxConfig().get_nginx_listen_ipv4()

    @staticmethod
    def test_replace_listen_ip_string(mocker: MockFixture, model: ReplaceListenIpString):
        assert NginxConfig()._replace_listen_ip_string(model.ip_li, model.line) == model.expect

    @staticmethod
    def test_get_nginx_listen_ipv6(mocker: MockFixture, model: GetNginxListenIpv6):
        mocker.patch.object(ExecCmd, "exec_cmd_use_pipe_symbol").return_value = model.exec_cmd
        if not model.check_ip:
            mocker.patch("socket.inet_pton").side_effect = Exception
        else:
            mocker.patch("socket.inet_pton")
        assert NginxConfig()._get_nginx_listen_ipv6("eth1") == model.expect

    @staticmethod
    def test_ubuntu_get_vlan_addr(mocker: MockFixture, model: UbuntuGetVlanAddr):
        assert NginxConfig()._ubuntu_get_vlan_addr(model.data) == model.expect

    @staticmethod
    def test_ubuntu_get_normal_addr(mocker: MockFixture, model: UbuntuGetNormalAddr):
        assert NginxConfig()._ubuntu_get_normal_addr(model.data, "eth1") == model.expect

    @staticmethod
    def test_update_old_tag_ini(mocker: MockFixture, model: UpdateOldTagIni):
        mocker.patch.object(CommonMethods, "write_net_tag_ini").return_value = model.write
        if model.exception:
            with pytest.raises(WebIpError):
                NginxConfig()._update_old_tag_ini("eth1", "1.1.1.1")
        else:
            NginxConfig()._update_old_tag_ini("eth1", "1.1.1.1")

    @staticmethod
    def test__get_ipv4_addr(mocker: MockFixture, model: GetIpv4Addr):
        mocker.patch("socket.inet_pton")
        mock_read_data(mocker, read_data=model.content)
        NginxConfig._get_ipv4_addr(Path("ifcfg-eth1"))

    @staticmethod
    def test_ubuntu_get_ipv4_addr(mocker: MockFixture, model: UbuntuGetIpv4Addr):
        if model.yaml:
            mocker.patch.object(YamlMethod, "load_yaml_info").side_effect = Exception
        else:
            mocker.patch.object(YamlMethod, "load_yaml_info").return_value = {}
        mocker.patch.object(NginxConfig, "_ubuntu_get_normal_addr").return_value = model.addr
        mocker.patch.object(NginxConfig, "_ubuntu_get_vlan_addr").return_value = model.addr
        assert model.expect == NginxConfig()._ubuntu_get_ipv4_addr("eth1")

    @staticmethod
    def test_get_link_state(mocker: MockFixture, model: GetLinkState):
        mocker.patch.object(ExecCmd, "exec_cmd_use_pipe_symbol").side_effect = model.exec_cmd
        if model.exception:
            with pytest.raises(WebIpError):
                assert model.expect == NginxConfig()._get_link_state()
        else:
            assert model.expect == NginxConfig()._get_link_state()

    @staticmethod
    def test_open_euler_get_all_ipv4_addr(mocker: MockFixture, model: OpenEulerGetAllIpv4Addr):
        mocker.patch.object(NginxConfig, "_euler_get_all_ipv4_addr").return_value = model.ip
        assert model.expect == NginxConfig()._open_euler_get_all_ipv4_addr("eth1")

    @staticmethod
    def test_get_ipv4_info_from_file(mocker: MockFixture, model: GetIpv4InfoFromFile):
        mocker.patch.object(NginxConfig, "_euler_get_all_ipv4_addr").return_value = "1.1.1.1"
        mocker.patch.object(NginxConfig, "_ubuntu_get_ipv4_addr").return_value = "1.1.1.1"
        mocker.patch.object(NginxConfig, "_open_euler_get_all_ipv4_addr").return_value = "1.1.1.1"
        old_os_name = cmd_constants.OS_NAME
        cmd_constants.OS_NAME = model.os_name
        if model.exception:
            with pytest.raises(model.exception):
                NginxConfig()._get_ipv4_info_from_file("eth1")
        else:
            assert model.expect == NginxConfig()._get_ipv4_info_from_file("eth1")
        cmd_constants.OS_NAME = old_os_name

    @staticmethod
    def test_euler_get_all_ipv4_addr(mocker: MockFixture, model: EulerGetAllIpv4Addr):
        mocker.patch.object(Path, "glob").return_value = (f"ifcfg-eth{i}" for i in range(0, 2))
        mocker.patch.object(NginxConfig, "_get_ipv4_addr").return_value = ["1.1.1.1"]
        assert model.expect == NginxConfig()._euler_get_all_ipv4_addr("eth1")

    @staticmethod
    def test_update_database(mocker: MockFixture, model: UpdateDatabase):
        mocker.patch.object(NginxConfig, "_get_link_state").return_value = model.link
        mocker.patch.object(NginxConfig, "_get_ipv4_info_from_file").return_value = model.addr
        mocker.patch.object(NetConfigMgr, "save_net_config")
        mocker.patch.object(NetConfigMgr, "delete_specific_eth_config")
        mocker.patch.object(NetConfigMgr, "query_info_with_condition").side_effect = [[NetConfig(ipv4="1.1.1.1")]]
        mocker.patch.object(NginxConfig, "_update_old_tag_ini")
        NginxConfig().update_database()

    @staticmethod
    def test_set_nginx_config(mocker: MockFixture, model: SetNginxConfig):
        mocker.patch.object(NginxConfig, "get_nginx_listen_ipv4").return_value = model.listen_ip
        mocker.patch.object(NginxConfig, "_get_nginx_listen_ipv6").return_value = model.ipv6
        mocker.patch.object(NginxConfig, "_write_nginx_config")
        mocker.patch.object(NginxConfig, "update_database")
        NginxConfig().set_nginx_config("nginx.conf")

    @staticmethod
    def test_write_nginx_config(mocker: MockFixture, model: WriteNginxConfig):
        mock_read_data(mocker, read_data=model.content)
        mocker.patch("os.fdopen")
        mocker.patch("os.open")
        NginxConfig()._write_nginx_config(["1.1.1.1"], "nginx.conf")

    @staticmethod
    def test_para_check(mocker: MockFixture, model: ParaCheck):
        mocker.patch("os.path.exists").return_value = model.path
        if not model.exception:
            para_check(model.argv)
        else:
            with pytest.raises(WebIpError):
                para_check(model.argv)

    @staticmethod
    def test_main(mocker: MockFixture, model: Main):
        mocker.patch("lib.Linux.systems.nic.config_web_ip.para_check").return_value = "nginx.conf"
        mocker.patch.object(NginxConfig, "update_database")
        mocker.patch.object(NginxConfig, "set_nginx_config")
        assert model.expect == main(["config_web_ip.py", "nginx.conf"])
