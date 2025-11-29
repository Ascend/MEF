# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from flask import Blueprint

from routes.route import Route


class AccountRoute(Route):

    def __init__(self, blueprint: Blueprint):
        super().__init__(blueprint)

    def add_route(self):
        from account_service.account_views import rf_get_account_info_collection
        from account_service.account_views import rf_get_account_password_expiration_days
        from account_service.account_views import rf_get_specified_account_info
        from account_service.account_views import rf_modify_account_password_expiration_days
        from account_service.account_views import rf_modify_specified_account_info
        self.blueprint.add_url_rule("", view_func=rf_get_account_password_expiration_days, methods=["GET"])
        self.blueprint.add_url_rule("", view_func=rf_modify_account_password_expiration_days, methods=["PATCH"])
        self.blueprint.add_url_rule("/Accounts", view_func=rf_get_account_info_collection, methods=["GET"])
        self.blueprint.add_url_rule("/Accounts/<member_id>", view_func=rf_get_specified_account_info,
                                    methods=["GET"])
        self.blueprint.add_url_rule("/Accounts/<member_id>",
                                    view_func=rf_modify_specified_account_info, methods=["PATCH"])
