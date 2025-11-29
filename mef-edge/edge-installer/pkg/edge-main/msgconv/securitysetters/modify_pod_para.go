// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package securitysetters
package securitysetters

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"edge-installer/pkg/edge-main/common/configpara"
)

func isPodGraceDelete(deletionTimestamp *metav1.Time) bool {
	return deletionTimestamp != nil
}

func setPodVolumes(podSpec, podSpecTmp *corev1.PodSpec) {
	for _, volume := range podSpec.Volumes {
		var volumeTmp = corev1.Volume{Name: volume.Name}
		if volume.HostPath != nil {
			volumeTmp.HostPath = volume.HostPath
		} else if volume.EmptyDir != nil {
			volumeTmp.EmptyDir = &corev1.EmptyDirVolumeSource{
				Medium: volume.EmptyDir.Medium,
			}
		} else if volume.ConfigMap != nil {
			volumeTmp.ConfigMap = &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: volume.ConfigMap.Name},
				DefaultMode:          volume.ConfigMap.DefaultMode}
		} else {
			continue
		}
		podSpecTmp.Volumes = append(podSpecTmp.Volumes, volumeTmp)
	}

}

func setPodTolerations(podSpec, podSpecTmp *corev1.PodSpec) {
	for _, toleration := range podSpec.Tolerations {
		podSpecTmp.Tolerations = append(podSpecTmp.Tolerations, corev1.Toleration{
			Key:      toleration.Key,
			Operator: toleration.Operator,
			Effect:   toleration.Effect})
	}
}

func setPodHostPID(_, podSpecTmp *corev1.PodSpec) {
	if configpara.GetPodConfig().HostPid == true {
		podSpecTmp.HostPID = true
	}
}

func modifyPodSpecCreatePara(pod *corev1.Pod) {
	var podSpecTmp = corev1.PodSpec{
		RestartPolicy:                 pod.Spec.RestartPolicy,
		TerminationGracePeriodSeconds: pod.Spec.TerminationGracePeriodSeconds,
		DNSPolicy:                     pod.Spec.DNSPolicy,
		ServiceAccountName:            pod.Spec.ServiceAccountName,
		HostNetwork:                   pod.Spec.HostNetwork,
		ImagePullSecrets:              pod.Spec.ImagePullSecrets,
		SecurityContext:               &corev1.PodSecurityContext{},
		SchedulerName:                 pod.Spec.SchedulerName,
	}

	var podSpecSetFuncs = []func(podSpec, podSpecTmp *corev1.PodSpec){
		setContainerPara,
		setPodVolumes,
		setPodTolerations,
		setPodHostPID,
	}

	for _, setFunc := range podSpecSetFuncs {
		setFunc(&pod.Spec, &podSpecTmp)
	}

	pod.Spec = podSpecTmp
}

func modifyPodSpecGraceDeletionPara(pod *corev1.Pod) {
	var spec corev1.PodSpec
	for i := 0; i < len(pod.Spec.Containers); i++ {
		spec.Containers = append(spec.Containers, corev1.Container{Name: pod.Spec.Containers[i].Name})
	}
	pod.Spec = spec
}

func modifyPodSpecDeletePara(pod *corev1.Pod) {
	var spec corev1.PodSpec
	for i := 0; i < len(pod.Spec.Containers); i++ {
		spec.Containers = append(spec.Containers, corev1.Container{Name: pod.Spec.Containers[i].Name})
	}
	pod.Spec = spec
}

func modifyPodMetaDataCreatePara(pod *corev1.Pod) {
	pod.ObjectMeta = *newObjectMeta(pod.ObjectMeta)
}

func modifyPodMetaDataGraceDeletionPara(pod *corev1.Pod) {
	newPodObjectMeta := *newObjectMeta(pod.ObjectMeta)
	newPodObjectMeta.DeletionTimestamp = pod.DeletionTimestamp
	newPodObjectMeta.DeletionGracePeriodSeconds = pod.DeletionGracePeriodSeconds

	pod.ObjectMeta = newPodObjectMeta
}

func modifyPodMetaDataDeletePara(pod *corev1.Pod) {
	modifyPodMetaDataCreatePara(pod)
}

func modifyPodStatusAndTypeMeta(pod *corev1.Pod) {
	pod.Status = corev1.PodStatus{}
	pod.TypeMeta = metav1.TypeMeta{}
}
