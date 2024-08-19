package fix

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogonRawData(t *testing.T) {
	privateKey, err := GetEd25519PrivateKeyFromFile("./sample/ed25519.pem")
	if err != nil {
		require.NoError(t, err)
	}

	// This test validates with Binance API doc.
	assert.Equal(
		t,
		"4MHXelVVcpkdwuLbl6n73HQUXUf1dse2PCgT1DYqW9w8AVZ1RACFGM+5UdlGPrQHrgtS3CvsRURC1oj73j8gCA==",
		GetLogonRawData(privateKey, "EXAMPLE", "SPOT", "20240627-11:17:25.223"),
	)
}
