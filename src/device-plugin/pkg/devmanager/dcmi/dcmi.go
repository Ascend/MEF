// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package dcmi this for dcmi manager
package dcmi

// #cgo LDFLAGS: -ldl
/*
   #include <stddef.h>
   #include <dlfcn.h>
   #include <stdlib.h>
   #include <stdio.h>

   #include "dcmi_interface_api.h"

   void *dcmiHandle;
   #define SO_NOT_FOUND  -99999
   #define FUNCTION_NOT_FOUND  -99998
   #define SUCCESS  0
   #define ERROR_UNKNOWN  -99997
   #define	CALL_FUNC(name,...) if(name##_func==NULL){return FUNCTION_NOT_FOUND;}return name##_func(__VA_ARGS__);

   // dcmi
   int (*dcmi_init_func)();
   int dcmi_init_new(){
   	CALL_FUNC(dcmi_init)
   }

   int (*dcmi_get_card_num_list_func)(int *card_num,int *card_list,int list_length);
   int dcmi_get_card_num_list_new(int *card_num,int *card_list,int list_length){
   	CALL_FUNC(dcmi_get_card_num_list,card_num,card_list,list_length)
   }

   int (*dcmi_get_device_num_in_card_func)(int card_id,int *device_num);
   int dcmi_get_device_num_in_card_new(int card_id,int *device_num){
   	CALL_FUNC(dcmi_get_device_num_in_card,card_id,device_num)
   }

   int (*dcmi_get_device_logic_id_func)(int *device_logic_id,int card_id,int device_id);
   int dcmi_get_device_logic_id_new(int *device_logic_id,int card_id,int device_id){
   	CALL_FUNC(dcmi_get_device_logic_id,device_logic_id,card_id,device_id)
   }

   int (*dcmi_get_device_info_func)(int card_id, int device_id, enum dcmi_main_cmd main_cmd, unsigned int sub_cmd,
   	void *buf, unsigned int *size);
   int dcmi_get_device_info(int card_id, int device_id, enum dcmi_main_cmd main_cmd, unsigned int sub_cmd, void *buf,
   	unsigned int *size){
   	CALL_FUNC(dcmi_get_device_info,card_id,device_id,main_cmd,sub_cmd,buf,size)
   }

   int (*dcmi_get_device_type_func)(int card_id,int device_id,enum dcmi_unit_type *device_type);
   int dcmi_get_device_type(int card_id,int device_id,enum dcmi_unit_type *device_type){
   	CALL_FUNC(dcmi_get_device_type,card_id,device_id,device_type)
   }

   int (*dcmi_get_device_errorcode_v2_func)(int card_id, int device_id, int *error_count, unsigned int *error_code_list,
    unsigned int list_len);
   int dcmi_get_device_errorcode_v2(int card_id, int device_id, int *error_count,
    unsigned int *error_code_list, unsigned int list_len){
    CALL_FUNC(dcmi_get_device_errorcode_v2,card_id,device_id,error_count,error_code_list,list_len)
   }

   int (*dcmi_get_device_health_func)(int card_id, int device_id, unsigned int *health);
   int dcmi_get_device_health(int card_id, int device_id, unsigned int *health){
   	CALL_FUNC(dcmi_get_device_health,card_id,device_id,health)
   }

   int (*dcmi_get_device_chip_info_func)(int card_id, int device_id, struct dcmi_chip_info *chip_info);
   int dcmi_get_device_chip_info(int card_id, int device_id, struct dcmi_chip_info *chip_info){
    CALL_FUNC(dcmi_get_device_chip_info,card_id,device_id,chip_info)
   }

   int (*dcmi_get_device_phyid_from_logicid_func)(unsigned int logicid, unsigned int *phyid);
   int dcmi_get_device_phyid_from_logicid(unsigned int logicid, unsigned int *phyid){
    CALL_FUNC(dcmi_get_device_phyid_from_logicid,logicid,phyid)
   }

   int (*dcmi_get_card_list_func)(int *card_num, int *card_list, int list_len);
   int dcmi_get_card_list(int *card_num, int *card_list, int list_len){
    CALL_FUNC(dcmi_get_card_list,card_num,card_list,list_len)
   }

   int (*dcmi_get_device_errorcode_func)(int card_id, int device_id, int *error_count, unsigned int *error_code,
   int *error_width);
   int dcmi_get_device_errorcode(int card_id, int device_id, int *error_count, unsigned int *error_code,
   int *error_width){
    CALL_FUNC(dcmi_get_device_errorcode,card_id,device_id,error_count,error_code,error_width)
   }

   int (*dcmi_get_card_id_device_id_from_logicid_func)(int *card_id, int *device_id, unsigned int device_logic_id);
   int dcmi_get_card_id_device_id_from_logicid(int *card_id, int *device_id, unsigned int device_logic_id){
    CALL_FUNC(dcmi_get_card_id_device_id_from_logicid,card_id,device_id,device_logic_id)
   }

   int (*dcmi_get_product_type_func)(int card_id, int device_id, char *product_type_str, int buf_size);
   int dcmi_get_product_type(int card_id, int device_id, char *product_type_str, int buf_size){
    CALL_FUNC(dcmi_get_product_type,card_id,device_id,product_type_str,buf_size)
   }

   // load .so files and functions
   int dcmiInit_dl(const char* dcmiLibPath){
   	if (dcmiLibPath == NULL) {
   	   	fprintf (stderr,"lib path is null\n");
   	   	return SO_NOT_FOUND;
   	}
   	dcmiHandle = dlopen(dcmiLibPath,RTLD_LAZY | RTLD_GLOBAL);
   	if (dcmiHandle == NULL){
   		fprintf (stderr,"%s\n",dlerror());
   		return SO_NOT_FOUND;
   	}

   	dcmi_init_func = dlsym(dcmiHandle,"dcmi_init");

   	dcmi_get_card_num_list_func = dlsym(dcmiHandle,"dcmi_get_card_num_list");

   	dcmi_get_device_num_in_card_func = dlsym(dcmiHandle,"dcmi_get_device_num_in_card");

   	dcmi_get_device_logic_id_func = dlsym(dcmiHandle,"dcmi_get_device_logic_id");

   	dcmi_get_device_info_func = dlsym(dcmiHandle,"dcmi_get_device_info");

   	dcmi_get_device_type_func = dlsym(dcmiHandle,"dcmi_get_device_type");

   	dcmi_get_device_health_func = dlsym(dcmiHandle,"dcmi_get_device_health");

   	dcmi_get_device_errorcode_v2_func = dlsym(dcmiHandle,"dcmi_get_device_errorcode_v2");

   	dcmi_get_device_chip_info_func = dlsym(dcmiHandle,"dcmi_get_device_chip_info");

   	dcmi_get_device_phyid_from_logicid_func = dlsym(dcmiHandle,"dcmi_get_device_phyid_from_logicid");

   	dcmi_get_card_list_func = dlsym(dcmiHandle,"dcmi_get_card_list");

   	dcmi_get_device_errorcode_func = dlsym(dcmiHandle,"dcmi_get_device_errorcode");

   	dcmi_get_card_id_device_id_from_logicid_func = dlsym(dcmiHandle,"dcmi_get_card_id_device_id_from_logicid");

	dcmi_get_product_type_func = dlsym(dcmiHandle,"dcmi_get_product_type");

   	return SUCCESS;
   }

   int dcmiShutDown(void){
   	if (dcmiHandle == NULL) {
   		return SUCCESS;
   	}
   	return (dlclose(dcmiHandle) ? ERROR_UNKNOWN : SUCCESS);
   }
*/
import "C"
import (
	"fmt"
	"unsafe"

	"huawei.com/mindx/common/hwlog"

	"Ascend-device-plugin/pkg/devmanager/common"
)

// DcDriverInterface interface for dcmi
type DcDriverInterface interface {
	DcInit() error
	DcShutDown() error

	DcGetDeviceCount() (int32, error)
	DcGetLogicIDList() (int32, []int32, error)
	DcGetDeviceHealth(int32, int32) (int32, error)
	DcGetDeviceErrorCode(int32, int32) (int32, int64, error)
	DcGetChipInfo(int32, int32) (*common.ChipInfo, error)
	DcGetPhysicIDFromLogicID(int32) (int32, error)
	DcGetDeviceLogicID(int32, int32) (int32, error)

	DcGetCardList() (int32, []int32, error)
	DcGetDeviceNumInCard(int32) (int32, error)
	DcGetCardIDDeviceID(int32) (int32, int32, error)
	DcGetVDeviceInfo(int32) (common.VirtualDevInfo, error)
	DcGetProductType(int32, int32) (string, error)
}

const (
	dcmiLibraryName = "libdcmi.so"
)

// DcManager for manager dcmi interface
type DcManager struct{}

// DcInit load symbol and initialize dcmi
func (d *DcManager) DcInit() error {
	dcmiLibPath, err := common.GetDriverLibPath(dcmiLibraryName)
	if err != nil {
		return err
	}
	cDcmiTemplateName := C.CString(dcmiLibPath)
	defer C.free(unsafe.Pointer(cDcmiTemplateName))
	if retCode := C.dcmiInit_dl(cDcmiTemplateName); retCode != C.SUCCESS {
		return fmt.Errorf("dcmi lib load failed, error code: %d", int32(retCode))
	}
	if retCode := C.dcmi_init_new(); retCode != C.SUCCESS {
		return fmt.Errorf("dcmi init failed, error code: %d", int32(retCode))
	}
	return nil
}

// DcShutDown clean the dynamically loaded resource
func (d *DcManager) DcShutDown() error {
	if retCode := C.dcmiShutDown(); retCode != C.SUCCESS {
		return fmt.Errorf("dcmi shut down failed, error code: %d", int32(retCode))
	}

	return nil
}

// DcGetCardList get card list
func (d *DcManager) DcGetCardList() (int32, []int32, error) {
	var ids [common.HiAIMaxCardNum]C.int
	var cNum C.int
	if retCode := C.dcmi_get_card_list(&cNum, &ids[0], common.HiAIMaxCardNum); int32(retCode) != common.Success {
		return common.RetError, nil, fmt.Errorf("get card list failed, error code: %d", int32(retCode))
	}
	// checking card's quantity
	if cNum <= 0 || cNum > common.HiAIMaxCardNum {
		return common.RetError, nil, fmt.Errorf("get error card quantity: %d", int32(cNum))
	}
	var cardNum = int32(cNum)
	var i int32
	var cardIDList []int32
	for i = 0; i < cardNum; i++ {
		cardID := int32(ids[i])
		if cardID < 0 {
			hwlog.RunLog.Errorf("get invalid card ID: %d", cardID)
			continue
		}
		cardIDList = append(cardIDList, cardID)
	}
	return cardNum, cardIDList, nil
}

// DcGetDeviceNumInCard get device number in the npu card
func (d *DcManager) DcGetDeviceNumInCard(cardID int32) (int32, error) {
	if !common.IsValidCardID(cardID) {
		return common.RetError, fmt.Errorf("cardID(%d) is invalid", cardID)
	}
	var deviceNum C.int
	if retCode := C.dcmi_get_device_num_in_card_new(C.int(cardID), &deviceNum); int32(retCode) != common.Success {
		return common.RetError, fmt.Errorf("get device count on the card failed, error code: %d", int32(retCode))
	}
	if !common.IsValidDevNumInCard(int32(deviceNum)) {
		return common.RetError, fmt.Errorf("get error device quantity: %d", int32(deviceNum))
	}
	return int32(deviceNum), nil
}

// DcGetDeviceLogicID get device logicID
func (d *DcManager) DcGetDeviceLogicID(cardID, deviceID int32) (int32, error) {
	if !common.IsValidCardIDAndDeviceID(cardID, deviceID) {
		return common.RetError, fmt.Errorf("cardID(%d) or deviceID(%d) is invalid", cardID, deviceID)
	}
	var logicID C.int
	if retCode := C.dcmi_get_device_logic_id_new(&logicID, C.int(cardID),
		C.int(deviceID)); int32(retCode) != common.Success {
		return common.RetError, fmt.Errorf("failed to get logicID by cardID(%d) and deviceID(%d), error code: %d",
			cardID, deviceID, int32(retCode))
	}

	// check whether logicID is invalid
	if !common.IsValidLogicIDOrPhyID(int32(logicID)) {
		return common.RetError, fmt.Errorf("get invalid logicID: %d", int32(logicID))
	}
	return int32(logicID), nil
}

func convertToString(cgoArr [dcmiVDevResNameLen]C.char) string {
	var charArr []rune
	for _, v := range cgoArr {
		if v == 0 {
			break
		}
		charArr = append(charArr, rune(v))
	}
	return string(charArr)
}

func convertBaseResource(cBaseResource C.struct_dcmi_base_resource) common.CgoBaseResource {
	baseResource := common.CgoBaseResource{
		Token:       uint64(cBaseResource.token),
		TokenMax:    uint64(cBaseResource.token_max),
		TaskTimeout: uint64(cBaseResource.task_timeout),
		VfgID:       uint32(cBaseResource.vfg_id),
		VipMode:     uint8(cBaseResource.vip_mode),
	}
	return baseResource
}

func convertComputingResource(cComputingResource C.struct_dcmi_computing_resource) common.CgoComputingResource {
	computingResource := common.CgoComputingResource{
		Aic:                float32(cComputingResource.aic),
		Aiv:                float32(cComputingResource.aiv),
		Dsa:                uint16(cComputingResource.dsa),
		Rtsq:               uint16(cComputingResource.rtsq),
		Acsq:               uint16(cComputingResource.acsq),
		Cdqm:               uint16(cComputingResource.cdqm),
		CCore:              uint16(cComputingResource.c_core),
		Ffts:               uint16(cComputingResource.ffts),
		Sdma:               uint16(cComputingResource.sdma),
		PcieDma:            uint16(cComputingResource.pcie_dma),
		MemorySize:         uint64(cComputingResource.memory_size),
		EventID:            uint32(cComputingResource.event_id),
		NotifyID:           uint32(cComputingResource.notify_id),
		StreamID:           uint32(cComputingResource.stream_id),
		ModelID:            uint32(cComputingResource.model_id),
		TopicScheduleAicpu: uint16(cComputingResource.topic_schedule_aicpu),
		HostCtrlCPU:        uint16(cComputingResource.host_ctrl_cpu),
		HostAicpu:          uint16(cComputingResource.host_aicpu),
		DeviceAicpu:        uint16(cComputingResource.device_aicpu),
		TopicCtrlCPUSlot:   uint16(cComputingResource.topic_ctrl_cpu_slot),
	}
	return computingResource
}

func convertMediaResource(cMediaResource C.struct_dcmi_media_resource) common.CgoMediaResource {
	mediaResource := common.CgoMediaResource{
		Jpegd: float32(cMediaResource.jpegd),
		Jpege: float32(cMediaResource.jpege),
		Vpc:   float32(cMediaResource.vpc),
		Vdec:  float32(cMediaResource.vdec),
		Pngd:  float32(cMediaResource.pngd),
		Venc:  float32(cMediaResource.venc),
	}
	return mediaResource
}

func convertVDevQueryInfo(cVDevQueryInfo C.struct_dcmi_vdev_query_info) common.CgoVDevQueryInfo {
	name := convertToString(cVDevQueryInfo.name)
	vDevQueryInfo := common.CgoVDevQueryInfo{
		Name:            string(name),
		Status:          uint32(cVDevQueryInfo.status),
		IsContainerUsed: uint32(cVDevQueryInfo.is_container_used),
		Vfid:            uint32(cVDevQueryInfo.vfid),
		VfgID:           uint32(cVDevQueryInfo.vfg_id),
		ContainerID:     uint64(cVDevQueryInfo.container_id),
		Base:            convertBaseResource(cVDevQueryInfo.base),
		Computing:       convertComputingResource(cVDevQueryInfo.computing),
		Media:           convertMediaResource(cVDevQueryInfo.media),
	}
	return vDevQueryInfo
}

func convertVDevQueryStru(cVDevQueryStru C.struct_dcmi_vdev_query_stru) common.CgoVDevQueryStru {
	vDevQueryStru := common.CgoVDevQueryStru{
		VDevID:    uint32(cVDevQueryStru.vdev_id),
		QueryInfo: convertVDevQueryInfo(cVDevQueryStru.query_info),
	}
	return vDevQueryStru
}

// DcGetDeviceVDevResource get virtual device resource info
func (d *DcManager) DcGetDeviceVDevResource(cardID, deviceID int32, vDevID uint32) (common.CgoVDevQueryStru, error) {
	if !common.IsValidCardIDAndDeviceID(cardID, deviceID) {
		return common.CgoVDevQueryStru{}, fmt.Errorf("cardID(%d) or deviceID(%d) is invalid", cardID, deviceID)
	}
	var cMainCmd = C.enum_dcmi_main_cmd(MainCmdVDevMng)
	subCmd := VmngSubCmdGetVDevResource
	var vDevResource C.struct_dcmi_vdev_query_stru
	size := C.uint(unsafe.Sizeof(vDevResource))
	vDevResource.vdev_id = C.uint(vDevID)
	if retCode := C.dcmi_get_device_info(C.int(cardID), C.int(deviceID), cMainCmd, C.uint(subCmd),
		unsafe.Pointer(&vDevResource), &size); int32(retCode) != common.Success {
		return common.CgoVDevQueryStru{}, fmt.Errorf("get device info failed, error is: %d", int32(retCode))
	}
	return convertVDevQueryStru(vDevResource), nil
}

func convertSocTotalResource(cSocTotalResource C.struct_dcmi_soc_total_resource) common.CgoSocTotalResource {
	socTotalResource := common.CgoSocTotalResource{
		VDevNum:   uint32(cSocTotalResource.vdev_num),
		VfgNum:    uint32(cSocTotalResource.vfg_num),
		VfgBitmap: uint32(cSocTotalResource.vfg_bitmap),
		Base:      convertBaseResource(cSocTotalResource.base),
		Computing: convertComputingResource(cSocTotalResource.computing),
		Media:     convertMediaResource(cSocTotalResource.media),
	}
	for i := uint32(0); i < uint32(cSocTotalResource.vdev_num) && i < dcmiMaxVdevNum; i++ {
		socTotalResource.VDevID = append(socTotalResource.VDevID, uint32(cSocTotalResource.vdev_id[i]))
	}
	return socTotalResource
}

// DcGetDeviceTotalResource get device total resource info
func (d *DcManager) DcGetDeviceTotalResource(cardID, deviceID int32) (common.CgoSocTotalResource, error) {
	if !common.IsValidCardIDAndDeviceID(cardID, deviceID) {
		return common.CgoSocTotalResource{}, fmt.Errorf("cardID(%d) or deviceID(%d) is invalid", cardID, deviceID)
	}
	var cMainCmd = C.enum_dcmi_main_cmd(MainCmdVDevMng)
	subCmd := VmngSubCmdGetTotalResource
	var totalResource C.struct_dcmi_soc_total_resource
	size := C.uint(unsafe.Sizeof(totalResource))
	if retCode := C.dcmi_get_device_info(C.int(cardID), C.int(deviceID), cMainCmd, C.uint(subCmd),
		unsafe.Pointer(&totalResource), &size); int32(retCode) != common.Success {
		return common.CgoSocTotalResource{}, fmt.Errorf("get device info failed, error is: %d", int32(retCode))
	}
	if uint32(totalResource.vdev_num) > dcmiMaxVdevNum {
		return common.CgoSocTotalResource{}, fmt.Errorf("get error virtual quantity: %d",
			uint32(totalResource.vdev_num))
	}

	return convertSocTotalResource(totalResource), nil
}

func convertSocFreeResource(cSocFreeResource C.struct_dcmi_soc_free_resource) common.CgoSocFreeResource {
	socFreeResource := common.CgoSocFreeResource{
		VfgNum:    uint32(cSocFreeResource.vfg_num),
		VfgBitmap: uint32(cSocFreeResource.vfg_bitmap),
		Base:      convertBaseResource(cSocFreeResource.base),
		Computing: convertComputingResource(cSocFreeResource.computing),
		Media:     convertMediaResource(cSocFreeResource.media),
	}
	return socFreeResource
}

// DcGetDeviceFreeResource get device free resource info
func (d *DcManager) DcGetDeviceFreeResource(cardID, deviceID int32) (common.CgoSocFreeResource, error) {
	if !common.IsValidCardIDAndDeviceID(cardID, deviceID) {
		return common.CgoSocFreeResource{}, fmt.Errorf("cardID(%d) or deviceID(%d) is invalid", cardID, deviceID)
	}
	var cMainCmd = C.enum_dcmi_main_cmd(MainCmdVDevMng)
	subCmd := VmngSubCmdGetFreeResource
	var freeResource C.struct_dcmi_soc_free_resource
	size := C.uint(unsafe.Sizeof(freeResource))
	if retCode := C.dcmi_get_device_info(C.int(cardID), C.int(deviceID), cMainCmd, C.uint(subCmd),
		unsafe.Pointer(&freeResource), &size); int32(retCode) != common.Success {
		return common.CgoSocFreeResource{}, fmt.Errorf("get device info failed, error is: %d", int32(retCode))
	}
	return convertSocFreeResource(freeResource), nil
}

// DcVGetDeviceInfo get vdevice resource info
func (d *DcManager) DcVGetDeviceInfo(cardID, deviceID int32) (common.VirtualDevInfo, error) {
	if !common.IsValidCardIDAndDeviceID(cardID, deviceID) {
		return common.VirtualDevInfo{}, fmt.Errorf("cardID(%d) or deviceID(%d) is invalid", cardID, deviceID)
	}
	var unitType C.enum_dcmi_unit_type
	if retCode := C.dcmi_get_device_type(C.int(cardID), C.int(deviceID), &unitType); int32(retCode) != 0 {
		return common.VirtualDevInfo{}, fmt.Errorf("get device type failed, error is: %d", int32(retCode))
	}
	if int32(unitType) != common.NpuType {
		return common.VirtualDevInfo{}, fmt.Errorf("not support unit type: %d", int32(unitType))
	}

	cgoDcmiSocTotalResource, err := d.DcGetDeviceTotalResource(cardID, deviceID)
	if err != nil {
		return common.VirtualDevInfo{}, fmt.Errorf("get device total resource failed, error is: %#v", err)
	}

	cgoDcmiSocFreeResource, err := d.DcGetDeviceFreeResource(cardID, deviceID)
	if err != nil {
		return common.VirtualDevInfo{}, fmt.Errorf("get device free resource failed, error is: %#v", err)
	}

	dcmiVDevInfo := common.VirtualDevInfo{
		TotalResource: cgoDcmiSocTotalResource,
		FreeResource:  cgoDcmiSocFreeResource,
	}
	for _, vDevID := range cgoDcmiSocTotalResource.VDevID {
		cgoVDevQueryStru, err := d.DcGetDeviceVDevResource(cardID, deviceID, vDevID)
		if err != nil {
			return common.VirtualDevInfo{}, fmt.Errorf("get device virtual resource failed, error is: %#v", err)
		}
		dcmiVDevInfo.VDevInfo = append(dcmiVDevInfo.VDevInfo, cgoVDevQueryStru)
	}
	return dcmiVDevInfo, nil
}

// DcGetCardIDDeviceID get card id and device id from logic id
func (d *DcManager) DcGetCardIDDeviceID(logicID int32) (int32, int32, error) {
	if !common.IsValidLogicIDOrPhyID(logicID) {
		return common.RetError, common.RetError, fmt.Errorf("input invalid logicID: %d", logicID)
	}
	var cardID, deviceID C.int
	if retCode := C.dcmi_get_card_id_device_id_from_logicid(&cardID, &deviceID,
		C.uint(logicID)); int32(retCode) != common.Success {
		return common.RetError, common.RetError,
			fmt.Errorf("failed to get card id and device id by logicID(%d), errorcode is: %d", logicID,
				int32(retCode))
	}
	if !common.IsValidCardIDAndDeviceID(int32(cardID), int32(deviceID)) {
		return common.RetError, common.RetError, fmt.Errorf("failed to get card id and device id, "+
			"cardID(%d) or deviceID(%d) is invalid", int32(cardID), int32(deviceID))
	}

	return int32(cardID), int32(deviceID), nil
}

// DcGetVDeviceInfo get virtual device info by logic id
func (d *DcManager) DcGetVDeviceInfo(logicID int32) (common.VirtualDevInfo, error) {
	if !common.IsValidLogicIDOrPhyID(logicID) {
		return common.VirtualDevInfo{}, fmt.Errorf("input invalid logicID: %d", logicID)
	}
	cardID, deviceID, err := d.DcGetCardIDDeviceID(logicID)
	if err != nil {
		return common.VirtualDevInfo{}, fmt.Errorf("get card id and device id failed, error is: %#v", err)
	}

	dcmiVDevInfo, err := d.DcVGetDeviceInfo(cardID, deviceID)
	if err != nil {
		return common.VirtualDevInfo{}, fmt.Errorf("get virtual device info failed, error is: %#v", err)
	}
	return dcmiVDevInfo, nil
}

// DcGetDeviceErrorCode get the error count and errorcode of the device,only return the first errorcode
func (d *DcManager) DcGetDeviceErrorCode(cardID, deviceID int32) (int32, int64, error) {
	if !common.IsValidCardIDAndDeviceID(cardID, deviceID) {
		return common.RetError, common.RetError, fmt.Errorf("cardID(%d) or deviceID(%d) is invalid", cardID,
			deviceID)
	}
	var errCount C.int
	var errCodeArray [common.MaxErrorCodeCount]C.uint
	if retCode := C.dcmi_get_device_errorcode_v2(C.int(cardID), C.int(deviceID), &errCount, &errCodeArray[0],
		common.MaxErrorCodeCount); int32(retCode) != common.Success {
		return common.RetError, common.RetError, fmt.Errorf("failed to obtain the device errorcode based on card_id("+
			"%d) and device_id(%d), error code: %d, error count: %d", cardID, deviceID, int32(retCode),
			int32(errCount))
	}

	if int32(errCount) < 0 || int32(errCount) > common.MaxErrorCodeCount {
		return common.RetError, common.RetError, fmt.Errorf("get wrong errorcode count, "+
			"card_id(%d) and device_id(%d), errorcode count: %d", cardID, deviceID, int32(errCount))
	}

	return int32(errCount), int64(errCodeArray[0]), nil
}

// DcGetDeviceCount get device count
func (d *DcManager) DcGetDeviceCount() (int32, error) {
	devNum, _, err := d.DcGetLogicIDList()
	if err != nil {
		return common.RetError, fmt.Errorf("get device count failed, error: %#v", err)
	}
	return devNum, nil
}

// DcGetLogicIDList get device logic id list
func (d *DcManager) DcGetLogicIDList() (int32, []int32, error) {
	var logicIDs []int32
	var totalNum int32
	_, cardList, err := d.DcGetCardList()
	if err != nil {
		return common.RetError, logicIDs, fmt.Errorf("get card list failed, error: %#v", err)
	}
	for _, cardID := range cardList {
		devNumInCard, err := d.DcGetDeviceNumInCard(cardID)
		if err != nil {
			return common.RetError, logicIDs, fmt.Errorf("get device num by cardID: %d failed, error: %#v",
				cardID, err)
		}
		totalNum += devNumInCard
		if totalNum > common.HiAIMaxDeviceNum*common.HiAIMaxCardNum {
			return common.RetError, nil, fmt.Errorf("get device num: %d greater than %d",
				totalNum, common.HiAIMaxDeviceNum*common.HiAIMaxCardNum)
		}
		for devID := int32(0); devID < devNumInCard; devID++ {
			logicID, err := d.DcGetDeviceLogicID(cardID, devID)
			if err != nil {
				return common.RetError, nil, fmt.Errorf("get device (cardID: %d, deviceID: %d) logic id "+
					"failed, error: %#v", cardID, devID, err)
			}
			logicIDs = append(logicIDs, logicID)
		}
	}
	return totalNum, logicIDs, nil
}

// DcGetDeviceHealth get device health
func (d *DcManager) DcGetDeviceHealth(cardID, deviceID int32) (int32, error) {
	if !common.IsValidCardIDAndDeviceID(cardID, deviceID) {
		return common.RetError, fmt.Errorf("cardID(%d) or deviceID(%d) is invalid", cardID, deviceID)
	}
	var health C.uint
	if retCode := C.dcmi_get_device_health(C.int(cardID), C.int(deviceID),
		&health); int32(retCode) != common.Success {
		return common.RetError, fmt.Errorf("get device (cardID: %d, deviceID: %d) health state failed, error "+
			"code: %d", cardID, deviceID, int32(retCode))
	}
	if common.IsGreaterThanOrEqualInt32(int64(health)) {
		return common.RetError, fmt.Errorf("get wrong health state , device (cardID: %d, deviceID: %d) "+
			"health: %d", cardID, deviceID, int64(health))
	}
	return int32(health), nil
}

func convertUCharToCharArr(cgoArr [maxChipNameLen]C.uchar) []byte {
	var charArr []byte
	for _, v := range cgoArr {
		if v == 0 {
			break
		}
		charArr = append(charArr, byte(v))
	}
	return charArr
}

// DcGetChipInfo get the chip info by cardID and deviceID
func (d *DcManager) DcGetChipInfo(cardID, deviceID int32) (*common.ChipInfo, error) {
	if !common.IsValidCardIDAndDeviceID(cardID, deviceID) {
		return nil, fmt.Errorf("cardID(%d) or deviceID(%d) is invalid", cardID, deviceID)
	}
	var chipInfo C.struct_dcmi_chip_info
	if rCode := C.dcmi_get_device_chip_info(C.int(cardID), C.int(deviceID), &chipInfo); int32(rCode) != common.Success {
		return nil, fmt.Errorf("get device ChipInfo information failed, cardID(%d), deviceID(%d),"+
			" error code: %d", cardID, deviceID, int32(rCode))
	}

	name := convertUCharToCharArr(chipInfo.chip_name)
	cType := convertUCharToCharArr(chipInfo.chip_type)
	ver := convertUCharToCharArr(chipInfo.chip_ver)

	chip := &common.ChipInfo{
		Name:    string(name),
		Type:    string(cType),
		Version: string(ver),
	}
	if !common.IsValidChipInfo(chip) {
		return nil, fmt.Errorf("get device ChipInfo information failed, chip info is empty,"+
			" cardID(%d), deviceID(%d)", cardID, deviceID)
	}

	return chip, nil
}

// DcGetPhysicIDFromLogicID get physicID from logicID
func (d *DcManager) DcGetPhysicIDFromLogicID(logicID int32) (int32, error) {
	if !common.IsValidLogicIDOrPhyID(logicID) {
		return common.RetError, fmt.Errorf("logicID(%d) is invalid", logicID)
	}
	var physicID C.uint
	if rCode := C.dcmi_get_device_phyid_from_logicid(C.uint(logicID), &physicID); int32(rCode) != common.Success {
		return common.RetError, fmt.Errorf("get physic id from logicID(%d) failed, error code: %d", logicID, int32(rCode))
	}
	if !common.IsValidLogicIDOrPhyID(int32(physicID)) {
		return common.RetError, fmt.Errorf("get wrong physicID(%d) from logicID(%d)", uint32(physicID), logicID)
	}
	return int32(physicID), nil
}

// DcGetProductType get product type by dcmi interface
func (d *DcManager) DcGetProductType(cardID, deviceID int32) (string, error) {
	cProductType := C.CString(string(make([]byte, productTypeLen)))
	defer C.free(unsafe.Pointer(cProductType))
	err := C.dcmi_get_product_type(C.int(cardID), C.int(deviceID), (*C.char)(cProductType), productTypeLen)
	if err != 0 {
		return "", fmt.Errorf("get product type failed, errCode: %d", err)
	}
	return C.GoString(cProductType), nil
}
