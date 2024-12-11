package main

import (
	"context"
	"os"
	"time"

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
	configFilePath = "./fix.conf"
	apiKey         = "YOUR_API_KEY_FOR_FIX_API"
	// Use your registered private key for Binance FIX API
	privateKeyFilePath = "./ed25519.pem"
	dropCopyFlag       = "Y"
)

// This is an example of how to use the Binance FIX API client for receiving drop-copy messages.
// It subscribes to the ExecutionReport<8> message and prints the order details to the console.
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
		logger, conf, fix.WithZapLogFactory(logger), fix.WithDropCopyFlagOpt(dropCopyFlag),
	)
	if err != nil {
		logger.Panicw("Failed to init client", "err", err)
	}

	logger.Info("Everything is ready!")

	// SUBSCRIBE TO EXECUTION REPORT
	client.SubscribeToExecutionReport(func(o *fix.Order) {
		logger.Infow("Received data from subscription", "order", o)
	})

	logger.Info("Subscribed to execution report!")

	time.Sleep(3 * time.Second)
}
