// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package types fore node
package types

// TypeMeta describes an individual object in an API response or request
// with strings representing the type of the object and its API schema version.
// Structures that are versioned or persisted should inline TypeMeta.
type TypeMeta struct {
	Kind       string `json:"kind,omitempty" binding:"max=64"`
	APIVersion string `json:"apiVersion,omitempty" binding:"max=64"`
}

// OwnerReference contains enough information to let you identify an owning
// object. An owning object must be in the same namespace as the dependent, or
// be cluster-scoped, so there is no namespace field.
type OwnerReference struct {
	APIVersion         string `json:"apiVersion" binding:"max=64"`
	Kind               string `json:"kind" binding:"max=64"`
	Name               string `json:"name" binding:"max=64"`
	UID                string `json:"uid" binding:"max=64"`
	Controller         *bool  `json:"controller,omitempty"`
	BlockOwnerDeletion *bool  `json:"blockOwnerDeletion,omitempty"`
}

// FieldsV1 stores a set of fields in a data structure like a Trie, in JSON format.
type FieldsV1 struct {
	Raw []byte `json:"-"`
}

// ManagedFieldsEntry is a workflow-id, a FieldSet and the group version of the resource
// that the fieldset applies to.
type ManagedFieldsEntry struct {
	Manager    string `json:"manager,omitempty" binding:"max=64"`
	Operation  string `json:"operation,omitempty" binding:"eq=Apply|eq=Update"`
	APIVersion string `json:"apiVersion,omitempty" binding:"max=64"`
	Time       string `json:"time,omitempty" binding:"max=64"`

	FieldsType  string    `json:"fieldsType,omitempty" binding:"max=64"`
	FieldsV1    *FieldsV1 `json:"fieldsV1,omitempty" binding:"omitempty"`
	Subresource string    `json:"subresource,omitempty" binding:"max=64"`
}

// ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create
type ObjectMeta struct {
	Name            string `json:"name,omitempty" validate:"^[a-z0-9]([a-z0-9-]{0,126}[a-z0-9]){0,1}$"`
	GenerateName    string `json:"generateName,omitempty" binding:"max=64"`
	Namespace       string `json:"namespace,omitempty" binding:"omitempty,oneof=websocket mef-user"`
	SelfLink        string `json:"selfLink,omitempty"  validate:"^[a-z0-9A-Z/-]{0,128}$"`
	ResourceVersion string `json:"resourceVersion,omitempty" validate:"^[a-zA-Z0-9]{0,64}$"`
	Generation      int64  `json:"generation,omitempty"`

	CreationTimestamp string  `json:"creationTimestamp,omitempty" binding:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	DeletionTimestamp *string `json:"deletionTimestamp,omitempty" binding:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`

	DeletionGracePeriodSeconds *int64 `json:"deletionGracePeriodSeconds,omitempty" binding:"omitempty,min=0,max=120"`

	Labels      map[string]string `json:"labels,omitempty" binding:"max=64,dive,keys,max=128,endkeys,max=64"`
	Annotations map[string]string `json:"annotations,omitempty" binding:"max=64,dive,keys,max=128,endkeys,max=64"`

	OwnerReferences []OwnerReference     `json:"ownerReferences,omitempty" binding:"max=64,dive"`
	Finalizers      []string             `json:"finalizers,omitempty" binding:"max=64,dive,max=64"`
	ClusterName     string               `json:"clusterName,omitempty" binding:"max=64"`
	ManagedFields   []ManagedFieldsEntry `json:"managedFields,omitempty" binding:"max=64,dive"`

	// The fd secret uid is not correct uid format. Ensure that the uid is verified only after other fields are verified
	// by validate tag successfully
	UID string `json:"uid,omitempty" validate:"^([0-9a-f]{8}(-[0-9a-f]{4}){3}-[0-9a-f]{12}){0,1}$"`
}

// Taint The node this Taint is attached to has the "effect" on any pod that does not tolerate the Taint.
type Taint struct {
	Key       string  `json:"key" binding:"max=64"`
	Value     string  `json:"value,omitempty" binding:"max=64"`
	Effect    string  `json:"effect" binding:"max=64"`
	TimeAdded *string `json:"timeAdded,omitempty" binding:"omitempty,max=64"`
}

// ConfigMapNodeConfigSource represents the config map of a node
type ConfigMapNodeConfigSource struct {
	Namespace string `json:"namespace" binding:"max=64"`
	Name      string `json:"name" binding:"max=64"`
	UID       string `json:"uid,omitempty" binding:"max=64"`

	ResourceVersion  string `json:"resourceVersion,omitempty" binding:"max=64"`
	KubeletConfigKey string `json:"kubeletConfigKey" binding:"max=64"`
}

// NodeConfigSource specifies a source of node configuration. Exactly one subfield (excluding metadata) must be non-nil.
// This API is deprecated since 1.22
type NodeConfigSource struct {
	ConfigMap *ConfigMapNodeConfigSource `json:"configMap,omitempty" binding:"omitempty"`
}

// NodeSpec describes the attributes that a node is created with.
type NodeSpec struct {
	PodCIDR            string            `json:"podCIDR,omitempty" binding:"max=64"`
	PodCIDRs           []string          `json:"podCIDRs,omitempty" binding:"max=64,dive,max=64"`
	ProviderID         string            `json:"providerID,omitempty" binding:"max=64"`
	Unschedulable      bool              `json:"unschedulable,omitempty"`
	Taints             []Taint           `json:"taints,omitempty" binding:"max=64,dive"`
	ConfigSource       *NodeConfigSource `json:"configSource,omitempty" binding:"omitempty"`
	DoNotUseExternalID string            `json:"externalID,omitempty" binding:"max=64"`
}

// NodeCondition contains condition information for a node.
type NodeCondition struct {
	Type               string `json:"type" binding:"max=64"`
	Status             string `json:"status" binding:"eq=True|eq=False|eq=Unknown"`
	LastHeartbeatTime  string `json:"lastHeartbeatTime,omitempty" binding:"max=64"`
	LastTransitionTime string `json:"lastTransitionTime,omitempty" binding:"max=64"`
	Reason             string `json:"reason,omitempty" binding:"max=1024"`
	Message            string `json:"message,omitempty" binding:"max=1024"`
}

// NodeAddress contains information for the node's address.
type NodeAddress struct {
	Type    string `json:"type" binding:"max=64"`
	Address string `json:"address" binding:"max=64"`
}

// DaemonEndpoint contains information about a single Daemon endpoint.
type DaemonEndpoint struct {
	Port int32 `json:"Port" binding:"gte=0,max=65535"`
}

// NodeDaemonEndpoints lists ports opened by daemons running on the Node.
type NodeDaemonEndpoints struct {
	// Endpoint on which Kubelet is listening.
	// +optional
	KubeletEndpoint DaemonEndpoint `json:"kubeletEndpoint,omitempty"`
}

// NodeSystemInfo is a set of ids/uuids to uniquely identify the node.
type NodeSystemInfo struct {
	MachineID               string `json:"machineID" binding:"max=64"`
	SystemUUID              string `json:"systemUUID" binding:"max=64"`
	BootID                  string `json:"bootID" binding:"max=64"`
	KernelVersion           string `json:"kernelVersion" binding:"max=64"`
	OSImage                 string `json:"osImage" binding:"max=64"`
	ContainerRuntimeVersion string `json:"containerRuntimeVersion" binding:"max=64"`
	KubeletVersion          string `json:"kubeletVersion" binding:"max=64"`
	KubeProxyVersion        string `json:"kubeProxyVersion" binding:"max=64"`
	OperatingSystem         string `json:"operatingSystem" binding:"max=64"`
	Architecture            string `json:"architecture" binding:"max=64"`
}

// ContainerImage Describe a container image
type ContainerImage struct {
	Names     []string `json:"names" binding:"max=64,dive,max=256"`
	SizeBytes int64    `json:"sizeBytes,omitempty"`
}

// AttachedVolume describes a volume attached to a node
type AttachedVolume struct {
	Name string `json:"name" binding:"max=64"`

	DevicePath string `json:"devicePath" binding:"max=64"`
}

// NodeConfigStatus describes the status of the config assigned by Node.Spec.ConfigSource.
type NodeConfigStatus struct {
	Assigned      *NodeConfigSource `json:"assigned,omitempty" binding:"omitempty"`
	Active        *NodeConfigSource `json:"active,omitempty" binding:"omitempty"`
	LastKnownGood *NodeConfigSource `json:"lastKnownGood,omitempty" binding:"omitempty"`
	Error         string            `json:"error,omitempty" binding:"max=1024"`
}

// NodeStatus is information about the current status of a node.
type NodeStatus struct {
	Capacity        map[string]string   `json:"capacity,omitempty" binding:"max=64,dive,keys,max=64,endkeys,max=64"`
	Allocatable     map[string]string   `json:"allocatable,omitempty" binding:"max=64,dive,keys,max=64,endkeys,max=64"`
	Phase           string              `json:"phase,omitempty" binding:"max=64"`
	Conditions      []NodeCondition     `json:"conditions,omitempty"  binding:"max=64,dive"`
	Addresses       []NodeAddress       `json:"addresses,omitempty" binding:"max=64,dive"`
	DaemonEndpoints NodeDaemonEndpoints `json:"daemonEndpoints,omitempty"`
	NodeInfo        NodeSystemInfo      `json:"nodeInfo,omitempty"`
	Images          []ContainerImage    `json:"images,omitempty" binding:"max=128,dive"`
	VolumesInUse    []string            `json:"volumesInUse,omitempty" binding:"max=64,dive,max=256"`
	VolumesAttached []AttachedVolume    `json:"volumesAttached,omitempty" binding:"max=64,dive"`
	Config          *NodeConfigStatus   `json:"config,omitempty" binding:"omitempty"`
}

// Node [struct] to define Node info
type Node struct {
	TypeMeta
	ObjectMeta `json:"metadata,omitempty"`
	Spec       NodeSpec `json:"spec,omitempty"`

	Status NodeStatus
}

// NodePatch [struct] to define NodePatch info
type NodePatch struct {
	Object Node
}
