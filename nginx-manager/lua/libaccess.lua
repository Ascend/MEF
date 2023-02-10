-- Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved

local _M = {}   -- 局部变量，模块名称
local cjson = require("cjson")
local common = require("common")


function _M.get_session_info()
    local sessionID, userID = common.parse_session_tag()
    if sessionID == nil or userID == nil then
        ngx.log(ngx.ERR, "cookie does not exist")
        return nil
    end
    local sess = ngx.shared.session_cache:get(userID)
    if sess == nil then
        ngx.log(ngx.ERR, "session does not exist")
        return nil
    end
    local sess_info = cjson.decode(sess)
    if sessionID ~= sess_info.SessionID then
        ngx.log(ngx.ERR, "session_id verify failed")
        return nil
    end
    local csrf_token = ngx.req.get_headers()["X-CSRF-Token"]
    if ngx.var.third_party ~= "true"  and (csrf_token == nil or csrf_token ~= sess_info.csrf_token) then
        ngx.log(ngx.ERR, "csrf_token verify failed")
        return nil
    end

    -- 更新session过期时间
    ngx.shared.session_cache:expire(userID, g_session_timeout)
    -- 更新cookie过期时间
    local cookie = "__Host-SessionID=" .. sessionID .. "." .. userID .. "; Max-Age=" .. g_session_timeout
    ngx.header["Set-Cookie"] = cookie .. "; Path=/; SameSite=Strict; Secure=true; HttpOnly"

    -- 将session中字段填到请求的header中
    for key, value in pairs(sess_info) do
        ngx.req.set_header(key, value)
    end
    return sess_info
end

function _M.del_session_by_id(session_id)
    local sess = ngx.shared.session_cache:get(session_id)
    if sess == nil then
        return
    end
    -- 更新session过期时间
    ngx.shared.session_cache:delete(session_id)
    ngx.header["Set-Cookie"] = "__Host-SessionID=; Max-Age=0; Path=/; SameSite=Strict; Secure=true; HttpOnly"
    return
end

function _M.handleLockResp(resp)
    if resp.data == nil then
        common.sendRespByBody(ngx.HTTP_BAD_REQUEST, nil, resp)
        return
    end

    if resp.data.userLocked == true then
        ngx.shared.user_failed_cache:set(1, true, 800)
        ngx.log(ngx.NOTICE, resp.data.userid .. " locked")
    end
    if resp.data.ipLocked == true then
        ngx.shared.ip_failed_cache:set(ngx.var.remote_addr, true, 800)
        ngx.log(ngx.NOTICE, resp.data.ip .. " locked")
    end

    if resp.data.ipLocked == true or resp.data.userLocked == true then
        common.sendResp(ngx.HTTP_UNAUTHORIZED, nil, g_error_lock_state, g_error_lock_state_info)
    else
        ngx.say(cjson.encode(resp))
    end
end

return _M
