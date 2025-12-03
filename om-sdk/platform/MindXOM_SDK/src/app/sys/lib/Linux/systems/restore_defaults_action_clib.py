#!/usr/bin/python3
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.

__all__ = ['authenticate']

from ctypes import c_void_p, c_uint, c_char_p, c_char, c_int, CDLL, POINTER, Structure, CFUNCTYPE, pointer, cast, sizeof
from ctypes.util import find_library

LIB_PAM = CDLL(find_library("pam"))
LIB_C = CDLL(find_library("c"))

CAL_LOC = LIB_C.calloc
CAL_LOC.restype = c_void_p
CAL_LOC.argtypes = [c_uint, c_uint]

STRD_UP = LIB_C.strdup
STRD_UP.argstypes = [c_char_p]
STRD_UP.restype = POINTER(c_char)  # NOT c_char_p !!!!

# Various constants
PAM_PROMPT_ECHO_OFF = 1
PAM_PROMPT_ECHO_ON = 2
PAM_ERROR_MSG = 3
PAM_TEXT_INFO = 4


class PamProcess(Structure):
    """wrapper class for pam_handle_t"""
    _fields_ = [("handle", c_void_p)]

    def __init__(self):
        super().__init__()
        self.handle = 0


class PamMessage(Structure):
    _fields_ = [("msg_style", c_int), ("msg", c_char_p), ]

    def __repr__(self):
        return "<Pam Msg %i '%s'>" % (self.msg_style, self.msg)


class PamResponse(Structure):
    _fields_ = [("resp", c_char_p), ("resp_retcode", c_int)]

    def __repr__(self):
        return "<Pam Resp %i '%s'>" % (self.resp_retcode, self.resp)


CONV_FUNC = CFUNCTYPE(c_int,
                      c_int, POINTER(POINTER(PamMessage)),
                      POINTER(POINTER(PamResponse)), c_void_p)


class PamConv(Structure):
    _fields_ = [("conv", CONV_FUNC), ("appdata_ptr", c_void_p)]


PY_PAM_START = LIB_PAM.pam_start
PY_PAM_START.restype = c_int
PY_PAM_START.argtypes = [c_char_p, c_char_p, POINTER(PamConv), POINTER(PamProcess)]

PY_PAM_AUTHENTICATE = LIB_PAM.pam_authenticate
PY_PAM_AUTHENTICATE.restype = c_int
PY_PAM_AUTHENTICATE.argtypes = [PamProcess, c_int]

PY_PAM_END = LIB_PAM.pam_end
PY_PAM_END.restype = c_int
PY_PAM_END.argtypes = [PamProcess, c_int]


def authenticate(username, password, service='login'):
    """
        Returns True if the given username and password authenticate for the given service.  Returns False otherwise
        ``username``: the username to authenticate
        ``password``: the password in plain text
        ``service``: the PAM service to authenticate against. Defaults to 'login'
    """
    # 参数编码
    username = username.encode("utf-8")
    service = service.encode("utf-8")

    @CONV_FUNC
    def my_conversation(n_messages, messages, p_response, app_data):
        addr = CAL_LOC(n_messages, sizeof(PamResponse))
        p_response[0] = cast(addr, POINTER(PamResponse))
        for i in range(n_messages):
            if messages[i].contents.msg_style == PAM_PROMPT_ECHO_OFF:
                pw_copy = STRD_UP(str(password).encode("utf-8"))
                p_response.contents[i].resp = cast(pw_copy, c_char_p)
                p_response.contents[i].resp_retcode = 0

        return 0

    handle = PamProcess()
    pam_conversation = PamConv(my_conversation, 0)
    ret_val = PY_PAM_START(service, username, pointer(pam_conversation), pointer(handle))
    if ret_val != 0:
        return False
    ret_val = PY_PAM_AUTHENTICATE(handle, 0)
    my_conversation = None
    PY_PAM_END(handle, 0)
    return ret_val == 0
