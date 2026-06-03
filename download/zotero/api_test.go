package zotero

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
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
		name                        string
		collectionName              string
		mockCollectionResponse      string // Mock response for user's collections request
		mockItemsResponse           string // Mock response for items request
		mockGroupResponse           string // Mock response for groups request
		mockGroupCollectionResponse string // Mock response for group's collections request
		mockError                   error
		expectedError               string
	}{
		{
			name:           "successful PDF download from user collection",
			collectionName: "collection",
			mockCollectionResponse: `[
                {"key":"123", "data":{"key":"123", "name":"collection", "parentCollection":false}}
            ]`,
			mockItemsResponse: `[
                {"key":"abc", "data":{"filename":"file.pdf"}}
            ]`,
		},
		{
			name:                   "successful PDF download from group collection",
			collectionName:         "TestGroup/collection",
			mockCollectionResponse: `[]`, // No collections in user's library
			mockGroupResponse: `[
                {"data":{"id":1, "name":"TestGroup"}}
            ]`,
			mockGroupCollectionResponse: `[
                {"key":"456", "data":{"key":"456", "name":"collection", "parentCollection":false}}
            ]`,
			mockItemsResponse: `[
                {"key":"def", "data":{"filename":"group_file.pdf"}}
            ]`,
		},
		{
			name:                   "API returns error",
			collectionName:         "collection",
			mockCollectionResponse: "",
			mockGroupResponse:      "",
			mockItemsResponse:      "",
			mockError:              errors.New("network error"),
			expectedError:          "network error",
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
					urlPath := req.URL.Path

					// Handle user's collections request
					if strings.Contains(urlPath, "/users/") && strings.Contains(urlPath, "/collections") && !strings.Contains(urlPath, "/items") {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewBufferString(tc.mockCollectionResponse)),
							Header:     make(http.Header),
						}, nil
					}
					// Handle groups list request
					if strings.Contains(urlPath, "/users/") && strings.HasSuffix(urlPath, "/groups") {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewBufferString(tc.mockGroupResponse)),
							Header:     make(http.Header),
						}, nil
					}
					// Handle group's collections request
					if strings.Contains(urlPath, "/groups/") && strings.Contains(urlPath, "/collections") && !strings.Contains(urlPath, "/items") {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewBufferString(tc.mockGroupCollectionResponse)),
							Header:     make(http.Header),
						}, nil
					}
					// Handle items request
					if strings.Contains(urlPath, "/items") && !strings.Contains(urlPath, "/file") {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewBufferString(tc.mockItemsResponse)),
							Header:     make(http.Header),
						}, nil
					}
					// Handle file download request
					if strings.Contains(urlPath, "/items/") && strings.Contains(urlPath, "/file") {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewBufferString("PDF content")),
							Header:     make(http.Header),
						}, nil
					}
					// Default case
					return &http.Response{
						StatusCode: http.StatusNotFound,
						Body:       io.NopCloser(bytes.NewBufferString(``)),
						Header:     make(http.Header),
					}, nil
				},
			}

			// Use t.TempDir() to create a temporary directory
			tempDir := t.TempDir()

			err := DownloadWithConfig(client, Config{
				Zotero: ZoteroConfig{
					User:      "user",
					APIKey:    "api_key",
					Group:     tc.collectionName,
					OutputDir: tempDir,
				},
			})
			if tc.expectedError != "" {
				if err == nil || !strings.Contains(err.Error(), tc.expectedError) {
					t.Errorf("expected error %v, got %v", tc.expectedError, err)
				}
			} else if err != nil {
				t.Errorf("expected no error, got %v", err)
			} else {
				expectedFile := filepath.Join(tempDir, "file.pdf")
				if strings.Contains(tc.collectionName, "TestGroup") {
					expectedFile = filepath.Join(tempDir, "group_file.pdf")
				}
				if _, err := os.Stat(expectedFile); err != nil {
					t.Errorf("expected downloaded file in output_dir, got %v", err)
				}
			}
		})
	}
}

func TestDownloadParsesTOMLConfig(t *testing.T) {
	tempDir := t.TempDir()
	client := &MockClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			urlPath := req.URL.Path
			if strings.Contains(urlPath, "/users/") && strings.Contains(urlPath, "/collections") && !strings.Contains(urlPath, "/items") {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(bytes.NewBufferString(`[
						{"key":"123", "data":{"key":"123", "name":"collection", "parentCollection":false}}
					]`)),
					Header: make(http.Header),
				}, nil
			}
			if strings.Contains(urlPath, "/items") && !strings.Contains(urlPath, "/file") {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(`[{"key":"abc", "data":{"filename":"file.pdf"}}]`)),
					Header:     make(http.Header),
				}, nil
			}
			if strings.Contains(urlPath, "/items/") && strings.Contains(urlPath, "/file") {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString("PDF content")),
					Header:     make(http.Header),
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(bytes.NewBufferString("")),
				Header:     make(http.Header),
			}, nil
		},
	}

	config := `
[zotero]
user = "user"
api_key = "api_key"
group = "collection"
output_dir = "` + tempDir + `"
`
	if err := Download(client, config); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(tempDir, "file.pdf")); err != nil {
		t.Fatalf("expected file.pdf in configured output_dir, got %v", err)
	}
}

func TestDownloadRequiresOutputDir(t *testing.T) {
	err := Download(&MockClient{}, `
[zotero]
user = "user"
api_key = "api_key"
group = "collection"
`)
	if err == nil || !strings.Contains(err.Error(), "zotero.output_dir is required") {
		t.Fatalf("expected output_dir validation error, got %v", err)
	}
}
