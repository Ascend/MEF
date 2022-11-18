// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package util to init util service
package util

// ListReq for common list request, PageNum and PageSize for slice page, Name for fuzzy query
type ListReq struct {
	PageNum  uint64
	PageSize uint64
	Name     string
}
