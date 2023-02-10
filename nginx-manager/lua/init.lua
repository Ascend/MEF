-- Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved
-- init const

require "resty.core"
local lrucache = require("resty.lrucache")
local libdns = require("libdns")

g_dns_cache = lrucache.new(100)
g_dns_servers = table.new(5, 0)
g_session_timeout = 600
g_success = "00000000"
g_not_logged_in = "10002001"
g_error_operate = "00002010"
g_upstream_not_found = "10002002"
g_error_lock_state = "10001006"
g_error_pass_or_user = "10001010"
g_error_need_firstlogin = "10001005"
g_not_logged_in_info = "please login"
g_error_operate_info = "failed to operate"
g_success_info = "success"
g_error_lock_state_info = "user or ip in lock state"

libdns.read_dns_servers_from_resolv_file()

g_service_map={}
g_service_map["edgemanager"] = "ascend-edge-manager.mef-center.svc.cluster.local"
g_service_map["softwaremanager"] = "ascend-software-manager.mef-center.svc.cluster.local"
g_service_map["usermanager"] ="ascend-nginx-manager.mef-center.svc.cluster.local"
g_service_map["certmanager"] ="ascend-cert-manager.mef-center.svc.cluster.local"

g_port_map={}
g_port_map["edgemanager"]     = os.getenv("EdgeMgrSvcPort")
g_port_map["softwaremanager"] = os.getenv("SoftwareMgrSvcPort")
g_port_map["usermanager"]    = os.getenv("UserMgrSvcPort")
g_port_map["certmanager"]    = os.getenv("CertMgrSvcPort")