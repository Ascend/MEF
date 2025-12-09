// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package util

import (
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestObjectWrapper(t *testing.T) {
	const (
		expect = 2
	)
	convey.Convey("Given a valid object wrapper", t, func() {
		obj := map[string]interface{}{
			"key1": "value1",
			"key2": 123,
			"key3": []interface{}{"value2", 456},
		}
		wrapper := NewWrapper(obj)

		convey.Convey("When calling GetString with an existing key", func() {
			value := wrapper.GetString("key1")

			convey.Convey("Then the returned value should be correct", func() {
				convey.So(value, convey.ShouldEqual, "value1")
			})
		})

		convey.Convey("When calling GetSlice with an existing key", func() {
			slice := wrapper.GetSlice("key3")

			convey.Convey("Then the returned slice should contain the correct objects", func() {
				convey.So(len(slice), convey.ShouldEqual, expect)
				convey.So(slice[1].GetData(), convey.ShouldEqual, 456)
			})
		})

	})
}
func TestNilObjectWrapper(t *testing.T) {
	const (
		expect = 1
	)

	convey.Convey("Given a valid object wrapper", t, func() {
		obj := map[string]interface{}{}
		wrapper := NewWrapper(obj)
		convey.Convey("When calling GetString with a non-existing key", func() {
			value := wrapper.GetString("key")
			convey.Convey("Then the returned value should be empty", func() {
				convey.So(value, convey.ShouldBeEmpty)
			})
		})
		convey.Convey("When calling GetSlice with a non-existing key", func() {
			slice := wrapper.GetSlice("key")
			convey.Convey("Then the returned slice should be empty", func() {
				convey.So(len(slice), convey.ShouldEqual, expect)
				convey.So(slice[0].GetData(), convey.ShouldBeNil)
			})
		})
	})
}

func TestInvalidObjectWrapper(t *testing.T) {
	convey.Convey("Given an invalid object wrapper", t, func() {
		wrapper := NewWrapper(nil)

		convey.Convey("When calling any method", func() {
			subWrapper := wrapper.GetObject("key")
			value := wrapper.GetString("key")
			slice := wrapper.GetSlice("key")

			convey.Convey("Then the returned value should be empty", func() {
				convey.So(subWrapper.GetData(), convey.ShouldBeNil)
				convey.So(value, convey.ShouldBeEmpty)
				convey.So(len(slice), convey.ShouldEqual, 1)
				convey.So(slice[0].GetData(), convey.ShouldBeNil)
			})
		})
	})
}

const notExistKey = "key"

var object = map[string]interface{}{
	"key1": "value1",
	"key2": 123,
	"key3": true,
	"key4": []interface{}{"value4", 123},
	"key5": []string{"value5"},
}

func TestGetObjectAndString(t *testing.T) {
	convey.Convey("test ObjectWrapper methods success", t, testObjectWrapper)
	convey.Convey("test ObjectWrapper methods, object is nil", t, testWrapperNilObj)
	convey.Convey("test ObjectWrapper methods, object type error", t, testWrapperErrObjType)
	convey.Convey("test ObjectWrapper methods, error wrapper key", t, testWrapperErrKey)
	convey.Convey("test ObjectWrapper method GetString/GetBool/GetSlice, value type error", t, testWrapperErrValType)
}

func testObjectWrapper() {
	wrapper := NewWrapper(object)
	// test 'GetObject'
	subWrapper := wrapper.GetObject("key1")
	convey.So(subWrapper.GetData(), convey.ShouldEqual, "value1")
	// test 'GetString'
	strVal := wrapper.GetString("key1")
	convey.So(strVal, convey.ShouldResemble, "value1")
	// test 'GetBool'
	boolVal, err := wrapper.GetBool("key3")
	convey.So(boolVal, convey.ShouldBeTrue)
	convey.So(err, convey.ShouldBeNil)
	// test 'GetSlice'
	slice := wrapper.GetSlice("key4")
	convey.So(slice[1].GetData(), convey.ShouldEqual, 123)
}

func testWrapperNilObj() {
	wrapper := NewWrapper(nil)
	// test 'GetObject'
	subWrapper := wrapper.GetObject(notExistKey)
	convey.So(subWrapper.GetData(), convey.ShouldBeNil)
	// test 'GetString'
	strValue := wrapper.GetString(notExistKey)
	convey.So(strValue, convey.ShouldResemble, "")
	// test 'GetBool'
	_, err := wrapper.GetBool(notExistKey)
	convey.So(err, convey.ShouldResemble, fmt.Errorf("val for %s nil", notExistKey))
	// test 'GetSlice'
	slice := wrapper.GetSlice("key4")
	convey.So(slice[0].GetData(), convey.ShouldBeNil)
}

func testWrapperErrObjType() {
	wrapper := NewWrapper("error type")
	// test 'GetObject'
	subWrapper := wrapper.GetObject(notExistKey)
	convey.So(subWrapper.GetData(), convey.ShouldBeNil)
	// test 'GetString'
	strVal := wrapper.GetString(notExistKey)
	convey.So(strVal, convey.ShouldResemble, "")
	// test 'GetBool'
	_, err := wrapper.GetBool(notExistKey)
	convey.So(err, convey.ShouldResemble, fmt.Errorf("not json map, cannot get key: %s", notExistKey))
	// test 'GetSlice'
	slice := wrapper.GetSlice(notExistKey)
	convey.So(slice[0].GetData(), convey.ShouldBeNil)
}

func testWrapperErrKey() {
	wrapper := NewWrapper(object)
	// test 'GetObject'
	subWrapper := wrapper.GetObject(notExistKey)
	convey.So(subWrapper.GetData(), convey.ShouldBeNil)
	// test 'GetString'
	strVal := wrapper.GetString(notExistKey)
	convey.So(strVal, convey.ShouldResemble, "")
	// test 'GetBool'
	_, err := wrapper.GetBool(notExistKey)
	convey.So(err, convey.ShouldResemble, fmt.Errorf("key %s not exist", notExistKey))
	// test 'GetSlice'
	slice := wrapper.GetSlice(notExistKey)
	convey.So(slice[0].GetData(), convey.ShouldBeNil)
}

func testWrapperErrValType() {
	wrapper := NewWrapper(object)
	// test 'GetString'
	strVal := wrapper.GetString("key2")
	convey.So(strVal, convey.ShouldResemble, "")
	// test 'GetBool'
	_, err := wrapper.GetBool("key1")
	convey.So(err, convey.ShouldResemble, fmt.Errorf("key %s is not bool", "key1"))
	// test 'GetSlice'
	slice := wrapper.GetSlice("key5")
	convey.So(slice[0].GetData(), convey.ShouldBeNil)
}
