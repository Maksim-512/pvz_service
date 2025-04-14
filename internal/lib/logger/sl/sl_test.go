package sl_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"pvz_service/internal/lib/logger/sl"
)

func TestErr(t *testing.T) {
	err := errors.New("some error")
	attr := sl.Err(err)

	assert.Equal(t, "error", attr.Key)
	assert.Equal(t, "some error", attr.Value.String())
}

func TestOp(t *testing.T) {
	op := "some operation"
	attr := sl.Op(op)

	assert.Equal(t, "op", attr.Key)
	assert.Equal(t, "some operation", attr.Value.String())
}
