package service

import (
	"github.com/gorilla/websocket"
	"net/http"
)

const (
	ReadBufferSize  = 1500
	WriteBufferSize = 1500
)

func CheckOrigin(*http.Request) bool {
	return true
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  ReadBufferSize,
	WriteBufferSize: WriteBufferSize,
	CheckOrigin:     CheckOrigin,
}

// 	--------------------------------------

type ClientConnManager struct {
	ConnMaps  [10]map[uint32]map[uint32]*websocket.Conn
	ServerKey uint32
}

const Server = 0 //	 正常用户的id大于0

func NewClientConnManager() *ClientConnManager {
	var maps [10]map[uint32]map[uint32]*websocket.Conn
	for i := 0; i < 10; i++ {
		maps[i] = make(map[uint32]map[uint32]*websocket.Conn)
	}
	return &ClientConnManager{ConnMaps: maps, ServerKey: Server}
}

var ConnManager = NewClientConnManager()

func (manager *ClientConnManager) getPosMap(curUserId uint32) *map[uint32]map[uint32]*websocket.Conn {
	index := curUserId % 10         //	 取个位数得到其在组中的位置
	return &manager.ConnMaps[index] //	得到其所在的map
}

func (manager *ClientConnManager) AddClientConn(curUserId uint32, targetId uint32, conn *websocket.Conn) {
	PosMap := manager.getPosMap(curUserId)
	targets, exists := (*PosMap)[curUserId]
	if exists {
		targets[targetId] = conn
	} else {
		(*PosMap)[curUserId] = make(map[uint32]*websocket.Conn)
		(*PosMap)[curUserId][targetId] = conn
	}
}

func (manager *ClientConnManager) DelClientConn(curUserId uint32, targetId uint32) {
	PosMap := manager.getPosMap(curUserId)
	targets, exists := (*PosMap)[curUserId]
	if !exists {
		return
	}
	_, targetExist := targets[targetId]
	if !targetExist {
		return
	}
	delete((*PosMap)[curUserId], targetId)
}

func (manager *ClientConnManager) LoadClientConn(curUserId uint32, targetId uint32) *websocket.Conn {
	PosMap := manager.getPosMap(curUserId)
	targets, exists := (*PosMap)[curUserId]
	if !exists {
		return nil
	}
	conn, targetExist := targets[targetId]
	if !targetExist {
		return nil
	}
	return conn
}

//  ---------------------------------------------------------------------------------

type GroupClientManager struct {
	Groups map[uint32]*GroupInfo
}

type GroupInfo struct {
	ConnMaps [10]map[uint32]*websocket.Conn
	C        chan []byte
}

var GroupManager = NewGroupClientManager()

func NewGroupClientManager() *GroupClientManager {
	return &GroupClientManager{Groups: make(map[uint32]*GroupInfo)}
}

func (groupManager *GroupClientManager) getPosMap(groupInfo *GroupInfo, curUserId uint32) *map[uint32]*websocket.Conn {
	index := curUserId % 10           //	 取个位数得到其在组中的位置
	return &groupInfo.ConnMaps[index] //	得到其所在的map
}

func (groupManager *GroupClientManager) AddUserToGroupConns(userId uint32, groupId uint32, conn *websocket.Conn) {
	group := groupManager.Groups[groupId] //  尝试获取群id对应的结果
	if group == nil {                     //  如果不存在，为其赋值
		var maps [10]map[uint32]*websocket.Conn
		for i := 0; i < 10; i++ {
			maps[i] = make(map[uint32]*websocket.Conn)
		}
		groupManager.Groups[groupId] = &GroupInfo{ConnMaps: maps}
		group = groupManager.Groups[groupId]
	}
	posMap := groupManager.getPosMap(group, userId) //  在当前群中 找到 该用户的位置
	(*posMap)[userId] = conn                        //  添加用户连接
}

func (groupManager *GroupClientManager) DelUserConnFromGroup(userId uint32, groupId uint32) {
	group := groupManager.Groups[groupId]
	if group == nil {
		return
	}
	posMap := groupManager.getPosMap(group, userId) //  在当前群中 找到 该用户的位置
	if posMap == nil {
		return
	}
	delete(*posMap, userId)
}

func (groupManager *GroupClientManager) GetGroup(groupId uint32) *GroupInfo {
	return groupManager.Groups[groupId]
}
