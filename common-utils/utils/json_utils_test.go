// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package utils
package utils

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestObjectConvert(t *testing.T) {
	convey.Convey("test Object", t, func() {
		var (
			inputNumber           = 1
			actualOutputNumbers   = []interface{}{int(0), int32(0), int64(0), float32(0), float64(0)}
			expectedOutputNumbers = []interface{}{int(1), int32(1), int64(1), float32(1), float64(1)}
		)
		for i := range actualOutputNumbers {
			convey.So(ObjectConvert(inputNumber, &actualOutputNumbers[i]), convey.ShouldBeNil)
			convey.So(actualOutputNumbers[i], convey.ShouldEqual, expectedOutputNumbers[i])
		}

		type TestObj struct {
			TestField string `json:"testField"`
		}
		var (
			inputObj    = map[string]interface{}{"testField": "1"}
			actualObj   TestObj
			expectedObj = TestObj{TestField: "1"}
		)
		convey.So(ObjectConvert(inputObj, &actualObj), convey.ShouldBeNil)
		convey.So(actualObj, convey.ShouldResemble, expectedObj)

		actualObj = TestObj{}
		convey.So(ObjectConvert(expectedObj, &actualObj), convey.ShouldBeNil)
		convey.So(actualObj, convey.ShouldResemble, expectedObj)

		convey.So(ObjectConvert(expectedObj, (*TestObj)(nil)), convey.ShouldBeError)
		convey.So(ObjectConvert(expectedObj, TestObj{}), convey.ShouldBeError)
	})
}

func TestJsonConvert(t *testing.T) {
	convey.Convey("test JSONConvert", t, func() {
		var (
			inputNumber           = "1"
			actualOutputNumbers   = []interface{}{int(0), int32(0), int64(0), float32(0), float64(0)}
			expectedOutputNumbers = []interface{}{int(1), int32(1), int64(1), float32(1), float64(1)}
		)
		for i := range actualOutputNumbers {
			convey.So(JsonConvert(inputNumber, &actualOutputNumbers[i]), convey.ShouldBeNil)
			convey.So(actualOutputNumbers[i], convey.ShouldEqual, expectedOutputNumbers[i])
		}

		var (
			inputNull   = "null"
			actualObj   interface{}
			expectedObj interface{}
		)
		convey.So(JsonConvert(inputNull, &actualObj), convey.ShouldBeNil)
		convey.So(actualObj, convey.ShouldEqual, expectedObj)
	})
}
