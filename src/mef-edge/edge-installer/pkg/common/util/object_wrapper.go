// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package util

import "fmt"

// ObjectWrapper to wrap a json obj
type ObjectWrapper struct {
	object interface{}
}

// NewWrapper create a wrapper within a object in it
func NewWrapper(object interface{}) ObjectWrapper {
	wrapper := ObjectWrapper{object: object}
	return wrapper
}

// GetObject get an object from a json like map
func (j ObjectWrapper) GetObject(key string) ObjectWrapper {
	if j.object == nil {
		return ObjectWrapper{}
	}
	obj, ok := j.object.(map[string]interface{})
	if !ok {
		return ObjectWrapper{}
	}
	target, ok := obj[key]
	if !ok {
		return ObjectWrapper{}
	}
	return NewWrapper(target)
}

// GetData get origin data in the wrapper
func (j ObjectWrapper) GetData() interface{} {
	return j.object
}

// GetString get a string from a json like map
func (j ObjectWrapper) GetString(key string) string {
	if j.object == nil {
		return ""
	}
	obj, ok := j.object.(map[string]interface{})
	if !ok {
		return ""
	}
	if _, exist := obj[key]; !exist {
		return ""
	}
	target, ok := obj[key].(string)
	if !ok {
		return ""
	}
	return target
}

// GetBool get a bool val from a json like map
func (j ObjectWrapper) GetBool(key string) (bool, error) {
	if j.object == nil {
		return false, fmt.Errorf("val for %s nil", key)
	}
	obj, ok := j.object.(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("not json map, cannot get key: %s", key)
	}
	if _, exist := obj[key]; !exist {
		return false, fmt.Errorf("key %s not exist", key)
	}
	target, ok := obj[key].(bool)
	if !ok {
		return false, fmt.Errorf("key %s is not bool", key)
	}
	return target, nil
}

// GetSlice get a slice from a json like map
func (j ObjectWrapper) GetSlice(key string) []ObjectWrapper {
	var wrappers []ObjectWrapper
	if j.object == nil {
		wrapper := ObjectWrapper{}
		return append(wrappers, wrapper)
	}

	obj, ok := j.object.(map[string]interface{})
	if !ok {
		wrapper := ObjectWrapper{}
		return append(wrappers, wrapper)
	}

	if _, exist := obj[key]; !exist {
		wrapper := ObjectWrapper{}
		return append(wrappers, wrapper)
	}

	objects, ok := obj[key].([]interface{})
	if !ok {
		wrapper := ObjectWrapper{}
		return append(wrappers, wrapper)
	}
	if len(objects) == 0 {
		wrapper := ObjectWrapper{}
		return append(wrappers, wrapper)
	}
	for _, v := range objects {
		if v == nil {
			continue
		}
		wrappers = append(wrappers, NewWrapper(v))
	}
	if len(wrappers) == 0 {
		wrapper := ObjectWrapper{}
		return append(wrappers, wrapper)
	} else {
		return wrappers
	}
}
