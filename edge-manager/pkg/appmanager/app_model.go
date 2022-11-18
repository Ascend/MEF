package appmanager

import (
	"sync"

	"gorm.io/gorm"

	"edge-manager/pkg/common"
	"edge-manager/pkg/database"
)

var (
	repositoryInitOnce sync.Once
	appRepository      AppRepository
)

// AppRepositoryImpl app service struct
type AppRepositoryImpl struct {
	db *gorm.DB
}

// AppRepository for app method to operate db
type AppRepository interface {
	CreateApp(*AppInfo, *AppContainer) error
	//ListApp(*AppInfo) error
	//DeleteAppByName(*AppInfo) error
	//GetAppsByName(uint64, uint64, string) (*[]AppInfo, error)
}

// GetTableCount get table count
func GetTableCount(tb interface{}) (int, error) {
	var total int64
	err := database.GetDb().Model(tb).Count(&total).Error
	if err != nil {
		return 0, err
	}
	return int(total), nil
}

// AppRepositoryInstance returns the singleton instance of application service
func AppRepositoryInstance() AppRepository {
	repositoryInitOnce.Do(func() {
		appRepository = &AppRepositoryImpl{db: database.GetDb()}
	})
	return appRepository
}

// CreateApp Create application Db
func (a *AppRepositoryImpl) CreateApp(appInfo *AppInfo, container *AppContainer) error {
	if err := a.db.Model(AppInfo{}).Create(appInfo).Error; err != nil {
		return err
	}
	return a.db.Model(AppContainer{}).Create(AppContainer{}).Error
}

//// DeleteApp Delete application Db
//func (a *AppRepositoryImpl) DeleteApp(appInfo *AppInfo) error {
//	return a.db.Model(AppInfo{}).Delete(appInfo).Error
//}

// UP
//func (a *AppRepositoryImpl) UpdateApp(appInfo *AppInfo) error {
//	return a.db.Model(AppInfo{}).Update(appInfo).Error
//}
//
//
//func (a *AppRepositoryImpl) ListApp(appInfo *AppInfo) error {
//	return a.db.Model(AppInfo{}).
//}
//
//// DeleteAppByName delete app
//func (a *AppRepositoryImpl) DeleteAppByName(appInfo *AppInfo) error {
//	return database.GetDb().Model(&AppInfo{}).Where("app_name = ?",
//		appInfo.AppName).Delete(appInfo).Error
//}
//
//// GetAppsByName return SQL result
//func (a *AppRepositoryImpl) GetAppsByName(page, pageSize uint64, appName string) (*[]AppInfo, error) {
//	var nodes []AppInfo
//	return &nodes,
//		database.GetDb().Scopes(getAppByLikeName(page, pageSize, appName)).
//			Find(&nodes).Error
//}
//
//func getAppByLikeName(page, pageSize uint64, appName string) func(db *gorm.DB) *gorm.DB {
//	return func(db *gorm.DB) *gorm.DB {
//		return db.Scopes(paginate(page, pageSize)).Where("app_name like ?", "%"+appName+"%")
//	}
//}

func paginate(page, pageSize uint64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page == 0 {
			page = common.DefaultPage
		}
		if pageSize > common.DefaultMaxPageSize {
			pageSize = common.DefaultMaxPageSize
		}
		offset := (page - 1) * pageSize
		return db.Offset(int(offset)).Limit(int(pageSize))
	}
}
