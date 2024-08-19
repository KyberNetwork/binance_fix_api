package fix

import (
	"github.com/quickfixgo/enum"
	"github.com/quickfixgo/field"
	"github.com/quickfixgo/quickfix"
	"go.uber.org/zap"
)

/* IMPLEMENT quickfix.Application INTERFACE */

// OnCreate implemented as part of Application interface.
func (c *Client) OnCreate(quickfix.SessionID) {}

// OnLogon notification of a session successfully logging on.
func (c *Client) OnLogon(quickfix.SessionID) {
	c.isConnected.Store(true)
	c.l.Info("Logon successfully!")
}

// OnLogout notification of a session logging off or disconnecting.
func (c *Client) OnLogout(quickfix.SessionID) {
	defer func() {
		if err := recover(); err != nil {
			c.l.Errorw("Recover from panic", "error", err)
		}
	}()

	c.isConnected.Store(false)
	c.l.Info("Logged out!")
	for _, call := range c.pending {
		call.done <- ErrClosed
		close(call.done)
	}
}

// ToAdmin notification of admin message being sent to target.
func (c *Client) ToAdmin(msg *quickfix.Message, _ quickfix.SessionID) {
	rawData := GetLogonRawData(c.privateKey, c.senderCompID, c.targetCompID, SendingTimeNow())

	msg.Body.Set(field.NewRawDataLength(len(rawData)))
	msg.Body.Set(field.NewRawData(rawData))
	msg.Body.Set(field.NewUsername(c.apiKey))
	msg.Body.Set(field.NewResetSeqNumFlag(true))
	msg.Body.SetInt(tagMessageHandling, int(c.options.messageHandling))
	msg.Body.SetInt(tagResponseMode, int(c.options.responseMode))
}

// ToApp notification of app message being sent to target.
func (c *Client) ToApp(msg *quickfix.Message, _ quickfix.SessionID) error {
	c.l.Infow("Sending message to server", "msg", msg)
	return nil
}

// FromAdmin notification of admin message being received from target.
func (c *Client) FromAdmin(msg *quickfix.Message, _ quickfix.SessionID) quickfix.MessageRejectError {
	c.l.Infow("FromAdmin message", "msg", msg)
	return nil
}

// FromApp notification of app message being received from target.
func (c *Client) FromApp(msg *quickfix.Message, s quickfix.SessionID) quickfix.MessageRejectError {
	// Process message according to message type.
	msgType, err := msg.MsgType()
	if err != nil {
		c.l.Errorw("Failed to get response message type", "error", err)
		return err
	}

	reqIDTag, err2 := getReqIDTagFromMsgType(enum.MsgType(msgType))
	if err2 != nil {
		c.l.Warnw("Could not get request ID tag", "msgType", msgType, "error", err2)
		return nil
	}

	id, err := msg.Body.GetString(reqIDTag)
	if err != nil {
		c.l.Errorw("Failed to get request ID", "tag", reqIDTag, "error", err)
		return err
	}

	c.mu.Lock()
	call := c.pending[id]
	delete(c.pending, id)
	c.mu.Unlock()

	if call != nil {
		c.l.Infow(
			"Matching response message",
			"id_tag", reqIDTag,
			"id", id,
			"request", call.request,
			"response", msg,
		)
		response, err2 := copyMessage(msg)
		if err2 != nil {
			c.l.Fatalw("Failed to copy response message", "error", err2)
		}
		call.response = response
		call.done <- nil
		close(call.done)
	}

	return nil
}

/* IMPLEMENT quickfix.Log INTERFACE */

type zapLog struct {
	logger *zap.SugaredLogger
}

func (l *zapLog) OnIncoming(data []byte) {
	l.logger.Infow("OnIncoming message", "data", string(data))
}

func (l *zapLog) OnOutgoing(data []byte) {
	l.logger.Infow("OnOutgoing message", "data", string(data))
}

func (l *zapLog) OnEvent(data string) {
	l.logger.Infow("OnEvent message", "data", data)
}

func (l *zapLog) OnEventf(data string, params ...interface{}) {
	l.logger.Infow("OnEventf message", "data", data, "params", params)
}

type zapLogFactory struct {
	logger *zap.SugaredLogger
}

func (f *zapLogFactory) Create() (quickfix.Log, error) {
	return &zapLog{f.logger}, nil
}

func (f *zapLogFactory) CreateSessionLog(sessionID quickfix.SessionID) (quickfix.Log, error) {
	return &zapLog{f.logger}, nil
}

func NewZapLogFactory(logger *zap.SugaredLogger) *zapLogFactory {
	return &zapLogFactory{logger}
}
