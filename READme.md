## Binance FIX API

Prepare those environment variables:

- configFilePath: FIX protocol config (see file path `sample/fix.conf`)
- apiKey: Binance FIX API KEY
- privateKeyFilePath: Binance Ed25519 pem file.

```go
package main

import (
	"context"
	"os"

	fix "github.com/KyberNetwork/binance_fix_api"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func SetupLogger() *zap.SugaredLogger {
	pConf := zap.NewProductionEncoderConfig()
	pConf.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder := zapcore.NewConsoleEncoder(pConf)
	level := zap.NewAtomicLevelAt(zap.DebugLevel)
	l := zap.New(zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level), zap.AddCaller())
	zap.ReplaceGlobals(l)
	return zap.S()
}

const (
	configFilePath     = "./sample/fix.conf"
	apiKey             = "your_api_key"
	privateKeyFilePath = "your_ed25519_key_pem"
)

func main() {
	logger := SetupLogger()
	logger.Infow("This is an fix-client example")
	settings, err := fix.LoadQuickfixSettings(configFilePath)
	if err != nil {
		logger.Panicw("Failed to LoadQuickfixSettings", "err", err)
	}

	conf := fix.Config{
		APIKey:             apiKey,
		PrivateKeyFilePath: privateKeyFilePath,
		Settings:           settings,
	}
	client, err := fix.NewClient(
		context.Background(),
		logger, conf, fix.WithZapLogFactory(logger),
	)
	if err != nil {
		logger.Panicw("Failed to init client", "err", err)
	}

	logger.Info("Everything is ready!")

	// GET LIMIT MESSAGE
	limit, err := client.NewGetLimitService().Do(context.Background())
	if err != nil {
		logger.Panicw("Failed to get LimitMessages", "err", err)
	}
	logger.Infow("Get limit message", "data", limit)

	// SUBSCRIBE TO EXECUTION REPORT
	client.SubscribeToExecutionReport(func(o *fix.Order) {
		logger.Infow("Received data from subscription", "order", o)
	})

	logger.Info("Subscribed to execution report!")

	// TRY TO PLACE ORDER
	time.Sleep(time.Second)
	order, err := client.NewOrderSingleService().
		Symbol("BNBUSDT").
		Side(enum.Side_BUY).
		Type(enum.OrdType_LIMIT).
		TimeInForce(enum.TimeInForce_GOOD_TILL_CANCEL).
		Quantity(0.01).
		Price(502).
		Do(context.Background())

	logger.Infow("NewOrderSingleService resp", "order", order, "err", err)

	time.Sleep(3 * time.Second)
}

```

## Order Entry Messages

1. ✅ `NewOrderSingle<D>`
   - Sent by the client to submit a new order for execution.
2. 🚫 `NewOrderList<E>`
   - Sent by the client to submit a list of orders for execution.
3. 🚫 `OrderCancelRequest<F>`
   - Sent by the client to cancel an order or an order list.
4. 🚫 `OrderCancelRequestAndNewOrderSingle<XCN>`
   - Sent by the client to cancel an order and submit a new one for execution.
5. 🚫 `OrderMassCancelRequest<q>`
   - Sent by the client to cancel all open orders on a symbol.
6. ✅ `ExecutionReport<8>`
   - Sent by the server whenever an order state changes.
7. 🚫 `OrderCancelReject<9>`
   - Sent by the server when OrderCancelRequest<F> has failed.
8. 🚫 `OrderMassCancelReport<r>`
   - Sent by the server in response to OrderMassCancelRequest<q>.
9. 🚫 `ListStatus<N>`
   - Sent by the server whenever an order list state changes.

## Limit message

- ✅ Sent by the client to query current limits.
- ✅ Sent by the server in response to LimitQuery<XLQ>.
