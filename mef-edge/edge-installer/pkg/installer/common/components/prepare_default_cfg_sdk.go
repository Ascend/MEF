// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.
//go:build !MEFEdge_A500

// Package components for prepare default config backup
package components

func (pi *PrepareInstaller) prepareDefaultCfgBackupDir() error {
	// !MEFEdge_A500 do not prepare default config backup dir
	return nil
}

func (peo *PrepareEdgeOm) prepareDefaultCfgBackupDir() error {
	// !MEFEdge_A500 do not prepare default config backup dir
	return nil
}

func (pem *PrepareEdgeMain) prepareDefaultCfgBackupDir() error {
	// !MEFEdge_A500 do not prepare default config backup dir
	return nil
}

func (pec *PrepareEdgeCore) prepareDefaultCfgBackupDir() error {
	// !MEFEdge_A500 do not prepare default config backup dir
	return nil
}
