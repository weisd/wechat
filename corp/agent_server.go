// @description wechat 是腾讯微信公众平台 api 的 golang 语言封装
// @link        https://github.com/chanxuehong/wechat for the canonical source repository
// @license     https://github.com/chanxuehong/wechat/blob/master/LICENSE
// @authors     chanxuehong(chanxuehong@gmail.com)

package corp

import (
	"bytes"
	"errors"
	"sync"
)

type AgentServer interface {
	CorpId() string // 获取应用所属的企业号Id
	AgentId() int64 // 获取应用的Id
	Token() string  // 获取应用的Token

	CurrentAESKey() [32]byte                // 获取当前有效的 AES 加密 Key
	LastAESKey() (key [32]byte, valid bool) // 获取上一个有效的 AES 加密 Key

	MessageHandler() MessageHandler // 获取 MessageHandler
}

var _ AgentServer = (*DefaultAgentServer)(nil)

type DefaultAgentServer struct {
	corpId  string
	agentId int64
	token   string

	rwmutex           sync.RWMutex
	currentAESKey     [32]byte // 当前的 AES Key
	lastAESKey        [32]byte // 最后一个 AES Key
	isLastAESKeyValid bool     // lastAESKey 是否有效, 如果 lastAESKey 是 zero 则无效

	messageHandler MessageHandler
}

// NewDefaultAgentServer 创建一个新的 DefaultAgentServer.
func NewDefaultAgentServer(corpId string, agentId int64, token string, aesKey []byte, handler MessageHandler) (srv *DefaultAgentServer) {
	if len(aesKey) != 32 {
		panic("the length of aesKey must equal to 32")
	}
	if handler == nil {
		panic("nil MessageHandler")
	}

	srv = &DefaultAgentServer{
		corpId:         corpId,
		agentId:        agentId,
		token:          token,
		messageHandler: handler,
	}
	copy(srv.currentAESKey[:], aesKey)
	return
}

func (srv *DefaultAgentServer) CorpId() string {
	return srv.corpId
}
func (srv *DefaultAgentServer) AgentId() int64 {
	return srv.agentId
}
func (srv *DefaultAgentServer) Token() string {
	return srv.token
}
func (srv *DefaultAgentServer) MessageHandler() MessageHandler {
	return srv.messageHandler
}
func (srv *DefaultAgentServer) CurrentAESKey() (key [32]byte) {
	srv.rwmutex.RLock()
	key = srv.currentAESKey
	srv.rwmutex.RUnlock()
	return
}
func (srv *DefaultAgentServer) LastAESKey() (key [32]byte, valid bool) {
	srv.rwmutex.RLock()
	key = srv.lastAESKey
	valid = srv.isLastAESKeyValid
	srv.rwmutex.RUnlock()
	return
}

func (srv *DefaultAgentServer) UpdateAESKey(aesKey []byte) (err error) {
	if len(aesKey) != 32 {
		return errors.New("the length of aesKey must equal to 32")
	}

	srv.rwmutex.Lock()
	defer srv.rwmutex.Unlock()

	if bytes.Equal(aesKey, srv.currentAESKey[:]) {
		return
	}

	srv.isLastAESKeyValid = true
	srv.lastAESKey = srv.currentAESKey
	copy(srv.currentAESKey[:], aesKey)
	return
}
