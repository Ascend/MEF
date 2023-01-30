-- Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved

local cjson = require("cjson")
local common = require("common")
local libaccess = require("libaccess")


if ngx.req.get_method() ~= "POST" then
    ngx.log(ngx.ERR, "Logout method " .. ngx.req.get_method() .. " is not allow")
    return ngx.exit(ngx.HTTP_NOT_ALLOWED)
end

-- get user info will check sessionID
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

-- delete session info
libaccess.del_session_by_id(session.UserID)
ngx.log(ngx.NOTICE, session.UserID .. " logout")
ngx.status = ngx.HTTP_OK
local resp={}
resp["status"] = g_success
resp["msg"] = "logout success"
ngx.say(cjson.encode(resp))
