// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nginxcom this file is for common constant or method
package nginxcom

// EnvEntry key-val pair
type EnvEntry struct {
	EnvKey string
	EnvVal string
}

// NginxConfItem nginx replace item info
type NginxConfItem struct {
	Key  string
	From string
	To   string
}
