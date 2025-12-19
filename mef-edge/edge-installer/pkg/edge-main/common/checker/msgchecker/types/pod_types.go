// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package types for pod
package types

import (
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// PodPatch defines a structure for verifying kubeedge Pod patch parameters.
type PodPatch struct {
	Object Pod `json:"object"`
}

// Pod defines a structure for verifying kubeedge Pod parameters.
type Pod struct {
	TypeMeta   `json:",inline"`
	ObjectMeta `json:"metadata,omitempty"`
	Spec       PodSpec   `json:"spec,omitempty" binding:"omitempty"`
	Status     podStatus `json:"status,omitempty" binding:"omitempty"`
}

// PodSpec [struct] to describe PodSpec info
type PodSpec struct {
	Volumes        []Volume    `json:"volumes,omitempty" binding:"max=256,dive"`
	InitContainers []Container `json:"initContainers" binding:"eq=0"`
	Containers     []Container `json:"containers" binding:"max=10,dive"`
	EphemeralConts []Container `json:"ephemeralContainers" binding:"eq=0"`
	RestartPolicy  string      `json:"restartPolicy,omitempty" binding:"omitempty,oneof=Always OnFailure Never"`

	TermGracePeriod  *int64                 `json:"terminationGracePeriodSeconds,omitempty" binding:"omitempty,eq=30"`
	DeadlineSeconds  *int64                 `json:"activeDeadlineSeconds,omitempty" binding:"isdefault,omitempty"`
	DNSPolicy        string                 `json:"dnsPolicy,omitempty" binding:"omitempty,eq=ClusterFirst"`
	NodeSelector     map[string]string      `json:"nodeSelector,omitempty" binding:"max=20,dive,keys,max=256,startswith=MEF-Node,endkeys,max=1024"`
	SvcAccountName   string                 `json:"serviceAccountName,omitempty" binding:"omitempty,eq=default"`
	SvcAccount       string                 `json:"serviceAccount,omitempty" binding:"omitempty,eq=default"`
	SvcAccountToken  *bool                  `json:"automountServiceAccountToken,omitempty" binding:"omitempty,eq=false"`
	NodeName         string                 `json:"nodeName,omitempty" binding:"omitempty,min=1,max=256"`
	HostNetwork      bool                   `json:"hostNetwork,omitempty" binding:"omitempty"`
	HostPID          bool                   `json:"hostPID,omitempty" binding:"omitempty"`
	HostIPC          bool                   `json:"hostIPC,omitempty" binding:"omitempty,eq=false"`
	ShareProcessNs   *bool                  `json:"shareProcessNamespace,omitempty" binding:"isdefault,omitempty"`
	SecurityContext  *podSecurityContext    `json:"securityContext,omitempty" binding:"omitempty"`
	ImagePullSecrets []localObjectReference `json:"imagePullSecrets,omitempty" binding:"omitempty,len=1,dive"`
	Hostname         string                 `json:"hostname,omitempty" binding:"omitempty,max=36"`
	Subdomain        string                 `json:"subdomain,omitempty" binding:"isdefault,omitempty"`
	Affinity         *affinity              `json:"affinity,omitempty" binding:"omitempty"`
	SchedulerName    string                 `json:"schedulerName,omitempty" binding:"omitempty,eq=default-scheduler"`
	Tolerations      []toleration           `json:"tolerations,omitempty" binding:"omitempty,max=16,dive"`
	HostAliases      []v1.HostAlias         `json:"hostAliases,omitempty" binding:"omitempty,len=0"`
	PriorityCName    string                 `json:"priorityClassName,omitempty" binding:"isdefault,omitempty"`
	Priority         *int32                 `json:"priority,omitempty" binding:"omitempty,eq=0"`
	DNSConfig        *v1.PodDNSConfig       `json:"dnsConfig,omitempty" binding:"isdefault,omitempty"`
	ReadinessGates   []v1.PodReadinessGate  `json:"readinessGates,omitempty" binding:"omitempty,len=0"`
	RuntimeClassName *string                `json:"runtimeClassName,omitempty" binding:"isdefault,omitempty"`
	ServiceLinks     *bool                  `json:"enableServiceLinks,omitempty" binding:"omitempty,eq=true"`
	PreemptionPolicy *string                `json:"preemptionPolicy,omitempty" binding:"omitempty,eq=PreemptLowerPriority"`
	Overhead         v1.ResourceList        `json:"overhead,omitempty" binding:"omitempty,len=0"`

	TSConstraints  []v1.TopologySpreadConstraint `json:"topologySpreadConstraints,omitempty" binding:"omitempty,len=0"`
	HostnameAsFQDN *bool                         `json:"setHostnameAsFQDN,omitempty" binding:"isdefault,omitempty"`
}

// Volume [struct] to describe volume info
type Volume struct {
	Name         string `json:"name" validate:"^[a-zA-Z0-9-]{1,63}$"`
	VolumeSource `json:",inline"`
}

// EmptyDirVolumeSource [struct] to describe empty dir info
type EmptyDirVolumeSource struct {
	Medium    string             `json:"medium,omitempty" binding:"isdefault,omitempty"`
	SizeLimit *resource.Quantity `json:"sizeLimit,omitempty" binding:"isdefault,omitempty"`
}

// ConfigMapVolumeSource [struct] to describe config map info
type ConfigMapVolumeSource struct {
	Name        string         `json:"name" validate:"^[a-z][a-z0-9-]{2,62}[a-z0-9]$"`
	Items       []v1.KeyToPath `json:"items,omitempty" binding:"isdefault,omitempty"`
	DefaultMode *int32         `json:"defaultMode" binding:"eq=0644"`
	Optional    *bool          `json:"optional,omitempty" binding:"isdefault,omitempty"`
}

// VolumeSource [struct] to describe volumeSource info
type VolumeSource struct {
	HostPath             *HostPathVolumeSource                 `json:"hostPath,omitempty" binding:"omitempty"`
	EmptyDir             *EmptyDirVolumeSource                 `json:"emptyDir,omitempty" binding:"omitempty"`
	ConfigMap            *ConfigMapVolumeSource                `json:"configMap,omitempty" binding:"omitempty"`
	GCEPersistentDisk    *v1.GCEPersistentDiskVolumeSource     `json:"gcePersistentDisk,omitempty" binding:"isdefault"`
	AWSElasticBlockStr   *v1.AWSElasticBlockStoreVolumeSource  `json:"awsElasticBlockStore,omitempty" binding:"isdefault"`
	GitRepo              *v1.GitRepoVolumeSource               `json:"gitRepo,omitempty" binding:"isdefault"`
	Secret               *v1.SecretVolumeSource                `json:"secret,omitempty" binding:"isdefault"`
	NFS                  *v1.NFSVolumeSource                   `json:"nfs,omitempty" binding:"isdefault"`
	ISCSI                *v1.ISCSIVolumeSource                 `json:"iscsi,omitempty" binding:"isdefault"`
	Glusterfs            *v1.GlusterfsVolumeSource             `json:"glusterfs,omitempty" binding:"isdefault"`
	PstVolumeClaim       *v1.PersistentVolumeClaimVolumeSource `json:"persistentVolumeClaim,omitempty" binding:"isdefault"`
	RBD                  *v1.RBDVolumeSource                   `json:"rbd,omitempty" binding:"isdefault"`
	FlexVolume           *v1.FlexVolumeSource                  `json:"flexVolume,omitempty" binding:"isdefault"`
	Cinder               *v1.CinderVolumeSource                `json:"cinder,omitempty" binding:"isdefault"`
	CephFS               *v1.CephFSVolumeSource                `json:"cephfs,omitempty" binding:"isdefault"`
	Flocker              *v1.FlockerVolumeSource               `json:"flocker,omitempty" binding:"isdefault"`
	DownwardAPI          *v1.DownwardAPIVolumeSource           `json:"downwardAPI,omitempty" binding:"isdefault"`
	FC                   *v1.FCVolumeSource                    `json:"fc,omitempty" binding:"isdefault"`
	AzureFile            *v1.AzureFileVolumeSource             `json:"azureFile,omitempty" binding:"isdefault"`
	VsphereVolume        *v1.VsphereVirtualDiskVolumeSource    `json:"vsphereVolume,omitempty" binding:"isdefault"`
	Quobyte              *v1.QuobyteVolumeSource               `json:"quobyte,omitempty" binding:"isdefault"`
	AzureDisk            *v1.AzureDiskVolumeSource             `json:"azureDisk,omitempty" binding:"isdefault"`
	PhotonPersistentDisk *v1.PhotonPersistentDiskVolumeSource  `json:"photonPersistentDisk,omitempty" binding:"isdefault"`
	Projected            *v1.ProjectedVolumeSource             `json:"projected,omitempty" binding:"isdefault"`
	PortworxVolume       *v1.PortworxVolumeSource              `json:"portworxVolume,omitempty" binding:"isdefault"`
	ScaleIO              *v1.ScaleIOVolumeSource               `json:"scaleIO,omitempty" binding:"isdefault"`
	StorageOS            *v1.StorageOSVolumeSource             `json:"storageos,omitempty" binding:"isdefault"`
	CSI                  *v1.CSIVolumeSource                   `json:"csi,omitempty" binding:"isdefault"`
	Ephemeral            *v1.EphemeralVolumeSource             `json:"ephemeral,omitempty" binding:"isdefault"`
}

// HostPathVolumeSource [struct] to describe host path volume info
type HostPathVolumeSource struct {
	Path string  `json:"path" binding:"min=2,max=1024,excludes=.." validate:"^/[a-z0-9A-Z_./-]+$"`
	Type *string `json:"type,omitempty" binding:"eq="`
}

type affinity struct {
	NodeAffinity *nodeAffinity `json:"nodeAffinity,omitempty" binding:"omitempty"`
}

type nodeAffinity struct {
	A *nodeSelector             `json:"requiredDuringSchedulingIgnoredDuringExecution,omitempty" binding:"omitempty"`
	B []preferredSchedulingTerm `json:"preferredDuringSchedulingIgnoredDuringExecution,omitempty" binding:"len=0,dive"`
}

type nodeSelector struct {
	NodeSelectorTerms []nodeSelectorTerm `json:"nodeSelectorTerms" binding:"max=1,dive"`
}

type nodeSelectorTerm struct {
	MatchExpressions []nodeSelectorRequirement `json:"matchExpressions,omitempty" binding:"len=0,dive"`
	MatchFields      []nodeSelectorRequirement `json:"matchFields,omitempty" binding:"max=1,dive"`
}

type nodeSelectorRequirement struct {
	Key      string   `json:"key" binding:"eq=metadata.name"`
	Operator string   `json:"operator" binding:"oneof=In NotIn Exists DoesNotExist Gt Lt"`
	Values   []string `json:"values,omitempty" binding:"max=1,dive,required,max=512"`
}

type preferredSchedulingTerm struct {
	Weight     int32            `json:"weight" binding:"isdefault"`
	Preference nodeSelectorTerm `json:"preference" binding:"isdefault"`
}

type podSecurityContext struct {
	SELinuxOptions *v1.SELinuxOptions                `json:"seLinuxOptions,omitempty" binding:"isdefault,omitempty"`
	WindowsOptions *v1.WindowsSecurityContextOptions `json:"windowsOptions,omitempty" binding:"isdefault,omitempty"`

	RunAsUser           *int64                     `json:"runAsUser,omitempty" binding:"isdefault,omitempty"`
	RunAsGroup          *int64                     `json:"runAsGroup,omitempty" binding:"isdefault,omitempty"`
	RunAsNonRoot        *bool                      `json:"runAsNonRoot,omitempty" binding:"isdefault,omitempty"`
	SupplementalGroups  []int64                    `json:"supplementalGroups,omitempty" binding:"isdefault,omitempty"`
	FSGroup             *int64                     `json:"fsGroup,omitempty" binding:"isdefault,omitempty"`
	Sysctls             []v1.Sysctl                `json:"sysctls,omitempty" binding:"isdefault,omitempty"`
	FSGroupChangePolicy *v1.PodFSGroupChangePolicy `json:"fsGroupChangePolicy,omitempty" binding:"isdefault,omitempty"`
	SeccompProfile      *v1.SeccompProfile         `json:"seccompProfile,omitempty" binding:"isdefault,omitempty"`
}

type localObjectReference struct {
	Name string `json:"name,omitempty" binding:"omitempty,oneof=fusion-director-docker-registry-secret image-pull-secret"`
}

// HTTPGetAction [struct] to describe HTTPGetActione info
type HTTPGetAction struct {
	Path        string             `json:"path,omitempty"`
	Port        intstr.IntOrString `json:"port"`
	Host        string             `json:"host,omitempty" binding:"max=64"`
	Scheme      string             `json:"scheme,omitempty" binding:"omitempty,oneof=HTTP HTTPS"`
	HTTPHeaders []v1.HTTPHeader    `json:"httpHeaders,omitempty" binding:"isdefault,omitempty"`
}

// ExecAction [struct] to describe ExecAction info
type ExecAction struct {
	Command []string `json:"command,omitempty" binding:"max=1,dive,max=1024,excludes=.." validate:"^/[a-z0-9A-Z_./-]+$"`
}

// ProbeHandler [struct] to describe ProbeHandler info
type ProbeHandler struct {
	Exec      *ExecAction         `json:"exec,omitempty" binding:"omitempty"`
	HTTPGet   *HTTPGetAction      `json:"httpGet,omitempty" binding:"omitempty"`
	TCPSocket *v1.TCPSocketAction `json:"tcpSocket,omitempty" binding:"isdefault,omitempty"`
	GRPC      *v1.GRPCAction      `json:"grpc,omitempty" binding:"isdefault,omitempty"`
}

// Probe [struct] to  describe Probe info
type Probe struct {
	ProbeHandler                  `json:",inline"`
	InitialDelaySeconds           int32  `json:"initialDelaySeconds,omitempty" binding:"omitempty,min=1,max=3600"`
	TimeoutSeconds                int32  `json:"timeoutSeconds,omitempty" binding:"omitempty,min=1,max=3600"`
	PeriodSeconds                 int32  `json:"periodSeconds,omitempty" binding:"omitempty,min=1,max=3600"`
	SuccessThreshold              int32  `json:"successThreshold,omitempty" binding:"omitempty,eq=1"`
	FailureThreshold              int32  `json:"failureThreshold,omitempty" binding:"omitempty,eq=3"`
	TerminationGracePeriodSeconds *int64 `json:"terminationGracePeriodSeconds,omitempty" binding:"isdefault,omitempty"`
}

type toleration struct {
	Key               string `json:"key,omitempty" binding:"startswith=node.kubernetes.io,max=128"`
	Operator          string `json:"operator,omitempty" binding:"oneof=Exists Equal"`
	Value             string `json:"value,omitempty" binding:"max=128"`
	Effect            string `json:"effect,omitempty" binding:"oneof=NoSchedule PreferNoSchedule NoExecute"`
	TolerationSeconds *int64 `json:"tolerationSeconds,omitempty" binding:"omitempty,max=120"`
}

// Container [struct] to describe container info
type Container struct {
	Name  string   `json:"name" validate:"^[a-z0-9]([a-z0-9-]{0,30}[a-z0-9]){0,1}$"`
	Image string   `json:"image,omitempty" validate:"^[a-zA-Z0-9:_/.-]{0,288}$"`
	Cmd   []string `json:"command,omitempty" binding:"max=16" validate:"^([a-zA-Z0-9 _./-]{0,255}[a-zA-Z0-9]){0,1}$"`
	Args  []string `json:"args,omitempty" binding:"max=16" validate:"^([a-zA-Z0-9 =_./-]{0,255}[a-zA-Z0-9]){0,1}$"`

	WorkingDir     string               `json:"workingDir,omitempty" binding:"isdefault,omitempty"`
	Ports          []ContainerPort      `json:"ports,omitempty" binding:"max=16,dive"`
	EnvFrom        []v1.EnvFromSource   `json:"envFrom,omitempty" binding:"omitempty,len=0"`
	Env            []envVar             `json:"env,omitempty" binding:"max=256,dive"`
	Resources      resourceRequirements `json:"resources,omitempty" binding:"required"`
	VolumeMounts   []volumeMount        `json:"volumeMounts,omitempty" binding:"max=256,dive"`
	VolumeDevices  []v1.VolumeDevice    `json:"volumeDevices,omitempty" binding:"omitempty,len=0"`
	LivenessProbe  *Probe               `json:"livenessProbe,omitempty" binding:"omitempty"`
	ReadinessProbe *Probe               `json:"readinessProbe,omitempty" binding:"omitempty"`
	StartupProbe   *Probe               `json:"startupProbe,omitempty" binding:"isdefault,omitempty"`
	Lifecycle      *v1.Lifecycle        `json:"lifecycle,omitempty" binding:"isdefault,omitempty"`

	TermMsgPath   string `json:"terminationMessagePath,omitempty" binding:"omitempty,eq=/dev/termination-log"`
	TermMsgPolicy string `json:"terminationMessagePolicy,omitempty" binding:"omitempty,eq=File"`

	PullPolicy string           `json:"imagePullPolicy,omitempty" binding:"omitempty,oneof=IfNotPresent Always"`
	SecContext *SecurityContext `json:"securityContext,omitempty" binding:"omitempty"`
	Stdin      bool             `json:"stdin,omitempty" binding:"isdefault,omitempty"`
	StdinOnce  bool             `json:"stdinOnce,omitempty" binding:"isdefault,omitempty"`
	TTY        bool             `json:"tty,omitempty" binding:"isdefault,omitempty"`
}

// ContainerPort [struct] to describe container port info
type ContainerPort struct {
	Name          string `json:"name,omitempty" validate:"^([a-z0-9]([a-z0-9-]{0,30}[a-z0-9]){0,1}){0,1}$"`
	HostPort      int32  `json:"hostPort,omitempty" binding:"min=1024,max=65535"`
	ContainerPort int32  `json:"containerPort" binding:"min=1,max=65535"`
	Protocol      string `json:"protocol,omitempty" binding:"oneof=TCP UDP"`
	HostIP        string `json:"hostIP,omitempty" binding:"ipv4"`
}

type envVar struct {
	Name  string `json:"name" validate:"^[a-zA-Z][a-zA-Z0-9._-]{0,30}[a-zA-Z0-9]$"`
	Value string `json:"value,omitempty" validate:"^[a-zA-Z0-9 _./:-]{1,512}$"`

	ValueFrom *v1.EnvVarSource `json:"valueFrom,omitempty" binding:"isdefault,omitempty"`
}

type resourceRequirements struct {
	Lim v1.ResourceList `json:"limits,omitempty" binding:"min=0,max=3,dive,keys,oneof=cpu memory huawei.com/Ascend310"`
	Req v1.ResourceList `json:"requests,omitempty" binding:"min=0,max=3,dive,keys,oneof=cpu memory huawei.com/Ascend310"`
}

type volumeMount struct {
	Name      string `json:"name" validate:"^[a-zA-Z0-9-]{1,63}$"`
	ReadOnly  bool   `json:"readOnly,omitempty" binding:"eq=true"`
	MountPath string `json:"mountPath" binding:"min=2,max=1024,excludes=.." validate:"^/[a-z0-9A-Z_./-]+$"`

	SubPath          string                   `json:"subPath,omitempty" binding:"isdefault,omitempty"`
	MountPropagation *v1.MountPropagationMode `json:"mountPropagation,omitempty"  binding:"isdefault,omitempty"`
	SubPathExpr      string                   `json:"subPathExpr,omitempty" binding:"isdefault,omitempty"`
}

// SecurityContext [struct] to describe securityContext info
type SecurityContext struct {
	Capabilities             *Capabilities `json:"capabilities,omitempty" binding:"omitempty"`
	Privileged               *bool         `json:"privileged,omitempty" binding:"omitempty"`
	RunAsUser                *int64        `json:"runAsUser,omitempty" binding:"omitempty,min=0,max=65535"`
	RunAsGroup               *int64        `json:"runAsGroup,omitempty" binding:"omitempty,min=0,max=65535"`
	RunAsNonRoot             *bool         `json:"runAsNonRoot,omitempty" binding:"omitempty"`
	ReadOnlyRootFilesystem   *bool         `json:"readOnlyRootFilesystem,omitempty" binding:"omitempty"`
	AllowPrivilegeEscalation *bool         `json:"allowPrivilegeEscalation,omitempty" binding:"omitempty"`

	SELinuxOptions *v1.SELinuxOptions                `json:"seLinuxOptions,omitempty" binding:"isdefault"`
	WindowsOptions *v1.WindowsSecurityContextOptions `json:"windowsOptions,omitempty" binding:"isdefault"`
	ProcMount      *v1.ProcMountType                 `json:"procMount,omitempty" binding:"isdefault"`
	SeccompProfile *SeccompProfile                   `json:"seccompProfile" binding:"omitempty"`
}

// SeccompProfile [struct] to describe seccompProfile info
type SeccompProfile struct {
	Type             string  `json:"type" binding:"eq=RuntimeDefault"`
	LocalhostProfile *string `json:"localhostProfile" binding:"isdefault,omitempty"`
}

// Capabilities [struct] to capabilities seccompProfile info
type Capabilities struct {
	Add  []string `json:"add,omitempty" binding:"max=5,dive,max=32"`
	Drop []string `json:"drop,omitempty" binding:"max=5,dive,max=32"`
}

type podStatus struct {
	Phase string `json:"phase,omitempty" binding:"omitempty,oneof=Pending Running Succeeded Failed Unknown"`

	Conditions []podCondition    `json:"conditions,omitempty" binding:"max=4,dive"`
	Message    string            `json:"message,omitempty" binding:"max=2048"`
	Reason     string            `json:"reason,omitempty" binding:"max=2048"`
	NodeName   string            `json:"nominatedNodeName,omitempty" binding:"max=1024"`
	HostIP     string            `json:"hostIP,omitempty" binding:"isdefault|ipv4"`
	PodIP      string            `json:"podIP,omitempty" binding:"isdefault|ipv4"`
	PodIPs     []podIP           `json:"podIPs,omitempty" binding:"max=2,dive"`
	StartTime  *string           `json:"startTime,omitempty" binding:"omitempty,max=64"`
	CStatus    []containerStatus `json:"containerStatuses,omitempty" binding:"max=20,dive"`
	QOSClass   string            `json:"qosClass,omitempty" binding:"isdefault|oneof=Guaranteed Burstable BestEffort"`
	ICStatuses []containerStatus `json:"initContainerStatuses,omitempty" binding:"omitempty,len=0"`
	ECStatuses []containerStatus `json:"ephemeralContainerStatuses,omitempty" binding:"omitempty,len=0"`
}

type podCondition struct {
	Type               string `json:"type" binding:"oneof=ContainersReady Initialized Ready PodScheduled"`
	Status             string `json:"status" binding:"oneof=True False Unknown"`
	LastProbeTime      string `json:"lastProbeTime,omitempty" binding:"omitempty,eq=null"`
	LastTransitionTime string `json:"lastTransitionTime,omitempty" binding:"max=64"`
	Reason             string `json:"reason,omitempty" binding:"max=2048"`
	Message            string `json:"message,omitempty" binding:"max=2048"`
}

type podIP struct {
	IP string `json:"ip,omitempty" binding:"isdefault|ipv4"`
}

type containerStatus struct {
	Name                 string         `json:"name" binding:"min=1,max=32"`
	State                containerState `json:"state,omitempty" binding:"omitempty"`
	LastTerminationState containerState `json:"lastState,omitempty" binding:"omitempty"`
	Ready                bool           `json:"ready" binding:"-"`
	RestartCount         int32          `json:"restartCount"`
	Image                string         `json:"image" binding:"min=3,max=299"`
	ImageID              string         `json:"imageID" binding:"max=1024"`
	ContainerID          string         `json:"containerID,omitempty" binding:"max=1024"`
	Started              *bool          `json:"started,omitempty" binding:"required"`
}

type containerState struct {
	Waiting    *containerStateWaiting    `json:"waiting,omitempty" binding:"omitempty"`
	Running    *containerStateRunning    `json:"running,omitempty" binding:"omitempty"`
	Terminated *containerStateTerminated `json:"terminated,omitempty" binding:"omitempty"`
}

type containerStateWaiting struct {
	Reason  string `json:"reason,omitempty" binding:"max=2048"`
	Message string `json:"message,omitempty" binding:"max=2048"`
}

type containerStateRunning struct {
	StartedAt string `json:"startedAt,omitempty" binding:"max=64"`
}

type containerStateTerminated struct {
	ExitCode    int32  `json:"exitCode"`
	Signal      int32  `json:"signal,omitempty"`
	Reason      string `json:"reason,omitempty" binding:"max=2048"`
	Message     string `json:"message,omitempty" binding:"max=2048"`
	StartedAt   string `json:"startedAt,omitempty" binding:"max=64"`
	FinishedAt  string `json:"finishedAt,omitempty" binding:"max=64"`
	ContainerID string `json:"containerID,omitempty" binding:"max=1024"`
}
