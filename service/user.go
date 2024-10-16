package service

import (
	"github.com/eatmoreapple/openwechat"
	"github.com/patrickmn/go-cache"
	"github.com/qingconglaixueit/wechatbot/config"
	"time"
)

// UserServiceInterface 用户业务接口
type UserServiceInterface interface {
	GetUserSessionContext() []UserDialogHistoryPair
	SetUserSessionContext(question, reply string)
	ClearUserSessionContext()
}

var _ UserServiceInterface = (*UserService)(nil)

// UserService 用戶业务
type UserService struct {
	// 缓存
	cache *cache.Cache
	// 用户
	user *openwechat.User
}

// NewUserService 创建新的业务层
func NewUserService(cache *cache.Cache, user *openwechat.User) UserServiceInterface {
	return &UserService{
		cache: cache,
		user:  user,
	}
}

// ClearUserSessionContext 清空GTP上下文，接收文本中包含`我要问下一个问题`，并且Unicode 字符数量不超过20就清空
func (s *UserService) ClearUserSessionContext() {
	s.cache.Delete(s.user.ID())
}

// GetUserSessionContext 获取用户会话上下文文本
func (s *UserService) GetUserSessionContext() []UserDialogHistoryPair {
	// 1.获取上次会话信息，如果没有直接返回空字符串
	session, ok := s.cache.Get(s.user.ID())
	if !ok {
		return make([]UserDialogHistoryPair, 0)
	}

	// 2.如果字符长度超过等于4000，强制清空会话（超过GPT会报错）。
	ret := make([]UserDialogHistoryPair, 0)
	dialogRecord := session.(*userDialogRecord)
	dialogLen := len(dialogRecord.histories)
	if dialogLen == 0 {
		return ret
	}
	startInd, totalLen := dialogLen, 0
	for startInd > 0 {
		curInd := startInd - 1
		totalLen +=
			len(dialogRecord.histories[curInd].BotReply) + len(dialogRecord.histories[curInd].UserMsg)
		if totalLen > 4000 {
			break
		}
		startInd--
	}
	if startInd < dialogLen {
		ret = dialogRecord.histories[startInd:]
	}
	// 3.返回上文
	return ret
}

// SetUserSessionContext 设置用户会话上下文文本，question用户提问内容，GTP回复内容
func (s *UserService) SetUserSessionContext(question, reply string) {
	session, ok := s.cache.Get(s.user.ID())
	if !ok {
		session = newUserDialogRecord(config.LoadConfig().MaxHistoryRound)
	}
	dialogRecord := session.(*userDialogRecord)
	dialogRecord.updateHistory(UserDialogHistoryPair{
		UserMsg:  question,
		BotReply: reply,
	})
	s.cache.Set(s.user.ID(), dialogRecord, time.Second*config.LoadConfig().SessionTimeout)
}

type userDialogRecord struct {
	histories        []UserDialogHistoryPair
	maxHistoryRounds int
}

type UserDialogHistoryPair struct {
	UserMsg  string
	BotReply string
}

func newUserDialogRecord(maxHistoryRounds int) *userDialogRecord {
	return &userDialogRecord{
		histories:        make([]UserDialogHistoryPair, 0, maxHistoryRounds),
		maxHistoryRounds: maxHistoryRounds,
	}
}

func (dialog *userDialogRecord) updateHistory(pair UserDialogHistoryPair) {
	dialog.histories = append(dialog.histories, pair)
	dialogLength := len(dialog.histories)
	if dialogLength > dialog.maxHistoryRounds {
		startIndex := dialogLength % dialog.maxHistoryRounds
		dialog.histories = dialog.histories[startIndex:]
	}
}
