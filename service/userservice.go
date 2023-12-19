package service

import (
	"fmt"
	"ginchat/models"
	"ginchat/utils"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// GetIndex
// @Summary 所有用户
// @Tags 用户模块
// @Success 200 {string} json{"code","message"}
// @Router /user/getUserList [get]
func GetUserList(c *gin.Context) {
	data := make([]*models.UserBasic, 10)
	data = models.GetUserList()

	c.JSON(200, gin.H{
		"code": 0, //0表示成功,-1表示失败
		"data": data,
	})
}

// CreateUser
// @Summary 新增用户
// @Tags 用户模块
// @param name query string false "用户名"
// @param password query string false "密码"
// @param repassword query string false "确认密码"
// @Success 200 {string} json{"code","message"}
// @Router /user/createUser [post]
func CreateUser(c *gin.Context) {
	user := models.UserBasic{}
	// user.Name = c.Query("name")
	// password := c.Query("password")
	// repassword := c.Query("repassword")
	user.Name = c.Request.FormValue("name")
	password := c.Request.FormValue("password")
	repassword := c.Request.FormValue("repassword")

	salt := fmt.Sprintf("%06d", rand.Int31())

	data := models.FindUserByName(user.Name)
	if user.Name == "" || password == "" || repassword == "" {
		c.JSON(200, gin.H{
			"code":    -1, //0表示成功,-1表示失败
			"message": "用户名或密码不能为空！",
			"data":    user,
		})
		return
	}
	if data.Name != "" {
		c.JSON(200, gin.H{
			"code":    -1, //0表示成功,-1表示失败
			"message": "用户名已存在",
			"data":    user,
		})
		return
	}
	if password != repassword {
		c.JSON(200, gin.H{
			"code":    -1, //0表示成功,-1表示失败
			"message": "两次密码不一致",
			"data":    user,
		})
		return
	}
	// user.PassWord = password
	user.PassWord = utils.MakePassword(password, salt)
	user.Salt = salt
	user.LoginTime = time.Now()
	user.LoginOutTime = time.Now()
	user.HeartbeatTime = time.Now()
	models.CreateUser(user)
	c.JSON(200, gin.H{
		"code":    0, //0表示成功,-1表示失败
		"message": "新增用户成功",
		"data":    user,
	})
}

// DeleteUser
// @Summary 删除用户
// @Tags 用户模块
// @param id query string false "id"
// @Success 200 {string} json{"code","message"}
// @Router /user/deleteUser [get]
func DeleteUser(c *gin.Context) {
	user := models.UserBasic{}
	id, _ := strconv.Atoi(c.Query("id"))
	user.ID = uint(id)
	models.DeleteUser(user)
	c.JSON(200, gin.H{
		"code":    0, //0表示成功,-1表示失败
		"message": "删除用户成功",
		"data":    user,
	})
}

// UpdateUser
// @Summary 修改用户
// @Tags 用户模块
// @param id formData string false "id"
// @param name formData string false "name"
// @param password formData string false "password"
// @param email formData string false "email"
// @param phone formData string false "phone"
// @Success 200 {string} json{"code","message"}
// @Router /user/updateUser [post]
func UpdateUser(c *gin.Context) {
	user := models.UserBasic{}
	id, _ := strconv.Atoi(c.PostForm("id"))
	user.ID = uint(id)
	user.Name = c.PostForm("name")
	user.PassWord = c.PostForm("password")
	user.Email = c.PostForm("email")
	user.Phone = c.PostForm("phone")
	fmt.Println(user)
	_, err := govalidator.ValidateStruct(user)
	if err != nil {
		fmt.Println(err)
		c.JSON(200, gin.H{
			"code":    -1, //0表示成功,-1表示失败
			"message": "修改用户失败",
			"data":    user,
		})
	} else {
		models.UpdateUser(user)
		c.JSON(200, gin.H{
			"code":    0, //0表示成功,-1表示失败
			"message": "修改用户成功",
			"data":    user,
		})
	}

}

// FindUserByNameAndPwd
// @Summary 校验用户
// @Tags 用户模块
// @param name query string false "name"
// @param password query string false "password"
// @Success 200 {string} json{"code","message"}
// @Router /user/findUserByNameAndPwd [post]
func FindUserByNameAndPwd(c *gin.Context) {
	data := models.UserBasic{}
	// name := c.Query("name")
	// password := c.Query("password")
	name := c.Request.FormValue("name")
	password := c.Request.FormValue("password")
	user := models.FindUserByName(name)
	if user.Name == "" {
		c.JSON(200, gin.H{
			"code":    -1, //0表示成功,-1表示失败
			"message": "用户不存在",
			"data":    data,
		})
		return
	}
	if !utils.ValidPassword(password, user.Salt, user.PassWord) {
		c.JSON(200, gin.H{
			"code":    -1, //0表示成功,-1表示失败
			"message": "密码错误",
			"data":    data,
		})
		return
	}
	pwd := utils.MakePassword(password, user.Salt)
	data = models.FindUserByNameAndPwd(name, pwd)
	c.JSON(200, gin.H{
		"code":    0, //0表示成功,-1表示失败
		"message": "登录成功",
		"data":    data,
	})
}

// 防止跨域站点伪造请求
var upGrade = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func SendMsg(c *gin.Context) {
	// 1. 从连接中获取ws
	ws, err := upGrade.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func(ws *websocket.Conn) {
		err = ws.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(ws)
	MsgHandler(ws, c)
}

func MsgHandler(ws *websocket.Conn, c *gin.Context) {
	for {
		msg, err := utils.Subscribe(c, utils.PublishKey)
		fmt.Println("发送消息")
		if err != nil {
			fmt.Println(err)
		}
		tm := time.Now().Format("2006-01-02 15:04:05")
		m := fmt.Sprintf("[ws][%s]:%s", tm, msg)
		err = ws.WriteMessage(1, []byte(m))
		if err != nil {
			fmt.Println(err)
		}
	}
}

func SendUserMsg(c *gin.Context) {
	models.Chat(c.Writer, c.Request)
}

func SearchFriends(c *gin.Context) {
	id, _ := strconv.Atoi(c.Request.FormValue("userId"))
	users := models.SearchFriend(uint(id))
	// c.JSON(200, gin.H{
	// 	"code":    0, //0表示成功,-1表示失败
	// 	"message": "搜索好友成功",
	// 	"data":    user,
	// })
	utils.RespOKList(c.Writer, users, len(users))
}

func AddFriend(c *gin.Context) {
	userId, _ := strconv.Atoi(c.Request.FormValue("userId"))
	targetId, _ := strconv.Atoi(c.Request.FormValue("targetId"))
	code, msg := models.AddFriend(uint(userId), uint(targetId))
	// c.JSON(200, gin.H{
	// 	"code":    0, //0表示成功,-1表示失败
	// 	"message": "搜索好友成功",
	// 	"data":    user,
	// })
	if code == 0 {
		utils.RespOK(c.Writer, code, msg)
	} else {
		utils.RespFail(c.Writer, msg)
	}
}

func CreateCommunity(c *gin.Context) {
	ownerId, _ := strconv.Atoi(c.Request.FormValue("ownerId"))
	Name := c.Request.FormValue("name")
	community := models.Community{OwnerId: uint(ownerId), Name: Name}
	code, msg := models.CreateCommunity(community)
	// c.JSON(200, gin.H{
	// 	"code":    0, //0表示成功,-1表示失败
	// 	"message": "搜索好友成功",
	// 	"data":    user,
	// })
	if code == 0 {
		utils.RespOK(c.Writer, code, msg)
	} else {
		utils.RespFail(c.Writer, msg)
	}
}
