package parachecker

import "math"

const (
	nameReg               = "^[a-z0-9]([a-z0-9-]{0,30}[a-z0-9]){0,1}$"
	imageNameReg          = "^[a-z0-9]([a-z0-9_./-]{0,30}[a-z0-9]){0,1}$"
	imageVerReg           = "^[a-zA-Z0-9_.-]{1,32}$"
	cmdAndArgsReg         = "^[a-zA-Z0-9 _./-]{0,255}[a-zA-Z0-9]$"
	descriptionReg        = "^[\\S ]{0,512}$"
	envNameReg            = "^[a-zA-Z][a-zA-z0-9._-]{0,30}[a-zA-Z0-9]$"
	envValueReg           = "^[a-zA-Z0-9 _./-]{1,512}$"
	localVolumeReg        = "^[a-z0-9]{1,63}$"
	configmapMountPathReg = `^/[a-zA-Z\d_\-/.]{1,1023}`
	configmapNameReg      = "^[a-zA-Z0-9][a-zA-Z0-9-_]{0,61}[a-zA-Z0-9]$"

	minAppId      = 1
	maxAppId      = math.MaxInt64
	minTemplateId = 1
	maxTemplateId = math.MaxInt64
)
