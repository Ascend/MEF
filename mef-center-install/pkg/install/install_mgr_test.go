// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package install

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"syscall"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/backuputils"
	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"

	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

func DoInstallMgrTest() {
	convey.Convey("doInstall func", DoInstallTest)
	convey.Convey("checkInstalled func", CheckInstalledTest)
	convey.Convey("preCheck func", PreCheckTest)
	convey.Convey("CheckUser func", CheckUserTest)
	convey.Convey("CheckDiskSpace func", CheckDiskSpaceTest)
	convey.Convey("CheckNecessaryTools func", CheckNecessaryToolsTest)
	convey.Convey("prepareMefUser func success situation", PrepareMefUserSuccessTest)
	convey.Convey("prepareMefUser func failed situation", PrepareMefUserCreateUserFailedTest)
	convey.Convey("prepareMefUser func userCheck failed", PrePareMefUserUserCheckFailedTest)
	convey.Convey("prepareK8sLabel func", PrepareK8sLabelTest)
	convey.Convey("prepareComponent func", PrepareComponentLogDirTest)
	convey.Convey("prepareComponentLogDir func", PrepareComponentLogDirRightCheckTest)
	convey.Convey("prepareCerts func", PrepareCertsTest)
	convey.Convey("prepareWorkingDir func", PrepareWorkingDirTest)
	convey.Convey("prepareYaml func", PrepareYamlTest)
	convey.Convey("setInstallJson func", SetInstallJsonTest)
	convey.Convey("componentInstall func", ComponentsInstallTest)
	convey.Convey("setCenterMode func", SetCenterModeTest)
	convey.Convey("prepareInstallPkgDir func", PrepareInstallPkgDirTest)
	convey.Convey("prepareLogDumpDir func", PrepareLogDumpDirTest)
	convey.Convey("prepareConfigDir func", PrepareConfigDirTest)
	convey.Convey("setConfigOwner func", SetConfigOwnerTest)
	convey.Convey("configBackup func", ConfigBackupTest)
}

func DoInstallTest() {
	ins, err := GetSftInstallMgrIns([]string{}, "", "", "")
	convey.So(err, convey.ShouldBeNil)
	convey.Convey("test doInstall func success", func() {
		p := gomonkey.ApplyPrivateMethod(ins, "checkInstalled", func(_ *SftInstallCtl) error { return nil }).
			ApplyPrivateMethod(ins, "preCheck", func(_ *SftInstallCtl) error { return nil }).
			ApplyPrivateMethod(ins, "prepareMefUser", func(_ *SftInstallCtl) error { return nil }).
			ApplyPrivateMethod(ins, "prepareK8sLabel", func(_ *SftInstallCtl) error { return nil }).
			ApplyPrivateMethod(ins, "prepareComponentLogDir", func(_ *SftInstallCtl) error { return nil }).
			ApplyPrivateMethod(ins, "prepareInstallPkgDir", func(_ *SftInstallCtl) error { return nil }).
			ApplyPrivateMethod(ins, "prepareCerts", func(_ *SftInstallCtl) error { return nil }).
			ApplyPrivateMethod(ins, "prepareWorkingDir", func(_ *SftInstallCtl) error { return nil }).
			ApplyPrivateMethod(ins, "prepareConfigDir", func(_ *SftInstallCtl) error { return nil }).
			ApplyPrivateMethod(ins, "prepareYaml", func(_ *SftInstallCtl) error { return nil }).
			ApplyPrivateMethod(ins, "setInstallJson", func(_ *SftInstallCtl) error { return nil }).
			ApplyPrivateMethod(ins, "componentsInstall", func(_ *SftInstallCtl) error { return nil }).
			ApplyPrivateMethod(ins, "prepareLogDumpDir", func(_ *SftInstallCtl) error { return nil }).
			ApplyPrivateMethod(ins, "configBackup", func(_ *SftInstallCtl) error { return nil }).
			ApplyPrivateMethod(ins, "setConfigOwner", func(_ *SftInstallCtl) error { return nil }).
			ApplyPrivateMethod(ins, "setCenterMode", func(_ *SftInstallCtl) error { return nil })
		defer p.Reset()
		convey.So(ins.DoInstall(), convey.ShouldBeNil)
	})

	convey.Convey("test doInstall func checkInstalled failed", func() {
		p := gomonkey.ApplyPrivateMethod(ins, "checkUser", func(_ *SftInstallCtl) error { return nil }).
			ApplyPrivateMethod(ins, "checkNecessaryTools", func(_ *SftInstallCtl) error { return nil }).
			ApplyFuncReturn(os.Stat, nil, nil)
		defer p.Reset()
		convey.So(ins.DoInstall(), convey.ShouldResemble, errors.New("the software has already been installed"))
	})

	k8sMgrIns := &util.K8sLabelMgr{}
	convey.Convey("test doInstall func k8s label exists failed", func() {
		p := gomonkey.ApplyPrivateMethod(ins, "checkUser", func(_ *SftInstallCtl) error { return nil }).
			ApplyPrivateMethod(ins, "checkNecessaryTools", func(_ *SftInstallCtl) error { return nil }).
			ApplyPrivateMethod(k8sMgrIns, "CheckK8sLabel",
				func(_ *util.K8sLabelMgr) (bool, error) { return true, nil })
		defer p.Reset()
		convey.So(ins.DoInstall(), convey.ShouldResemble,
			errors.New("the software has already been installed since k8s label exists"))
	})

	convey.Convey("test doInstall func install failed", func() {
		p := gomonkey.ApplyPrivateMethod(ins, "preCheck", func(_ *SftInstallCtl) error { return ErrTest }).
			ApplyPrivateMethod(ins, "clearAll", func(_ *SftInstallCtl) { return })
		defer p.Reset()
		convey.So(ins.DoInstall(), convey.ShouldResemble, ErrTest)
	})

}

func CheckInstalledTest() {
	k8sMgr := &util.K8sLabelMgr{}
	ins, err := GetSftInstallMgrIns([]string{}, "", "", "")
	convey.So(err, convey.ShouldBeNil)
	convey.Convey("checkInstalled func success", func() {
		p := gomonkey.ApplyMethodReturn(k8sMgr, "CheckK8sLabel", false, nil)
		defer p.Reset()
		convey.So(ins.checkInstalled(), convey.ShouldBeNil)
	})

	convey.Convey("checkInstalled func failed by installed", func() {
		p := gomonkey.ApplyMethodReturn(k8sMgr, "CheckK8sLabel", true, nil)
		defer p.Reset()
		convey.So(ins.checkInstalled(), convey.ShouldResemble,
			errors.New("the software has already been installed since k8s label exists"))
	})

	convey.Convey("checkInstalled func failed", func() {
		p := gomonkey.ApplyFuncReturn(os.Stat, nil, nil)
		defer p.Reset()
		convey.So(ins.checkInstalled(), convey.ShouldResemble,
			errors.New("the software has already been installed"))
	})
}

func PreCheckTest() {
	ins, err := GetSftInstallMgrIns([]string{}, "", "", "")
	convey.So(err, convey.ShouldBeNil)
	convey.Convey("test preCheck func success", func() {
		p := gomonkey.ApplyPrivateMethod(ins, "checkUser", func(_ *SftInstallCtl) error { return nil }).
			ApplyPrivateMethod(ins, "checkDiskSpace", func(_ *SftInstallCtl) error { return nil }).
			ApplyPrivateMethod(ins, "checkInstalled", func(_ *SftInstallCtl) error { return nil }).
			ApplyPrivateMethod(ins, "checkNecessaryTools", func(_ *SftInstallCtl) error { return nil }).
			ApplyMethodReturn(&util.DockerDealer{}, "CheckImageExists", true, nil)
		defer p.Reset()
		convey.So(ins.preCheck(), convey.ShouldBeNil)
	})

	convey.Convey("test preCheck func failed", func() {
		p := gomonkey.ApplyPrivateMethod(ins, "checkUser", func(_ *SftInstallCtl) error { return ErrTest })
		defer p.Reset()
		convey.So(ins.preCheck(), convey.ShouldResemble, ErrTest)
	})
}

func CheckUserTest() {
	ins, err := GetSftInstallMgrIns([]string{}, "", "", "")
	convey.So(err, convey.ShouldBeNil)
	convey.Convey("test CheckUser func check user failed", func() {
		p := gomonkey.ApplyFuncReturn(user.Current, nil, ErrTest)
		defer p.Reset()
		resp := fmt.Errorf("get current user info failed: %s", ErrTest.Error())
		convey.So(ins.checkUser(), convey.ShouldResemble, resp)
	})

	userName := "notExist"
	convey.Convey(fmt.Sprintf("test CheckUser func user is %s", userName), func() {
		userName = "notExist"
		p := gomonkey.ApplyFuncReturn(user.Current, &user.User{Username: userName}, nil)
		defer p.Reset()
		resp := fmt.Errorf("the current user must be root, can not be %s", userName)
		convey.So(ins.checkUser(), convey.ShouldResemble, resp)
	})

	userName = "root"
	convey.Convey(fmt.Sprintf("test CheckUser func user is %s", userName), func() {
		userName = "root"
		p := gomonkey.ApplyFuncReturn(user.Current, &user.User{Username: userName}, nil)
		defer p.Reset()
		convey.So(ins.checkUser(), convey.ShouldBeNil)
	})
}

func CheckDiskSpaceTest() {
	ins, err := GetSftInstallMgrIns([]string{}, "", "", "")
	convey.So(err, convey.ShouldBeNil)
	convey.Convey("test Check Disk Space func get disk free failed", func() {
		p := gomonkey.ApplyFunc(syscall.Statfs, func(_ string, _ *syscall.Statfs_t) error {
			return ErrTest
		})
		defer p.Reset()
		convey.So(ins.checkDiskSpace(), convey.ShouldResemble, errors.New("check install disk space failed"))
	})

	convey.Convey("test Check Disk Space func Disk Free enough", func() {
		p := gomonkey.ApplyFuncReturn(envutils.GetFileDevNum, uint64(1), nil).
			ApplyFunc(syscall.Statfs, func(_ string,
				fs *syscall.Statfs_t) error {
				fs.Bavail = util.InstallDiskSpace
				fs.Bsize = 1
				return nil
			})
		defer p.Reset()
		convey.So(ins.checkDiskSpace(), convey.ShouldBeNil)
	})

	convey.Convey("test Check Disk Space func Disk Free not enough", func() {
		p := gomonkey.ApplyFunc(syscall.Statfs, func(_ string, fs *syscall.Statfs_t) error {
			fs.Bavail = util.InstallDiskSpace - 1
			fs.Bsize = 1
			return nil
		})
		defer p.Reset()
		convey.So(ins.checkDiskSpace(), convey.ShouldResemble, errors.New("check install disk space failed"))
	})
}

func CheckNecessaryToolsTest() {
	ins, err := GetSftInstallMgrIns([]string{}, "", "", "")
	convey.So(err, convey.ShouldBeNil)
	convey.Convey("test CheckNecessaryTools func failed", func() {
		p := gomonkey.ApplyFuncReturn(exec.LookPath, "", ErrTest)
		defer p.Reset()
		convey.So(ins.checkNecessaryTools(), convey.ShouldNotBeNil)
	})

	convey.Convey("test CheckNecessaryTools func success", func() {
		p := gomonkey.ApplyFuncReturn(exec.LookPath, "", nil)
		defer p.Reset()
		convey.So(ins.checkNecessaryTools(), convey.ShouldBeNil)
	})
}

func PrepareMefUserSuccessTest() {
	ins, err := GetSftInstallMgrIns([]string{}, "", "", "")
	convey.So(err, convey.ShouldBeNil)
	convey.Convey("test PrepareMefUser func create user 8000 uid success", func() {
		p := gomonkey.ApplyFuncReturn(exec.LookPath, "", nil).
			ApplyFuncReturn(user.Lookup, nil, ErrTest).
			ApplyFuncReturn(user.LookupGroup, nil, ErrTest).
			ApplyFuncReturn(fileutils.IsExist, false).
			ApplyFuncReturn(user.LookupId, nil, nil).
			ApplyFuncReturn(envutils.RunCommand, "", nil)
		defer p.Reset()
		convey.So(ins.prepareMefUser(), convey.ShouldBeNil)
	})

	convey.Convey("test PrepareMefUser func create user auto uid success", func() {
		p := gomonkey.ApplyFuncReturn(exec.LookPath, "", nil).
			ApplyFuncReturn(user.Lookup, nil, ErrTest).
			ApplyFuncReturn(user.LookupGroup, nil, ErrTest).
			ApplyFuncReturn(fileutils.IsExist, false).
			ApplyFuncReturn(user.LookupId, nil, errors.New("user: unknown userid")).
			ApplyFuncReturn(envutils.RunCommand, "", nil)
		defer p.Reset()
		convey.So(ins.prepareMefUser(), convey.ShouldBeNil)
	})
}

func PrepareMefUserCreateUserFailedTest() {
	ins, err := GetSftInstallMgrIns([]string{}, "", "", "")
	convey.So(err, convey.ShouldBeNil)
	convey.Convey("test PrepareMefUser func create user command exec failed", func() {
		p := gomonkey.ApplyFuncReturn(exec.LookPath, "", nil).
			ApplyFuncReturn(user.Lookup, nil, ErrTest).
			ApplyFuncReturn(user.LookupGroup, nil, ErrTest).
			ApplyFuncReturn(fileutils.IsExist, false).
			ApplyFuncReturn(user.LookupId, nil, errors.New("user: unknown userid")).
			ApplyFuncReturn(envutils.RunCommand, "", ErrTest)
		defer p.Reset()
		convey.So(ins.prepareMefUser(), convey.ShouldResemble, errors.New("exec useradd command failed"))
	})

	convey.Convey("test PrepareMefUser func create user no nologin cmd", func() {
		p := gomonkey.ApplyFuncReturn(exec.LookPath, "", ErrTest)
		defer p.Reset()
		convey.So(ins.prepareMefUser(), convey.ShouldResemble, errors.New("look path of nologin failed"))
	})

	convey.Convey("test PrepareMefUser func create user home dir exists", func() {
		p := gomonkey.ApplyFuncReturn(exec.LookPath, "", nil).
			ApplyFuncReturn(user.Lookup, nil, ErrTest).
			ApplyFuncReturn(user.LookupGroup, nil, ErrTest).
			ApplyFuncReturn(fileutils.IsExist, true)
		defer p.Reset()
		convey.So(ins.prepareMefUser(), convey.ShouldResemble, errors.New("user home dir exists"))
	})

	convey.Convey("test PrepareMefUser func create user user exists but not group", func() {
		p := gomonkey.ApplyFuncReturn(exec.LookPath, "", nil).
			ApplyFuncReturn(user.Lookup, nil, nil).
			ApplyFuncReturn(user.LookupGroup, nil, ErrTest)
		defer p.Reset()
		convey.So(ins.prepareMefUser(), convey.ShouldResemble, errors.New("the user name or group name is in use"))
	})

}

func PrePareMefUserUserCheckFailedTest() {
	ins, err := GetSftInstallMgrIns([]string{}, "", "", "")
	convey.So(err, convey.ShouldBeNil)
	applyFunc := func(_ *envutils.UserMgr) error { return ErrTest }
	convey.Convey("test PrepareMefUser func get groupIds failed", func() {
		p := gomonkey.ApplyFuncReturn(exec.LookPath, "", nil).
			ApplyFuncReturn(user.Lookup, &user.User{}, nil).
			ApplyFuncReturn(user.LookupGroup, &user.Group{Gid: "8000"}, nil).
			ApplyFuncReturn(fileutils.IsExist, false).
			ApplyMethodReturn(&user.User{}, "GroupIds", nil, ErrTest)
		defer p.Reset()
		convey.So(ins.prepareMefUser(), convey.ShouldNotBeNil)
	})

	convey.Convey("test PrepareMefUser func gid not in user group failed", func() {
		p := gomonkey.ApplyFuncReturn(exec.LookPath, "", nil).
			ApplyFuncReturn(user.Lookup, &user.User{}, nil).
			ApplyFuncReturn(user.LookupGroup, &user.Group{Gid: "8000"}, nil).
			ApplyFuncReturn(fileutils.IsExist, false).
			ApplyMethodReturn(&user.User{}, "GroupIds", []string{"9000"}, nil)
		defer p.Reset()
		convey.So(ins.prepareMefUser(), convey.ShouldNotBeNil)
	})

	convey.Convey("test PrepareMefUser func check nologin function failed", func() {
		p := gomonkey.ApplyFuncReturn(exec.LookPath, "", nil).
			ApplyFuncReturn(user.Lookup, &user.User{}, nil).
			ApplyFuncReturn(user.LookupGroup, &user.Group{Gid: "8000"}, nil).
			ApplyFuncReturn(fileutils.IsExist, false).
			ApplyMethodReturn(&user.User{}, "GroupIds", []string{"8000"}, nil).
			ApplyPrivateMethod(&envutils.UserMgr{}, "checkNoLogin", applyFunc)
		defer p.Reset()
		convey.So(ins.prepareMefUser(), convey.ShouldNotBeNil)
	})

	convey.Convey("test PrepareMefUser func check user does not have nologin attribute failed", func() {
		p := gomonkey.ApplyFuncReturn(exec.LookPath, "", nil).
			ApplyFuncReturn(user.Lookup, &user.User{}, nil).
			ApplyFuncReturn(user.LookupGroup, &user.Group{Gid: "8000"}, nil).
			ApplyFuncReturn(fileutils.IsExist, false).
			ApplyMethodReturn(&user.User{}, "GroupIds", []string{"8000"}, nil).
			ApplyPrivateMethod(&envutils.UserMgr{}, "checkNoLogin", applyFunc)
		defer p.Reset()
		convey.So(ins.prepareMefUser(), convey.ShouldNotBeNil)
	})
}

func PrepareK8sLabelTest() {
	ins, err := GetSftInstallMgrIns([]string{}, "", "", "")
	convey.So(err, convey.ShouldBeNil)

	k8sMgr := &util.K8sLabelMgr{}
	convey.Convey("test prepareK8sLabel func success", func() {
		p := gomonkey.ApplyMethodReturn(k8sMgr, "PrepareK8sLabel", nil)
		defer p.Reset()
		convey.So(ins.prepareK8sLabel(), convey.ShouldBeNil)
	})

	convey.Convey("test prepareK8sLabel func failed", func() {
		p := gomonkey.ApplyMethodReturn(k8sMgr, "PrepareK8sLabel", ErrTest)
		defer p.Reset()
		convey.So(ins.prepareK8sLabel(), convey.ShouldResemble, ErrTest)
	})
}

func PrepareComponentLogDirTest() {
	convey.Convey("test prepareComponentLogDir", func() {
		currentPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
		convey.So(err, convey.ShouldBeNil)
		ins, err := GetSftInstallMgrIns([]string{"edge-manager"}, "", currentPath, "")
		convey.So(err, convey.ShouldBeNil)

		convey.Convey("test prepareComponentLogDir func success", func() {
			p := gomonkey.ApplyFuncReturn(envutils.GetUid, uint32(8000), nil).
				ApplyFuncReturn(envutils.GetGid, uint32(8000), nil).
				ApplyFuncReturn(fileutils.SetPathOwnerGroup, nil)
			defer ResetAndClearDir(p, ins.logPathMgr.GetComponentLogPath("edge-manager"))
			convey.So(ins.prepareComponentLogDir(), convey.ShouldBeNil)
		})

		convey.Convey("test prepareComponentLogDir func pathMgr is nil", func() {
			var tempIns = &SftInstallCtl{SoftwareMgr: util.SoftwareMgr{Components: []string{"edge-manager"}}}
			convey.So(tempIns.prepareComponentLogDir(), convey.ShouldResemble,
				errors.New("pointer pathMgr is nil"))
		})

		convey.Convey("test prepareComponentLogDir func makedir failed", func() {
			p := gomonkey.ApplyFuncReturn(fileutils.CreateDir, ErrTest)
			defer p.Reset()
			convey.So(ins.prepareComponentLogDir(), convey.ShouldResemble,
				errors.New("prepare component [edge-manager] log dir failed"))
		})

		convey.Convey("test prepareComponentLogDir func chown failed", func() {
			p := gomonkey.ApplyFuncReturn(envutils.GetUid, uint32(8000), nil).
				ApplyFuncReturn(envutils.GetGid, uint32(8000), nil).
				ApplyFuncReturn(fileutils.SetPathOwnerGroup, ErrTest)
			defer ResetAndClearDir(p, ins.logPathMgr.GetComponentLogPath("edge-manager"))
			convey.So(ins.prepareComponentLogDir(), convey.ShouldResemble,
				errors.New("set run script path owner failed"))
		})
	})
}

func PrepareComponentLogDirRightCheckTest() {
	convey.Convey("test prepareComponentLogDir", func() {
		currentPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
		convey.So(err, convey.ShouldBeNil)
		ins, err := GetSftInstallMgrIns([]string{"edge-manager"}, "", currentPath, "")
		convey.So(err, convey.ShouldBeNil)
		convey.Convey("test prepareComponentLogDir func get mef-center uid failed", func() {
			p := gomonkey.ApplyFuncReturn(user.Lookup, nil, ErrTest)
			defer ResetAndClearDir(p, ins.logPathMgr.GetComponentLogPath("edge-manager"))
			convey.So(ins.prepareComponentLogDir(), convey.ShouldResemble, errors.New("get mef uid or gid failed"))
		})

		convey.Convey("test prepareComponentLogDir func transfer mef-center uid to int failed", func() {
			p := gomonkey.ApplyFuncReturn(user.Lookup, &user.User{Uid: "string"}, nil)
			defer ResetAndClearDir(p, ins.logPathMgr.GetComponentLogPath("edge-manager"))
			convey.So(ins.prepareComponentLogDir(), convey.ShouldResemble, errors.New("get mef uid or gid failed"))
		})

		convey.Convey("test prepareComponentLogDir func get mef-center gid failed", func() {
			p := gomonkey.ApplyFuncReturn(user.Lookup, &user.User{Uid: "8000"}, nil).
				ApplyFuncReturn(user.LookupGroup, nil, ErrTest)
			defer ResetAndClearDir(p, ins.logPathMgr.GetComponentLogPath("edge-manager"))
			convey.So(ins.prepareComponentLogDir(), convey.ShouldResemble, errors.New("get mef uid or gid failed"))
		})

		convey.Convey("test prepareComponentLogDir func transfer mef-center gid to int failed", func() {
			p := gomonkey.ApplyFuncReturn(user.Lookup, &user.User{Uid: "8000"}, nil).
				ApplyFuncReturn(user.LookupGroup, &user.Group{Gid: "string"}, nil)
			defer ResetAndClearDir(p, ins.logPathMgr.GetComponentLogPath("edge-manager"))
			convey.So(ins.prepareComponentLogDir(), convey.ShouldResemble, errors.New("get mef uid or gid failed"))
		})
	})
}

func PrepareCertsTest() {
	ins, err := GetSftInstallMgrIns([]string{"edge-manager"}, "", "", "")
	convey.So(err, convey.ShouldBeNil)
	var certCtlIns *certPrepareCtl
	convey.Convey("test prepareCerts func success", func() {
		p := gomonkey.ApplyPrivateMethod(certCtlIns, "doPrepare", func(_ *certPrepareCtl) error { return nil })
		defer p.Reset()
		convey.So(ins.prepareCerts(), convey.ShouldBeNil)
	})

	convey.Convey("test prepareCerts func failed", func() {
		p := gomonkey.ApplyPrivateMethod(certCtlIns, "doPrepare",
			func(_ *certPrepareCtl) error { return ErrTest })
		defer p.Reset()
		convey.So(ins.prepareCerts(), convey.ShouldResemble, errors.New("prepare certs failed"))
	})
}

func PrepareWorkingDirTest() {
	ins, err := GetSftInstallMgrIns([]string{"edge-manager"}, "", "", "")
	convey.So(err, convey.ShouldBeNil)
	var workDirMgtCtl *WorkingDirCtl
	convey.Convey("test prepareWorkingDir func success", func() {
		p := gomonkey.ApplyPrivateMethod(workDirMgtCtl, "DoInstallPrepare",
			func(_ *WorkingDirCtl) error { return nil })
		defer p.Reset()
		convey.So(ins.prepareWorkingDir(), convey.ShouldBeNil)
	})

	convey.Convey("test prepareWorkingDir func failed", func() {
		p := gomonkey.ApplyPrivateMethod(workDirMgtCtl, "DoInstallPrepare",
			func(_ *WorkingDirCtl) error { return ErrTest })
		defer p.Reset()
		convey.So(ins.prepareWorkingDir(), convey.ShouldResemble, ErrTest)
	})
}

func PrepareYamlTest() {
	ins, err := GetSftInstallMgrIns([]string{"edge-manager"}, "", "", "")
	convey.So(err, convey.ShouldBeNil)
	var yamlMgrCtl *YamlMgr
	convey.Convey("test prepareYaml func success", func() {
		p := gomonkey.ApplyMethodReturn(yamlMgrCtl, "EditSingleYaml", nil).
			ApplyFuncReturn(util.ModifyEndpointYaml, nil).
			ApplyPrivateMethod(backuputils.NewBackupFileMgr(""), "BackUp", func() error { return nil })
		defer p.Reset()
		convey.So(ins.prepareYaml(), convey.ShouldBeNil)
	})

	convey.Convey("test prepareYaml func failed", func() {
		p := gomonkey.ApplyMethodReturn(yamlMgrCtl, "EditSingleYaml", ErrTest)
		defer p.Reset()
		convey.So(ins.prepareYaml(), convey.ShouldResemble, ErrTest)
	})
}

func SetInstallJsonTest() {
	ins, err := GetSftInstallMgrIns([]string{"edge-manager"}, "", "", "")
	convey.So(err, convey.ShouldBeNil)
	var jsonHandlerIns *util.InstallParamJsonTemplate
	convey.Convey("test setInstallJson func success", func() {
		p := gomonkey.ApplyMethodReturn(jsonHandlerIns, "SetInstallParamJsonInfo", nil)
		defer p.Reset()
		convey.So(ins.setInstallJson(), convey.ShouldBeNil)
	})

	convey.Convey("test setInstallJson func failed", func() {
		p := gomonkey.ApplyMethodReturn(jsonHandlerIns, "SetInstallParamJsonInfo", ErrTest)
		defer p.Reset()
		convey.So(ins.setInstallJson(), convey.ShouldResemble, ErrTest)
	})
}

func ComponentsInstallTest() {
	ins, err := GetSftInstallMgrIns([]string{"edge-manager"}, "", "", "")
	convey.So(err, convey.ShouldBeNil)
	var componentMgrIns *util.ComponentMgr
	convey.Convey("test componentsInstall func success", func() {
		p := gomonkey.ApplyMethodReturn(componentMgrIns, "LoadAndSaveImage", nil).
			ApplyMethodReturn(componentMgrIns, "ClearDockerFile", nil).
			ApplyMethodReturn(componentMgrIns, "ClearLibDir", nil)
		defer p.Reset()
		convey.So(ins.componentsInstall(), convey.ShouldBeNil)
	})

	convey.Convey("test componentsInstall func LoadAndSaveImage failed", func() {
		p := gomonkey.ApplyMethodReturn(componentMgrIns, "LoadAndSaveImage", ErrTest)
		defer p.Reset()
		convey.So(ins.componentsInstall(), convey.ShouldResemble,
			fmt.Errorf("install component [edge-manager] failed: %s", ErrTest.Error()))
	})

	convey.Convey("test componentsInstall func ClearDockerFile failed", func() {
		p := gomonkey.ApplyMethodReturn(componentMgrIns, "LoadAndSaveImage", nil).
			ApplyMethodReturn(componentMgrIns, "ClearDockerFile", ErrTest)
		defer p.Reset()
		convey.So(ins.componentsInstall(), convey.ShouldResemble,
			fmt.Errorf("clear component [edge-manager]'s docker file failed: %s", ErrTest.Error()))
	})

	convey.Convey("test componentsInstall func ClearLibDir failed", func() {
		p := gomonkey.ApplyMethodReturn(componentMgrIns, "LoadAndSaveImage", nil).
			ApplyMethodReturn(componentMgrIns, "ClearDockerFile", nil).
			ApplyMethodReturn(componentMgrIns, "ClearLibDir", ErrTest)
		defer p.Reset()
		convey.So(ins.componentsInstall(), convey.ShouldResemble,
			fmt.Errorf("clear component [edge-manager]'s lib dir failed: %s", ErrTest.Error()))
	})
}

// SetCenterModeTest tests the success and failure scenarios of setting the center mode.
func SetCenterModeTest() {
	ins, err := GetSftInstallMgrIns([]string{"edge-manager"}, "", "", "")
	convey.So(err, convey.ShouldBeNil)
	centerModeMgrIns := &util.CenterModeMgr{}

	convey.Convey("test setCenterMode success", func() {
		patches := gomonkey.ApplyMethodReturn(centerModeMgrIns, "SetWorkDirMode", nil).
			ApplyMethodReturn(centerModeMgrIns, "SetConfigDirMode", nil).
			ApplyMethodReturn(centerModeMgrIns, "SetOutter755Mode", nil)
		defer patches.Reset()
		convey.So(ins.setCenterMode(), convey.ShouldBeNil)
	})

	convey.Convey("test setCenterMode set work dir mode failed", func() {
		patch := gomonkey.ApplyMethodReturn(centerModeMgrIns, "SetWorkDirMode", ErrTest)
		defer patch.Reset()
		convey.So(ins.setCenterMode(), convey.ShouldResemble, errors.New("set work dir mode failed"))
	})

	convey.Convey("test setCenterMode set config dir mode failed", func() {
		patches := gomonkey.ApplyMethodReturn(centerModeMgrIns, "SetWorkDirMode", nil).
			ApplyMethodReturn(centerModeMgrIns, "SetConfigDirMode", ErrTest)
		defer patches.Reset()
		convey.So(ins.setCenterMode(), convey.ShouldResemble, errors.New("set config dir mode failed"))
	})

	convey.Convey("test setCenterMode set path mode failed", func() {
		patches := gomonkey.ApplyMethodReturn(centerModeMgrIns, "SetWorkDirMode", nil).
			ApplyMethodReturn(centerModeMgrIns, "SetConfigDirMode", nil).
			ApplyMethodReturn(centerModeMgrIns, "SetOutter755Mode", ErrTest)
		defer patches.Reset()
		convey.So(ins.setCenterMode(), convey.ShouldResemble, ErrTest)
	})
}

// PrepareInstallPkgDirTest tests the success and failure scenarios of preparing the installation package directory.
func PrepareInstallPkgDirTest() {
	ins, err := GetSftInstallMgrIns([]string{"edge-manager"}, "", "", "")
	convey.So(err, convey.ShouldBeNil)

	convey.Convey("test prepareInstallPkgDir success", func() {
		patch := gomonkey.ApplyFuncReturn(fileutils.MakeSureDir, nil)
		defer patch.Reset()
		convey.So(ins.prepareInstallPkgDir(), convey.ShouldBeNil)
	})

	convey.Convey("test prepareInstallPkgDir prepare install_package dir failed", func() {
		patch := gomonkey.ApplyFuncReturn(fileutils.MakeSureDir, ErrTest)
		defer patch.Reset()
		convey.So(ins.prepareInstallPkgDir(), convey.ShouldResemble,
			errors.New("prepare install_package dir failed"))
	})
}

// PrepareLogDumpDirTest tests the success and failure scenarios of preparing the log dump directory.
func PrepareLogDumpDirTest() {
	ins, err := GetSftInstallMgrIns([]string{"edge-manager"}, "", "", "")
	convey.So(err, convey.ShouldBeNil)

	convey.Convey("test prepare log dump dir success", func() {
		patch := gomonkey.ApplyFuncReturn(util.PrepareLogDumpDir, nil)
		defer patch.Reset()
		convey.So(ins.prepareLogDumpDir(), convey.ShouldBeNil)
	})

	convey.Convey("test prepare log dump dir failed", func() {
		patch := gomonkey.ApplyFuncReturn(util.PrepareLogDumpDir, ErrTest)
		defer patch.Reset()
		convey.So(ins.prepareLogDumpDir(), convey.ShouldResemble,
			fmt.Errorf("prepare log dump dir failed, %v", ErrTest))
	})
}

// PrepareConfigDirTest tests the success and failure scenarios of preparing the configuration directory.
func PrepareConfigDirTest() {
	ins, err := GetSftInstallMgrIns([]string{"edge-manager"}, "", "", "")
	convey.So(err, convey.ShouldBeNil)

	configMgr := &util.ConfigMgr{}
	convey.Convey("test prepare config dir success", func() {
		patch := gomonkey.ApplyMethodReturn(configMgr, "DoPrepare", nil)
		defer patch.Reset()
		convey.So(ins.prepareConfigDir(), convey.ShouldBeNil)
	})

	convey.Convey("test prepare config dir failed", func() {
		patch := gomonkey.ApplyMethodReturn(configMgr, "DoPrepare", ErrTest)
		defer patch.Reset()
		convey.So(ins.prepareConfigDir(), convey.ShouldResemble,
			errors.New("prepare config dir failed"))
	})
}

// SetConfigOwnerTest tests the success and failure scenarios of setting the configuration owner.
func SetConfigOwnerTest() {
	ins, err := GetSftInstallMgrIns([]string{"edge-manager"}, "", "", "")
	convey.So(err, convey.ShouldBeNil)

	var ownerMgr = util.GetOwnerMgr(ins.InstallPathMgr.ConfigPathMgr)
	convey.Convey("test set owner for mef-center config success", func() {
		patch := gomonkey.ApplyMethodReturn(ownerMgr, "SetConfigOwner", nil)
		defer patch.Reset()
		convey.So(ins.setConfigOwner(), convey.ShouldBeNil)
	})

	convey.Convey("test set owner for mef-center config failed", func() {
		patch := gomonkey.ApplyMethodReturn(ownerMgr, "SetConfigOwner", ErrTest)
		defer patch.Reset()
		convey.So(ins.setConfigOwner(), convey.ShouldResemble,
			fmt.Errorf("set owner for mef-center config faild, %v", ErrTest))
	})
}

// ConfigBackupTest tests the success and failure scenarios of creating a backup for the configuration.
func ConfigBackupTest() {
	ins, err := GetSftInstallMgrIns([]string{"edge-manager"}, "", "", "")
	convey.So(err, convey.ShouldBeNil)

	configPath := ins.InstallPathMgr.ConfigPathMgr.GetConfigPath()
	backup := backuputils.NewBackupDirMgr(configPath, backuputils.JsonFileType, backuputils.CrtFileType,
		backuputils.CrlFileType, backuputils.KeyFileType)
	convey.Convey("test create backup for mef-center config success", func() {
		patch := gomonkey.ApplyMethodReturn(backup, "BackUp", nil)
		defer patch.Reset()
		convey.So(ins.configBackup(), convey.ShouldBeNil)
	})

	convey.Convey("test create backup for mef-center config failed", func() {
		patch := gomonkey.ApplyMethodReturn(backup, "BackUp", ErrTest)
		defer patch.Reset()
		convey.So(ins.configBackup(), convey.ShouldResemble,
			fmt.Errorf("back up mef-center config failed, %v", ErrTest))
	})
}
