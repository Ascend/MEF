// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package common for parameter validate
package common

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"huawei.com/mindx/common/hwlog"
)

// Validator to provide basic parameter verification methods
type Validator struct {
	errs   []string
	unique map[string]struct{}
}

// NewValidator generate a new validator
func NewValidator() *Validator {
	return &Validator{
		errs:   []string{},
		unique: make(map[string]struct{}),
	}
}

// Error collect and return all verification error information
func (v *Validator) Error() error {
	if v != nil && len(v.errs) != 0 {
		errs := strings.Join(v.errs, ", ")
		return errors.New(errs)
	}
	v.errs = []string{}
	v.unique = make(map[string]struct{})
	return nil
}

// ValidateStringLength check string length valid
func (v *Validator) ValidateStringLength(paramName, value string, minLength, maxLength int) *Validator {
	length := len(value)
	if length > maxLength || length < minLength {
		hwlog.RunLog.Errorf("%v (len %v) is too long or too short.", paramName, length)
		v.errs = append(v.errs, paramName+" is too long or too short")
	}
	return v
}

// ValidateStringRegex check string regex valid
func (v *Validator) ValidateStringRegex(paramName, value, regPattern string) *Validator {
	if regPattern != "" {
		if match, err := regexp.MatchString(regPattern, value); !match || err != nil {
			hwlog.RunLog.Errorf("param %v not meet requirement.", paramName)
			v.errs = append(v.errs, paramName+" invalid")
		}
	}
	return v
}

// ValidateInt check int valid
func (v *Validator) ValidateInt(paramName, value string, min, max int) *Validator {
	if val, err := strconv.Atoi(value); err != nil || val < min || val > max {
		hwlog.RunLog.Errorf("param (%v) value invalid.", paramName)
		v.errs = append(v.errs, paramName+" value is invalid")
	}
	return v
}

// ValidateFloat check float valid
func (v *Validator) ValidateFloat(paramName, value string, min, max float64, decimalsNum int) *Validator {
	floatStrs := strings.Split(value, ".")
	const decimalsIndex = 1
	const floatLength = 2
	if len(floatStrs) < floatLength || (len(floatStrs) == floatLength && len(floatStrs[decimalsIndex]) <= decimalsNum) {
		val, err := strconv.ParseFloat(value, BitSize64)
		if err == nil && val >= min && val <= max {
			return v
		}
	}
	hwlog.RunLog.Errorf("param (%v) value invalid.", paramName)
	v.errs = append(v.errs, paramName+" value is invalid")
	return v
}

// ValidateCount check count valid
func (v *Validator) ValidateCount(paramName string, count, min, max int) *Validator {
	if count < min || count > max {
		hwlog.RunLog.Errorf("param (%v) count invalid.", paramName)
		v.errs = append(v.errs, paramName+" count is invalid")
	}
	return v
}

// ValidateGtEq check value is greater than or equal to compare value
func (v *Validator) ValidateGtEq(paramName, compareName string, value, compareValue string) *Validator {
	if a, err := strconv.ParseFloat(value, BitSize64); err == nil {
		if b, err := strconv.ParseFloat(compareValue, BitSize64); err == nil && a >= b {
			return v
		}
	}
	hwlog.RunLog.Errorf("param (%v) must be greater than or equal to param (%v).", paramName, compareName)
	v.errs = append(v.errs, paramName+" and "+compareName+" are invalid")
	return v
}

// ValidateIn check value is in options
func (v *Validator) ValidateIn(paramName, value string, options []string) *Validator {
	if len(options) == 0 {
		return v
	}
	for _, option := range options {
		if option == value {
			return v
		}
	}
	hwlog.RunLog.Errorf("param (%v) value invalid.", paramName)
	v.errs = append(v.errs, paramName+" value is invalid")
	return v
}

// ValidateUnique check value is unique
func (v *Validator) ValidateUnique(paramName, uniqueName, value string) *Validator {
	mapName := uniqueName + value
	if _, ok := v.unique[mapName]; ok {
		hwlog.RunLog.Errorf("param (%v) value must be unique.", paramName)
		v.errs = append(v.errs, paramName+" value must be unique")
	} else {
		v.unique[mapName] = struct{}{}
	}
	return v
}

// AppendError append custom error
func (v *Validator) AppendError(paramName, info string) *Validator {
	hwlog.RunLog.Errorf("param (%v) value invalid,%v.", paramName, info)
	v.errs = append(v.errs, paramName+" "+info)
	return v
}
