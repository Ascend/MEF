<template>
  <div v-loading="loading">
    <div class="info-block">
      <span class="tab-title" style="margin-bottom: 0;">
        {{ $t("disk.localDiskManagement.driveList") }}
      </span>
      <img class="info-img" alt="" src="@/assets/img/manager/abnormal.svg" />
      <span class="disk-tip">
        {{ $t("disk.localDiskManagement.diskListTip") }}
      </span>
    </div>
    <div class="data-list">
      <div class="refresh">
        <el-button style="margin-left: 10px; width: 32px; height: 32px" @click="refreshWholeTable">
          <img
            style="width: 14px; height: 14px"
            alt=""
            src="@/assets/img/common/refresh.svg"
          />
        </el-button>
      </div>
      <el-table
        :data="diskData"
        style="width: 97%; height: 86%;"
        row-key="driveLetter"
        @expand-change="initPartitionTable"
      >
        <el-table-column type="expand">
          <template #default="props">
            <el-table :data="props.row.partitions">
              <el-table-column
                :label="$t('disk.localDiskManagement.partitionName')"
                prop="partitionName"
              />
              <el-table-column
                :label="$t('disk.localDiskManagement.systemPartition')"
                prop="systemPartition"
              />
              <el-table-column
                :label="$t('disk.common.capacityBytes')"
                prop="capacityBytes"
              />
              <el-table-column
                :label="$t('disk.localDiskManagement.fileSystem')"
                prop="fileSystem"
              />
              <el-table-column
                :label="$t('disk.common.freeBytes')"
                prop="freeBytes"
              />
              <el-table-column
                :label="$t('disk.common.mountPath')"
                prop="mountPath"
              />
              <el-table-column prop="address" :label="$t('common.operations')">
                <template #default="scope">
                  <el-button link type="primary" size="small" @click="clickMountPartition(scope.row, props.$index)">
                    {{ scope.row.mountPath ? $t("disk.common.unmount") : $t("disk.localDiskManagement.mount") }}
                  </el-button>
                  <el-button
                    link
                    type="primary"
                    size="small"
                    :disabled="isEmmc(props.row)"
                    @click="clickDeletePartition(scope.row, props.$index)"
                    >{{ $t("common.delete") }}</el-button
                  >
                </template>
              </el-table-column>
            </el-table>
          </template>
        </el-table-column>
        <el-table-column
          :label="$t('disk.localDiskManagement.driveLetter')"
          prop="driveLetter"
        />
        <el-table-column
          :label="$t('disk.localDiskManagement.deviceType')"
          prop="deviceType"
        />
        <el-table-column
          :label="$t('disk.common.capacityBytes')"
          prop="capacityBytes"
        />
        <el-table-column
          :label="$t('disk.localDiskManagement.interfaceType')"
          prop="interfaceType"
        />
        <el-table-column
          :label="$t('disk.localDiskManagement.location')"
          prop="location"
        />
        <el-table-column
          :label="$t('disk.common.healthStatus')"
          prop="healthStatus"
        >
          <template #default="scope">
            <el-tag effect="dark" v-if="scope.row.healthStatus === 'OK'" type="success">
              {{ $t("disk.common.normal") }}
            </el-tag>
            <el-tag effect="dark" v-else type="danger">{{ $t("disk.common.abnormal") }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="$t('common.operations')">
          <template #default="scope">
            <el-button
              link
              type="primary"
              size="small"
              :disabled="isEmmc(scope.row)"
              @click="clickCreatePartition(scope.row, scope.$index)"
            >
              {{ $t("disk.localDiskManagement.createPartition") }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination
        layout="total, sizes, prev, pager, next"
        :current-page="pagination.pageNum"
        :total="initData?.length"
        :page-size="pagination.pageSize"
        :page-sizes="[1, 5, 10]"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
        style="position: absolute; bottom: 20px;"
      />
    </div>
  </div>
  <create-partition
    v-if="createPartitionData.isShow"
    :create-partition-data="createPartitionData"
    @cancelCreatePartition="cancelCreatePartition"
    @createPartitionSuccessfully="createPartitionSuccessfully"
  />
  <delete-partition
    :delete-partition-data="deletePartitionData"
    @cancelDeletePartition="cancelDeletePartition"
    @deletePartitionSuccessfully="deletePartitionSuccessfully"
  />
  <mount-partition
    :mount-partition-data="mountPartitionData"
    @cancelMountAndUnmountPartition="cancelMountAndUnmountPartition"
    @mountPartitionSuccessfully="mountPartitionSuccessfully"
  />
</template>

<script>
import { defineComponent, ref, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { AppMessage, showErrorAlert, convertToGB, isFulfilled, checkUrlsResponse } from '@/utils/commonMethods';

import {
  queryAllPartitionsInfo,
  queryAllSimpleStoragesInfo
} from '@/api/drive';
import { getUrls } from '@/api/http';
import CreatePartition from '@/views/manager/disk/CreatePartition.vue';
import DeletePartition from '@/views/manager/disk/DeletePartition.vue';
import MountPartition from '@/views/manager/disk/MountPartition.vue';

export default defineComponent({
  name: 'LocalDisk',
  components: {
    CreatePartition,
    DeletePartition,
    MountPartition,
  },
  setup() {
    const { t } = useI18n();
    let diskData = ref([]);
    let initData = ref([]);
    const expandedIndexOfDiskData = ref(-1);
    const pagination = ref({
      pageNum: 1,
      pageSize: 5,
    })
    const createPartitionData = ref({
      isShow: false,
      currDisk: null,
      diskIndex: null,
    });

    const deletePartitionData = ref({
      isShow: false,
      partitionId: '',
      diskIndex: null,
    });

    const mountPartitionData = ref({
      isShow: false,
      partitionId: '',
      location: '',
      diskIndex: null,
    });

    const clickDeletePartition = (row, diskIndex) => {
      deletePartitionData.value = {
        isShow: true,
        partitionId: row.partitionName,
        diskIndex,
      };
    };

    const clickMountPartition = (row, diskIndex) => {
      mountPartitionData.value = {
        isShow: true,
        partitionId: row.partitionName,
        mountPath: row.mountPath,
        diskIndex,
      };
    }

    const clickCreatePartition = (row, index) => {
      createPartitionData.value = {
        isShow: true,
        currDisk: row,
        diskIndex: index,
      };
    };

    const isEmmc = (row) => row?.driveLetter === '/dev/mmcblk0'

    const cancelCreatePartition = () => {
      createPartitionData.value = {
        isShow: false,
        currDisk: null,
        diskIndex: null,
      };
    };

    const cancelDeletePartition = () => {
      deletePartitionData.value = {
        isShow: false,
        partitionId: null,
        driveLetter: null,
        diskIndex: null,
      };
    };

    const cancelMountAndUnmountPartition = () => {
      mountPartitionData.value = {
        isShow: false,
        partitionId: null,
        location: null,
        diskIndex: null,
      };
    }

    const updateExpandedRowPartitions = async (expandedDiskRow) => {
      if (!expandedDiskRow) {
        return;
      }
      expandedDiskRow.partitions = await fetchPartitionsDetailByExpandedIndex(
        expandedDiskRow
      );
    }

    const createPartitionSuccessfully = async () => {
      cancelCreatePartition();
      await refreshWholeTable();
      AppMessage.success(t('common.saveSuccessfully'));
    };

    const deletePartitionSuccessfully = async () => {
      cancelDeletePartition();
      await refreshWholeTable();
      AppMessage.success(
        t('disk.localDiskManagement.deletePartitionTip.deleteSuccess', {
          partition: deletePartitionData.value.partitionId,
        })
      );
    };

    const mountPartitionSuccessfully = async (operate) => {
      let diskIndex = mountPartitionData.value.diskIndex
      cancelMountAndUnmountPartition();
      await updateExpandedRowPartitions(diskData.value[diskIndex]);
      let operateMapper = {
        'mount': 'disk.localDiskManagement.mountPartitionTip.mountSuccess',
        'unmount': 'disk.localDiskManagement.mountPartitionTip.unmountSuccess',
      }
      AppMessage.success(
        t(operateMapper[operate], {
          partition: mountPartitionData.value.partitionId,
        })
      );
    }

    const loading = ref(false);
    const initDiskTable = async () => {
      loading.value = true;
      let devices = [];
      try {
        let { data } = await queryAllSimpleStoragesInfo(false);
        let params = data?.Members.map(item => (
          {
            url: item['@odata.id'],
            isShowLoading: false,
          }
        ))

        let allResponse = await getUrls(params);
        allResponse.filter(item => isFulfilled(item)).map(item => {
          let { data: singleRes } = item.value;
          if (!singleRes) {
            return;
          }

          for (let i = 0; i < singleRes?.Devices?.length; i++) {
            devices.push({
              id: singleRes?.Id,
              driveLetter: singleRes?.Devices[i].Name,
              deviceType: singleRes?.Devices[i].Model,
              capacityBytes: convertToGB(singleRes?.Devices[i].CapacityBytes),
              interfaceType: singleRes?.Name,
              location: singleRes?.Devices[i].Location ?? '',
              healthStatus: singleRes?.Devices[i].Status.Health,
              leftBytes: convertToGB(singleRes?.Devices[i].LeftBytes),
              partitions: [],
            });
          }
        })
        checkUrlsResponse(allResponse);
      } catch (e) {
        showErrorAlert(t('common.requestError'))
      } finally {
        loading.value = false;
      }

      return devices;
    };

    const fetchPartitionsDetailByExpandedIndex = async (expandedRow) => {
      // fetch all partition url and filter by device letter
      loading.value = true;
      let partitionsDetail = [];
      try {
        let { data } = await queryAllPartitionsInfo(false);
        let driveLetter = expandedRow.driveLetter;
        let params = [];
        const deviceLetter = driveLetter?.split('/dev/')[1];
        for (let i = 0; i < data?.Members?.length; i++) {
          if (data.Members[i]['@odata.id'].indexOf(deviceLetter) > -1) {
            params.push({
              url: data.Members[i]['@odata.id'],
              isShowLoading: false,
            });
          }
        }

        // fetch partition details
        let allResponse = await getUrls(params);
        allResponse.filter(item => isFulfilled(item)).map(item => {
          let { data: partitionRes } = item.value;
          if (!partitionRes) {
            return;
          }
          partitionsDetail.push({
            partitionName: partitionRes?.Name,
            systemPartition: partitionRes?.Primary,
            capacityBytes: convertToGB(partitionRes?.CapacityBytes),
            fileSystem: partitionRes?.FileSystem,
            freeBytes: convertToGB(partitionRes?.FreeBytes),
            mountPath: partitionRes?.MountPath,
          });
        })
        checkUrlsResponse(allResponse);
      } catch (e) {
        showErrorAlert(t('common.requestError'))
      } finally {
        loading.value = false;
      }

      return partitionsDetail;
    };

    let expandedRow = [];
    const initPartitionTable = async (row, expandedRows) => {
      let rows = [];
      for (let i = 0; i < expandedRows?.length; i++) {
        for (let j = 0; j < diskData.value?.length; j++) {
          if (diskData.value[j].driveLetter === expandedRows[i].driveLetter) {
            if (expandedRow.indexOf(j) === -1) {
              await updateExpandedRowPartitions(diskData.value[j]);
            }
            rows.push(j)
          }
        }
      }

      expandedRow = rows
    };

    const refreshWholeTable = async () => {
      initData.value = await initDiskTable();
      diskData.value = initData.value.slice(
        pagination.value.pageSize * (pagination.value.pageNum - 1),
        pagination.value.pageSize * pagination.value.pageNum
      );

      for (let j = 0; j < expandedRow?.length; j++) {
        await updateExpandedRowPartitions(diskData.value[expandedRow[j]]);
      }
    }

    const handleSizeChange = (value) => {
      pagination.value.pageSize = value
      diskData.value = initData.value.slice(
        pagination.value.pageSize * (pagination.value.pageNum - 1),
        pagination.value.pageSize * pagination.value.pageNum
      );
    }

    const handleCurrentChange = (value) => {
      pagination.value.pageNum = value
      diskData.value = initData.value.slice(
        pagination.value.pageSize * (pagination.value.pageNum - 1),
        pagination.value.pageSize * pagination.value.pageNum
      );
    }

    onMounted(async () => {
      initData.value = await initDiskTable();
      handleCurrentChange(pagination.value.pageNum);
    });

    return {
      createPartitionData,
      deletePartitionData,
      mountPartitionData,
      diskData,
      initData,
      pagination,
      loading,
      cancelCreatePartition,
      clickCreatePartition,
      isEmmc,
      createPartitionSuccessfully,
      clickDeletePartition,
      cancelDeletePartition,
      deletePartitionSuccessfully,
      clickMountPartition,
      mountPartitionSuccessfully,
      cancelMountAndUnmountPartition,
      initPartitionTable,
      refreshWholeTable,
      handleSizeChange,
      handleCurrentChange,
    };
  },
});
</script>

<style lang="scss" scoped>
.info-block {
  display: flex;
  align-items: center;
}

.disk-tip {
  font-size: 12px;
  color: #d3dce9;
  line-height: 16px;
  font-weight: 400;
}

.info-img {
  margin-left: 26px;
  margin-right: 10px;
}

.data-list {
  width: 90%;
  overflow: auto;
  height: 74vh;
  background-color: var(--el-bg-color-overlay);
  margin-top: 20px;
  padding: 24px;
  border-radius: 4px;
}

.refresh {
  width: 6%;
  float: right;
  margin-bottom: 20px;
  display: flex;
  align-items: center;
}
</style>
