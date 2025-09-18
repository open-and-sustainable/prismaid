#!/bin/bash

# Function to clean test outputs
clean_test_outputs() {
    echo "Cleaning test outputs..."
    # Clean all files but preserve .gitkeep files and directory structure
    find projects/test/outputs -type f ! -name '.gitkeep' -delete
    find projects/test/outputs -type d -empty ! -path projects/test/outputs -delete 2>/dev/null || true
    # Ensure directories exist
    mkdir -p projects/test/outputs/{screening,review,download}
}

# Clean before starting tests
echo "###### Preparing test environment ######"
clean_test_outputs

echo "###### Testing CODE ######"
go test -v ./...

echo "###### Testing SCREENING ######"
go test -v ./screening/logic -run TestScreeningWithBasicConfig
go test -v ./screening/filters -run TestLanguageDetection
go test -v ./screening/filters -run TestArticleTypeClassification
go test -v ./screening/filters -run TestDeduplication

echo "###### Testing SCREENING with config ######"
go run cmd/main.go --screening projects/test/configs/screening_test.toml

echo "###### Testing DOWNLOAD-URL ######"
# Create a temporary directory for downloads to avoid polluting test inputs
TEMP_DOWNLOAD_DIR=$(mktemp -d)
cp projects/test/inputs/download/url_list_test.txt "$TEMP_DOWNLOAD_DIR/"
go run cmd/main.go --download-URL "$TEMP_DOWNLOAD_DIR/url_list_test.txt"
# Move downloaded files to test output
mv "$TEMP_DOWNLOAD_DIR"/*.pdf projects/test/outputs/download/ 2>/dev/null || true
rm -rf "$TEMP_DOWNLOAD_DIR"

echo "###### Testing DOWNLOAD-ZOTERO ######"
go run cmd/main.go --download-zotero projects/test/configs/zotero_test.toml

#echo "###### Testing CONVERSION ######"
# conversion is already tested in go tests

echo "###### Testing REVIEW ######"
go run cmd/main.go --project projects/test/configs/proj_test.toml

# Final cleanup
echo "###### Final cleanup ######"
clean_test_outputs
echo "All tests completed and outputs cleaned"
