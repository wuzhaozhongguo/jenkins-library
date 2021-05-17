package protecode

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"

	piperHttp "github.com/SAP/jenkins-library/pkg/http"
)

func TestProtecode_deleteResult(t *testing.T) {
	testURL := "https://example.org"

	t.Run("success", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		uploader := &piperHttp.Client{}
		uploader.SetOptions(piperHttp.ClientOptions{UseDefaultTransport: true})
		pc := &Protecode{
			serverURL: testURL,
			client:    uploader,
		}
		// add response handler
		httpmock.RegisterResponder(http.MethodDelete, fmt.Sprintf(endpointProduct, 12345),
			// httpmock.RegisterResponder(http.MethodDelete, testURL+fmt.Sprintf(endpointProduct, 12345),
			func(req *http.Request) (*http.Response, error) {
				assert.Empty(t, req.Header)
				assert.Equal(t, http.MethodDelete, req.Method)
				// id := httpmock.MustGetSubmatchAsUint(req, 1) // 1=first regexp submatch
				return httpmock.NewStringResponse(200, `{"meta": {"code": 200}}`), nil
			})
		// test
		err := pc.deleteResult(12345)
		// assert
		assert.NoError(t, err)
		assert.Equal(t, 1, httpmock.GetTotalCallCount(), "unexpected number of requests")
	})
	t.Run("failure", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		uploader := &piperHttp.Client{}
		uploader.SetOptions(piperHttp.ClientOptions{UseDefaultTransport: true})
		pc := &Protecode{
			serverURL: testURL,
			client:    uploader,
		}
		// add response handler
		httpmock.RegisterResponder(http.MethodDelete, testURL+fmt.Sprintf(endpointProduct, -1),
			httpmock.NewStringResponder(http.StatusBadRequest, `{"meta": {"code": 200}}`))
		// test
		err := pc.deleteResult(-1)
		// assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete result")
		// assert.Equal(t, 111, count)
		assert.Equal(t, 1, httpmock.GetTotalCallCount(), "unexpected number of requests")
	})
}
