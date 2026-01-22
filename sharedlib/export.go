package main

/*
#include <stdlib.h>
*/
import "C"

import (
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
	return prismaid.Review(goInput)
}

func runDownloadZoteroPDFs(username, apiKey, collectionName, parentDir *C.char) error {
	goUsername := C.GoString(username)
	goApiKey := C.GoString(apiKey)
	goCollectionName := C.GoString(collectionName)
	goParentDir := C.GoString(parentDir)
	return prismaid.DownloadZoteroPDFs(goUsername, goApiKey, goCollectionName, goParentDir)
}

func runDownloadURLList(path *C.char) error {
	goPath := C.GoString(path)
	return prismaid.DownloadURLList(goPath)
}

func runConvert(inputDir, selectedFormats, tikaAddress, singleFile, ocrOnly *C.char) error {
	goInputDir := C.GoString(inputDir)
	goSelectedFormats := C.GoString(selectedFormats)
	goTikaAddress := C.GoString(tikaAddress)
	goSingleFile := C.GoString(singleFile)
	goOcrOnly := strings.TrimSpace(strings.ToLower(C.GoString(ocrOnly)))
	ocrOnlyEnabled := goOcrOnly == "1" || goOcrOnly == "true" || goOcrOnly == "yes"
	return prismaid.Convert(goInputDir, goSelectedFormats, prismaid.ConvertOptions{
		TikaServer: goTikaAddress,
		PDF: prismaid.PDFOptions{
			SingleFile: goSingleFile,
			OCROnly:    ocrOnlyEnabled,
		},
	})
}

func runScreening(input *C.char) error {
	goInput := C.GoString(input)
	return prismaid.Screening(goInput)
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

//export DownloadZoteroPDFsPython
func DownloadZoteroPDFsPython(username, apiKey, collectionName, parentDir *C.char) *C.char {
	defer handlePanic()
	if err := runDownloadZoteroPDFs(username, apiKey, collectionName, parentDir); err != nil {
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

//export DownloadZoteroPDFsR
func DownloadZoteroPDFsR(username, apiKey, collectionName, parentDir *C.char) *C.char {
	defer handlePanic()
	if err := runDownloadZoteroPDFs(username, apiKey, collectionName, parentDir); err != nil {
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

// Free memory function used by both interfaces
//
//export FreeCString
func FreeCString(str *C.char) {
	C.free(unsafe.Pointer(str))
}

func main() {}
