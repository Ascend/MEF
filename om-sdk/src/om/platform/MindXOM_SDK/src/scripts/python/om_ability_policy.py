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
import json
import os
import signal
import sys

from common.backup_restore_service.backup import Backup
from common.constants.base_constants import CommonConstants
from common.utils.ability_policy import OmAbility, AbilityConfig, SwitchVal, Cmd, CmdDisableAllOpt, HighRiskOpPolicyCli
from logger import terminal_print


def signal_handler(sig_num=None, frame=None):
    """
    信号处理函数
    :param sig_num: 信号值
    :param frame: 栈帧
    :return: None
    """
    sys.exit(1)


class OmCli(HighRiskOpPolicyCli):
    """
    Om cli类
    """
    pass


OM_CMDS_CONF = {
    "description": "config om ability policy.",
    "prog": "om_ability_policy.sh",
    "sub_cmds": [
        {
            "cmd": Cmd.DISABLE_ALL.value,
            "help": "disable all ops.",
            "opts": [
                {
                    "name": CmdDisableAllOpt.ON.value,
                    "default": SwitchVal.OFF,
                    "help": "disable all ability.",
                }
            ]
        },
        {
            "cmd": Cmd.ALLOW.value,
            "help": "allow ops.",
            "opts": [
                {
                    "name": OmAbility.MEF_CONFIG.value,
                    "default": SwitchVal.OFF,
                    "help": "to enable net config.",
                },
            ],
        }
    ]
}


class HelpException(Exception):
    pass


def check_param(cli: HighRiskOpPolicyCli):
    """
    命令行参数预检查

    :param cli:
    :return:
    """
    help_opts = ("-h", "--help")
    arg_num = len(sys.argv)
    # 总命令带help或不带help
    if arg_num == 1 or (arg_num == 2 and sys.argv[1] in help_opts):
        cli.print_help()
        raise HelpException()
    # 子命令带help或不带help
    if arg_num == 2 or (arg_num == 3 and (sys.argv[2] in help_opts)):
        cmd = sys.argv[1]
        if cli.print_help(cmd):
            raise HelpException()

    if arg_num >= 2 and sys.argv[1] not in Cmd.values():
        cli.print_help()
        raise HelpException()


def cli_set_config(config_file):
    """
    从命令行读取，写入配置文件
    :return:
    """
    cli = OmCli(OM_CMDS_CONF, OmAbility)
    cli.gen_cli_parser()
    check_param(cli)

    dto = cli.parse_args()
    json_str = json.dumps((dto.dump_to_json()), indent=2)
    with os.fdopen(os.open(config_file, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o644), "w") as file:
        file.write(json_str)
    terminal_print.info("Write to config file success")
    terminal_print.info(json_str)
    terminal_print.info("Warning: You also need to reboot system to take effect.")
    return 0


def main():
    try:
        ret_code = cli_set_config(AbilityConfig.CONFIG_FILE)
    except HelpException:
        ret_code = 1
    except Exception as err:
        ret_code = 2
        terminal_print.error("run cmd failed, %s", err)
    except SystemExit:
        ret_code = 3

    # 能力项设置成功之后需要重启生效，生效之前执行配置文件备份
    if ret_code == 0:
        Backup(CommonConstants.MONITOR_BACKUP_DIR, AbilityConfig.CONFIG_FILE).entry()

    return ret_code


if __name__ == "__main__":
    # 注册退出信号的中断处理函数
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)
    sys.exit(main())
