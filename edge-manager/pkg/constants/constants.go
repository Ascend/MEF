// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package constants for edge-manager
package constants

import (
	"time"

	"huawei.com/mindxedge/base/common"
)

const (
	// ServerCertPath server cert path
	ServerCertPath = "/home/data/config/mef-certs/edge-manager.crt"
	// ServerKeyPath server encrypt key path
	ServerKeyPath = "/home/data/config/mef-certs/edge-manager.key"
	// RootCaPath root ca path
	RootCaPath = "/home/data/inner-root-ca/RootCA.crt"
)

const (
	// LogUploadUrl resource for uploading log
	LogUploadUrl = "/logmgmt/dump/upload"
	// ResLogDumpTask resource for task of log dumping
	ResLogDumpTask = "/logmgmt/dump/task"
	// ResLogDumpError resource for error of log dumping
	ResLogDumpError = "/logmgmt/dump/error"
	// LogDumpUrlPrefix prefix for url of log dumping
	LogDumpUrlPrefix = "/edgemanager/v1/logmgmt/dump"
	// ResTask resource for task
	ResTask = "/task"
	// ResDownload resource for downloading
	ResDownload = "/download"
	// EdgeNodesTarGzFileName edgeNodes.tar.gz
	EdgeNodesTarGzFileName = "edgeNodes.tar.gz"

	// LogDumpTempDir the temp dir for dumping log
	LogDumpTempDir = "/home/MEFCenter/mef_logcollect/temp"
	// LogDumpPublicDir the public dir for dumping log
	LogDumpPublicDir = "/home/MEFCenter/mef_logcollect/public"
	// taskIdRegexpStr task id regexp
	taskIdRegexpStr = `[-_a-zA-Z0-9.]{1,128}`
	// DumpMultiNodesLogTaskName the single node task name
	DumpMultiNodesLogTaskName = `dumpMultiNodesLog`
	// DumpSingleNodeLogTaskName the multiple nodes task name
	DumpSingleNodeLogTaskName = `dumpSingleNodeLog`
	// SingleNodeTaskIdRegexpStr the single node task regexp
	SingleNodeTaskIdRegexpStr = "^" + DumpSingleNodeLogTaskName + taskIdRegexpStr + "$"
	// LogUploadMaxSize max size for single node uploading log
	LogUploadMaxSize = 200 * common.MB
)

const (
	// NodeSerialNumber resource for serial number of node
	NodeSerialNumber = "nodeSerialNumber"
	// LogManagerName LogManagerName
	LogManagerName = "LogManager"
)

// const for query interface
const (
	IdKey       = "id"
	SnKey       = "sn"
	KeySymbol   = "key"
	ValueSymbol = "value"
)

// const for init server
const (
	ServerInitRetryInterval = 5 * time.Second
	ServerInitRetryCount    = 3
)
