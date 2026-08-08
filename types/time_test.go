package types

import (
	"testing"
	"time"

	json "github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnixMilliTimeUnmarshal(t *testing.T) {
	var ut UnixMilliTime
	err := json.Unmarshal([]byte(`1700000000000`), &ut)
	assert.NoError(t, err)
	assert.False(t, ut.IsZero())
	expected := time.Unix(0, 1700000000000*int64(time.Millisecond))
	assert.True(t, ut.Equal(expected))
}

func TestUnixMilliTimeUnmarshalZero(t *testing.T) {
	var ut UnixMilliTime
	err := json.Unmarshal([]byte(`0`), &ut)
	assert.NoError(t, err)
	assert.True(t, ut.IsZero())
}

func TestUnixMilliTimeUnmarshalNull(t *testing.T) {
	var ut UnixMilliTime
	err := json.Unmarshal([]byte(`null`), &ut)
	assert.NoError(t, err)
	assert.True(t, ut.IsZero())
}

// TestUnixMilliTimeUnmarshalEmptyString verifies UnmarshalJSON handles an
// empty string.
func TestUnixMilliTimeUnmarshalEmptyString(t *testing.T) {
	var ut UnixMilliTime
	err := json.Unmarshal([]byte(`""`), &ut)
	assert.NoError(t, err)
	assert.True(t, ut.IsZero())
}

func TestUnixMilliTimeMarshal(t *testing.T) {
	ut := UnixMilliTime{Time: time.Unix(0, 1700000000000*int64(time.Millisecond))}
	b, err := json.Marshal(ut)
	assert.NoError(t, err)
	assert.Equal(t, `1700000000000`, string(b))
}

func TestUnixMilliTimeMarshalZero(t *testing.T) {
	ut := UnixMilliTime{}
	b, err := json.Marshal(ut)
	assert.NoError(t, err)
	assert.Equal(t, `0`, string(b))
}

// TestUnixMilliTimeRoundTrip verifies marshal→unmarshal preserves the value.
func TestUnixMilliTimeRoundTrip(t *testing.T) {
	original := UnixMilliTime{Time: time.Unix(0, 1700000000123*int64(time.Millisecond))}
	b, err := json.Marshal(original)
	require.NoError(t, err)
	var round UnixMilliTime
	require.NoError(t, json.Unmarshal(b, &round))
	assert.True(t, original.Equal(round.Time))
}

// TestUnixMilliTimeRoundTripZero verifies the zero value round-trips to 0.
func TestUnixMilliTimeRoundTripZero(t *testing.T) {
	original := UnixMilliTime{}
	b, err := json.Marshal(original)
	require.NoError(t, err)
	assert.Equal(t, `0`, string(b))
	var round UnixMilliTime
	require.NoError(t, json.Unmarshal(b, &round))
	assert.True(t, round.IsZero())
}

// TestUnixMilliTimeBoundaryEpoch verifies the Unix epoch (0 ms) round-trips.
func TestUnixMilliTimeBoundaryEpoch(t *testing.T) {
	ut := UnixMilliTime{Time: time.Unix(0, 0)}
	b, err := json.Marshal(ut)
	require.NoError(t, err)
	// time.Unix(0,0).IsZero() is true → marshals to 0.
	assert.Equal(t, `0`, string(b))
}

// TestUnixMilliTimeBoundaryOneMilli verifies 1 ms after the epoch.
func TestUnixMilliTimeBoundaryOneMilli(t *testing.T) {
	ut := UnixMilliTime{Time: time.Unix(0, int64(time.Millisecond))}
	b, err := json.Marshal(ut)
	require.NoError(t, err)
	assert.Equal(t, `1`, string(b))
	var round UnixMilliTime
	require.NoError(t, json.Unmarshal(b, &round))
	assert.True(t, ut.Equal(round.Time))
}

// TestUnixMilliTimeUnmarshalInvalid verifies UnmarshalJSON returns an error
// for a non-integer, non-null value.
func TestUnixMilliTimeUnmarshalInvalid(t *testing.T) {
	var ut UnixMilliTime
	err := json.Unmarshal([]byte(`"not a number"`), &ut)
	assert.Error(t, err)
}

// TestUnixMilliTimeUnmarshalNegative verifies UnmarshalJSON handles a
// negative timestamp.
func TestUnixMilliTimeUnmarshalNegative(t *testing.T) {
	var ut UnixMilliTime
	err := json.Unmarshal([]byte(`-1`), &ut)
	assert.NoError(t, err)
	assert.False(t, ut.IsZero())
}
