package fix

import (
	"github.com/quickfixgo/enum"
	"github.com/quickfixgo/quickfix"
	"github.com/quickfixgo/tag"
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

	tagCumQuoteQty       quickfix.Tag = 25017
	tagOrderCreationTime quickfix.Tag = 25018
	tagWorkingTime       quickfix.Tag = 25023
)

const (
	msgType_LIMIT_REQUEST  enum.MsgType = "XLQ"
	msgType_LIMIT_RESPONSE enum.MsgType = "XLR"
)

var mappedMsgTypeTag = map[enum.MsgType]quickfix.Tag{
	msgType_LIMIT_RESPONSE:        tagGetLimitReqID,
	enum.MsgType_EXECUTION_REPORT: tag.ClOrdID,
}

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

type OrderStatus string

const (
	OrderStatusNew             OrderStatus = "NEW"
	OrderStatusPartiallyFilled OrderStatus = "PARTIALLY_FILLED"
	OrderStatusFilled          OrderStatus = "FILLED"
	OrderStatusCanceled        OrderStatus = "CANCELED"
	OrderStatusPendingCancel   OrderStatus = "PENDING_CANCEL"
	OrderStatusRejected        OrderStatus = "REJECTED"
	OrderStatusPendingNew      OrderStatus = "PENDING_NEW"
	OrderStatusExpired         OrderStatus = "EXPIRED"
)

var mappedOrderStatus = map[enum.OrdStatus]OrderStatus{
	enum.OrdStatus_NEW:              OrderStatusNew,
	enum.OrdStatus_PARTIALLY_FILLED: OrderStatusPartiallyFilled,
	enum.OrdStatus_FILLED:           OrderStatusFilled,
	enum.OrdStatus_CANCELED:         OrderStatusCanceled,
	enum.OrdStatus_PENDING_CANCEL:   OrderStatusPendingCancel,
	enum.OrdStatus_REJECTED:         OrderStatusRejected,
	enum.OrdStatus_PENDING_NEW:      OrderStatusPendingNew,
	enum.OrdStatus_EXPIRED:          OrderStatusExpired,
}

type TimeInForce string

const (
	TimeInForceGTC TimeInForce = "GOOD_TILL_CANCEL"
	TimeInForceIOC TimeInForce = "IMMEDIATE_OR_CANCEL"
	TimeInForceFOK TimeInForce = "FILL_OR_KILL"
)

var mappedTimeInForce = map[enum.TimeInForce]TimeInForce{
	enum.TimeInForce_GOOD_TILL_CANCEL:    TimeInForceGTC,
	enum.TimeInForce_IMMEDIATE_OR_CANCEL: TimeInForceIOC,
	enum.TimeInForce_FILL_OR_KILL:        TimeInForceFOK,
}

type OrderType string

const (
	OrderTypeMarket    OrderType = "MARKET"
	OrderTypeLimit     OrderType = "LIMIT"
	OrderTypeStop      OrderType = "STOP"
	OrderTypeStopLimit OrderType = "STOP_LIMIT"
)

var mappedOrderType = map[enum.OrdType]OrderType{
	enum.OrdType_MARKET:     OrderTypeMarket,
	enum.OrdType_LIMIT:      OrderTypeLimit,
	enum.OrdType_STOP:       OrderTypeStop,
	enum.OrdType_STOP_LIMIT: OrderTypeStopLimit,
}

type SideType string

const (
	SideTypeBuy  SideType = "BUY"
	SideTypeSell SideType = "SELL"
)

var mappedSideType = map[enum.Side]SideType{
	enum.Side_BUY:  SideTypeBuy,
	enum.Side_SELL: SideTypeSell,
}
