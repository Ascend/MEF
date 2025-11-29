/* Copyright(C) 2022. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Package server holds the implementation of registration to kubelet, k8s device plugin interface and grpc service.
package server

import (
	"context"
	"fmt"

	"huawei.com/mindx/common/hwlog"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/device"
)

func (ps *PluginServer) stopListAndWatch() {
	if ps.isRunning.Load() {
		ps.stop <- struct{}{}
	}
}

// Notify is called when device status changed, to notify ListAndWatch
func (ps *PluginServer) Notify(devices []*common.NpuDevice) bool {
	if ps == nil {
		hwlog.RunLog.Error("invalid interface receiver")
		return false
	}
	if ps.isRunning.Load() {
		ps.deepCopyDevice(devices)
		ps.reciChan <- struct{}{}
		return true
	}
	return false
}

func sendToKubelet(stream v1beta1.DevicePlugin_ListAndWatchServer, resp *v1beta1.ListAndWatchResponse) error {
	return stream.Send(resp)
}

func (ps *PluginServer) responseToKubelet() *v1beta1.ListAndWatchResponse {
	resp := new(v1beta1.ListAndWatchResponse)
	ps.cachedLock.RLock()
	for _, dev := range ps.cachedDevices {
		hwlog.RunLog.Infof("ListAndWatch resp devices: %s %s", dev.DeviceName, dev.Health)
		resp.Devices = append(resp.Devices, &v1beta1.Device{ID: dev.DeviceName, Health: dev.Health})
	}

	ps.cachedLock.RUnlock()
	return resp
}

func (ps *PluginServer) deepCopyDevice(cachedDevices []*common.NpuDevice) {
	ps.cachedLock.Lock()
	ps.cachedDevices = ps.cachedDevices[:0]
	for _, dev := range cachedDevices {
		ps.cachedDevices = append(ps.cachedDevices, common.NpuDevice{
			DeviceName: dev.DeviceName,
			Health:     dev.Health,
			PhyID:      dev.PhyID,
		})
	}
	ps.cachedLock.Unlock()
}

// ListAndWatch is to send device info to kubelet
func (ps *PluginServer) ListAndWatch(empty *v1beta1.Empty, stream v1beta1.DevicePlugin_ListAndWatchServer) error {
	send := func(stream v1beta1.DevicePlugin_ListAndWatchServer) {
		if err := sendToKubelet(stream, ps.responseToKubelet()); err != nil {
			hwlog.RunLog.Errorf("send to kubelet failed, error is %#v", err)
		}
	}
	ps.isRunning.Store(true)
	send(stream)
	for {
		select {
		case <-ps.stop:
			ps.isRunning.Store(false)
			return nil
		case _, ok := <-ps.reciChan:
			if ok {
				send(stream)
			}
		}
	}
}

func (ps *PluginServer) deviceExists(id string) bool {
	ps.cachedLock.RLock()
	defer ps.cachedLock.RUnlock()
	for _, d := range ps.cachedDevices {
		if d.DeviceName == id {
			return true
		}
	}
	return false
}

func (ps *PluginServer) checkAllocateRequest(requests *v1beta1.AllocateRequest) error {
	if requests == nil {
		return fmt.Errorf("invalid requests")
	}
	if len(requests.ContainerRequests) > common.MaxContainerLimit {
		return fmt.Errorf("the number of container request %d exceeds the upper limit",
			len(requests.ContainerRequests))
	}
	for _, rqt := range requests.ContainerRequests {
		if len(rqt.DevicesIDs) > common.MaxDevicesNum*common.MinAICoreNum {
			return fmt.Errorf("the devices can't bigger than %d", common.MaxDevicesNum)
		}
		for _, deviceName := range rqt.DevicesIDs {
			if len(deviceName) > common.MaxDeviceNameLen {
				return fmt.Errorf("length of device name %d is invalid", len(deviceName))
			}
			if !ps.deviceExists(deviceName) {
				return fmt.Errorf("plugin doesn't have device %s", deviceName)
			}
			if common.IsVirtualDev(deviceName) && len(rqt.DevicesIDs) > common.MaxRequestVirtualDeviceNum {
				return fmt.Errorf("request more than %d virtual device, current is %d",
					common.MaxRequestVirtualDeviceNum, len(rqt.DevicesIDs))
			}
			if common.IsVirtualDev(deviceName) {
				ps.ascendRuntimeOptions = common.VirtualDev
				return nil
			}
		}
	}
	return nil
}

func getDevPath(id, ascendRuntimeOptions string) (string, string) {
	containerPath := fmt.Sprintf("%s%s", "/dev/davinci", id)
	hostPath := containerPath
	if ascendRuntimeOptions == common.VirtualDev {
		hostPath = fmt.Sprintf("%s%s", "/dev/vdavinci", id)
	}
	return containerPath, hostPath
}

func mountDevice(resp *v1beta1.ContainerAllocateResponse, devices []int, ascendRuntimeOptions string) {
	for _, deviceID := range devices {
		containerPath, hostPath := getDevPath(fmt.Sprintf("%d", deviceID), ascendRuntimeOptions)
		resp.Devices = append(resp.Devices, &v1beta1.DeviceSpec{
			HostPath:      hostPath,
			ContainerPath: containerPath,
			Permissions:   "rw",
		})
	}
}

func mountDefaultDevice(resp *v1beta1.ContainerAllocateResponse, defaultDevs []string) {
	// mount default devices
	for _, d := range defaultDevs {
		resp.Devices = append(resp.Devices, &v1beta1.DeviceSpec{
			HostPath:      d,
			ContainerPath: getDeviceContainerPath(d),
			Permissions:   "rw",
		})
	}
}

func getDeviceContainerPath(hostPath string) string {
	if hostPath == common.HiAIManagerDeviceDocker {
		return common.HiAIManagerDevice
	}
	return hostPath
}

// Allocate is called by kubelet to mount device to k8s pod.
func (ps *PluginServer) Allocate(ctx context.Context, requests *v1beta1.AllocateRequest) (*v1beta1.AllocateResponse,
	error) {
	if err := ps.checkAllocateRequest(requests); err != nil {
		hwlog.RunLog.Error(err)
		return nil, err
	}
	resps := new(v1beta1.AllocateResponse)
	for _, rqt := range requests.ContainerRequests {
		var err error
		allocateDevices := rqt.DevicesIDs
		hwlog.RunLog.Infof("request: %#v", rqt.DevicesIDs)

		_, ascendVisibleDevices, err := common.GetDeviceListID(allocateDevices, ps.ascendRuntimeOptions)
		if err != nil {
			hwlog.RunLog.Error(err)
			return nil, err
		}

		resp := new(v1beta1.ContainerAllocateResponse)

		hwlog.RunLog.Info("device-plugin will use origin mount way")
		mountDefaultDevice(resp, ps.defaultDevs)
		mountDevice(resp, ascendVisibleDevices, ps.ascendRuntimeOptions)

		resps.ContainerResponses = append(resps.ContainerResponses, resp)
	}
	return resps, nil
}

// GetPreferredAllocation implement the kubelet device plugin interface
func (ps *PluginServer) GetPreferredAllocation(context.Context, *v1beta1.PreferredAllocationRequest) (
	*v1beta1.PreferredAllocationResponse, error) {
	return nil, fmt.Errorf("not support")
}

// GetDevicePluginOptions is Standard interface to kubelet.
func (ps *PluginServer) GetDevicePluginOptions(ctx context.Context, e *v1beta1.Empty) (*v1beta1.DevicePluginOptions,
	error) {
	return &v1beta1.DevicePluginOptions{}, nil
}

// PreStartContainer is Standard interface to kubelet with empty implement.
func (ps *PluginServer) PreStartContainer(ctx context.Context,
	r *v1beta1.PreStartContainerRequest) (*v1beta1.PreStartContainerResponse, error) {
	hwlog.RunLog.Info("PreStart just call in UT.")
	return &v1beta1.PreStartContainerResponse{}, nil
}

// NewPluginServer returns an initialized PluginServer
func NewPluginServer(deviceType string, devices []*common.NpuDevice, defaultDevs []string,
	manager device.DevManager) *PluginServer {
	ps := &PluginServer{
		restart:        true,
		reciChan:       make(chan interface{}),
		deviceType:     deviceType,
		defaultDevs:    defaultDevs,
		stop:           make(chan interface{}),
		klt2RealDevMap: make(map[string]string, common.MaxDevicesNum),
		isRunning:      common.NewAtomicBool(false),
		manager:        manager,
	}
	ps.deepCopyDevice(devices)
	return ps
}
