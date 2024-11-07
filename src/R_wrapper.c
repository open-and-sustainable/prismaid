#include <R.h>
#include <Rinternals.h>
#include "_cgo_export.h"  // Header to access Go functions

extern char* __cdecl RunReviewR(char* input);

SEXP RunReviewR_wrap(SEXP input) {
    const char *c_input = CHAR(STRING_ELT(input, 0));
    const char *c_result = RunReviewR((char *)c_input);  // Call the Go function
    SEXP result = PROTECT(mkString(c_result));
    UNPROTECT(1);
    return result;
}
