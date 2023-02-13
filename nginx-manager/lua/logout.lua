-- Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved

local cjson = require("cjson")
local common = require("common")
local libaccess = require("libaccess")

common.check_method("POST")
common.check_locked()

-- get user info will check sessionID
local session = libaccess.get_session_info()
if session == nil then
    ngx.log(ngx.NOTICE, "invalid cookie")
    common.sendResp(ngx.HTTP_UNAUTHORIZED, "application/json", g_not_logged_in, g_not_logged_in_info)
    return ngx.exit(ngx.HTTP_UNAUTHORIZED)
end

-- delete session info
libaccess.del_session_by_id(session.UserID)
ngx.log(ngx.NOTICE, session.UserID .. " logout")
common.sendResp(ngx.HTTP_OK, nil, g_success, g_success_info)
