// _cgo_export.h

#ifndef CGO_EXPORT_H
#define CGO_EXPORT_H

#ifdef _WIN32
    #define EXPORT __declspec(dllimport)
#else
    #define EXPORT
#endif

#ifdef __cplusplus
extern "C" {
#endif

EXPORT const char* __stdcall RunReviewR(char* input);

#ifdef __cplusplus
}
#endif

#endif // CGO_EXPORT_H
