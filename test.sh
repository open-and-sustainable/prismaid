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
echo "==> Testing screening functionality..."
if go run cmd/main.go --screening projects/test/configs/screening_test.toml > /dev/null 2>&1; then
    echo "    ✓ Screening command executed"
    if [ -f "projects/test/outputs/screening/test_screening_output.csv" ]; then
        echo "    ✓ Screening output file created"
    else
        echo "    ⚠ Warning: Screening output file not found"
    fi
else
    echo "    ✗ Screening test failed"
fi

echo "###### Testing DOWNLOAD-URL ######"
echo "==> Testing URL downloads from TXT file..."
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

    # Check if failed URLs log was generated
    if [ -f "$TEMP_DOWNLOAD_DIR/url_list_test_failed.txt" ]; then
        echo "    ✓ Failed URLs log generated"
        # Move failed URLs log to outputs for inspection
        cp "$TEMP_DOWNLOAD_DIR/url_list_test_failed.txt" projects/test/outputs/download/
        FAILED_COUNT=$(grep -c "^https" "$TEMP_DOWNLOAD_DIR/url_list_test_failed.txt" 2>/dev/null || echo "0")
        if [ "$FAILED_COUNT" -gt 0 ]; then
            echo "    ✓ Logged $FAILED_COUNT failed URLs"
        fi
    fi
else
    echo "    ✗ URL download test failed"
fi
rm -rf "$TEMP_DOWNLOAD_DIR"

echo "###### Testing DOWNLOAD-CSV ######"
echo "==> Testing CSV downloads with problematic URL detection..."
# Create a temporary directory for CSV downloads
TEMP_CSV_DIR=$(mktemp -d)
TEMP_CSV_LOG="$TEMP_CSV_DIR/download_log.txt"
cp projects/test/inputs/download/csv_test.csv "$TEMP_CSV_DIR/"
if go run cmd/main.go --download-URL "$TEMP_CSV_DIR/csv_test.csv" 2>&1 | tee "$TEMP_CSV_LOG"; then
    echo "    ✓ CSV download command executed"

    # Check if report was generated
    if [ -f "$TEMP_CSV_DIR/csv_test_report.csv" ]; then
        echo "    ✓ Download report generated"
        # Move report to outputs for inspection
        cp "$TEMP_CSV_DIR/csv_test_report.csv" projects/test/outputs/download/

        # Count successful downloads in the report (excluding header)
        SUCCESS_COUNT=$(grep -c ",true," "$TEMP_CSV_DIR/csv_test_report.csv" 2>/dev/null || echo "0")
        TOTAL_COUNT=$(tail -n +2 "$TEMP_CSV_DIR/csv_test_report.csv" | wc -l 2>/dev/null || echo "0")
        echo "    ✓ Downloaded $SUCCESS_COUNT out of $TOTAL_COUNT papers"

        # Check for problematic URL detection in the log output
        if [ -f "$TEMP_CSV_LOG" ]; then
            PROBLEMATIC_COUNT=$(grep -c "Detected problematic URL" "$TEMP_CSV_LOG" 2>/dev/null || echo "0")
            CROSSREF_COUNT=$(grep -c "Found DOI via Crossref" "$TEMP_CSV_LOG" 2>/dev/null || echo "0")
            if [ "$PROBLEMATIC_COUNT" -gt 0 ]; then
                echo "    ✓ Detected $PROBLEMATIC_COUNT problematic URLs (Dimensions/ResearchGate/Academia/SemanticScholar)"
                if [ "$CROSSREF_COUNT" -gt 0 ]; then
                    echo "    ✓ Resolved $CROSSREF_COUNT DOIs via Crossref API"
                fi
            fi
        fi
    else
        echo "    ⚠ Warning: Download report not found"
    fi

    # Check if enhanced CSV with download status was generated
    if [ -f "$TEMP_CSV_DIR/csv_test_with_status.csv" ]; then
        echo "    ✓ Enhanced CSV with download status generated"
        # Move enhanced CSV to outputs for inspection
        cp "$TEMP_CSV_DIR/csv_test_with_status.csv" projects/test/outputs/download/
    else
        echo "    ⚠ Warning: Enhanced CSV with download status not found"
    fi

    # Move downloaded PDFs to test output
    if ls "$TEMP_CSV_DIR"/*.pdf 2>/dev/null | head -1 > /dev/null; then
        PDF_COUNT=$(ls "$TEMP_CSV_DIR"/*.pdf 2>/dev/null | wc -l)
        mv "$TEMP_CSV_DIR"/*.pdf projects/test/outputs/download/ 2>/dev/null
        echo "    ✓ $PDF_COUNT PDF files saved to output directory"

        # Verify intelligent file naming (should contain year, author, title)
        FIRST_PDF=$(ls projects/test/outputs/download/*.pdf 2>/dev/null | head -1)
        if [ -n "$FIRST_PDF" ]; then
            BASENAME=$(basename "$FIRST_PDF")
            if [[ "$BASENAME" =~ [0-9]{4}_ ]]; then
                echo "    ✓ Intelligent file naming working (detected year prefix)"
            fi
        fi
    else
        echo "    ⚠ Warning: No PDF files were downloaded from CSV"
    fi
else
    echo "    ✗ CSV download test failed"
fi
rm -rf "$TEMP_CSV_DIR"

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

# Final cleanup
echo ""
echo "###### Final cleanup ######"
clean_test_outputs
echo ""
echo "======================================"
echo "All tests completed and outputs cleaned"
echo "======================================"
