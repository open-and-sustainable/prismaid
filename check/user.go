package check

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// RunUserCheck interacts with the user to confirm whether they want to proceed given an estimated cost.
//
// Parameters:
//   - totalCost: A string representation of the estimated cost.
//   - provider: The name of the AI provider (e.g., "GoogleAI", "Anthropic").
//
// Returns:
//   - An error if the user chooses not to proceed or if an input reading error occurs.
//
// Example:
//   > err := RunUserCheck("100", "GoogleAI")
//   > if err != nil {
//   >     log.Println("Operation aborted:", err)
//   > }
func RunUserCheck(totalCost string, provider string) error {
	if  provider == "GoogleAI" {
		fmt.Println("Unless you are using a free tier with Google AI, the total cost (USD - $) to run this review is at least:", totalCost)
	} else if provider == "Anthropic" {
		fmt.Println("Anthropic tokenizer is not available, hence the estimation of the number of token is very imprecise.\nThe total cost (USD - $) to run this review should be at least:", totalCost)
	} else if provider == "DeepSeek" {
		fmt.Println("DeepSeek tokenizer is not available, hence the estimation of the number of token is very imprecise.\nThe total cost (USD - $) to run this review should be at least:", totalCost)
	} else {
		fmt.Println("The total cost (USD - $) to run this review is at least:", totalCost)
	}
	fmt.Println("This value is an estimate of the total cost of input tokens only.")
	fmt.Println("Eventual requests for CoT justifications and summaries increase the cost and are not included here.")
	// Ask the user if they want to continue
	fmt.Print("Do you want to continue? (y/n): ")
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %v", err)
	}

	// Normalize and check response
	response = strings.TrimSpace(strings.ToLower(response))
	if response != "y" {
		return fmt.Errorf("operation aborted by the user")
	}

	return nil // No error, operation continues
}
