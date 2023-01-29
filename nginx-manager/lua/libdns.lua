-- Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved

local resolver = require("resty.dns.resolver")

local _M = {}

function _M.read_file_data(path)
    local res, err = io.open(path, 'r')
    if not res or err then
        return nil, err
    end
    local data = res:read('*all')
    res:close()
    return data, nil
end

function _M.read_dns_servers_from_resolv_file()
    local text = _M.read_file_data('/etc/resolv.conf')

    local captures, it, err
    it, err = ngx.re.gmatch(text, [[^nameserver\s+(\d+?\.\d+?\.\d+?\.\d+$)]], "jomi")

    for captures, err in it do
        if not err then
            g_dns_servers[#g_dns_servers + 1] = captures[1]
            ngx.log(ngx.INFO, "read resolv.conf server = ".. captures[1])
        end
    end
    ngx.log(ngx.INFO, "read resolv.conf success")
end

function _M.is_addr(hostname)
    return ngx.re.find(hostname, [[\d+?\.\d+?\.\d+?\.\d+$]], "jo")
end

function _M.query(hostname)
    local r, err = resolver:new({
        nameservers = g_dns_servers,
        retrans = 3,     -- 3 retransmissions on receive timeout
        timeout = 3000,  -- 3 sec
    })
    if not r or err then
        return nil
    end

    local answers, err = r:query(hostname, {qtype = r.TYPE_A})

    if not answers or answers.errcode or err then
        return nil
    end

    for i, ans in ipairs(answers) do
        if ans.address then
            return ans.address
        end
    end
    return nil
end

function _M.get_addr(hostname)
    if _M.is_addr(hostname) then
        -- hostname本身是ip地址
        return hostname, hostname
    end

    -- 从cache查找
    local addr = g_dns_cache:get(hostname)
    if addr then
        return addr, hostname
    end

    -- 从DNS服务器查询
    addr = _M.query(hostname)
    if addr then
        -- 存到cache
        g_dns_cache:set(hostname, addr, 300)
        ngx.log(ngx.WARN, "DNS query ["..hostname.."] -> ["..addr.."]")
    else
        ngx.log(ngx.ERR, "DNS query ["..hostname.."] failed")
    end
    return addr, hostname
end

return  _M
