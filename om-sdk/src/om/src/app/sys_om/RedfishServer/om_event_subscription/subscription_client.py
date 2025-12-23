#!/usr/bin/python3
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.

import os
from functools import partial

from urllib3 import PoolManager

from common.file_utils import FileCopy
from common.kmc_lib.tlsconfig import TlsConfig
from om_event_subscription.subscription_mgr import SubscriptionsCertMgr


def create_clients():
    """
    功能描述：创建上报告警事件客户端
    返回值：cert_crl_manager(同时具有证书和吊销列表客户端)
        cert_manager(只有证书的客户端)
    """
    cert_crl_manager = None
    cert_manager = None
    # 1. 获取订阅的证书和吊销列表
    subs_cert = SubscriptionsCertMgr().get_first()

    # 2. 根据证书创建ssl_context，对于没有crl的，创建ssl_context时，不加VERIFY_CRL_CHECK_CHAIN
    cert_crl_ssl_ctx, cert_ssl_ctx = get_ssl_ctx_by_cert_crl(subs_cert)

    # 3. 根据ssl上下文 创建连接池
    if cert_crl_ssl_ctx:
        cert_crl_manager = partial(PoolManager, ssl_context=cert_crl_ssl_ctx, assert_hostname=False, timeout=20)

    if cert_ssl_ctx:
        cert_manager = partial(PoolManager, ssl_context=cert_ssl_ctx, assert_hostname=False, timeout=20)
    return cert_crl_manager, cert_manager


def get_ssl_ctx_by_cert_crl(cert_with_crl):
    tmp_cert_file = "/run/web/tmp_subs_crt.crt"
    tmp_crl_file = "/run/web/tmp_subs_crl.crl"

    cert_crl_ssl_ctx = None
    cert_ssl_ctx = None

    flags = os.O_WRONLY | os.O_CREAT | os.O_TRUNC

    with os.fdopen(os.open(tmp_cert_file, flags=flags, mode=0o600), "w") as tmp_file:
        tmp_file.write(cert_with_crl.cert_contents)

    if not cert_with_crl.crl_contents:
        res, ssl_ctx = TlsConfig.get_client_ssl_context(tmp_cert_file)
        FileCopy.remove_path(tmp_cert_file)
        if res:
            cert_ssl_ctx = ssl_ctx
        return cert_crl_ssl_ctx, cert_ssl_ctx

    with os.fdopen(os.open(tmp_crl_file, flags=flags, mode=0o600), "w") as tmp_file:
        tmp_file.write(cert_with_crl.crl_contents)

    res, ssl_ctx = TlsConfig.get_client_ssl_context(tmp_cert_file, tmp_crl_file)
    for file in tmp_cert_file, tmp_crl_file:
        FileCopy.remove_path(file)
    if res:
        cert_crl_ssl_ctx = ssl_ctx
    return cert_crl_ssl_ctx, cert_ssl_ctx
