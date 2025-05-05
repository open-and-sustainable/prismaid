// _cgo_export.h

#ifndef CGO_EXPORT_H
#define CGO_EXPORT_H

// Declaration of Go functions exposed to C
char* RunReviewR(char* input);
char* DownloadZoteroPDFsR(char* username, char* apiKey, char* collectionName, char* parentDir);
char* DownloadURLListR(char* path);
char* ConvertR(char* inputDir, char* selectedFormats);
void FreeCString(char* str);

#endif // CGO_EXPORT_H
