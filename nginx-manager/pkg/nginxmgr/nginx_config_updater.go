// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nginxmgr this package is for manager the nginx
package nginxmgr

import (
	"bytes"
	"fmt"
	"os"
	"strconv"

	"huawei.com/mindx/common/checker"
	"huawei.com/mindx/common/fileutils"
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
	if err := prepareRootCert(nginxcom.IcsCaPath, common.IcsCertName, false); err != nil {
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
	for _, conf := range n.confItems {
		content = bytes.ReplaceAll(content, []byte(conf.From), []byte(conf.To))
	}
	icsContent := n.icsConf()
	content = bytes.ReplaceAll(content, []byte("$icsConf"), []byte(icsContent))

	err = fileutils.WriteData(nginxcom.NginxConfigPath, content)
	if err != nil {
		hwlog.RunLog.Errorf("writeFile failed. error:%s", err.Error())
		return fmt.Errorf("writeFile failed. error:%s", err.Error())
	}
	return nil
}

func getIcsConfContent() string {
	data, err := getIcsData()
	if err != nil {
		return ""
	}
	res := fmt.Sprintf(`location /icsmanager {
            proxy_set_header X-Forwarded-For $remote_addr;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_pass_request_headers on;
            proxy_pass_request_body on;
            proxy_request_buffering off;
			proxy_ssl_certificate /home/data/config/mef-certs/nginx-manager.crt;
			proxy_ssl_certificate_key /home/MEFCenter/pipe/client_pipe_5;
            proxy_ssl_trusted_certificate /home/data/config/mef-certs/ics-root.crt;
            proxy_ssl_verify on;
            proxy_ssl_session_reuse on;
            proxy_ssl_protocols TLSv1.3;
            proxy_ssl_ciphers "ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384";
            proxy_pass https://ascend-ics-manager.ics-center.svc.cluster.local:%v;
			
			location /icsmanager/v1/inclearning/label/upload {
                limit_conn per_addr_upload_conn_zone 1;
                limit_conn global_upload_conn_zone 10;
                client_max_body_size 500m;
                client_body_timeout   600;
                proxy_http_version 1.1;
                proxy_request_buffering off;

                proxy_pass https://ascend-ics-manager.ics-center.svc.cluster.local:%v;
            }
            location /icsmanager/v1/inclearning/label/download {
                limit_conn global_download_conn_zone 1;
                proxy_send_timeout 7200;
                proxy_pass https://ascend-ics-manager.ics-center.svc.cluster.local:%v;
            }
        }

        location /icscertmgnt {
            proxy_set_header X-Forwarded-For $remote_addr;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_pass_request_headers on;
            proxy_pass_request_body on;
            proxy_request_buffering off;
			proxy_ssl_certificate /home/data/config/mef-certs/nginx-manager.crt;
			proxy_ssl_certificate_key /home/MEFCenter/pipe/client_pipe_6;
            proxy_ssl_trusted_certificate /home/data/config/mef-certs/ics-root.crt;
            proxy_ssl_verify on;
            proxy_ssl_session_reuse on;
            proxy_ssl_protocols TLSv1.3;
            proxy_ssl_ciphers "ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384";
            proxy_pass https://ascend-ics-cert-manager.ics-center.svc.cluster.local:%v;
        }`, data.icsPort, data.icsPort, data.icsPort, data.icsCertPort)
	return res
}

func getIcsResolverConfContent() string {
	data, err := getIcsData()
	if err != nil {
		return ""
	}
	return fmt.Sprintf(`location /icsmanager {
            proxy_set_header X-Forwarded-For $remote_addr;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_pass_request_headers on;
            proxy_pass_request_body on;
            proxy_request_buffering off;
			proxy_ssl_certificate /home/data/config/mef-certs/nginx-manager.crt;
			proxy_ssl_certificate_key /home/MEFCenter/pipe/client_pipe_5;
            proxy_ssl_trusted_certificate /home/data/config/mef-certs/ics-root.crt;
            proxy_ssl_verify on;
            proxy_ssl_session_reuse on;
            proxy_ssl_protocols TLSv1.3;
            proxy_ssl_ciphers "ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384";
			set $IcsMgrSvc https://ascend-ics-manager.ics-center.svc.cluster.local;
            proxy_pass https://$IcsMgrSvc:%v;
			location /icsmanager/v1/inclearning/label/upload {
                limit_conn per_addr_upload_conn_zone 1;
                limit_conn global_upload_conn_zone 10;
                client_max_body_size 500m;
                client_body_timeout   600;
                proxy_http_version 1.1;
                proxy_request_buffering off;
				set $IcsMgrSvc https://ascend-ics-manager.ics-center.svc.cluster.local;
                proxy_pass https://$IcsMgrSvc:%v;}
			location /icsmanager/v1/inclearning/label/download {
                limit_conn global_download_conn_zone 1;
                proxy_send_timeout 7200;
				set $IcsMgrSvc https://ascend-ics-manager.ics-center.svc.cluster.local;
                proxy_pass https://$IcsMgrSvc:%v;}
        }
        location /icscertmgnt {
            proxy_set_header X-Forwarded-For $remote_addr;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_pass_request_headers on;
            proxy_pass_request_body on;
            proxy_request_buffering off;
			proxy_ssl_certificate /home/data/config/mef-certs/nginx-manager.crt;
			proxy_ssl_certificate_key /home/MEFCenter/pipe/client_pipe_6;
            proxy_ssl_trusted_certificate /home/data/config/mef-certs/ics-root.crt;
            proxy_ssl_verify on;
            proxy_ssl_session_reuse on;
            proxy_ssl_protocols TLSv1.3;
            proxy_ssl_ciphers "ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384";
            set $IcsMgrSvc https://ascend-ics-cert-manager.ics-center.svc.cluster.local;
            proxy_pass https://$IcsMgrSvc:%v;}`, data.icsPort, data.icsPort, data.icsPort, data.icsCertPort)
}

type icsData struct {
	icsPort     int
	icsCertPort int
}

func getIcsData() (icsData, error) {
	port, err := strconv.Atoi(os.Getenv("IcsPort"))
	if err != nil {
		hwlog.RunLog.Error("cannot convert ics port")
		return icsData{}, err
	}
	if res := checker.GetIntChecker("", common.MinPort, common.MaxPort, true).Check(port); !res.Result {
		hwlog.RunLog.Errorf("ics port %d is not in [%d, %d]", port, common.MinPort, common.MaxPort)
		return icsData{}, err
	}
	certPort, err := strconv.Atoi(os.Getenv("IcsCertPort"))
	if err != nil {
		hwlog.RunLog.Error("cannot convert ics cert port")
		return icsData{}, err
	}
	if res := checker.GetIntChecker("", common.MinPort, common.MaxPort, true).Check(certPort); !res.Result {
		hwlog.RunLog.Errorf("ics cert port %d is not in [%d, %d]", certPort, common.MinPort, common.MaxPort)
		return icsData{}, err
	}
	return icsData{
		icsCertPort: certPort,
		icsPort:     port,
	}, nil
}
