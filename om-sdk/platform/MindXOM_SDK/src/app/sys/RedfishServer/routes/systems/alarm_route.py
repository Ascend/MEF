# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from flask import Blueprint

from routes.route import Route


class AlarmRoute(Route):

    def __init__(self, blueprint: Blueprint):
        super().__init__(blueprint)

    def add_route(self):
        # Alarm类的URL
        from system_service.alarm_views import rf_get_system_alarm
        from system_service.alarm_views import rf_get_system_alarm_info
        from system_service.alarm_views import rf_get_system_alarm_shield
        from system_service.alarm_views import rf_increase_system_alarm_shield
        from system_service.alarm_views import rf_decrease_system_alarm_shield

        self.blueprint.add_url_rule("/Alarm", view_func=rf_get_system_alarm, methods=["GET"])
        self.blueprint.add_url_rule("/Alarm/AlarmInfo", view_func=rf_get_system_alarm_info, methods=["GET"])
        self.blueprint.add_url_rule("/Alarm/AlarmShield", view_func=rf_get_system_alarm_shield, methods=["GET"])
        self.blueprint.add_url_rule("/Alarm/AlarmShield/Increase",
                                    view_func=rf_increase_system_alarm_shield, methods=["PATCH"])
        self.blueprint.add_url_rule("/Alarm/AlarmShield/Decrease",
                                    view_func=rf_decrease_system_alarm_shield, methods=["PATCH"])
