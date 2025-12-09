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
import os
from typing import Iterable, Union

from sqlalchemy import Column, Integer, String, Boolean, ForeignKey

from common.constants.base_constants import UserManagerConstants
from common.db.base_models import Base, SaveDefaultsMixin, ClearOnStartMixin, RebuildMixin
from common.file_utils import FileUtils
from common.utils.date_utils import DateUtils


class User(Base, SaveDefaultsMixin):
    __tablename__ = "users"

    id = Column(Integer, primary_key=True)
    username_db = Column(String(32), index=True, default="admin")
    pword_hash = Column(String(64))
    create_time = Column(String(32), default=DateUtils.default_time)
    pword_modify_time = Column(String(32), default=DateUtils.default_time)
    pword_wrong_times = Column(Integer, default=0)
    start_lock_time = Column(Integer, default=0)
    last_login_success_time = Column(String(32), default="")
    last_login_failure_time = Column(String(32), default="")
    account_insecure_prompt = Column(Boolean, default=True)
    config_navigator_prompt = Column(Boolean, default=True)
    log_in_time = Column(String(32), default="")
    lock_state = Column(Boolean, default=False)
    enabled = Column(Boolean, default=True)
    role_id = Column(Integer, default=1)

    @classmethod
    def default_instance_generator(cls) -> Iterable[Union["User", "HisPwd"]]:
        default_user_id = 1
        modify_time = DateUtils.default_time()
        pw_hash = default_pw()
        yield cls(id=default_user_id, pword_hash=pw_hash, pword_modify_time=modify_time)  # noqa: Mixin导致入参提示有误
        yield HisPwd(user_id=default_user_id, history_pword_hash=pw_hash, pword_modify_time=modify_time)


class Session(Base, ClearOnStartMixin, RebuildMixin):
    __tablename__ = "session_info"

    id = Column(Integer, primary_key=True)
    dialog_id = Column(String(64), unique=True)
    token = Column(String(64), unique=True)
    user_id = Column(Integer)
    request_ip = Column(String(32))
    visit_times = Column(Integer)
    create_time = Column(String(32))
    reset_time = Column(String(32))


class HisPwd(Base):
    __tablename__ = "history_password"

    id = Column(Integer, primary_key=True)
    user_id = Column(Integer, ForeignKey("users.id"))
    history_pword_hash = Column(String(64))  # 数据库中变量声明
    pword_modify_time = Column(Integer)


class EdgeConfig(Base, SaveDefaultsMixin):
    __tablename__ = "edge_config"

    id = Column(Integer, primary_key=True)
    default_lock_times = Column(Integer, default=UserManagerConstants.DEFAULT_LOCKING_TIMES, comment="密码容错次数")
    lock_duration = Column(Integer, default=UserManagerConstants.DEFAULT_LOCKING_DURATION, comment="锁定时长，单位s")
    min_password_length = Column(Integer, default=UserManagerConstants.MIN_PASSWORD_LENGTH, comment="密码最小长度")
    max_password_length = Column(Integer, default=UserManagerConstants.MAX_PASSWORD_LENGTH, comment="密码最大长度")
    token_timeout = Column(Integer, default=UserManagerConstants.DEFAULT_TOKEN_TIMEOUT, comment="token有效期")
    default_expiration_days = Column(Integer, default=UserManagerConstants.DEFAULT_EXPIRATION_DAYS, comment="密码有效期")


class LastLoginInfo(Base):
    __tablename__ = "last_login_info"

    id = Column(Integer, primary_key=True, comment="至多只有一条数据")
    ip = Column(String, comment="上一次登入的IP")


def default_pw():
    """获取默认密码"""
    cur_dir = os.path.dirname(os.path.realpath(__file__))
    file_path = os.path.join(cur_dir, "config", "paw.ini")
    file_content = FileUtils.read_file(file_path, "r")
    if not file_content:
        raise ValueError("invalid default pwd context.")
    return file_content[0]
