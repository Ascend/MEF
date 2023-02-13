-- Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved
-- access control

local libaccess = require("libaccess")
local libdynamic = require("libdynamic")
local common = require("common")

-- 检查是否锁定
common.check_locked()

-- 检查是否登录
local session = libaccess.get_session_info()
if session == nil then
    ngx.log(ngx.NOTICE, "invalid cookie")
    common.sendResp(ngx.HTTP_UNAUTHORIZED, "application/json", g_not_logged_in, g_not_logged_in_info)
    return ngx.exit(ngx.HTTP_UNAUTHORIZED)
end

-- 设置路由
if string.len(ngx.var.service_name) > 0 then
    libdynamic.set_upstream(ngx.var.service_name)
end
