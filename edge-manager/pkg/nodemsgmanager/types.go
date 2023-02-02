package nodemsgmanager

// EdgeUpgradeInfoReq software upgrade req
type EdgeUpgradeInfoReq struct {
	NodeIDs         []uint64     `json:"nodeIDs"`
	SNs             []string     `json:"sns"`
	SoftWareName    string       `json:"softWareName"`
	SoftWareVersion string       `json:"softWarVersion"`
	DownloadInfo    DownloadInfo `json:"downloadInfo"`
}

type DownloadInfo struct {
	Url      string `json:"url"`
	UserName string `json:"userName"`
	Password string `json:"password"`
}

// EffectInfoReq effect software
type EffectInfoReq struct {
	NodeIDs []uint64 `json:"nodeIDs"`
	SNs     []string `json:"sns"`
}

// SoftwareVersionInfoReq software version info req
type SoftwareVersionInfoReq struct {
	NodeIDs []uint64 `json:"nodeIDs"`
	SNs     []string `json:"sns"`
}
