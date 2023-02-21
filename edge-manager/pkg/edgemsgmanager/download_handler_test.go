package edgemsgmanager

import (
	"encoding/json"
	"testing"

	"edge-manager/pkg/types"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/modulemanager"
	"huawei.com/mindxedge/base/modulemanager/model"
)

func setup() {
	var err error
	logConfig := &hwlog.LogConfig{OnlyToStdout: true}
	if err = common.InitHwlogger(logConfig, logConfig); err != nil {
		hwlog.RunLog.Errorf("init hwlog failed, %v", err)
	}

	nodesProgress = make(map[string]types.ProgressInfo, 0)
}

func teardown() {

}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	hwlog.RunLog.Infof("exit_code=%d\n", code)
}

func createBaseData() SoftwareDownloadInfo {
	baseContent := `{
    "serialNumbers": ["2102312NSF10K8000130"],
    "softWareName": "edge-installer",
    "softWarVersion": "1.0",
    "downLoadInfo": {
        "package": "GET https://Ascend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz",
        "signFile": "GET https://Ascend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz.cms",
        "crlFile": "GET https://Ascend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz.crl",
        "userName": "FileTransferAccount",
        "password": [118,103,56,115,42,98,35,118,120,54,111]
    	}
	}`

	var req SoftwareDownloadInfo
	err := json.Unmarshal([]byte(baseContent), &req)
	if err != nil {
		hwlog.RunLog.Errorf("unmarshal failed")
		return req
	}
	return req
}
func testDownloadInfo() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed")
	}

	req := createBaseData()
	content, err := json.Marshal(req)
	if err != nil {
		hwlog.RunLog.Errorf("marshal failed")
	}

	msg.FillContent(string(content))
	var p2 = gomonkey.ApplyFunc(modulemanager.SendMessage, func(m *model.Message) error {
		return nil
	})
	defer p2.Reset()

	resp := downloadSoftware(msg)

	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testDownloadInfoSerialNumbersInvalid() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed")
	}

	var p2 = gomonkey.ApplyFunc(modulemanager.SendMessage, func(m *model.Message) error {
		return nil
	})
	defer p2.Reset()

	req := createBaseData()

	dataCases := [][]string{
		{},
		{"_2102312NSF10K8000130"},
		{"-2102312NSF10K8000130"},
		{"2102312NSF10K8000130_"},
		{"2102312NSF10K8000130-"},
		{"2102312NSF10K8000130", "2102312NSF10K8000130"},
		{"21!02312NSF10K800013$0"},
		{"2102312NSF10K80001302102312NSF10K80001302102312NSF10K800013021023"},
	}
	for _, dataCase := range dataCases {
		req.SerialNumbers = dataCase
		content, err := json.Marshal(req)
		if err != nil {
			hwlog.RunLog.Errorf("marshal failed")
		}
		msg.FillContent(string(content))

		resp := downloadSoftware(msg)

		convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
	}

}

func testDownloadInfoSoftWareNameInvalid() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed")
	}

	var p2 = gomonkey.ApplyFunc(modulemanager.SendMessage, func(m *model.Message) error {
		return nil
	})
	defer p2.Reset()

	req := createBaseData()

	dataCases := []string{
		"",
		"AtlasEdge",
	}
	for _, dataCase := range dataCases {
		req.SoftwareName = dataCase
		content, err := json.Marshal(req)
		if err != nil {
			hwlog.RunLog.Errorf("marshal failed")
		}
		msg.FillContent(string(content))

		resp := downloadSoftware(msg)

		convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
	}

}

func testDownloadInfoPackageInvalid() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed")
	}

	var p2 = gomonkey.ApplyFunc(modulemanager.SendMessage, func(m *model.Message) error {
		return nil
	})
	defer p2.Reset()

	req := createBaseData()

	dataCases := []string{
		"",
		" ",
		"GET ",
		"https://Ascend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz",
		"GET https://A!scend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz",
		"GET https://A\nscend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz",
		"GET https://A$scend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz",
		"GET https://A\\scend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz",
		"GET https://A;scend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz",
		"GET https://A&scend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz",
		"GET https://A<scend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz",
		"GET https://A>scend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz",
		"GET https://Ascend -mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz",
	}
	for _, dataCase := range dataCases {
		req.DownloadInfo.Package = dataCase
		content, err := json.Marshal(req)
		if err != nil {
			hwlog.RunLog.Errorf("marshal failed")
		}
		msg.FillContent(string(content))

		resp := downloadSoftware(msg)

		convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
	}
}

func testDownloadInfoSignFileInvalid() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed")
	}

	var p2 = gomonkey.ApplyFunc(modulemanager.SendMessage, func(m *model.Message) error {
		return nil
	})
	defer p2.Reset()

	req := createBaseData()

	failDataCases := []string{
		"",
		" ",
		"GET ",
		"https://Ascend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz",
		"GET https://A!scend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz",
		"GET https://A\nscend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz",
		"GET https://A$scend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz",
		"GET https://A\\scend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz",
		"GET https://A;scend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz",
		"GET https://A&scend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz",
		"GET https://A<scend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz",
		"GET https://A>scend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz",
		"GET https://Ascend -mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz",
	}
	for _, dataCase := range failDataCases {
		req.DownloadInfo.SignFile = &dataCase
		content, err := json.Marshal(req)
		if err != nil {
			hwlog.RunLog.Errorf("marshal failed")
		}
		msg.FillContent(string(content))

		resp := downloadSoftware(msg)

		convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
	}

	req.DownloadInfo.SignFile = nil
	content, err := json.Marshal(req)
	if err != nil {
		hwlog.RunLog.Errorf("marshal failed")
	}
	msg.FillContent(string(content))

	resp := downloadSoftware(msg)

	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testDownloadInfoUserNameInvalid() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed")
	}

	var p2 = gomonkey.ApplyFunc(modulemanager.SendMessage, func(m *model.Message) error {
		return nil
	})
	defer p2.Reset()

	req := createBaseData()
	failDataCases := []string{
		"_FileTransferAccount",
		"-FileTransferAccount",
		"0FileTransferAccount",
		"FileTransferAccountFileTransferAccountFileTransferAccountFileTransf",
	}
	for _, dataCase := range failDataCases {
		req.DownloadInfo.UserName = dataCase
		content, err := json.Marshal(req)
		if err != nil {
			hwlog.RunLog.Errorf("marshal failed")
		}
		msg.FillContent(string(content))

		resp := downloadSoftware(msg)

		convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
	}

	req.DownloadInfo.SignFile = nil
	content, err := json.Marshal(req)
	if err != nil {
		hwlog.RunLog.Errorf("marshal failed")
	}
	msg.FillContent(string(content))

	resp := downloadSoftware(msg)

	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testDownloadInfoPasswordInvalid() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed")
	}

	var p2 = gomonkey.ApplyFunc(modulemanager.SendMessage, func(m *model.Message) error {
		return nil
	})
	defer p2.Reset()

	req := createBaseData()
	failDataCases := [][]byte{
		nil,
	}
	for _, dataCase := range failDataCases {
		req.DownloadInfo.Password = dataCase
		content, err := json.Marshal(req)
		if err != nil {
			hwlog.RunLog.Errorf("marshal failed")
		}
		msg.FillContent(string(content))

		resp := downloadSoftware(msg)

		convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
	}

	req.DownloadInfo.SignFile = nil
	content, err := json.Marshal(req)
	if err != nil {
		hwlog.RunLog.Errorf("marshal failed")
	}
	msg.FillContent(string(content))

	resp := downloadSoftware(msg)

	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func TestDownloadInfo(t *testing.T) {
	convey.Convey("test download info", t, func() {
		convey.Convey("test download info serialNumbers", func() {
			convey.Convey("create configmap should success", testDownloadInfo)
			convey.Convey("test invalid serialNumbers", testDownloadInfoSerialNumbersInvalid)
			convey.Convey("test invalid softWareName", testDownloadInfoSoftWareNameInvalid)
			convey.Convey("test invalid Package", testDownloadInfoPackageInvalid)
			convey.Convey("test invalid SignFile", testDownloadInfoSignFileInvalid)
			convey.Convey("test invalid UserName", testDownloadInfoUserNameInvalid)
			convey.Convey("test invalid Password", testDownloadInfoPasswordInvalid)
		})
	})
}
