// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
	// ErrorNodeGroupNotFound node group not found
	ErrorNodeGroupNotFound = "40012019"

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
	// ErrorGetConfigData failed to get token
	ErrorGetConfigData = "40012017"

	// ErrorGetNodeSoftwareVersion failed to get node software version
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

	// ErrorLogDumpBusy edge is busy
	ErrorLogDumpBusy = "40052001"
	// ErrorLogDumpBusiness business error
	ErrorLogDumpBusiness = "40052002"
	// ErrorLogDumpNodeInfoError parameter error
	ErrorLogDumpNodeInfoError = "40051002"

	// ErrorUpdateSoftwareDownloadProgress update software download progress error
	ErrorUpdateSoftwareDownloadProgress = "40062001"
	// ErrorGetSoftwareDownloadProgress get software download progress error
	ErrorGetSoftwareDownloadProgress = "40062002"

	// ErrorListCenterNodeAlarm failed to list center node alarm info
	ErrorListCenterNodeAlarm = "50011001"
	// ErrorListEdgeNodeAlarm  failed to list edge node alarm info
	ErrorListEdgeNodeAlarm = "50011002"
	// ErrorListGroupAlarm  failed to list specific group node alarm info
	ErrorListGroupAlarm = "50011003"
	// ErrorListAlarm failed to list alarms
	ErrorListAlarm = "50011004"
	// ErrorListGroupNodeFromEdgeMgr failed to query nodeGroup information,please recheck provided groupId is valid
	ErrorListGroupNodeFromEdgeMgr = "50011005"
	// ErrorDecodeRespFromEdgeMgr failed to unmarshal response from edge-manager
	ErrorDecodeRespFromEdgeMgr = "50011006"
	// ErrorGetAlarmDetail failed to get alarm detail in db
	ErrorGetAlarmDetail = "50011007"

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
	// ErrorGetRootCaInfo failed to get root ca info by cert name
	ErrorGetRootCaInfo = "60001010"
	// ErrorSaveCrl failed to  save crl content
	ErrorSaveCrl = "60001011"
	// ErrorGetImportedCertsInfo failed to get imported certs info
	ErrorGetImportedCertsInfo = "60001012"
	// ErrorExportToken export token failed
	ErrorExportToken = "60002001"
	// ErrorContentTypeError message content type error
	ErrorContentTypeError = "60002002"
	// ErrorCertTypeError message content type error
	ErrorCertTypeError = "60002003"
	// ErrorNodeNotFound message content type error
	ErrorNodeNotFound = "60002004"
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
	// ErrorParamConvert convert request error
	ErrorParamConvert: "convert request error",

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
	// ErrorGetConfigData failed to get token
	ErrorGetConfigData: "failed to get token",
	// ErrorNodeGroupNotFound node group not found
	ErrorNodeGroupNotFound: "node group not found",

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
	// ErrorGetImportedCertsInfo failed to get imported certs info
	ErrorGetImportedCertsInfo: "failed to get imported certs info",

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

	// ErrorLogDumpBusy edge is busy
	ErrorLogDumpBusy: "edge node is busy",
	// ErrorLogDumpBusiness business error
	ErrorLogDumpBusiness: "failed to collect log due to business error",
	// ErrorLogDumpNodeInfoError parameter error
	ErrorLogDumpNodeInfoError: "failed to collect log due to abnormal node info",
	// ErrorListCenterNodeAlarm failed to list center node alarm info
	ErrorListCenterNodeAlarm: "failed to list center node alarm info",
	// ErrorListEdgeNodeAlarm  failed to list edge node alarm info
	ErrorListEdgeNodeAlarm: "failed to list edge node alarm info",
	// ErrorListGroupAlarm  failed to list specific group node alarm info
	ErrorListGroupAlarm: "failed to list specific group node alarm info",
	// ErrorListGroupNodeFromEdgeMgr failed to query nodeGroup information,please recheck provided groupId
	ErrorListGroupNodeFromEdgeMgr: "failed to query nodeGroup information,please recheck provided groupId is valid",
	// ErrorListAlarm failed to list alarms
	ErrorListAlarm: "ErrorListAlarm failed to list alarms",
	// ErrorDecodeRespFromEdgeMgr failed to unmarshal response from edge-manager
	ErrorDecodeRespFromEdgeMgr: "failed to unmarshal response from edge-manager",
	// ErrorGetAlarmDetail failed to get alarm detail in db
	ErrorGetAlarmDetail: "failed to get alarm detail in db",

	ErrorExportToken: "export token failed",

	ErrorContentTypeError: "message content type error",

	ErrorCertTypeError: "cert type error",

	ErrorNodeNotFound: "empty node found",
}
