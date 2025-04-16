package zotero

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "strings"

	"github.com/open-and-sustainable/alembica/utils/logger"
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
        return downloadPDFsFromGroup(client, username, apiKey, collectionName, parentDir)
    } else {
        logger.Info("Collection key:", collectionKey)
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
            logger.Error("Error creating request for file: %v\n", err)
            continue
        }
        req.Header.Add("Zotero-API-Key", apiKey)

        resp, err := client.Do(req)
        if err != nil {
            logger.Error("Error downloading file: %v\n", err)
            continue
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
            logger.Error("Error: received non-200 response status for file: %s\n", resp.Status)
            continue
        }

        outputPath := filepath.Join(outputDir, item.Data.Filename)
        outFile, err := os.Create(outputPath)
        if err != nil {
            logger.Error("Error creating file: %v\n", err)
            continue
        }
        defer outFile.Close()

        _, err = io.Copy(outFile, resp.Body)
        if err != nil {
            logger.Error("Error saving file: %v\n", err)
            continue
        }

        logger.Info("Downloaded:", item.Data.Filename)
    }

    return nil
}

type Collection struct {
    Key  string `json:"key"`
    Data struct {
        Key              string      `json:"key"`
        Name             string      `json:"name"`
        ParentCollection interface{} `json:"parentCollection"`
    } `json:"data"`
}

type CollectionsResponse []Collection  // Since the root is an array


// getCollectionKey fetches the key of a collection by its name and nested structure
func getCollectionKey(client HttpClient, username, apiKey, collectionPath string) (string, error) {
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
        return "", fmt.Errorf("received non-200 response status: %s", resp.Status)
    }

    var collections []Collection
    if err := json.NewDecoder(resp.Body).Decode(&collections); err != nil {
        return "", fmt.Errorf("failed to decode response: %v", err)
    }

    pathParts := strings.Split(collectionPath, "/")
    var parentKey string // Start with empty parent key for top-level collections

    for _, part := range pathParts {
        matches, err := findCollectionByParent(collections, parentKey, part)
        if err != nil {
            return "", err
        }

        if len(matches) == 0 {
            return "", fmt.Errorf("collection with name '%s' not found under parent '%s'", part, parentKey)
        } else if len(matches) > 1 {
            return "", fmt.Errorf("multiple collections with name '%s' found under parent '%s'", part, parentKey)
        }

        // Since there's only one match, update parentKey to the current collection's key
        parentKey = matches[0].Data.Key
    }

    return parentKey, nil
}


func getParentCollectionKey(pc interface{}) (string, error) {
    switch v := pc.(type) {
    case string:
        return v, nil
    case bool:
        if !v { // if false, it's a root collection
            return "", nil
        }
        return "", fmt.Errorf("unexpected boolean value for parentCollection: true")
    case nil:
        return "", nil
    default:
        return "", fmt.Errorf("unknown type for parentCollection")
    }
}

func findCollectionByParent(collections []Collection, parentKey string, name string) ([]Collection, error) {
    var result []Collection
    for _, collection := range collections {
        collectionParentKey, err := getParentCollectionKey(collection.Data.ParentCollection)
        if err != nil {
            return nil, err
        }
        if collection.Data.Name == name && collectionParentKey == parentKey {
            result = append(result, collection)
        }
    }
    return result, nil
}

func downloadPDFsFromGroup(client HttpClient, username, apiKey, collectionName, parentDir string) error {
    const baseURL = "https://api.zotero.org"
    userID := username

    // Split collectionName into parts
    pathParts := strings.Split(collectionName, "/")
    if len(pathParts) == 0 {
        return fmt.Errorf("collectionName is empty")
    }

    groupName := pathParts[0]
    collectionPath := strings.Join(pathParts[1:], "/") // This could be empty if there's no collection path

    // Get the list of groups the user is a member of
    groupsURL := fmt.Sprintf("%s/users/%s/groups?format=json", baseURL, userID)
    req, err := http.NewRequest("GET", groupsURL, nil)
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

    var groups []Group
    if err := json.NewDecoder(resp.Body).Decode(&groups); err != nil {
        return fmt.Errorf("error decoding JSON: %v", err)
    }

    // Find the group with the matching name
    var groupID string
    groupFound := false
    for _, group := range groups {
        logger.Info("Fetched group: '%s' with ID: %d\n", group.Data.Name, group.Data.ID)
        if group.Data.Name == groupName {
            groupID = fmt.Sprintf("%d", group.Data.ID)
            groupFound = true
            break
        }
    }

    if !groupFound {
        return fmt.Errorf("group '%s' not found", groupName)
    }

    // If collectionPath is empty, download items from the group's library root
    var collectionKey string
    if collectionPath != "" {
        // Find the collection within the group
        collectionKey, err = getGroupCollectionKey(client, groupID, apiKey, collectionPath)
        if err != nil {
            return err
        } else {
            logger.Info("Collection key found in group '%s': %s", groupName, collectionKey)
        }
    }

    // Now download the PDFs
    var itemsURL string
    if collectionKey != "" {
        // Download items from the specific collection
        itemsURL = fmt.Sprintf("%s/groups/%s/collections/%s/items?format=json&itemType=attachment", baseURL, groupID, collectionKey)
    } else {
        // Download items from the group's root library
        itemsURL = fmt.Sprintf("%s/groups/%s/items?format=json&itemType=attachment", baseURL, groupID)
    }

    req, err = http.NewRequest("GET", itemsURL, nil)
    if err != nil {
        return fmt.Errorf("error creating request: %v", err)
    }
    req.Header.Add("Zotero-API-Key", apiKey)

    resp, err = client.Do(req)
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

    outputDir := filepath.Join(parentDir, "zotero")
    if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
        return fmt.Errorf("error creating directory: %v", err)
    }

    for _, item := range items {
        downloadURL := fmt.Sprintf("%s/groups/%s/items/%s/file", baseURL, groupID, item.Key)
        req, err := http.NewRequest("GET", downloadURL, nil)
        if err != nil {
            logger.Error("Error creating request for file: %v\n", err)
            continue
        }
        req.Header.Add("Zotero-API-Key", apiKey)

        resp, err := client.Do(req)
        if err != nil {
            logger.Error("Error downloading file: %v\n", err)
            continue
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
            logger.Error("Error: received non-200 response status for file: %s\n", resp.Status)
            continue
        }

        outputPath := filepath.Join(outputDir, item.Data.Filename)
        outFile, err := os.Create(outputPath)
        if err != nil {
            logger.Error("Error creating file: %v\n", err)
            continue
        }
        defer outFile.Close()

        _, err = io.Copy(outFile, resp.Body)
        if err != nil {
            logger.Error("Error saving file: %v\n", err)
            continue
        }

        logger.Info("Downloaded:", item.Data.Filename)
    }

    return nil // Successfully downloaded from group
}

func getGroupCollectionKey(client HttpClient, groupID, apiKey, collectionPath string) (string, error) {
    const baseURL = "https://api.zotero.org"

    collectionsURL := fmt.Sprintf("%s/groups/%s/collections?format=json", baseURL, groupID)

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
        return "", fmt.Errorf("received non-200 response status: %s", resp.Status)
    }

    var collections []Collection
    if err := json.NewDecoder(resp.Body).Decode(&collections); err != nil {
        return "", fmt.Errorf("failed to decode response: %v", err)
    }

    // Find the collection by path
    pathParts := strings.Split(collectionPath, "/")
    var parentKey string // Empty for top-level collections

    for _, part := range pathParts {
        matches, err := findGroupCollectionByParent(collections, parentKey, part)
        if err != nil {
            return "", err
        }

        if len(matches) == 0 {
            return "", fmt.Errorf("collection with name '%s' not found under parent '%s'", part, parentKey)
        } else if len(matches) > 1 {
            return "", fmt.Errorf("multiple collections with name '%s' found under parent '%s'", part, parentKey)
        }

        // Update parentKey for next iteration
        parentKey = matches[0].Data.Key
    }

    return parentKey, nil // Return the final collection key
}

func findGroupCollectionByParent(collections []Collection, parentKey string, name string) ([]Collection, error) {
    var result []Collection
    for _, collection := range collections {
        collectionParentKey, err := getGroupParentCollectionKey(collection.Data.ParentCollection)
        if err != nil {
            return nil, err
        }

        if collection.Data.Name == name && collectionParentKey == parentKey {
            result = append(result, collection)
        }
    }
    return result, nil
}

func getGroupParentCollectionKey(pc interface{}) (string, error) {
    switch v := pc.(type) {
    case string:
        return v, nil
    case bool:
        if !v { // false indicates a top-level collection
            return "", nil
        }
        return "", fmt.Errorf("unexpected boolean value for parentCollection: true")
    case nil:
        return "", nil
    default:
        return "", fmt.Errorf("unknown type for parentCollection")
    }
}

type Group struct {
    Data struct {
        ID   int    `json:"id"`
        Name string `json:"name"`
        // Include other fields from "data" as needed
    } `json:"data"`
    // Include other top-level fields if necessary
}
