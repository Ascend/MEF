// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package types Lease
package types

// LeaseSpec is a specification of a Lease.
type LeaseSpec struct {
	HolderIdentity       *string `json:"holderIdentity,omitempty" binding:"omitempty,max=64"`
	LeaseDurationSeconds *int32  `json:"leaseDurationSeconds,omitempty"`
	AcquireTime          *string `json:"acquireTime,omitempty" binding:"omitempty,max=64"`
	RenewTime            *string `json:"renewTime,omitempty" binding:"omitempty,max=64"`
	LeaseTransitions     *int32  `json:"leaseTransitions,omitempty"`
}

// Lease defines a Lease concept.
type Lease struct {
	TypeMeta   `json:",inline"`
	ObjectMeta `json:"metadata,omitempty"  binding:"omitempty"`
	Spec       LeaseSpec `json:"spec,omitempty"  binding:"omitempty"`
}
