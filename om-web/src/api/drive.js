/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
   MindEdge is licensed under Mulan PSL v2.
   You can use this software according to the terms and conditions of the Mulan PSL v2.
   You may obtain a copy of Mulan PSL v2 at:
            http://license.coscl.org.cn/MulanPSL2
   THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
   EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
   MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
   See the Mulan PSL v2 for more details.
 */

import { $get, $post, $patch, $delete } from '@/api/http';

export function queryAllSimpleStoragesInfo(isShowLoading = true) {
  // 功能描述：查询简单存储集合信息
  const url = '/redfish/v1/Systems/SimpleStorages'
  return $get(url,
    {},
    {
      customParams: { isShowLoading },
    })
}

export function queryAllPartitionsInfo(isShowLoading = true) {
  // 功能描述：查询磁盘分区集合信息
  const url = '/redfish/v1/Systems/Partitions'
  return $get(url,
    {},
    {
      customParams: { isShowLoading },
    })
}

export function createPartitions(params) {
  // 功能描述：查询磁盘分区集合信息
  const url = '/redfish/v1/Systems/Partitions'
  return $post(url, {
    ...params,
  },
  {
    timeout: 10 * 60 * 1000,
  });
}

export function deletePartition(partitionId) {
  // 功能描述：删除磁盘分区
  const url = `/redfish/v1/Systems/Partitions/${partitionId}`;
  return $delete(url);
}

export function mountPartitions(params) {
  // 功能描述：挂载/解挂磁盘分区
  const url = '/redfish/v1/Systems/Partitions/Mount'
  return $patch(url, {
    ...params,
  });
}

export function unmountPartitions(params) {
  // 功能描述：挂载/解挂磁盘分区
  const url = '/redfish/v1/Systems/Partitions/Unmount'
  return $patch(url, {
    ...params,
  });
}

export function queryNfsInfo(isShowLoading = true, AutoRefresh = false) {
  // 功能描述：查询NFS分区信息
  const url = '/redfish/v1/Systems/NfsManage'
  return $get(url,
    {},
    {
      customParams: { isShowLoading },
      headers: { AutoRefresh },
    })
}

export function mountNfs(params, isShowLoading = true) {
  // 功能描述：挂载NFS分区
  const url = '/redfish/v1/Systems/NfsManage/Actions/NfsManage.Mount'
  return $post(url, {
    ...params,
  },
  {
    customParams: { isShowLoading },
  });
}

export function unmountNfs(params, isShowLoading = true) {
  // 功能描述：解挂NFS分区
  const url = '/redfish/v1/Systems/NfsManage/Actions/NfsManage.Unmount'
  return $post(url, {
    ...params,
  },
  {
    customParams: { isShowLoading },
  });
}
