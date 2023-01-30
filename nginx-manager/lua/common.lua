-- Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved

local _M = {}   -- 局部变量，模块名称
local b64 = require("ngx.base64")

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


return  _M
