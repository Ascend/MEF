-- Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved
-- access control

local cjson = require("cjson")
local libaccess = require("libaccess")
local libdynamic = require("libdynamic")

ngx.log(ngx.INFO, "begin to access-check.")

-- get session info
local session = libaccess.get_session_info()
if session == nil then
    ngx.log(ngx.NOTICE, "invalid cookie")
    local resp = {}
    ngx.header["Content-Type"] = "application/json"
    resp.status= g_not_logged_in
    resp.msg = "please login"
    ngx.status = ngx.HTTP_UNAUTHORIZED
    ngx.say(cjson.encode(resp))
    return ngx.exit(ngx.HTTP_UNAUTHORIZED)
end

-- set upstream
if string.len(ngx.var.service_name) > 0 then
    libdynamic.set_upstream(ngx.var.service_name)
end
