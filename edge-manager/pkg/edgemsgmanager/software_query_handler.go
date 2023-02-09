package edgemsgmanager

import (
	"errors"

	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/modulemanager/model"
)

func getNodeVersionInfo(nodeName string) (map[string]map[string]string, error) {
	nodeInfo, err := getNodeInfo(nodeName)
	if err != nil {
		hwlog.RunLog.Error("get node version failed")
		return map[string]map[string]string{}, errors.New("get node version failed")
	}

	return nodeInfo.SoftwareInfo, nil
}

// queryEdgeSoftwareVersion [method] query edge software version
func queryEdgeSoftwareVersion(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start query edge software version")
	message, ok := input.(*model.Message)
	if !ok {
		hwlog.RunLog.Error("get message failed")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "get message failed", Data: nil}
	}

	uniqueName, ok := message.GetContent().(string)
	if !ok {
		hwlog.RunLog.Error("query edge software version failed: para type not valid")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "query edge software version " +
			"request convert error", Data: nil}
	}

	nodeVersionInfo, err := getNodeVersionInfo(uniqueName)
	if err != nil {
		return common.RespMsg{Status: common.ErrorGetNodeVersion, Msg: "", Data: nil}
	}

	return common.RespMsg{Status: common.Success, Msg: "", Data: nodeVersionInfo}
}
