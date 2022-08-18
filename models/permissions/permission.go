package permissions

import "gorm.io/gorm"

const TableName = "allowed_users"

type PermissionTable struct {
	OriginUser int `json:"originUser" gorm:"column:user_origin"`
	ChildUser int `json:"childUser" gorm:"column:user_child"`
	ChildUsername string `json:"childUsername" gorm:"column:username_child"`
	AllowUseCaptures bool `json:"allowCaptures" gorm:"column:allow_use_captures"`
}

func (p *PermissionTable) TableName() string {
	return TableName
}

func GetPermissionsForUser(db *gorm.DB, childUserID int) ([]*PermissionTable, error) {
	var permissTable []*PermissionTable
	err := db.Where("user_child = ?", childUserID).Find(&permissTable).Error
	if err != nil {
		return nil, err
	}

	return permissTable, nil
}

func GetAllowedUsers(db *gorm.DB, originUserID int) ([]*PermissionTable, error)  {
	var permissTable []*PermissionTable
	err := db.Where("user_origin = ?", originUserID).Find(&permissTable).Error
	if err != nil {
		return nil, err
	}

	return permissTable, nil
}

func CreatePermission(db *gorm.DB, table PermissionTable) error  {
	return db.Create(&table).Error
}

func UpdatePermissionTable(db *gorm.DB, table PermissionTable) error {
	return db.Where("user_origin = ? AND user_child = ?", table.OriginUser, table.ChildUser).Save(&table).Error
}

func DeletePermissionTable(db *gorm.DB, table PermissionTable) error  {
	return db.Where("user_origin = ? AND user_child = ?", table.OriginUser, table.ChildUser).Delete(&table).Error
}