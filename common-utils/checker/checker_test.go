// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package checker

import (
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

const (
	minListLen  = 0
	maxListLen  = 10
	minAge      = 18
	maxAge      = 28
	minLike     = 1
	maxLike     = 40
	maxClassNum = 8
	maxIncome   = 1000
)

type classStruct struct {
	topClass bool
	number   int64
	income   float64
	name     string
}
type Student struct {
	sex   *string
	class []classStruct
	like  []int64
	age   int64
	name  string
}

type StudentChecker struct {
	modelChecker ModelChecker
}

func GetStudentChecker(field string) *StudentChecker {
	return &StudentChecker{
		modelChecker: ModelChecker{Field: field},
	}
}

func (nc *StudentChecker) init() {
	nc.modelChecker.Checker = GetAndChecker(
		GetRegChecker("name", "^[a-z]{1,10}$", true),
		GetIntChecker("age", minAge, maxAge, true),
	)
}

func (nc *StudentChecker) Check(data interface{}) CheckResult {
	nc.init()
	checkResult := nc.modelChecker.Check(data)
	if !checkResult.Result {
		return NewFailedResult(fmt.Sprintf("student checker failed: %v", checkResult.Reason))
	}
	return NewSuccessResult()
}

func GetClassChecker(field string) *ClassChecker {
	return &ClassChecker{
		modelChecker: ModelChecker{Field: field},
	}
}

type ClassChecker struct {
	modelChecker ModelChecker
}

func (cc *ClassChecker) init() {
	cc.modelChecker.Checker = GetAndChecker(
		GetRegChecker("name", "^[a-z0-9]{1,10}$", true),
		GetIntChecker("number", 1, maxClassNum, true),
	)
}

func (cc *ClassChecker) Check(data interface{}) CheckResult {
	cc.init()
	checkResult := cc.modelChecker.Check(data)
	if !checkResult.Result {
		return NewFailedResult(fmt.Sprintf("class checker failed: %v", checkResult.Reason))
	}
	return NewSuccessResult()
}
func TestCheck(t *testing.T) {
	convey.Convey("test checker", t, func() {
		var sexValue string
		sexValue = "woman"
		studentData := Student{
			name:  "abcde",
			age:   18,
			sex:   &sexValue,
			class: []classStruct{{name: "high", number: 7, topClass: true, income: 100}},
			like:  []int64{10, 20, 30}}
		checkResult := GetStudentChecker("").Check(studentData)
		expectCheckResult := NewSuccessResult()
		convey.So(checkResult.String(), convey.ShouldEqual, expectCheckResult.String())
	})
}
