-- Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved

local cjson = require("cjson")
local common = require("common")
local libdynamic = require("libdynamic")
local libaccess = require("libaccess")

local function create_session(resp)
    local session = {}
    session.UserID     = resp.userid        -- user id
    session.LoginTime  = ngx.now()
    session.SessionID  = common.get_random_string(32)
    session.csrf_token  = common.get_random_string(32)
    session.RemoteAddr = ngx.var.remote_addr
    return session
end

if ngx.req.get_method() ~= "POST" then
    ngx.log(ngx.ERR, "Login method " .. ngx.req.get_method() .. " is not allow")
    return ngx.exit(ngx.HTTP_NOT_ALLOWED)
end

ngx.req.read_body()
local body = ngx.req.get_body_data()
libdynamic.set_upstream("usermanager")

local res, err = ngx.location.capture("/internal/login", { method=ngx.HTTP_POST, body=body, ctx=ngx.ctx })
local ok, resp = pcall(cjson.decode, res.body)
if res.status == ngx.HTTP_OK and ok and resp.status == g_success then
    local session = create_session(resp.data)

    -- session 10分钟超时
    ngx.shared.session_cache:set(session.UserID, cjson.encode(session), g_session_timeout)

    resp.data["csrf_token"] = session.csrf_token
    local cookie = "__Host-SessionID=" .. session.SessionID .. "." .. session.UserID .. "; Max-Age=" .. g_session_timeout
    ngx.header["Set-Cookie"] = cookie .. "; Path=/; SameSite=Strict; Secure=true; HttpOnly"
    ngx.header["Content-Type"] = "application/json"
    ngx.say(cjson.encode(resp))
    ngx.log(ngx.NOTICE, "Auth success: user id = " .. session.UserID)
elseif resp.status == g_error_pass_or_user then
    libaccess.handleLockResp(resp)
else
    resp={}
    ngx.status = ngx.HTTP_UNAUTHORIZED
    if res and res.body then
        resp = res.body
    else
        resp["info"] = "failed"
        resp["msg"] = res.body
    end
    ngx.say(cjson.encode(resp))
    ngx.log(ngx.NOTICE, "Auth failed: ip addr =  " .. ngx.var.remote_addr)
end
