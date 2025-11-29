#[==[
#Provides the following variables:

#  Cmscbb_INCLUDE_DIRS - the Cmscbb include directories
#  Cmscbb_LIBRARIES - link these to use Cmscbb

#The following components are supported:

#  * `cmscbb`

#Note that only components requested with `COMPONENTS` or `OPTIONAL_COMPONENTS`
#are guaranteed to set these variables or provide targets.
#]==]

set(Cmscbb_PATH ${CMAKE_CURRENT_LIST_DIR}/../../platform/cmscbb_csec/)

if (NOT DEFINED Cmscbb_PATH)
  message(FATAL_ERROR "please define environment variable:Cmscbb_PATH")
endif()

set(Cmscbb_INCLUDE_DIRS ${Cmscbb_PATH}/include)
set(Cmscbb_INCLUDE_SRCS ${Cmscbb_PATH}/src)