package restretries

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrors(t *testing.T) {
	err := NewPermanentError(nil)
	assert.Error(t, err)
	assert.Equal(t, errMustStop, err.(interface{ Unwrap() []error }).Unwrap()[0])
	assert.True(t, IsPermanentError(err))

	err = NewPermanentError(context.Canceled)
	assert.Error(t, err)
	assert.True(t, IsPermanentError(err))
	assert.True(t, errors.Is(err, context.Canceled))
	assert.True(t, errors.Is(err, errMustStop))
}
