<template>
  <el-row style="width: 100%; height: calc(100vh - 48px); margin-top: -12px; background: #181b20;">
    <el-col :span="5" style="padding: 28px;">
      <div class="block" style="height: 36vh;">
        <div class="sub-title">
          <img src="@/assets/img/home/setting.svg" alt=""/>
          {{ $t("home.system.title") }}
        </div>
        <el-row class="sub-block">
          <el-col :span="12">
            <div class="sub-block-title">{{ $t("home.system.hostname") }}</div>
            <div class="sub-block-content">
              <el-popover placement="right" trigger="hover">
                <template #reference>
                  {{ formatText(systemInfo.hostname ?? constants.DEFAULT_EMPTY_TEXT, 15) }}
                </template>
                <div style="word-break: break-word;">
                  {{ systemInfo.hostname ?? constants.DEFAULT_EMPTY_TEXT }}
                </div>
              </el-popover>
            </div>
          </el-col>
          <el-col :span="12">
            <div class="sub-block-title">{{ $t("home.system.uptime") }}</div>
            <div class="sub-block-content">
              <el-popover placement="right" trigger="hover">
                <template #reference>
                  {{ formatText(systemInfo.runtime ?? constants.DEFAULT_EMPTY_TEXT, 16) }}
                </template>
                <div style="word-break: break-word;">
                  {{ systemInfo.runtime ?? constants.DEFAULT_EMPTY_TEXT }}
                </div>
              </el-popover>
            </div>
          </el-col>
        </el-row>
        <el-row class="sub-block">
          <el-col :span="12">
            <div class="sub-block-title">{{ $t("home.system.os") }}</div>
            <div class="sub-block-content">
              <el-popover placement="right" trigger="hover">
                <template #reference>
                  {{ formatText(systemInfo.os ?? constants.DEFAULT_EMPTY_TEXT, 15) }}
                </template>
                <div style="word-break: break-word;">
                  {{ systemInfo.os ?? constants.DEFAULT_EMPTY_TEXT }}
                </div>
              </el-popover>
            </div>
          </el-col>
          <el-col :span="12">
            <div class="sub-block-title">{{ $t("home.system.firmwareVersion") }}</div>
            <div class="sub-block-content">
              <el-popover placement="right" trigger="hover">
                <template #reference>
                  {{ formatText(systemInfo.firmwareVersion ?? constants.DEFAULT_EMPTY_TEXT, 15) }}
                </template>
                <div style="word-break: break-word;">
                  {{ systemInfo.firmwareVersion ?? constants.DEFAULT_EMPTY_TEXT }}
                </div>
              </el-popover>
            </div>
          </el-col>
        </el-row>
        <el-row class="sub-block">
          <el-col :span="12">
            <div class="sub-block-title">{{ $t("home.system.productAssetLabel") }}</div>
            <div class="sub-block-content">
              <el-popover placement="right" trigger="hover">
                <template #reference>
                  {{ formatText(systemInfo.assetLabel ?? constants.DEFAULT_EMPTY_TEXT, 15) }}
                </template>
                <div style="word-break: break-word;">
                  {{ systemInfo.assetLabel ?? constants.DEFAULT_EMPTY_TEXT }}
                </div>
              </el-popover>
            </div>
          </el-col>
          <el-col :span="12">
            <div class="sub-block-title">{{ $t("home.system.npuVersion") }}</div>
            <div class="sub-block-content">
              <el-popover placement="right" trigger="hover">
                <template #reference>
                  {{ formatText(systemInfo.npuVersion ?? constants.DEFAULT_EMPTY_TEXT, 15) }}
                </template>
                <div style="word-break: break-word;">
                  {{ systemInfo.npuVersion ?? constants.DEFAULT_EMPTY_TEXT }}
                </div>
              </el-popover>
            </div>
          </el-col>
        </el-row>
        <el-row class="sub-block">
          <el-col :span="12">
            <div class="sub-block-title">{{ $t("home.common.sn") }}</div>
            <div class="sub-block-content">
              <el-popover placement="right" trigger="hover">
                <template #reference>
                  {{ formatText(systemInfo.sn, 15) }}
                </template>
                <div style="word-break: break-word;">
                  {{ systemInfo.sn }}
                </div>
              </el-popover>
            </div>
          </el-col>
          <el-col :span="12" v-if="isA500()">
            <div class="sub-block-title">{{ $t("home.system.mcuVersion") }}</div>
            <div class="sub-block-content">
              <el-popover placement="right" trigger="hover">
                <template #reference>
                  {{ formatText(systemInfo.mcuVersion ?? constants.DEFAULT_EMPTY_TEXT, 15) }}
                </template>
                <div style="word-break: break-word;">
                  {{ systemInfo.mcuVersion ?? constants.DEFAULT_EMPTY_TEXT }}
                </div>
              </el-popover>
            </div>
          </el-col>
        </el-row>
      </div>
      <div class="block" style="height: 22vh;">
        <div class="sub-title">
          <img src="@/assets/img/home/information.svg" alt=""/>
          {{ $t("home.alarm.title") }}
        </div>
        <el-row class="alarm-sub-block" @click="clickToAlarm" style="cursor: pointer;">
          <el-col :span="12" class="alarm-col">
            <img src="@/assets/img/home/alarm-red.svg" alt=""/>
            <div class="sub-block-title">{{ $t("home.alarm.emergency") }}</div>
            <div class="sub-block-content">{{ alarmInfo.emergency }}</div>
          </el-col>
          <el-col :span="12" class="alarm-col">
            <img src="@/assets/img/home/alarm-orange.svg" alt=""/>
            <div class="sub-block-title">{{ $t("home.alarm.serious") }}</div>
            <div class="sub-block-content">{{ alarmInfo.serious }}</div>
          </el-col>
        </el-row>
        <el-row class="alarm-sub-block" @click="clickToAlarm" style="margin-top: 20px; cursor: pointer;">
          <el-col :span="12" class="alarm-col">
            <img src="@/assets/img/home/alarm-yellow.svg" alt=""/>
            <div class="sub-block-title">{{ $t("home.alarm.normal") }}</div>
            <div class="sub-block-content">{{ alarmInfo.normal }}</div>
          </el-col>
        </el-row>
      </div>
      <div>
        <div class="sub-title">
          <img src="@/assets/img/home/temperature.svg" alt=""/>
          {{ $t("home.temperature.title") }}
        </div>
        <el-row class="sub-block" v-if="isA500()">
          <el-col :span="12">
            <div class="sub-block-title">{{ $t("home.temperature.diskHeating") }}</div>
            <div class="sub-block-content">{{ temperatureInfo.diskHeating ?? constants.DEFAULT_EMPTY_TEXT }}</div>
          </el-col>
          <el-col :span="12">
            <div class="sub-block-title">{{ $t("home.temperature.boxTemperature") }}</div>
            <div class="sub-block-content">{{ temperatureInfo.box ?? constants.DEFAULT_EMPTY_TEXT }} ℃</div>
          </el-col>
        </el-row>
        <el-row class="sub-block">
          <el-col :span="12" v-if="isA500()">
            <div class="sub-block-title">{{ $t("home.temperature.cpuHeating") }}</div>
            <div class="sub-block-content">{{ temperatureInfo.cpuHeating ?? constants.DEFAULT_EMPTY_TEXT }}</div>
          </el-col>
          <el-col :span="12">
            <div class="sub-block-title">{{ $t("home.temperature.aiProcessorTemperature") }}</div>
            <div class="sub-block-content">{{ temperatureInfo.aiSpeeding ?? constants.DEFAULT_EMPTY_TEXT }} ℃</div>
          </el-col>
        </el-row>
        <el-row class="sub-block" v-if="isA500()">
          <el-col :span="12">
            <div class="sub-block-title">{{ $t("home.temperature.usbHubHeating") }}</div>
            <div class="sub-block-content">{{ temperatureInfo.usbHubHeating ?? constants.DEFAULT_EMPTY_TEXT }}</div>
          </el-col>
        </el-row>
      </div>
    </el-col>
    <el-col :span="14" style="padding: 28px;">
      <div class="title">{{ homeTitle }}</div>
      <div class="status-block">
        <span class="status-title">{{ $t("home.systemRunningStatus") }}</span>
        <span class="status-content" :class="{'error-status': $t('home.system.abnormal') === systemInfo.healthStatus }">
          {{ systemInfo.healthStatus }}
        </span>
        <span class="status-title" style="margin-left: 64px;">{{ $t("home.externalDevice.title") }}</span>
        <span class="status-content">{{ externalDevicesInfo.num }}</span>
        <el-button
          link
          type="primary"
          style="margin-left: 8px; cursor: pointer;"
          @click="clickToInformation"
        >
          {{ $t("common.seeMore") }}
        </el-button>
      </div>
      <div class="host">
        <img id="circle" src="@/assets/img/home/circle.svg">
        <img id="host" :src="$t('img.host')">
        <img id="model" src="/WhiteboxConfig/img/model.svg">
        <el-popover placement="right" trigger="hover" v-for="(device, index) in externalDevicesInfo.list" :key="device.Name" :width="400">
          <template #reference>
            <div class="device" :class="'device' + (index + 1)">
              <el-tag size="small" color="#3E4551" effect="dark" round style="margin-top: 80px;" type="info">{{ device.DeviceClass }}</el-tag>
              {{ formatText(device.DeviceName, 10) }}
            </div>
          </template>
          <div>
            <el-tag size="small" color="#3E4551" effect="dark" round type="info">{{ device.DeviceClass }}</el-tag>
            {{ formatText(device.DeviceName, 30) }}
          </div>
          <el-row class="popover-block">
            <el-col :span="8">
              <div class="sub-block-title">{{ $t("home.externalDevice.category") }}</div>
              <div class="sub-block-content">{{ device.DeviceClass ?? constants.DEFAULT_EMPTY_TEXT }}</div>
            </el-col>
            <el-col :span="8">
              <div class="sub-block-title">{{ $t("home.externalDevice.location") }}</div>
              <div class="sub-block-content">{{ device.Location ?? constants.DEFAULT_EMPTY_TEXT }}</div>
            </el-col>
            <el-col :span="8">
              <div class="sub-block-title">{{ $t("home.common.model") }}</div>
              <div class="sub-block-content">
                {{ formatText(device.Model, 10) }}
              </div>
            </el-col>
          </el-row>
          <el-row class="popover-block">
            <el-col :span="8">
              <div class="sub-block-title">{{ $t("home.common.sn") }}</div>
              <div class="sub-block-content">
                {{ formatText(device.SerialNumber, 10) }}
              </div>
            </el-col>
            <el-col :span="16">
              <div class="sub-block-title">{{ $t("home.common.vendor") }}</div>
              <div class="sub-block-content">{{ device.Manufacturer ?? constants.DEFAULT_EMPTY_TEXT }}</div>
            </el-col>
          </el-row>
        </el-popover>
      </div>
    </el-col>
    <el-col :span="5" style="padding: 28px;">
      <div class="block" style="height: 58vh;">
        <div class="sub-title">
          <img src="@/assets/img/home/unfold.svg" alt="" />
          {{ $t("home.stationInfo") }}
        </div>
        <el-row class="station-block">
          <el-col :span="12">
            <div class="chart">
              <el-progress type="dashboard" :percentage="stationInfo.cpuUtilizationRate.rate" :color="[{color: '#0077FF', percentage: 100}]">
                <template #default="{ percentage }">
                  <div class="percentage-label">{{ $t("home.common.utilization") }}</div>
                  <div class="percentage-value">{{ percentage }}%</div>
                </template>
              </el-progress>
            </div>
          </el-col>
          <el-col :span="12">
            <div class="station-sub-title">{{ $t("home.cpuUtilizationRate.title") }}</div>
            <div class="station-block-content">
              <span class="station-block-title">{{ $t("home.common.vendor") }}</span>
              <el-popover placement="left" trigger="hover">
                <template #reference>
                  {{ formatText(stationInfo.cpuUtilizationRate.vendor ?? constants.DEFAULT_EMPTY_TEXT, 15) }}
                </template>
                <div style="word-break: break-word;">
                  {{ stationInfo.cpuUtilizationRate.vendor ?? constants.DEFAULT_EMPTY_TEXT }}
                </div>
              </el-popover>
            </div>
            <div class="station-block-content">
              <span class="station-block-title">{{ $t("home.common.model") }}</span>
              <el-popover placement="left" trigger="hover">
                <template #reference>
                  {{ formatText(stationInfo.cpuUtilizationRate.model ?? constants.DEFAULT_EMPTY_TEXT, 15) }}
                </template>
                <div style="word-break: break-word;">
                  {{ stationInfo.cpuUtilizationRate.model ?? constants.DEFAULT_EMPTY_TEXT }}
                </div>
              </el-popover>
            </div>
            <div class="station-block-content">
              <span class="station-block-title">{{ $t("home.cpuUtilizationRate.coreNum") }}</span>
              <el-popover placement="left" trigger="hover">
                <template #reference>
                  {{ formatText(stationInfo.cpuUtilizationRate.coreNum ?? constants.DEFAULT_EMPTY_TEXT, 15) }}
                </template>
                <div style="word-break: break-word;">
                  {{ stationInfo.cpuUtilizationRate.coreNum ?? constants.DEFAULT_EMPTY_TEXT }}
                </div>
              </el-popover>
            </div>
          </el-col>
        </el-row>
        <el-row class="station-block">
          <el-col :span="12">
            <div class="chart">
              <el-progress type="dashboard" :percentage="stationInfo.ramUtilizationRate.rate" :color="[{color: '#0077FF', percentage: 100}]">
                <template #default="{ percentage }">
                  <div class="percentage-label">{{ $t("home.common.utilization") }}</div>
                  <div class="percentage-value">{{ percentage }}%</div>
                </template>
              </el-progress>
            </div>
          </el-col>
          <el-col :span="12">
            <div class="station-sub-title">{{ $t("home.ramUtilizationRate.title") }}</div>
            <div class="station-block-content">
              <span class="station-block-title">{{ $t("home.ramUtilizationRate.totalRam") }}</span>
              <el-popover placement="left" trigger="hover">
                <template #reference>
                  {{ formatText(stationInfo.ramUtilizationRate.totalRam ?? constants.DEFAULT_EMPTY_TEXT, 15) }} GB
                </template>
                <div style="word-break: break-word;">
                  {{ stationInfo.ramUtilizationRate.totalRam ?? constants.DEFAULT_EMPTY_TEXT }} GB
                </div>
              </el-popover>
            </div>
            <div class="station-block-content">
              <span class="station-block-title">{{ $t("home.ramUtilizationRate.usedRam") }}</span>
              <el-popover placement="left" trigger="hover">
                <template #reference>
                  {{ formatText(stationInfo.ramUtilizationRate.usedRam ?? constants.DEFAULT_EMPTY_TEXT, 15) }} GB
                </template>
                <div style="word-break: break-word;">
                  {{ stationInfo.ramUtilizationRate.usedRam ?? constants.DEFAULT_EMPTY_TEXT }} GB
                </div>
              </el-popover>
            </div>
          </el-col>
        </el-row>
        <el-row class="station-block">
          <el-col :span="12">
            <el-popover placement="left" trigger="hover" :width="360">
              <template #reference>
                <div class="chart">
                  <el-progress type="dashboard" :percentage="stationInfo.npuUtilizationRate.rate" :color="[{color: '#0077FF', percentage: 100}]">
                    <template #default="{ percentage }">
                      <div class="percentage-label">{{ $t("home.common.utilization") }}</div>
                      <div class="percentage-value">{{ percentage }}%</div>
                    </template>
                  </el-progress>
                </div>
              </template>
              <div>
                <div>{{ $t("home.npuUtilizationRate.title") }}</div>
                <el-row class="popover-block">
                  <el-col :span="12">
                    <div class="sub-block-title">{{ $t("home.npuUtilizationRate.aiCpu") }}</div>
                    <div class="sub-block-content">{{ stationInfo.npuUtilizationRate.aiCpu }}%</div>
                  </el-col>
                  <el-col :span="12">
                    <div class="sub-block-title">{{ $t("home.npuUtilizationRate.ctrlCpu") }}</div>
                    <div class="sub-block-content">{{ stationInfo.npuUtilizationRate.ctrlCpu }}%</div>
                  </el-col>
                </el-row>
                <el-row class="popover-block">
                  <el-col :span="12">
                    <div class="sub-block-title">{{ $t("home.npuUtilizationRate.ddrUsage") }}</div>
                    <div class="sub-block-content">{{ stationInfo.npuUtilizationRate.ddrUsage }}%</div>
                  </el-col>
                  <el-col :span="12">
                    <div class="sub-block-title">{{ $t("home.npuUtilizationRate.ddrBandWidth") }}</div>
                    <div class="sub-block-content">{{ stationInfo.npuUtilizationRate.ddrBandWidth }}%</div>
                  </el-col>
                </el-row>
              </div>
            </el-popover>
          </el-col>
          <el-col :span="12">
            <div class="station-sub-title">{{ $t("home.npuUtilizationRate.title") }}</div>
            <div class="station-block-content">
              <span class="station-block-title">{{ $t("home.common.vendor") }}</span>
              <el-popover placement="left" trigger="hover">
                <template #reference>
                  {{ formatText(stationInfo.npuUtilizationRate.vendor ?? constants.DEFAULT_EMPTY_TEXT, 15) }}
                </template>
                <div style="word-break: break-word;">
                  {{ stationInfo.npuUtilizationRate.vendor ?? constants.DEFAULT_EMPTY_TEXT }}
                </div>
              </el-popover>
            </div>
            <div class="station-block-content">
              <span class="station-block-title">{{ $t("home.common.model") }}</span>
              <el-popover placement="left" trigger="hover">
                <template #reference>
                  {{ formatText(stationInfo.npuUtilizationRate.model ?? constants.DEFAULT_EMPTY_TEXT, 15) }}
                </template>
                <div style="word-break: break-word;">
                  {{ stationInfo.npuUtilizationRate.model ?? constants.DEFAULT_EMPTY_TEXT }}
                </div>
              </el-popover>
            </div>
            <div class="station-block-content">
              <span class="station-block-title">{{ $t("home.npuUtilizationRate.computingAbility") }}</span>
              <el-popover placement="left" trigger="hover">
                <template #reference>
                  {{ formatText(stationInfo.npuUtilizationRate.computingAbility ?? constants.DEFAULT_EMPTY_TEXT, 15) }}
                </template>
                <div style="word-break: break-word;">
                  {{ stationInfo.npuUtilizationRate.computingAbility ?? constants.DEFAULT_EMPTY_TEXT }}
                </div>
              </el-popover>
            </div>
            <div class="station-block-content">
              <span class="station-block-title">{{ $t("home.npuUtilizationRate.ddrCapacity") }}</span>
              <el-popover placement="left" trigger="hover">
                <template #reference>
                  {{ formatText(stationInfo.npuUtilizationRate.ddrCapacity ?? constants.DEFAULT_EMPTY_TEXT, 15) }} GB
                </template>
                <div style="word-break: break-word;">
                  {{ stationInfo.npuUtilizationRate.ddrCapacity ?? constants.DEFAULT_EMPTY_TEXT }} GB
                </div>
              </el-popover>
            </div>
          </el-col>
        </el-row>
      </div>
      <div>
        <div class="sub-title">
          <img src="@/assets/img/home/train.svg" alt=""/>
          {{ $t("home.power.title") }}
        </div>
        <div>
          <el-row style="margin-bottom: 20px;">
            <el-col :span="18" class="power-title">{{ $t("home.power.currPower") }}</el-col>
            <el-col :span="6">
              <span class="power-value">{{ powerInfo.currPower }}</span>
              <span class="power-unit"> W</span>
            </el-col>
          </el-row>
          <el-progress :percentage="powerInfo.currPower" :show-text="false" :color="[{color: '#0077FF', percentage: 100}]" />
        </div>
      </div>
    </el-col>
  </el-row>
</template>

<script>
import { defineComponent, ref, onMounted, onBeforeUnmount } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import { querySystemsSourceInfo, queryCpuInfo, queryMemoryInfo, queryAiProcessorInfo, queryAllExtendedDevicesInfo } from '@/api/information';
import { queryAlarmSourceService } from '@/api/alarm';
import constants from '@/utils/constants';
import {
  getLanguageDefaultChinese, checkModuleType, isA500, isFulfilled, checkUrlsResponse
} from '@/utils/commonMethods';
import { getUrls } from '@/api/http';

export default defineComponent({
  setup() {
    const { t } = useI18n();
    let homeTitle = ref('');

    const systemInfo = ref({
      hostname: null,
      runtime: null,
      os: null,
      firmwareVersion: null,
      assetLabel: null,
      mcuVersion: null,
      sn: '',
      npuVersion: null,
      healthStatus: null,
    })

    const alarmInfo = ref({
      emergency: null,
      serious: null,
      normal: null,
    })

    const temperatureInfo = ref({
      box: null,
      aiSpeeding: null,
      cpuHeating: null,
      diskHeating: null,
      usbHubHeating: null,
    })

    const powerInfo = ref({
      currPower: null,
      maxPower: null,
    })

    const stationInfo = ref({
      cpuUtilizationRate: {
        vendor: null,
        model: null,
        coreNum: null,
        rate: null,
      },
      ramUtilizationRate: {
        totalRam: null,
        usedRam: null,
        rate: null,
      },
      npuUtilizationRate: {
        vendor: null,
        model: null,
        computingAbility: null,
        ddrCapacity: null,
        rate: null,
        aiCpu: null,
        ctrlCpu: null,
        ddrUsage: null,
        ddrBandWidth: null,
      },
    })

    const externalDevicesInfo = ref({
      num: 0,
      list: [],
    });

    const fetchSystemInfo = async (isShowLoading = false, AutoRefresh = false) => {
      let { data } = await querySystemsSourceInfo(isShowLoading, AutoRefresh);
      return data;
    }

    const initSystemInfo = (data) => {
      if (!data) {
        return;
      };
      systemInfo.value.hostname = data?.HostName;
      systemInfo.value.runtime = data?.Oem?.Uptime
      systemInfo.value.os = data?.Oem?.OSVersion;
      systemInfo.value.assetLabel = data?.AssetTag;
      systemInfo.value.sn = data?.SerialNumber ?? constants.DEFAULT_EMPTY_TEXT;
      systemInfo.value.healthStatus = {
        OK: t('home.system.normal'),
        error: t('home.system.abnormal'),
        'General warning': t('home.system.warning'),
        'Important warning': t('home.system.alarm'),
        'Urgent warning': t('home.system.critical'),
        'Unknown mistake': t('home.system.unknown'),
      }[data?.Status?.Health ?? 'OK'] ;
      stationInfo.value.cpuUtilizationRate.rate = data?.Oem?.CpuUsage ?? 0;

      let versions = data?.Oem?.Firmware;
      for (let i = 0; i < versions?.length; i++) {
        if (checkModuleType(versions[i].Module, 'mcu')) {
          systemInfo.value.mcuVersion = versions[i].Version || constants.DEFAULT_EMPTY_TEXT;
        } else if (checkModuleType(versions[i].Module, 'om') || checkModuleType(versions[i].Module, 'firmware')) {
          systemInfo.value.firmwareVersion = versions[i].Version || constants.DEFAULT_EMPTY_TEXT;
        }
      }
    }

    const fetchAlarmInfo = async (isShowLoading, AutoRefresh) => {
      let { data } = await queryAlarmSourceService(isShowLoading, AutoRefresh);
      return data;
    }

    const initAlarmInfo = (data) => {
      if (!data || !data.AlarMessages) {
        return;
      }
      let emergencyNum = 0;
      let seriousNum = 0;
      let normalNum = 0;

      data.AlarMessages.forEach(item => {
        if (item.PerceivedSeverity === '0') {
          emergencyNum++;
        } else if (item.PerceivedSeverity === '1') {
          seriousNum++;
        } else if (item.PerceivedSeverity === '2') {
          normalNum++;
        }
      })
      alarmInfo.value.emergency = emergencyNum;
      alarmInfo.value.serious = seriousNum;
      alarmInfo.value.normal = normalNum;
    }

    const initTemperatureInfo = (data) => {
      if (!data || !data.Oem) {
        return;
      }
      temperatureInfo.value.box = data.Oem.Temperature;
      temperatureInfo.value.aiSpeeding = data.Oem.AiTemperature;
      temperatureInfo.value.cpuHeating = data.Oem.CpuHeating;
      temperatureInfo.value.diskHeating = data.Oem.DiskHeating;
      temperatureInfo.value.usbHubHeating = data.Oem.UsbHubHeating;
    }

    const initPowerInfo = (data) => {
      if (!data || !data.Oem) {
        return;
      }
      powerInfo.value.currPower = data.Oem.Power ?? 0;
      powerInfo.value.maxPower = null;
    }

    const fetchCpuInfo = async (isShowLoading, AutoRefresh) => {
      let { data } = await queryCpuInfo(isShowLoading, AutoRefresh);
      return data;
    }

    const initStationCpuInfo = (data) => {
      if (!data || !data.Oem) {
        return;
      }
      stationInfo.value.cpuUtilizationRate.vendor = data.Manufacturer;
      stationInfo.value.cpuUtilizationRate.model = data.Oem.CPUModel;
      stationInfo.value.cpuUtilizationRate.coreNum = data.Oem.Count;
    }

    const fetchRamInfo = async (isShowLoading, AutoRefresh) => {
      let { data } = await queryMemoryInfo(isShowLoading, AutoRefresh);
      return data;
    }

    const initStationRamInfo = (systemData, ramData) => {
      if (!ramData || !ramData.Oem) {
        return;
      }
      stationInfo.value.ramUtilizationRate.totalRam = ramData.Oem.TotalSystemMemoryGiB;

      if (!systemData || !systemData.Oem) {
        return;
      }
      stationInfo.value.ramUtilizationRate.usedRam = (systemData.Oem.MemoryUsage * ramData.Oem.TotalSystemMemoryGiB / 100).toFixed(2);
      stationInfo.value.ramUtilizationRate.rate = systemData?.Oem.MemoryUsage ?? 0;
    }

    const fetchAiProcessorInfo = async (isShowLoading, AutoRefresh) => {
      let { data } = await queryAiProcessorInfo(isShowLoading, AutoRefresh);
      return data;
    }

    const initStationNpuInfo = (data) => {
      if (!data || !data.Oem) {
        return;
      }

      systemInfo.value.npuVersion = data?.NpuVersion || constants.DEFAULT_EMPTY_TEXT;
      stationInfo.value.npuUtilizationRate.vendor = stationInfo.value.cpuUtilizationRate.vendor;
      stationInfo.value.npuUtilizationRate.model = data.Model ?? constants.DEFAULT_EMPTY_TEXT;
      stationInfo.value.npuUtilizationRate.computingAbility = data.Oem?.Capability.Calc;
      stationInfo.value.npuUtilizationRate.ddrCapacity = data.Oem?.Capability.Ddr;
      stationInfo.value.npuUtilizationRate.rate = data?.Oem?.OccupancyRate.AiCore ?? 0;
      stationInfo.value.npuUtilizationRate.aiCpu = data?.Oem?.OccupancyRate.AiCpu ?? 0;
      stationInfo.value.npuUtilizationRate.ctrlCpu = data?.Oem?.OccupancyRate.CtrlCpu ?? 0;
      stationInfo.value.npuUtilizationRate.ddrUsage = data?.Oem?.OccupancyRate.DdrUsage ?? 0;
      stationInfo.value.npuUtilizationRate.ddrBandWidth = data?.Oem?.OccupancyRate.DdrBandWidth ?? 0;
    }

    const fetchExternalDevice = async (isShowLoading, AutoRefresh) => {
      let { data } = await queryAllExtendedDevicesInfo(isShowLoading, AutoRefresh);
      return data;
    }

    const initExternalDevicesInfo = async (deviceData, isShowLoading, AutoRefresh) => {
      if (!deviceData) {
        return;
      }
      externalDevicesInfo.value.num = deviceData['Members@odata.count'] ?? 0;
      if (!deviceData.Members) {
        return;
      }

      let odataList = deviceData.Members.length > 10 ? deviceData.Members.slice(0, 10) : deviceData.Members.slice(0);

      let params = odataList.map(item => (
        {
          url: item['@odata.id'],
          isShowLoading,
          AutoRefresh,
        }
      ))
      let allResponse = await getUrls(params);
      externalDevicesInfo.value.list = allResponse.filter(item => isFulfilled(item)).map(item => item.value?.data);
      checkUrlsResponse(allResponse, AutoRefresh)
    }

    const init = async (isShowLoading = false, AutoRefresh = false) => {
      let urls = [
        '/WhiteboxConfig/json/config.json',
        '/redfish/v1/Systems',
        '/redfish/v1/Systems/Alarm/AlarmInfo',
        '/redfish/v1/Systems/Processors/CPU',
        '/redfish/v1/Systems/Memory',
        '/redfish/v1/Systems/Processors/AiProcessor',
        '/redfish/v1/Systems/ExtendedDevices',
      ]

      let params = urls.map(url => (
        {
          url,
          isShowLoading,
          AutoRefresh,
        }
      ))

      let allResponse = await getUrls(params);
      checkUrlsResponse(allResponse, AutoRefresh);

      homeTitle.value = allResponse[0].value?.data.websiteTitle[getLanguageDefaultChinese()];

      if (isFulfilled(allResponse[1])) {
        let systemInfoData = allResponse[1].value?.data;
        initSystemInfo(systemInfoData);
        initTemperatureInfo(systemInfoData);
        initPowerInfo(systemInfoData);
      }

      if (isFulfilled(allResponse[2])) {
        let alarmInfoData = allResponse[2].value?.data;
        initAlarmInfo(alarmInfoData);
      }

      if (isFulfilled(allResponse[3])) {
        let cpuInfo = allResponse[3].value?.data;
        initStationCpuInfo(cpuInfo);
      }

      if (isFulfilled(allResponse[1]) && isFulfilled(allResponse[4])) {
        let ramInfo = allResponse[4].value?.data;
        initStationRamInfo(allResponse[1].value?.data, ramInfo);
      }

      if (isFulfilled(allResponse[5])) {
        let aiProcessorInfo = allResponse[5].value?.data;
        initStationNpuInfo(aiProcessorInfo);
      }

      if (isFulfilled(allResponse[6])) {
        let externalDeviceInfo = allResponse[6].value?.data;
        await initExternalDevicesInfo(externalDeviceInfo, isShowLoading, AutoRefresh);
      }
    }

    onMounted(async () => {
      startRefreshTimer();
      await init();
    })

    let autoRefreshTimer;

    const startRefreshTimer = () => {
      autoRefreshTimer = setInterval(() => {
        init(false, true)
      }
      , constants.AUTO_REFRESH);
    }

    const stopRefreshTimer = () => {
      clearInterval(autoRefreshTimer);
      autoRefreshTimer = null;
    }

    onBeforeUnmount(() => {
      // 清除定时器
      stopRefreshTimer();
    });

    const router = useRouter();

    const clickToInformation = () => {
      router.push('/manager/information');
    }

    const clickToAlarm = () => {
      router.push('/manager/alarm');
    }

    const formatText = (text, length) => text?.length > length ? text.slice(0, length) + '...' : text

    return {
      constants,
      homeTitle,
      systemInfo,
      alarmInfo,
      temperatureInfo,
      powerInfo,
      stationInfo,
      externalDevicesInfo,
      clickToInformation,
      clickToAlarm,
      formatText,
      isA500,
    }
  },
});
</script>

<style lang="scss" scoped>
.sub-title {
  display: flex;
  align-items: center;
  font-size: 16px;
  color: var(--el-text-color-regular);
  letter-spacing: 0;
  line-height: 24px;
  font-weight: 400;
  margin-bottom: 24px;
}

.sub-title > img {
  margin-right: 8px;
  width: 16px;
  height: 16px;
}

.block {
  border-bottom: 1px solid var(--el-border-color);
  margin-bottom: 20px;
}

.title {
  font-size: 32px;
  letter-spacing: 0.89px;
  font-weight: 400;
  text-align: center;
}

.sub-block > div {
  border-left: 1px solid var(--el-border-color);
  padding-left: 14px;
  margin-bottom: 32px;
}

.popover-block {
  border-left: 1px solid var(--el-border-color);
  padding-left: 14px;
  margin-top: 20px;
}

.sub-block-title {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  letter-spacing: 0;
  line-height: 16px;
  font-weight: 200;
  font-family: 'HarmonyOS_Sans_SC_Light', serif;
}

.sub-block-content {
  font-size: 14px;
  color: var(--el-text-color-regular);
  letter-spacing: 0;
  line-height: 16px;
  font-weight: 400;
  margin-top: 8px;
}

.alarm-sub-block {
  margin-bottom: 32px;
}

.alarm-col {
  border-left: 1px solid var(--el-border-color);
  padding-left: 12px;
}

.alarm-sub-block img {
  float: left;
  margin-right: 10px;
  width: 32px;
  height: 32px;
  margin-top: 4px;
}

.station-block {
  padding-left: 14px;
  margin-bottom: 52px;
}

.station-sub-title {
  font-size: 16px;
  letter-spacing: 0;
  line-height: 24px;
  font-weight: 400;
  margin-bottom: 24px;
}

.station-block-title {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  letter-spacing: 0;
  line-height: 16px;
  font-weight: 200;
  margin-right: 8px;
  font-family: 'HarmonyOS_Sans_SC_Light', serif;
}

.station-block-content {
  font-size: 14px;
  color: var(--el-text-color-regular);
  letter-spacing: 0;
  line-height: 16px;
  font-weight: 400;
  margin-top: 8px;
}

.status-block {
  margin-top: 64px;
  text-align: center;
}

.status-title {
  font-size: 18px;
  color: var(--el-text-color-secondary);
  letter-spacing: 0;
  line-height: 24px;
  font-weight: 400;
}

.status-content {
  font-size: 18px;
  color: var(--el-text-color-regular);
  letter-spacing: 0;
  line-height: 24px;
  font-weight: 500;
  margin-left: 16px;
}

.error-status {
  color: var(--el-color-danger);
}

.chart {
  height: 12vh;
  width: 12vw;
}

.percentage-label {
  font-size: 8px;
  color: var(--el-text-color-secondary);
  text-align: center;
  line-height: 12px;
  font-weight: 400;
}

.percentage-value {
  font-size: 16px;
  color: var(--el-text-color-regular);
  text-align: center;
  line-height: 20px;
  font-weight: 400;
  margin-top: 8px;
}

.power-title {
  font-size: 16px;
  color: #D3DCE9;
  line-height: 24px;
  font-weight: 400;
  margin-top: 10px;
}

.power-value {
  font-size: 40px;
  color: var(--el-text-color-regular);
  text-align: right;
  line-height: 48px;
  font-weight: 400;
}

.power-unit {
  font-size: 12px;
  color: #D3DCE9;
  text-align: right;
  line-height: 16px;
  font-weight: 400;
}

.host {
  width: 70%;
  height: 60%;
  position: absolute;
  left: calc(50vw - 35%);
}

#circle {
  position: absolute;
  width: 1000px;
  height: 800px;
  left: calc(50% - 500px);
}

#host {
  position: absolute;
  width: 700px;
  height: 600px;
  left: calc(50% - 350px);
  z-index: 1;
}

#model {
  position: absolute;
  width: 56px;
  height: 10px;
  top: 430px;
  left: calc(51% - 37px);
  z-index: 2;
}

.device {
  position: absolute;
  background-image: url("@/assets/img/home/device.svg"), url("@/assets/img/home/device-shadow.svg");
  background-repeat: no-repeat, no-repeat;
  background-position: 0% 0%, 100% 100%;
  background-size: 115px 100px;
  width: 120px;
  height: 120px;
  cursor: pointer;
  z-index: 1000;
  font-size: 12px;
}

.device1 {
  left: 29%;
  top: 29%;
}

.device2 {
  left: 23%;
  top: 41%;
}

.device3 {
  left: 25%;
  top: 55%;
}

.device4 {
  left: 28%;
  top: 68%;
}

.device5 {
  left: 39%;
  top: 74%;
}

.device6 {
  left: 51%;
  top: 74%;
}

.device7 {
  left: 63%;
  top: 68%;
}

.device8 {
  left: 69%;
  top: 55%;
}

.device9 {
  left: 67%;
  top: 41%;
}

.device10 {
  left: 62%;
  top: 31%;
}

</style>