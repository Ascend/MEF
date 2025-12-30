#!/bin/bash

CUR_DIR=$(dirname "$(readlink -f "$0")")
Dev_list="dev_out_list.txt"

function restore_one() {
  local file_info=$1
  local info_arr dir_path path st_type major_id minor_id st_mode st_uid st_gid
  read -r -a info_arr <<<"$file_info"
  path=$CUR_DIR/${info_arr[0]}
  st_type=${info_arr[1]}
  major_id=${info_arr[2]}
  minor_id=${info_arr[3]}
  st_mode=${info_arr[4]}
  st_uid=${info_arr[5]}
  st_gid=${info_arr[6]}

  if [ -z "$path" ] ||
    [ -z "$st_type" ] ||
    [ -z "$major_id" ] ||
    [ -z "$minor_id" ] ||
    [ -z "$st_mode" ] ||
    [ -z "$st_uid" ] ||
    [ -z "$st_gid" ]; then
    echo "the input file info not valid, so skipped this info "
    return 1
  fi

  dir_path="$(dirname "$path")"
  mkdir -p "$dir_path"
  rm -f "$path"
  mknod "$path" "$st_type" "$major_id" "$minor_id" -m "$st_mode"
  chown "$st_uid":"$st_gid" "$path"

}

function unmount() {
  for ((i = 0; i < 120; i++)); do
    umount --recursive /dev && break
    echo "umount failed, retry $i times ..."
    sleep 1
  done
}

function start_docker() {
  /usr/bin/dockerd "$@"
  return $?
}

function restore_dev_file() {
  local one_info_path
  if [ ! -f "$Dev_list" ]; then
    echo "the dev into file [$Dev_list] not exists"
    return 1
  fi

  while IFS= read -r one_info_path; do
    echo "input file info=$one_info_path"
    restore_one "$one_info_path"
  done <"$Dev_list"
  return 0

}

function main() {
  unmount
  restore_dev_file
  start_docker "$@"
  return $?
}

main "$@"
exit $?
