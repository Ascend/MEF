// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package innercommands
package innercommands

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"syscall"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/backuputils"
	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/kmc"
	"huawei.com/mindx/common/test"
	"huawei.com/mindx/common/x509/certutils"

	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/installer/edgectl/common"
)

var resolveIP = net.IPv4(127, 0, 0, 1)

const testKeyData = "12345"

func TestPrepareEdgeCoreFlow(t *testing.T) {
	p1 := gomonkey.ApplyPrivateMethod(&PrepareEdgecoreFlow{}, "prepareConfig",
		func() error { return nil }).
		ApplyFuncReturn(certutils.GetKeyContentWithBackup, []byte(testKeyData), nil).
		ApplyFuncReturn(certutils.GetCertContentWithBackup, []byte{}, nil).
		ApplyFuncReturn(util.StartBackupEdgeCoreDb, context.Background(), nil).
		ApplyPrivateMethod(backuputils.NewBackupFileMgr(""), "BackUp", func() error { return nil }).
		ApplyFuncReturn(filepath.EvalSymlinks, "/tmp", nil)
	defer p1.Reset()

	convey.Convey("test prepare edge core flow successful", t, prepareEdgeCoreSuccess)
	convey.Convey("test prepare edge core flow failed", t, func() {
		convey.Convey("check local host failed", checkLocalHostFailed)
		convey.Convey("create pipe failed", createPipeFailed)
		convey.Convey("start edge core failed", startEdgeCoreFailed)
		convey.Convey("write info failed", writeInfoFailed)
		convey.Convey("delete pipe failed", deletePipeFailed)
		convey.Convey("monitor edge core failed", monitorEdgeCoreFailed)
	})
}

func prepareEdgeCoreSuccess() {
	prepareFlow := &PrepareEdgecoreFlow{}
	p := gomonkey.ApplyFuncReturn(net.ResolveIPAddr,
		&net.IPAddr{IP: resolveIP}, nil).
		ApplyFuncReturn(fileutils.IsLexist, false).
		ApplyFuncReturn(syscall.Mkfifo, nil).
		ApplyFuncReturn(fileutils.DeleteFile, nil).
		ApplyFuncReturn(envutils.RunResidentCmd, 1, nil).
		ApplyFuncReturn(fileutils.IsExist, true).
		ApplyMethodReturn(WriteEdgecoreInfoTask{}, "Run", nil).
		ApplyPrivateMethod(&PrepareEdgecoreFlow{}, "deletePipe", func() error { return nil }).
		ApplyPrivateMethod(&PrepareEdgecoreFlow{}, "minitorEdgecore", func() error { return nil }).
		ApplyPrivateMethod(prepareFlow, "checkDocker", func() error { return nil }).
		ApplyPrivateMethod(prepareFlow, "addPortLimitRule", func() error { return nil })
	defer p.Reset()
	err := PrepareEdgecoreCmd().Execute(&common.Context{})
	convey.So(err, convey.ShouldBeNil)
	PrepareEdgecoreCmd().PrintOpLogOk("root", "localhost")
}

func checkLocalHostFailed() {
	convey.Convey("check localhost ip failed", func() {
		p := gomonkey.ApplyFuncReturn(net.ResolveIPAddr, nil, test.ErrTest).
			ApplyPrivateMethod(&PrepareEdgecoreFlow{}, "deletePipeOnce", func() error { return nil })
		defer p.Reset()
		err := NewPrepareEdgecore().Run()
		expectErr := errors.New("check localhost ip failed")
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("localhost ip is incorrect", func() {
		p := gomonkey.ApplyFuncReturn(net.ResolveIPAddr,
			&net.IPAddr{IP: net.IP{49}}, nil).
			ApplyPrivateMethod(&PrepareEdgecoreFlow{}, "deletePipeOnce", func() error { return nil })
		defer p.Reset()
		err := NewPrepareEdgecore().Run()
		expectErr := errors.New("localhost ip is incorrect")
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}

func createPipeFailed() {
	prepareFlow := &PrepareEdgecoreFlow{}
	p := gomonkey.ApplyFuncReturn(net.ResolveIPAddr,
		&net.IPAddr{IP: resolveIP}, nil).
		ApplyPrivateMethod(prepareFlow, "checkDocker", func() error { return nil })
	defer p.Reset()

	convey.Convey("pipe file exists and delete it failed", func() {
		p1 := gomonkey.ApplyFuncSeq(fileutils.IsLexist, []gomonkey.OutputCell{
			{Values: gomonkey.Params{true}},
			{Values: gomonkey.Params{true}},
		}).
			ApplyFuncSeq(fileutils.DeleteFile, []gomonkey.OutputCell{
				{Values: gomonkey.Params{test.ErrTest}},
				{Values: gomonkey.Params{test.ErrTest}},
			})
		defer p1.Reset()
		err := NewPrepareEdgecore().Run()
		expectErr := errors.New("pipe file exists and delete it failed")
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("create pipe failed", func() {
		p2 := gomonkey.ApplyFuncReturn(fileutils.IsLexist, false).
			ApplyFuncReturn(syscall.Mkfifo, test.ErrTest).
			ApplyPrivateMethod(&PrepareEdgecoreFlow{}, "deletePipeOnce", func() error { return nil })
		defer p2.Reset()
		err := NewPrepareEdgecore().Run()
		expectErr := errors.New("create pipe failed")
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}

func startEdgeCoreFailed() {
	prepareFlow := &PrepareEdgecoreFlow{}
	p := gomonkey.ApplyPrivateMethod(&PrepareEdgecoreFlow{}, "checkLocalHost", func() error { return nil }).
		ApplyPrivateMethod(&PrepareEdgecoreFlow{}, "createPipe", func() error { return nil }).
		ApplyPrivateMethod(prepareFlow, "checkDocker", func() error { return nil })
	defer p.Reset()

	convey.Convey("get current work path failed", func() {
		p1 := gomonkey.ApplyFuncReturn(path.GetWorkPathMgr, nil, test.ErrTest).
			ApplyPrivateMethod(&PrepareEdgecoreFlow{}, "deletePipeOnce", func() error { return nil })
		defer p1.Reset()
		err := NewPrepareEdgecore().Run()
		expectErr := fmt.Errorf("get work path manager failed: %s", test.ErrTest.Error())
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("kubelet.sock file exists and delete it failed", func() {
		p2 := gomonkey.ApplyFuncReturn(fileutils.DeleteFile, test.ErrTest).
			ApplyPrivateMethod(&PrepareEdgecoreFlow{}, "deletePipeOnce", func() error { return nil })
		defer p2.Reset()
		err := NewPrepareEdgecore().Run()
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	convey.Convey("start edge core run cmd failed", func() {
		p3 := gomonkey.ApplyFuncReturn(fileutils.DeleteFile, nil).
			ApplyFuncReturn(envutils.RunResidentCmd, 0, test.ErrTest).
			ApplyPrivateMethod(&PrepareEdgecoreFlow{}, "deletePipeOnce", func() error { return nil })
		defer p3.Reset()
		err := NewPrepareEdgecore().Run()
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	convey.Convey("make sure edge core start failed", func() {
		convey.Convey("edge core process does not exist", func() {
			p4 := gomonkey.ApplyFuncReturn(fileutils.DeleteFile, nil).
				ApplyFuncReturn(envutils.RunResidentCmd, 1, nil).
				ApplyFuncReturn(fileutils.IsExist, false).
				ApplyFuncReturn(os.Stat, nil, os.ErrNotExist).
				ApplyPrivateMethod(&PrepareEdgecoreFlow{}, "deletePipeOnce", func() error { return nil })
			defer p4.Reset()
			err := NewPrepareEdgecore().Run()
			expectErr := fmt.Errorf("edge core process does not exist")
			convey.So(err, convey.ShouldResemble, expectErr)
		})

		convey.Convey("get edge core process status failed", func() {
			p5 := gomonkey.ApplyFuncReturn(fileutils.DeleteFile, nil).
				ApplyFuncReturn(envutils.RunResidentCmd, 1, nil).
				ApplyFuncReturn(fileutils.IsExist, false).
				ApplyFuncReturn(os.Stat, nil, test.ErrTest).
				ApplyPrivateMethod(&PrepareEdgecoreFlow{}, "deletePipeOnce", func() error { return nil })
			defer p5.Reset()
			err := NewPrepareEdgecore().Run()
			expectErr := fmt.Errorf("get edge core process status failed")
			convey.So(err, convey.ShouldResemble, expectErr)
		})
	})
}

func writeInfoFailed() {
	prepareFlow := &PrepareEdgecoreFlow{}
	p := gomonkey.ApplyPrivateMethod(&PrepareEdgecoreFlow{}, "checkLocalHost", func() error { return nil }).
		ApplyPrivateMethod(&PrepareEdgecoreFlow{}, "createPipe", func() error { return nil }).
		ApplyPrivateMethod(&PrepareEdgecoreFlow{}, "startEdgecore", func() error { return nil }).
		ApplyPrivateMethod(prepareFlow, "checkDocker", func() error { return nil })
	defer p.Reset()

	convey.Convey("write edge core info failed, and delete edge core pipePath failed", func() {
		p1 := gomonkey.ApplyMethodReturn(WriteEdgecoreInfoTask{}, "Run", test.ErrTest).
			ApplyFuncReturn(fileutils.DeleteFile, test.ErrTest)
		defer p1.Reset()
		err := NewPrepareEdgecore().Run()
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})
}

func deletePipeFailed() {
	prepareFlow := &PrepareEdgecoreFlow{}
	p := gomonkey.ApplyPrivateMethod(&PrepareEdgecoreFlow{}, "checkLocalHost", func() error { return nil }).
		ApplyPrivateMethod(&PrepareEdgecoreFlow{}, "createPipe", func() error { return nil }).
		ApplyPrivateMethod(&PrepareEdgecoreFlow{}, "startEdgecore", func() error { return nil }).
		ApplyPrivateMethod(&PrepareEdgecoreFlow{}, "writeInfo", func() error { return nil }).
		ApplyPrivateMethod(prepareFlow, "checkDocker", func() error { return nil }).
		ApplyPrivateMethod(prepareFlow, "addPortLimitRule", func() error { return nil })
	defer p.Reset()

	convey.Convey("delete edge core pipePath failed", func() {
		p1 := gomonkey.ApplyFuncReturn(fileutils.DeleteFile, test.ErrTest)
		defer p1.Reset()
		err := NewPrepareEdgecore().Run()
		convey.So(err, convey.ShouldBeNil)
	})
}

func monitorEdgeCoreFailed() {
	prepareFlow := &PrepareEdgecoreFlow{}
	p := gomonkey.ApplyPrivateMethod(&PrepareEdgecoreFlow{}, "checkLocalHost", func() error { return nil }).
		ApplyPrivateMethod(&PrepareEdgecoreFlow{}, "createPipe", func() error { return nil }).
		ApplyPrivateMethod(&PrepareEdgecoreFlow{}, "startEdgecore", func() error { return nil }).
		ApplyPrivateMethod(&PrepareEdgecoreFlow{}, "writeInfo", func() error { return nil }).
		ApplyPrivateMethod(&PrepareEdgecoreFlow{}, "deletePipe", func() error { return nil }).
		ApplyPrivateMethod(prepareFlow, "checkDocker", func() error { return nil }).
		ApplyPrivateMethod(prepareFlow, "addPortLimitRule", func() error { return nil })
	defer p.Reset()

	convey.Convey("check process exists failed", func() {
		p1 := gomonkey.ApplyFuncReturn(checkProcessExists, test.ErrTest)
		defer p1.Reset()
		err := NewPrepareEdgecore().Run()
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("check edge core conn failed", func() {
		convey.Convey("get edge core's port by its pid failed", func() {
			p2 := gomonkey.ApplyFuncReturn(checkProcessExists, nil).
				ApplyMethodReturn(&envutils.ProcessPortMgr{}, "GetPortByPid", nil, test.ErrTest)
			defer p2.Reset()
			err := NewPrepareEdgecore().Run()
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("get edge core's port is nil", func() {
			p3 := gomonkey.ApplyFuncReturn(checkProcessExists, nil).
				ApplyMethodReturn(&envutils.ProcessPortMgr{}, "GetPortByPid", nil, nil)
			defer p3.Reset()
			err := NewPrepareEdgecore().Run()
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("established port and ports are not deep equal", func() {
			p4 := gomonkey.ApplyFuncReturn(checkProcessExists, nil).
				ApplyMethodReturn(&envutils.ProcessPortMgr{}, "GetPortByPid", []int{100}, nil).
				ApplyFuncReturn(reflect.DeepEqual, false)
			defer p4.Reset()
			err := NewPrepareEdgecore().Run()
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestWriteEdgeCoreInfoTask(t *testing.T) {
	p := gomonkey.ApplyFuncReturn(certutils.GetKeyContentWithBackup, []byte(testKeyData), nil).
		ApplyFuncReturn(certutils.GetCertContentWithBackup, []byte{}, nil).
		ApplyPrivateMethod(backuputils.NewBackupFileMgr(""), "BackUp", func() error { return nil })
	defer p.Reset()

	convey.Convey("test write edge core info task successful", t, writeEdgeCoreInfoTaskSuccess)
	convey.Convey("test write edge core info task failed", t, func() {
		convey.Convey("write edge core info run failed", writeEdgeCoreInfoRunFailed)
		convey.Convey("prepare key data failed", prepareKeyDataFailed)
		convey.Convey("write info data to pipe failed", writeInfoDataToPipe)
	})
}

func writeEdgeCoreInfoTaskSuccess() {
	writeLen := 5
	writeTask := NewWriteEdgecoreInfoTask("/run/edgecore.pipe", 1)
	p := gomonkey.ApplyFuncReturn(fileutils.IsExist, true).
		ApplyFuncReturn(kmc.InitKmcCfg, nil).
		ApplyFuncReturn(util.GetKmcConfig, &kmc.SubConfig{}, nil).
		ApplyFuncReturn(fileutils.LoadFile, []byte{}, nil).
		ApplyFuncReturn(fileutils.CheckOriginPath, "", nil).
		ApplyFuncReturn(os.Stat, nil, nil).
		ApplyFuncReturn(os.OpenFile, &os.File{}, nil).
		ApplyMethodReturn(&fileutils.FileLinkChecker{}, "Check", nil).
		ApplyMethodReturn(&os.File{}, "Write", writeLen, nil).
		ApplyMethodReturn(&os.File{}, "Close", nil)
	defer p.Reset()
	err := writeTask.Run()
	convey.So(err, convey.ShouldBeNil)
}

func writeEdgeCoreInfoRunFailed() {
	writeTask := NewWriteEdgecoreInfoTask("/run/edgecore.pipe", 1)
	convey.Convey("get install root dir failed", func() {
		p := gomonkey.ApplyFuncReturn(path.GetConfigPathMgr, nil, test.ErrTest)
		defer p.Reset()
		err := writeTask.Run()
		expectErr := errors.New("get config path manager failed")
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("edge core tls key file not exist", func() {
		p := gomonkey.ApplyFuncReturn(fileutils.IsExist, false)
		defer p.Reset()
		err := writeTask.Run()
		expectErr := fmt.Errorf("edgecore tls key file not exist")
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("get kmc config failed", func() {
		p := gomonkey.ApplyFuncReturn(fileutils.IsExist, true).
			ApplyFuncReturn(kmc.InitKmcCfg, test.ErrTest).
			ApplyFuncReturn(util.GetKmcConfig, nil, test.ErrTest)
		defer p.Reset()
		err := writeTask.Run()
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})
}

func prepareKeyDataFailed() {
	writeTask := NewWriteEdgecoreInfoTask("/run/edgecore.pipe", 1)
	p := gomonkey.ApplyFuncReturn(fileutils.IsExist, true).
		ApplyFuncReturn(kmc.InitKmcCfg, nil).
		ApplyFuncReturn(util.GetKmcConfig, &kmc.SubConfig{}, nil)
	defer p.Reset()

	convey.Convey("load edge core tls key file failed", func() {
		p1 := gomonkey.ApplyFuncReturn(certutils.GetKeyContentWithBackup, nil, test.ErrTest)
		defer p1.Reset()
		err := writeTask.Run()
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})
}

func writeInfoDataToPipe() {
	writeTask := NewWriteEdgecoreInfoTask("/run/edgecore.pipe", 1)
	expectErr := errors.New("write key into pipe file failed")
	p := gomonkey.ApplyFuncReturn(fileutils.IsExist, true).
		ApplyFuncReturn(kmc.InitKmcCfg, nil).
		ApplyFuncReturn(util.GetKmcConfig, &kmc.SubConfig{}, nil).
		ApplyFuncReturn(fileutils.LoadFile, []byte{}, nil).
		ApplyFuncReturn(kmc.DecryptContent, []byte{1, 2, 3, 4, 5}, nil)
	defer p.Reset()

	convey.Convey("do pipe write operation failed", func() {
		convey.Convey("check pipe path failed", func() {
			p1 := gomonkey.ApplyFuncReturn(fileutils.CheckOriginPath, "", test.ErrTest)
			defer p1.Reset()
			err := writeTask.Run()
			convey.So(err, convey.ShouldResemble, expectErr)
		})

		convey.Convey("check process exists failed", func() {
			p2 := gomonkey.ApplyFuncReturn(fileutils.CheckOriginPath, "", nil).
				ApplyFuncReturn(os.Stat, nil, os.ErrNotExist)
			defer p2.Reset()
			err := writeTask.Run()
			convey.So(err, convey.ShouldResemble, expectErr)
		})

		convey.Convey("open edge core pipe file failed", func() {
			p3 := gomonkey.ApplyFuncReturn(fileutils.CheckOriginPath, "", nil).
				ApplyFuncReturn(os.Stat, nil, nil).
				ApplyFuncReturn(os.OpenFile, nil, test.ErrTest)
			defer p3.Reset()
			err := writeTask.Run()
			convey.So(err, convey.ShouldResemble, expectErr)
		})

		convey.Convey("write edge core pipe file failed", func() {
			p4 := gomonkey.ApplyFuncReturn(fileutils.CheckOriginPath, "", nil).
				ApplyFuncReturn(os.Stat, nil, nil).
				ApplyFuncReturn(os.OpenFile, &os.File{}, nil).
				ApplyMethodReturn(&fileutils.FileLinkChecker{}, "Check", nil).
				ApplyMethodReturn(&os.File{}, "Write", 0, test.ErrTest).
				ApplyMethodReturn(&os.File{}, "Close", test.ErrTest)
			defer p4.Reset()
			err := writeTask.Run()
			convey.So(err, convey.ShouldResemble, expectErr)
		})

		convey.Convey("write edge core pipe data not correct failed", func() {
			p5 := gomonkey.ApplyFuncReturn(fileutils.CheckOriginPath, "", nil).
				ApplyFuncReturn(os.Stat, nil, nil).
				ApplyFuncReturn(os.OpenFile, &os.File{}, nil).
				ApplyMethodReturn(&fileutils.FileLinkChecker{}, "Check", nil).
				ApplyMethodReturn(&os.File{}, "Write", 0, nil).
				ApplyMethodReturn(&os.File{}, "Close", nil)
			defer p5.Reset()
			err := writeTask.Run()
			convey.So(err, convey.ShouldResemble, expectErr)
		})
	})
}
