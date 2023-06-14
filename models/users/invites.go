package users

import (
	"errors"
	"gorm.io/gorm"
)

type InviteCode struct {
	Invite string `json:"invite" gorm:"column:invite"`
	Valid  bool   `json:"-" gorm:"column:valid"`
}

const tableNameInvites = "invite"

func (i *InviteCode) TableName() string {
	return tableNameInvites
}

func (i *InviteCode) ValidateInvite(db *gorm.DB) (bool, error) {
	err := db.Find(&i).Error
	if err != nil {
		return false, errors.New("invite code doesn't exist")
	}

	return i.Valid, nil
}

func (i *InviteCode) SetUnValid(db *gorm.DB) error {
	i.Valid = false
	return db.Save(i).Error
}
