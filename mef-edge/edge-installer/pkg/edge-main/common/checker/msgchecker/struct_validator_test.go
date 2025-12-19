// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package msgchecker

import (
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/hwlog"
)

type People struct {
	Name       string  `validate:"^[a-zA-Z]{1,16}$"`
	ID         string  `validate:"^[\\d]{5}$"`
	MotherName *string `validate:"^[a-zA-Z]{3,16}$"`
	Age        int
}

type Student struct {
	People
	Grade string `validate:"^[a-z]{3,5}$"`
}

type HeadTeacher struct {
	TeacherName string `validate:"^[a-zA-Z]{1,16}$"`
}

type Class struct {
	ClassName     string `validate:"^[a-z]{1,8}$"`
	Students      []Student
	Courses       []string          `validate:"^[a-z]{1,8}$"`
	Teachers      map[string]string `validate:"^[a-z]{3,16}$;^[a-zA-Z]{3,8}$"`
	MasterTeacher *HeadTeacher
}

type structValidateTestcase struct {
	description string
	class       Class
	shouldErr   bool
	assert      convey.Assertion
	expected    interface{}
}

var motherNameA = "li"
var motherNameB = "lifang"
var headTeacher = HeadTeacher{TeacherName: "zhang ming"}

var testcase = []structValidateTestcase{
	{
		description: "test class name failed",
		class:       Class{ClassName: "class1"},
		shouldErr:   true,
		assert:      convey.ShouldContainSubstring,
		expected:    "Class.ClassName",
	},
	{
		description: "test student Name failed",
		class: Class{ClassName: "five",
			Students: []Student{
				{
					Grade:  "12",
					People: People{Name: ""},
				},
			},
		},
		shouldErr: true,
		assert:    convey.ShouldContainSubstring,
		expected:  "Class.Students.People.Name",
	},
	{
		description: "test student MotherName failed",
		class: Class{ClassName: "five",
			Students: []Student{
				{
					Grade: "12",
					People: People{
						Name:       "lifei",
						ID:         "12345",
						MotherName: &motherNameA,
					},
				},
			},
		},
		shouldErr: true,
		assert:    convey.ShouldContainSubstring,
		expected:  "Class.Students.People.MotherName",
	},

	{
		description: "test student grade failed",
		class: Class{ClassName: "five",
			Students: []Student{
				{
					Grade: "12",
					People: People{
						Name:       "lifei",
						ID:         "12345",
						MotherName: &motherNameB,
					},
				},
			},
		},
		shouldErr: true,
		assert:    convey.ShouldContainSubstring,
		expected:  "Class.Students.Grade",
	},

	{
		description: "test class Courses failed",
		class: Class{
			ClassName: "five",
			Students: []Student{
				{
					Grade: "three",
					People: People{
						Name:       "lifei",
						ID:         "12345",
						MotherName: &motherNameB,
					},
				},
			},
			Courses: []string{"math", "psychology"},
		},
		shouldErr: true,
		assert:    convey.ShouldContainSubstring,
		expected:  "Class.Courses",
	},

	{
		description: "test class Teachers key failed",
		class: Class{
			ClassName: "five",
			Students: []Student{
				{
					Grade: "three",
					People: People{
						Name:       "lifei",
						ID:         "12345",
						MotherName: &motherNameB,
					},
				},
			},
			Courses:  []string{"math", "english"},
			Teachers: map[string]string{"li": "math"},
		},
		shouldErr: true,
		assert:    convey.ShouldContainSubstring,
		expected:  "Class.Teachers",
	},

	{
		description: "test class Teachers value failed",
		class: Class{
			ClassName: "five",
			Students: []Student{
				{
					Grade: "three",
					People: People{
						Name:       "lifei",
						ID:         "12345",
						MotherName: &motherNameB,
					},
				},
			},
			Courses:  []string{"math", "english"},
			Teachers: map[string]string{"lifang": "math", "zhangsan": "psychology"},
		},
		shouldErr: true,
		assert:    convey.ShouldContainSubstring,
		expected:  "Class.Teachers",
	},
	{
		description: "test class MasterTeacher value failed",
		class: Class{
			ClassName: "five",
			Students: []Student{
				{
					Grade: "three",
					People: People{
						Name:       "lifei",
						ID:         "12345",
						MotherName: &motherNameB,
					},
				},
			},
			Courses:       []string{"math", "english"},
			Teachers:      map[string]string{"lifang": "math", "zhangsan": "chinese"},
			MasterTeacher: &headTeacher,
		},
		shouldErr: true,
		assert:    convey.ShouldContainSubstring,
		expected:  "Class.MasterTeacher.TeacherName",
	},
	{
		description: "test class ok",
		class: Class{
			ClassName: "five",
			Students: []Student{
				{
					Grade: "three",
					People: People{
						Name:       "lifei",
						ID:         "12345",
						MotherName: &motherNameB,
					},
				},
			},
			Courses:  []string{"math", "english"},
			Teachers: map[string]string{"lifang": "math", "zhangsan": "english"},
		},
		shouldErr: false,
		assert:    convey.ShouldEqual,
		expected:  nil,
	},
}

func TestStructValidate(t *testing.T) {
	var err error
	convey.Convey("test class struct para", t, func() {
		for _, tc := range testcase {
			fmt.Println(tc.description)

			err = validateStruct(tc.class)
			if err != nil {
				hwlog.RunLog.Errorf("check struct para failed: %v", err)
			}

			if tc.shouldErr {
				convey.So(err.Error(), tc.assert, tc.expected)
			} else {
				convey.So(err, tc.assert, tc.expected)
			}
		}
	})
}
