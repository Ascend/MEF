// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nginxmgr this package is for manager the nginx
package nginxmgr

import (
	"bytes"
	"fmt"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"

	"huawei.com/mindxedge/base/common"

	"nginx-manager/pkg/nginxcom"
)

const (
	nginxDefaultConfigPath  = "/home/MEFCenter/conf/nginx_default.conf"
	nginxResolverConfigPath = "/home/MEFCenter/conf/nginx_resolver.conf"
)

type nginxConfUpdater struct {
	confItems []nginxcom.NginxConfItem
	confPath  string
	icsConf   func() string
}

// NewNginxConfUpdater create an updater to modify nginx configuration file
func NewNginxConfUpdater(confItems []nginxcom.NginxConfItem) (*nginxConfUpdater, error) {
	enableResolverVal, err := nginxcom.GetEnvManager().Get(nginxcom.EnableResolverKey)
	if err != nil {
		return nil, err
	}
	config := nginxConfUpdater{}
	if enableResolverVal == "true" {
		config.confPath = nginxResolverConfigPath
		config.icsConf = getIcsResolverConfContent
	} else {
		config.confPath = nginxDefaultConfigPath
		config.icsConf = getIcsConfContent
	}
	if err := prepareRootCert(nginxcom.IcsCaPath, common.IcsCertName, "IcsRoot"); err != nil {
		config.icsConf = func() string { return "" }
	}
	config.confItems = confItems
	return &config, nil
}

func loadConf(path string) ([]byte, error) {
	b, err := utils.LoadFile(path)
	if err != nil {
		hwlog.RunLog.Errorf("failed to read file. error:%s", err.Error())
		return nil, fmt.Errorf("failed to read file. error:%s", err.Error())
	}
	return b, nil
}

func calculatePipeCount() (int, error) {
	content, err := loadConf(nginxcom.NginxConfigPath)
	if err != nil {
		return 0, err
	}
	return bytes.Count(content, []byte(nginxcom.ClientPipePrefix)), nil
}

// Update do the modify nginx configuration file job
func (n *nginxConfUpdater) Update() error {
	content, err := loadConf(n.confPath)
	if err != nil {
		return err
	}
	return n.updateUrl(content)
}

func (n *nginxConfUpdater) updateUrl(content []byte) error {
	for _, conf := range n.confItems {
		content = bytes.ReplaceAll(content, []byte(conf.From), []byte(conf.To))
	}
	content = bytes.ReplaceAll(content, []byte("$icsConf"), []byte(n.icsConf()))

	err := common.WriteData(nginxcom.NginxConfigPath, content)
	if err != nil {
		hwlog.RunLog.Errorf("writeFile failed. error:%s", err.Error())
		return fmt.Errorf("writeFile failed. error:%s", err.Error())
	}
	return nil
}

func getIcsConfContent() string {
	return `location /icsmanager {
            proxy_set_header X-Forwarded-For $remote_addr;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_pass_request_headers on;
            proxy_pass_request_body on;
            proxy_request_buffering off;
            proxy_ssl_trusted_certificate /home/data/config/mef-certs/ics-root.crt;
            proxy_ssl_verify on;
            proxy_ssl_session_reuse on;
            proxy_ssl_protocols TLSv1.2 TLSv1.3;
            proxy_ssl_ciphers "ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384";
            proxy_pass https://ascend-ics-manager.mef-center.svc.cluster.local:8111;
        }

        location /icscertmgnt {
            proxy_set_header X-Forwarded-For $remote_addr;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_pass_request_headers on;
            proxy_pass_request_body on;
            proxy_request_buffering off;
            proxy_ssl_trusted_certificate /home/data/config/mef-certs/ics-root.crt;
            proxy_ssl_verify on;
            proxy_ssl_session_reuse on;
            proxy_ssl_protocols TLSv1.2 TLSv1.3;
            proxy_ssl_ciphers "ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384";
            proxy_pass https://ascend-ics-manager.mef-center.svc.cluster.local:8112;
        }`
}

func getIcsResolverConfContent() string {
	return `location /icsmanager {
            proxy_set_header X-Forwarded-For $remote_addr;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_pass_request_headers on;
            proxy_pass_request_body on;
            proxy_request_buffering off;
            proxy_ssl_trusted_certificate /home/data/config/mef-certs/ics-root.crt;
            proxy_ssl_verify on;
            proxy_ssl_session_reuse on;
            proxy_ssl_protocols TLSv1.2 TLSv1.3;
            proxy_ssl_ciphers "ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384";
			set $IcsMgrSvc https://ascend-ics-manager.mef-center.svc.cluster.local:8111;
            proxy_pass https://$IcsMgrSvc:8111;
        }

        location /icscertmgnt {
            proxy_set_header X-Forwarded-For $remote_addr;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_pass_request_headers on;
            proxy_pass_request_body on;
            proxy_request_buffering off;
            proxy_ssl_trusted_certificate /home/data/config/mef-certs/ics-root.crt;
            proxy_ssl_verify on;
            proxy_ssl_session_reuse on;
            proxy_ssl_protocols TLSv1.2 TLSv1.3;
            proxy_ssl_ciphers "ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384";
            set $IcsMgrSvc https://ascend-ics-manager.mef-center.svc.cluster.local:8112;
            proxy_pass https://$IcsMgrSvc:8112;
        }`
}
