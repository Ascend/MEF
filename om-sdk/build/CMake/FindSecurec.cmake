#[==[
#Provides the following variables:

#  Securec_INCLUDE_DIRS - the Securec include directories
#  Securec_LIBRARIES - link these to use Securec

#The following components are supported:

#  * `securec`

#Note that only components requested with `COMPONENTS` or `OPTIONAL_COMPONENTS`
#are guaranteed to set these variables or provide targets.
#]==]

set(Securec_PATH ${CMAKE_CURRENT_LIST_DIR}/../../platform/cpp/secure)

if (NOT DEFINED Securec_PATH)
  message(FATAL_ERROR "please define environment variable:Securec_PATH")
endif()

set(Securec_INCLUDE_DIRS ${Securec_PATH}/include)

set(components securec)

foreach(component ${components})

  set(Securec_LIB_PATH PATHS ${Securec_PATH} NO_SYSTEM_ENVIRONMENT_PATH NO_CMAKE_ENVIRONMENT_PATH NO_CMAKE_PACKAGE_REGISTRY NO_CMAKE_FIND_ROOT_PATH)
  set(Securec_LIB_SUFFIXES "lib")
  find_library(${component}_library NAMES ${component} NAMES_PER_DIR ${Securec_LIB_PATH} PATH_SUFFIXES ${Securec_LIB_SUFFIXES} NO_CMAKE_SYSTEM_PATH)
  if(${component}_library)
    list(APPEND Securec_LIBRARIES ${${component}_library})
  endif()

endforeach()
