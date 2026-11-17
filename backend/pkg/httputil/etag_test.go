package httputil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseIfMatch(t *testing.T) {
	tests := []struct {
		in      string
		want    int32
		wantErr error
	}{
		{`"7"`, 7, nil},
		{` "12" `, 12, nil},
		{``, 0, ErrMissingIfMatch},
		{`7`, 0, ErrInvalidIfMatch},
		{`W/"7"`, 0, ErrInvalidIfMatch},
		{`"0"`, 0, ErrInvalidIfMatch},
		{`"abc"`, 0, ErrInvalidIfMatch},
		{`"99999999999"`, 0, ErrInvalidIfMatch},
	}
	for _, tt := range tests {
		got, err := ParseIfMatch(tt.in)
		assert.ErrorIs(t, err, tt.wantErr, tt.in)
		assert.Equal(t, tt.want, got, tt.in)
	}
	assert.Equal(t, `"7"`, ETag(7))
}
