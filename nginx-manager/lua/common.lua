-- Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved

local _M = {}   -- 局部变量，模块名称
local b64 = require("ngx.base64")
local cjson = require("cjson")
local libdynamic = require("libdynamic")

-- 获取经过base64编码的随机字符串
function _M.get_random_string(length)
    local frandom = assert(io.open("/dev/random", "rb"))
    local buf = frandom:read(length)
    io.close(frandom)
    return b64.encode_base64url(buf)
end

-- 解析session标识,将解析sessionID分成sessionID和userID
function _M.parse_session_tag()
    local cookie_name = "cookie___Host-SessionID"
    local sessionTag = ngx.var[cookie_name]
    if sessionTag == nil or string.len(sessionTag) == 0 then
        return nil, nil
    end
    local m = string.gmatch(sessionTag, "([%w_-]+)")
    local sessionID = m()
    local userID = m()
    return sessionID, userID
end

function _M.is_locked()
    local ip_json = ngx.shared.ip_failed_cache:get(ngx.var.remote_addr)
    local user_json = ngx.shared.user_failed_cache:get(1)
    if ip_json == nil and user_json == nil then
        return false
    end
    postBody = {}
    postBody["targetIp"] = ngx.var.remote_addr
    libdynamic.set_upstream("usermanager")
    local res, err = ngx.location.capture("/internal/islocked", {method=ngx.HTTP_POST, body=cjson.encode(postBody), ctx=ngx.ctx})
    local ok, resp = pcall(cjson.decode, res.body)
    if res.status == ngx.HTTP_OK and ok and resp.status == g_success then
        if resp.data.userLocked == false then
            ngx.shared.user_failed_cache:delete(1)
        end
        if resp.data.ipLocked == false then
            ngx.shared.ip_failed_cache:delete(ngx.var.remote_addr)
        end
        if resp.data.userLocked == true or resp.data.ipLocked == true then
            return true
        end
        return false
    end
end

function _M.check_locked()
    if _M.is_locked() then
        _M.sendResp(ngx.HTTP_FORBIDDEN, nil, g_error_lock_state, g_error_lock_state_info)
        return ngx.exit(ngx.HTTP_FORBIDDEN)
    end
end

function _M.sendResp(httpCode, contentType, status, msg)
    if httpCode and httpCode ~= nil then
        ngx.status = httpCode
    end
    if contentType and contentType ~= nil then
        ngx.header["Content-Type"] = contentType
    end
    local resp={}
    resp["status"] = status
    resp["msg"] = msg
    ngx.say(cjson.encode(resp))
end

function _M.sendRespByBody(httpCode, contentType, body)
    if httpCode and httpCode ~= nil then
        ngx.status = httpCode
    end
    if contentType and contentType ~= nil then
        ngx.header["Content-Type"] = contentType
    end
    ngx.say(cjson.encode(body))
end

function _M.sendRespByCapture(res)
    if res and res.status ~= nil then
        _M.sendRespByBody(res.status, nil, res.body)
        return
    end
    _M.sendResp(ngx.HTTP_BAD_REQUEST, nil, g_error_operate, g_error_operate_info)
end

return  _M
