package tgbotapi

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

func (bot *BotAPI) FindCbk(chatID, userID int64, msgID int) (*Callback, bool) {
	if ch, ok := bot.callbackMap.Load(fmt.Sprintf("%v:%v:%v", chatID, msgID, userID)); ok {
		return ch.(*Callback), true
	} else {
		return nil, false
	}
}

func (bot *BotAPI) FindMsgCbk(chatID, userID int64) (*Msg, bool) {
	if ch, ok := bot.msgMap.Load(fmt.Sprintf("%v:%v", chatID, userID)); ok {
		return ch.(*Msg), true
	} else {
		return nil, false
	}
}

type CallBackChan chan *ChanData

type MsgChan chan *Message

type ChanData struct {
	Key   string
	Value string
}

type Callback struct {
	cbkMap         *sync.Map
	chatID, userID int64
	msgID          int
	cbChan         CallBackChan
	once           sync.Once
}

func ParseCbkData(cbkData *CallbackQuery) *ChanData {
	before, after, found := strings.Cut(cbkData.Data, ":")
	data := &ChanData{}
	data.Key = before
	if found {
		data.Value = after
	} else {
		data.Value = before
	}
	return data
}

type Msg struct {
	cbkMap         *sync.Map
	chatID, userID int64
	msgChan        MsgChan
	closeOnce      sync.Once
}

func (cb *Callback) Close() {
	cb.once.Do(func() {
		cb.cbkMap.Delete(fmt.Sprintf("%v:%v:%v", cb.chatID, cb.msgID, cb.userID))
		close(cb.cbChan)
	})
}

func (bot *BotAPI) CloseCbkChan(chatID, userID int64, msgID int) error {
	if value, loaded := bot.callbackMap.LoadAndDelete(fmt.Sprintf("%v:%v:%v", chatID, msgID, userID)); loaded {
		value.(*Callback).Close()
		return nil
	} else {
		return errors.New("not found")
	}
}

func (cb *Callback) Chan() CallBackChan {
	return cb.cbChan
}

func (cb *Callback) MsgID() int {
	return cb.msgID
}

func (cb *Callback) ChatID() int64 {
	return cb.chatID
}

func (cb *Callback) UserID() int64 {
	return cb.userID
}

func newCallbackChan(buffer int) CallBackChan {
	return make(CallBackChan, buffer)
}

func (bot *BotAPI) NewCbk(chatID, userID int64, messageID int) (*Callback, error) {
	if _, ok := bot.callbackMap.Load(fmt.Sprintf("%v:%v:%v", chatID, messageID, userID)); ok {
		return nil, errors.New("user cbk arlready exists")
	}
	cbk := Callback{once: sync.Once{}, cbkMap: bot.callbackMap, chatID: chatID, userID: userID, msgID: messageID, cbChan: newCallbackChan(0)}
	bot.callbackMap.Store(fmt.Sprintf("%v:%v:%v", chatID, messageID, userID), &cbk)
	return &cbk, nil
}

func newMsgChan(buffer int) MsgChan {
	return make(MsgChan, buffer)
}

func (bot *BotAPI) NewMsgCbk(chatID, userID int64) (*Msg, error) {
	if _, ok := bot.msgMap.Load(fmt.Sprintf("%v:%v", chatID, userID)); ok {
		return nil, errors.New("user msg cbk already exists")
	}
	msg := Msg{closeOnce: sync.Once{}, cbkMap: bot.msgMap, chatID: chatID, userID: userID, msgChan: newMsgChan(0)}
	bot.msgMap.Store(fmt.Sprintf("%v:%v", chatID, userID), &msg)
	return &msg, nil
}

func (cb *Msg) Close() {
	cb.closeOnce.Do(func() {
		cb.cbkMap.Delete(fmt.Sprintf("%v:%v", cb.chatID, cb.userID))
		close(cb.msgChan)
	})
}

func (bot *BotAPI) CloseMsgChan(chatID, userID int64) error {
	if value, loaded := bot.msgMap.LoadAndDelete(fmt.Sprintf("%v:%v", chatID, userID)); loaded {
		value.(*Msg).Close()
		return nil
	} else {
		return errors.New("not found")
	}
}

func (cb *Msg) MsgChan() MsgChan {
	return cb.msgChan
}

func (cb *Msg) ChatID() int64 {
	return cb.chatID
}
