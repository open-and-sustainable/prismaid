// File: pythonlib/export.go
package main

/*
#include <stdlib.h>
*/
import "C"

import (
    "fmt"
    "github.com/open-and-sustainable/prismaid"
    "unsafe"
)

//export RunReviewPython
func RunReviewPython(input *C.char) *C.char {
    // Recover from panic, but do not return from defer
    defer func() {
        if r := recover(); r != nil {
            fmt.Println("Recovered from panic:", r)
            // Directly return the panic message
            return // No need to return a C string here
        }
    }()

    // Convert the C string to a Go string
    goInput := C.GoString(input)

    // Call your Go function
    err := prismaid.RunReview(goInput)
    if err != nil {
        fmt.Println("Error in RunReview:", err)
        return C.CString(err.Error())
    }

    return nil
}

//export FreeCString
func FreeCString(str *C.char) {
    C.free(unsafe.Pointer(str))
}

func main() {}
