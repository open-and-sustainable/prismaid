// Include Go headers only when native libraries are available
#ifdef NATIVE_LIBS_AVAILABLE
#include "_cgo_export.h"
#endif

typedef void* SEXP;

// Manually define R API functions without including R headers
extern SEXP Rf_mkString(const char* str);
extern SEXP Rf_protect(SEXP);
extern void Rf_unprotect(int count);

// Manually define PROTECT and UNPROTECT without including R headers
#define PROTECT(s) Rf_protect(s)
#define UNPROTECT(n) Rf_unprotect(n)

// Error message for unsupported platforms
static const char* UNSUPPORTED_PLATFORM_MSG = 
    "Error: prismaid native libraries not available on this platform.\n"
    "Supported platforms: Linux x86_64, Windows x86_64, macOS ARM64.\n"
    "Please use the command-line binary or other language bindings on this platform.";

#ifdef NATIVE_LIBS_AVAILABLE

// Native implementation - call actual Go functions
SEXP RunReviewR_wrap(SEXP input) {
    const char *c_input = (const char*)input;
    const char *c_result = RunReviewR((char *)c_input);
    SEXP result = Rf_mkString(c_result);
    if (c_result) {
        FreeCString((char *)c_result);
    }
    PROTECT(result);
    UNPROTECT(1);
    return result;
}

SEXP DownloadZoteroPDFsR_wrap(SEXP username, SEXP apiKey, SEXP collectionName, SEXP parentDir) {
    const char *c_username = (const char*)username;
    const char *c_apiKey = (const char*)apiKey;
    const char *c_collectionName = (const char*)collectionName;
    const char *c_parentDir = (const char*)parentDir;

    const char *c_result = DownloadZoteroPDFsR(
        (char *)c_username,
        (char *)c_apiKey,
        (char *)c_collectionName,
        (char *)c_parentDir
    );

    SEXP result = Rf_mkString(c_result);
    if (c_result) {
        FreeCString((char *)c_result);
    }
    PROTECT(result);
    UNPROTECT(1);
    return result;
}

SEXP DownloadURLListR_wrap(SEXP path) {
    const char *c_path = (const char*)path;
    const char *c_result = DownloadURLListR((char *)c_path);

    SEXP result = Rf_mkString(c_result);
    if (c_result) {
        FreeCString((char *)c_result);
    }
    PROTECT(result);
    UNPROTECT(1);
    return result;
}

SEXP ConvertR_wrap(SEXP inputDir, SEXP selectedFormats, SEXP tikaAddress, SEXP singleFile, SEXP ocrOnly) {
    const char *c_inputDir = (const char*)inputDir;
    const char *c_selectedFormats = (const char*)selectedFormats;
    const char *c_tikaAddress = (const char*)tikaAddress;
    const char *c_singleFile = (const char*)singleFile;
    const char *c_ocrOnly = (const char*)ocrOnly;

    const char *c_result = ConvertR((char *)c_inputDir, (char *)c_selectedFormats, (char *)c_tikaAddress, (char *)c_singleFile, (char *)c_ocrOnly);

    SEXP result = Rf_mkString(c_result);
    if (c_result) {
        FreeCString((char *)c_result);
    }
    PROTECT(result);
    UNPROTECT(1);
    return result;
}

SEXP ScreeningR_wrap(SEXP input) {
    const char *c_input = (const char*)input;
    const char *c_result = ScreeningR((char *)c_input);
    
    SEXP result = Rf_mkString(c_result);
    if (c_result) {
        FreeCString((char *)c_result);
    }
    PROTECT(result);
    UNPROTECT(1);
    return result;
}

#else

// Stub implementation for unsupported platforms
// These functions return informative error messages
SEXP RunReviewR_wrap(SEXP input) {
    SEXP result = Rf_mkString(UNSUPPORTED_PLATFORM_MSG);
    PROTECT(result);
    UNPROTECT(1);
    return result;
}

SEXP DownloadZoteroPDFsR_wrap(SEXP username, SEXP apiKey, SEXP collectionName, SEXP parentDir) {
    SEXP result = Rf_mkString(UNSUPPORTED_PLATFORM_MSG);
    PROTECT(result);
    UNPROTECT(1);
    return result;
}

SEXP DownloadURLListR_wrap(SEXP path) {
    SEXP result = Rf_mkString(UNSUPPORTED_PLATFORM_MSG);
    PROTECT(result);
    UNPROTECT(1);
    return result;
}

SEXP ConvertR_wrap(SEXP inputDir, SEXP selectedFormats, SEXP tikaAddress, SEXP singleFile, SEXP ocrOnly) {
    SEXP result = Rf_mkString(UNSUPPORTED_PLATFORM_MSG);
    PROTECT(result);
    UNPROTECT(1);
    return result;
}

SEXP ScreeningR_wrap(SEXP input) {
    SEXP result = Rf_mkString(UNSUPPORTED_PLATFORM_MSG);
    PROTECT(result);
    UNPROTECT(1);
    return result;
}

#endif

// Platform detection function for R to call
SEXP check_platform_support() {
#ifdef NATIVE_LIBS_AVAILABLE
    SEXP result = Rf_mkString("supported");
#else
    SEXP result = Rf_mkString("unsupported");
#endif
    PROTECT(result);
    UNPROTECT(1);
    return result;
}
