// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package checker

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"reflect"
	"sort"

	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindx/common/checker/valuer"
)

// GetUniqueListChecker [method] for get unique list checker, reflect.DeepEqual will be used for comparing
func GetUniqueListChecker(field string, elemChecker checkerIntf, minLen, maxLen int, required bool) *UniqueListChecker {
	return &UniqueListChecker{
		ListChecker: *GetListChecker(field, elemChecker, minLen, maxLen, required),
	}
}

// UniqueListChecker [struct] for unique list checker
type UniqueListChecker struct {
	ListChecker
}

// Check [method] for do unique list checker
func (lc *UniqueListChecker) Check(data interface{}) CheckResult {
	if checkResult := lc.ListChecker.Check(data); !checkResult.Result {
		return checkResult
	}

	value, err := lc.ListChecker.valuer.GetValue(data, lc.ListChecker.field)
	if err != nil {
		if valuer.CheckIsFieldNotExistErr(err) && !lc.ListChecker.required {
			return NewSuccessResult()
		}
		return NewFailedResult(fmt.Sprintf("unique list checker get field [%s] value failed:%v", lc.field, err))
	}

	conflicts, err := lc.checkUniquenessBySort(*value)
	if err == nil {
		if len(conflicts) <= 1 {
			return NewSuccessResult()
		} else {
			return NewFailedResult(fmt.Sprintf("unique list checker unique check failed, [%d]==[%d]",
				conflicts[0], conflicts[1]))
		}
	}
	if conflicts = lc.checkUniquenessByCompare(*value); len(conflicts) <= 1 {
		return NewSuccessResult()
	} else {
		return NewFailedResult(fmt.Sprintf("unique list checker unique check failed, [%d]==[%d]",
			conflicts[0], conflicts[1]))
	}
}

func newSortableList(listValue reflect.Value) (*sortableList, error) {
	var sl sortableList
	for i := 0; i < listValue.Len(); i++ {
		elementValue := listValue.Index(i)
		if elementValue.Kind() == reflect.Ptr && elementValue.IsNil() {
			return nil, fmt.Errorf("can't encode nil pointer, %T", elementValue.Type().String())
		}
		var buffer bytes.Buffer
		encoder := gob.NewEncoder(&buffer)
		if err := encoder.EncodeValue(elementValue); err != nil {
			return nil, err
		}
		sl = append(sl, listElement{
			binaryData:    buffer.Bytes(),
			originalIndex: i,
		})
	}
	return &sl, nil
}

func (lc *UniqueListChecker) checkUniquenessBySort(listValue reflect.Value) ([]int, error) {
	sl, err := newSortableList(listValue)
	if err != nil {
		return nil, err
	}
	sort.Sort(sl)
	for i := 1; i < sl.Len(); i++ {
		if !sl.Less(i-1, i) {
			return []int{sl.getOriginalIndex(i), sl.getOriginalIndex(i - 1)}, nil
		}
	}
	return nil, nil
}

func (lc *UniqueListChecker) checkUniquenessByCompare(listValue reflect.Value) []int {
	for i := 0; i < listValue.Len(); i++ {
		for j := i + 1; j < listValue.Len(); j++ {
			elementI := listValue.Index(i).Interface()
			elementJ := listValue.Index(j).Interface()
			if reflect.DeepEqual(elementI, elementJ) {
				return []int{i, j}
			}
		}
	}
	return nil
}

type listElement struct {
	originalIndex int
	binaryData    []byte
}

type sortableList []listElement

func (s sortableList) Len() int {
	return len(s)
}

func (s sortableList) Less(i, j int) bool {
	if i >= len(s) || j >= len(s) {
		hwlog.RunLog.Error("UniqueListChecker: failed to compare element due to index out of bound")
		return false
	}
	return bytes.Compare(s[i].binaryData, s[j].binaryData) < 0
}

func (s sortableList) Swap(i, j int) {
	if i >= len(s) || j >= len(s) {
		hwlog.RunLog.Error("UniqueListChecker: failed to swap element due to index out of bound")
		return
	}
	s[i], s[j] = s[j], s[i]
}

func (s sortableList) getOriginalIndex(i int) int {
	if i >= len(s) {
		hwlog.RunLog.Error("UniqueListChecker: failed to get element due to index out of bound")
		return 0
	}
	return s[i].originalIndex
}
