#include "_cgo_export.h"  // Header to access Go functions

typedef void* SEXP;

// Manually define R API functions without including R headers
extern SEXP Rf_mkString(const char* str);
extern SEXP Rf_protect(SEXP);
extern void Rf_unprotect(int count);

// Manually define PROTECT and UNPROTECT without including R headers
#define PROTECT(s) Rf_protect(s)
#define UNPROTECT(n) Rf_unprotect(n)

SEXP RunReviewR_wrap(SEXP input) {
    const char *c_input = (const char*)input;  // Cast input as a string
    const char *c_result = RunReviewR((char *)c_input);  // Call the Go function
    SEXP result = Rf_mkString(c_result);
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
    PROTECT(result);
    UNPROTECT(1);
    return result;
}

SEXP DownloadURLListR_wrap(SEXP path) {
    const char *c_path = (const char*)path;
    const char *c_result = DownloadURLListR((char *)c_path);

    SEXP result = Rf_mkString(c_result);
    PROTECT(result);
    UNPROTECT(1);
    return result;
}

SEXP ConvertR_wrap(SEXP inputDir, SEXP selectedFormats) {
    const char *c_inputDir = (const char*)inputDir;
    const char *c_selectedFormats = (const char*)selectedFormats;

    const char *c_result = ConvertR((char *)c_inputDir, (char *)c_selectedFormats);

    SEXP result = Rf_mkString(c_result);
    PROTECT(result);
    UNPROTECT(1);
    return result;
}
