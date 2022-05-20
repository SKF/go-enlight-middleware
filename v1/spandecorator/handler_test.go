package spandecorator

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SKF/go-utility/v2/useridcontext"
	"github.com/SKF/go-utility/v2/uuid"
	"github.com/stretchr/testify/require"

	"github.com/SKF/go-enlight-middleware/v1/spandecorator/internal"
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
