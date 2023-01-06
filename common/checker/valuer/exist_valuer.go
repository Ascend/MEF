// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

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
