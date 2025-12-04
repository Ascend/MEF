# coding: utf-8
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import json
from typing import Union
import abc
import argparse
import enum
from typing import Type
from common.log.logger import run_log


class EnumBase(enum.Enum):
    @classmethod
    def values(cls):
        return [i.value for i in cls]


class SwitchVal(EnumBase):
    """开关值枚举"""
    ON = True
    OFF = False

    def __bool__(self):
        return self.value


class ArgDto:
    """命令行参数dto"""
    def __init__(self, **kwargs):
        self.name = kwargs.get("name")
        self.default = kwargs.get("default")
        self.help = kwargs.get("help")
        self._choices = kwargs.get("choices")
        self.value = kwargs.get("value")


class SubCmdDto:
    """
    子命令 dto
    """
    def __init__(self, cmd, help_doc, opts):
        self.cmd = cmd
        self.help = help_doc
        self.opts = [ArgDto(**opt) for opt in opts]


class OpEnumBase(EnumBase):
    """
    高危操作枚举类型基类
    """
    pass


class Cmd(EnumBase):
    """
    支持的子命令
    """
    DISABLE_ALL = 'disable_all'
    ALLOW = 'allow'


class CmdDisableAllOpt(EnumBase):
    """
    disable_all子命令, optional参数
    """
    ON = 'on'


class HighRiskOpPolicyDto:
    """
    高危操作策略dto
    """
    def __init__(self, risk_op_enum: Type[OpEnumBase], disable_all=SwitchVal.ON, allow=None):
        self.risk_op_enum = risk_op_enum
        self.disable_all = disable_all
        allow = allow or {}
        self.allow = {
            risk_op_enum(risk_op): SwitchVal(allow.get(risk_op, SwitchVal.OFF))
            for risk_op in risk_op_enum.values()
        }

    @classmethod
    def load_from_json(cls, risk_op_enum, json_conf):
        """
        将json对象转换为dto对象

        :param risk_op_enum: 高危操作枚举
        :param json_conf: json对象
        :return:
        """
        disable_all = json_conf.get(Cmd.DISABLE_ALL.value, SwitchVal.ON)
        allow = json_conf.get(Cmd.ALLOW.value, {})
        return cls(risk_op_enum, disable_all, allow)

    def dump_to_json(self):
        """
        将dto对象转换为json对象

        :return:
        """
        return {
            Cmd.DISABLE_ALL.value: self.disable_all.value,
            Cmd.ALLOW.value: {
                risk_op.value: self.allow.get(risk_op, SwitchVal.OFF).value
                for risk_op in self.risk_op_enum
            }
        }

    def check_allow(self, srv: OpEnumBase):
        """
        检查特性开关是否打开

        :param srv:
        :return:
        """
        if self.disable_all:
            return False

        return bool(self.allow.get(srv, SwitchVal.OFF))


class HighRiskOpPolicyCli(abc.ABC):
    """
    高危操作cli类
    """
    def __init__(self, cmds_conf, risk_op_enum, config_file=None):
        self.cmds_conf = cmds_conf
        self.risk_op_enum = risk_op_enum
        self.description = cmds_conf.get("description", "")
        self.config_file = config_file

        disable_cmd, allow_cmd = cmds_conf["sub_cmds"]
        self.disable_cmd = SubCmdDto(disable_cmd['cmd'], disable_cmd['help'], disable_cmd['opts'])
        self.allow_cmd = SubCmdDto(allow_cmd['cmd'], allow_cmd['help'], allow_cmd['opts'])
        self._cli_parser = None
        self._sub_parsers = {}

    def gen_cli_parser(self):
        parser = argparse.ArgumentParser(prog=self.cmds_conf.get("prog"),
                                         description=self.description,
                                         allow_abbrev=False)
        subparsers = parser.add_subparsers(title="", help='sub-command help')

        # disable_all 子命令
        parser_sub = subparsers.add_parser(self.disable_cmd.cmd, allow_abbrev=False, help=self.disable_cmd.help)
        opt = self.disable_cmd.opts[0]
        name = f"--{opt.name}"
        parser_sub.add_argument(
            name,
            default=opt.default,
            help=opt.help,
            action="store_const",
            const=SwitchVal.ON,
        )
        self._sub_parsers[self.disable_cmd.cmd] = parser_sub

        # allow 子命令
        parser_sub = subparsers.add_parser(self.allow_cmd.cmd, allow_abbrev=False, help=self.allow_cmd.help)
        for opt in self.allow_cmd.opts:
            name = f"--{opt.name}"
            parser_sub.add_argument(
                name,
                default=opt.default,
                help=opt.help,
                action="store_const",
                const=SwitchVal.ON,
            )
        self._sub_parsers[self.allow_cmd.cmd] = parser_sub

        self._cli_parser = parser

        return self._cli_parser

    def print_help(self, sub_cmd=None):
        """
        输出帮助信息, sub_cmd为None时打印总命令帮助信息

        :param sub_cmd:
        :return: 是否成功输出帮助信息
        """
        if sub_cmd is None:
            self._cli_parser.print_help()
            return True

        sub_parser = self._sub_parsers.get(sub_cmd)
        if sub_parser:
            sub_parser.print_help()
            return True

        return False

    def parse_args(self):
        """
        从命令行加载规则

        :return:
        """
        if not self._cli_parser:
            self.gen_cli_parser()

        config = self._cli_parser.parse_args()
        return self.convert_to_dto(config)

    def convert_to_dto(self, config) -> HighRiskOpPolicyDto:
        """
        将从 命令行解析的参数 转换到 dto
        :param config:
        :return:
        """
        if hasattr(config, CmdDisableAllOpt.ON.value):
            disable_all = getattr(config, CmdDisableAllOpt.ON.value)
            return HighRiskOpPolicyDto(self.risk_op_enum, disable_all=disable_all)

        allow_opt = {
            risk_op: getattr(config, risk_op, SwitchVal.OFF)
            for risk_op in self.risk_op_enum.values()
        }
        return HighRiskOpPolicyDto(
            self.risk_op_enum,
            disable_all=SwitchVal.OFF,
            allow=allow_opt
        )


class OmAbility(OpEnumBase):
    """
    OM 能力列表
    """
    MEF_CONFIG = "mef_config"


class AbilityConfig:
    POLICY: Union[HighRiskOpPolicyDto, None] = None
    CONFIG_FILE = "/home/data/config/ability_policy.json"


def init(config_file):
    """
    初始化读取能力开关配置文件

    :return:
    """
    if AbilityConfig.POLICY:
        return

    try:
        with open(config_file, "r") as file:
            config_json = json.load(file)
    except Exception as err:
        run_log.error("load json file failed, %s", err)
        raise err

    try:
        AbilityConfig.POLICY = HighRiskOpPolicyDto.load_from_json(OmAbility, config_json)
    except Exception as err:
        run_log.error("init policy failed, %s", err)
        raise err


def is_allow(srv: OmAbility):
    """
    能力开关检查

    :param srv:
    :return:
    """
    if not AbilityConfig.POLICY.check_allow(srv):
        run_log.error("check ability policy failed, feature %s is disable", srv.value)
        return False

    return True
