# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from flask import Blueprint

from routes.route import Route


class SecurityServiceRoute(Route):

    def __init__(self, blueprint: Blueprint):
        super().__init__(blueprint)

    def add_route(self):
        # 添加安全服务相关的URL
        from system_service.security_service_views import rf_get_system_https_cert_alarm_time
        from system_service.security_service_views import rf_import_system_puny_dict
        from system_service.security_service_views import rf_export_system_puny_dict
        from system_service.security_service_views import rf_delete_system_puny_dict
        from system_service.security_service_views import rf_https_cert
        from system_service.security_service_views import rf_import_custom_certificate
        from system_service.security_service_views import rf_modify_system_https_cert_alarm_time
        from system_service.security_service_views import rf_security_service
        from system_service.security_service_views import rf_download_csr_file
        from system_service.security_service_views import rf_get_security_load_config_info
        from system_service.security_service_views import rf_modify_security_load_config_info
        from system_service.security_service_views import rf_export_security_load
        from system_service.security_service_views import rf_import_security_load

        self.blueprint.add_url_rule("/SecurityService", view_func=rf_security_service, methods=['GET'])
        self.blueprint.add_url_rule("/SecurityService/HttpsCert", view_func=rf_https_cert, methods=['GET'])
        self.blueprint.add_url_rule("/SecurityService/HttpsCert/Actions/HttpsCert.ImportServerCertificate",
                                    view_func=rf_import_custom_certificate, methods=['POST'])
        self.blueprint.add_url_rule("/SecurityService/Actions/SecurityService.PunyDictImport",
                                    view_func=rf_import_system_puny_dict, methods=['POST'])
        self.blueprint.add_url_rule("/SecurityService/Actions/SecurityService.PunyDictExport",
                                    view_func=rf_export_system_puny_dict, methods=['POST'])
        self.blueprint.add_url_rule("/SecurityService/Actions/SecurityService.PunyDictDelete",
                                    view_func=rf_delete_system_puny_dict, methods=['POST'])
        self.blueprint.add_url_rule("/SecurityService/HttpsCertAlarmTime",
                                    view_func=rf_get_system_https_cert_alarm_time, methods=['GET'])
        self.blueprint.add_url_rule("/SecurityService/HttpsCertAlarmTime",
                                    view_func=rf_modify_system_https_cert_alarm_time, methods=["PATCH"])
        self.blueprint.add_url_rule("/SecurityService/downloadCSRFile",
                                    view_func=rf_download_csr_file, methods=["POST"])
        self.blueprint.add_url_rule("/SecurityService/SecurityLoad",
                                    view_func=rf_get_security_load_config_info, methods=["GET"])
        self.blueprint.add_url_rule("/SecurityService/SecurityLoad",
                                    view_func=rf_modify_security_load_config_info, methods=["PATCH"])
        self.blueprint.add_url_rule("/SecurityService/SecurityLoad/Actions/SecurityLoad.Export",
                                    view_func=rf_export_security_load, methods=["POST"])
        self.blueprint.add_url_rule("/SecurityService/SecurityLoad/Actions/SecurityLoad.Import",
                                    view_func=rf_import_security_load, methods=["POST"])
