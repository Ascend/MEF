#!/bin/bash
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
set -ex

# 设置npm源
npm config set registry https://cmc.centralrepo.rnd.huawei.com/artifactory/api/npm/npm-central-repo/

# 回到项目根目录
cd ..

# 安装依赖
npm ci

# 构建
npm run build

# 创建output文件夹
mkdir output

# 打包
zip -vr mindxom-web.zip dist

# 移动
mv mindxom-web.zip ./output/
