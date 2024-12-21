package entities_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/entities"
)

func Test_NewCoin_Error(t *testing.T) {
	t.Parallel()

	// error checking
	c, err := entities.NewCoin("", 1.1)

	require.ErrorIs(t, err, entities.ErrInvalidParam)
	require.Contains(t, err.Error(), "coin name is empty")
	require.Nil(t, c)

	c, err = entities.NewCoin("BTC", -12.3)

	require.ErrorIs(t, err, entities.ErrInvalidParam)
	require.Contains(t, err.Error(), "negative coin rate")
	require.Nil(t, c)

	// checking the values ​​are correct
	c, err = entities.NewCoin("ETH", 250.467)

	require.Contains(t, c.Title, "ETH")
	require.EqualValues(t, c.Cost, 250.467)
	require.Nil(t, err)

	c, err = entities.NewCoin("BTC", 0.127)

	require.Contains(t, c.Title, "BTC")
	require.EqualValues(t, c.Cost, 0.127)
	require.Nil(t, err)
}

// assert.EqualValues(t, uint32(123), int32(123))
// assert.Exactly(t, int32(123), int64(123))
// assert.InDelta(t, math.Pi, 22/7.0, 0.01)
