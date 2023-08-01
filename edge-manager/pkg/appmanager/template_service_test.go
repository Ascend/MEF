// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.
// Package appmanager for
package appmanager

import (
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"edge-manager/pkg/types"
	"huawei.com/mindxedge/base/common"
)

// TestCreateTemplate test unit for creating templates with normal parameters and abnormal parameters
func TestCreateTemplate(t *testing.T) {
	convey.Convey("create app template with normal para should success", t, testCreateTemplate)
	convey.Convey("create app template with same name", t, testCreateDuplicateTemplate)
	convey.Convey("create app template empty input", t, testCreateTemplateError)
	convey.Convey("create app template invalid input", t, testCreateTemplateInvalid)
	convey.Convey("create app template with two same container name", t, testCreateTemplateDuplicatedContainer)
}

// TestCreateTemplate test unit for updating templates with normal\empty\invalid parameters
func TestUpdateTemplate(t *testing.T) {
	convey.Convey("update app template should success", t, testUpdateTemplate)
	convey.Convey("update app template to a exist name", t, testUpdateTemplateWithExistingName)
	convey.Convey("update app template empty input", t, testUpdateTemplateError)
	convey.Convey("update app template invalid input", t, testUpdateTemplateInvalid)
	convey.Convey("update app template with noneExist id", t, testUpdateNoneExistId)
}

func TestGetTemplate(t *testing.T) {
	convey.Convey("get app template should success", t, testGetTemplate)
	convey.Convey("get app template with not int param", t, testGetTemplateInvalid)
}

func TestListTemplate(t *testing.T) {
	convey.Convey("list app template should success", t, testListTemplates)
	convey.Convey("list app template empty input", t, testListTemplatesError)
	convey.Convey("list app template error input", t, testListTemplatesInvalid)
}

func TestDeleteTemplate(t *testing.T) {
	convey.Convey("delete app template should success", t, testDeleteTemplate)
	convey.Convey("delete app template error input", t, testDeleteTemplateInvalid)
	convey.Convey("delete app template empty success", t, testDeleteTemplateError)
	convey.Convey("delete batch app templates should success", t, testBatchDeleteTemplates)
}

var containerJson = `{
"args":[],
"command":[],
"containerPort":[],
"memRequest": 1024,
"cpuRequest": 1,
"env":[],
"groupId":1024,
"image":"euler_image",
"imageVersion":"2.0",
"name":"afafda",
"userId":1024
}`

// testCreateTemplate test createTemplate for single and multiple containers
func testCreateTemplate() {
	checkTemplateName := []string{`"template1"`, `"template1-with-muti-containers"`}
	var reqDataStd = []string{fmt.Sprintf(`{
    "name":%s,
    "description":"",
    "containers":[%s]}`, checkTemplateName[0], containerJson), fmt.Sprintf(`{
    "name":%s,
    "description":"",
    "containers":[%s,{
			"name":"afafda-2",
            "args":[],
            "command":[],
            "containerPort":[],
			"memRequest": 1024,
            "cpuRequest": 1,
            "env":[],
            "groupId":1024,
            "image":"euler_image",
            "imageVersion":"2.0",
            "userId":1024
}]}`, checkTemplateName[1], containerJson)}

	for i := 0; i < len(reqDataStd); i++ {
		if i >= len(checkTemplateName) {
			break
		}
		resp := createTemplate(reqDataStd[i])
		id, ok := resp.Data.(uint64)
		convey.So(ok, convey.ShouldEqual, true)
		defer RepositoryInstance().deleteTemplates([]uint64{id})
		template, err := RepositoryInstance().getTemplate(id)
		convey.So(err, convey.ShouldBeNil)
		convey.So(fmt.Sprintf(`"%s"`, template.TemplateName), convey.ShouldEqual, checkTemplateName[i])
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
	}
}

func testCreateDuplicateTemplate() {
	sameName := `"template-name-same"`
	templateJson := fmt.Sprintf(`{
    "name":%s,
    "description":"",
    "containers":[%s]}`, sameName, containerJson)
	resp := createTemplate(templateJson)
	id, ok := resp.Data.(uint64)
	convey.So(ok, convey.ShouldEqual, true)
	defer RepositoryInstance().deleteTemplates([]uint64{id})
	resp = createTemplate(templateJson)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorCreateAppTemplate)
}

func testCreateTemplateError() {
	reqData := ""
	resp := createTemplate(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
}

// testCreateTemplateInvalid test for exceeded cpuRequests 100001
func testCreateTemplateInvalid() {
	reqData := `{
    "name":"template2",
    "containers":[{
			"memRequest": 1024,
            "cpuRequest": 100001,
            "groupId":1024,
            "image":"euler_image",
            "imageVersion":"2.0",
            "name":"afafda",
            "userId":1024
}]}`
	resp := createTemplate(reqData)
	convey.So(resp.Status, convey.ShouldNotEqual, common.Success)

	// cpu limit lower than request
	reqData = `{
    "name":"template2",
    "containers":[{
			"memRequest": 1024,
            "cpuRequest": 10,
			"cpuLimit": 9,
            "groupId":1024,
            "image":"euler_image",
            "imageVersion":"2.0",
            "name":"afafda",
            "userId":1024
}]}`
	resp = createTemplate(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorCheckAppTemplateParams)
}

func testCreateTemplateDuplicatedContainer() {
	resp := createTemplate(fmt.Sprintf(`{
    "name":"same-container-name",
    "description":"",
    "containers":[%s,%s]}`, containerJson, containerJson))
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorCheckAppTemplateParams)
}

func testUpdateTemplateWithExistingName() {
	reqData := fmt.Sprintf(`{
	"id":1,
    "name":"template1",
    "description":"",
    "containers":[%s]}`, containerJson)
	resp := createTemplate(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
	defer RepositoryInstance().deleteTemplates([]uint64{resp.Data.(uint64)})

	reqData = fmt.Sprintf(`{
	"id":1,
    "name":"template2",
    "description":"",
    "containers":[%s]}`, containerJson)
	resp = createTemplate(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
	defer RepositoryInstance().deleteTemplates([]uint64{resp.Data.(uint64)})

	id := resp.Data
	reqData = fmt.Sprintf(`{
	"id":%v,
    "name":"template1",
    "description":"",
    "containers":[%s]}`, id, containerJson)
	resp = updateTemplate(reqData)
	convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
}

func testUpdateTemplate() {
	reqData := fmt.Sprintf(`{
	"id":1,
    "name":"template1",
    "description":"",
    "containers":[%s]}`, containerJson)
	resp := createTemplate(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
	defer RepositoryInstance().deleteTemplates([]uint64{resp.Data.(uint64)})

	id := resp.Data
	reqData = fmt.Sprintf(`{
	"id":%v,
    "name":"template-new-name",
    "description":"",
    "containers":[%s]}`, id, containerJson)
	resp = updateTemplate(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
	tempDB, _ := RepositoryInstance().getTemplate(id.(uint64))
	convey.So(tempDB.TemplateName, convey.ShouldEqual, "template-new-name")

	// patch a new container with same name
	reqData = fmt.Sprintf(`{
	"id":%v,
    "name":"template-new-name",
    "description":"",
    "containers":[%s,%s]}`, id, containerJson, containerJson)
	resp = updateTemplate(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorCheckAppTemplateParams)
}

func testUpdateNoneExistId() {
	reqData := fmt.Sprintf(`{
	"id":%v,
    "name":"template-new-name",
    "description":"",
    "containers":[%s]}`, MaxAppTemplate+1, containerJson)
	resp := updateTemplate(reqData)
	convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
}

func testUpdateTemplateError() {
	reqData := ""
	resp := updateTemplate(reqData)
	convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
}

// test updateTemplate for invalid id=0 and id:1 but without missing key:name
func testUpdateTemplateInvalid() {
	reqData := []string{`{
    "id":0
}`, `{
    "id":1
}`}
	for _, req := range reqData {
		resp := updateTemplate(req)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorCheckAppTemplateParams)
	}
}

func testGetTemplate() {
	resp := createTemplate(fmt.Sprintf(`{
    "name":"template-name",
    "description":"",
    "containers":[%s]}`, containerJson))
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
	defer RepositoryInstance().deleteTemplates([]uint64{resp.Data.(uint64)})
	reqData, ok := resp.Data.(uint64)
	if !ok {
		convey.So(ok, convey.ShouldEqual, true, "expectd assert to be true,but get %v", ok)
	}
	resp = getTemplate(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testGetTemplateInvalid() {
	reqData := []string{"string", "12s", "-1", "213"}
	var resp common.RespMsg
	for _, key := range reqData {
		resp = getTemplate(key)
		convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
	}
	reqUnExsitTemp := uint64(1001)
	resp = getTemplate(reqUnExsitTemp)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorTemplateNotFind)
}

func testListTemplates() {
	var reqData = types.ListReq{
		PageNum:  1,
		PageSize: 1,
		Name:     "template1",
	}
	resp := listTemplates(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testListTemplatesError() {
	reqData := ""
	resp := listTemplates(reqData)
	convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
}

func testListTemplatesInvalid() {
	var reqData = types.ListReq{
		PageNum:  1,
		PageSize: exceedPageSize,
		Name:     "template1",
	}
	resp := listTemplates(reqData)
	convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
}

func testDeleteTemplate() {
	reqData := fmt.Sprintf(`{
	"id":1,
    "name":"template1",
    "description":"",
    "containers":[%s]}`, containerJson)
	resp := createTemplate(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
	defer RepositoryInstance().deleteTemplates([]uint64{resp.Data.(uint64)})

	reqData = `{
		"ids": [1]
 	}`
	resp = deleteTemplate(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

// test delete an invalid(negative num) id template
func testDeleteTemplateInvalid() {
	var reqData = `{
		"ids": [-1]
 	}`
	resp := deleteTemplate(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamConvert)
}

// test for batch delete templates
func testBatchDeleteTemplates() {
	templateJson := fmt.Sprintf(`{
    "name":"name1",
    "description":"",
    "containers":[%s]}`, containerJson)
	resp1 := createTemplate(templateJson)
	convey.So(resp1.Status, convey.ShouldEqual, common.Success)

	templateJson2 := fmt.Sprintf(`{
    "name":"name2",
    "description":"",
    "containers":[%s]}`, containerJson)
	resp2 := createTemplate(templateJson2)
	convey.So(resp2.Status, convey.ShouldEqual, common.Success)

	reqData := types.ListReq{
		PageNum:  1,
		PageSize: 10,
		Name:     "",
	}
	resp := listTemplates(reqData)
	fmt.Println(resp.Data)

	delReqData := fmt.Sprintf(`{
		"ids": [%v,%v,3,4,5,6]
 	}`, resp1.Data, resp2.Data)
	resp = deleteTemplate(delReqData)
	fmt.Println(resp.Data)
	convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
}

func testDeleteTemplateError() {
	reqData := ""
	resp := deleteTemplate(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamConvert)
}
