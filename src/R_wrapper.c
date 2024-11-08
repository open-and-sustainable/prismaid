#include "_cgo_export.h"  // Header to access Go functions

typedef void* SEXP;

extern SEXP mkString(const char* str);
extern void PROTECT(SEXP);
extern void UNPROTECT(int count);

SEXP RunReviewR_wrap(SEXP input) {
    const char *c_input = (const char*)input;  // Cast input as a string
    const char *c_result = RunReviewR((char *)c_input);  // Call the Go function
    SEXP result = mkString(c_result);
    PROTECT(result);
    UNPROTECT(1);
    return result;
}
