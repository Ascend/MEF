// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package rand implement the security io.Reader
package rand

import (
	"io"
)

// Reader rand reader to generate security random bytes
var Reader io.Reader

// Read is a helper function that calls Reader.Read using io.ReadFull.
// If the second return value is nil, this function reads exactly
// len(b) bytes and the first return value is always equal to len(b).
func Read(b []byte) (int, error) {
	return io.ReadFull(Reader, b)
}
