// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package main to start edge-manager server
package main

import (
	"flag"
	"fmt"

	"edge-manager/pkg/common"
	"edge-manager/pkg/database"

	"github.com/gin-gonic/gin"
	"huawei.com/mindx/common/hwlog"
)

const (
	portConst      = 8101
	runLogFile     = "/var/log/mindx-edge/edge-manager/run.log"
	operateLogFile = "/var/log/mindx-edge/edge-manager/operate.log"
)

var (
	serverRunConf = &hwlog.LogConfig{LogFileName: runLogFile}
	serverOpConf  = &hwlog.LogConfig{LogFileName: operateLogFile}
	// BuildName the program name
	BuildName string
	// BuildVersion the program version
	BuildVersion string
	port         int
	ip           string
	version      bool
)

func main() {
	flag.Parse()
	if version {
		fmt.Printf("%s version: %s\n", BuildName, BuildVersion)
		return
	}
	if err := common.InitHwlogger(serverRunConf, serverOpConf); err != nil {
		fmt.Printf("initialize hwlog failed, %s\n.", err.Error())
		return
	}
	if err := common.BaseParamValid(port, ip); err != nil {
		hwlog.RunLog.Error(err)
		return
	}
	if err := initResourse(); err != nil {
		return
	}
	r := initGin()
	hwlog.RunLog.Info("start http server now...")
	r.Run()
}

func init() {
	flag.IntVar(&port, "port", portConst,
		"The server port of the http service,range[1025-40000]")
	flag.StringVar(&ip, "ip", "",
		"The listen ip of the service,0.0.0.0 is not recommended when install on Multi-NIC host")
	flag.BoolVar(&version, "version", false, "Output the program version")

	// hwOpLog configuration
	flag.IntVar(&serverOpConf.LogLevel, "operateLogLevel", 0,
		"Operation log level, -1-debug, 0-info, 1-warning, 2-error, 3-dpanic, 4-panic, 5-fatal (default 0)")
	flag.IntVar(&serverOpConf.MaxAge, "operateLogMaxAge", hwlog.DefaultMinSaveAge,
		"Maximum number of days for backup operation log files, must be greater than or equal to 7 days")
	flag.StringVar(&serverOpConf.LogFileName, "operateLogFile", operateLogFile,
		"Operation log file path. If the file size exceeds 20MB, will be rotated")
	flag.IntVar(&serverOpConf.MaxBackups, "operateLogMaxBackups", hwlog.DefaultMaxBackups,
		"Maximum number of backup operation logs, range (0, 30]")

	// hwRunLog configuration
	flag.IntVar(&serverRunConf.LogLevel, "runLogLevel", 0,
		"Run log level, -1-debug, 0-info, 1-warning, 2-error, 3-dpanic, 4-panic, 5-fatal (default 0)")
	flag.IntVar(&serverRunConf.MaxAge, "runLogMaxAge", hwlog.DefaultMinSaveAge,
		"Maximum number of days for backup run log files, must be greater than or equal to 7 days")
	flag.StringVar(&serverRunConf.LogFileName, "runLogFile", runLogFile,
		"Run log file path. If the file size exceeds 20MB, will be rotated")
	flag.IntVar(&serverRunConf.MaxBackups, "runLogMaxBackups", hwlog.DefaultMaxBackups,
		"Maximum number of backup run logs, range (0, 30]")
}

func initGin() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.GET("/edgemanager/v1/version", versionQuery)
	return r
}

func initResourse() error {
	if err := database.InitDB(); err != nil {
		hwlog.RunLog.Info("init database failed")
		return err
	}
	return nil
}

func versionQuery(c *gin.Context) {
	msg := fmt.Sprintf("%s version: %s", BuildName, BuildVersion)
	hwlog.OpLog.Infof("query edge manager version: %s successfully", msg)
	common.ConstructResp(c, common.Success, "", msg)
}
