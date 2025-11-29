# Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
from flask import Flask

from om_event_subscription.subscription_blueprint import event_subscription_bp
from om_system_service.digital_warranty_service_views import https_digital_warranty_service_bp
from om_system_service.om_actions_views import om_actions_service_bp


def register_om_blueprint(app: Flask):
    """
    功能描述：注册蓝图
    app: Flask实例
    """
    # 注册OM特有的接口蓝图
    app.register_blueprint(https_digital_warranty_service_bp)
    app.register_blueprint(event_subscription_bp)
    app.register_blueprint(om_actions_service_bp)
