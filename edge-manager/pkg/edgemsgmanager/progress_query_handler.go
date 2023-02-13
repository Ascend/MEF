package edgemsgmanager

import (
	"errors"

	"huawei.com/mindx/common/hwlog"

	"edge-manager/pkg/types"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/modulemanager/model"
)

func getNodeUpgradeProgressInfo(nodeName string) (types.ProgressInfo, error) {
	nodeInfo, err := getNodeInfo(nodeName)
	if err != nil {
		hwlog.RunLog.Error("get node upgrade progress failed")
		return types.ProgressInfo{}, errors.New("get node upgrade progress failed")
	}

	return nodeInfo.UpgradeResult, nil
}

// queryEdgeSoftwareUpgradeProgress [method] query edge software upgrade progress
func queryEdgeSoftwareUpgradeProgress(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start query edge software upgrade progress")
	message, ok := input.(*model.Message)
	if !ok {
		hwlog.RunLog.Error("get message failed")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "get message failed", Data: nil}
	}

	uniqueName, ok := message.GetContent().(string)
	if !ok {
		hwlog.RunLog.Error("query edge software upgrade progress failed: para type not valid")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "query edge software upgrade progress" +
			" convert error", Data: nil}
	}

	var processInfo types.ProgressInfo
	if nodeProgress, ok := nodesProgress[uniqueName]; !ok {
		processInfo = nodeProgress
	}

	return common.RespMsg{Status: common.Success, Msg: "", Data: processInfo}
}
