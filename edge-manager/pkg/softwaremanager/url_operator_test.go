// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package softwaremanager module test
package softwaremanager

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindxedge/base/common"
)

func testUrlUnique() {
	var oldUrlInfos = []UrlInfo{
		{
			Type:      common.EdgeCore,
			Url:       "xxxx",
			CreatedAt: "2023-02-01 10:56:59",
			Version:   "v1.12.1",
		},
		{
			Type:      common.EdgeCore,
			Url:       "xxxx",
			CreatedAt: "2023-02-01 10:56:59",
			Version:   "v1.12.1",
		},
	}
	urlOpr := newUrlOperator(oldUrlInfos, opAdd)
	urlOpr.unique()
	convey.So(len(urlOpr.urlInfos), convey.ShouldEqual, 1)
}

func testUrlSort() {
	const urlCount = 4
	var oldUrlInfos = []UrlInfo{
		{
			Type:      common.EdgeCore,
			Url:       "xxxx",
			CreatedAt: "2022-02-01 10:56:59",
			Version:   "v1.12.1",
		},
		{
			Type:      common.EdgeCore,
			Url:       "xxxx",
			CreatedAt: "2022-02-01 10:56:59",
			Version:   "v1.12.2",
		},
		{
			Type:      common.EdgeCore,
			Url:       "xxxx",
			CreatedAt: "2023-02-01 10:56:59",
			Version:   "v1.12.1",
		},
		{
			Type:      common.EdgeCore,
			Url:       "xxxx",
			CreatedAt: "2023-02-01 10:56:59",
			Version:   "v1.12.2",
		},
	}
	urlOpr := newUrlOperator(oldUrlInfos, opAdd)
	urlOpr.sort()

	convey.So(len(urlOpr.urlInfos), convey.ShouldEqual, urlCount)
	convey.So(urlOpr.urlInfos[0], convey.ShouldResemble, UrlInfo{
		Type:      common.EdgeCore,
		Url:       "xxxx",
		CreatedAt: "2023-02-01 10:56:59",
		Version:   "v1.12.2",
	})
	convey.So(urlOpr.urlInfos[1], convey.ShouldResemble, UrlInfo{
		Type:      common.EdgeCore,
		Url:       "xxxx",
		CreatedAt: "2022-02-01 10:56:59",
		Version:   "v1.12.2",
	})

	convey.So(urlOpr.urlInfos[3], convey.ShouldResemble, UrlInfo{
		Type:      common.EdgeCore,
		Url:       "xxxx",
		CreatedAt: "2022-02-01 10:56:59",
		Version:   "v1.12.1",
	})
}

func testUrlAddOperate() {
	const urlCount = 2
	var oldUrlInfos = []UrlInfo{
		{
			Type:      common.EdgeCore,
			Url:       "xxxx",
			CreatedAt: "2023-02-01 10:56:59",
			Version:   "v1.12.1",
		},
	}
	urlOpr := newUrlOperator(oldUrlInfos, opAdd)

	var urlInfos = []UrlInfo{
		{
			Type:      common.EdgeCore,
			Url:       "xxxx",
			CreatedAt: "2023-03-01 10:56:59",
			Version:   "v1.12.2",
		},
	}
	err := urlOpr.operate(urlInfos)
	convey.So(err, convey.ShouldEqual, nil)
	convey.So(len(urlOpr.urlInfos), convey.ShouldEqual, urlCount)
}

func testUrlAddOperateByDuplicate() {
	var oldUrlInfos = []UrlInfo{
		{
			Type:      common.EdgeCore,
			Url:       "xxxx",
			CreatedAt: "2023-02-01 10:56:59",
			Version:   "v1.12.1",
		},
	}
	urlOpr := newUrlOperator(oldUrlInfos, opAdd)

	var urlInfos = []UrlInfo{
		{
			Type:      common.EdgeCore,
			Url:       "xxxx",
			CreatedAt: "2023-02-01 10:56:59",
			Version:   "v1.12.1",
		},
	}
	err := urlOpr.operate(urlInfos)
	convey.So(err, convey.ShouldEqual, nil)
	convey.So(len(urlOpr.urlInfos), convey.ShouldEqual, 1)
}

func testUrlDeleteOperate() {
	const urlCount = 2
	var oldUrlInfos = []UrlInfo{
		{
			Type:      common.EdgeCore,
			Url:       "xxxx",
			CreatedAt: "2021-01-01 10:56:59",
			Version:   "v1.12.1",
		},
		{ // will be deleted
			Type:      common.EdgeCore,
			Url:       "xxxx",
			CreatedAt: "2022-02-01 10:56:59",
			Version:   "v1.12.1",
		},
		{
			Type:      common.EdgeCore,
			Url:       "xxxx",
			CreatedAt: "2023-03-01 10:56:59",
			Version:   "v1.12.2",
		},
	}
	urlOpr := newUrlOperator(oldUrlInfos, opDelete)

	var urlInfos = []UrlInfo{
		{
			Type:      common.EdgeCore,
			Url:       "xxxx",
			CreatedAt: "2022-02-01 10:56:59",
			Version:   "v1.12.1",
		},
		{ // no impact
			Type:      common.EdgeCore,
			Url:       "xxxx",
			CreatedAt: "2024-02-01 10:56:59",
			Version:   "v1.12.1",
		},
	}
	err := urlOpr.operate(urlInfos)
	convey.So(err, convey.ShouldEqual, nil)
	convey.So(len(urlOpr.urlInfos), convey.ShouldEqual, urlCount)
	convey.So(urlOpr.urlInfos[0], convey.ShouldResemble, UrlInfo{
		Type:      common.EdgeCore,
		Url:       "xxxx",
		CreatedAt: "2023-03-01 10:56:59",
		Version:   "v1.12.2",
	})
}

func testUrlSyncOperate() {
	var oldUrlInfos = []UrlInfo{
		{
			Type:      common.EdgeCore,
			Url:       "xxxx",
			CreatedAt: "2021-01-01 10:56:59",
			Version:   "v1.12.1",
		},
		{
			Type:      common.EdgeCore,
			Url:       "xxxx",
			CreatedAt: "2022-02-01 10:56:59",
			Version:   "v1.12.1",
		},
		{
			Type:      common.EdgeCore,
			Url:       "xxxx",
			CreatedAt: "2023-03-01 10:56:59",
			Version:   "v1.12.2",
		},
	}
	urlOpr := newUrlOperator(oldUrlInfos, opSync)

	var urlInfos = []UrlInfo{
		{ // cover old data
			Type:      common.EdgeCore,
			Url:       "xxxx",
			CreatedAt: "2022-02-01 10:56:59",
			Version:   "v1.12.1",
		},
	}
	err := urlOpr.operate(urlInfos)
	convey.So(err, convey.ShouldEqual, nil)
	convey.So(len(urlOpr.urlInfos), convey.ShouldEqual, 1)
	convey.So(urlOpr.urlInfos[0], convey.ShouldResemble, UrlInfo{
		Type:      common.EdgeCore,
		Url:       "xxxx",
		CreatedAt: "2022-02-01 10:56:59",
		Version:   "v1.12.1",
	})
}

func TestUrlOperator(t *testing.T) {
	convey.Convey("test software url operate", t, func() {
		convey.Convey("test software url operate", func() {
			convey.Convey("test software url unique", testUrlUnique)
			convey.Convey("test software url sort", testUrlSort)
			convey.Convey("test software url add operate", testUrlAddOperate)
			convey.Convey("test software url add operate by remove duplicate", testUrlAddOperateByDuplicate)
			convey.Convey("test software url delete operate", testUrlDeleteOperate)
			convey.Convey("test software url sync operate", testUrlSyncOperate)
		})
	})
}
