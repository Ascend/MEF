package edgemsgmanager

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/hwlog"

	"edge-manager/pkg/types"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/modulemanager/model"
)

func testSoftwareQueryValid() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed")
	}

	var p2 = gomonkey.ApplyFunc(common.SendSyncMessageByRestful, func(input interface{},
		router *common.Router) common.RespMsg {
		var rsp common.RespMsg
		rsp.Status = common.Success
		var softwareInfo types.InnerSoftwareInfoResp
		softwareInfo.SoftwareInfo = append(softwareInfo.SoftwareInfo, types.SoftwareInfo{
			InactiveVersion: "v1.12",
			Name:            "edgecore",
			Version:         "v1.12"})

		rsp.Data = softwareInfo
		return rsp
	})
	defer p2.Reset()

	dataCases := []string{
		"2102312NSF10K8000130",
	}

	for _, dataCase := range dataCases {
		msg.FillContent(dataCase)
		resp := queryEdgeSoftwareVersion(msg)

		convey.So(resp.Status, convey.ShouldEqual, common.Success)
	}

}

func testSoftwareQueryInvalid() {
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
		resp := queryEdgeSoftwareVersion(msg)

		convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
	}

}

func TestSoftwareQueryPara(t *testing.T) {
	convey.Convey("test software query info", t, func() {
		convey.Convey("test software query info", func() {
			convey.Convey("software query should success", testSoftwareQueryValid)
			convey.Convey("test invalid software query para", testSoftwareQueryInvalid)
		})
	})
}
