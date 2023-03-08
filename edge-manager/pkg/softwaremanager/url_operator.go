// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package softwaremanager for url operator
package softwaremanager

import (
	"errors"
	"sort"
	"strings"
)

const (
	opAdd    = "ADD"
	opSync   = "SYNC"
	opDelete = "DELETE"
)

type urlOperator struct {
	urlInfos []UrlInfo
	option   string
}

func newUrlOperator(urlInfos []UrlInfo, option string) urlOperator {
	return urlOperator{urlInfos: urlInfos, option: option}
}

func (uo *urlOperator) operate(urlInfos []UrlInfo) error {
	switch uo.option {
	case opAdd:
		uo.add(urlInfos)
	case opDelete:
		uo.delete(urlInfos)
	case opSync:
		uo.sync(urlInfos)
	default:
		return errors.New("option unknown")
	}
	return nil
}

func (uo *urlOperator) add(urlInfos []UrlInfo) {
	var tmpUrlInfo = uo.urlInfos
	for _, urlInfo := range urlInfos {
		tmpUrlInfo = append(tmpUrlInfo, urlInfo)
	}
	uo.urlInfos = tmpUrlInfo

	uo.unique()
	uo.sort()
	uo.limitCount()
}

func (uo *urlOperator) unique() {
	inResult := make(map[UrlInfo]struct{})
	var res []UrlInfo
	for _, urlInfo := range uo.urlInfos {
		if _, ok := inResult[urlInfo]; !ok {
			res = append(res, urlInfo)
			inResult[urlInfo] = struct{}{}
		}
	}
	uo.urlInfos = res
}

func (uo *urlOperator) delete(urlInfos []UrlInfo) {
	inResult := make(map[UrlInfo]struct{})

	for _, urlInfo := range uo.urlInfos {
		inResult[urlInfo] = struct{}{}
	}

	for _, urlInfo := range urlInfos {
		if _, ok := inResult[urlInfo]; ok {
			delete(inResult, urlInfo)
		}
	}

	var res []UrlInfo
	for urlInfo := range inResult {
		res = append(res, urlInfo)
	}

	uo.urlInfos = res
	uo.sort()
	uo.limitCount()
}

func (uo *urlOperator) sync(urlInfos []UrlInfo) {
	uo.urlInfos = urlInfos
	uo.unique()
	uo.sort()
	uo.limitCount()
}

func (uo urlOperator) Len() int { return len(uo.urlInfos) }
func (uo urlOperator) Swap(i, j int) {
	if i >= len(uo.urlInfos) || j >= len(uo.urlInfos) {
		return
	}
	uo.urlInfos[i], uo.urlInfos[j] = uo.urlInfos[j], uo.urlInfos[i]
}
func (uo urlOperator) Less(i, j int) bool {
	if i >= len(uo.urlInfos) || j >= len(uo.urlInfos) {
		return false
	}
	if strings.Compare(uo.urlInfos[i].Version, uo.urlInfos[j].Version) > 0 {
		return true
	}

	if strings.Compare(uo.urlInfos[i].Version, uo.urlInfos[j].Version) < 0 {
		return false
	}

	if strings.Compare(uo.urlInfos[i].CreatedAt, uo.urlInfos[j].CreatedAt) > 0 {
		return true
	} else {
		return false
	}

}

func (uo *urlOperator) sort() {
	sort.Sort(uo)
}

func (uo *urlOperator) limitCount() {
	if len(uo.urlInfos) > maxSftUrlCount {
		uo.urlInfos = uo.urlInfos[0:maxSftUrlCount]
	}
}
