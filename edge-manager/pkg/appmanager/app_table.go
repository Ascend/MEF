package appmanager

// AppInfo is app db table info
type AppInfo struct {
	ID          uint64 `gorm:"type:integer;primaryKey;autoIncrement:true"`
	AppName     string `gorm:"type:char(128);unique;not null"`
	Description string `gorm:"type:char(255);" json:"description"`
	AppGroupID  uint64 `gorm:"type:integer;not null"`
	CreatedAt   string `gorm:"type:char(19);not null"`
	ModifiedAt  string `gorm:"type:char(19)lnot null"`
}

// AppContainer containers belonging to the same App share the same AppName.
type AppContainer struct {
	Id            uint64            `gorm:"type:integer;not_null;primary_key;auto_increment"`
	GroupId       uint64            `gorm:"type:integer;not_null"`
	AppName       string            `gorm:"type:varchar(32);not_null"`
	CreatedAt     string            `gorm:"type:char(19);not_null"`
	ModifiedAt    string            `gorm:"type:char(19);not_null"`
	ContainerName string            `gorm:"type:varchar(32);not_null"`
	ImageName     string            `gorm:"type:varchar(64);not_null"`
	ImageVersion  string            `gorm:"type:varchar(16);not_null"`
	CpuRequest    string            `gorm:"type:varchar(7);not_null"`
	CpuLimit      string            `gorm:"type:varchar(7)"`
	MemoryRequest string            `gorm:"type:varchar(7);not_null"`
	MemoryLimit   string            `gorm:"type:varchar(7)"`
	Npu           string            `gorm:"type:varchar(5)"`
	Env           map[string]string `gorm:"type:text"`
	ContainerUser UserInfo          `gorm:"type:text"`
	ContainerPort string            `gorm:"type:varchar(5)"`
	ContainerHost HostAddr          `gorm:"type:text"`
	Command       []string          `gorm:"type:text"`
}

type AppInstance struct {
	ID          int64  `gorm:"type:Integer;primaryKey;autoIncrement:true"`
	PodName     string `gorm:"type:char(42);unique;not null"`
	NodeName    string `gorm:"type:char(64);not null"`
	NodeGroupID uint64 `gorm:"type:Integer;not null"`
	status      string `gorm:"type:char(50);not null"`
	VersionID   string `gorm:"not null"`
	CreatedAt   string `gorm:"type:time;not null"`
	ChangedAt   string `gorm:"type:time;not null"`
	//ReserverdField
}

type UserInfo struct {
	UserId      int
	UserGroupId int
}

type HostAddr struct {
	HostIp   string
	HostPort int
}
