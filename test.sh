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

echo "###### Testing R WRAPPER STRING BRIDGE ######"
echo "==> Testing R .Call string conversion..."
if command -v R >/dev/null 2>&1 && command -v Rscript >/dev/null 2>&1; then
    R_WRAPPER_TEST_DIR=$(mktemp -d)
    cp r-package/src/R_wrapper.c r-package/src/_cgo_export.h "$R_WRAPPER_TEST_DIR/"
    cat > "$R_WRAPPER_TEST_DIR/native_stubs.c" <<'EOF'
#include <stdarg.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static char* copy_string(const char *value) {
    if (value == NULL) {
        value = "<null>";
    }
    size_t size = strlen(value) + 1;
    char *out = (char*)malloc(size);
    if (out == NULL) {
        return NULL;
    }
    memcpy(out, value, size);
    return out;
}

static char* format_string(const char *format, ...) {
    va_list args;
    va_start(args, format);
    int needed = vsnprintf(NULL, 0, format, args);
    va_end(args);
    if (needed < 0) {
        return copy_string("format error");
    }
    char *out = (char*)malloc((size_t)needed + 1);
    if (out == NULL) {
        return NULL;
    }
    va_start(args, format);
    vsnprintf(out, (size_t)needed + 1, format, args);
    va_end(args);
    return out;
}

char* RunReviewR(char* input) { return format_string("review:%s", input); }
char* DownloadZoteroR(char* input) { return format_string("zotero:%s", input); }
char* DownloadURLListR(char* path) { return format_string("url:%s", path); }
char* ConvertR(char* inputDir, char* selectedFormats, char* tikaAddress, char* singleFile, char* ocrOnly) {
    return format_string("convert:%s|%s|%s|%s|%s", inputDir, selectedFormats, tikaAddress, singleFile, ocrOnly);
}
char* ScreeningR(char* input) { return format_string("screening:%s", input); }
char* ValidateConfigR(char* configType, char* input) { return format_string("validate:%s:%s", configType, input); }
char* CheckConformanceR(char* record, char* protocol) { return format_string("conformance:%s:%s", record, protocol); }
char* ProtocolGuidanceR(char* protocol) { return format_string("guidance:%s", protocol); }
char* GenerateRevAIseRecordR(char* paramsJson) { return format_string("generate:%s", paramsJson); }
char* RevAIseSchemaR(char* paramsJson) { return format_string("schema:%s", paramsJson); }
char* MergeRecordStageR(char* record, char* stage) { return format_string("merge:%s:%s", record, stage); }
char* ValidateRecordR(char* record) { return format_string("record:%s", record); }
void FreeCString(char* str) { free(str); }
EOF
    if (cd "$R_WRAPPER_TEST_DIR" && PKG_CPPFLAGS="-DNATIVE_LIBS_AVAILABLE" R CMD SHLIB R_wrapper.c native_stubs.c > build.log 2>&1); then
        if WRAPPER_TEST_DIR="$R_WRAPPER_TEST_DIR" Rscript - <<'EOF'
dyn.load(file.path(Sys.getenv("WRAPPER_TEST_DIR"), paste0("R_wrapper", .Platform$dynlib.ext)))

expect_equal <- function(actual, expected) {
  if (!identical(actual, expected)) {
    stop(sprintf("Expected %s, got %s", shQuote(expected), shQuote(actual)), call. = FALSE)
  }
}

toml <- paste(
  "[zotero]",
  'user = "u"',
  'api_key = "k"',
  'group = "dummy_records"',
  'output_dir = "./papers_zotero"',
  sep = "\n"
)
screening <- paste("[project]", 'name = "screening"', sep = "\n")
record <- '{"review_id":"r1","stages":[]}'
stage <- '{"stage_type":"search"}'
params <- '{"title":"My review","include_manual_stage_stubs":true}'

expect_equal(.Call("check_platform_support"), "supported")
expect_equal(.Call("RunReviewR_wrap", toml), paste0("review:", toml))
expect_equal(.Call("DownloadZoteroR_wrap", toml), paste0("zotero:", toml))
expect_equal(.Call("DownloadURLListR_wrap", "urls.txt"), "url:urls.txt")
expect_equal(.Call("ConvertR_wrap", "papers", "pdf", "localhost:9998", "one.pdf", "true"), "convert:papers|pdf|localhost:9998|one.pdf|true")
expect_equal(.Call("ScreeningR_wrap", screening), paste0("screening:", screening))
expect_equal(.Call("ValidateConfigR_wrap", "zotero", toml), paste0("validate:zotero:", toml))
expect_equal(.Call("CheckConformanceR_wrap", record, "prisma-2020"), paste0("conformance:", record, ":prisma-2020"))
expect_equal(.Call("ProtocolGuidanceR_wrap", "prisma-2020"), "guidance:prisma-2020")
expect_equal(.Call("GenerateRevAIseRecordR_wrap", params), paste0("generate:", params))
expect_equal(.Call("RevAIseSchemaR_wrap", params), paste0("schema:", params))
expect_equal(.Call("MergeRecordStageR_wrap", record, stage), paste0("merge:", record, ":", stage))
expect_equal(.Call("ValidateRecordR_wrap", record), paste0("record:", record))

err <- tryCatch(.Call("ValidateConfigR_wrap", NA_character_, toml), error = conditionMessage)
if (!grepl("configType must not be NA", err, fixed = TRUE)) {
  stop("Expected NA configType validation error, got: ", err, call. = FALSE)
}
EOF
        then
            echo "    ✓ R wrapper string bridge passed"
        else
            echo "    ✗ R wrapper string bridge failed"
        fi
    else
        echo "    ✗ R wrapper string bridge build failed"
        cat "$R_WRAPPER_TEST_DIR/build.log"
    fi
    rm -rf "$R_WRAPPER_TEST_DIR"
else
    echo "    ℹ R not available; skipping R wrapper string bridge test"
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

    # Check if download results file was generated
    if [ -f "$TEMP_DOWNLOAD_DIR/url_list_test_download.csv" ]; then
        echo "    ✓ Download results file generated"
        # Move download results file to outputs for inspection
        cp "$TEMP_DOWNLOAD_DIR/url_list_test_download.csv" projects/test/outputs/download/
        FAILED_COUNT=$(grep -c ",false," "$TEMP_DOWNLOAD_DIR/url_list_test_download.csv" 2>/dev/null || echo "0")
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

    # Check if download results file was generated
    if [ -f "$TEMP_CSV_DIR/csv_test_download.csv" ]; then
        echo "    ✓ Download results file generated"
        # Move download results file to outputs for inspection
        cp "$TEMP_CSV_DIR/csv_test_download.csv" projects/test/outputs/download/

        # Count successful downloads in the results file (excluding header)
        SUCCESS_COUNT=$(grep -c ",true," "$TEMP_CSV_DIR/csv_test_download.csv" 2>/dev/null || echo "0")
        TOTAL_COUNT=$(tail -n +2 "$TEMP_CSV_DIR/csv_test_download.csv" | wc -l 2>/dev/null || echo "0")
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
        echo "    ⚠ Warning: Download results file not found"
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
# Copy the config to the output directory; output_dir controls the download destination
cp projects/test/configs/zotero_test.toml projects/test/outputs/download/zotero_test_temp.toml
REVAISE_RECORD="projects/test/outputs/download/zotero_revaise_record.json"
cat >> projects/test/outputs/download/zotero_test_temp.toml <<EOF

[revaise]
enabled = true
record_file = "$REVAISE_RECORD"
format = "json"

[revaise.stage]
stage_type = "search"
stage_label = "Live Zotero full-text download"
EOF
# Run the download with the TOML configuration
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
    if [ -f "$REVAISE_RECORD" ]; then
        echo "    ✓ RevAIse record created for live Zotero download"
        if grep -q '"kind": "fulltexts"' "$REVAISE_RECORD" && grep -q '"resource_uri": "file://' "$REVAISE_RECORD" && grep -q 'projects/test/outputs/download/zotero' "$REVAISE_RECORD"; then
            echo "    ✓ RevAIse record includes Zotero full-text output"
        else
            echo "    ⚠ Warning: RevAIse record missing Zotero full-text output"
        fi
    else
        echo "    ⚠ Warning: RevAIse record not created"
    fi
else
    echo "    ✗ Zotero download test failed"
fi
echo "###### Testing CONVERSION ######"
echo "==> Testing file conversion functionality..."

# Check if Tika server is available
TIKA_AVAILABLE=false
if curl -s --max-time 2 http://localhost:9998/tika > /dev/null 2>&1; then
    TIKA_AVAILABLE=true
    echo "    ℹ Tika server detected - will be used as fallback when needed"
else
    echo "    ℹ Tika server not available (conversion will use standard methods only)"
fi

# Test PDF conversion on Zotero downloads (before cleanup)
if [ -d "projects/test/outputs/download/zotero" ]; then
    PDF_COUNT=$(ls projects/test/outputs/download/zotero/*.pdf 2>/dev/null | wc -l)
    if [ "$PDF_COUNT" -gt 0 ]; then
        echo "==> Converting $PDF_COUNT Zotero PDFs to TXT..."
        if go run cmd/main.go --convert-pdf projects/test/outputs/download/zotero --tika-server localhost:9998 2>&1 | tee /tmp/conversion_output.log > /dev/null; then
            TXT_COUNT=$(ls projects/test/outputs/download/zotero/*.txt 2>/dev/null | wc -l)
            echo "    ✓ PDF conversion executed"
            echo "    ✓ Converted $TXT_COUNT out of $PDF_COUNT PDFs to TXT"

            # Check if Tika fallback was used
            if grep -q "attempting Tika OCR fallback" /tmp/conversion_output.log; then
                TIKA_USED=$(grep -c "attempting Tika OCR fallback" /tmp/conversion_output.log)
                echo "    ✓ Tika OCR fallback triggered $TIKA_USED times"
            fi

            # Show text file sizes
            for txtfile in projects/test/outputs/download/zotero/*.txt; do
                if [ -f "$txtfile" ]; then
                    SIZE=$(wc -c < "$txtfile")
                    echo "    ✓ $(basename "$txtfile"): $SIZE bytes"
                fi
            done
        else
            echo "    ✗ PDF conversion failed"
        fi

        if [ "$TIKA_AVAILABLE" = true ]; then
            FIRST_PDF=$(ls projects/test/outputs/download/zotero/*.pdf 2>/dev/null | head -1)
            if [ -n "$FIRST_PDF" ]; then
                echo "==> OCR-only conversion for a single PDF..."
                if go run cmd/main.go --convert-pdf projects/test/outputs/download/zotero --single-file "$FIRST_PDF" --ocr-only --tika-server localhost:9998 > /dev/null 2>&1; then
                    echo "    ✓ OCR-only single-file conversion executed"
                else
                    echo "    ✗ OCR-only single-file conversion failed"
                fi
            fi
        fi
    else
        echo "    ℹ No PDFs found in Zotero download directory"
    fi
fi

# Clean up the temporary config and conversion outputs
rm -f projects/test/outputs/download/zotero_test_temp.toml
rm -f /tmp/conversion_output.log
echo "    ✓ Conversion tests completed"

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
