# coding: UTF-8
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import contextlib
import shlex
import subprocess
from typing import List
from typing import Union


class ExecCmd(object):
    @staticmethod
    def exec_cmd(cmd, wait=60) -> int:
        """
        :param cmd: 命令
        :param wait: 超时时间
        :return: int 成功or失败
        """
        ret_code, _ = ExecCmd.exec_cmd_get_output(cmd, wait=wait)
        return ret_code

    @staticmethod
    def exec_cmd_get_output(cmd, stand_in=subprocess.PIPE, stand_out=subprocess.PIPE,
                            stand_err=subprocess.PIPE, wait=60) -> List[Union[int, str]]:
        """
        通过执行cmd命令获取信息
        :param cmd: 输入命令
        :param stand_in: 标准输入重定项
        :param stand_out: 标准输出重定项
        :param stand_err: 标准错误重定项
        :param wait: 命令执行超时时间
        :return: list,成功失败，以及对应的输出信息
        """
        try:
            ret = subprocess.run(cmd, stdin=stand_in, stdout=stand_out, stderr=stand_err, timeout=wait, shell=False)
        except Exception as err:
            return [-1000, f'call linux command error {err}']

        out_msg = ret.stdout.decode("utf-8") if stand_out == subprocess.PIPE else ""
        err_msg = ret.stderr.decode("utf-8") if stand_err == subprocess.PIPE else ""
        ret_msg = out_msg if ret.returncode == 0 else out_msg + err_msg
        return [ret.returncode, ret_msg]

    @staticmethod
    def change_cmd_format(cmds: str):
        """
        分割管道命令
        :param cmds: 命令字符串(管道符需要前后有空格)
        :return: 命令分割后的列表
        """
        cmds_list = []
        result = []
        if " | " in cmds:
            cmds_list = cmds.split(" | ")
        else:
            cmds_list.append(cmds)

        for cmd in cmds_list:
            cmd = shlex.split(cmd)
            result.append(cmd)

        return result

    @staticmethod
    def exec_cmd_use_pipe_symbol(cmds: str, wait=60):
        """
        执行linux命令字符串。
        :param cmds: 命令字符串(管道符需要前后有空格)
        :param wait: 等待时间
        :return: [执行情况，输出内容]
        """
        cmds_list = ExecCmd.change_cmd_format(cmds.strip())
        try:
            with contextlib.ExitStack() as stack:
                for index, cmd in enumerate(cmds_list):
                    stdin = None if index == 0 else proc.stdout
                    proc = stack.enter_context(subprocess.Popen(cmd, stdin=stdin, stdout=subprocess.PIPE,
                                                                stderr=subprocess.PIPE, shell=False))

                out_data, err_data = proc.communicate(timeout=wait)
                ret_msg = out_data.decode() if proc.returncode == 0 else out_data.decode() + err_data.decode()
        except Exception as err:
            return [-1000, f'call linux command error {err}']

        return [proc.returncode, ret_msg]