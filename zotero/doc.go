// Package zotero provides a set of functions and types for interacting
// with the Zotero API. Zotero is a free, easy-to-use tool to help you
// collect, organize, cite, and share research.
//
// The zotero package simplifies the process of fetching collection keys,
// downloading attachments, and handling user authentication for API requests.
// It leverages standard library components for HTTP requests and provides
// utility functions to handle common tasks such as:
//
// - Retrieving collection keys based on collection names, supporting nested structures.
// - Downloading all PDFs from specified Zotero collections or shared groups, including nested collections.
// - Automatically managing API request headers and response status codes.
// - Converting downloaded PDFs into text files automatically.
// - Reviewing converted text files through API calls to AI models.
//
// **Nested Structures and Shared Groups**
//
// This package supports nested collections and groups by allowing the use of
// path-like expressions to represent nested structures. Collections and groups
// can be specified using a forward-slash ('/') to denote levels of nesting.
// For example, to access a sub-collection named "SubCollection" within a collection
// named "Collection", you would use `"Collection/SubCollection"`. Similarly, groups
// can be specified in the same way, with the group name followed by any nested
// collections.
//
// **PDF Conversion and AI Review**
//
// After downloading PDFs from Zotero, the package automatically converts them into
// text files. These text files can then be processed and reviewed via API calls to
// AI models, facilitating tasks such as text analysis, summarization, or other
// AI-driven functionalities.
//
// **Usage**
//
// To use this package, create a client instance and use it to call methods
// provided by the package. For example, to download PDFs from a specific
// collection or group:
//
//     client := zotero.NewClient(apiKey, userID)
//     err := client.DownloadPDFs("collectionName", "parentDir")
//     if err != nil {
//         log.Fatal(err)
//     }
//
// **Parameters:**
//
// - `collectionName`: The name or path of the collection or group. It can be a simple collection name, a nested
//   collection path (e.g., `"Collection/SubCollection"`), or a group name with optional nested collections
//   (e.g., `"GroupName/Collection"`). The package will attempt to find the collection in the user's library first;
//   if not found, it will search in the user's groups.
// - `parentDir`: The directory where the downloaded PDF files will be stored. After downloading, the PDFs in this
//   directory will be automatically converted into text files for further processing. These text files can then be
//   reviewed and analyzed using API calls to AI models or other processing tools.
//
// **Example with Nested Collection and Group**
//
//     client := zotero.NewClient(apiKey, userID)
//     err := client.DownloadPDFs("GroupName/Collection/SubCollection", "/path/to/download/directory")
//     if err != nil {
//         log.Fatal(err)
//     }
//
// In this example, PDFs from the specified sub-collection within the group `"GroupName"`
// are downloaded to the directory `"/path/to/download/directory"`, converted into text files,
// and prepared for review or analysis.
//
// **Note on API Limits**
//
// The Zotero API has rate limits and usage restrictions. Ensure that you are
// aware of these when using this package to make frequent or large numbers of
// requests. For more information, refer to the official Zotero API documentation
// at https://www.zotero.org/support/dev/web_api/v3/start.
package zotero
