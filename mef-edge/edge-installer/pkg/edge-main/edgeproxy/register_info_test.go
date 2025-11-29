// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package edgeproxy

import (
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"edge-installer/pkg/common/constants"
)

func TestGetMsgDest(t *testing.T) {
	mesg := getAsyncMessage()
	convey.Convey("Gives a message that does not meet all requirements\n", t, func() {
		_, err := GetMsgDest(constants.EdgeCore, mesg)
		convey.So(err, convey.ShouldResemble, fmt.Errorf("msg destination not found"))
	})

}
