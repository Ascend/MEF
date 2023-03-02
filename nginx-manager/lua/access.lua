-- Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved
-- access control

local libaccess = require("libaccess")
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
