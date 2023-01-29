-- Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved

local _M = {}   -- 局部变量，模块名称
local cjson = require("cjson")
local libdns = require("libdns")

-- 设置动态上游，如果dns查询失败，则返回失败
function _M.set_upstream(service_name)
    local service_domain = g_service_map[service_name]
    if service_domain == nil then
        -- service_name -> domain failed
        ngx.log(ngx.ERR, "no such service " .. service_name)
        ngx.status = 401
        local resp={}
        resp["status"] = g_upstream_not_found
        resp["msg"] = "no such service " .. service_name
        ngx.say(cjson.encode(resp))
        return
    end

    -- get backend service ip and port
    ngx.ctx.target_ip = libdns.get_addr(service_domain)
    ngx.ctx.target_port = g_port_map[service_name]
    ngx.var.target_host = service_domain
    if ngx.ctx.target_ip == nil then
        -- dns lookup failed
        ngx.log(ngx.ERR, "dns lookup failed for " .. service_domain)
        ngx.status = 500
        local resp={}
        resp["status"] = g_upstream_not_found
        resp["msg"] = "can not find service " .. service_name
        ngx.say(cjson.encode(resp))
        return
    end
end

return _M
