package init

import (
    "fmt"
    "os"
    "strings"
    "testing"
)

func TestMyInteractiveFunction(t *testing.T) {
    // Create a temporary file
    tmpfile, err := os.CreateTemp("", "test")
    if err != nil {
        t.Fatalf("Failed to create temp file: %v", err)
    }
    defer os.Remove(tmpfile.Name()) // Clean up after the test

    // Write test data to the temp file
    testData := strings.Join([]string{
        "./test_config.toml",
        "Test Project",
        "Test Author",
        "1.0",
        "no",
        "./input",
        "",
        "./results",
        "csv",
        "low",
        "no",
        "no",
        "no",
        "no",
        "yes",
        "yes",
        "yes",
        "yes",
        "no",
    }, "\n")
    if _, err := tmpfile.WriteString(testData); err != nil {
        tmpfile.Close()
        t.Fatalf("Failed to write to temp file: %v", err)
    }
    if _, err := tmpfile.Seek(0, 0); err != nil {
        tmpfile.Close()
        t.Fatalf("Failed to seek temp file: %v", err)
    }

    // Redirect stdin
    originalStdin := os.Stdin
    defer func() {
        os.Stdin = originalStdin
        tmpfile.Close()
    }()
    os.Stdin = tmpfile

    // Call the function that reads from stdin
    output, err := myInteractiveFunction()
    if err != nil {
        t.Errorf("Failed during function execution: %v", err)
    }

    // Assert the expected output or state
    expectedOutput := "./test_config.toml"
    if output != expectedOutput {
        t.Errorf("Expected %s, got %s instead", expectedOutput, output)
    }
}

// myInteractiveFunction is an example function that reads from stdin and returns some output
func myInteractiveFunction() (string, error) {
    var userInput string
    _, err := fmt.Scanln(&userInput)
    if err != nil {
        return "", err
    }
    return userInput, nil
}
