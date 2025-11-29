<template>
  <div v-loading="loading">
    <div class="tab-title title">{{ $t("time.time.tabName") }}</div>
    <div class="container">
      <div class="container-title">
        {{ $t("time.time.basicInfo") }}
      </div>
      <app-tip :tip-type="'info'" :tip-text="$t('time.time.saveTip')" />
      <el-form
        ref="timeFormRef"
        :model="timeForm"
        :rules="timeFormRule"
        style="width: 440px; margin-top: 10px;"
        :class="{isEnglish: isEnglish()}"
        class="add-ip-form"
        :hide-required-asterisk="true"
        label-position='left'
        label-width="auto"
      >
        <el-form-item prop="currDeviceTime" class="not-required" :label="$t('time.time.deviceTime')">
          <el-input style="width: 220px;" v-model="timeForm.currDeviceTime" disabled />
        </el-form-item>
        <el-form-item prop="currTimeZone" class="not-required" :label="$t('time.time.timeZone')">
          <el-input style="width: 220px;" v-model="timeForm.currTimeZone" disabled />
        </el-form-item>
        <el-form-item prop="region" class="not-required" :label="$t('time.time.region')">
          <el-select
            style="width: 220px;"
            v-model="timeForm.region"
            :placeholder="$t('common.pleaseSelect')"
            @change="timeFormRef.validate()"
          >
            <el-option v-for="item in areaOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
        <el-form-item prop="timeZone" class="not-required" :label="$t('time.time.timeZone')">
          <el-select
            style="width: 220px;"
            v-model="timeForm.timeZone"
            :placeholder="$t('common.pleaseSelect')"
            @change="changeTimeZone"
          >
            <el-option v-for="item in timeZoneOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
        <el-form-item prop="time" class="not-required" :label="$t('time.time.time')">
          <el-date-picker
            v-model="timeForm.time"
            type="datetime"
            :placeholder="$t('common.pleaseSelect')"
            @visible-change="visibleDatePicker"
            :disabled-date="setDisableDate"
          />
        </el-form-item>
        <el-form-item prop="ntpSwitch" class="not-required" :label="$t('time.time.ntpSwitch')">
          <el-switch v-model="timeForm.ntpSwitch"/>
        </el-form-item>
        <el-form-item v-if="timeForm.ntpSwitch" class="required" prop="preferredAddress" :label="$t('time.time.PreferredNtpServerAddress')">
          <el-row :gutter="10" style="width: 100%;">
            <el-col :span="19">
              <el-input style="width: 220px;" v-model="timeForm.preferredAddress"/>
            </el-col>
            <el-col :span="1" style="padding-top: 6px;">
              <app-popover class="tooltip" :text="$t('time.time.configTimeTip')" />
            </el-col>
          </el-row>
        </el-form-item>
        <el-form-item v-if="timeForm.ntpSwitch" class="not-required" prop="alternateAddress" :label="$t('time.time.AlternateNtpServerAddress')">
          <el-row :gutter="10" style="width: 100%;">
            <el-col :span="19">
              <el-input style="width: 220px;" v-model="timeForm.alternateAddress"/>
            </el-col>
            <el-col :span="1" style="padding-top: 6px;">
              <app-popover class="tooltip" :text="$t('time.time.configNtpTip')" />
            </el-col>
          </el-row>
        </el-form-item>
        <el-form-item class="not-required">
          <el-button type="primary" @click="confirmModifySysTime(timeFormRef)">
            {{ $t("common.save") }}
          </el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script>
import { defineComponent, ref, reactive, onMounted, watch, onUnmounted } from 'vue';
import { useI18n } from 'vue-i18n';

import { querySystemTime, queryNTPService, configNTPService } from '@/api/time';
import { querySystemsSourceInfo, modifySystemsSourceInfo } from '@/api/information';
import { validateIp } from '@/utils/validator';
import AppTip from '@/components/AppTip.vue';
import AppPopover from '@/components/AppPopover.vue';
import timeZone from '@/utils/timeZone';
import timeZoneMapper from '@/utils/timeZoneMapper';
import {
  dateFormat, getLanguageDefaultChinese, AppMessage, showErrorAlert,
  switchToTargetTimezone, handleOperationResponseError
} from '@/utils/commonMethods';

export default defineComponent({
  name: 'ManagerTime',
  components: {
    AppTip,
    AppPopover,
  },
  setup() {
    const { t } = useI18n();
    const timeFormRef = ref();
    const timeZoneOptions = ref([]);
    let isUpdateTime = true;
    let clockId;
    let pickerClockId;
    let lastTimeZone;
    let initFormTimeZone;
    let isInit = true;
  
    const isEnglish = () => getLanguageDefaultChinese() !== 'zh'

    const currConfig = ref({
      time: null,
      timeZone: '',
      ntpRemoteServers: '',
      ntpRemoteServersBak: '',
      clientEnabled: '',
      ntpLocalServers: '',
      serverEnabled: '',
    })

    let timeForm = ref({
      currDeviceTime: null,
      currTimeZone: '',
      region: '',
      timeZone: '',
      time: null,
      ntpSwitch: false,
      preferredAddress: '',
      alternateAddress: '',
    })

    const validateTimeZone = (rule, value, callback) => {
      if (!value) {
        callback()
        return;
      }

      let region = timeForm.value.region;
      let lang = getLanguageDefaultChinese();
      let timeZoneData = timeZone[lang][region];
      for (let area in timeZoneData) {
        if (area.indexOf(value?.toUpperCase()) > -1) {
          callback()
          return;
        }
      }
      callback(new Error(t('time.time.timeZoneErrorTip')))
    }

    const validateAlternateAddress = (rule, value, callback) => {
      if (!value) {
        callback()
        return;
      }
      if (value === timeForm.value.preferredAddress) {
        callback(new Error(t('time.time.alternateAddressErrorTip')))
        return;
      }
      callback();
    }

    const timeFormRule = reactive({
      preferredAddress: [
        {
          required: true,
          message: t('common.notEmpty'),
          trigger: 'blur',
        },
        {
          validator: validateIp,
          trigger: 'blur',
        },
      ],
      alternateAddress: [
        {
          validator: validateIp,
          trigger: 'blur',
        },
        {
          validator: validateAlternateAddress,
          trigger: 'blur',
        },
      ],
      timeZone: [
        {
          validator: validateTimeZone,
          trigger: 'blur',
        },
      ],
    })

    const areaOptions = ref([
      {
        value: 'GMT',
        label: t('time.time.area.GMT'),
      },
      {
        value: 'Pacific',
        label: t('time.time.area.Pacific'),
      },
      {
        value: 'Indian',
        label: t('time.time.area.Indian'),
      },
      {
        value: 'Europe',
        label: t('time.time.area.Europe'),
      },
      {
        value: 'Australia',
        label: t('time.time.area.Australia'),
      },
      {
        value: 'Atlantic',
        label: t('time.time.area.Atlantic'),
      },
      {
        value: 'Asia',
        label: t('time.time.area.Asia'),
      },
      {
        value: 'Arctic',
        label: t('time.time.area.Arctic'),
      },
      {
        value: 'Antarctica',
        label: t('time.time.area.Antarctica'),
      },
      {
        value: 'America',
        label: t('time.time.area.America'),
      },
      {
        value: 'Africa',
        label: t('time.time.area.Africa'),
      },
    ])
    const loading = ref(false);

    const initTimeForm = async () => {
      if (clockId) {
        clearInterval(clockId);
        clockId = null;
      }
      if (pickerClockId) {
        clearInterval(pickerClockId);
        pickerClockId = null;
      }

      let systemTime;
      let systemInfo;
      let ntpInfo;
      try {
        loading.value = true;
        let systemTimeInfo = await querySystemTime(false)
        let systemsSourceInfo = await querySystemsSourceInfo(false)
        let ntpServiceInfo = await queryNTPService(false)
        systemTime = systemTimeInfo?.data
        systemInfo = systemsSourceInfo?.data
        ntpInfo = ntpServiceInfo?.data
        loading.value = false;
      } catch (err) {
        loading.value = false;
        if (err.response.status === 400) {
          showErrorAlert(t('common.requestError'))
        }
      }

      if (systemTime) {
        timeForm.value.currDeviceTime = new Date(systemTime.Datetime).toLocaleString();
        timeForm.value.time = systemTime.Datetime;
        currConfig.value.time = systemTime.Datetime;
      }

      if (systemInfo) {
        let localOffset = systemInfo?.Oem?.DateTimeLocalOffset;
        timeForm.value.currTimeZone = localOffset;
        let area = localOffset?.split(' ')[0];
        currConfig.value.timeZone = area?.split('/')[1];
        if (localOffset?.indexOf('Etc/GMT') !== -1) {
          timeForm.value.region = 'GMT';
          let pattern = /^Etc\/(GMT[\s\S]+)\(/;
          let lang = getLanguageDefaultChinese();
          let timeZoneData = timeZone[lang].GMT;
          for (let areaStr in timeZoneData) {
            if (timeZoneData[areaStr] === pattern.exec(localOffset)[1].trim()) {
              currConfig.value.timeZone = areaStr;
              break;
            }
          }
        } else if (localOffset.indexOf('UTC') !== -1) {
          timeForm.value.region = 'GMT';
        } else {
          timeForm.value.region = area.split('/')[0];
        }
      }

      if (ntpInfo) {
        currConfig.value.ntpRemoteServers = ntpInfo.NTPRemoteServers;
        currConfig.value.ntpRemoteServersBak = ntpInfo.NTPRemoteServersbak;
        currConfig.value.clientEnabled = ntpInfo.ClientEnabled;
        currConfig.value.ntpLocalServers = ntpInfo.NTPLocalServers;
        currConfig.value.serverEnabled = ntpInfo.ServerEnabled;

        timeForm.value.ntpSwitch = currConfig.value.clientEnabled;
        timeForm.value.preferredAddress = currConfig.value.ntpRemoteServers;
        timeForm.value.alternateAddress = currConfig.value.ntpRemoteServersBak;
      }
    }

    const computeTimeZoneOptionsByRegion = (region, currTimeZone) => {
      let lang = getLanguageDefaultChinese();
      let timeZoneData = timeZone[lang][region];
      if (timeZoneData === undefined) {
        return null;
      }

      let timeZoneOptionsList = [];
      let initTimeZoneSelectValue = Object.keys(timeZoneData)[0];
      Object.keys(timeZoneData).forEach(area => {
        if (area.indexOf(currTimeZone?.toUpperCase()) > -1) {
          initTimeZoneSelectValue = area;
        }
        timeZoneOptionsList.push({
          value: area,
          label: timeZoneData[area],
        })
      })

      return {
        initValue: initTimeZoneSelectValue,
        options: timeZoneOptionsList,
      }
    }

    const updateClock = (timeDifference) => {
      let originTime = timeDifference + new Date().getTime();
      timeForm.value.currDeviceTime = new Date(originTime).toLocaleString();
    }

    const updatePickerTime = (timeDifference) => {
      let originTime = timeDifference + new Date().getTime();
      if (isUpdateTime) {
        timeForm.value.time = new Date(originTime).toLocaleString();
      }
      if (isInit) {
        currConfig.value.time = timeForm.value.time;
      }
    }

    const startClock = (start, interval, callback) => {
      if (!start || start.length === 0) {
        return null;
      }

      let origin = new Date(start).getTime();
      let timeDifference = origin - new Date().getTime();
      return setInterval(callback, interval, timeDifference);
    }

    const init = async () => {
      await initTimeForm();

      clockId = startClock(
        timeForm.value.currDeviceTime,
        1000,
        updateClock
      );

      pickerClockId = startClock(
        timeForm.value.time,
        1000,
        updatePickerTime
      )

      let res = computeTimeZoneOptionsByRegion(
        timeForm.value.region,
        currConfig.value.timeZone
      );
      timeForm.value.timeZone = res?.initValue;
      lastTimeZone = timeForm.value.timeZone;
      initFormTimeZone = timeForm.value.timeZone;
      timeZoneOptions.value = res?.options;
    }

    onMounted(async () => {
      await init();
    })

    onUnmounted(() => {
      clearInterval(clockId);
      clockId = null;
      clearInterval(pickerClockId);
      pickerClockId = null;
    })

    const hasChangeTimeZoneInfo = () => timeForm.value.timeZone?.indexOf(currConfig.value.timeZone?.toUpperCase()) === -1 ?? false

    const hasChangeTimeInfo = () => {
      let currTime = new Date(currConfig.value.time);
      let setTime = new Date(timeForm.value.time);
      return currTime.getFullYear() !== setTime.getFullYear() ||
          currTime.getMonth() !== setTime.getMonth() ||
          currTime.getDate() !== setTime.getDate() ||
          currTime.getHours() !== setTime.getHours()
    }

    const setTime = async () => {
      if (!hasChangeTimeInfo() && !hasChangeTimeZoneInfo()) {
        return;
      }

      let systemParams = Object();
      if (hasChangeTimeInfo()) {
        systemParams.DateTime = dateFormat(new Date(timeForm.value.time), 'yyyy-MM-dd hh:mm:ss')
      }
      if (hasChangeTimeZoneInfo()) {
        systemParams.DateTimeLocalOffset = timeZoneMapper[timeForm.value.timeZone]
      }
      await modifySystemsSourceInfo(systemParams, false);
    }

    const hasChangeNtpInfo = () => {
      let isChangeClientEnabled = currConfig.value.clientEnabled !== timeForm.value.ntpSwitch;
      let isChangeNtp = currConfig.value.ntpRemoteServers !== timeForm.value.preferredAddress ||
        currConfig.value.ntpRemoteServersBak !== timeForm.value.alternateAddress;
      return Boolean(isChangeClientEnabled || isChangeNtp)
    }

    const setNtp = async () => {
      // set Ntp server
      if (!hasChangeNtpInfo()) {
        return;
      }

      let ntpParams = {
        ClientEnabled: timeForm.value.ntpSwitch,
        NTPRemoteServers: timeForm.value.preferredAddress,
        NTPRemoteServersbak: timeForm.value.alternateAddress ?? '',
        NTPLocalServers: currConfig.value.ntpLocalServers,
        ServerEnabled: currConfig.value.serverEnabled,
        Target: 'Client',
      }
      await configNTPService(ntpParams, false);
    }

    const confirmModifySysTime = async (formElement) => {
      if (!formElement) {
        return;
      }
      await formElement.validate(async (valid) => {
        if (!valid) {
          return;
        }
        try {
          if (!hasChangeTimeInfo() && !hasChangeTimeZoneInfo() && !hasChangeNtpInfo()) {
            AppMessage.info(t('common.noModifyTip'));
            return;
          }
          loading.value = true;
          await setTime();
          await setNtp();
          loading.value = false;
          AppMessage.success(t('common.saveSuccessfully'));
        } catch (err) {
          handleOperationResponseError(err)
        } finally {
          formElement.resetFields();
          await init();
        }
      })
    }

    watch(timeForm.value, () => {
      let res = computeTimeZoneOptionsByRegion(
        timeForm.value.region,
        currConfig.value.timeZone
      );
      timeZoneOptions.value = res?.options;
    })

    const visibleDatePicker = (isShow) => {
      isUpdateTime = !isShow
      isInit = false;
      if (!isShow) {
        pickerClockId = startClock(
          timeForm.value.time,
          1000,
          updatePickerTime
        )
      } else {
        clearInterval(pickerClockId);
        pickerClockId = null;
      }
    }

    const changeTimeZone = async (val) => {
      await timeFormRef.value.validate(async (valid) => {
        if (!valid) {
          return;
        }

        let currDate = new Date(timeForm.value.time);
        let currTz = timeZoneMapper[lastTimeZone];
        let targetTz = timeZoneMapper[val];

        clearInterval(pickerClockId);
        pickerClockId = null;
        timeForm.value.time = switchToTargetTimezone(currDate, currTz, targetTz);
        pickerClockId = startClock(
          timeForm.value.time,
          1000,
          updatePickerTime
        )
        lastTimeZone = val
      })
    }

    const setDisableDate = (value) => (
      value.getTime() > new Date(2099, 11, 31, 23, 59, 59) ||
      value.getTime() < new Date(2015, 0, 1)
    )

    return {
      areaOptions,
      timeZoneOptions,
      timeForm,
      timeFormRule,
      timeFormRef,
      loading,
      confirmModifySysTime,
      visibleDatePicker,
      changeTimeZone,
      setDisableDate,
      isEnglish,
    }
  },
});
</script>

<style lang="scss" scoped>
.title {
  margin-top: 10px;
}

.container {
  margin-top: 20px;
  padding: 24px;
  background: var(--el-bg-color-overlay);
  height: 80vh;
  border-radius: 4px;
}

.sub-title {
  font-size: 16px;
  letter-spacing: 0;
  line-height: 24px;
  font-weight: 400;
}

.isEnglish {
  width: 500px !important;
}
</style>