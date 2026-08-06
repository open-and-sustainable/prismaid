// Include Go headers only when native libraries are available
#ifdef NATIVE_LIBS_AVAILABLE
#include "_cgo_export.h"
#endif

#include <Rinternals.h>
#include <string.h>

static const char *as_c_string(SEXP value, const char *name) {
    if (TYPEOF(value) != STRSXP || XLENGTH(value) < 1) {
        Rf_error("%s must be a character vector with at least one value", name);
    }
    SEXP element = STRING_ELT(value, 0);
    if (element == NA_STRING) {
        Rf_error("%s must not be NA", name);
    }
    return CHAR(element);
}

static SEXP string_result(const char *message) {
    SEXP result = PROTECT(Rf_mkString(message ? message : ""));
    UNPROTECT(1);
    return result;
}

// Error message for unsupported platforms
static const char* UNSUPPORTED_PLATFORM_MSG = 
    "Error: prismaid native libraries not available on this platform.\n"
    "Supported platforms: Linux x86_64, Windows x86_64, macOS ARM64.\n"
    "Please use the command-line binary or other language bindings on this platform.";

#ifdef NATIVE_LIBS_AVAILABLE

static SEXP native_result(const char *c_result) {
    SEXP result = PROTECT(Rf_mkString(c_result ? c_result : "Unknown native error"));
    if (c_result) {
        FreeCString((char *)c_result);
    }
    UNPROTECT(1);
    return result;
}

// Native implementation - call actual Go functions
SEXP RunReviewR_wrap(SEXP input) {
    const char *c_input = as_c_string(input, "input");
    const char *c_result = RunReviewR((char *)c_input);
    return native_result(c_result);
}

SEXP DownloadZoteroR_wrap(SEXP input) {
    const char *c_input = as_c_string(input, "input");
    const char *c_result = DownloadZoteroR((char *)c_input);
    return native_result(c_result);
}

SEXP DownloadURLListR_wrap(SEXP path) {
    const char *c_path = as_c_string(path, "path");
    const char *c_result = DownloadURLListR((char *)c_path);
    return native_result(c_result);
}

SEXP ConvertR_wrap(SEXP inputDir, SEXP selectedFormats, SEXP tikaAddress, SEXP singleFile, SEXP ocrOnly) {
    const char *c_inputDir = as_c_string(inputDir, "inputDir");
    const char *c_selectedFormats = as_c_string(selectedFormats, "selectedFormats");
    const char *c_tikaAddress = as_c_string(tikaAddress, "tikaAddress");
    const char *c_singleFile = as_c_string(singleFile, "singleFile");
    const char *c_ocrOnly = as_c_string(ocrOnly, "ocrOnly");

    const char *c_result = ConvertR((char *)c_inputDir, (char *)c_selectedFormats, (char *)c_tikaAddress, (char *)c_singleFile, (char *)c_ocrOnly);
    return native_result(c_result);
}

SEXP ScreeningR_wrap(SEXP input) {
    const char *c_input = as_c_string(input, "input");
    const char *c_result = ScreeningR((char *)c_input);
    return native_result(c_result);
}

SEXP ValidateConfigR_wrap(SEXP configType, SEXP input) {
    const char *c_configType = as_c_string(configType, "configType");
    const char *c_input = as_c_string(input, "input");
    const char *c_result = ValidateConfigR((char *)c_configType, (char *)c_input);
    return native_result(c_result);
}

SEXP CheckConformanceR_wrap(SEXP record, SEXP protocol) {
    const char *c_record = as_c_string(record, "record");
    const char *c_protocol = as_c_string(protocol, "protocol");
    const char *c_result = CheckConformanceR((char *)c_record, (char *)c_protocol);
    return native_result(c_result);
}

SEXP ProtocolGuidanceR_wrap(SEXP protocol) {
    const char *c_protocol = as_c_string(protocol, "protocol");
    const char *c_result = ProtocolGuidanceR((char *)c_protocol);
    return native_result(c_result);
}

SEXP GenerateRevAIseRecordR_wrap(SEXP paramsJson) {
    const char *c_params = as_c_string(paramsJson, "paramsJson");
    const char *c_result = GenerateRevAIseRecordR((char *)c_params);
    return native_result(c_result);
}

SEXP RevAIseSchemaR_wrap(SEXP paramsJson) {
    const char *c_params = as_c_string(paramsJson, "paramsJson");
    const char *c_result = RevAIseSchemaR((char *)c_params);
    return native_result(c_result);
}

SEXP MergeRecordStageR_wrap(SEXP record, SEXP stage) {
    const char *c_record = as_c_string(record, "record");
    const char *c_stage = as_c_string(stage, "stage");
    const char *c_result = MergeRecordStageR((char *)c_record, (char *)c_stage);
    return native_result(c_result);
}

SEXP ValidateRecordR_wrap(SEXP record) {
    const char *c_record = as_c_string(record, "record");
    const char *c_result = ValidateRecordR((char *)c_record);
    return native_result(c_result);
}

#else

// Stub implementation for unsupported platforms
// These functions return informative error messages
SEXP RunReviewR_wrap(SEXP input) {
    return string_result(UNSUPPORTED_PLATFORM_MSG);
}

SEXP DownloadZoteroR_wrap(SEXP input) {
    return string_result(UNSUPPORTED_PLATFORM_MSG);
}

SEXP DownloadURLListR_wrap(SEXP path) {
    return string_result(UNSUPPORTED_PLATFORM_MSG);
}

SEXP ConvertR_wrap(SEXP inputDir, SEXP selectedFormats, SEXP tikaAddress, SEXP singleFile, SEXP ocrOnly) {
    return string_result(UNSUPPORTED_PLATFORM_MSG);
}

SEXP ScreeningR_wrap(SEXP input) {
    return string_result(UNSUPPORTED_PLATFORM_MSG);
}

SEXP ValidateConfigR_wrap(SEXP configType, SEXP input) {
    return string_result(UNSUPPORTED_PLATFORM_MSG);
}

SEXP CheckConformanceR_wrap(SEXP record, SEXP protocol) {
    return string_result(UNSUPPORTED_PLATFORM_MSG);
}

SEXP ProtocolGuidanceR_wrap(SEXP protocol) {
    return string_result(UNSUPPORTED_PLATFORM_MSG);
}

SEXP GenerateRevAIseRecordR_wrap(SEXP paramsJson) {
    return string_result(UNSUPPORTED_PLATFORM_MSG);
}

SEXP RevAIseSchemaR_wrap(SEXP paramsJson) {
    return string_result(UNSUPPORTED_PLATFORM_MSG);
}

SEXP MergeRecordStageR_wrap(SEXP record, SEXP stage) {
    return string_result(UNSUPPORTED_PLATFORM_MSG);
}

SEXP ValidateRecordR_wrap(SEXP record) {
    return string_result(UNSUPPORTED_PLATFORM_MSG);
}

#endif

// Platform detection function for R to call
SEXP check_platform_support() {
#ifdef NATIVE_LIBS_AVAILABLE
    return string_result("supported");
#else
    return string_result("unsupported");
#endif
}
