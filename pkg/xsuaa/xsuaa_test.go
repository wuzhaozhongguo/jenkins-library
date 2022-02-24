package xsuaa

import (
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestXSUAA_GetBearerToken(t *testing.T) {
	type (
		args struct {
			oauthUrlPath string
			clientID     string
			clientSecret string
		}
		want struct {
			authToken AuthToken
			errRegex  string
		}
		response struct {
			statusCode int
			bodyText   string
		}
	)
	tests := []struct {
		name     string
		args     args
		want     want
		response response
	}{
		{
			name: "Straight forward",
			args: args{
				clientID:     "myClientID",
				clientSecret: "secret",
			},
			want: want{
				authToken: AuthToken{
					TokenType:   "bearer",
					AccessToken: "1234",
					ExpiresIn:   9876,
				}},
			response: response{
				bodyText: `{"access_token": "1234", "expires_in": 9876, "token_type": "bearer"}`,
			},
		},
		{
			name: "OAuth Url with path",
			args: args{
				oauthUrlPath: "/oauth/token?grant_type=client_credentials",
				clientID:     "myClientID",
				clientSecret: "secret",
			},
			want: want{
				authToken: AuthToken{
					TokenType:   "bearer",
					AccessToken: "1234",
					ExpiresIn:   9876,
				}},
			response: response{
				bodyText: `{"access_token": "1234", "expires_in": 9876, "token_type": "bearer"}`,
			},
		},
		{
			name: "No token type",
			args: args{
				clientID:     "myClientID",
				clientSecret: "secret",
			},
			want: want{
				authToken: AuthToken{
					TokenType:   "bearer",
					AccessToken: "1234",
					ExpiresIn:   9876,
				}},
			response: response{
				bodyText: `{"access_token": "1234", "expires_in": 9876}`,
			},
		},
		{
			name: "HTTP error",
			args: args{
				clientID:     "myClientID",
				clientSecret: "secret",
			},
			want: want{errRegex: `HTTP GET request failed: request to .*/oauth/token\?grant_type=client_credentials&response_type=token returned with response 401 Unauthorized`},
			response: response{
				statusCode: 401,
				bodyText:   `{"error": "unauthorized"}`,
			},
		},
		{
			name: "Wrong response code",
			want: want{errRegex: `expected response code 200, got '201', response body: '{"success": "created"}'`},
			response: response{
				statusCode: 201,
				bodyText:   `{"success": "created"}`,
			},
		},
		{
			name: "No 'access_token' field in json response",
			want: want{errRegex: `expected authToken field 'access_token' in json response; response body: '{"authToken": "1234"}'`},
			response: response{
				bodyText: `{"authToken": "1234"}`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			xsuaa := XSUAA{client: piperhttp.Client{}}
			var requestedUrlPath string
			// Start a local HTTP server
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				requestedUrlPath = req.URL.String()
				if tt.response.statusCode != 0 {
					rw.WriteHeader(tt.response.statusCode)
				}
				rw.Write([]byte(tt.response.bodyText))
			}))
			// Close the server when test finishes
			defer server.Close()

			oauthUrl := server.URL + tt.args.oauthUrlPath
			err := xsuaa.GetBearerToken(oauthUrl, tt.args.clientID, tt.args.clientSecret)
			if tt.want.errRegex != "" {
				require.Error(t, err, "Error expected")
				assert.Regexp(t, tt.want.errRegex, err.Error(), "")
				return
			}
			require.NoError(t, err, "No error expected")
			assert.Equal(t, tt.want.authToken, xsuaa.AuthToken, "Did not receive expected authToken.")
			wantUrlPath := "/oauth/token?grant_type=client_credentials&response_type=token"
			assert.Equal(t, wantUrlPath, requestedUrlPath)
		})
	}
}

func TestClient_SetBearerToken(t *testing.T) {
	type (
		args struct {
			clientID     string
			clientSecret string
		}
		want struct {
			token    string
			errRegex string
		}
		response struct {
			statusCode int
			bodyText   string
		}
	)
	tests := []struct {
		name     string
		args     args
		want     want
		response response
	}{
		{
			name: "Straight forward",
			args: args{
				clientID:     "myClientID",
				clientSecret: "secret",
			},
			want: want{token: "bearer 1234"},
			response: response{
				bodyText: `{"access_token": "1234", "expires_in": 9876, "token_type": "bearer"}`,
			},
		},
		{
			name: "Error case",
			args: args{
				clientID:     "myClientID",
				clientSecret: "secret",
			},
			want: want{errRegex: `HTTP GET request failed: request to .*/oauth/token\?grant_type=client_credentials&response_type=token returned with response 401 Unauthorized`},
			response: response{
				statusCode: 401,
				bodyText:   `{"error": "unauthorized"}`,
			},
		},
		{
			name: "Different token type",
			args: args{
				clientID:     "myClientID",
				clientSecret: "secret",
			},
			want: want{token: "jwt 1234"},
			response: response{
				bodyText: `{"access_token": "1234", "expires_in": 9876, "token_type": "jwt"}`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			xsuaa := XSUAA{client: piperhttp.Client{}}
			var headers http.Header
			// Start a local HTTP server
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				headers = req.Header
				if tt.response.statusCode != 0 {
					rw.WriteHeader(tt.response.statusCode)
				}
				rw.Write([]byte(tt.response.bodyText))
			}))
			// Close the server when test finishes
			defer server.Close()

			err := xsuaa.SetBearerToken(server.URL, tt.args.clientID, tt.args.clientSecret)
			if tt.want.errRegex != "" {
				require.Error(t, err, "Error expected")
				assert.Regexp(t, tt.want.errRegex, err.Error(), "")
				return
			}
			require.NoError(t, err, "No error expected")
			_, err = xsuaa.client.SendRequest(http.MethodGet, server.URL, nil, nil, nil)
			require.NoError(t, err, "Client should work without error")
			assert.Equal(t, tt.want.token, headers.Get("Authorization"))
		})
	}
}

func Test_readResponseBody(t *testing.T) {
	tests := []struct {
		name        string
		response    *http.Response
		want        []byte
		wantErrText string
	}{
		{
			name:     "Straight forward",
			response: httpmock.NewStringResponse(200, "test string"),
			want:     []byte("test string"),
		},
		{
			name:        "No response error",
			wantErrText: "did not retrieve an HTTP response",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readResponseBody(tt.response)
			if tt.wantErrText != "" {
				require.Error(t, err, "Error expected")
				assert.EqualError(t, err, tt.wantErrText, "Error is not equal")
				return
			}
			require.NoError(t, err, "No error expected")
			assert.Equal(t, tt.want, got, "Did not receive expected body")
		})
	}
}
