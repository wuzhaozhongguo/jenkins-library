package protecode

import (
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

const (
	statusBusy   = "B"
	statusReady  = "R"
	statusFailed = "F"

	endpointApps      = "/api/apps/%s/"
	endpointProduct   = "/api/product/%v/"
	endpointPdfReport = "/api/product/%v/pdf-report"
	endpointUpload    = "/api/upload/%v"
	endpointFetch     = "/api/fetch/"
)

func (pc *Protecode) loadProduct(group string) (*io.ReadCloser, error) {
	protecodeURL := pc.createURL(fmt.Sprintf(endpointApps, group), "", "")
	headers := map[string][]string{
		"acceptType": {"application/json"},
	}
	// send request
	r, err := pc.send(http.MethodGet, protecodeURL, headers)
	if err != nil {
		return r, errors.Wrapf(err, "failed to load product: %s", protecodeURL)
	}
	return r, nil
}

// UploadScanFile upload the scan file to the protecode server
func (pc *Protecode) triggerWithFileUpload(group, filePath, fileName string, deleteBinary bool) (*io.ReadCloser, error) {
	protecodeURL := pc.createURL(fmt.Sprintf(endpointUpload, fileName), "", "")
	headers := map[string][]string{
		"Group":         {group},
		"Delete-Binary": {fmt.Sprintf("%v", deleteBinary)},
	}
	// send request
	r, err := pc.upload(http.MethodPut, protecodeURL, filePath, headers)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to trigger scan with file upload", protecodeURL)
	}
	return r, nil
}

// DeclareFetchURL configures the fetch url for the protecode scan
func (pc *Protecode) triggerWithFetchUrl(group, fetchURL string, deleteBinary bool) (*io.ReadCloser, error) {
	protecodeURL := pc.createURL(endpointFetch, "", "")
	headers := map[string][]string{
		"Content-Type":  {"application/json"},
		"Group":         {group},
		"Delete-Binary": {fmt.Sprintf("%v", deleteBinary)},
		"Url":           {fetchURL},
	}
	// send request
	r, err := pc.send(http.MethodPost, protecodeURL, headers)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to trigger scan with fetch-url", protecodeURL)
	}
	return r, nil
}

func (pc *Protecode) loadResult(productID int) (*io.ReadCloser, error) {
	protecodeURL := pc.createURL(fmt.Sprintf(endpointProduct, productID), "", "")
	headers := map[string][]string{
		"acceptType": {"application/json"},
	}
	// send request
	r, err := pc.send(http.MethodGet, protecodeURL, headers)
	if err != nil {
		return r, errors.Wrapf(err, "failed to load results", protecodeURL)
	}
	return r, nil
}

func (pc *Protecode) loadResultAsPdf(productID int, reportFileName string) (*io.ReadCloser, error) {
	protecodeURL := pc.createURL(fmt.Sprintf(endpointPdfReport, productID), "", "")
	headers := map[string][]string{
		"Cache-Control": {"no-cache, no-store, must-revalidate"},
		"Pragma":        {"no-cache"},
		"Outputfile":    {reportFileName},
	}
	// send request
	readCloser, err := pc.send(http.MethodGet, protecodeURL, headers)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load result as PDF: %s", protecodeURL)
	}
	return readCloser, nil
}

func (pc *Protecode) deleteResult(productID int) error {
	protecodeURL := pc.createURL(fmt.Sprintf(endpointProduct, productID), "", "")
	headers := map[string][]string{}
	// send request
	_, err := pc.send(http.MethodDelete, protecodeURL, headers)
	if err != nil {
		return errors.Wrapf(err, "failed to delete result: %s", protecodeURL)
	}
	return nil
}

func (pc *Protecode) send(method string, url string, headers map[string][]string) (*io.ReadCloser, error) {
	r, err := pc.client.SendRequest(method, url, nil, headers, nil)
	if err != nil {
		return nil, err
	}
	return &r.Body, nil
}

func (pc *Protecode) upload(method string, url string, filename string, headers map[string][]string) (*io.ReadCloser, error) {
	r, err := pc.client.UploadRequest(method, url, filename, "file", headers, nil)
	if err != nil {
		return nil, err
	}
	return &r.Body, nil
}
