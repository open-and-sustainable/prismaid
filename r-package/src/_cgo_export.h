// _cgo_export.h

#ifndef CGO_EXPORT_H
#define CGO_EXPORT_H

// Declaration of Go functions exposed to C
char* RunReviewR(char* input);
char* DownloadZoteroR(char* input);
char* DownloadURLListR(char* path);
char* ConvertR(char* inputDir, char* selectedFormats, char* tikaAddress, char* singleFile, char* ocrOnly);
char* ScreeningR(char* input);
char* ValidateConfigR(char* configType, char* input);
char* CheckConformanceR(char* record, char* protocol);
char* ProtocolGuidanceR(char* protocol);
void FreeCString(char* str);

#endif // CGO_EXPORT_H
