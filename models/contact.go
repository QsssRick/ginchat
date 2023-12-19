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

func AddFriend(userId, targetId uint) (int, string) {
	user := UserBasic{}
	if targetId != 0 {
		user = FindByid(targetId)
		// fmt.Println("user     ", user)
		if user.PassWord != "" {
			if targetId == userId {
				return -1, "不能添加自己为好友"
			}
			contact0 := Contact{}
			utils.DB.Where("owner_id = ? and target_id = ? and type = 1", userId, targetId).First(&contact0)
			if contact0.ID != 0 {
				return -1, "已经是好友"
			}
			tx := utils.DB.Begin() //开启事务,如果有错误就回滚
			//事务一旦开始，不论任何异常都回滚
			defer func() {
				if r := recover(); r != nil {
					tx.Rollback()
				}
			}()
			// fmt.Println("user     ", user)
			contact := Contact{}
			contact.OwnerId = userId
			contact.TargetId = targetId
			contact.Type = 1
			if err := utils.DB.Create(&contact).Error; err != nil {
				tx.Rollback()
				return -1, "添加好友失败"
			}
			contact1 := Contact{}
			contact1.OwnerId = targetId
			contact1.TargetId = userId
			contact1.Type = 1
			if err := utils.DB.Create(&contact1).Error; err != nil {
				tx.Rollback()
				return -1, "添加好友失败"
			}
			tx.Commit()
			return 0, "添加好友成功"
		}
		return -1, "该用户不存在"
	}
	return -1, "用户id不能为空"
}
