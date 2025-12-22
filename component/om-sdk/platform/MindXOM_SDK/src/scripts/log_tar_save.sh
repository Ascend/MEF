#!/bin/bash

#***********************************************
#  Description: 日志转储
#    Parameter:
#        Input: $1 -- 日志源目录
#        Input: $2 -- 日志名称
#        Input: $3 -- 日志转储目录
#        Input: $4 -- 转储目录允许的最大文件数量
#       Output: NA
#***********************************************

# 添加环境变量
OM_WORK_DIR="/usr/local/mindx/MindXOM"
export LD_LIBRARY_PATH="${OM_WORK_DIR}"/lib:"${LD_LIBRARY_PATH}"
export PYTHONPATH="${OM_WORK_DIR}"/software/ibma:"${OM_WORK_DIR}"/software/ibma/opensource/python:"${OM_WORK_DIR}"/scripts/python
# 执行日志转储
python3 -u "${OM_WORK_DIR}"/scripts/python/log_tar_save.py "$@"
exit $?

