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
from configparser import RawConfigParser

from pytest_mock import MockerFixture

from common.file_utils import FileUtils
from common.utils.exec_cmd import ExecCmd
from common.utils.policy_based_route import AddRouteCfg, AddRuleCfg, SaveRouteCfg, SaveRuleCfg, PersistRouteCfg, \
    DeleteDefRouteCfg, SaveGatewayAndRouteTableID, DeleteEulerGateway, DeleteOpenEulerGateway, DeleteUbuntuGateway
from common.yaml.yaml_methods import YamlMethod


class TestAddRouteCfg:

    @staticmethod
    def test_config(mocker: MockerFixture):
        mocker.patch.object(ExecCmd, "exec_cmd_get_output").return_value = (0, "success")
        AddRouteCfg("0", "127.0.0.1", "127.0.0.0").config()

    @staticmethod
    def test_rollback(mocker: MockerFixture):
        mocker.patch.object(ExecCmd, "exec_cmd_get_output").return_value = (0, "127.0.0.1 success")
        AddRouteCfg("0", "127.0.0.1", "127.0.0.0").rollback()


class TestAddRuleCfg:

    @staticmethod
    def test_config(mocker: MockerFixture):
        mocker.patch.object(ExecCmd, "exec_cmd_get_output").return_value = (0, "success")
        AddRuleCfg("1", "127.0.0.1").config()

    @staticmethod
    def test_rollback(mocker: MockerFixture):
        mocker.patch.object(ExecCmd, "exec_cmd_get_output").return_value = (0, "127.0.0.1 success")
        AddRuleCfg("1", "127.0.0.1").rollback()


class TestSaveRouteCfg:

    @staticmethod
    def test_config(mocker: MockerFixture):
        mocker.patch.object(ExecCmd, "exec_cmd_get_output").return_value = (0, "success")
        mocker.patch.object(ExecCmd, "exec_cmd_use_pipe_symbol").return_value = (0, "2 success")
        mocker.patch.object(FileUtils, "check_script_file_valid").return_value = True
        SaveRouteCfg("2").config()

    @staticmethod
    def test_rollback(mocker: MockerFixture):
        mocker.patch.object(ExecCmd, "exec_cmd_get_output").return_value = (0, "127.0.0.1 success")
        mocker.patch.object(ExecCmd, "exec_cmd_use_pipe_symbol").return_value = (0, "2 success")
        mocker.patch.object(FileUtils, "check_script_file_valid").return_value = True
        SaveRouteCfg("2").rollback()


class TestSaveRuleCfg:

    @staticmethod
    def test_config(mocker: MockerFixture):
        mocker.patch.object(ExecCmd, "exec_cmd_get_output").return_value = (0, "success")
        mocker.patch.object(ExecCmd, "exec_cmd_use_pipe_symbol").return_value = (0, "3 success")
        mocker.patch.object(FileUtils, "check_script_file_valid").return_value = True
        SaveRuleCfg("3").config()

    @staticmethod
    def test_rollback(mocker: MockerFixture):
        mocker.patch.object(ExecCmd, "exec_cmd_get_output").return_value = (0, "127.0.0.1 success")
        mocker.patch.object(ExecCmd, "exec_cmd_use_pipe_symbol").return_value = (0, "3 success")
        mocker.patch.object(FileUtils, "check_script_file_valid").return_value = True
        SaveRuleCfg("3").rollback()


class TestPersistRouteCfg:

    @staticmethod
    def test_config(mocker: MockerFixture):
        mocker.patch.object(ExecCmd, "exec_cmd_get_output").return_value = (0, "success")
        mocker.patch.object(ExecCmd, "exec_cmd_use_pipe_symbol").return_value = (0, "4 success")
        mocker.patch.object(FileUtils, "check_script_file_valid").return_value = True
        PersistRouteCfg("4").config()

    @staticmethod
    def test_rollback(mocker: MockerFixture):
        mocker.patch.object(ExecCmd, "exec_cmd_get_output").return_value = (0, "success")
        mocker.patch.object(ExecCmd, "exec_cmd_use_pipe_symbol").return_value = (0, "4 success")
        mocker.patch.object(FileUtils, "check_script_file_valid").return_value = True
        PersistRouteCfg("4").rollback()


class TestDeleteDefRouteCfg:

    @staticmethod
    def test_config(mocker: MockerFixture):
        mocker.patch.object(ExecCmd, "exec_cmd_get_output").return_value = (0, "default dev")
        DeleteDefRouteCfg("5", "127.0.0.1").config()

    @staticmethod
    def test_rollback(mocker: MockerFixture):
        mocker.patch.object(ExecCmd, "exec_cmd_get_output").return_value = (0, "default")
        DeleteDefRouteCfg("5", "127.0.0.1").rollback()


class TestSaveGatewayAndRouteTableID:

    @staticmethod
    def test_config(mocker: MockerFixture):
        mocker.patch.object(FileUtils, "check_section_exist").return_value = True
        mocker.patch.object(FileUtils, "get_option_list").return_value = {
            "route_table_id": "6", "eth_gw_ip": "127.0.0.1", "eth_name": "eth1",
        }
        mocker.patch.object(FileUtils, "modify_one_option")
        SaveGatewayAndRouteTableID("6", "127.0.0.1", "eth0").config()

    @staticmethod
    def test_rollback(mocker: MockerFixture):
        mocker.patch.object(ExecCmd, "exec_cmd_get_output").return_value = (0, "default")
        mocker.patch.object(FileUtils, "get_option_list").return_value = {
            "route_table_id": "6", "eth_gw_ip": "127.0.0.1", "eth_name": "eth0",
        }
        mocker.patch.object(FileUtils, "get_config_parser")
        mocker.patch.object(RawConfigParser, "remove_section")
        mocker.patch.object(FileUtils, "write_file_with_lock")
        SaveGatewayAndRouteTableID("6", "127.0.0.1", "eth0").rollback()


class TestDeleteEulerGateway:

    @staticmethod
    def test_config(mocker: MockerFixture):
        mocker.patch.object(DeleteEulerGateway, "_get_gateway_by_normal").return_value = "127.0.0.1"
        mocker.patch.object(DeleteEulerGateway, "_set_gateway_by_normal").return_value = True
        DeleteEulerGateway("7", "127.0.0.1", "eth0").config()

    @staticmethod
    def test_rollback(mocker: MockerFixture):
        mocker.patch.object(DeleteEulerGateway, "_get_gateway_by_normal").return_value = ""
        mocker.patch.object(DeleteEulerGateway, "_set_gateway_by_normal").return_value = True
        DeleteEulerGateway("7", "127.0.0.1", "eth0").rollback()

    @staticmethod
    def test_get_gateway_by_normal():
        tmp_cfg = "/tmp/euler_ifcfg"
        with os.fdopen(os.open(tmp_cfg, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600), "w") as fd:
            fd.write("GATEWAY=127.0.0.1\n")

        ret = DeleteEulerGateway("7", "127.0.0.1", "eth0")._get_gateway_by_normal(tmp_cfg)
        assert ret == "127.0.0.1"
        os.unlink(tmp_cfg)

    @staticmethod
    def test_set_gateway_by_normal():
        tmp_cfg = "/tmp/euler_ifcfg"
        with os.fdopen(os.open(tmp_cfg, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600), "w") as fd:
            fd.write("")

        assert DeleteEulerGateway("7", "127.0.0.1", "eth0")._set_gateway_by_normal(tmp_cfg, "127.0.0.1")
        with open(tmp_cfg) as fd_in:
            value = fd_in.read()

        assert value == "GATEWAY=127.0.0.1\n"
        os.unlink(tmp_cfg)


class TestDeleteOpenEulerGateway:

    @staticmethod
    def test_config(mocker: MockerFixture):
        mocker.patch.object(DeleteOpenEulerGateway, "_get_gateway_by_normal").return_value = "127.0.0.1"
        mocker.patch.object(DeleteOpenEulerGateway, "_set_gateway_by_normal").return_value = True
        DeleteOpenEulerGateway("7", "127.0.0.1", "eth0").config()

    @staticmethod
    def test_rollback(mocker: MockerFixture):
        mocker.patch.object(DeleteOpenEulerGateway, "_get_gateway_by_normal").return_value = ""
        mocker.patch.object(DeleteOpenEulerGateway, "_set_gateway_by_normal").return_value = True
        DeleteOpenEulerGateway("7", "127.0.0.1", "eth0").rollback()


class TestDeleteUbuntuGateway:
    TMP_YAML = "/tmp/ubuntu_ifcfg.yaml"

    def setup_method(self):
        with os.fdopen(os.open(self.TMP_YAML, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600), "w") as fd:
            fd.write("")

    def teardown_method(self):
        os.unlink(self.TMP_YAML)

    def test_config(self, mocker: MockerFixture):
        content = {
            'network': {
                'ethernets': {
                    'eth0': {
                        'gateway4': '127.0.0.1',
                        'routes': [{
                            'to': 'default',
                        }],
                    },
                },
            },
        }
        YamlMethod.dumps_yaml_file(content, self.TMP_YAML)
        mocker.patch.object(DeleteUbuntuGateway, "NETCFG_YAML", self.TMP_YAML)
        DeleteUbuntuGateway("8", "127.0.0.1", "eth0").config()
        assert DeleteUbuntuGateway.NETCFG_YAML == self.TMP_YAML
        with open(self.TMP_YAML) as fd_in:
            value = fd_in.read()
        assert "gateway4" not in value

    def test_rollback(self, mocker: MockerFixture):
        content = {
            'network': {
                'ethernets': {
                    'eth0': {
                        'gateway4': '',
                        'routes': [{
                            'to': 'default',
                            'via': '',
                        }],
                    },
                },
            },
        }
        YamlMethod.dumps_yaml_file(content, self.TMP_YAML)
        mocker.patch.object(DeleteUbuntuGateway, "NETCFG_YAML", self.TMP_YAML)
        DeleteUbuntuGateway("8", "127.0.0.1", "eth0").rollback()
        with open(self.TMP_YAML) as fd_in:
            value = fd_in.read()
        assert "127.0.0.1" in value
