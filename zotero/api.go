package zotero

import (
    "encoding/json"
    "fmt"
    "log"
    "io"
    "net/http"
    "os"
    "path/filepath"
)


type HttpClient interface {
    Do(req *http.Request) (*http.Response, error)
}

type Item struct {
    Key  string `json:"key"`
    Data struct {
        Filename string `json:"filename"`
    } `json:"data"`
}

// DownloadPDFs downloads all PDFs from the specified Zotero group or collection
func DownloadPDFs(client HttpClient, username, apiKey, collectionName, parentDir string) error {
    const baseURL = "https://api.zotero.org"
    userID := username

    collectionKey, err := getCollectionKey(client, username, apiKey, collectionName)
    if err != nil {
        return err
    }

    // Construct the URL for the collection
    collectionURL := fmt.Sprintf("%s/users/%s/collections/%s/items?format=json&itemType=attachment", baseURL, userID, collectionKey)
    req, err := http.NewRequest("GET", collectionURL, nil)
    if err != nil {
        return fmt.Errorf("error creating request: %v", err)
    }

    req.Header.Add("Zotero-API-Key", apiKey)
    resp, err := client.Do(req)
    if err != nil {
        return fmt.Errorf("error making request: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("error: received non-200 response status: %s", resp.Status)
    }

    var items []Item
    if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
        return fmt.Errorf("error decoding JSON: %v", err)
    }

    outputDir := parentDir + "/zotero"
    if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
        return fmt.Errorf("error creating directory: %v", err)
    }

    for _, item := range items {
        downloadURL := fmt.Sprintf("%s/users/%s/items/%s/file", baseURL, userID, item.Key)
        req, err := http.NewRequest("GET", downloadURL, nil)
        if err != nil {
            log.Printf("Error creating request for file: %v\n", err)
            continue
        }
        req.Header.Add("Zotero-API-Key", apiKey)

        resp, err := client.Do(req)
        if err != nil {
            log.Printf("Error downloading file: %v\n", err)
            continue
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
            log.Printf("Error: received non-200 response status for file: %s\n", resp.Status)
            continue
        }

        outputPath := filepath.Join(outputDir, item.Data.Filename)
        outFile, err := os.Create(outputPath)
        if err != nil {
            log.Printf("Error creating file: %v\n", err)
            continue
        }
        defer outFile.Close()

        _, err = io.Copy(outFile, resp.Body)
        if err != nil {
            log.Printf("Error saving file: %v\n", err)
            continue
        }

        log.Println("Downloaded:", item.Data.Filename)
    }

    return nil
}

type Collection struct {
    Key    string `json:"key"`
    Name   string `json:"name"`
}

type CollectionsResponse struct {
    Data []Collection `json:"data"`
}


// getCollectionKey fetches the key of a collection by its name
func getCollectionKey(client HttpClient, username, apiKey, collectionName string) (string, error) {
    const baseURL = "https://api.zotero.org"
    collectionsURL := fmt.Sprintf("%s/users/%s/collections?format=json", baseURL, username)

    req, err := http.NewRequest("GET", collectionsURL, nil)
    if err != nil {
        return "", fmt.Errorf("error creating request: %v", err)
    }

    req.Header.Add("Zotero-API-Key", apiKey)
    resp, err := client.Do(req)
    if err != nil {
        return "", fmt.Errorf("error making request: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("error: received non-200 response status: %s", resp.Status)
    }

    var response CollectionsResponse
    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        return "", fmt.Errorf("error decoding JSON: %v", err)
    }

    for _, collection := range response.Data {
        if collection.Name == collectionName {
            return collection.Key, nil
        }
    }

    return "", fmt.Errorf("collection with name '%s' not found", collectionName)
}
