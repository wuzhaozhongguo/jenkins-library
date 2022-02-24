package xsuaa

import (
	"encoding/json"
	"fmt"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/url"
)

type XSUAA struct {
	client    piperhttp.Client
	AuthToken AuthToken
}

// Sender provides an interface to the piper http client for uid/pwd and token authenticated requests
// It includes a SetBearerToken function that will retrieve a token from the XSUAA service and store it in the client
type Sender interface {
	piperhttp.Sender
	SetBearerToken(oauthBaseUrl, clientID, clientSecret string) error
}

// AuthToken provides a structure for the XSUAA auth token to be marshalled into
type AuthToken struct {
	TokenType   string `json:"token_type"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

func (x *XSUAA) SetBearerToken(oauthBaseUrl, clientID, clientSecret string) error {
	err := x.GetBearerToken(oauthBaseUrl, clientID, clientSecret)
	if err != nil {
		return err
	}

	x.client.SetOptions(piperhttp.ClientOptions{Token: fmt.Sprintf("%s %s", x.AuthToken.TokenType, x.AuthToken.AccessToken)})
	return nil
}

// GetBearerToken authenticates to and retrieves the auth information from the provided XSUAA oAuth base url. The following path
// and query is always used: /oauth/token?grant_type=client_credentials&response_type=token. The gotten JSON string is marshalled
// into an AuthToken struct and returned. If no 'access_token' field was present in the JSON response, an error is returned.
func (x *XSUAA) GetBearerToken(oauthUrl, clientID, clientSecret string) (err error) {
	const method = http.MethodGet
	const urlPathAndQuery = "oauth/token?grant_type=client_credentials&response_type=token"

	oauthBaseUrl, err := url.Parse(oauthUrl)
	if err != nil {
		return
	}
	entireUrl := fmt.Sprintf("%s://%s/%s", oauthBaseUrl.Scheme, oauthBaseUrl.Host, urlPathAndQuery)

	clientOptions := piperhttp.ClientOptions{
		Username: clientID,
		Password: clientSecret,
	}
	x.client.SetOptions(clientOptions)

	header := make(http.Header)
	header.Add("Accept", "application/json")

	response, httpErr := x.client.SendRequest(method, entireUrl, nil, header, nil)
	if httpErr != nil {
		log.SetErrorCategory(log.ErrorService)
		err = errors.Wrapf(httpErr, "HTTP %s request failed", method)
		return
	}

	bodyText, err := readResponseBody(response)
	if err != nil {
		return
	}

	if response.StatusCode != http.StatusOK {
		err = errors.Errorf("expected response code 200, got '%d', response body: '%s'", response.StatusCode, bodyText)
		return
	}

	parsingErr := json.Unmarshal(bodyText, &x.AuthToken)
	if err != nil {
		err = errors.Wrapf(parsingErr, "HTTP response body could not be parsed as JSON: %s", bodyText)
		return
	}

	if x.AuthToken.AccessToken == "" {
		err = errors.Errorf("expected authToken field 'access_token' in json response; response body: '%s'", bodyText)
		return
	}

	if x.AuthToken.TokenType == "" {
		x.AuthToken.TokenType = "bearer"
	}

	return
}

func readResponseBody(response *http.Response) ([]byte, error) {
	if response == nil {
		return nil, errors.Errorf("did not retrieve an HTTP response")
	}
	if response.Body != nil {
		defer response.Body.Close()
	}
	bodyText, readErr := ioutil.ReadAll(response.Body)
	if readErr != nil {
		return nil, errors.Wrap(readErr, "HTTP response body could not be read")
	}
	return bodyText, nil
}
