// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
