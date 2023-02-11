local cjson = require("cjson")

common.check_method("PATCH")
--校验是否被锁定
common.check_locked()

-- 校验是否登录
local session = libaccess.get_session_info()
if session == nil then
    ngx.log(ngx.NOTICE, "invalid cookie")
    common.sendResp(ngx.HTTP_UNAUTHORIZED, "application/json", g_not_logged_in, "please login")
    return ngx.exit(ngx.HTTP_UNAUTHORIZED)
end

libdynamic.set_upstream("usermanager")
ngx.req.read_body()
local body = ngx.req.get_body_data()
local res, err = ngx.location.capture("/internal/change", { method=ngx.HTTP_POST, body=body, ctx=ngx.ctx })
local ok, resp = pcall(cjson.decode, res.body)
if resp and resp.status == g_error_pass_or_user then
    libaccess.handleLockResp(resp)
else
    common.sendRespByCapture(res)
end
