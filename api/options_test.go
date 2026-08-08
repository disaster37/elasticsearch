package api

import (
	"encoding/json"
	"testing"

	jsoniter "github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
)

// ---- rawBody ----

func TestUnitRawBody(t *testing.T) {
	rb := rawBody(`{"f":"v"}`)
	assert.Equal(t, jsoniter.RawMessage(`{"f":"v"}`), rb)
	// A nil string still produces an (empty) RawMessage, not nil.
	assert.NotNil(t, rawBody(""))
}

func TestUnitRawBodyNil(t *testing.T) {
	// rawBody wraps any string as json.RawMessage; passing the zero string
	// yields an empty (but non-nil) RawMessage.
	var s string
	rb := rawBody(s)
	assert.Equal(t, jsoniter.RawMessage(""), rb)
}

// ---- isNilValue ----

func TestUnitIsNilValueNilPointer(t *testing.T) {
	var p *int
	assert.True(t, isNilValue(p))
}

func TestUnitIsNilValueNonNilPointer(t *testing.T) {
	n := 5
	p := &n
	assert.False(t, isNilValue(p))
}

func TestUnitIsNilValueNilInterface(t *testing.T) {
	var i interface{}
	assert.False(t, isNilValue(i)) // untyped nil is not a typed nil
}

func TestUnitIsNilValueTypedNilInterface(t *testing.T) {
	var p *int
	var i interface{} = p // typed nil
	assert.True(t, isNilValue(i))
}

func TestUnitIsNilValueNonPointer(t *testing.T) {
	assert.False(t, isNilValue(42))
	assert.False(t, isNilValue("foo"))
	assert.False(t, isNilValue(struct{}{}))
}

// Ensure the standard library json.RawMessage is compatible with the
// goccy RawMessage returned by rawBody (they share the same underlying type).
func TestUnitRawBodyJSONCompat(t *testing.T) {
	rb := rawBody(`{"a":1}`)
	var m map[string]int
	assert.NoError(t, json.Unmarshal(rb, &m))
	assert.Equal(t, 1, m["a"])
}
