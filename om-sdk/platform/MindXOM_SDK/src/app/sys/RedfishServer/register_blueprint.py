# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import importlib

from flask import Flask

from account_service.account_blueprint import account_service_bp
from common.log.logger import run_log
from net_manager.request_api import net_manager_bp
from redfish_service.redfish_views import public_bp
from session_service.session_blueprint import session_service_bp
from system_service.systems_blueprint import system_bp
from upgrade_service.upgrade_blueprint import https_upgrade_service_bp


def register_blueprint(app: Flask, extend_register_func_path: str):
    """
    功能描述：注册蓝图
    app: Flask实例
    extend_register_func_path：注册蓝图扩展函数的路径
    """
    # 注册蓝图
    app.register_blueprint(https_upgrade_service_bp)
    app.register_blueprint(account_service_bp)
    app.register_blueprint(session_service_bp)
    app.register_blueprint(system_bp)
    app.register_blueprint(net_manager_bp)
    app.register_blueprint(public_bp)

    # 注册扩展蓝图
    if not extend_register_func_path:
        return
    try:
        func_index = extend_register_func_path.rfind(".")
        func_name = extend_register_func_path[func_index + 1:]
        model_path = extend_register_func_path[:func_index]
        model = importlib.import_module(model_path)
        if not model or not hasattr(model, func_name):
            run_log.warning("Register extend blueprint failed. reason: model or function not exists")
            return
        extend_register_func = getattr(model, func_name)
        extend_register_func(app)
    except Exception as err:
        run_log.warning("Register extend blueprint failed. reason: %s", err)
