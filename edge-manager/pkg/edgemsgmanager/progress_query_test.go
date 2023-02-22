package edgemsgmanager

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/modulemanager/model"
)

func testProgressQueryValid() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed")
	}

	dataCases := []string{
		"2102312NSF10K8000130",
	}

	for _, dataCase := range dataCases {
		msg.FillContent(dataCase)
		resp := queryEdgeDownloadProgress(msg)

		convey.So(resp.Status, convey.ShouldEqual, common.Success)
	}

}

func testProgressQueryInvalid() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed")
	}

	dataCases := []string{
		"_2102312NSF10K8000130",
		"-2102312NSF10K8000130",
		"2102312NSF10K8000130_",
		"2102312NSF10K8000130-",
		"21!02312NSF10K800013$0",
		"2102312NSF10K80001302102312NSF10K80001302102312NSF10K800013021023",
	}

	for _, dataCase := range dataCases {
		msg.FillContent(dataCase)
		resp := queryEdgeDownloadProgress(msg)

		convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
	}

}

func TestProgressQueryPara(t *testing.T) {
	convey.Convey("test download info", t, func() {
		convey.Convey("test progress query info", func() {
			convey.Convey("progress query should success", testProgressQueryValid)
			convey.Convey("test invalid progress query para", testProgressQueryInvalid)
		})
	})
}
