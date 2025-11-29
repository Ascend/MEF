#!/bin/bash
# Copyright(C) Huawei Technologies Co.,Ltd. 2022. All rights reserved.
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
