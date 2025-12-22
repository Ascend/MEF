#!/bin/bash

ENS_HOME=$(dirname "$(dirname "$(readlink -f "$0")")")
export ENS_HOME=${ENS_HOME}
export LD_LIBRARY_PATH=${ENS_HOME}/lib:${OM_WORK_DIR}/lib:/usr/local/lib:/usr/lib64/:${LD_LIBRARY_PATH}
export PATH=${ENS_HOME}/bin:$PATH
ensd "$@" &
