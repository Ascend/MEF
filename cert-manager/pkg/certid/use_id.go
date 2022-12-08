// Copyright (c) 2021. Huawei Technologies Co., Ltd. All rights reserved.

// Package certid  cert use id mgr
package certid

const (
	certMgrID      = "1"
	edgeMgrHubSer  = "2"
	edgeMgrHubClt  = "3"
	softwareMgrSer = "4"
	imageMgrSer    = "5"
	resMgrSer      = "6"
	edgeMgrNgx     = "7"
	edgeCoreValid  = "8"
)

var useIdMap = map[string]string{certMgrID: "cert_inner", edgeMgrHubSer: "edge_mgr_hub_ser",
	edgeMgrHubClt: "edge_mgr_hub_clt", softwareMgrSer: "software_mgr_ser", imageMgrSer: "image_mgr_ser",
	resMgrSer: "resource_mgr_ser", edgeMgrNgx: "nginx_ser", edgeCoreValid: "edge_core_valid"}

// CheckUseId check use id if valid
func CheckUseId(id string) bool {
	if _, ok := useIdMap[id]; ok {
		return true
	}
	return false
}

// GetUseIdName get use name with id
func GetUseIdName(id string) string {
	return useIdMap[id]
}
