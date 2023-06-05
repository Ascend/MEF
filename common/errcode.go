// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package common to init error code
package common

const (
	// Success code
	Success = "00000000"
	// ErrorParseBody parse body failed
	ErrorParseBody = "00001001"
	// ErrorGetResponse get response failed
	ErrorGetResponse = "00001002"
	// ErrorsSendSyncMessageByRestful send sync message by restful failed
	ErrorsSendSyncMessageByRestful = "00001003"
	// ErrorResourceOptionNotFound module resource or option not found
	ErrorResourceOptionNotFound = "00001004"
	// ErrorParamInvalid parameter invalid
	ErrorParamInvalid = "00001005"
	// ErrorParamConvert parameter convert error
	ErrorParamConvert = "00001006"
	// ErrorTypeAssert parameter type assert error
	ErrorTypeAssert = "00001007"
	// ErrorNewMsg new msg error
	ErrorNewMsg = "00001008"

	// ErrorCreateAppTemplate failed to create app template
	ErrorCreateAppTemplate = "00002005"
	// ErrorDeleteAppTemplate failed to delete app template
	ErrorDeleteAppTemplate = "00002006"
	// ErrorModifyAppTemplate failed to modify app template
	ErrorModifyAppTemplate = "00002007"
	// ErrorGetAppTemplates failed to get app templates
	ErrorGetAppTemplates = "00002008"
	// ErrorGetAppTemplateDetail failed to get app template detail
	ErrorGetAppTemplateDetail = "00002009"
	// ErrorOperate fail to operate
	ErrorOperate = "00002010"

	// ErrorCreateUser fail to create user
	ErrorCreateUser = "10001001"
	// ErrorCreateUserToDb fail to insert user to db
	ErrorCreateUserToDb = "10001002"
	// ErrorChangePassword fail to modify password
	ErrorChangePassword = "10001003"
	// ErrorLogin fail to login
	ErrorLogin = "10001004"
	// ErrorNeedFirstLogin need first login
	ErrorNeedFirstLogin = "10001005"
	// ErrorLockState in lock state
	ErrorLockState = "10001006"
	// ErrorPasswordRepeat pass repeat in history
	ErrorPasswordRepeat = "10001008"
	// ErrorUserAlreadyFirstLogin user already first login
	ErrorUserAlreadyFirstLogin = "10001009"
	// ErrorPassOrUser username or password error
	ErrorPassOrUser = "10001010"
	// ErrorQueryLock query lock info error
	ErrorQueryLock = "10001011"
	// ErrorQueryHisPassword query history pass error
	ErrorQueryHisPassword = "10001012"

	// ErrorCheckNodeMrgSize failed to check data size while creating
	ErrorCheckNodeMrgSize = "40011000"
	// ErrorNodeMrgDuplicate failed to create/update data cause of duplicate
	ErrorNodeMrgDuplicate = "40012000"
	// ErrorCreateNodeGroup failed to create node group
	ErrorCreateNodeGroup = "40012001"
	// ErrorListNodeGroups failed to list node group
	ErrorListNodeGroups = "40012002"
	// ErrorGetNodeGroup failed to get node group detail
	ErrorGetNodeGroup = "40012003"
	// ErrorModifyNodeGroup failed to modify node group
	ErrorModifyNodeGroup = "40012004"
	// ErrorCountNodeGroup failed to count node groups
	ErrorCountNodeGroup = "40012005"
	// ErrorDeleteNodeGroup failed to delete node groups
	ErrorDeleteNodeGroup = "40012006"

	// ErrorGetNode failed to get node detail
	ErrorGetNode = "40012007"
	// ErrorModifyNode failed to modify node
	ErrorModifyNode = "40012008"
	// ErrorCountNodeByStatus failed to count node by status
	ErrorCountNodeByStatus = "40012009"
	// ErrorListNode failed to list node
	ErrorListNode = "40012010"
	// ErrorListUnManagedNode failed to list node
	ErrorListUnManagedNode = "40012011"
	// ErrorAddNodeToGroup failed to add node to group
	ErrorAddNodeToGroup = "40012012"
	// ErrorAddUnManagedNode failed to add unmanaged node mef
	ErrorAddUnManagedNode = "40012013"
	// ErrorDeleteNode failed to delete node
	ErrorDeleteNode = "40012014"
	// ErrorDeleteNodeFromGroup failed to delete node from group
	ErrorDeleteNodeFromGroup = "40012015"

	// ErrorSendMsgToNode failed to send msg to node
	ErrorSendMsgToNode = "40012016"
	// ErrorGetToken failed to get token
	ErrorGetToken = "40012017"

	// ErrorGetNodeSoftwareVersion failed to delete node groups
	ErrorGetNodeSoftwareVersion = "40012018"

	// ErrorCheckAppMrgSize failed to check data size while creating
	ErrorCheckAppMrgSize = "40021000"
	// ErrorAppParamConvertDb failed to convert request param to db
	ErrorAppParamConvertDb = "40021001"
	// ErrorUnmarshalContainer failed to unmarshal container param from db
	ErrorUnmarshalContainer = "40021002"
	// ErrorAppMrgDuplicate failed to create/update data cause of duplicate
	ErrorAppMrgDuplicate = "40022000"
	// ErrorAppMrgRecodeNoFound failed to query data cause of db data no found
	ErrorAppMrgRecodeNoFound = "40022001"
	// ErrorCreateApp failed to create app
	ErrorCreateApp = "40022002"
	// ErrorQueryApp failed to query app
	ErrorQueryApp = "40022003"
	// ErrorListApp failed to list app
	ErrorListApp = "40022004"
	// ErrorDeployApp failed to deploy app
	ErrorDeployApp = "40022005"
	// ErrorUnDeployApp failed to undeploy app
	ErrorUnDeployApp = "40022006"
	// ErrorUpdateApp failed to update app
	ErrorUpdateApp = "40022007"
	// ErrorDeleteApp failed to delete app
	ErrorDeleteApp = "40022008"
	// ErrorListAppInstancesByID failed to list app instances by id
	ErrorListAppInstancesByID = "40022009"
	// ErrorListAppInstancesByNode failed to list app instances by node
	ErrorListAppInstancesByNode = "40022010"
	// ErrorListAppInstances failed to list app instances
	ErrorListAppInstances = "40022011"
	// ErrorGetAppInstanceCountByNodeGroup failed to count app instances by node group
	ErrorGetAppInstanceCountByNodeGroup = "40022012"

	// ErrorAccountOrPassword incorrect account or password
	ErrorAccountOrPassword = "40031000"
	// ErrorSetEdgeAccountPassword failed to set edge account password
	ErrorSetEdgeAccountPassword = "40031001"
	// ErrorSetEdgeAccount failed to set edge account
	ErrorSetEdgeAccount = "40031002"
	// ErrorAccountOrPasswordEmpty both account and password are required
	ErrorAccountOrPasswordEmpty = "40031003"

	// ErrorMaxEdgeClientsReached max mef-edge clients connection reached
	ErrorMaxEdgeClientsReached = "40041000"

	// ErrorQueryCrt failed to get crt from cert manager
	ErrorQueryCrt = "40042002"

	// ErrorLogCollectEdgeBusy edge is busy
	ErrorLogCollectEdgeBusy = "40052001"
	// ErrorLogCollectEdgeBusiness business error
	ErrorLogCollectEdgeBusiness = "40052002"
	// ErrorLogCollectEdgeParamInvalid parameter error
	ErrorLogCollectEdgeParamInvalid = "40051002"

	// ErrorExportToken export token failed
	ErrorExportToken = "40061001"

	// ErrorGetRootCa failed to get root ca by cert name
	ErrorGetRootCa = "60001001"
	// ErrorIssueSrvCert failed to issue service certificate
	ErrorIssueSrvCert = "60001002"
	// ErrorInValidCaContent failed to valid ca content
	ErrorInValidCaContent = "60001003"
	// ErrorSaveCa failed to  save ca content
	ErrorSaveCa = "60001004"
	// ErrorDeleteRootCa  failed to delete cert file
	ErrorDeleteRootCa = "60001005"
	// ErrorDistributeRootCa failed to distribute cert file
	ErrorDistributeRootCa = "60001006"
	// ErrorGetSecret failed to get secret
	ErrorGetSecret = "60001007"
	// ErrorCreateSecret failed to create secret
	ErrorCreateSecret = "60001008"
	// ErrorExportRootCa failed to export root ca
	ErrorExportRootCa = "60001009"
)

// ErrorMap error code and error msg map
var ErrorMap = map[string]string{
	// Success success code
	Success: "success",
	// ErrorParseBody parse body failed
	ErrorParseBody: "parse request body failed",
	// ErrorGetResponse get response failed
	ErrorGetResponse: "get response failed",
	// ErrorsSendSyncMessageByRestful send sync message by restful failed
	ErrorsSendSyncMessageByRestful: "send sync message by restful failed",
	// ErrorResourceOptionNotFound module resource or option not found info
	ErrorResourceOptionNotFound: "module resource or option not found",
	// ErrorParamInvalid parameter invalid info
	ErrorParamInvalid: "parameter invalid",
	// ErrorTypeAssert parameter type assert error
	ErrorTypeAssert: "parameter type assert error",
	// ErrorCreateAppTemplate failed to create app template info
	ErrorCreateAppTemplate: "failed to create app template",
	// ErrorDeleteAppTemplate failed to delete app template info
	ErrorDeleteAppTemplate: "failed to delete app template",
	// ErrorModifyAppTemplate failed to modify app template info
	ErrorModifyAppTemplate: "failed to modify app template",
	// ErrorGetAppTemplates failed to get app templates
	ErrorGetAppTemplates: "failed to get app templates",
	// ErrorGetAppTemplateDetail failed to get app template detail info
	ErrorGetAppTemplateDetail: "failed to get app template detail",
	// ErrorOperate fail to operate
	ErrorOperate: "failed to operate",

	// ErrorCreateUser fail to create user
	ErrorCreateUser: "failed to create user",
	// ErrorCreateUserToDb fail to insert user to db
	ErrorCreateUserToDb: "fail to insert user to db",
	// ErrorModifyPassword fail to modify password
	ErrorChangePassword: "fail to modify password",
	// ErrorLogin fail to login
	ErrorLogin: "fail to login",
	// ErrorNeedFirstLogin need first login
	ErrorNeedFirstLogin: "need first login",
	// ErrorLockState in lock state
	ErrorLockState: "user or ip in lock state",
	// ErrorPasswordRepeat password same with history
	ErrorPasswordRepeat: "password cannot be the same within recent 5 times",
	// ErrorUserAlreadyFirstLogin user already first login
	ErrorUserAlreadyFirstLogin: "user already first login",
	// ErrorPassOrUser username or password error
	ErrorPassOrUser: "username or password error",
	// ErrorParamConvert convert request error
	ErrorParamConvert: "convert request error",
	// ErrorQueryLock query lock info error
	ErrorQueryLock: "query lock info error",
	// ErrorQueryHisPassword query history password error
	ErrorQueryHisPassword: "query history password error",

	// ErrorCheckNodeMrgSize failed to check data size while creating
	ErrorCheckNodeMrgSize: "failed to check data size while creating",
	// ErrorNodeMrgDuplicate failed to create/update data cause of duplicate
	ErrorNodeMrgDuplicate: "failed to create/update data cause of duplicate",
	// ErrorCreateNodeGroup failed to create node group
	ErrorCreateNodeGroup: "failed to create node group",
	// ErrorListNodeGroups failed to list node group
	ErrorListNodeGroups: "failed to list node group",
	// ErrorGetNodeGroup failed to get node group detail
	ErrorGetNodeGroup: "failed to get node group detail",
	// ErrorModifyNodeGroup failed to modify node group
	ErrorModifyNodeGroup: "failed to modify node group",
	// ErrorCountNodeGroup failed to count node groups
	ErrorCountNodeGroup: "failed to count node groups",
	// ErrorDeleteNodeGroup failed to delete node groups
	ErrorDeleteNodeGroup: "failed to delete node groups",
	// ErrorSendMsgToNode failed to send msg to node
	ErrorSendMsgToNode: "failed to send msg to node",
	// ErrorDeleteNodeFromGroup failed to send msg to node
	ErrorGetNodeSoftwareVersion: "failed to get node version",
	// ErrorGetToken failed to get token
	ErrorGetToken: "failed to get token",

	// ErrorGetNode failed to get node detail
	ErrorGetNode: "failed to get node detail",
	// ErrorModifyNode failed to modify node
	ErrorModifyNode: "failed to modify node",
	// ErrorCountNodeByStatus failed to count node by status
	ErrorCountNodeByStatus: "failed to count node by status",
	// ErrorListNode failed to list node
	ErrorListNode: "failed to list node",
	// ErrorListUnManagedNode failed to list node
	ErrorListUnManagedNode: "failed to list node",
	// ErrorAddNodeToGroup failed to add node to group
	ErrorAddNodeToGroup: "failed to add node to group",
	// ErrorAddUnManagedNode failed to add unmanaged node mef
	ErrorAddUnManagedNode: "failed to add unmanaged node mef",
	// ErrorDeleteNode failed to delete node
	ErrorDeleteNode: "failed to delete node",
	// ErrorDeleteNodeFromGroup failed to delete node from group
	ErrorDeleteNodeFromGroup: "failed to delete node from group",

	// ErrorAppMrgDuplicate failed to check data size while creating
	ErrorCheckAppMrgSize: "failed to check data size while creating",
	// ErrorAppParamConvertDb failed to convert request param to db
	ErrorAppParamConvertDb: "failed to convert request param to db",
	// ErrorUnmarshalContainer failed to unmarshal container param from db
	ErrorUnmarshalContainer: "failed to unmarshal container param from db",
	// ErrorAppMrgDuplicate failed to create/update data cause of duplicate
	ErrorAppMrgDuplicate: "failed to create/update data cause of duplicate",
	// ErrorAppMrgRecodeNoFound failed to query data cause of db data no found
	ErrorAppMrgRecodeNoFound: "failed to query data cause of db data no found",
	// ErrorCreateApp failed to create app
	ErrorCreateApp: "failed to create app",
	// ErrorQueryApp failed to query app
	ErrorQueryApp: "failed to query app",
	// ErrorListApp failed to list app
	ErrorListApp: "failed to list app",
	// ErrorDeployApp failed to deploy app
	ErrorDeployApp: "failed to deploy app",
	// ErrorUnDeployApp failed to undeploy app
	ErrorUnDeployApp: "failed to undeploy app",
	// ErrorUpdateApp failed to update app
	ErrorUpdateApp: "failed to update app",
	// ErrorDeleteApp failed to delete app
	ErrorDeleteApp: "failed to delete app",
	// ErrorListAppInstancesByID failed to list app instances by id
	ErrorListAppInstancesByID: "failed to list app instances by id",
	// ErrorListAppInstancesByNode failed to list app instances by node
	ErrorListAppInstancesByNode: "failed to list app instances by node",
	// ErrorListAppInstances failed to list app instances
	ErrorListAppInstances: "failed to list app instances",
	// ErrorGetAppInstanceCountByNodeGroup failed to count app instances by node group
	ErrorGetAppInstanceCountByNodeGroup: "failed to count app instances by node group",
	// ErrorGetRootCa failed to get root ca by cert name
	ErrorGetRootCa: "failed to get root ca by cert name",
	// ErrorIssueSrvCert failed to issue service certificate
	ErrorIssueSrvCert: "failed to issue service certificate",
	// ErrorInValidCaContent failed to valid ca content
	ErrorInValidCaContent: "failed to valid ca content",
	// ErrorSaveCa failed to save ca content
	ErrorSaveCa: "failed to save ca content",
	// ErrorDeleteRootCa  failed to delete cert file
	ErrorDeleteRootCa: "failed to delete cert file",
	// ErrorDistributeRootCa failed to distribute cert file
	ErrorDistributeRootCa: "failed to distribute cert file",
	// ErrorGetSecret failed to get secret
	ErrorGetSecret: "failed to get secret",
	// ErrorCreateSecret failed to create secret
	ErrorCreateSecret: "failed to create secret",
	// ErrorExportRootCa failed to export root ca
	ErrorExportRootCa: "failed to export root ca",

	// ErrorAccountOrPassword incorrect account or password
	ErrorAccountOrPassword: "incorrect account or password",
	// ErrorSetEdgeAccountPassword failed to set edge account password
	ErrorSetEdgeAccountPassword: "failed to set edge account password",
	// ErrorSetEdgeAccount failed to set edge account
	ErrorSetEdgeAccount: "failed to set edge account",

	// ErrorQueryCrt failed to get crt from cert manager
	ErrorQueryCrt: "failed to get crt from cert manager",
	// ErrorAccountOrPasswordEmpty both account and password are required
	ErrorAccountOrPasswordEmpty: "account or password is empty",
	// ErrorMaxEdgeClientsReached max mef-edge clients connection reached
	ErrorMaxEdgeClientsReached: "max edge client connection reached, please try again later",

	// ErrorLogCollectEdgeBusy edge is busy
	ErrorLogCollectEdgeBusy: "edge node is busy",
	// ErrorLogCollectEdgeBusiness business error
	ErrorLogCollectEdgeBusiness: "failed to collect log due to business error",
	// ErrorLogCollectEdgeParamInvalid parameter error
	ErrorLogCollectEdgeParamInvalid: "failed to collect log due to parameter invalid",

	ErrorExportToken: "export token failed",
}
