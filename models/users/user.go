package users

import (
	"gorm.io/gorm"
	"newTradingBot/api/security"
)

const tableName = "users"

type User struct {
	ID int `json:"id" gorm:"column:id"`
	Username string `json:"username" gorm:"column:username"`
	Email string `json:"email" gorm:"column:email"`
	Password string `json:"-" gorm:"column:password"`
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

func Create(db *gorm.DB, newUser *NewUser) (*User, error) {
	err := newUser.hashPassword()
	if err != nil {
		return nil, err
	}

	user := &User{
		0, newUser.Username, newUser.Email, newUser.Password,
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