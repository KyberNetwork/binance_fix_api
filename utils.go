package fix

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/quickfixgo/enum"
	"github.com/quickfixgo/field"
	"github.com/quickfixgo/quickfix"
)

const (
	utcTimestampMillisFmt = "20060102-15:04:05.000"
	utcTimestampMicrosFmt = "20060102-15:04:05.000000"
	blockTypePrivateKey   = "PRIVATE KEY"
)

var (
	ErrClosed = errors.New("connection is closed")

	ErrNilPrivateKeyValue  = errors.New("nil private key value")
	ErrInvalidEd25519Key   = errors.New("invalid key ed25519 key")
	ErrInvalidRequestIDTag = errors.New("request id tag not found")
)

func ParseEd25519PrivateKey(data []byte) (ed25519.PrivateKey, error) {
	block, _ := pem.Decode((data))
	if block == nil || block.Type != blockTypePrivateKey {
		return nil, ErrNilPrivateKeyValue
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	ret, ok := privateKey.(ed25519.PrivateKey)
	if !ok {
		return nil, ErrInvalidEd25519Key
	}

	return ret, nil
}

func GetEd25519PrivateKeyFromFile(path string) (ed25519.PrivateKey, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return ParseEd25519PrivateKey(data)
}

func GetLogonRawData(
	privateKey ed25519.PrivateKey,
	senderCompID, targetCompID, sendingTime string,
) string {
	method := string(enum.MsgType_LOGON)
	msgSeqNum := "1" // Logon is the first request of fix protocol.
	payload := strings.Join([]string{method, senderCompID, targetCompID, msgSeqNum, sendingTime}, "\x01")
	data := ed25519.Sign(privateKey, []byte(payload))

	return base64.StdEncoding.EncodeToString(data)
}

func SendingTimeNow() string {
	return time.Now().UTC().Format(utcTimestampMillisFmt)
}

func copyMessage(msg *quickfix.Message) (*quickfix.Message, error) {
	out := quickfix.NewMessage()
	err := quickfix.ParseMessage(out, bytes.NewBufferString(msg.String()))
	if err != nil {
		return nil, err
	}
	return out, nil
}

type call struct {
	request  *quickfix.Message
	response *quickfix.Message
	done     chan error
}

type waiter struct {
	*call
}

// wait for the response message of an ongoing FIX call.
func (w waiter) wait(ctx context.Context) (*quickfix.Message, error) {
	select {
	case err, ok := <-w.call.done:
		if !ok {
			err = ErrClosed
		}
		if err != nil {
			return nil, err
		}
		return w.call.response, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func LoadQuickfixSettings(filePath string) (*quickfix.Settings, error) {
	cfg, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer cfg.Close()

	data, err := io.ReadAll(cfg)
	if err != nil {
		return nil, err
	}

	appSettings, err := quickfix.ParseSettings(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	return appSettings, nil
}

func floatToString(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

type Order struct {
	Symbol            string
	OrderID           int64
	ClientOrderID     string
	Price             float64
	OrderQty          float64
	CumQty            float64
	CumQuoteQty       float64
	Status            OrderStatus
	TimeInForce       TimeInForce
	Type              OrderType
	Side              SideType
	IcebergQuantity   float64
	TransactTime      time.Time // Timestamp when this event occurred.
	OrderCreationTime time.Time
	WorkingTime       time.Time // When this order appeared on the order book.
}

func decodeExecutionReport(msg *quickfix.Message) (Order, error) {
	status, err := getOrderStatus(msg)
	if err != nil {
		return Order{}, err
	}

	if status == OrderStatusRejected {
		reason, err := getText(msg)
		if err != nil {
			return Order{}, err
		}
		if reason != "" {
			return Order{}, errors.New(reason)
		}
	}

	symbol, err := getSymbol(msg)
	if err != nil {
		return Order{}, err
	}

	orderID, err := getOrderID(msg)
	if err != nil {
		return Order{}, err
	}

	clientOrderID, err := getClientOrderID(msg)
	if err != nil {
		return Order{}, err
	}

	price, err := getPrice(msg)
	if err != nil {
		return Order{}, err
	}

	orderQty, err := getOrderQty(msg)
	if err != nil {
		return Order{}, err
	}

	cumQty, err := getCumQty(msg)
	if err != nil {
		return Order{}, err
	}

	cumQuoteQty, err := getCumQuoteQty(msg)
	if err != nil {
		return Order{}, err
	}

	timeInForce, err := getTimeInForce(msg)
	if err != nil {
		return Order{}, err
	}

	orderType, err := getOrdType(msg)
	if err != nil {
		return Order{}, err
	}

	side, err := getSide(msg)
	if err != nil {
		return Order{}, err
	}

	maxFloor, err := getMaxFloor(msg)
	if err != nil {
		return Order{}, err
	}

	transactTime, err := getTransactTime(msg)
	if err != nil {
		return Order{}, err
	}

	orderCreationTime, err := getOrderCreationTime(msg)
	if err != nil {
		return Order{}, err
	}

	workingTime, err := getWorkingTime(msg)
	if err != nil {
		return Order{}, err
	}

	return Order{
		Symbol:            symbol,
		OrderID:           orderID,
		ClientOrderID:     clientOrderID,
		Price:             price,
		OrderQty:          orderQty,
		CumQty:            cumQty,
		CumQuoteQty:       cumQuoteQty,
		Status:            status,
		TimeInForce:       timeInForce,
		Type:              orderType,
		Side:              side,
		IcebergQuantity:   maxFloor,
		TransactTime:      transactTime,
		OrderCreationTime: orderCreationTime,
		WorkingTime:       workingTime,
	}, nil
}

func getText(msg *quickfix.Message) (v string, err error) {
	var f field.TextField
	if msg.Body.Has(f.Tag()) {
		if err = msg.Body.Get(&f); err == nil {
			v = f.Value()
		}
	}
	return
}

func getSymbol(msg *quickfix.Message) (v string, err error) {
	var f field.SymbolField
	if err = msg.Body.Get(&f); err == nil {
		v = f.Value()
	}
	return
}

func getOrderID(msg *quickfix.Message) (v int64, err error) {
	var f field.OrderIDField
	if msg.Body.Has(f.Tag()) {
		if err = msg.Body.Get(&f); err != nil {
			return
		}
	}

	return strconv.ParseInt(f.Value(), 10, 64)
}

func getClientOrderID(msg *quickfix.Message) (v string, err error) {
	var f field.ClOrdIDField
	if msg.Body.Has(f.Tag()) {
		if err = msg.Body.Get(&f); err == nil {
			v = f.Value()
		}
	}
	return
}

func getOrderStatus(msg *quickfix.Message) (v OrderStatus, err error) {
	var f field.OrdStatusField
	if err = msg.Body.Get(&f); err == nil {
		v = mappedOrderStatus[f.Value()]
	}
	return
}

func getOrdType(msg *quickfix.Message) (v OrderType, err error) {
	var f field.OrdTypeField
	if err = msg.Body.Get(&f); err == nil {
		v = mappedOrderType[f.Value()]
	}
	return
}

func getSide(msg *quickfix.Message) (v SideType, err error) {
	var f field.SideField
	if err = msg.Body.Get(&f); err == nil {
		v = mappedSideType[f.Value()]
	}
	return
}

func getTimeInForce(msg *quickfix.Message) (v TimeInForce, err error) {
	var f field.TimeInForceField
	if msg.Body.Has(f.Tag()) {
		if err = msg.Body.Get(&f); err == nil {
			v = mappedTimeInForce[f.Value()]
		}
	}
	return
}

func getPrice(msg *quickfix.Message) (float64, error) {
	var f field.PriceField
	if msg.Body.Has(f.Tag()) {
		if err := msg.Body.Get(&f); err != nil {
			return 0, err
		}
		return f.InexactFloat64(), nil
	}
	return 0, nil
}

func getOrderQty(msg *quickfix.Message) (float64, error) {
	var f field.OrderQtyField
	if msg.Body.Has(f.Tag()) {
		if err := msg.Body.Get(&f); err != nil {
			return 0, err
		}
		return f.InexactFloat64(), nil
	}
	return 0, nil
}

func getCumQty(msg *quickfix.Message) (float64, error) {
	var f field.CumQtyField
	if msg.Body.Has(f.Tag()) {
		if err := msg.Body.Get(&f); err != nil {
			return 0, err
		}
		return f.InexactFloat64(), nil
	}
	return 0, nil
}

func getCumQuoteQty(msg *quickfix.Message) (float64, error) {
	if msg.Body.Has(tagCumQuoteQty) {
		str, err := msg.Body.GetString(tagCumQuoteQty)
		if err != nil {
			return 0, err
		}
		return strconv.ParseFloat(str, 64)
	}
	return 0, nil
}

func getMaxFloor(msg *quickfix.Message) (float64, error) {
	var f field.MaxFloorField
	if msg.Body.Has(f.Tag()) {
		if err := msg.Body.Get(&f); err != nil {
			return 0, err
		}
		return f.InexactFloat64(), nil
	}
	return 0, nil
}

func getTransactTime(msg *quickfix.Message) (v time.Time, err error) {
	var f field.TransactTimeField
	if msg.Body.Has(f.Tag()) {
		if err = msg.Body.Get(&f); err == nil {
			v = f.Value()
		}
	}
	return
}

func getOrderCreationTime(msg *quickfix.Message) (time.Time, error) {
	if msg.Body.Has(tagOrderCreationTime) {
		str, err := msg.Body.GetString(tagOrderCreationTime)
		if err != nil {
			return time.Time{}, err
		}
		return time.Parse(utcTimestampMicrosFmt, str)
	}
	return time.Time{}, nil
}

func getWorkingTime(msg *quickfix.Message) (time.Time, error) {
	if msg.Body.Has(tagWorkingTime) {
		str, err := msg.Body.GetString(tagWorkingTime)
		if err != nil {
			return time.Time{}, err
		}
		return time.Parse(utcTimestampMicrosFmt, str)
	}
	return time.Time{}, nil
}
