package models

import (
	"fmt"
	"ginchat/utils"

	"gorm.io/gorm"
)

// 关系
type Contact struct {
	gorm.Model
	OwnerId  uint   //谁的关系信息
	TargetId uint   //对应的谁
	Type     int    //类型  1好友 2群组 3
	Desc     string //描述
}

func (table *Contact) TableName() string {
	return "contact"
}

// 查找好友
func SearchFriend(userId uint) []UserBasic {
	contact := make([]Contact, 0)
	objIds := make([]uint, 0)
	utils.DB.Where("owner_id = ? and type = 1", userId).Find(&contact)
	for _, v := range contact {
		fmt.Println(v)
		objIds = append(objIds, v.TargetId)
	}
	users := make([]UserBasic, 0)
	utils.DB.Where("id in ?", objIds).Find(&users)
	return users
}

func AddFriend(userId, targetId uint) int {
	user := UserBasic{}
	if targetId != 0 {
		user = FindByid(targetId)
		// fmt.Println("user     ", user)
		if user.PassWord != "" {
			contact := Contact{}
			contact.OwnerId = userId
			contact.TargetId = targetId
			contact.Type = 1
			utils.DB.Create(&contact)
			return 0
		}
		return -1
	}
	return -1
}
