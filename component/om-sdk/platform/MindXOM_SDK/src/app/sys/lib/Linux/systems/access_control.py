# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import configparser
import os
import threading

from common.file_utils import FileCheck
from common.init_cmd import cmd_constants
from common.log.logger import operate_log
from common.log.logger import run_log
from common.utils.exec_cmd import ExecCmd
from common.common_methods import CommonMethods


class AccessControlError(Exception):
    pass


class AccessControl:
    ACCESS_LOCK = threading.Lock()
    ACCESS_INI = "/home/data/ies/access_control.ini"
    CONFIG_PARSER = configparser.RawConfigParser()

    def __init__(self):
        self.web_access = None
        self.ssh_access = None

    @staticmethod
    def _turn_on_web_access():
        cmd = f"{cmd_constants.OS_CMD_SYSTEMCTL} restart start-nginx"
        ret = ExecCmd.exec_cmd_use_pipe_symbol(cmd, wait=30)
        if ret[0] != 0:
            raise AccessControlError("Restart start-nginx failed")

    @staticmethod
    def _turn_off_web_access():
        cmd = f"{cmd_constants.OS_CMD_SYSTEMCTL} stop start-nginx"
        ret = ExecCmd.exec_cmd_use_pipe_symbol(cmd, wait=30)
        if ret[0] != 0:
            raise AccessControlError("Stop start-nginx failed")

    @staticmethod
    def _turn_on_ssh_access():
        sshd_service = "ssh" if cmd_constants.OS_NAME == "Ubuntu" else "sshd"
        cmd = f"{cmd_constants.OS_CMD_SYSTEMCTL} enable {sshd_service}"
        ret = ExecCmd.exec_cmd_use_pipe_symbol(cmd, wait=30)
        if ret[0] != 0:
            raise AccessControlError("Enable ssh failed")

        cmd = f"{cmd_constants.OS_CMD_SYSTEMCTL} restart {sshd_service}"
        ret = ExecCmd.exec_cmd_use_pipe_symbol(cmd, wait=30)
        if ret[0] != 0:
            raise AccessControlError("Restart ssh failed")

    @staticmethod
    def _turn_off_ssh_access():
        sshd_service = "ssh" if cmd_constants.OS_NAME == "Ubuntu" else "sshd"
        cmd = f"{cmd_constants.OS_CMD_SYSTEMCTL} stop {sshd_service}"
        ret = ExecCmd.exec_cmd_use_pipe_symbol(cmd, wait=30)
        if ret[0] != 0:
            raise AccessControlError("Stop ssh failed")

        cmd = f"{cmd_constants.OS_CMD_SYSTEMCTL} disable {sshd_service}"
        ret = ExecCmd.exec_cmd_use_pipe_symbol(cmd, wait=30)
        if ret[0] != 0:
            raise AccessControlError("Disable ssh failed")
    
    def check_and_parse_ini_file(self):
        ret = FileCheck.check_path_is_exist_and_valid(self.ACCESS_INI)
        if not ret:
            run_log.error("access control ini file is invalid")
            raise AccessControlError(ret.error)
        self.CONFIG_PARSER.read(self.ACCESS_INI)

    def write_access_ini_file(self, access_type, access_value):
        self.CONFIG_PARSER.set("access_control", access_type, str(access_value))
        flags = os.O_WRONLY | os.O_CREAT | os.O_TRUNC
        try:
            FileCheck.check_is_link_exception(self.ACCESS_INI)
            with os.fdopen(os.open(self.ACCESS_INI, flags, 0o640), "w+") as file:
                self.CONFIG_PARSER.write(file)
        except Exception as err:
            err_msg = "write config to access control ini file failed."
            run_log.error(err_msg)
            raise AccessControlError(err_msg) from err

    def patch_request(self, request_data):
        if self.ACCESS_LOCK.locked():
            run_log.warning("Access control process is busy")
            return [CommonMethods.ERROR, "Access control process is busy"]

        with self.ACCESS_LOCK:
            username = request_data.get("_User")
            request_ip = request_data.get("_Xip")
            switch_mapper = {
                True: "on",
                False: "off"
            }
            if request_data.get("web_access") is not None:
                message = f"Turn access control[web] {switch_mapper.get(request_data.get('web_access'))} successfully."

                try:
                    self._effect_access_control(request_data, "web_access")
                except Exception as err:
                    message = f"Turn access control[web] {switch_mapper.get(request_data.get('web_access'))} failed"
                    return [CommonMethods.ERROR, message]
                finally:
                    operate_log.info("[%s@%s] %s", username, request_ip, message)

            if request_data.get("ssh_access") is not None:
                message = f"Turn access control[ssh] {switch_mapper.get(request_data.get('ssh_access'))} successfully."
                try:
                    self._effect_access_control(request_data, "ssh_access")
                except Exception as err:
                    message = f"Turn access control[ssh] {switch_mapper.get(request_data.get('ssh_access'))} failed."
                    return [CommonMethods.ERROR, message]
                finally:
                    operate_log.info("[%s@%s] %s", username, request_ip, message)

            self.get_all_info()
            return [CommonMethods.OK, ""]
                    
    def get_all_info(self):
        if self.ACCESS_LOCK.locked():
            run_log.warning("Access control process is busy")
            return [CommonMethods.ERROR, "Access control process is busy"]

        with self.ACCESS_LOCK:
            try:
                self.check_and_parse_ini_file()
            except Exception as err:
                run_log.error("access control ini file is invalid")
                return [CommonMethods.ERROR, "access control ini file is invalid"]

            self.web_access = self.CONFIG_PARSER.getboolean("access_control", "web_access")
            self.ssh_access = self.CONFIG_PARSER.getboolean("access_control", "ssh_access")
            return [CommonMethods.OK, ""]

    def _effect_access_control(self, request_data, access_type):
        access_value = request_data.get(access_type)
        if not isinstance(access_value, bool):
            message = f"config {access_type} failed, its format is invalid."
            run_log.error(message)
            raise AccessControlError(message)

        self.check_and_parse_ini_file()
        prev_access_value = self.CONFIG_PARSER.getboolean("access_control", access_type)
        self.write_access_ini_file(access_type, access_value)

        access_type_value = f"{access_type} {access_value}"
        try:
            {
                "web_access True": self._turn_on_web_access,
                "web_access False": self._turn_off_web_access,
                "ssh_access True": self._turn_on_ssh_access,
                "ssh_access False": self._turn_off_ssh_access,
            }.get(access_type_value)()
        except Exception as err:
            message = f"switch {access_type} failed."
            run_log.error("%s the reason is %s.", message, err)
            run_log.info("%s Now rollback to the previous configuration.", message)
            try:
                self.write_access_ini_file(access_type, prev_access_value)
            except Exception:
                run_log.error("Roll backing to the previous configuration failed.")
            raise AccessControlError(message) from err
        message = f"switch {access_type} successfully."
        run_log.info(message)
