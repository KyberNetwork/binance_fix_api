package fix

import (
	"context"
	"crypto/ed25519"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/quickfixgo/field"
	"github.com/quickfixgo/quickfix"
	"go.uber.org/zap"
)

const logonTimeout = 30 * time.Second

type Config struct {
	APIKey             string
	PrivateKeyFilePath string
	Settings           *quickfix.Settings
}

type Options struct {
	messageHandling MessageHandling
	responseMode    ResponseMode
	fixLogFactory   quickfix.LogFactory
}

func defaultOpts() Options {
	return Options{
		messageHandling: MessageHandlingSequential,
		responseMode:    ResponseModeEverything,
		fixLogFactory:   quickfix.NewNullLogFactory(),
	}
}

type NewClientOption func(o *Options)

func WithMessageHandlingOpt(mh MessageHandling) NewClientOption {
	return func(o *Options) {
		o.messageHandling = mh
	}
}

func WithResponseModeOpt(rm ResponseMode) NewClientOption {
	return func(o *Options) {
		o.responseMode = rm
	}
}

func WithZapLogFactory(logger *zap.SugaredLogger) NewClientOption {
	return func(o *Options) {
		o.fixLogFactory = NewZapLogFactory(logger)
	}
}

type Client struct {
	l           *zap.SugaredLogger
	mu          sync.Mutex
	isConnected atomic.Bool
	initiator   *quickfix.Initiator
	pending     map[string]*call

	apiKey       string
	privateKey   ed25519.PrivateKey
	beginString  string
	targetCompID string
	senderCompID string

	options Options
}

func NewClient(ctx context.Context, l *zap.SugaredLogger, conf Config, opts ...NewClientOption) (*Client, error) {
	// Get BeginString, TargetCompID and SenderCompID from settings.
	if conf.Settings == nil {
		return nil, errors.New("empty quickfix settings")
	}

	globalSettings := conf.Settings.GlobalSettings()
	beginString, err := globalSettings.Setting("BeginString")
	if err != nil {
		l.Errorw("Failed to read BeginString from settings", "error", err)
		return nil, err
	}
	targetCompID, err := globalSettings.Setting("TargetCompID")
	if err != nil {
		l.Errorw("Failed to read TargetCompID from settings", "error", err)
		return nil, err
	}
	senderCompID, err := globalSettings.Setting("SenderCompID")
	if err != nil {
		l.Errorw("Failed to read SenderCompID from settings", "error", err)
		return nil, err
	}

	privateKey, err := GetEd25519PrivateKeyFromFile(conf.PrivateKeyFilePath)
	if err != nil {
		l.Errorw("Failed to GetEd25519PrivateKeyFromFile", "error", err)
		return nil, err
	}

	options := defaultOpts()
	for _, opt := range opts {
		opt(&options)
	}

	// Create a new Client object.
	client := &Client{
		l:            l,
		pending:      make(map[string]*call),
		apiKey:       conf.APIKey,
		privateKey:   privateKey,
		beginString:  beginString,
		targetCompID: targetCompID,
		senderCompID: senderCompID,
		options:      options,
	}

	// Init session and logon to Binance FIX API server.
	client.initiator, err = quickfix.NewInitiator(
		client,
		quickfix.NewMemoryStoreFactory(),
		conf.Settings,
		options.fixLogFactory,
	)
	if err != nil {
		client.l.Errorw("Failed to create new initiator", "error", err)
		return nil, err
	}

	err = client.Start(ctx)
	if err != nil {
		client.l.Errorw("Failed to start fix connection", "error", err)
		return nil, err
	}

	return client, nil
}

func (c *Client) Start(ctx context.Context) error {
	if err := c.initiator.Start(); err != nil {
		c.l.Errorw("Failed to initialize initiator", "error", err)
		return err
	}

	// Wait for the session to be authorized by the server.
	timeoutCtx, cancel := context.WithTimeout(ctx, logonTimeout)
	defer cancel()

	for {
		select {
		case <-timeoutCtx.Done():
			return errors.New("logon timed out")
		default:
			if c.IsConnected() {
				return nil
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func (c *Client) IsConnected() bool {
	return c.isConnected.Load()
}

// Stop closes underlying connection.
func (c *Client) Stop() {
	c.initiator.Stop()
}

// Call initiates a FIX call and wait for the response.
func (c *Client) Call(
	ctx context.Context, id string, msg *quickfix.Message,
) (*quickfix.Message, error) {
	call, err := c.send(id, msg)
	if err != nil {
		return nil, err
	}

	return call.wait(ctx)
}

func (c *Client) addCommonHeaders(msg *quickfix.Message) {
	msg.Header.Set(field.NewBeginString(c.beginString))
	msg.Header.Set(field.NewTargetCompID(c.targetCompID))
	msg.Header.Set(field.NewSenderCompID(c.senderCompID))
	msg.Header.Set(field.NewSendingTime(time.Now().UTC()))
}

func (c *Client) send(
	id string, msg *quickfix.Message,
) (waiter, error) {
	if !c.isConnected.Load() {
		return waiter{}, ErrClosed
	}

	c.addCommonHeaders(msg)
	cc := &call{request: msg, done: make(chan error, 1)}
	c.pending[id] = cc

	if err := quickfix.Send(msg); err != nil {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return waiter{}, err
	}

	return waiter{cc}, nil
}
