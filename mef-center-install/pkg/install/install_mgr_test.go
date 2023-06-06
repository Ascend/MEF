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

	. "github.com/agiledragon/gomonkey/v2"
	. "github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

func DoInstallMgrTest() {
	Convey("doInstall func", DoInstallTest)
	Convey("checkInstalled func", CheckInstalledTest)
	Convey("preCheck func", PreCheckTest)
	Convey("CheckUser func", CheckUserTest)
	Convey("CheckDiskSpace func", CheckDiskSpaceTest)
	Convey("CheckNecessaryTools func", CheckNecessaryToolsTest)
	Convey("prepareMefUser func success situation", PrepareMefUserSuccessTest)
	Convey("prepareMefUser func failed situation", PrepareMefUserCreateUserFailedTest)
	Convey("prepareMefUser func userCheck failed", PrePareMefUserUserCheckFailedTest)
	Convey("prepareK8sLabel func", PrepareK8sLabelTest)
	Convey("prepareComponent func", PrepareComponentLogDirTest)
	Convey("prepareComponentLogDir func", PrepareComponentLogDirRightCheckTest)
	Convey("prepareCerts func", PrepareCertsTest)
	Convey("prepareWorkingDir func", PrepareWorkingDirTest)
	Convey("prepareYaml func", PrepareYamlTest)
	Convey("setInstallJson func", SetInstallJsonTest)
	Convey("componentInstall func", ComponentsInstallTest)
}

func DoInstallTest() {
	ins := GetSftInstallMgrIns([]string{}, "", "", "", "")
	Convey("test doInstall func success", func() {
		p := ApplyPrivateMethod(ins, "checkInstalled", func(_ *SftInstallCtl) error { return nil }).
			ApplyPrivateMethod(ins, "preCheck", func(_ *SftInstallCtl) error { return nil }).
			ApplyPrivateMethod(ins, "prepareMefUser", func(_ *SftInstallCtl) error { return nil }).
			ApplyPrivateMethod(ins, "prepareK8sLabel", func(_ *SftInstallCtl) error { return nil }).
			ApplyPrivateMethod(ins, "prepareComponentLogDir", func(_ *SftInstallCtl) error { return nil }).
			ApplyPrivateMethod(ins, "prepareInstallPkgDir", func(_ *SftInstallCtl) error { return nil }).
			ApplyPrivateMethod(ins, "prepareCerts", func(_ *SftInstallCtl) error { return nil }).
			ApplyPrivateMethod(ins, "copyCloudCoreCa", func(_ *SftInstallCtl) error { return nil }).
			ApplyPrivateMethod(ins, "prepareWorkingDir", func(_ *SftInstallCtl) error { return nil }).
			ApplyPrivateMethod(ins, "prepareConfigDir", func(_ *SftInstallCtl) error { return nil }).
			ApplyPrivateMethod(ins, "prepareYaml", func(_ *SftInstallCtl) error { return nil }).
			ApplyPrivateMethod(ins, "setInstallJson", func(_ *SftInstallCtl) error { return nil }).
			ApplyPrivateMethod(ins, "componentsInstall", func(_ *SftInstallCtl) error { return nil }).
			ApplyPrivateMethod(ins, "setCenterMode", func(_ *SftInstallCtl) error { return nil })
		defer p.Reset()
		So(ins.DoInstall(), ShouldBeNil)
	})

	Convey("test doInstall func checkInstalled failed", func() {
		p := ApplyPrivateMethod(ins, "checkUser", func(_ *SftInstallCtl) error { return nil }).
			ApplyPrivateMethod(ins, "checkNecessaryTools", func(_ *SftInstallCtl) error { return nil }).
			ApplyFuncReturn(os.Stat, nil, nil)
		defer p.Reset()
		So(ins.DoInstall(), ShouldResemble, errors.New("the software has already been installed"))
	})

	k8sMgrIns := &util.K8sLabelMgr{}
	Convey("test doInstall func k8s label exists failed", func() {
		p := ApplyPrivateMethod(ins, "checkUser", func(_ *SftInstallCtl) error { return nil }).
			ApplyPrivateMethod(ins, "checkNecessaryTools", func(_ *SftInstallCtl) error { return nil }).
			ApplyPrivateMethod(k8sMgrIns, "CheckK8sLabel", func(_ *util.K8sLabelMgr) (bool, error) { return true, nil })
		defer p.Reset()
		So(ins.DoInstall(), ShouldResemble, errors.New("the software has already been installed since k8s label exists"))
	})

	Convey("test doInstall func install failed", func() {
		p := ApplyPrivateMethod(ins, "preCheck", func(_ *SftInstallCtl) error { return ErrTest }).
			ApplyPrivateMethod(ins, "clearAll", func(_ *SftInstallCtl) { return })
		defer p.Reset()
		So(ins.DoInstall(), ShouldResemble, ErrTest)
	})

}

func CheckInstalledTest() {
	k8sMgr := &util.K8sLabelMgr{}
	ins := GetSftInstallMgrIns([]string{}, "", "", "", "")
	Convey("checkInstalled func success", func() {
		p := ApplyMethodReturn(k8sMgr, "CheckK8sLabel", false, nil)
		defer p.Reset()
		So(ins.checkInstalled(), ShouldBeNil)
	})

	Convey("checkInstalled func failed by installed", func() {
		p := ApplyMethodReturn(k8sMgr, "CheckK8sLabel", true, nil)
		defer p.Reset()
		So(ins.checkInstalled(), ShouldResemble,
			errors.New("the software has already been installed since k8s label exists"))
	})

	Convey("checkInstalled func failed", func() {
		p := ApplyFuncReturn(os.Stat, nil, nil)
		defer p.Reset()
		So(ins.checkInstalled(), ShouldResemble, errors.New("the software has already been installed"))
	})
}

func PreCheckTest() {
	ins := GetSftInstallMgrIns([]string{}, "", "", "", "")
	Convey("test preCheck func success", func() {
		p := ApplyPrivateMethod(ins, "checkUser", func(_ *SftInstallCtl) error { return nil }).
			ApplyPrivateMethod(ins, "checkDiskSpace", func(_ *SftInstallCtl) error { return nil }).
			ApplyPrivateMethod(ins, "checkInstalled", func(_ *SftInstallCtl) error { return nil }).
			ApplyPrivateMethod(ins, "checkNecessaryTools", func(_ *SftInstallCtl) error { return nil })
		defer p.Reset()
		So(ins.preCheck(), ShouldBeNil)
	})

	Convey("test preCheck func failed", func() {
		p := ApplyPrivateMethod(ins, "checkUser", func(_ *SftInstallCtl) error { return ErrTest })
		defer p.Reset()
		So(ins.preCheck(), ShouldResemble, ErrTest)
	})
}

func CheckUserTest() {
	ins := GetSftInstallMgrIns([]string{}, "", "", "", "")
	Convey("test CheckUser func check user failed", func() {
		p := ApplyFuncReturn(user.Current, nil, ErrTest)
		defer p.Reset()
		resp := fmt.Errorf("get current user info failed: %s", ErrTest.Error())
		So(ins.checkUser(), ShouldResemble, resp)
	})

	userName := "notExist"
	Convey(fmt.Sprintf("test CheckUser func user is %s", userName), func() {
		userName = "notExist"
		p := ApplyFuncReturn(user.Current, &user.User{Username: userName}, nil)
		defer p.Reset()
		resp := fmt.Errorf("install failed: the install user must be root, can not be %s", userName)
		So(ins.checkUser(), ShouldResemble, resp)
	})

	userName = "root"
	Convey(fmt.Sprintf("test CheckUser func user is %s", userName), func() {
		userName = "root"
		p := ApplyFuncReturn(user.Current, &user.User{Username: userName}, nil)
		defer p.Reset()
		So(ins.checkUser(), ShouldBeNil)
	})
}

func CheckDiskSpaceTest() {
	ins := GetSftInstallMgrIns([]string{}, "", "", "", "")
	Convey("test Check Disk Space func get disk free failed", func() {
		p := ApplyFunc(syscall.Statfs, func(_ string, _ *syscall.Statfs_t) error {
			return ErrTest
		})
		defer p.Reset()
		So(ins.checkDiskSpace(), ShouldResemble, errors.New("check install disk space failed"))
	})

	Convey("test Check Disk Space func Disk Free enough", func() {
		p := ApplyFuncReturn(common.GetFileDevNum, uint64(1), nil).
			ApplyFunc(syscall.Statfs, func(_ string,
				fs *syscall.Statfs_t) error {
				fs.Bavail = util.InstallDiskSpace
				fs.Bsize = 1
				return nil
			})
		defer p.Reset()
		So(ins.checkDiskSpace(), ShouldBeNil)
	})

	Convey("test Check Disk Space func Disk Free not enough", func() {
		p := ApplyFunc(syscall.Statfs, func(_ string, fs *syscall.Statfs_t) error {
			fs.Bavail = util.InstallDiskSpace - 1
			fs.Bsize = 1
			return nil
		})
		defer p.Reset()
		So(ins.checkDiskSpace(), ShouldResemble, errors.New("check install disk space failed"))
	})
}

func CheckNecessaryToolsTest() {
	ins := GetSftInstallMgrIns([]string{}, "", "", "", "")
	Convey("test CheckNecessaryTools func failed", func() {
		p := ApplyFuncReturn(exec.LookPath, "", ErrTest)
		defer p.Reset()
		So(ins.checkNecessaryTools(), ShouldNotBeNil)
	})

	Convey("test CheckNecessaryTools func success", func() {
		p := ApplyFuncReturn(exec.LookPath, "", nil)
		defer p.Reset()
		So(ins.checkNecessaryTools(), ShouldBeNil)
	})
}

func PrepareMefUserSuccessTest() {
	ins := GetSftInstallMgrIns([]string{}, "", "", "", "")
	Convey("test PrepareMefUser func create user 8000 uid success", func() {
		p := ApplyFuncReturn(user.Lookup, nil, ErrTest).
			ApplyFuncReturn(user.LookupGroup, nil, ErrTest).
			ApplyFuncReturn(utils.IsExist, false).
			ApplyFuncReturn(exec.LookPath, "", nil).
			ApplyFuncReturn(user.LookupId, nil, nil).
			ApplyFuncReturn(common.RunCommand, "", nil)
		defer p.Reset()
		So(ins.prepareMefUser(), ShouldBeNil)
	})

	Convey("test PrepareMefUser func create user auto uid success", func() {
		p := ApplyFuncReturn(user.Lookup, nil, ErrTest).
			ApplyFuncReturn(user.LookupGroup, nil, ErrTest).
			ApplyFuncReturn(utils.IsExist, false).
			ApplyFuncReturn(exec.LookPath, "", nil).
			ApplyFuncReturn(user.LookupId, nil, ErrTest).
			ApplyFuncReturn(common.RunCommand, "", nil)
		defer p.Reset()
		So(ins.prepareMefUser(), ShouldBeNil)
	})

	Convey("test PrepareMefUser func user exists success", func() {
		p := ApplyFuncReturn(user.Lookup, &user.User{}, nil).
			ApplyFuncReturn(user.LookupGroup, &user.Group{Gid: "8000"}, nil).
			ApplyMethodReturn(&user.User{}, "GroupIds", []string{"8000"}, nil).
			ApplyFuncReturn(common.RunCommand, "nologin", nil)
		defer p.Reset()
		So(ins.prepareMefUser(), ShouldBeNil)
	})
}

func PrepareMefUserCreateUserFailedTest() {
	ins := GetSftInstallMgrIns([]string{}, "", "", "", "")
	Convey("test PrepareMefUser func create user command exec failed", func() {
		p := ApplyFuncReturn(user.Lookup, nil, ErrTest).
			ApplyFuncReturn(user.LookupGroup, nil, ErrTest).
			ApplyFuncReturn(utils.IsExist, false).
			ApplyFuncReturn(exec.LookPath, "", nil).
			ApplyFuncReturn(user.LookupId, nil, ErrTest).
			ApplyFuncReturn(common.RunCommand, "", ErrTest)
		defer p.Reset()
		So(ins.prepareMefUser(), ShouldResemble, errors.New("exec useradd command failed"))
	})

	Convey("test PrepareMefUser func create user no nologin cmd", func() {
		p := ApplyFuncReturn(user.Lookup, nil, ErrTest).
			ApplyFuncReturn(user.LookupGroup, nil, ErrTest).
			ApplyFuncReturn(utils.IsExist, false).
			ApplyFuncReturn(exec.LookPath, "", ErrTest)
		defer p.Reset()
		So(ins.prepareMefUser(), ShouldResemble, errors.New("look path of nologin failed"))
	})

	Convey("test PrepareMefUser func create user home dir exists", func() {
		p := ApplyFuncReturn(user.Lookup, nil, ErrTest).
			ApplyFuncReturn(user.LookupGroup, nil, ErrTest).
			ApplyFuncReturn(utils.IsExist, true)
		defer p.Reset()
		So(ins.prepareMefUser(), ShouldResemble, errors.New("user home dir exists"))
	})

	Convey("test PrepareMefUser func create user user exists but not group", func() {
		p := ApplyFuncReturn(user.Lookup, nil, ErrTest).
			ApplyFuncReturn(user.LookupGroup, nil, nil)
		ret, err := user.LookupGroup("test")
		fmt.Printf("\nret = %v, err = %v\n", ret, err)
		defer p.Reset()
		So(ins.prepareMefUser(), ShouldResemble, errors.New("the user name or group name is in use"))
	})

}

func PrePareMefUserUserCheckFailedTest() {
	ins := GetSftInstallMgrIns([]string{}, "", "", "", "")
	Convey("test PrepareMefUser func get groupIds failed", func() {
		p := ApplyFuncReturn(user.Lookup, &user.User{}, nil).
			ApplyFuncReturn(user.LookupGroup, &user.Group{Gid: "8000"}, nil).
			ApplyMethodReturn(&user.User{}, "GroupIds", nil, ErrTest)
		defer p.Reset()
		So(ins.prepareMefUser(), ShouldNotBeNil)
	})

	Convey("test PrepareMefUser func gid not in user group failed", func() {
		p := ApplyFuncReturn(user.Lookup, &user.User{}, nil).
			ApplyFuncReturn(user.LookupGroup, &user.Group{Gid: "8000"}, nil).
			ApplyMethodReturn(&user.User{}, "GroupIds", []string{"9000"}, nil)
		defer p.Reset()
		So(ins.prepareMefUser(), ShouldNotBeNil)
	})

	Convey("test PrepareMefUser func check is user nologin command exec failed", func() {
		p := ApplyFuncReturn(user.Lookup, &user.User{}, nil).
			ApplyFuncReturn(user.LookupGroup, &user.Group{Gid: "8000"}, nil).
			ApplyMethodReturn(&user.User{}, "GroupIds", []string{"8000"}, nil).
			ApplyFuncReturn(common.RunCommand, "", ErrTest)
		defer p.Reset()
		So(ins.prepareMefUser(), ShouldResemble, errors.New("exec check nologin command failed"))
	})

	Convey("test PrepareMefUser func check is user nologin failed", func() {
		p := ApplyFuncReturn(user.Lookup, &user.User{}, nil).
			ApplyFuncReturn(user.LookupGroup, &user.Group{Gid: "8000"}, nil).
			ApplyMethodReturn(&user.User{}, "GroupIds", []string{"8000"}, nil).
			ApplyFuncReturn(common.RunCommand, "-100", nil)
		defer p.Reset()
		So(ins.prepareMefUser(), ShouldNotBeNil)
	})
}

func PrepareK8sLabelTest() {
	ins := GetSftInstallMgrIns([]string{}, "", "", "", "")

	k8sMgr := &util.K8sLabelMgr{}
	Convey("test prepareK8sLabel func success", func() {
		p := ApplyMethodReturn(k8sMgr, "PrepareK8sLabel", nil)
		defer p.Reset()
		So(ins.prepareK8sLabel(), ShouldBeNil)
	})

	Convey("test prepareK8sLabel func failed", func() {
		p := ApplyMethodReturn(k8sMgr, "PrepareK8sLabel", ErrTest)
		defer p.Reset()
		So(ins.prepareK8sLabel(), ShouldResemble, ErrTest)
	})
}

func PrepareComponentLogDirTest() {
	Convey("test prepareComponentLogDir", func() {
		currentPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
		So(err, ShouldBeNil)
		var ins = GetSftInstallMgrIns([]string{"edge-manager"}, "", currentPath, "", "")
		Convey("test prepareComponentLogDir func success", func() {
			p := ApplyFuncReturn(user.Lookup, &user.User{Uid: "8000"}, nil).
				ApplyFuncReturn(user.LookupGroup, &user.Group{Gid: "8000"}, nil).
				ApplyFuncReturn(os.Chown, nil)
			defer ResetAndClearDir(p, ins.logPathMgr.GetComponentLogPath("edge-manager"))
			So(ins.prepareComponentLogDir(), ShouldBeNil)
		})

		Convey("test prepareComponentLogDir func pathMgr is nil", func() {
			var tempIns = &SftInstallCtl{SoftwareMgr: util.SoftwareMgr{Components: []string{"edge-manager"}}}
			So(tempIns.prepareComponentLogDir(), ShouldResemble, errors.New("pointer pathMgr is nil"))
		})

		Convey("test prepareComponentLogDir func makedir failed", func() {
			p := ApplyFuncReturn(os.MkdirAll, ErrTest)
			defer p.Reset()
			So(ins.prepareComponentLogDir(), ShouldResemble,
				errors.New("prepare component [edge-manager] log dir failed"))
		})

		Convey("test prepareComponentLogDir func chown failed", func() {
			p := ApplyFuncReturn(user.Lookup, &user.User{Uid: "8000"}, nil).
				ApplyFuncReturn(user.LookupGroup, &user.Group{Gid: "8000"}, nil).
				ApplyFuncReturn(os.Chown, ErrTest)
			defer ResetAndClearDir(p, ins.logPathMgr.GetComponentLogPath("edge-manager"))
			So(ins.prepareComponentLogDir(), ShouldResemble, errors.New("set run script path owner failed"))
		})
	})
}

func PrepareComponentLogDirRightCheckTest() {
	Convey("test prepareComponentLogDir", func() {
		currentPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
		So(err, ShouldBeNil)
		var ins = GetSftInstallMgrIns([]string{"edge-manager"}, "", currentPath, "", "")
		Convey("test prepareComponentLogDir func get mef-center uid failed", func() {
			p := ApplyFuncReturn(user.Lookup, nil, ErrTest)
			defer ResetAndClearDir(p, ins.logPathMgr.GetComponentLogPath("edge-manager"))
			So(ins.prepareComponentLogDir(), ShouldResemble, errors.New("get mef uid or gid failed"))
		})

		Convey("test prepareComponentLogDir func transfer mef-center uid to int failed", func() {
			p := ApplyFuncReturn(user.Lookup, &user.User{Uid: "string"}, nil)
			defer ResetAndClearDir(p, ins.logPathMgr.GetComponentLogPath("edge-manager"))
			So(ins.prepareComponentLogDir(), ShouldResemble, errors.New("get mef uid or gid failed"))
		})

		Convey("test prepareComponentLogDir func get mef-center gid failed", func() {
			p := ApplyFuncReturn(user.Lookup, &user.User{Uid: "8000"}, nil).
				ApplyFuncReturn(user.LookupGroup, nil, ErrTest)
			defer ResetAndClearDir(p, ins.logPathMgr.GetComponentLogPath("edge-manager"))
			So(ins.prepareComponentLogDir(), ShouldResemble, errors.New("get mef uid or gid failed"))
		})

		Convey("test prepareComponentLogDir func transfer mef-center gid to int failed", func() {
			p := ApplyFuncReturn(user.Lookup, &user.User{Uid: "8000"}, nil).
				ApplyFuncReturn(user.LookupGroup, &user.Group{Gid: "string"}, nil)
			defer ResetAndClearDir(p, ins.logPathMgr.GetComponentLogPath("edge-manager"))
			So(ins.prepareComponentLogDir(), ShouldResemble, errors.New("get mef uid or gid failed"))
		})
	})
}

func PrepareCertsTest() {
	var ins = GetSftInstallMgrIns([]string{"edge-manager"}, "", "", "", "")
	var certCtlIns *certPrepareCtl
	Convey("test prepareCerts func success", func() {
		p := ApplyPrivateMethod(certCtlIns, "doPrepare", func(_ *certPrepareCtl) error { return nil })
		defer p.Reset()
		So(ins.prepareCerts(), ShouldBeNil)
	})

	Convey("test prepareCerts func failed", func() {
		p := ApplyPrivateMethod(certCtlIns, "doPrepare",
			func(_ *certPrepareCtl) error { return ErrTest })
		defer p.Reset()
		So(ins.prepareCerts(), ShouldResemble, errors.New("prepare certs failed"))
	})
}

func PrepareWorkingDirTest() {
	var ins = GetSftInstallMgrIns([]string{"edge-manager"}, "", "", "", "")
	var workDirMgtCtl *WorkingDirCtl
	Convey("test prepareWorkingDir func success", func() {
		p := ApplyPrivateMethod(workDirMgtCtl, "DoInstallPrepare", func(_ *WorkingDirCtl) error { return nil })
		defer p.Reset()
		So(ins.prepareWorkingDir(), ShouldBeNil)
	})

	Convey("test prepareWorkingDir func failed", func() {
		p := ApplyPrivateMethod(workDirMgtCtl, "DoInstallPrepare",
			func(_ *WorkingDirCtl) error { return ErrTest })
		defer p.Reset()
		So(ins.prepareWorkingDir(), ShouldResemble, ErrTest)
	})
}

func PrepareYamlTest() {
	var ins = GetSftInstallMgrIns([]string{"edge-manager"}, "", "", "", "")
	var yamlMgrCtl *YamlMgr
	Convey("test prepareYaml func success", func() {
		p := ApplyMethodReturn(yamlMgrCtl, "EditSingleYaml", nil)
		defer p.Reset()
		So(ins.prepareYaml(), ShouldBeNil)
	})

	Convey("test prepareYaml func failed", func() {
		p := ApplyMethodReturn(yamlMgrCtl, "EditSingleYaml", ErrTest)
		defer p.Reset()
		So(ins.prepareYaml(), ShouldResemble, ErrTest)
	})
}

func SetInstallJsonTest() {
	var ins = GetSftInstallMgrIns([]string{"edge-manager"}, "", "", "", "")
	var jsonHandlerIns *util.InstallParamJsonTemplate
	Convey("test setInstallJson func success", func() {
		p := ApplyMethodReturn(jsonHandlerIns, "SetInstallParamJsonInfo", nil)
		defer p.Reset()
		So(ins.setInstallJson(), ShouldBeNil)
	})

	Convey("test setInstallJson func failed", func() {
		p := ApplyMethodReturn(jsonHandlerIns, "SetInstallParamJsonInfo", ErrTest)
		defer p.Reset()
		So(ins.setInstallJson(), ShouldResemble, ErrTest)
	})
}

func ComponentsInstallTest() {
	var ins = GetSftInstallMgrIns([]string{"edge-manager"}, "", "", "", "")
	var componentMgrIns *util.ComponentMgr
	Convey("test componentsInstall func success", func() {
		p := ApplyMethodReturn(componentMgrIns, "LoadAndSaveImage", nil).
			ApplyMethodReturn(componentMgrIns, "ClearDockerFile", nil).
			ApplyMethodReturn(componentMgrIns, "ClearLibDir", nil)
		defer p.Reset()
		So(ins.componentsInstall(), ShouldBeNil)
	})

	Convey("test componentsInstall func LoadAndSaveImage failed", func() {
		p := ApplyMethodReturn(componentMgrIns, "LoadAndSaveImage", ErrTest)
		defer p.Reset()
		So(ins.componentsInstall(), ShouldResemble,
			fmt.Errorf("install component [edge-manager] failed: %s", ErrTest.Error()))
	})

	Convey("test componentsInstall func ClearDockerFile failed", func() {
		p := ApplyMethodReturn(componentMgrIns, "LoadAndSaveImage", nil).
			ApplyMethodReturn(componentMgrIns, "ClearDockerFile", ErrTest)
		defer p.Reset()
		So(ins.componentsInstall(), ShouldResemble,
			fmt.Errorf("clear component [edge-manager]'s docker file failed: %s", ErrTest.Error()))
	})

	Convey("test componentsInstall func ClearLibDir failed", func() {
		p := ApplyMethodReturn(componentMgrIns, "LoadAndSaveImage", nil).
			ApplyMethodReturn(componentMgrIns, "ClearDockerFile", nil).
			ApplyMethodReturn(componentMgrIns, "ClearLibDir", ErrTest)
		defer p.Reset()
		So(ins.componentsInstall(), ShouldResemble,
			fmt.Errorf("clear component [edge-manager]'s lib dir failed: %s", ErrTest.Error()))
	})
}
