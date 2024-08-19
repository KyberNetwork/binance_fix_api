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

/*
	switch msgType {
		case enum.MsgType_SECURITY_LIST:
			return tag.SecurityReqID, nil
		case enum.MsgType_MARKET_DATA_REQUEST:
			return tag.MDReqID, nil
		case enum.MsgType_MARKET_DATA_REQUEST_REJECT:
			return tag.MDReqID, nil
		case enum.MsgType_MARKET_DATA_SNAPSHOT_FULL_REFRESH:
			return tag.MDReqID, nil
		case enum.MsgType_MARKET_DATA_INCREMENTAL_REFRESH:
			return tag.MDReqID, nil
		case enum.MsgType_EXECUTION_REPORT:
			return tag.OrigClOrdID, nil
		case enum.MsgType_ORDER_CANCEL_REJECT:
			return tag.ClOrdID, nil
		case enum.MsgType_ORDER_MASS_CANCEL_REPORT:
			return tag.OrderID, nil
		case enum.MsgType_POSITION_REPORT:
			return tag.PosReqID, nil
		case enum.MsgType_USER_RESPONSE:
			return tag.UserRequestID, nil
		case enum.MsgType_SECURITY_STATUS:
			return tag.SecurityStatusReqID, nil
	}
*/
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
