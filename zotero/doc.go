
// Package zotero provides a set of functions and types for interacting
// with the Zotero API. Zotero is a free, easy-to-use tool to help you
// collect, organize, cite, and share research.
//
// The zotero package simplifies the process of fetching collection keys,
// downloading attachments, and handling user authentication for API requests.
// It leverages standard library components for HTTP requests and provides
// utility functions to handle common tasks such as:
//
// - Retrieving collection keys based on collection names.
// - Downloading all PDFs from a specified Zotero group or collection.
// - Automatically managing API request headers and response status codes.
//
// Usage:
//
// To use this package, create a client instance and use it to call methods
// provided by the package. For example, to download PDFs from a specific
// collection:
//
//     client := zotero.NewClient(apiKey, userID)
//     err := client.DownloadPDFs("collectionName", "parentDir")
//     if err != nil {
//         log.Fatal(err)
//     }
//
// The client handles setting up HTTP requests, error handling, and the parsing
// of JSON responses from the Zotero API. Users of the package need to provide
// their own API key and user ID to authenticate requests.
//
// This package requires a minimum of Go 1.11 due to the use of modules.
//
// Note:
//
// The Zotero API has rate limits and usage restrictions. Ensure that you are
// aware of these when using this package to make frequent or large numbers of
// requests. For more information, refer to the official Zotero API documentation
// at https://www.zotero.org/support/dev/web_api/v3/start.
package zotero
