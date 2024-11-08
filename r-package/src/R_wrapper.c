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
