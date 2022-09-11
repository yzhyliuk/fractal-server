package users

import (
	"gorm.io/gorm"
	"newTradingBot/api/helpers"
	"newTradingBot/api/security"
)

const tableName = "users"
const ConfirmationCodeLength = 6
const ResetCodeLength = 11

type User struct {
	ID int `json:"id" gorm:"column:id"`
	Username string `json:"username" gorm:"column:username"`
	Email string `json:"email" gorm:"column:email"`
	Password string `json:"-" gorm:"column:password"`
	ProfilePhoto *string `json:"profilePhoto" gorm:"column:profilephoto"`
	Verified bool `json:"verified" gorm:"column:verified"`
	ResetCode *string `json:"-" gorm:"column:reset_code"`
	ConfirmationCode *string `json:"-" gorm:"column:confirmation_code"`
}

type NewUser struct {
	Username string `json:"username" gorm:"column:username"`
	Email string `json:"email" gorm:"column:email"`
	Password string `json:"password" gorm:"column:password"`
}

type UserCredentials struct {
	Email string `json:"email" gorm:"column:email"`
	Password string `json:"password" gorm:"column:password"`
}

func (u *User) TableName() string {
	return tableName
}

func (n *NewUser) TableName() string  {
	return tableName
}

func (n *UserCredentials) TableName() string  {
	return tableName
}

func (n *NewUser) hashPassword() error {
	hashPass, err := security.GetHashedString(n.Password)
	if err != nil {
		return err
	}
	n.Password = hashPass
	return nil
}

func ListAllUsers(db *gorm.DB) ([]*User, error) {
	var list []*User
	err := db.Find(&list).Error
	if err != nil {
		return nil, err
	}

	return list, nil
}

func Create(db *gorm.DB, newUser *NewUser) (*User, error) {
	err := newUser.hashPassword()
	if err != nil {
		return nil, err
	}

	confirmationCode := helpers.GenerateCode(ConfirmationCodeLength)

	user := &User{
		0, newUser.Username, newUser.Email, newUser.Password, nil, false, nil, &confirmationCode,
	}

	result := db.Create(user)
	if result.Error != nil {
		return nil, result.Error
	}

	return user, nil
}

func GetByEmail(db *gorm.DB, email string) (*User, error)  {
	var user User
	err := db.Where("email = ?", email).Find(&user).Error
	return &user, err
}

func GetUserByUsername(db *gorm.DB, username string) (*User, error) {
	var user User
	err := db.Where("username = ?", username).Find(&user).Error
	return &user, err
}

func GetUserByID(db *gorm.DB, userID int) (*User, error) {
	var user User
	err := db.Where("id = ?", userID).Find(&user).Error
	return &user, err
}

func UpdateUserName(db *gorm.DB, userID int, username string) error {
	return db.Table(tableName).Where("id = ?", userID).Update("username", username).Error
}

func UpdateProfilePhoto(db *gorm.DB, userID int, filename string) error {
	return db.Table(tableName).Where("id = ?", userID).Update("profilephoto", filename).Error
}