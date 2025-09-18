#!/bin/bash

# Function to clean test outputs
clean_test_outputs() {
    echo "==> Cleaning test outputs..."
    # Clean all files but preserve .gitkeep files and directory structure
    find projects/test/outputs -type f ! -name '.gitkeep' -delete
    # Clean any zotero directories that may have been created in wrong places
    rm -rf projects/test/configs/zotero 2>/dev/null || true
    # Clean the zotero subdirectory in download outputs
    rm -rf projects/test/outputs/download/zotero 2>/dev/null || true
    find projects/test/outputs -type d -empty ! -path projects/test/outputs -delete 2>/dev/null || true
    # Ensure directories exist
    mkdir -p projects/test/outputs/{screening,review,download}
    echo "    ✓ Test outputs cleaned"
}

# Clean before starting tests
echo "###### Preparing test environment ######"
clean_test_outputs

echo "###### Testing CODE ######"
if go test -v ./...; then
    echo "    ✓ Code tests passed"
else
    echo "    ✗ Code tests failed"
fi

echo "###### Testing SCREENING ######"
echo "==> Running screening logic tests..."
if go test -v ./screening/logic -run TestScreeningWithBasicConfig; then
    echo "    ✓ Screening logic test passed"
else
    echo "    ✗ Screening logic test failed"
fi

echo "==> Running language detection tests..."
if go test -v ./screening/filters -run TestLanguageDetection; then
    echo "    ✓ Language detection test passed"
else
    echo "    ✗ Language detection test failed (may not be implemented yet)"
fi

echo "==> Running article type classification tests..."
if go test -v ./screening/filters -run TestArticleTypeClassification; then
    echo "    ✓ Article type classification test passed"
else
    echo "    ✗ Article type classification test failed (may not be implemented yet)"
fi

echo "==> Running deduplication tests..."
if go test -v ./screening/filters -run TestDeduplication; then
    echo "    ✓ Deduplication test passed"
else
    echo "    ✗ Deduplication test failed (may not be implemented yet)"
fi

echo "###### Testing SCREENING with config ######"
if go run cmd/main.go --screening projects/test/configs/screening_test.toml; then
    echo "    ✓ Screening with config test passed"
    if [ -f "projects/test/outputs/screening/test_screening_output.csv" ]; then
        echo "    ✓ Screening output file created"
    else
        echo "    ⚠ Warning: Screening output file not found"
    fi
else
    echo "    ✗ Screening with config test failed"
fi

: <<'COMMENT'
echo "###### Testing DOWNLOAD-URL ######"
echo "==> Testing URL downloads..."
# Create a temporary directory for downloads to avoid polluting test inputs
TEMP_DOWNLOAD_DIR=$(mktemp -d)
cp projects/test/inputs/download/url_list_test.txt "$TEMP_DOWNLOAD_DIR/"
if go run cmd/main.go --download-URL "$TEMP_DOWNLOAD_DIR/url_list_test.txt"; then
    echo "    ✓ URL download command executed"
    # Move downloaded files to test output
    if ls "$TEMP_DOWNLOAD_DIR"/*.pdf 2>/dev/null | head -1 > /dev/null; then
        mv "$TEMP_DOWNLOAD_DIR"/*.pdf projects/test/outputs/download/ 2>/dev/null
        echo "    ✓ PDF files downloaded successfully"
    else
        echo "    ⚠ Warning: No PDF files were downloaded"
    fi
else
    echo "    ✗ URL download test failed"
fi
rm -rf "$TEMP_DOWNLOAD_DIR"

echo "###### Testing DOWNLOAD-ZOTERO ######"
echo "==> Testing Zotero downloads..."
# Copy the config to the output directory so files are downloaded there
cp projects/test/configs/zotero_test.toml projects/test/outputs/download/zotero_test_temp.toml
# Run the download with the config in the output directory
if go run cmd/main.go --download-zotero projects/test/outputs/download/zotero_test_temp.toml; then
    echo "    ✓ Zotero download command executed"
    # Check if zotero directory was created in the output directory
    if [ -d "projects/test/outputs/download/zotero" ]; then
        echo "    ✓ Zotero files downloaded to correct location"
        # List the downloaded files for verification
        FILE_COUNT=$(ls projects/test/outputs/download/zotero/ 2>/dev/null | wc -l)
        echo "    ✓ Downloaded $FILE_COUNT files from Zotero"
    else
        echo "    ⚠ Warning: Zotero directory not found in expected location"
    fi
else
    echo "    ✗ Zotero download test failed"
fi
# Clean up the temporary config
rm -f projects/test/outputs/download/zotero_test_temp.toml

#echo "###### Testing CONVERSION ######"
# conversion is already tested in go tests

echo "###### Testing REVIEW ######"
echo "==> Testing review functionality..."
if go run cmd/main.go --project projects/test/configs/proj_test.toml; then
    echo "    ✓ Review command executed"
    if [ -f "projects/test/outputs/review/test_results.csv" ]; then
        echo "    ✓ Review output file created"
    else
        echo "    ⚠ Warning: Review output file not found"
    fi
else
    echo "    ✗ Review test failed"
fi

COMMENT
# Final cleanup
echo ""
echo "###### Final cleanup ######"
clean_test_outputs
echo ""
echo "======================================"
echo "All tests completed and outputs cleaned"
echo "======================================"
