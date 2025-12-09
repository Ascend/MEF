#! /bin/bash
OM_WORK_DIR="/usr/local/mindx/MindXOM"

source "/home/data/config/os_cmd.conf"
source "${OM_WORK_DIR}/scripts/log_print.sh"

init_log_path "/var/plog/manager" "manager_run.log"

source "${OM_WORK_DIR}/software/ibma/lib/Linux/systems/disk/common_collect_log.sh"

LOG_PATH="/run/collect_log"
if [ -d "${LOG_PATH}" ]; then
    rm -rf "${LOG_PATH}"
fi
mkdir -p "${LOG_PATH}"

if [ -z "$1" ] ;then
    logger_error "input error!"
    exit 1
fi

while true
do
    case "$1" in
        "NPU")
            npu_log
            shift
            ;;
        "MindXOM")
            MindXOM_log
            shift
            ;;
        "OSDivers")
            OS_log
            shift
            ;;
        "MEF")
            MEF_log "all"
            shift
            ;;
        *)
            break
            ;;
    esac
done

#get pack
log_name="log_collect.tar.gz"
if [ -f "/run/${log_name}" ];then
    rm -f "/run/${log_name}"
fi

touch "/run/${log_name}"
chmod 644 "/run/${log_name}"

if ! tar -czPf "/run/${log_name}" -C /run/ collect_log; then
    echo "Collect ${log_name} failed"
    rm -rf "${LOG_PATH}"
    exit 1
else
    echo "Collect ${log_name} successfully"
    rm -rf "${LOG_PATH}"
fi
exit 0
