// Package list provides functionality for downloading PDFs from URL lists, CSV files, or TSV files.
//
// The package supports three input formats:
//
// 1. Plain text files (.txt or no extension):
//   - One URL per line
//   - Lines starting with '#' are treated as comments
//   - Empty lines are ignored
//
// 2. CSV files (.csv):
//   - Comma-separated values with automatic column detection
//   - Intelligent parsing of paper metadata
//   - Generates meaningful filenames from metadata
//
// 3. TSV files (.tsv):
//   - Tab-separated values with automatic column detection
//   - Same features as CSV support
//
// Column Detection:
//
// The package automatically detects and uses the following columns if present:
//   - URL/Link columns: BestLink, BestURL, URL, Link, href (prioritizes "best" variants)
//   - DOI column: Automatically converts DOIs to resolvable URLs if no direct URL is found
//   - Title column: ArticleTitle, Article_Title, Paper_Title, Title
//   - Authors column: Authors, Creator, Contributor
//   - Year column: PublicationYear, Publication_Year, Year
//   - Journal column: SourceTitle, Source_Title, Journal, Source, Publication
//   - Abstract column: Abstract (for future use)
//
// File Naming:
//
// For CSV/TSV inputs, the package generates intelligent filenames using available metadata:
//   - Format: [Year]_[FirstAuthorLastName]_[TruncatedTitle].pdf
//   - Example: 2023_Smith_Climate_change_impacts.pdf
//   - Falls back to row ID if metadata is insufficient
//
// Output:
//
// For CSV/TSV inputs, the package generates a download report with:
//   - Original metadata
//   - Download success status
//   - Generated filename
//   - Error messages for failed downloads
//
// The report is saved as [input_filename]_report.csv in the same directory.
//
// Usage:
//
//	// Download from plain text URL list
//	err := list.DownloadURLList("urls.txt")
//
//	// Download from CSV with metadata
//	err := list.DownloadURLList("papers.csv")
//
//	// Download from TSV with metadata
//	err := list.DownloadURLList("papers.tsv")
//
// Example CSV Input:
//
//	ArticleTitle,Authors,PublicationYear,BestLink,DOI
//	"Climate Change Impacts","Smith, J.; Jones, M.",2023,https://example.com/paper1.pdf,10.1234/abc
//	"Machine Learning Review","Brown, A.",2024,,10.5678/def
//
// The package will:
// 1. Parse the CSV/TSV structure
// 2. Detect column mappings automatically
// 3. Use BestLink when available, or resolve DOI to URL
// 4. Generate filename like "2023_Smith_Climate_Change_Impacts.pdf"
// 5. Download PDFs to the same directory as the input file
// 6. Create a report showing success/failure for each paper
package list
