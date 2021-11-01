package clientid_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/SKF/go-rest-utility/problems"
	"github.com/SKF/go-utility/v2/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	client_id "github.com/SKF/go-enlight-middleware/client-id"
	"github.com/SKF/go-enlight-middleware/client-id/models"
	"github.com/SKF/go-enlight-middleware/client-id/store"
)

var (
	ClientA = client_id.ClientID{
		Identifier: uuid.New(),
	}
	ClientB = client_id.ClientID{
		Identifier:   uuid.New(),
		Environments: models.Environments{models.Sandbox},
	}
	ClientC = client_id.ClientID{
		Identifier: uuid.New(),
		NotBefore:  time.Now().Add(1 * time.Hour),
	}
	ClientD = client_id.ClientID{
		Identifier: uuid.New(),
		Expires:    time.Now().Add(-1 * time.Hour),
	}
)

type ClientIDEcho struct {
	Found    bool      `json:"found"`
	ClientID uuid.UUID `json:"cid"`
}

func doRequest(request *http.Request, mw *client_id.Middleware) *http.Response {
	endpoint := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var response ClientIDEcho

		cid, found := client_id.FromContext(r.Context())

		if found {
			response = ClientIDEcho{
				Found:    found,
				ClientID: cid.Identifier,
			}
		}

		json.NewEncoder(w).Encode(response) //nolint
	})

	w := httptest.NewRecorder()

	r := mux.NewRouter()
	r.Use(mw.Middleware())
	r.Handle("/", endpoint)

	r.ServeHTTP(w, request)

	return w.Result()
}

type ResponseTester interface {
	TestResponse(*testing.T, *http.Response)
}

type (
	Problem struct {
		Type   string
		Status int
	}
)

func (expected Problem) TestResponse(t *testing.T, response *http.Response) {
	require.Equal(t, expected.Status, response.StatusCode)

	var actual problems.BasicProblem
	err := json.NewDecoder(response.Body).Decode(&actual)
	require.NoError(t, err)

	require.Equal(t, "application/problem+json", response.Header.Get("Content-Type"))
	require.Equal(t, expected.Type, actual.ProblemType())
}

func (expected ClientIDEcho) TestResponse(t *testing.T, response *http.Response) {
	var actual ClientIDEcho
	err := json.NewDecoder(response.Body).Decode(&actual)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, response.StatusCode)
	require.Equal(t, expected.Found, actual.Found)

	if expected.Found {
		require.Equal(t, expected.ClientID, actual.ClientID)
	}
}

func Test(t *testing.T) {
	cids := store.NewLocal().
		Add(ClientA).
		Add(ClientB).
		Add(ClientC).Add(ClientD)

	testCases := []struct {
		desc       string
		requestCID uuid.UUID
		required   bool
		shouldBe   ResponseTester
	}{
		{
			desc:       "Valid non-required cid should be found",
			requestCID: ClientA.Identifier,
			shouldBe:   ClientIDEcho{Found: true, ClientID: ClientA.Identifier},
		},
		{
			desc:     "Empty non-required cid should not be found, but allowed",
			shouldBe: ClientIDEcho{Found: false},
		},
		{
			desc:       "Unknown cid should be allowed if not required",
			requestCID: uuid.New(),
			shouldBe:   ClientIDEcho{Found: false},
		},
		{
			desc:       "Cid not active in production environment should not be found",
			requestCID: ClientB.Identifier,
			shouldBe:   ClientIDEcho{Found: false},
		},
		{
			desc:       "Cid not yet activated should not be found",
			requestCID: ClientC.Identifier,
			shouldBe:   ClientIDEcho{Found: false},
		},
		{
			desc:       "Cid that has expired should not be found",
			requestCID: ClientD.Identifier,
			shouldBe:   ClientIDEcho{Found: false},
		},
		{
			desc:       "Valid required cid should be found",
			requestCID: ClientA.Identifier,
			required:   true,
			shouldBe:   ClientIDEcho{Found: true, ClientID: ClientA.Identifier},
		},
		{
			desc:     "Empty cid should not be allowed when required",
			required: true,
			shouldBe: Problem{Type: "/problems/missing-client-id", Status: http.StatusUnauthorized},
		},
		{
			desc:       "Unknown cid should not be allowed when required",
			requestCID: uuid.New(),
			required:   true,
			shouldBe:   Problem{Type: "/problems/unknown-client-id", Status: http.StatusUnauthorized},
		},
		{
			desc:       "Cid not active in production environment should not be allowed when required",
			requestCID: ClientB.Identifier,
			required:   true,
			shouldBe:   Problem{Type: "/problems/unauthorized-client-id", Status: http.StatusForbidden},
		},
		{
			desc:       "Cid not yet activated should not be allowed when required",
			requestCID: ClientC.Identifier,
			required:   true,
			shouldBe:   Problem{Type: "/problems/not-yet-active-client-id", Status: http.StatusForbidden},
		},
		{
			desc:       "Cid that has expired should not be allowed when required",
			requestCID: ClientD.Identifier,
			required:   true,
			shouldBe:   Problem{Type: "/problems/expired-client-id", Status: http.StatusForbidden},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			options := []client_id.Option{
				client_id.WithStore(cids),
				client_id.WithHeaderExtractor("X-Client-ID"),
			}

			if tC.required {
				options = append(options, client_id.WithRequired())
			}

			middleware := client_id.New(options...)

			request := httptest.NewRequest(http.MethodGet, "/", nil)
			if tC.requestCID != "" {
				request.Header.Set("X-Client-ID", tC.requestCID.String())
			}

			response := doRequest(request, middleware)
			defer response.Body.Close()

			tC.shouldBe.TestResponse(t, response)
		})
	}
}

func Test_Options(t *testing.T) {
	options := []client_id.Option{
		client_id.WithStore(store.NewLocal()),
		client_id.WithHeaderExtractor("X-Client-ID"),
		client_id.WithRequired(),
	}

	middleware := client_id.New(options...)

	request := httptest.NewRequest(http.MethodOptions, "/", nil)

	response := doRequest(request, middleware)
	defer response.Body.Close()

	require.Equal(t, http.StatusOK, response.StatusCode)
}
