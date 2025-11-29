// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.
//go:build MEFEdge_SDK

// Package config for
package config

import (
	"edge-installer/pkg/common/constants"
)

// GetDomainCfg get the mapping relation of domain and ip
func GetDomainCfg() (*DomainConfigs, error) {
	var cfg DomainConfigs
	dbMgr, err := GetComponentDbMgr(constants.EdgeOm)
	if err != nil {
		return nil, err
	}
	if err = dbMgr.GetConfig(constants.DomainCfgKey, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// SetDomainCfg set the mapping relation of domain and ip for usage of image registry
func SetDomainCfg(cfg *DomainConfigs) error {
	dbMgr, err := GetComponentDbMgr(constants.EdgeOm)
	if err != nil {
		return err
	}
	return dbMgr.SetConfig(constants.DomainCfgKey, cfg)
}

// SetImageCfg set the mapping relation of domain and port for usage of image registry
func SetImageCfg(imageCfg *ImageConfig) error {
	dbMgr, err := GetComponentDbMgr(constants.EdgeOm)
	if err != nil {
		return err
	}
	return dbMgr.SetConfig(constants.ImageCfgKey, imageCfg)
}

// GetImageCfg get the mapping relation of domain and port
func GetImageCfg() (*ImageConfig, error) {
	var imageCfg ImageConfig
	dbMgr, err := GetComponentDbMgr(constants.EdgeOm)
	if err != nil {
		return nil, err
	}
	if err = dbMgr.GetConfig(constants.ImageCfgKey, &imageCfg); err != nil {
		return nil, err
	}
	return &imageCfg, nil
}
