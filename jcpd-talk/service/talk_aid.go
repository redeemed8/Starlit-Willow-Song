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

type GroupClient struct {
	msgChan chan []byte
	ConnMap map[uint32]*websocket.Conn
}

type GroupClientManager struct {
	Clients [10]map[uint32]*GroupClient
}

func NewGroupClientManager() *GroupClientManager {
	var maps [10]map[uint32]*GroupClient
	for i := 0; i < 10; i++ {
		maps[i] = make(map[uint32]*GroupClient)
	}
	return &GroupClientManager{Clients: maps}
}

var GroupManager = NewGroupClientManager()
