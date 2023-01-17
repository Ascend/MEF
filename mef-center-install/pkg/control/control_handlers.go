package control

import (
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

type SftControlMgr struct {
	componentFlag      string
	operate            string
	installedComponent []string
	installPathMgr     *util.InstallDirPathMgr
	componentList      []*util.CtlComponent
}

func (scm *SftControlMgr) DoControl() error {
	var installTasks = []func() error{
		scm.init,
		scm.deal,
	}

	for _, function := range installTasks {
		if err := function(); err != nil {
			return err
		}
	}

	return nil
}

func (scm *SftControlMgr) init() error {
	// if all, then construct a full componentFlag list. (Does not support batch configuration)
	if scm.componentFlag == "all" {
		for _, c := range scm.installedComponent {
			component := &util.CtlComponent{
				Name:           c,
				Operation:      scm.operate,
				InstallPathMgr: scm.installPathMgr,
			}
			scm.componentList = append(scm.componentList, component)
		}

		// if just a certain componentFlag, then construct a single-element componentFlag list
	} else {
		component := &util.CtlComponent{
			Name:           scm.componentFlag,
			Operation:      scm.operate,
			InstallPathMgr: scm.installPathMgr,
		}
		scm.componentList = append(scm.componentList, component)
	}

	hwlog.RunLog.Info("init componentFlag list successful")
	return nil
}

func (scm *SftControlMgr) deal() error {
	for _, component := range scm.componentList {
		if err := component.Operate(); err != nil {
			return err
		}
	}
	return nil
}

func InitSftControlMgr(component, operate string,
	installComponents []string, installPathMgr *util.InstallDirPathMgr) *SftControlMgr {
	return &SftControlMgr{
		componentFlag:      component,
		operate:            operate,
		installedComponent: installComponents,
		installPathMgr:     installPathMgr,
	}
}
