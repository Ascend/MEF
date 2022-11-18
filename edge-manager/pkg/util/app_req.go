package util

// CreateAppReq Create application
type CreateAppReq struct {
	AppName       string            `json:"appName"`
	ContainerName string            `json:"containerName"`
	CpuRequest    string            `json:"cpuRequest"`
	CpuLimit      string            `json:"cpuLimit"`
	MemRequest    string            `json:"memRequest"`
	MemLimit      string            `json:"memLimit"`
	Npu           string            `json:"npu"`
	ImageName     string            `json:"imageName"`
	ImageVersion  string            `json:"imageVersion"`
	Command       []string          `json:"command"`
	Env           map[string]string `json:"env"`
	ContainerPort string            `json:"ContainerPort"`
	HostIp        string            `json:"hostIp"`
	HostPort      int               `json:"hostPort"`
	UserId        int               `json:"userId"`
	GroupId       int               `json:"groupId"`
	Description   string            `json:"description"`
}

type UpdateAppReq struct {
	AppID     uint64 `json:"appID"`
	ImageName string `json:"imageName"`
}

type DeleteAppReq struct {
	AppName string `json:"appName"`
}

type DeployAppReq struct {
	AppName       string `json:"appName"`
	NodeGroupName string `json:"nodeGroupName"`
}

type UndeployAppReq struct {
	AppName       string `json:"appName"`
	NodeGroupName string `json:"nodeGroupName"`
}

type GetNodeByAppIdReq struct {
	AppId int `json:"appId"`
}
