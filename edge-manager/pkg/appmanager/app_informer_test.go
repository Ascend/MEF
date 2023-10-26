// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

package appmanager

import (
	"strconv"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"huawei.com/mindxedge/base/common"
)

var (
	pod       *corev1.Pod
	pod2      *corev1.Pod
	daemonSet *appv1.DaemonSet
)

func TestInformer(t *testing.T) {
	initInformerTestEnv()
	convey.Convey("addPod pod should success", t, testAddPod)
	convey.Convey("update pod pod should success", t, testUpdatePod)
	convey.Convey("delete pod pod should success", t, testDeletePod)
}

func initInformerTestEnv() {
	initAppStatusService()
	initTestPod()
	initTestDaemonSet()
}

func initAppStatusService() {
	appStatusService.podStatusCache = make(map[string]string)
	appStatusService.containerStatusCache = make(map[string]containerStatus)
}

func initTestPod() {
	container := corev1.Container{
		Name:            "ut-container1",
		Image:           "ubuntu:22.04",
		Command:         nil,
		Args:            nil,
		Ports:           nil,
		Env:             nil,
		Resources:       corev1.ResourceRequirements{},
		ImagePullPolicy: corev1.PullIfNotPresent,
	}
	spec := corev1.PodSpec{
		Containers:    []corev1.Container{container},
		RestartPolicy: corev1.RestartPolicyOnFailure,
		NodeSelector:  map[string]string{"MEF-Node1": ""},
		NodeName:      "",
	}

	pod = &corev1.Pod{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: "ut-app-1-1",
			Labels: map[string]string{
				common.AppManagerName: AppLabel,
				AppName:               "ut-app",
				AppId:                 strconv.FormatInt(int64(1), DecimalScale),
			},
		},
		Spec: spec,
		Status: corev1.PodStatus{
			Phase:                      corev1.PodRunning,
			Conditions:                 nil,
			ContainerStatuses:          nil,
			EphemeralContainerStatuses: nil,
		},
	}
}

func initTestDaemonSet() {
	reqData := `{
  	"appName":"test-app-for-daemonset",
  	"description":"",
	    "containers":[{
  	        "args":[],
      	    "command":[],
          	"containerPort":[],
				"memRequest": 1024,
  	        "cpuRequest": 1,
      	    "env":[],
	            "groupId":1024,
  	        "image":"euler_image",
      	    "imageVersion":"2.0",
	            "memRequest": 1024,
  	        "name":"afafda",
      	    "userId":1024
		}]
	}`
	resp := createApp(reqData)
	var appInfo AppInfo
	gormInstance.Model(AppInfo{}).Where("id = ?", resp.Data.(uint64)).Find(&appInfo)
	daemonSet, _ = initDaemonSet(&appInfo, 1)
}

func testAddPod() {
	var p1 = gomonkey.ApplyFuncReturn(getNodeInfoByUniqueName,
		uint64(1), "ut-node", nil,
	)
	defer p1.Reset()

	var count1, count2 int64
	gormInstance.Model(AppInstance{}).Count(&count1)
	appStatusService.addPod(pod)
	gormInstance.Model(AppInstance{}).Count(&count2)
	convey.So(count1, convey.ShouldNotEqual, count2)
}

func testUpdatePod() {
	pod2 = pod
	pod2.Status.Phase = corev1.PodPending
	status1 := appStatusService.getPodStatusFromCache("ut-app-1-1", "ready")
	appStatusService.updatePod(pod, pod2)
	status2 := appStatusService.getPodStatusFromCache("ut-app-1-1", "ready")
	convey.So(status1, convey.ShouldNotEqual, status2)
}

func testDeletePod() {
	var count1, count2 int64
	gormInstance.Model(AppInstance{}).Count(&count1)
	appStatusService.deletePod(pod)
	gormInstance.Model(AppInstance{}).Count(&count2)
	convey.So(count1, convey.ShouldNotEqual, count2)
}
