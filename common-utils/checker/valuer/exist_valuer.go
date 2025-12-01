// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package valuer

// ExistValuer [struct] for exist valuer
type ExistValuer struct {
}

// GetValue [method] for if can get value
func (bv *ExistValuer) GetValue(data interface{}, name string) (bool, error) {
	_, err := GetReflectValueByName(data, name)
	if err != nil {
		return false, err
	}
	return true, nil
}
