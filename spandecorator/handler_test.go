package spandecorator

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SKF/go-utility/v2/useridcontext"
	"github.com/SKF/go-utility/v2/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SKF/go-enlight-middleware/spandecorator/internal"
)

func TestGetAttributes(t *testing.T) {
	// ARRANGE
	userID := uuid.New()
	ctx := useridcontext.NewContext(context.Background(), userID.String())

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("My-Header", "apa")
	request.Header.Set("X-Forwarded-For", "myIp")

	request = request.WithContext(ctx)

	// ACT
	attrs := extractAttributes(request)

	// ASSERT
	require.Equal(t, "", attrs["authorization"])
	require.Equal(t, userID.String(), attrs[internal.UserIDKey])
	require.Equal(t, "apa", attrs["header.My-Header"])
	require.Equal(t, "", attrs["X-Forwarded-For"])
}

func TestIgnoreStrangeCaseAuthorization(t *testing.T) {
	// ARRANGE
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("aUthOrization", "sensitive token")

	// ACT
	attrs := extractAttributes(request)

	// ASSERT
	require.Len(t, attrs, 0)
}

func TestWithBody_Happy(t *testing.T) {
	// ARRANGE
	jsonStr := `{
		"isTest": true,
		"aNumber": 123,
		"is": "5f2394d2-c07e-4bca-85b8-4441dcb8eb27",
		"nested": {
			"isTest": true,
			"aNumber": 123,
			"is": "5f2394d2-c07e-4bca-85b8-4441dcb8eb27"
		}
	}`
	request := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(jsonStr))
	span := testSpan{make(map[string]string)}

	// ACT
	err1 := decorateWithBody(request, &span)
	forwardedBody, err2 := io.ReadAll(request.Body)

	// ASSERT
	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.Equal(t, jsonStr, span.attributes["http.request.body"])
	assert.Equal(t, jsonStr, string(forwardedBody))
}

func TestWithBody_OverLimit(t *testing.T) {
	// ARRANGE
	input := ""
	for i := 0; i < 5000; i++ {
		input += "1"
	}
	for i := 0; i < 5000; i++ {
		input += "2"
	}
	request := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(input))
	span := testSpan{make(map[string]string)}

	// ACT
	err1 := decorateWithBody(request, &span)
	forwardedBody, err2 := io.ReadAll(request.Body)

	// ASSERT
	require.NoError(t, err2)
	require.NoError(t, err1)
	assert.Len(t, span.attributes["http.request.body"], 5000)
	assert.NotContains(t, span.attributes["http.request.body"], "2")
	assert.Equal(t, input, string(forwardedBody))
}

type testSpan struct {
	attributes map[string]string
}

func (s *testSpan) End() {}

func (s *testSpan) AddStringAttribute(name, value string) {
	s.attributes[name] = value
}

func (s *testSpan) Empty() bool {
	return false
}

func (s *testSpan) Internal() any {
	return nil
}
