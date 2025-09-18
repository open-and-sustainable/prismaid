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
echo "==> Running screening unit tests..."
if go test ./screening/logic ./screening/filters -v > /dev/null 2>&1; then
    echo "    ✓ All screening unit tests passed"
else
    echo "    ✗ Some screening unit tests failed"
fi

echo "###### Testing SCREENING with config ######"
echo "==> Running screening on test manuscripts..."
if go run cmd/main.go --screening projects/test/configs/screening_test.toml > /tmp/screening_output.log 2>&1; then
    echo "    ✓ Screening completed successfully"

    OUTPUT_FILE="projects/test/outputs/screening/test_screening_output.csv"
    if [ -f "$OUTPUT_FILE" ]; then
        echo ""
        echo "╔════════════════════════════════════════════════════════════╗"
        echo "║              MANUSCRIPT SCREENING SUMMARY                   ║"
        echo "╚════════════════════════════════════════════════════════════╝"

        # Use Python for accurate CSV parsing
        python3 -c "
import csv
import sys

with open('$OUTPUT_FILE', 'r') as f:
    reader = csv.DictReader(f)
    rows = list(reader)

    total = len(rows)
    duplicates = sum(1 for r in rows if r.get('tag_is_duplicate') == 'true')
    included = sum(1 for r in rows if r.get('include') == 'true')

    # Count exclusion reasons
    lang_excluded = sum(1 for r in rows if 'Language not accepted' in r.get('exclusion_reason', ''))
    editorial_excluded = sum(1 for r in rows if 'Editorial' in r.get('exclusion_reason', ''))
    letter_excluded = sum(1 for r in rows if 'Letter' in r.get('exclusion_reason', ''))
    review_excluded = sum(1 for r in rows if 'Review' in r.get('exclusion_reason', ''))

    article_excluded = editorial_excluded + letter_excluded + review_excluded

    print('')
    print('  📚 INITIAL POOL:')
    print(f'     Total manuscripts loaded: {total}')

    print('')
    print('  🔄 DEDUPLICATION FILTER:')
    print(f'     Duplicates removed: {duplicates}')
    print(f'     Remaining: {total - duplicates}')

    print('')
    print('  🌐 LANGUAGE FILTER:')
    print(f'     Non-English removed: {lang_excluded}')
    print(f'     Remaining: {total - duplicates - lang_excluded}')

    if article_excluded > 0:
        print('')
        print('  📝 ARTICLE TYPE FILTER:')
        print(f'     Excluded types: {article_excluded}')
        if editorial_excluded > 0:
            print(f'       - Editorials: {editorial_excluded}')
        if letter_excluded > 0:
            print(f'       - Letters: {letter_excluded}')
        if review_excluded > 0:
            print(f'       - Reviews: {review_excluded}')
        print(f'     Remaining: {total - duplicates - lang_excluded - article_excluded}')

    print('')
    print('  ═══════════════════════════════════════════════════════════')
    print('')
    print('  📊 FINAL RESULTS:')
    print(f'     ✅ Manuscripts included: {included}')
    print(f'     ❌ Total excluded: {total - included}')
    if total > 0:
        percentage = int(included * 100 / total)
        print(f'     📈 Inclusion rate: {percentage}%')

    print('')
    print('  ═══════════════════════════════════════════════════════════')
    print('')
    print(f'  💾 Output saved to: $OUTPUT_FILE')
    print('')
"
    else
        echo "    ⚠ Warning: Screening output file not found"
    fi
else
    echo "    ✗ Screening with config test failed"
    echo "    Check /tmp/screening_output.log for details"
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
