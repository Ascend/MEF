# coding: utf-8
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
from io import StringIO
from unittest.mock import patch

import pytest

from common.utils.ability_policy import OpEnumBase, HighRiskOpPolicyDto, SwitchVal, CmdDisableAllOpt, Cmd, \
    HighRiskOpPolicyCli


class OpTest(OpEnumBase):
    """
    使用时需要定义高危特性枚举类
    """
    OP1 = "op1"
    OP2 = "op2"
    OP3 = "op3"


class TestHighRiskPolicyDto:
    """
    测试作为校验策略时的使用场景
    """
    default_json_policy = {
        "disable_all": True,
        "allow": {
            "op1": False,
            "op2": False,
            "op3": False,
        }
    }

    @staticmethod
    def test_dump_to_json_success():
        dto = HighRiskOpPolicyDto(OpTest, disable_all=SwitchVal.OFF, allow=None)
        json_conf = dto.dump_to_json()
        expected = {
            "disable_all": False,
            "allow": {
                "op1": False,
                "op2": False,
                "op3": False,
            }
        }
        assert json_conf == expected

    def test_load_from_json_success(self):
        dto = HighRiskOpPolicyDto.load_from_json(OpTest, self.default_json_policy)
        assert dto.disable_all is True
        assert dto.allow[OpTest.OP1] == SwitchVal.OFF
        assert dto.allow[OpTest.OP2] == SwitchVal.OFF
        assert dto.allow[OpTest.OP3] == SwitchVal.OFF

    def test_check_allow_disable_all_success(self):
        json_conf = copy.deepcopy(self.default_json_policy)
        json_conf["disable_all"] = True
        dto = HighRiskOpPolicyDto.load_from_json(OpTest, json_conf)
        assert not dto.check_allow(OpTest.OP1)
        assert not dto.check_allow(OpTest.OP2)
        assert not dto.check_allow(OpTest.OP3)

    def test_check_allow_srv_enable_success(self):
        json_conf = copy.deepcopy(self.default_json_policy)
        json_conf["disable_all"] = False
        try:
            json_conf["allow"]["op1"] = True
        except KeyError:
            assert False
        dto = HighRiskOpPolicyDto.load_from_json(OpTest, json_conf)
        assert dto.check_allow(OpTest.OP1)
        assert not dto.check_allow(OpTest.OP2)
        assert not dto.check_allow(OpTest.OP3)

    def test_check_allow_srv_disable_success(self):
        json_conf = copy.deepcopy(self.default_json_policy)
        json_conf["disable_all"] = False
        try:
            json_conf["allow"]["op2"] = False
        except KeyError:
            assert False
        dto = HighRiskOpPolicyDto.load_from_json(OpTest, json_conf)
        assert not dto.check_allow(OpTest.OP2)


class TestHighRiskOpPolicyCli:
    """
    测试作为命令行使用的场景
    """

    @classmethod
    def setup_class(cls):
        # 命令行参数定义
        cmd_conf = {
            "description": "config om high risk operations policy.",
            "sub_cmds": [
                {
                    'cmd': Cmd.DISABLE_ALL.value,
                    "help": "disable all ops.",
                    "opts": [
                        {
                            "name": CmdDisableAllOpt.ON.value,
                            "default": SwitchVal.OFF,
                            "help": "disable all high risk operations.",
                        }
                    ]
                },
                {
                    'cmd': Cmd.ALLOW.value,
                    "help": "allow ops.",
                    "opts": [
                        {
                            "name": OpTest.OP1.value,
                            "default": SwitchVal.OFF,
                            "help": "to enable op1 feature.",
                        },
                        {
                            "name": OpTest.OP2.value,
                            "default": SwitchVal.OFF,
                            "help": "to enable op2 feature.",
                        },
                        {
                            "name": OpTest.OP3.value,
                            "default": SwitchVal.OFF,
                            "help": "to enable op3 feature.",
                        },
                    ]
                },
            ]
        }
        cls.cli = HighRiskOpPolicyCli(cmd_conf, OpTest)
        cls.parser = cls.cli.gen_cli_parser()

    def test_parse_args_disable_all_success(self):
        cmd = "disable_all --on"
        dto = self.cli.convert_to_dto(self.parser.parse_args(cmd.split(" ")))
        assert dto.disable_all
        assert dto.allow.get(OpTest.OP1) == SwitchVal.OFF
        assert dto.allow.get(OpTest.OP2) == SwitchVal.OFF
        assert dto.allow.get(OpTest.OP3) == SwitchVal.OFF

    def test_parse_args_allow_one_success(self):
        cmd = "allow --op1"
        dto = self.cli.convert_to_dto(self.parser.parse_args(cmd.split(" ")))
        assert not dto.disable_all
        assert dto.allow[OpTest.OP1]
        assert not dto.allow[OpTest.OP2]
        assert not dto.allow[OpTest.OP3]

    def test_parse_args_allow_part_success(self):
        cmd = "allow --op3 --op1"
        dto = self.cli.convert_to_dto(self.parser.parse_args(cmd.split(" ")))
        assert not dto.disable_all
        assert dto.allow[OpTest.OP1]
        assert not dto.allow[OpTest.OP2]
        assert dto.allow[OpTest.OP3]

    def test_parse_args_allow_all_success(self):
        cmd = "allow --op1 --op3 --op2"
        dto = self.cli.convert_to_dto(self.parser.parse_args(cmd.split(" ")))
        assert not dto.disable_all
        assert dto.allow[OpTest.OP1]
        assert dto.allow[OpTest.OP2]
        assert dto.allow[OpTest.OP3]

    @patch('sys.stderr', new_callable=StringIO)
    def test_parse_args_default_help_success(self, mock_stderr):
        cmd = ""
        expected = "[-h] {disable_all,allow}"
        with pytest.raises(SystemExit):
            self.cli.convert_to_dto(self.parser.parse_args(cmd.split(" ")))
        assert expected in mock_stderr.getvalue()

    @patch('sys.stdout', new_callable=StringIO)
    def test_parse_args_with_opt_help_success(self, mock_stdout):
        cmds = ["-h", "--help"]
        expected = "[-h] {disable_all,allow}"
        for cmd in cmds:
            with pytest.raises(SystemExit):
                self.cli.convert_to_dto(self.parser.parse_args(cmd.split(" ")))
            assert expected in mock_stdout.getvalue()

    @patch('sys.stdout', new_callable=StringIO)
    def test_parse_args_disable_with_opt_help_success(self, mock_stdout):
        cmds = ["disable_all -h", "disable_all --help"]
        expected = "disable_all [-h] [--on]"
        for cmd in cmds:
            with pytest.raises(SystemExit):
                self.cli.convert_to_dto(self.parser.parse_args(cmd.split(" ")))
            assert expected in mock_stdout.getvalue()

    @patch('sys.stdout', new_callable=StringIO)
    def test_parse_args_allow_default_help_success(self, mock_stdout):
        cmds = ["allow -h", "allow --help"]
        expected = "allow [-h] [--op1] [--op2] [--op3]"
        for cmd in cmds:
            with pytest.raises(SystemExit):
                self.cli.convert_to_dto(self.parser.parse_args(cmd.split(" ")))
            assert expected in mock_stdout.getvalue()

    @patch('sys.stderr', new_callable=StringIO)
    def test_parse_args_disable_sub_cmd_abbrev_success(self, mock_stderr):
        cmds = ['a', 'd']
        for cmd in cmds:
            expected = f"invalid choice: '{cmd}'"
            with pytest.raises(SystemExit):
                self.cli.convert_to_dto(self.parser.parse_args(cmd.split(" ")))
            assert expected in mock_stderr.getvalue()

    @patch('sys.stderr', new_callable=StringIO)
    def test_parse_args_disable_option_arg_abbrev_success(self, mock_stderr):
        cmds = ['disable_all -o', 'allow -o']
        expected = f"unrecognized arguments: "
        for cmd in cmds:
            with pytest.raises(SystemExit):
                self.cli.convert_to_dto(self.parser.parse_args(cmd.split(" ")))
            assert expected in mock_stderr.getvalue()

    @patch('sys.stderr', new_callable=StringIO)
    def test_parse_args_failed_with_invalid_sub_cmd_choice(self, mock_stderr):
        invalid_choice = "xxx"
        expected = f"invalid choice: '{invalid_choice}'"
        cmd = f"{invalid_choice} --op1 --op3 --op2"

        with pytest.raises(SystemExit):
            self.cli.convert_to_dto(self.parser.parse_args(cmd.split(" ")))
        assert expected in mock_stderr.getvalue()

    @patch('sys.stderr', new_callable=StringIO)
    def test_parse_args_failed_with_invalid_disable_opt(self, mock_stderr):
        invalid_opt = "--onn"
        expected = f"unrecognized arguments: {invalid_opt}"
        cmd = f"disable_all {invalid_opt}"

        with pytest.raises(SystemExit):
            self.cli.convert_to_dto(self.parser.parse_args(cmd.split(" ")))
        assert expected in mock_stderr.getvalue()

    @patch('sys.stderr', new_callable=StringIO)
    def test_parse_args_failed_with_invalid_allow_opt(self, mock_stderr):
        invalid_opt = "--op4"
        expected = f"unrecognized arguments: {invalid_opt}"
        cmd = f"allow --op1 {invalid_opt} --op2"

        with pytest.raises(SystemExit):
            self.cli.convert_to_dto(self.parser.parse_args(cmd.split(" ")))
        assert expected in mock_stderr.getvalue()
