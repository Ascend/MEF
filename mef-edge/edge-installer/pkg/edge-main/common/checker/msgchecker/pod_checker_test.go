// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

package msgchecker

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/edge-main/common/checker/msgchecker/types"
	"edge-installer/pkg/edge-main/common/configpara"
	"edge-installer/pkg/edge-main/common/msglistchecker"
)

func TestMsgContent(t *testing.T) {
	patches := gomonkey.ApplyFunc(configpara.GetPodConfig, MockPodConfig).
		ApplyFuncReturn(configpara.GetNetType, constants.FDWithOM, nil).
		ApplyPrivateMethod(&MsgValidator{}, "checkSystemResources", func() error { return nil })
	defer patches.Reset()

	convey.Convey("test restart Pod success", t, testRestartPod)
	convey.Convey("test delete pods success", t, testDeletePods)

	convey.Convey("test Pod security para", t, func() {
		convey.Convey("test Pod hostNetwork failed", testPodHostNetWork)
		convey.Convey("test Pod hostPid failed", testPodHostPid)
		convey.Convey("tet pod available resource", testPodAvailableResources)
	})
	convey.Convey("test container secure para", t, testContainerSecContextPara)
	convey.Convey("test container port map", t, testPodPortMapPara)

	convey.Convey("test pod security para", t, testCheckPodSecurityContext)
	convey.Convey("test container security para", t, testCheckContainerSecurityContext)
}

func testRestartPod() {
	var msg model.Message
	msg.KubeEdgeRouter = model.MessageRoute{
		Source:    "controller",
		Group:     "resource",
		Operation: "restart",
		Resource:  "websocket/pod/test-eadea3ed-62ed-4b24-95a2-d2a4d98607e5",
	}
	msg.Header.ID = "90fca461-8d3f-43d7-9f44-0090b8d3389d"
	msg.Header.Timestamp = 1704373672
	msg.Header.ResourceVersion = ""
	msg.Header.Sync = true

	msgValidator := NewMsgValidator(msglistchecker.NewFdMsgHeaderValidator())

	var err error
	if err = msgValidator.Check(&msg); err != nil {
		hwlog.RunLog.Errorf("check msg failed: %v", err)
	}
	convey.So(err, convey.ShouldResemble, nil)
}

func testDeletePods() {
	var msg model.Message
	msg.KubeEdgeRouter = model.MessageRoute{
		Source:    "controller",
		Group:     "resource",
		Operation: "delete",
		Resource:  "websocket/pods_data",
	}
	msg.Header.ID = "90fca461-8d3f-43d7-9f44-0090b8d3389d"
	msg.Header.Timestamp = 1704373672
	msg.Header.ResourceVersion = ""
	msg.Header.Sync = false

	data := make([]byte, 64)
	_, err := rand.Read(data)
	if err != nil {
		fmt.Printf("read read data failed: %v", err)
		return
	}
	msg.FillContent(data)
	msgValidator := NewMsgValidator(msglistchecker.NewFdMsgHeaderValidator())
	if err = msgValidator.Check(&msg); err != nil {
		hwlog.RunLog.Errorf("check msg failed: %v", err)
	}
	convey.So(err, convey.ShouldResemble, nil)
}

var podDataBase = `{"kind":"Pod","metadata":{"creationTimestamp":"2023-12-18T13:45:57Z",
"name":"test-eadea3ed-62ed-4b24-95a2-d2a4d98607e5","namespace":"websocket","resourceVersion":"7565301",
"uid":"eadea3ed-62ed-4b24-95a2-d2a4d98607e5"},"spec":{"containers":
[{"image":"fd.fusiondirector.huawei.com:443/library/wyj:3.0","imagePullPolicy":"IfNotPresent","name":"container-0",
"resources":{"limits":{"cpu":"50m","memory":"128Mi"},"requests":{"cpu":"50m","memory":"128Mi"}},
"securityContext":{"allowPrivilegeEscalation":false,"capabilities":{"drop":["All"]},"privileged":false,
"readOnlyRootFilesystem":true,"runAsGroup":1001,"runAsNonRoot":true,"runAsUser":1001,
"seccompProfile":{"type":"RuntimeDefault"}},"terminationMessagePolicy":"File"}],
"dnsPolicy":"ClusterFirst","enableServiceLinks":true,
"imagePullSecrets":[{"name":"fusion-director-docker-registry-secret"}],"nodeName":"2102314nmv10p7100006",
"restartPolicy":"Always","schedulerName":"default-scheduler","securityContext":{},"serviceAccountName":"default",
"terminationGracePeriodSeconds":30,"tolerations":[{"effect":"NoExecute","key":"node.kubernetes.io/unreachable",
"operator":"Exists"},{"effect":"NoExecute","key":"node.kubernetes.io/not-ready","operator":"Exists"},
{"effect":"NoExecute","key":"node.kubernetes.io/network-unavailable","operator":"Exists"}]},"status":{}}`

func getPodInfo() types.Pod {
	var basePod types.Pod
	err := json.Unmarshal([]byte(podDataBase), &basePod)
	if err != nil {
		hwlog.RunLog.Infof("unmarshal Pod data failed:%v", err)
	}
	return basePod
}

func setFdPodMsg(msg *model.Message, podInfo types.Pod) {
	data, err := json.Marshal(podInfo)
	if err != nil {
		fmt.Printf("marshal Pod failed:%v", err)
		return
	}
	msg.KubeEdgeRouter = model.MessageRoute{
		Source:    "controller",
		Group:     "resource",
		Operation: "update",
		Resource:  "websocket/pod/test-d1aefae9-6ead-4942-9269-04ebae160521",
	}
	msg.Header.ID = "90fca461-8d3f-43d7-9f44-0090b8d3389d"
	msg.Header.Timestamp = 1704373672
	msg.Header.ResourceVersion = ""
	msg.Header.Sync = true

	msg.FillContent(data)
}

func setMefPodMsg(msg *model.Message, podInfo types.Pod) {
	data, err := json.Marshal(podInfo)
	if err != nil {
		fmt.Printf("marshal Pod failed:%v", err)
		return
	}
	msg.KubeEdgeRouter = model.MessageRoute{
		Source:    "edgecontroller",
		Group:     "resource",
		Operation: "update",
		Resource:  "mef-user/pod/test-2-d1aef",
	}
	msg.Header.ID = "90fca461-8d3f-43d7-9f44-0090b8d3389d"
	msg.Header.Timestamp = 1704373672
	msg.Header.ResourceVersion = ""
	msg.Header.Sync = false

	msg.FillContent(data)
}
func testPodHostNetWork() {
	var basePod = getPodInfo()
	basePod.Spec.HostNetwork = true

	var msg model.Message
	setFdPodMsg(&msg, basePod)

	msgValidator := NewMsgValidator(msglistchecker.NewFdMsgHeaderValidator())
	var err error
	if err = msgValidator.Check(&msg); err != nil {
		hwlog.RunLog.Errorf("check msg failed: %v", err)
	}
	convey.So(err, convey.ShouldNotEqual, nil)
	convey.So(err.Error(), convey.ShouldContainSubstring, "cur config not support pod host network")
}

func testPodHostPid() {
	var basePod = getPodInfo()
	basePod.Spec.HostPID = true

	var msg model.Message
	setFdPodMsg(&msg, basePod)

	msgValidator := NewMsgValidator(msglistchecker.NewFdMsgHeaderValidator())
	var err error
	if err = msgValidator.Check(&msg); err != nil {
		hwlog.RunLog.Errorf("check msg failed: %v", err)
	}
	convey.So(err, convey.ShouldNotEqual, nil)
	convey.So(err.Error(), convey.ShouldContainSubstring, "cur config not support pod host pid")
}

type securityContextTestCase struct {
	description     string
	securityContext types.SecurityContext
	shouldErr       bool
	assert          convey.Assertion
	expected        interface{}
}

var privileged = true
var runAsUser = int64(0)
var capabilities = types.Capabilities{Add: []string{"cap_chown"}}

var runAsGroup = int64(0)
var readOnlyRootFilesystem = false
var allowPrivilegeEscalation = true
var seccompProfile = types.SeccompProfile{Type: "Localhost"}

var securityTestCase = []securityContextTestCase{
	{
		description:     "container securityContext contain capability",
		securityContext: types.SecurityContext{Capabilities: &capabilities},
		shouldErr:       true,
		assert:          convey.ShouldContainSubstring,
		expected:        "check container Capability failed, cur config not support",
	}, {
		description:     "container securityContext contain privileged",
		securityContext: types.SecurityContext{Privileged: &privileged},
		shouldErr:       true,
		assert:          convey.ShouldContainSubstring,
		expected:        "check container Privileged failed, cur config not support",
	}, {
		description:     "container securityContext contain runAsUser",
		securityContext: types.SecurityContext{RunAsUser: &runAsUser},
		shouldErr:       true,
		assert:          convey.ShouldContainSubstring,
		expected:        "check container run as user failed, cur config not support",
	}, {
		description:     "container securityContext contain runAsGroup",
		securityContext: types.SecurityContext{RunAsGroup: &runAsGroup},
		shouldErr:       true,
		assert:          convey.ShouldContainSubstring,
		expected:        "check container run as group failed, cur config not support",
	}, {
		description:     "container securityContext contain readOnlyRootFilesystem",
		securityContext: types.SecurityContext{ReadOnlyRootFilesystem: &readOnlyRootFilesystem},
		shouldErr:       true,
		assert:          convey.ShouldContainSubstring,
		expected:        "check container ReadOnlyRootFilesystem failed, not support",
	}, {
		description:     "container securityContext contain allowPrivilegeEscalation",
		securityContext: types.SecurityContext{AllowPrivilegeEscalation: &allowPrivilegeEscalation},
		shouldErr:       true,
		assert:          convey.ShouldContainSubstring,
		expected:        "check container AllowPrivilegeEscalation failed, cur config not support",
	}, {
		description:     "container securityContext contain seccompProfile",
		securityContext: types.SecurityContext{SeccompProfile: &seccompProfile},
		shouldErr:       true,
		assert:          convey.ShouldContainSubstring,
		expected:        "Error:Field validation for 'Type'",
	},
}

func testContainerSecContextPara() {
	var basePod = getPodInfo()

	var err error
	msgValidator := NewMsgValidator(msglistchecker.NewFdMsgHeaderValidator())

	for _, tc := range securityTestCase {
		hwlog.RunLog.Infof("--------------------%s-------------------", tc.description)

		var msg model.Message
		basePod.Spec.Containers[0].SecContext = &tc.securityContext
		setFdPodMsg(&msg, basePod)

		if err = msgValidator.Check(&msg); err != nil {
			hwlog.RunLog.Errorf("check msg failed: %v", err)
		}

		if tc.shouldErr {
			convey.So(err.Error(), tc.assert, tc.expected)
		} else {
			convey.So(err, tc.assert, tc.expected)
		}
	}

}

type portMapTestCase struct {
	description string
	ports       []types.ContainerPort
	shouldErr   bool
	assert      convey.Assertion
	expected    interface{}
}

var portTestCase = []portMapTestCase{
	{
		description: "host ip name contain nil char",
		ports:       []types.ContainerPort{{HostPort: 1024, ContainerPort: 1024, Protocol: "TCP", HostIP: "127.0.0.1"}},
		shouldErr:   false,
		assert:      convey.ShouldEqual,
		expected:    nil,
	}, {
		description: "host ip name contain invalid char",
		ports: []types.ContainerPort{
			{Name: "-1", HostPort: 1024, ContainerPort: 1024, Protocol: "TCP", HostIP: "0.0.0.0"}},
		shouldErr: true,
		assert:    convey.ShouldContainSubstring,
		expected:  "Pod.Spec.Containers.Ports.Name",
	}, {
		description: "host protocol name contain invalid protocol",
		ports: []types.ContainerPort{
			{Name: "abc", HostPort: 1024, ContainerPort: 1024, Protocol: "SCP", HostIP: "0.0.0.0"}},
		shouldErr: true,
		assert:    convey.ShouldContainSubstring,
		expected:  "Pod.Spec.Containers[0].Ports[0].Protocol",
	}, {
		description: "host port contain invalid value 10",
		ports: []types.ContainerPort{
			{Name: "abc", HostPort: 10, ContainerPort: 1024, Protocol: "TCP", HostIP: "0.0.0.0"},
		},
		shouldErr: true,
		assert:    convey.ShouldContainSubstring,
		expected:  "Pod.Spec.Containers[0].Ports[0].HostPort",
	}, {
		description: "host port name contain invalid value 65536",
		ports: []types.ContainerPort{
			{Name: "abc", HostPort: 655356, ContainerPort: 1024, Protocol: "TCP", HostIP: "0.0.0.0"}},
		shouldErr: true,
		assert:    convey.ShouldContainSubstring,
		expected:  "Pod.Spec.Containers[0].Ports[0].HostPort",
	}, {
		description: "test port dupicate failed",
		ports: []types.ContainerPort{
			{Name: "1", HostPort: 1024, ContainerPort: 1024, Protocol: "TCP", HostIP: "127.0.0.1"},
			{Name: "1", HostPort: 1024, ContainerPort: 1024, Protocol: "TCP", HostIP: "127.0.0.1"},
		},
		shouldErr: true,
		assert:    convey.ShouldContainSubstring,
		expected:  "duplicated host port",
	}, {
		description: "host ip all zero error",
		ports: []types.ContainerPort{
			{Name: "1", HostPort: 1024, ContainerPort: 1024, Protocol: "TCP", HostIP: "0.0.0.0"},
		},
		shouldErr: true,
		assert:    convey.ShouldContainSubstring,
		expected:  "IP can't be an all zeros address",
	}, {
		description: "host ip all 255 error",
		ports: []types.ContainerPort{
			{Name: "1", HostPort: 1024, ContainerPort: 1024, Protocol: "TCP", HostIP: "255.255.255.255"},
		},
		shouldErr: true,
		assert:    convey.ShouldContainSubstring,
		expected:  "IP can't be a broadcast address",
	},
}

func testPodPortMapPara() {
	var basePod = getPodInfo()

	var err error
	msgValidator := NewMsgValidator(msglistchecker.NewFdMsgHeaderValidator())

	for _, tc := range portTestCase {
		hwlog.RunLog.Infof("--------------------%s-------------------", tc.description)

		var msg model.Message
		basePod.Spec.Containers[0].Ports = tc.ports
		setFdPodMsg(&msg, basePod)

		if err = msgValidator.Check(&msg); err != nil {
			hwlog.RunLog.Errorf("check msg failed: %v", err)
		}

		if tc.shouldErr {
			convey.So(err.Error(), tc.assert, tc.expected)
		} else {
			convey.So(err, tc.assert, tc.expected)
		}
	}

}

func testPodAvailableResources() {

	var basePod = getPodInfo()

	msgValidator := NewMsgValidator(msglistchecker.NewFdMsgHeaderValidator())

	// 节点可用的容器cpu数量为3，配置为3，npu总量为2， 配置为2，memory总量为10240Mi, 配置10240Mi校验通过
	basePod.Spec.Containers[0].Resources.Req["cpu"] = resource.MustParse("3")
	basePod.Spec.Containers[0].Resources.Lim["cpu"] = resource.MustParse("3")
	basePod.Spec.Containers[0].Resources.Req["memory"] = resource.MustParse("10240Mi")
	basePod.Spec.Containers[0].Resources.Lim["memory"] = resource.MustParse("10240Mi")
	basePod.Spec.Containers[0].Resources.Req["huawei.com/Ascend310"] = resource.MustParse("2")
	basePod.Spec.Containers[0].Resources.Lim["huawei.com/Ascend310"] = resource.MustParse("2")

	var msg model.Message
	setFdPodMsg(&msg, basePod)

	var err error
	if err = msgValidator.Check(&msg); err != nil {
		hwlog.RunLog.Errorf("check msg failed: %v", err)
	}
	convey.So(err, convey.ShouldEqual, nil)

	// 节点可用的容器cpu数量为3，配置为3.01，校验失败
	basePod.Spec.Containers[0].Resources.Req["cpu"] = resource.MustParse("3.01")
	basePod.Spec.Containers[0].Resources.Lim["cpu"] = resource.MustParse("3.01")
	setFdPodMsg(&msg, basePod)

	if err = msgValidator.Check(&msg); err != nil {
		hwlog.RunLog.Errorf("check msg failed: %v", err)
	}
	convey.So(err, convey.ShouldNotEqual, nil)

	// 节点可用的容器memory数量为10240Mi，配置为10241Mi，校验失败
	basePod.Spec.Containers[0].Resources.Req["memory"] = resource.MustParse("10241Mi")
	basePod.Spec.Containers[0].Resources.Lim["memory"] = resource.MustParse("10241Mi")
	setFdPodMsg(&msg, basePod)

	if err = msgValidator.Check(&msg); err != nil {
		hwlog.RunLog.Errorf("check msg failed: %v", err)
	}
	convey.So(err, convey.ShouldNotEqual, nil)

	// 节点可用的容器cpu数量为2，配置为2.01，校验通过
	basePod.Spec.Containers[0].Resources.Req["huawei.com/Ascend310"] = resource.MustParse("2.01")
	basePod.Spec.Containers[0].Resources.Lim["huawei.com/Ascend310"] = resource.MustParse("2.01")
	setFdPodMsg(&msg, basePod)

	if err = msgValidator.Check(&msg); err != nil {
		hwlog.RunLog.Errorf("check msg failed: %v", err)
	}
	convey.So(err, convey.ShouldNotEqual, nil)
}

type podTestCase struct {
	description string
	deltaFunc   func(pod *types.Pod)
	shouldErr   bool
	assert      convey.Assertion
	expected    interface{}
}

var trueValue = true
var rootUid int64 = 0
var errDeletionGracePeriodSeconds int64 = 121
var podSecurityContextTestcases = []podTestCase{
	{
		description: "HostIPC should not be allowed",
		deltaFunc:   func(pod *types.Pod) { pod.Spec.HostIPC = true },
		shouldErr:   true, assert: convey.ShouldContainSubstring, expected: "Pod.Spec.HostIPC",
	},
	{
		description: "ShareProcessNs should not be allowed",
		deltaFunc:   func(pod *types.Pod) { pod.Spec.ShareProcessNs = &trueValue },
		shouldErr:   true, assert: convey.ShouldContainSubstring,
		expected: "Field validation for 'ShareProcessNs' failed on the 'isdefault' tag",
	},
	{
		description: "ServiceAccount should not be allowed",
		deltaFunc:   func(pod *types.Pod) { pod.Spec.SvcAccountToken = &trueValue },
		shouldErr:   true, assert: convey.ShouldContainSubstring, expected: "Pod.Spec.SvcAccountToken",
	},
	{
		description: "ServiceAccount name should be empty",
		deltaFunc:   func(pod *types.Pod) { pod.Spec.SvcAccountName = "serviceAccount" },
		shouldErr:   true, assert: convey.ShouldContainSubstring, expected: "Pod.Spec.SvcAccountName",
	},
	{
		description: "SELinuxOptions should be empty",
		deltaFunc:   func(pod *types.Pod) { pod.Spec.SecurityContext.SELinuxOptions = &v1.SELinuxOptions{} },
		shouldErr:   true, assert: convey.ShouldContainSubstring,
		expected: "Field validation for 'SELinuxOptions' failed on the 'isdefault' tag",
	},
	{
		description: "WindowsOptions should be empty",
		deltaFunc:   func(pod *types.Pod) { pod.Spec.SecurityContext.WindowsOptions = &v1.WindowsSecurityContextOptions{} },
		shouldErr:   true, assert: convey.ShouldContainSubstring,
		expected: "Field validation for 'WindowsOptions' failed on the 'isdefault' tag",
	},
	{
		description: "RunAsUser should not be set",
		deltaFunc:   func(pod *types.Pod) { pod.Spec.SecurityContext.RunAsUser = &rootUid },
		shouldErr:   true, assert: convey.ShouldContainSubstring,
		expected: "Field validation for 'RunAsUser' failed on the 'isdefault' tag",
	},
	{
		description: "RunAsGroup should not be set",
		deltaFunc:   func(pod *types.Pod) { pod.Spec.SecurityContext.RunAsGroup = &rootUid },
		shouldErr:   true, assert: convey.ShouldContainSubstring,
		expected: "Field validation for 'RunAsGroup' failed on the 'isdefault' tag",
	},
	{
		description: "SupplementalGroups should not be set",
		deltaFunc:   func(pod *types.Pod) { pod.Spec.SecurityContext.SupplementalGroups = []int64{rootUid} },
		shouldErr:   true, assert: convey.ShouldContainSubstring,
		expected: "Field validation for 'SupplementalGroups' failed on the 'isdefault' tag",
	},
	{
		description: "FSGroup should not be set",
		deltaFunc:   func(pod *types.Pod) { pod.Spec.SecurityContext.FSGroup = &rootUid },
		shouldErr:   true, assert: convey.ShouldContainSubstring,
		expected: "Field validation for 'FSGroup' failed on the 'isdefault' tag",
	},
	{
		description: "Sysctls should not be set",
		deltaFunc:   func(pod *types.Pod) { pod.Spec.SecurityContext.Sysctls = []v1.Sysctl{{Name: "a", Value: "b"}} },
		shouldErr:   true, assert: convey.ShouldContainSubstring,
		expected: "Field validation for 'Sysctls' failed on the 'isdefault' tag",
	},
	{
		description: "SeccompProfile should not be set",
		deltaFunc:   func(pod *types.Pod) { pod.Spec.SecurityContext.SeccompProfile = &v1.SeccompProfile{} },
		shouldErr:   true, assert: convey.ShouldContainSubstring,
		expected: "Field validation for 'SeccompProfile' failed on the 'isdefault' tag",
	},
	{
		description: "initContainers should not be allowed",
		deltaFunc:   func(pod *types.Pod) { pod.Spec.InitContainers = []types.Container{{}} },
		shouldErr:   true, assert: convey.ShouldContainSubstring,
		expected: "Pod.Spec.InitContainers",
	},
	{
		description: "ephemeralContainers should not be allowed",
		deltaFunc:   func(pod *types.Pod) { pod.Spec.EphemeralConts = []types.Container{{}} },
		shouldErr:   true, assert: convey.ShouldContainSubstring,
		expected: "Pod.Spec.EphemeralConts",
	},
	{
		description: "deletionGracePeriodSeconds is invalid",
		deltaFunc:   func(pod *types.Pod) { pod.ObjectMeta.DeletionGracePeriodSeconds = &errDeletionGracePeriodSeconds },
		shouldErr:   true, assert: convey.ShouldContainSubstring,
		expected: "Field validation for 'DeletionGracePeriodSeconds' failed on the 'max' tag",
	},
}

func testCheckPodSecurityContext() {
	var err error
	msgValidator := NewMsgValidator(msglistchecker.NewFdMsgHeaderValidator())

	for _, tc := range podSecurityContextTestcases {
		hwlog.RunLog.Infof("--------------------%s-------------------", tc.description)

		basePod := getPodInfo()
		var msg model.Message
		if tc.deltaFunc != nil {
			tc.deltaFunc(&basePod)
		}
		setFdPodMsg(&msg, basePod)

		if err = msgValidator.Check(&msg); err != nil {
			hwlog.RunLog.Errorf("check msg failed: %v", err)
		}

		if tc.shouldErr {
			convey.So(err.Error(), tc.assert, tc.expected)
		} else {
			convey.So(err, tc.assert, tc.expected)
		}
	}
}

var procMountTypeDefault = v1.DefaultProcMount

type containerTestCase struct {
	description string
	deltaFunc   func(container *types.Container)
	shouldErr   bool
	assert      convey.Assertion
	expected    interface{}
}

var containerSecurityContextTestcases = []containerTestCase{
	{
		description: "SELinuxOptions should be empty",
		deltaFunc:   func(container *types.Container) { container.SecContext.SELinuxOptions = &v1.SELinuxOptions{} },
		shouldErr:   true, assert: convey.ShouldContainSubstring,
		expected: "Field validation for 'SELinuxOptions' failed on the 'isdefault' tag",
	},
	{
		description: "WindowsOptions should be empty",
		deltaFunc: func(container *types.Container) {
			container.SecContext.WindowsOptions = &v1.WindowsSecurityContextOptions{}
		},
		shouldErr: true, assert: convey.ShouldContainSubstring,
		expected: "Field validation for 'WindowsOptions' failed on the 'isdefault' tag",
	},
	{
		description: "ProMount should not be set",
		deltaFunc:   func(container *types.Container) { container.SecContext.ProcMount = &procMountTypeDefault },
		shouldErr:   true, assert: convey.ShouldContainSubstring,
		expected: "Field validation for 'ProcMount' failed on the 'isdefault' tag",
	},
}

func testCheckContainerSecurityContext() {
	var err error
	msgValidator := NewMsgValidator(msglistchecker.NewFdMsgHeaderValidator())

	for _, tc := range containerSecurityContextTestcases {
		hwlog.RunLog.Infof("--------------------%s-------------------", tc.description)

		basePod := getPodInfo()
		var msg model.Message
		if tc.deltaFunc != nil {
			tc.deltaFunc(&basePod.Spec.Containers[0])
		}
		setFdPodMsg(&msg, basePod)

		if err = msgValidator.Check(&msg); err != nil {
			hwlog.RunLog.Errorf("check msg failed: %v", err)
		}

		if tc.shouldErr {
			convey.So(err.Error(), tc.assert, tc.expected)
		} else {
			convey.So(err, tc.assert, tc.expected)
		}
	}
}
