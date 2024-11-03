package zotero

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
)

type Item struct {
    Key  string `json:"key"`
    Data struct {
        Filename string `json:"filename"`
    } `json:"data"`
}

// DownloadPDFs downloads all PDFs from the specified Zotero group or collection
func DownloadPDFs(username, apiKey, collectionKey string) error {
    const baseURL = "https://api.zotero.org"
    userID := username

    // Construct the URL for the collection
    collectionURL := fmt.Sprintf("%s/users/%s/collections/%s/items?format=json&itemType=attachment", baseURL, userID, collectionKey)

    // Create a new HTTP request
    req, err := http.NewRequest("GET", collectionURL, nil)
    if err != nil {
        return fmt.Errorf("error creating request: %v", err)
    }

    // Add the API key to the request header
    req.Header.Add("Zotero-API-Key", apiKey)

    // Send the request
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return fmt.Errorf("error making request: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("error: received non-200 response status: %s", resp.Status)
    }

    // Parse the JSON response
    var items []Item
    if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
        return fmt.Errorf("error decoding JSON: %v", err)
    }

    // Create a directory to save the PDFs
    outputDir := "zotero_pdfs"
    if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
        return fmt.Errorf("error creating directory: %v", err)
    }

    // Download each PDF
    for _, item := range items {
        downloadURL := fmt.Sprintf("%s/users/%s/items/%s/file", baseURL, userID, item.Key)
        req, err := http.NewRequest("GET", downloadURL, nil)
        if err != nil {
            fmt.Printf("Error creating request for file: %v\n", err)
            continue
        }
        req.Header.Add("Zotero-API-Key", apiKey)

        resp, err := client.Do(req)
        if err != nil {
            fmt.Printf("Error downloading file: %v\n", err)
            continue
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
            fmt.Printf("Error: received non-200 response status for file: %s\n", resp.Status)
            continue
        }

        // Create the file
        outputPath := filepath.Join(outputDir, item.Data.Filename)
        outFile, err := os.Create(outputPath)
        if err != nil {
            fmt.Printf("Error creating file: %v\n", err)
            continue
        }
        defer outFile.Close()

        // Write the response body to the file
        _, err = io.Copy(outFile, resp.Body)
        if err != nil {
            fmt.Printf("Error saving file: %v\n", err)
            continue
        }

        fmt.Println("Downloaded:", item.Data.Filename)
    }

    return nil
}
