-- Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved
-- apigw: select target ip address for macro service

local balancer = require("ngx.balancer")
balancer.set_current_peer(ngx.ctx.target_ip, ngx.ctx.target_port)
