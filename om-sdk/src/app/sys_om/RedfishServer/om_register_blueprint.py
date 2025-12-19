# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
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
