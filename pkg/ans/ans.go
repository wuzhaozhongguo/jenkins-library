package ans

import (
	"bytes"
	"encoding/json"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/xsuaa"
	"github.com/pkg/errors"
	"net/http"
)

type ANSServiceKey struct {
	Url          string `json:"url"`
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	OauthUrl     string `json:"oauth_url"`
}

// ReadANSServiceKey unmarshalls the given json service key string.
func ReadANSServiceKey(serviceKeyJSON string) (ansServiceKey ANSServiceKey, err error) {
	// parse
	err = json.Unmarshal([]byte(serviceKeyJSON), &ansServiceKey)
	if err != nil {
		err = errors.Wrap(err, "error unmarshalling ANS serviceKey")
		return
	}

	log.Entry().Info("ANS serviceKey read successfully")
	return
}

const body = `{
"eventTimestamp": 1535618178,
"resource": {
"resourceName": "web-shop",
"resourceType": "app",
"tags": {
"env": "prod"
}
},
"severity": "INFO",
"category": "ALERT",
"subject": "Overloaded external dependency of My Web Shop external dependency",
"body": "External dependency showing recommendations does not respond on time. Stop some clients to reduce the load.",
"tags": {
"ans:correlationId": "30118",
"ans:status": "CREATE_OR_UPDATE",
"customTag": "42"
}
}`

func Send(serviceKeyString string) error {
	serviceKey, err:= ReadANSServiceKey(serviceKeyString)
	if err != nil {
		return err
	}
	xsuaa := &xsuaa.XSUAA{Client: piperhttp.Client{}}
	err = xsuaa.SetBearerToken(serviceKey.OauthUrl,
		serviceKey.ClientId,
		serviceKey.ClientSecret)
	if err != nil {
		return err
	}
	header := make(http.Header)
	header.Add("Content-Type", "application/json")
	response, err := xsuaa.Client.SendRequest(http.MethodPost,
		"https://clm-sl-ans-live-ans-service-api.cfapps.eu10.hana.ondemand.com/cf/producer/v1/resource-events",
		bytes.NewBuffer([]byte(body)), header, nil)
	if err != nil {
		return err
	}

	log.Entry().Infof("XXXXXXXXX Status code: %d", response.StatusCode)

	return nil
}
