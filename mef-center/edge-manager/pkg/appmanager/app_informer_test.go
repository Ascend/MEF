// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package appmanager

import (
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/test"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"edge-manager/pkg/kubeclient"
	"huawei.com/mindxedge/base/common"
)

var (
	pod       *corev1.Pod
	pod2      *corev1.Pod
	daemonSet *appv1.DaemonSet
)

func TestInformer(t *testing.T) {
	initInformerTestEnv()
	convey.Convey("test addPod\n", t, testAddPod)
	convey.Convey("test updatePod\n", t, testUpdatePod)
	convey.Convey("test deletePod\n", t, testDeletePod)
	convey.Convey("test deleteTerminatingPod\n", t, testDeleteTerminatingPod)
	convey.Convey("test initAppStatusService\n", t, testInitAppStatusService)
	convey.Convey("test initDefaultImagePullSecret\n", t, testInitDefaultImagePullSecret)
	convey.Convey("test getContainerInfo\n", t, testGetContainerInfo)
	convey.Convey("test getContainerStatus \n", t, testGetContainerStatus)
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
  	        "name":"afafda",
      	    "userId":1024
		}]
	}`
	resp := createApp(&model.Message{Content: []byte(reqData)})
	var appInfo AppInfo
	test.MockGetDb().Model(AppInfo{}).Where("id = ?", resp.Data.(uint64)).Find(&appInfo)
	daemonSet, _ = initDaemonSet(&appInfo, 1)
}

func testAddPod() {
	var p1 = gomonkey.ApplyFuncReturn(getNodeInfoByUniqueName,
		uint64(1), "ut-node", nil,
	)
	defer p1.Reset()

	var count1, count2 int64
	test.MockGetDb().Model(AppInstance{}).Count(&count1)
	appStatusService.addPod(pod)
	test.MockGetDb().Model(AppInstance{}).Count(&count2)
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
	test.MockGetDb().Model(AppInstance{}).Count(&count1)
	appStatusService.deletePod(pod)
	test.MockGetDb().Model(AppInstance{}).Count(&count2)
	convey.So(count1, convey.ShouldNotEqual, count2)
}

func testDeleteTerminatingPod() {
	testFlag := "test flag"
	appStatusService.podInformer = informers.NewSharedInformerFactory(&kubernetes.Clientset{}, time.Second).
		Core().V1().Pods().Informer()
	store := appStatusService.podInformer.GetStore()
	toDeletePod := *pod
	toDeletePod.DeletionTimestamp = &metav1.Time{Time: time.Now()}
	var patches = gomonkey.ApplyMethodReturn(store, "List", []interface{}{&toDeletePod}).
		ApplyMethod(&kubeclient.Client{}, "DeletePodByForce",
			func(client *kubeclient.Client, pod *corev1.Pod) error {
				if pod != nil {
					pod.APIVersion = testFlag
				}
				return test.ErrTest
			})
	defer patches.Reset()
	appStatusService.deleteTerminatingPod()
	convey.So(toDeletePod.APIVersion, convey.ShouldEqual, testFlag)
}

func testInitAppStatusService() {
	GetClientSetOps := []gomonkey.OutputCell{
		{Values: gomonkey.Params{nil}, Times: 1},
		{Values: gomonkey.Params{&kubernetes.Clientset{}}, Times: 5},
	}
	CreateNamespaceOps := []gomonkey.OutputCell{
		{Values: gomonkey.Params{nil, test.ErrTest}, Times: 1},
		{Values: gomonkey.Params{&corev1.Namespace{}, nil}, Times: 4},
	}
	WaitForCacheSyncOps := []gomonkey.OutputCell{
		{Values: gomonkey.Params{false}, Times: 1},
		{Values: gomonkey.Params{true}, Times: 2},
	}
	initDefaultImagePullSecretOps := []gomonkey.OutputCell{
		{Values: gomonkey.Params{test.ErrTest}, Times: 1},
		{Values: gomonkey.Params{nil}, Times: 1},
	}
	patches := gomonkey.ApplyMethodSeq(&kubeclient.Client{}, "GetClientSet", GetClientSetOps).
		ApplyMethodSeq(&kubeclient.Client{}, "CreateNamespace", CreateNamespaceOps).
		ApplyPrivateMethod(&AppRepositoryImpl{}, "deleteAllRemainingInstance", func() error { return nil }).
		ApplyFuncSeq(cache.WaitForCacheSync, WaitForCacheSyncOps).
		ApplyFuncSeq(initDefaultImagePullSecret, initDefaultImagePullSecretOps).
		ApplyPrivateMethod(informers.NewSharedInformerFactory(nil, time.Second), "Start", func(stopCh <-chan struct{}) {})
	defer patches.Reset()

	fmt.Println("\ncase: get k8s client set failed")
	err := appStatusService.initAppStatusService()
	convey.So(err, convey.ShouldResemble, errors.New("get k8s client set failed"))

	fmt.Println("\ncase: create default user namespace failed")
	err = appStatusService.initAppStatusService()
	convey.So(err, convey.ShouldResemble, fmt.Errorf("create default user namespace failed: %v", test.ErrTest))

	fmt.Println("\ncase: run infomer failed")
	err = appStatusService.initAppStatusService()
	convey.So(err, convey.ShouldResemble, errors.New("sync app status service pod caches error"))

	fmt.Println("\ncase: init image pull secret failed")
	err = appStatusService.initAppStatusService()
	convey.So(err, convey.ShouldResemble, test.ErrTest)

	fmt.Println("\ncase: all successful")
	err = appStatusService.initAppStatusService()
	convey.So(err, convey.ShouldBeNil)
}

func testInitDefaultImagePullSecret() {
	GetSecretOps := []gomonkey.OutputCell{
		{Values: gomonkey.Params{&corev1.Secret{}, nil}, Times: 1},
		{Values: gomonkey.Params{&corev1.Secret{Data: map[string][]byte{corev1.DockerConfigJsonKey: nil}}, nil}, Times: 1},
		{Values: gomonkey.Params{nil, test.ErrTest}, Times: 1},
		{Values: gomonkey.Params{nil, errors.New(kubeclient.K8sNotFoundErrorFragment)}, Times: 2},
	}
	CreateOrUpdateSecretOps := []gomonkey.OutputCell{
		{Values: gomonkey.Params{nil, test.ErrTest}, Times: 1},
		{Values: gomonkey.Params{&corev1.Secret{}, nil}, Times: 1},
	}
	patches := gomonkey.ApplyMethodSeq(&kubeclient.Client{}, "GetSecret", GetSecretOps).
		ApplyMethodSeq(&kubeclient.Client{}, "CreateOrUpdateSecret", CreateOrUpdateSecretOps)
	defer patches.Reset()

	fmt.Println("\ncase: empty secret")
	err := initDefaultImagePullSecret()
	convey.So(err, convey.ShouldBeNil)

	fmt.Println("\ncase: normal secret")
	err = initDefaultImagePullSecret()
	convey.So(err, convey.ShouldBeNil)

	fmt.Println("\ncase: get secret form k8s failed")
	err = initDefaultImagePullSecret()
	convey.So(err, convey.ShouldResemble, errors.New("check image pull secret failed"))

	fmt.Println("\ncase: create secret to k8s failed")
	err = initDefaultImagePullSecret()
	convey.So(err, convey.ShouldResemble, errors.New("create default image pull secret failed"))

	fmt.Println("\ncase: create new secret successful")
	err = initDefaultImagePullSecret()
	convey.So(err, convey.ShouldBeNil)
}

func testGetContainerInfo() {
	instance := AppInstance{
		ContainerInfo: `[{"name":"testContainer","image":"testImage:v1","status":"running","restartCount":0}]`,
	}
	fmt.Println("\ncase: create new secret successful")
	info, err := appStatusService.getContainerInfos(instance, nodeStatusUnknown)
	convey.So(err, convey.ShouldBeNil)
	convey.So(len(info), convey.ShouldEqual, 1)
	convey.So(info[0].Status, convey.ShouldEqual, containerStateUnknown)

	fmt.Println("\ncase: unmarshal failed")
	_, err = appStatusService.getContainerInfos(AppInstance{ContainerInfo: "error-json-string"}, nodeStatusUnknown)
	convey.So(err, convey.ShouldResemble, errors.New("unmarshal app container info failed"))
}

func testGetContainerStatus() {
	k8sContainerStatus := corev1.ContainerStatus{}
	status := getContainerStatus(k8sContainerStatus)
	convey.So(status, convey.ShouldEqual, containerStateUnknown)

	k8sContainerStatus.State = corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{}}
	status = getContainerStatus(k8sContainerStatus)
	convey.So(status, convey.ShouldEqual, containerStateWaiting)

	k8sContainerStatus.State = corev1.ContainerState{Running: &corev1.ContainerStateRunning{}}
	status = getContainerStatus(k8sContainerStatus)
	convey.So(status, convey.ShouldEqual, containerStateRunning)

	k8sContainerStatus.State = corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{}}
	status = getContainerStatus(k8sContainerStatus)
	convey.So(status, convey.ShouldEqual, containerStateTerminated)
}
