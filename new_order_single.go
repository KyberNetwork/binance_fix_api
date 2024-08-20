package fix

import (
	"context"

	"github.com/google/uuid"
	"github.com/quickfixgo/enum"
	"github.com/quickfixgo/field"
	"github.com/quickfixgo/quickfix"
	"github.com/quickfixgo/tag"
	"go.uber.org/zap"
)

/*
Tag     Name                    Type    Required    Description
11.     ClOrdID                 STRING  Y           ClOrdID to be assigned to the order.
38.     OrderQty                QTY     N           Quantity of the order
40.     OrdType                 CHAR    Y           1: MARKET, 2: LIMIT, 3: STOP, 4: STOP_LIMIT
18      ExecInst                CHAR    N           6: PARTICIPATE_DONT_INITIATE
44.     Price                   PRICE   N           Price of the order
54.     Side                    CHAR    Y           1: BUY, 2: SELL
55.     Symbol                  STRING  Y           Symbol to place the order on.
59.     TimeInForce             CHAR    N           1: GOOD_TILL_CANCEL, 3: IMMEDIATE_OR_CANCEL, 4: FILL_OR_KILL
111     MaxFloor                QTY     N           Used for iceberg orders, this specifies the visible quantity of the order on the book.
152     CashOrderQty            QTY     N           Quantity of the order specified in the quote asset units, for reverse market orders.
847     TargetStrategy          INT     N
7940    StrategyID              INT     N           The value cannot be less than 1000000.
25001   SelfTradePreventionMode CHAR    N           1: NONE, 2: EXPIRE_TAKER, 3: EXPIRE_MAKER, 4: EXPIRE_BOTH
1100    TriggerType             CHAR    N           4: PRICE_MOVEMENT
1101    TriggerAction           CHAR    N           1: ACTIVATE
1102    TriggerPrice            PRICE   N           Activation price for contingent orders. See table
1107    TriggerPriceType        CHAR    N           2: LAST_TRADE
1109    TriggerPriceDirection   CHAR    N           U or D
25009   TriggerTrailingDeltaBps INT     N           Provide to create trailing orders.
25032   SOR                     BOOLEAN N           Whether to activate SOR for this order.
*/

// NewOrderSingleService uses uuid to generate unique ClOrdID.
type NewOrderSingleService struct {
	c           *Client
	symbol      string
	side        enum.Side
	orderType   enum.OrdType
	timeInForce *enum.TimeInForce
	quantity    *float64
	price       *float64
}

func (c *Client) NewOrderSingleService() *NewOrderSingleService {
	return &NewOrderSingleService{
		c: c,
	}
}

// Symbol set symbol
func (s *NewOrderSingleService) Symbol(symbol string) *NewOrderSingleService {
	s.symbol = symbol
	return s
}

// Side set side
func (s *NewOrderSingleService) Side(side enum.Side) *NewOrderSingleService {
	s.side = side
	return s
}

// Type set type
func (s *NewOrderSingleService) Type(orderType enum.OrdType) *NewOrderSingleService {
	s.orderType = orderType
	return s
}

// TimeInForce set timeInForce
func (s *NewOrderSingleService) TimeInForce(timeInForce enum.TimeInForce) *NewOrderSingleService {
	s.timeInForce = &timeInForce
	return s
}

// Quantity set quantity
func (s *NewOrderSingleService) Quantity(quantity float64) *NewOrderSingleService {
	s.quantity = &quantity
	return s
}

// Price set price
func (s *NewOrderSingleService) Price(price float64) *NewOrderSingleService {
	s.price = &price
	return s
}

func (s *NewOrderSingleService) Do(ctx context.Context) (Order, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return Order{}, err
	}

	msg := quickfix.NewMessage()
	msg.Header.Set(field.NewMsgType(enum.MsgType_ORDER_SINGLE))

	msg.Body.Set(field.NewClOrdID(id.String()))
	msg.Body.Set(field.NewSymbol(s.symbol))
	msg.Body.Set(field.NewSide(s.side))
	msg.Body.Set(field.NewOrdType(s.orderType))
	if s.quantity != nil {
		msg.Body.SetString(tag.OrderQty, floatToString(*s.quantity))
	}
	if s.price != nil {
		msg.Body.SetString(tag.Price, floatToString(*s.price))
	}
	if s.timeInForce != nil {
		msg.Body.Set(field.NewTimeInForce(*s.timeInForce))
	}

	resp, err := s.c.Call(ctx, id.String(), msg)
	if err != nil {
		zap.S().Errorw("Failed to create new order", "request", msg, "err", err)
		return Order{}, err
	}

	order, err := decodeExecutionReport(resp)
	if err != nil {
		zap.S().Errorw("Failed to decode ExecutionReport message", "request", msg, "response", resp, "error", err)
		return Order{}, err
	}

	return order, nil
}
