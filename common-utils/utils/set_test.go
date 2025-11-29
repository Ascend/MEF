// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package utils test for string set operations
package utils

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestSet(t *testing.T) {
	convey.Convey("test set add", t, func() {
		s1 := NewSet()
		s1.Add("a1", "a2", "a3")
		s2 := NewSet("a1", "a2", "a3")
		convey.So(s1, convey.ShouldResemble, s2)
	})

	convey.Convey("test set find", t, func() {
		s1 := NewSet()
		s1.Add("a1", "a2", "a3")
		convey.So(s1.Find("a2"), convey.ShouldEqual, true)
		convey.So(s1.Find("b2"), convey.ShouldEqual, false)
	})

	convey.Convey("test set delete", t, func() {
		s1 := NewSet("a1", "a2", "a3")
		s1.Delete("a1")
		s2 := NewSet("a2", "a3")
		convey.So(s1, convey.ShouldResemble, s2)
	})

	convey.Convey("test set intersection, union, difference", t, func() {
		s1 := NewSet("a1", "a2")
		s2 := NewSet("a2", "a3")
		convey.So(s1.Intersection(s2), convey.ShouldResemble, NewSet("a2"))
		convey.So(s1.Union(s2), convey.ShouldResemble, NewSet("a1", "a2", "a3"))
		convey.So(s1.Difference(s2), convey.ShouldResemble, NewSet("a1"))
		convey.So(s2.Difference(s1), convey.ShouldResemble, NewSet("a3"))
	})

	convey.Convey("test set intersection, union, difference when set is nil", t, func() {
		s1 := NewSet("a1", "a2")
		convey.So(s1.Intersection(nil), convey.ShouldResemble, NewSet())
		convey.So(s1.Union(nil), convey.ShouldResemble, s1)
		convey.So(s1.Difference(nil), convey.ShouldResemble, s1)
	})
}
