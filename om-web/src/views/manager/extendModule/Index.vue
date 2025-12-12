<template>
  <el-tabs
    v-model="activeModuleOdata"
    :tab-change="clickTabItem"
    style="height: 100%"
  >
    <el-empty v-if="isExtendModuleEmpty()" :description='$t("common.empty")' style="margin-top: 20vh;" />
    <el-tab-pane v-else v-for="module in modulesObjectsList" :label="module.name" :name="module.odata" :key="module.name">
      <div class="container">
        <div v-for="device in devicesObjectsList" class="device-card">
          <div>
            <span>{{ device.name }}</span>
            <el-button style="float: right;" link type="primary" @click="clickSeeMore(device)">{{ $t("common.seeMore") }}</el-button>
          </div>
          <div style="margin-top: 36px;">
            <el-row>
              <el-col :span="8">
                <img src="@/assets/img/manager/device.svg" style="height: 72px; width: 88px;"/>
              </el-col>
              <el-col :span="16">
                <el-row class="field-info">
                  <el-col :span="12">
                    <div class="field-title">name</div>
                    <div class="field-value">{{ getValueByKey(device.attributeValue, 'name') }}</div>
                  </el-col>
                  <el-col :span="12">
                    <div class="field-title">class</div>
                    <div class="field-value">{{ getValueByKey(device.attributeValue, 'class') }}</div>
                  </el-col>
                </el-row>
                <el-row class="field-info">
                  <el-col :span="12">
                    <div class="field-title">present</div>
                    <div class="field-value">{{ getValueByKey(device.attributeValue, 'present') }}</div>
                  </el-col>
                  <el-col :span="12">
                    <div class="field-title">temperature</div>
                    <div class="field-value">{{ getValueByKey(device.attributeValue, 'temperature') }}</div>
                  </el-col>
                </el-row>
              </el-col>
            </el-row>
          </div>
        </div>
      </div>
      <el-pagination
        layout="total, sizes, prev, pager, next"
        :current-page="pagination.pageNum"
        :total="originObjectsList.length"
        :page-size="pagination.pageSize"
        :page-sizes="[10, 20, 50, 100]"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
        style="margin-top: 20px;"
      />
    </el-tab-pane>
  </el-tabs>

  <el-drawer v-model="isShowDrawer" destroy-on-close :title="clickedDevice?.name"  :size="800" >
    <app-tip :tip-text="$t('time.time.saveTip')" />
    <el-table
      :data="clickedDevice.attributeValue"
      max-height="900"
      :header-cell-style="{ background: '#2a2e34' }"
      :cell-style="{ background: '#2a2e34', height: '34px' }"
      :highlight-current-row="false"
    >
      <el-table-column width="150" prop="attributesName" :label="$t('extendModule.attributesName')" />
      <el-table-column width="150" prop="valueType" :label="$t('extendModule.attributesType')" />
      <el-table-column prop="attributesValue" :label="$t('extendModule.attributesValue')">
        <template #default="scope">
          <div v-if="!scope.row.isEdit" style="display: flex; align-items: center;">
            <div v-if="!isJsonAttr(scope.row)" style="display: inline; margin-right: 20px;">
              {{ formatAttributesValue(scope.row) }}
            </div>
            <div v-if="isJsonAttr(scope.row)" style="margin-right: 20px; width: 70%;">
              <el-row
                v-for="key in Object.keys(scope.row.subAttributes)"
                :key="key"
                style="display: flex; align-items: center; width: 100%; height: 40px;"
              >
                <el-col :span="10" class="json-key">{{ key }}</el-col>
                <el-col :span="12" class="json-val">{{ formatJsonAttributesValue(scope.row, key) }}</el-col>
              </el-row>
            </div>
            <img
              src="@/assets/img/edit.svg" alt=""
              v-if="!scope.row.isEdit && isEditableAttr(scope.row)"
              @click="clickEdit(clickedDevice.name, scope.row)"
              style="width: 16px; height: 16px; cursor: pointer;"
            />
          </div>
          <el-form
            ref="attrFormRef"
            :model="attrForm"
            v-if="scope.row.isEdit && !isJsonAttr(scope.row)"
          >
            <el-form-item
              :prop="scope.row.attributesName"
              :error="scope.row.errorText"
            >
              <span style="width: 30%;" v-if="isHiddenAttr(scope.row)">
                <el-input
                  v-model="scope.row.attributesValue"
                  :disabled="!isEditableAttr(scope.row)"
                  type='password'
                  autocomplete="off"
                  @copy.native.capture.prevent="()=>{}"
                  @paste.native.capture.prevent="()=>{}"
                  @cut.native.capture.prevent="()=>{}"
                />
              </span>
              <span style="width: 30%;" v-else>
                <el-input
                  v-if="isStrAttr(scope.row) || isFloatAttr(scope.row)"
                  v-model="scope.row.attributesValue"
                  :disabled="!isEditableAttr(scope.row)"
                />
                <el-input
                  v-if="isNumAttr(scope.row)"
                  v-model.number="scope.row.attributesValue"
                  :disabled="!isEditableAttr(scope.row)"
                />
                <el-switch
                  v-if="isBoolAttr(scope.row)"
                  v-model="scope.row.attributesValue"
                  :disabled="!isEditableAttr(scope.row)"
                />
              </span>
              <el-button
                size="small"
                type="primary"
                @click="clickSave(clickedDevice.odata, clickedDevice.name, scope.row)"
                style="margin-left: 20px;"
              >
                {{ $t('common.save') }}</el-button>
              <el-button
                size="small"
                @click="clickCancel(clickedDevice.odata, scope.row)"
              >
                {{ $t('common.cancel') }}
              </el-button>
            </el-form-item>
          </el-form>
          <el-form v-if="scope.row.isEdit && isJsonAttr(scope.row)" v-model="scope.row.attributesValue" >
            <el-form-item
              v-for="(attr_value, attr_key) in scope.row.subAttributes"
              :label="attr_key"
              :key="attr_key"
              :prop="attr_key"
              :error="scope.row.subAttributes[attr_key].errorText"
            >
              <span style="width: 60%;" v-if="isHiddenAttr(scope.row.subAttributes[attr_key])">
                <el-input
                  v-model="scope.row.attributesValue[attr_key]"
                  :disabled="!isEditableAttr(scope.row.subAttributes[attr_key])"
                  type='password'
                  autocomplete="off"
                  @copy.native.capture.prevent="()=>{}"
                  @paste.native.capture.prevent="()=>{}"
                  @cut.native.capture.prevent="()=>{}"
                />
              </span>
              <span style="width: 60%;" v-else>
                <el-input
                  v-if="
                    isStrAttr(scope.row.subAttributes[attr_key]) ||
                    isFloatAttr(scope.row.subAttributes[attr_key])
                  "
                  v-model="scope.row.attributesValue[attr_key]"
                  :disabled="!isEditableAttr(scope.row.subAttributes[attr_key])"
                />
                <el-input
                  v-if="isNumAttr(scope.row.subAttributes[attr_key])"
                  v-model.number="scope.row.attributesValue[attr_key]"
                  :disabled="!isEditableAttr(scope.row.subAttributes[attr_key])"
                />
                <el-switch
                  v-if="isBoolAttr(scope.row.subAttributes[attr_key])"
                  v-model="scope.row.attributesValue"
                  :disabled="!isEditableAttr(scope.row.subAttributes[attr_key])"
                />
              </span>
            </el-form-item>
            <el-form-item>
              <el-button
                size="small"
                type="primary"
                @click="clickSave(clickedDevice.odata, clickedDevice.name, scope.row)"
              >
                {{ $t('common.save') }}</el-button>
              <el-button
                size="small"
                @click="clickCancel(clickedDevice.odata, scope.row)"
              >
                {{ $t('common.cancel') }}
              </el-button>
            </el-form-item>
          </el-form>
        </template>
      </el-table-column>
    </el-table>
  </el-drawer>
</template>

<script>
import { ref, defineComponent, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';

import { queryAllModules } from '@/api/extendModule';
import { queryByOdataUrl, modifyByOdataUrl } from '@/api/common';
import AppTip from '@/components/AppTip.vue';
import { AppMessage, showErrorAlert, deepCopyDictList } from '@/utils/commonMethods';
import { validateFloat, validateNonFloatNumber } from '@/utils/validator';
import constants from '@/utils/constants';

export default defineComponent({
  name: 'ExtendedModule',
  components: {
    AppTip,
  },
  setup() {
    const { t } = useI18n();

    const modulesObjectsList = ref([]);
    const activeModuleOdata = ref('');
    let currModuleAttributes = [];
    const pagination = ref({
      pageNum: 1,
      pageSize: 10,
    })
    const originObjectsList = ref([]);
    const devicesObjectsList = ref([]);
    const deviceAttributesFormRef = ref();
    const activeDeviceName = ref();
    const attrForm = ref();
    const attrFormRef = ref();
    const attrFormRules = ref();
    const isShowDrawer = ref(false);

    const getAllModulesOdata = async () => {
      // 获取所有模组的 odata
      let { data } = await queryAllModules();
      if (!data?.Members) {
        return;
      }

      // 保存所有模组的 odata
      modulesObjectsList.value = data?.Members.map(item => ({
        odata: item['@odata.id'],
        name: item['@odata.id']?.split('/').pop(),
      }));

      // 记录当前模组（tab）的 odata
      activeModuleOdata.value = data?.Members[0]['@odata.id'];
    }

    const getDevicesByActiveModuleOdata = async () => {
      let { data } = await queryByOdataUrl(activeModuleOdata.value);
      
      // 获取当前模组的 device
      if (!data?.Members || !data?.ModuleInfo) {
        return
      };
      originObjectsList.value = data?.Members.map(item => ({
        odata: item['@odata.id'],
        name: item['@odata.id'].split('/').pop(),
        attributeValue: [],
      }));
      handleCurrentChange(pagination.value.pageNum);

      // 获取当前模组的 attributes
      let attributes = data?.ModuleInfo?.attributes;
      Object.keys(attributes).forEach(key => {
        currModuleAttributes.push({
          attributesName: key,
          valueType: attributes[key].valueType,
          attributesValue: null,
          accessMode: attributes[key].accessMode,
          subAttributes: attributes[key]?.subAttributes,
          errorText: null,
        })
      })

    }

    const isHiddenAttr = (row) => Boolean(row?.accessMode.toLowerCase().indexOf('hide') > -1 && row?.valueType !== 'bool')

    const isNumAttr = (row) => Boolean(row?.valueType === 'int' || row?.valueType === 'long long')

    const isFloatAttr = (row) => Boolean(row?.valueType === 'float')

    const isStrAttr = (row) => Boolean(row?.valueType === 'string')

    const isBoolAttr = (row) => Boolean(row?.valueType === 'bool')

    const isJsonAttr = (row) => Boolean(row?.valueType === 'json')

    const isEditableAttr = (row) => {
      let isBoolRead = Boolean(row.valueType === 'bool' && row.accessMode === 'Read');
      let isWrite = row.accessMode?.toLowerCase().indexOf('write') > -1
      return Boolean(isBoolRead || isWrite);
    }

    const clickTabItem = async (moduleOdata) => {
      activeModuleOdata.value = moduleOdata;
      await getDevicesByActiveModuleOdata();
    }

    const fetchDeviceValue = async (deviceOdata) => {
      if (!deviceOdata) {
        return;
      }
      let { data } = await queryByOdataUrl(deviceOdata);
      if (!data) {
        return
      }
      let attributesList = deepCopyDictList(currModuleAttributes);
      for (let i = 0; i < attributesList.length; i++) {
        let attributeName = attributesList[i].attributesName;
        attributesList[i].attributesValue = data.Attributes[attributeName];
        attributesList[i].isEdit = attributesList[i]?.isEdit ? attributesList[i].isEdit : false;
      }
      for (let j = 0; j < devicesObjectsList.value.length; j++) {
        if (deviceOdata.indexOf(devicesObjectsList.value[j].name) !== -1) {
          devicesObjectsList.value[j].attributeValue = deepCopyDictList(attributesList);
        }
      }
    }

    const clickEdit = (deviceName, attrName) => {
      let deviceIndex;
      for (let j = 0; j < devicesObjectsList.value.length; j++) {
        if (deviceName === devicesObjectsList.value[j].name) {
          deviceIndex = j;
          break;
        }
      }

      for (let i = 0; i < devicesObjectsList.value[deviceIndex].attributeValue.length; i++) {
        if (devicesObjectsList.value[deviceIndex].attributeValue[i].attributesName === attrName.attributesName) {
          devicesObjectsList.value[deviceIndex].attributeValue[i].isEdit = true;
        }
      }
    }

    const clickCancel = async (odata, attrName) => {
      let { data } = await queryByOdataUrl(odata);
      if (!data || attrName.accessMode?.startsWith('Write')) {
        attrName.isEdit = false;
        attrName.attributesValue = null;
        return
      }
      Object.keys(data?.Attributes).forEach(key => {
        if (key === attrName.attributesName) {
          attrName.attributesValue = data?.Attributes[key];
          attrName.isEdit = false;
          attrName.errorText = null;
        }
      })
    }

    const validateAttr = (attr, func) => {
      attr.errorText = func(attr.attributesValue)
      return !attr.errorText;
    }

    const validateJsonAttr = (attr) => {
      for (const subAttr of Object.keys(attr.subAttributes)) {
        if (attr.subAttributes[subAttr].valueType === 'float') {
          attr.subAttributes[subAttr].errorText = validateFloat(
            attr.attributesValue[subAttr]
          );
          if (attr.errorText) {
            return false;
          }
        }
      }
      return true;
    }
    const clickSave = async (odata, deviceName, currDevice) => {
      if (isFloatAttr(currDevice)) {
        if (!validateAttr(currDevice, validateFloat)) {
          return;
        }
        currDevice.attributesValue = parseFloat(currDevice.attributesValue)
      } else if (isNumAttr(currDevice)) {
        if (!validateAttr(currDevice, validateNonFloatNumber)) {
          return;
        }
        currDevice.attributesValue = parseInt(currDevice.attributesValue)
      } else if (currDevice.valueType === 'json') {
        // Traverse the sub-properties in json to determine whether the user input meets the requirements
        if (!validateJsonAttr(currDevice)) {
          return;
        }
      }

      let params = {
        Attributes: {
          [currDevice.attributesName]: currDevice.attributesValue,
        },
      }
      try {
        await modifyByOdataUrl(odata, params);
        AppMessage.success(t('extendModule.deviceSaveSuccessTip', { device: deviceName }));
      } catch (err) {
        showErrorAlert(t('extendModule.deviceSaveFailedTip', { device: deviceName }));
      } finally {
        await fetchDeviceValue(odata);
        currDevice.isEdit = false;
      }
    }

    onMounted(async () => {
      await getAllModulesOdata();
      await getDevicesByActiveModuleOdata();

      devicesObjectsList.value.map(async (item) => {
        await fetchDeviceValue(item.odata)
      });
    })

    const handleSizeChange = (value) => {
      pagination.value.pageSize = value
      devicesObjectsList.value = originObjectsList.value.slice(
        pagination.value.pageSize * (pagination.value.pageNum - 1),
        pagination.value.pageSize * pagination.value.pageNum
      );
    }

    const handleCurrentChange = (value) => {
      pagination.value.pageNum = value
      devicesObjectsList.value = originObjectsList.value.slice(
        pagination.value.pageSize * (pagination.value.pageNum - 1),
        pagination.value.pageSize * pagination.value.pageNum
      );
    }

    const isExtendModuleEmpty = () => Boolean(modulesObjectsList.value.length === 0)

    const getValueByKey = (data, key) => data?.filter(item => item.attributesName === key)[0]?.attributesValue ?? constants.DEFAULT_EMPTY_TEXT;

    const clickedDevice = ref();
    const clickSeeMore = (device) => {
      isShowDrawer.value = true;
      clickedDevice.value = device
    }

    const formatAttributesValue = (row) =>
      row.accessMode?.toLowerCase().indexOf('hide') > -1 ? '*'.repeat(row.attributesValue?.toString().length) : row.attributesValue;
    
    const formatJsonAttributesValue = (jsonData, key) =>
      jsonData.subAttributes[key].accessMode?.toLowerCase().indexOf('hide') > -1
        ? '*'.repeat(jsonData.attributesValue[key]?.toString().length)
        : jsonData.attributesValue[key];

    return {
      attrForm,
      attrFormRef,
      attrFormRules,
      activeModuleOdata,
      activeDeviceName,
      modulesObjectsList,
      devicesObjectsList,
      deviceAttributesFormRef,
      pagination,
      originObjectsList,
      isHiddenAttr,
      isNumAttr,
      isFloatAttr,
      isStrAttr,
      isBoolAttr,
      isJsonAttr,
      isEditableAttr,
      clickTabItem,
      clickEdit,
      clickCancel,
      clickSave,
      handleSizeChange,
      handleCurrentChange,
      isExtendModuleEmpty,
      getValueByKey,
      isShowDrawer,
      clickSeeMore,
      clickedDevice,
      formatAttributesValue,
      formatJsonAttributesValue,
    }
  },
})
</script>

<style scoped>
.container {
  background: var(--el-bg-color-overlay);
  border-radius: 4px;
  padding: 24px 48px;
  height: 70vh;
  overflow: auto;
}

.json-key {
  font-family: 'HarmonyOS_Sans_SC_Regular', serif;
  color: var(--el-text-color-secondary);
  text-align: left;
  line-height: 16px;
  font-weight: 400;
}

.device-card {
  border-radius: 4px;
  background-color: var(--dialog-div-background);
  width: calc(25% - 20px);
  float: left;
  margin-right: 20px;
  padding: 24px;
}

.field-info > div{
  border-left: 1px solid var(--el-border-color);
  padding-left: 14px;
  margin-bottom: 20px;
}

.field-title {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  letter-spacing: 0;
  line-height: 16px;
  font-weight: 200;
  font-family: 'HarmonyOS_Sans_SC_Light', serif;
}

.field-value {
  font-size: 14px;
  color: var(--el-text-color-regular);
  letter-spacing: 0;
  line-height: 16px;
  font-weight: 400;
  margin-top: 8px;
}
</style>
