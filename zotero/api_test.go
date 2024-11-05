package zotero

import (
    "bytes"
    "errors"
    "io"
    "net/http"
    "strings"
    "testing"
)

type MockClient struct {
    DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
    return m.DoFunc(req)
}

func TestDownloadPDFs(t *testing.T) {
    tests := []struct {
        name                    string
        mockCollectionResponse  string  // Corrected mock response for collections request
        mockItemsResponse       string  // Mock response for items request
        mockError               error
        expectedError           string
    }{
        {
            name: "successful PDF download",
            mockCollectionResponse: `{
                "data": [
                    {"key": "123", "name": "collection"}
                ]
            }`,
            mockItemsResponse: `[
        		{"key":"abc", "data":{"filename":"file.pdf"}}
    		]`,
        },
        {
            name: "API returns error",
            mockCollectionResponse: "",
            mockItemsResponse: "",
            mockError: errors.New("network error"),
            expectedError: "network error",
        },
        // Include other test scenarios as needed
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            client := &MockClient{
                DoFunc: func(req *http.Request) (*http.Response, error) {
                    if tc.mockError != nil {
                        return nil, tc.mockError
                    }
                    if strings.Contains(req.URL.Path, "collections") {
                        return &http.Response{
                            StatusCode: http.StatusOK,
                            Body:       io.NopCloser(bytes.NewBufferString(tc.mockCollectionResponse)),
                        }, nil
                    } else if strings.Contains(req.URL.Path, "items") {
                        return &http.Response{
                            StatusCode: http.StatusOK,
                            Body:       io.NopCloser(bytes.NewBufferString(tc.mockItemsResponse)),
                        }, nil
                    }
                    return nil, nil  // Default to no error if not specified
                },
            }

            err := DownloadPDFs(client, "user", "api_key", "collection", "/non/existent/directory")
            if tc.expectedError != "" {
                if err == nil || !strings.Contains(err.Error(), tc.expectedError) {
                    t.Errorf("expected error %v, got %v", tc.expectedError, err)
                }
            } else if err != nil {
                t.Errorf("expected no error, got %v", err)
            }
        })
    }
}
