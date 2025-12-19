# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import logging
import os
import re
from logging.handlers import WatchedFileHandler

MAX_TARGET_LENGTH = 1024
LOG_INJECTION_VALUES = (
    # 转义字符编码
    "\r", "\n", "\f", "\v", "\t", "\b",
    # unicode编码
    "\u000D", "\u000d", "\u000A", "\u000a", "\u000C", "\u000c", "\u000B", "\u000b", "\u0009", "\u0008", "\u007F",
    "\u007f",
    # 16进制码
    "\x0D", "\x0d", "\x0A", "\x0a", "\x0C", "\x0c", "\x0B", "\x0b", "\x09", "\x08", "\x7F", "\x7f",
    # url编码
    "%0D", "%0d", "%0A", "%0a", "%0C", "%0c", "%0B", "%0b", "%09", "%08", "%7F", "%7f",
)


class TrimSensitiveLoggerHandler:
    @staticmethod
    def _path_mask(matched):
        data = matched.group()
        return f"***{data[3:]}" if len(data) > 3 else data

    @staticmethod
    def _remove_abspath(msg):
        if isinstance(msg, str):
            # 增加msg长度限制，超过 MAX_TARGET_LENGTH 个字符就截断，防止正则表达式引起的ReDos攻击
            msg = re.sub(r"((/[0-9a-zA-Z_.-]+)+/)", TrimSensitiveLoggerHandler._path_mask, msg[:MAX_TARGET_LENGTH])
        return msg

    @staticmethod
    def _remove_url_pword(msg):
        if isinstance(msg, str):
            msg = re.sub(r"://.*?@", "://:******@", str(msg))
        return msg

    @staticmethod
    def _remove_pword(msg):
        if isinstance(msg, str):
            # 匹配password、pword、pwd关键字，并将该关键字后面最多20个字符变为*
            rex = r"password(\s*)(.){0,20}|pword(\s*)(.){0,20}|pwd(\s*)(.){0,20}|passwd(\s*)(.){0,20}"
            msg = re.sub(rex, "pwd'******'", str(msg))
        return msg

    @staticmethod
    def _remove_pbkdf2(msg):
        if isinstance(msg, str):
            # 匹配pbkdf2关键字，并将该关键字后面最多160个字符变为*
            msg = re.sub(r"pbkdf2:(\s*)(.){0,160}", "pbkdf2:******", str(msg))
        return msg

    @staticmethod
    def _remove_crlf(msg):
        if not isinstance(msg, str):
            return msg
        # 防止日志注入
        for seq in LOG_INJECTION_VALUES:
            msg = msg.replace(seq, "*")
        return msg

    def handle_sensitive_log(self, msg):
        msg = self._remove_abspath(msg)
        msg = self._remove_url_pword(msg)
        msg = self._remove_pword(msg)
        msg = self._remove_crlf(msg)
        msg = self._remove_pbkdf2(msg)
        return msg


class WatchedFileSoftLinkHandler(WatchedFileHandler, TrimSensitiveLoggerHandler):
    def emit(self, record):
        try:
            if os.path.islink(self.baseFilename):
                os.remove(self.baseFilename)
            super().emit(record)
        except Exception:
            self.handleError(record)

    def format(self, record: logging.LogRecord) -> str:
        return self.handle_sensitive_log(super().format(record))
