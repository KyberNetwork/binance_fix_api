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
	"strings"
	"time"

	"github.com/quickfixgo/enum"
	"github.com/quickfixgo/quickfix"
)

const (
	sendingTimeFmt      = "20060102-15:04:05.000"
	blockTypePrivateKey = "PRIVATE KEY"
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
	return time.Now().UTC().Format(sendingTimeFmt)
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
