-- Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved
-- init const

require "resty.core"
-- user or ip lock duration, default 600 seconds
local default_lock_time = 600
-- unlock user or ip every 30s automatic
local default_lock_offset = 30
local lrucache = require("resty.lrucache")
local common = require("common")
g_dns_cache = lrucache.new(100)
g_dns_servers = table.new(5, 0)
g_default_session_timeout = 600
g_session_timeout = common.getEnvWithDefault("TokenExpireTime", g_default_session_timeout)
g_lock_time = common.getEnvWithDefault("LockTime", default_lock_time) + default_lock_offset
g_success = "00000000"
g_error_operate = "00002010"
g_error_need_firstlogin = "10001005"
g_error_lock_state = "10001006"
g_error_pass_or_user = "10001010"
g_not_logged_in = "10002001"
g_upstream_not_found = "10002002"
g_not_logged_in_info = "please login"
g_error_operate_info = "failed to operate"
g_success_info = "success"
g_error_lock_state_info = "user or ip in lock state"