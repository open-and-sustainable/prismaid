package zotero

import (
    "bytes"
    "errors"
    "io"
    "net/http"
    "strings"
    "testing"
)

// MockClient is the mock of the HTTP client
type MockClient struct {
    DoFunc func(req *http.Request) (*http.Response, error)
}

// Do is the mock client's method that intercepts real HTTP requests
func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
    return m.DoFunc(req)
}

// TestDownloadPDFs tests the DownloadPDFs function
func TestDownloadPDFs(t *testing.T) {
    tests := []struct {
        name          string
        mockResponse  string
        mockError     error
        expectedError string
    }{
        {
            name:         "successful PDF download",
            mockResponse: `[{"key":"123","data":{"filename":"test.pdf"}}]`,
            mockError:    nil,
        },
        {
            name:          "API returns error",
            mockResponse: "",
            mockError:     errors.New("network error"),
            expectedError: "network error",
        },
        {
            name:          "API returns non-200 status",
            mockResponse: "error response",
            mockError:     nil,
            expectedError: "received non-200 response status",
        },
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            client := &MockClient{
                DoFunc: func(req *http.Request) (*http.Response, error) {
                    if tc.mockError != nil {
                        return nil, tc.mockError
                    }
                    return &http.Response{
                        StatusCode: http.StatusOK,
                        Body:       io.NopCloser(bytes.NewBufferString(tc.mockResponse)),
                    }, nil
                },
            }

            // Using a test directory that does not actually exist to ensure files are not written
            err := DownloadPDFs("user", "api_key", "collection", "/non/existent/directory")
            if (err != nil && tc.expectedError == "") || (err == nil && tc.expectedError != "") {
                t.Errorf("DownloadPDFs() error = %v, expectedError %v", err, tc.expectedError)
            }
            if tc.expectedError != "" && err != nil && !strings.Contains(err.Error(), tc.expectedError) {
                t.Errorf("DownloadPDFs() error = %v, expected to contain %v", err.Error(), tc.expectedError)
            }
        })
    }
}
