package models

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"

	"github.com/fatih/set"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

// Message 消息
type Message struct {
	gorm.Model
	FormId   int64  // 发送者
	TargetId int64  // 接收者
	Type     int    // 消息类型 群聊 私聊 广播
	Content  string // 消息内容
	Media    int    // 发送类型 图片 语音 文字
	Pic      string // 图片
	Url      string // 链接
	Desc     string // 描述
	Amount   int    // 其他数字统计
}

func (table *Message) TableName() string {
	return "message"
}

type Node struct {
	Conn       *websocket.Conn
	DataQueueu chan []byte
	GroupSets  set.Interface
}

// 映射关系
var clientMap map[int64]*Node = make(map[int64]*Node, 0)

// 读写锁
var rwLocker sync.RWMutex

// 需要：发送者id ，接收者id, 消息类型，发送的内容，发送类型
func Chat(writer http.ResponseWriter, request *http.Request) {
	//1、校验 token等合法性
	// token:=query.GET("token")
	query := request.URL.Query()
	Id := query.Get("userId")
	userId, _ := strconv.ParseInt(Id, 10, 64)
	// targetId := query.Get("targetId")
	// msgType := query.Get("type")
	// context := query.Get("context")
	isvalida := true
	conn, err := (&websocket.Upgrader{
		//token 校验
		CheckOrigin: func(r *http.Request) bool {
			return isvalida
		},
	}).Upgrade(writer, request, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	//2、获取conn
	node := &Node{
		Conn:       conn,
		DataQueueu: make(chan []byte, 50),
		GroupSets:  set.New(set.ThreadSafe),
	}

	//3、用户关系

	//4、userid和node绑定并且加锁
	rwLocker.Lock()
	clientMap[userId] = node
	rwLocker.Unlock()

	//5、完成发送逻辑
	go sendProc(node)
	//6、完成接收逻辑
	go revcProc(node)
	sendMsg(userId, []byte("欢迎进入聊天室"))
}

func sendProc(node *Node) {
	for {
		select {
		case data := <-node.DataQueueu:
			err := node.Conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}

func revcProc(node *Node) {
	for {
		_, data, err := node.Conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			return
		}
		broadMsg(data) // 广播
		fmt.Println("[ws]revcProc <<<< ", string(data))
	}
}

var udpsendChan chan []byte = make(chan []byte, 1024)

func broadMsg(data []byte) {
	udpsendChan <- data
}

func init() {
	go udpSendProc()
	go udpRecvProc()
}

// udp数据发送协程
func udpSendProc() {
	con, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4(192, 168, 0, 255),
		Port: 3000,
	})
	defer con.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	for {
		select {
		case data := <-udpsendChan:
			_, err := con.Write(data)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}

// udp数据接受协程
func udpRecvProc() {
	con, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: 3000,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer con.Close()
	for {
		var buf [512]byte
		n, err := con.Read(buf[:])
		if err != nil {
			fmt.Println(err)
			return
		}
		dispatch(buf[:n])
	}
}

// 后端调度逻辑处理
func dispatch(data []byte) {
	msg := Message{}
	err := json.Unmarshal(data, &msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	switch msg.Type {
	case 1: //私信
		sendMsg(msg.TargetId, data)
		// case 2: //群聊
		// 	sendGroupMsg(msg)
		// case 3: //广播
		// 	sendAllMsg(msg)
		// case 4:

	}
}

func sendMsg(userId int64, msg []byte) {
	rwLocker.RLock()
	node, ok := clientMap[userId]
	rwLocker.RUnlock()
	if ok {
		node.DataQueueu <- msg
	}

}
