package appmanager

// AppParam app param
type AppParam struct {
	AppId       uint64      `json:"appId"`
	AppName     string      `json:"appName"`
	Description string      `json:"description"`
	Containers  []Container `json:"containers"`
}

// TemplateParam app param
type TemplateParam struct {
	Id          uint64      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Containers  []Container `json:"containers"`
}

// Container encapsulate container request
type Container struct {
	Name         string          `json:"name"`
	Image        string          `json:"image"`
	ImageVersion string          `json:"imageVersion"`
	CpuRequest   string          `json:"cpuRequest"`
	CpuLimit     string          `json:"cpuLimit"`
	MemRequest   string          `json:"memRequest"`
	MemLimit     string          `json:"memLimit"`
	Npu          string          `json:"npu"`
	Command      []string        `json:"command"`
	Args         []string        `json:"args"`
	Env          []EnvVar        `json:"env"`
	Ports        []ContainerPort `json:"containerPort"`
	UserId       int             `json:"userId"`
	GroupId      int             `json:"groupId"`
}

// EnvVar encapsulate env request
type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ContainerPort provide ports mapping
type ContainerPort struct {
	Name          string `json:"name"`
	Proto         string `json:"proto"`
	ContainerPort int32  `json:"containerPort"`
	HostIp        string `json:"hostIp"`
	HostPort      int32  `json:"hostPort"`
}
