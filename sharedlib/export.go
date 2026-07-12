package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"encoding/json"
	"fmt"
	"strings"
	"unsafe"

	"github.com/open-and-sustainable/prismaid"
)

// Common error handling and memory management functions
func handlePanic() *C.char {
	if r := recover(); r != nil {
		fmt.Println("Recovered from panic:", r)
		return C.CString(fmt.Sprint(r))
	}
	return nil
}

// common logic as an helper function
func runReview(input *C.char) error {
	goInput := C.GoString(input)
	_, err := prismaid.Review(goInput)
	return err
}

func runDownloadZotero(input *C.char) error {
	goInput := C.GoString(input)
	_, err := prismaid.DownloadZotero(goInput)
	return err
}

func runDownloadURLList(path *C.char) error {
	goPath := C.GoString(path)
	_, err := prismaid.DownloadURLList(goPath)
	return err
}

func runConvert(inputDir, selectedFormats, tikaAddress, singleFile, ocrOnly *C.char) error {
	goInputDir := C.GoString(inputDir)
	goSelectedFormats := C.GoString(selectedFormats)
	goTikaAddress := C.GoString(tikaAddress)
	goSingleFile := C.GoString(singleFile)
	goOcrOnly := strings.TrimSpace(strings.ToLower(C.GoString(ocrOnly)))
	ocrOnlyEnabled := goOcrOnly == "1" || goOcrOnly == "true" || goOcrOnly == "yes"
	_, err := prismaid.Convert(goInputDir, goSelectedFormats, prismaid.ConvertOptions{
		TikaServer: goTikaAddress,
		PDF: prismaid.PDFOptions{
			SingleFile: goSingleFile,
			OCROnly:    ocrOnlyEnabled,
		},
	})
	return err
}

func runScreening(input *C.char) error {
	goInput := C.GoString(input)
	_, err := prismaid.Screening(goInput)
	return err
}

func runValidate(configType, input *C.char) error {
	goConfigType := C.GoString(configType)
	goInput := C.GoString(input)
	return prismaid.ValidateConfig(goConfigType, goInput)
}

// runCheckConformance runs a protocol conformance check and returns the report
// as a JSON string. On error it returns a JSON object with an "error" field.
func runCheckConformance(record, protocol *C.char) string {
	report, err := prismaid.CheckConformance(C.GoString(record), C.GoString(protocol))
	if err != nil {
		data, _ := json.Marshal(map[string]string{"error": err.Error()})
		return string(data)
	}
	data, err := json.Marshal(report)
	if err != nil {
		errData, _ := json.Marshal(map[string]string{"error": err.Error()})
		return string(errData)
	}
	return string(data)
}

// runProtocolGuidance returns a protocol's requirement checklist as a JSON
// string. On error it returns a JSON object with an "error" field.
func runProtocolGuidance(protocol *C.char) string {
	guidance, err := prismaid.ProtocolGuidance(C.GoString(protocol))
	if err != nil {
		data, _ := json.Marshal(map[string]string{"error": err.Error()})
		return string(data)
	}
	data, err := json.Marshal(guidance)
	if err != nil {
		errData, _ := json.Marshal(map[string]string{"error": err.Error()})
		return string(errData)
	}
	return string(data)
}

// Python-specific function
//
//export RunReviewPython
func RunReviewPython(input *C.char) *C.char {
	defer handlePanic()
	if err := runReview(input); err != nil {
		return C.CString(err.Error())
	}
	return nil
}

//export DownloadZoteroPython
func DownloadZoteroPython(input *C.char) *C.char {
	defer handlePanic()
	if err := runDownloadZotero(input); err != nil {
		return C.CString(err.Error())
	}
	return nil
}

//export DownloadURLListPython
func DownloadURLListPython(path *C.char) *C.char {
	defer handlePanic()
	if err := runDownloadURLList(path); err != nil {
		return C.CString(err.Error())
	}
	return nil
}

//export ConvertPython
func ConvertPython(inputDir, selectedFormats, tikaAddress, singleFile, ocrOnly *C.char) *C.char {
	defer handlePanic()
	if err := runConvert(inputDir, selectedFormats, tikaAddress, singleFile, ocrOnly); err != nil {
		return C.CString(err.Error())
	}
	return nil
}

//export ScreeningPython
func ScreeningPython(input *C.char) *C.char {
	defer handlePanic()
	if err := runScreening(input); err != nil {
		return C.CString(err.Error())
	}
	return nil
}

//export ValidateConfigPython
func ValidateConfigPython(configType, input *C.char) *C.char {
	defer handlePanic()
	if err := runValidate(configType, input); err != nil {
		return C.CString(err.Error())
	}
	return nil
}

//export CheckConformancePython
func CheckConformancePython(record, protocol *C.char) *C.char {
	defer handlePanic()
	return C.CString(runCheckConformance(record, protocol))
}

//export ProtocolGuidancePython
func ProtocolGuidancePython(protocol *C.char) *C.char {
	defer handlePanic()
	return C.CString(runProtocolGuidance(protocol))
}

// R-specific exports
//
//export RunReviewR
func RunReviewR(input *C.char) *C.char {
	defer handlePanic()
	if err := runReview(input); err != nil {
		return C.CString(err.Error())
	}
	return C.CString("Review completed successfully")
}

//export DownloadZoteroR
func DownloadZoteroR(input *C.char) *C.char {
	defer handlePanic()
	if err := runDownloadZotero(input); err != nil {
		return C.CString(err.Error())
	}
	return C.CString("Download completed successfully")
}

//export DownloadURLListR
func DownloadURLListR(path *C.char) *C.char {
	defer handlePanic()
	if err := runDownloadURLList(path); err != nil {
		return C.CString(err.Error())
	}
	return C.CString("URL list download completed")
}

//export ConvertR
func ConvertR(inputDir, selectedFormats, tikaAddress, singleFile, ocrOnly *C.char) *C.char {
	defer handlePanic()
	if err := runConvert(inputDir, selectedFormats, tikaAddress, singleFile, ocrOnly); err != nil {
		return C.CString(err.Error())
	}
	return C.CString("Conversion completed successfully")
}

//export ScreeningR
func ScreeningR(input *C.char) *C.char {
	defer handlePanic()
	if err := runScreening(input); err != nil {
		return C.CString(err.Error())
	}
	return C.CString("Screening completed successfully")
}

//export ValidateConfigR
func ValidateConfigR(configType, input *C.char) *C.char {
	defer handlePanic()
	if err := runValidate(configType, input); err != nil {
		return C.CString(err.Error())
	}
	return C.CString("Configuration is valid")
}

//export CheckConformanceR
func CheckConformanceR(record, protocol *C.char) *C.char {
	defer handlePanic()
	return C.CString(runCheckConformance(record, protocol))
}

//export ProtocolGuidanceR
func ProtocolGuidanceR(protocol *C.char) *C.char {
	defer handlePanic()
	return C.CString(runProtocolGuidance(protocol))
}

// Free memory function used by both interfaces
//
//export FreeCString
func FreeCString(str *C.char) {
	C.free(unsafe.Pointer(str))
}

func main() {}
