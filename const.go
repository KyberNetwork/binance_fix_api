package fix

import (
	"github.com/quickfixgo/enum"
	"github.com/quickfixgo/quickfix"
)

const (
	tagMessageHandling quickfix.Tag = 25035
	tagResponseMode    quickfix.Tag = 25036
	tagGetLimitReqID   quickfix.Tag = 6136

	tagNoLimitIndicators            quickfix.Tag = 25003
	tagLimitType                    quickfix.Tag = 25004
	tagLimitCount                   quickfix.Tag = 25005
	tagLimitMax                     quickfix.Tag = 25006
	tagLimitResetInterval           quickfix.Tag = 25007
	tagLimitResetIntervalResolution quickfix.Tag = 25008
)

const (
	msgType_LIMIT_REQUEST  enum.MsgType = "XLQ"
	msgType_LIMIT_RESPONSE enum.MsgType = "XLR"
)

var (
	mappedMsgTypeTag = map[enum.MsgType]quickfix.Tag{
		msgType_LIMIT_RESPONSE: tagGetLimitReqID,
	}
)

func getReqIDTagFromMsgType(msgType enum.MsgType) (quickfix.Tag, error) {
	if tag, ok := mappedMsgTypeTag[msgType]; ok {
		return tag, nil
	}

	return 0, ErrInvalidRequestIDTag
}

type MessageHandling int

const (
	MessageHandlingUnordered  MessageHandling = 1
	MessageHandlingSequential MessageHandling = 2
)

type ResponseMode int

const (
	ResponseModeEverything ResponseMode = 1
	ResponseModeOnlyAcks   ResponseMode = 2
)
