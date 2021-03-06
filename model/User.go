package model

import (
	"GoBIMS/utils"
	"GoBIMS/utils/errmsg"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	UserName string `gorm:"type:varchar(20);not null;columns:user_name" json:"user_name" validate:"required,min=4,max=12" label:"用户名"`
	PassWord string `gorm:"type:varchar(500);not null columns:pass_word" json:"pass_word" validate:"required,min=6,max=120" label:"密码"`
	Role     int    `gorm:"type:int;default:2;columns:role" json:"role" validate:"required" label:"角色码"`
}

// CheckLogin 后台登录验证
func CheckLogin(username string, password string) (User, int) {
	var user User
	var PasswordErr error

	db.Where("user_name = ?", username).First(&user)

	PasswordErr = bcrypt.CompareHashAndPassword([]byte(user.PassWord), []byte(password))

	if user.ID == 0 {
		return user, errmsg.ErrorUserNotExist
	}
	if PasswordErr != nil {
		return user, errmsg.ErrorPasswordWrong
	}
	if user.Role != 1 {
		return user, errmsg.ErrorUserNoRight
	}
	return user, errmsg.SUCCESS
}

// CheckLoginFront 前台登录
func CheckLoginFront(username string, password string) (User, int) {
	var user User
	var PasswordErr error

	db.Where("user_name = ?", username).First(&user)

	PasswordErr = bcrypt.CompareHashAndPassword([]byte(user.PassWord), []byte(password))
	if user.ID == 0 {
		return user, errmsg.ErrorUserNotExist
	}
	if PasswordErr != nil {
		return user, errmsg.ErrorPasswordWrong
	}
	return user, errmsg.SUCCESS
}

// CheckUser 查询用户是否存在
func CheckUser(user User) (code int) {
	db.Select("id").Where("user_name = ?", user.UserName).First(&user)
	if len(user.UserName) == 0 {
		user.UserName = utils.RandString(10)
	}
	if len(user.PassWord) < 6 {
		return errmsg.ErrorPasswordLessThan6
	}
	if user.ID > 0 {
		return errmsg.ErrorUsernameUsed //1001
	}
	return errmsg.SUCCESS //200
}

// CreatUser 新增用户
func CreatUser(user *User) (code int) {
	err := db.Create(&user).Error
	if err != nil {
		return errmsg.ERROR //500
	}
	return errmsg.SUCCESS //200
}

// CheckUserPage 查询用户列表
func CheckUserPage(username string, pageSize int, pageNum int) ([]User, int, int64) {
	var user []User
	var total int64
	if username != "" {
		// fmt.Println("ffff")
		db.Select("id, created_at, updated_at, deleted_at, user_name, pass_word, role").
			Where("user_name LIKE ?", "%"+username+"%").
			Limit(pageSize).
			Offset((pageNum - 1) * pageSize).
			Find(&user)
		db.Model(&user).Where(
			"user_name LIKE ?", "%"+username+"%",
		).Count(&total)
		return user, errmsg.SUCCESS, total
	}
	db.Select("id, created_at, updated_at, deleted_at, user_name, pass_word, role").Limit(pageSize).Offset((pageNum - 1) * pageSize).Find(&user)
	db.Model(&user).Count(&total)

	if err != nil {
		return user, errmsg.ERROR, 0
	}
	return user, errmsg.SUCCESS, total
}

// BeforeCreate 密码加密&权限控制
func (u *User) BeforeCreate(_ *gorm.DB) (err error) {
	u.PassWord = ScryptPassWord(u.PassWord)
	// u.Role = 2
	return nil
}

func (u *User) BeforeUpdate(_ *gorm.DB) (err error) {
	u.PassWord = ScryptPassWord(u.PassWord)
	return nil
}

// ScryptPassWord 密码加密
func ScryptPassWord(password string) string {
	const cost = 10
	HashPassWord, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		log.Fatal(err)
	}
	return string(HashPassWord)
}

// EditUser 编辑用户信息
func EditUser(id int, data *User) int {
	var user User
	var maps = make(map[string]interface{})
	maps["user_name"] = data.UserName
	maps["role"] = data.Role
	err = db.Model(&user).Where("id = ? ", id).Updates(maps).Error
	if err != nil {
		return errmsg.ERROR
	}
	return errmsg.SUCCESS
}

// DeleteUser 删除用户
func DeleteUser(id int) int {
	var user User
	err = db.Where("id = ? ", id).Delete(&user).Error
	if err != nil {
		return errmsg.ERROR
	}
	return errmsg.SUCCESS
}

// CheckUpUser 更新查询
func CheckUpUser(id int, name string) (code int) {
	var user User
	db.Select("id, username").Where("username = ?", name).First(&user)
	if user.ID == uint(id) {
		return errmsg.SUCCESS
	}
	if user.ID > 0 {
		return errmsg.ErrorUsernameUsed //1001
	}
	return errmsg.SUCCESS
}
