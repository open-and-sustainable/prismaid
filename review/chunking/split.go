package chunking

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/open-and-sustainable/prismaid/review/config"
	"github.com/pkoukk/tiktoken-go"
)

// Counter estimates the size of a generated prompt against a user-provided
// input-context limit. Method identifies whether the count is model-specific
// or the conservative text-based fallback.
type Counter interface {
	Count(string) int
	Method() string
}

type byteCounter struct{}

func (byteCounter) Count(text string) int { return len([]byte(text)) }
func (byteCounter) Method() string        { return "conservative UTF-8 byte estimate" }

type tiktokenCounter struct {
	encoding *tiktoken.Tiktoken
}

func (c tiktokenCounter) Count(text string) int { return len(c.encoding.EncodeOrdinary(text)) }
func (c tiktokenCounter) Method() string        { return "OpenAI tiktoken" }

// CounterForModels returns an OpenAI tokenizer only for a single configured
// OpenAI model whose encoding can be loaded. Ensembles and all other
// configurations use the conservative UTF-8 byte estimate so one model's
// encoding is never applied to another model.
func CounterForModels(models map[string]config.LLMItem) Counter {
	if len(models) != 1 {
		return byteCounter{}
	}

	keys := make([]string, 0, len(models))
	for key := range models {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var encoding *tiktoken.Tiktoken
	for _, key := range keys {
		model := models[key]
		if model.Provider != "OpenAI" {
			return byteCounter{}
		}
		current, err := tiktoken.EncodingForModel(model.Model)
		if err != nil {
			return byteCounter{}
		}
		if encoding == nil {
			encoding = current
		}
	}
	if encoding == nil {
		return byteCounter{}
	}
	return tiktokenCounter{encoding: encoding}
}

// Chunk is one contiguous prompt-text fragment. Text can include an overlap
// copied from the preceding chunk; CoreText identifies only newly covered text.
// EstimatedTokens includes the fixed prompt Prefix used by Split.
type Chunk struct {
	Text            string
	CoreText        string
	EstimatedTokens int
}

var paragraphBoundary = regexp.MustCompile(`\r?\n[\t ]*\r?\n`)
var sentenceBoundary = regexp.MustCompile(`(?s).*?[.!?](?:\s+|$)|.+$`)

// Split divides text into roughly equal, bounded chunks. It preferentially
// ends chunks at paragraph boundaries, then sentence boundaries, and finally
// at a rune boundary. Prefix must be the fixed portion of every generated
// prompt and is included in each capacity check. The returned CoreText values,
// concatenated in order, reproduce the complete source text exactly.
func Split(text, prefix string, inputContextTokens, overlapTokens int, counter Counter) ([]Chunk, error) {
	if inputContextTokens <= 0 {
		return nil, fmt.Errorf("input context tokens must be greater than zero")
	}
	if overlapTokens < 0 {
		return nil, fmt.Errorf("overlap tokens cannot be negative")
	}
	if counter.Count(prefix) >= inputContextTokens {
		return nil, fmt.Errorf("fixed review prompt uses %d tokens, leaving no room within the configured %d-token input context", counter.Count(prefix), inputContextTokens)
	}
	if counter.Count(prefix+text) <= inputContextTokens {
		return []Chunk{{Text: text, CoreText: text, EstimatedTokens: counter.Count(prefix + text)}}, nil
	}

	available := inputContextTokens - counter.Count(prefix)
	if overlapTokens >= available {
		return nil, fmt.Errorf("overlap of %d tokens leaves no room for new text within the configured input context", overlapTokens)
	}

	documentTokens := counter.Count(text)
	chunkCount := 1
	covered := available
	for covered < documentTokens {
		chunkCount++
		covered += available - overlapTokens
	}
	targetCoreTokens := (documentTokens + chunkCount - 1) / chunkCount

	units := paragraphUnits(text)
	chunks := make([]Chunk, 0, chunkCount)
	previousCore := ""

	for len(units) > 0 {
		overlap := tail(previousCore, overlapTokens, counter)
		for overlap != "" && counter.Count(prefix+overlap) >= inputContextTokens {
			overlap = dropFirstRune(overlap)
		}

		core := ""
		for len(units) > 0 {
			candidateCore := core + units[0]
			if counter.Count(prefix+overlap+candidateCore) <= inputContextTokens {
				core = candidateCore
				units = units[1:]
				if counter.Count(core) >= targetCoreTokens && len(units) > 0 {
					break
				}
				continue
			}

			if core != "" {
				break
			}
			sentences := sentenceUnits(units[0])
			if len(sentences) > 1 {
				units = append(sentences, units[1:]...)
				continue
			}

			piece, remainder := largestFittingPrefix(units[0], prefix+overlap, inputContextTokens, counter)
			if piece == "" {
				return nil, fmt.Errorf("unable to fit even one character of a document chunk within the configured input context")
			}
			core = piece
			if remainder == "" {
				units = units[1:]
			} else {
				units[0] = remainder
			}
			break
		}

		if core == "" {
			return nil, fmt.Errorf("unable to build a non-empty document chunk")
		}
		chunkText := overlap + core
		chunks = append(chunks, Chunk{
			Text:            chunkText,
			CoreText:        core,
			EstimatedTokens: counter.Count(prefix + chunkText),
		})
		previousCore = core
	}

	return chunks, nil
}

func paragraphUnits(text string) []string {
	boundaries := paragraphBoundary.FindAllStringIndex(text, -1)
	if len(boundaries) == 0 {
		return sentenceUnits(text)
	}

	units := make([]string, 0, len(boundaries)+1)
	start := 0
	for _, boundary := range boundaries {
		units = append(units, text[start:boundary[1]])
		start = boundary[1]
	}
	if start < len(text) {
		units = append(units, text[start:])
	}
	return withoutEmpty(units)
}

func sentenceUnits(text string) []string {
	return withoutEmpty(sentenceBoundary.FindAllString(text, -1))
}

func withoutEmpty(units []string) []string {
	result := make([]string, 0, len(units))
	for _, unit := range units {
		if unit != "" {
			result = append(result, unit)
		}
	}
	return result
}

func largestFittingPrefix(text, prefix string, limit int, counter Counter) (string, string) {
	runes := []rune(text)
	low, high, best := 1, len(runes), 0
	for low <= high {
		middle := low + (high-low)/2
		candidate := string(runes[:middle])
		if counter.Count(prefix+candidate) <= limit {
			best = middle
			low = middle + 1
		} else {
			high = middle - 1
		}
	}
	if best == 0 {
		return "", text
	}
	return string(runes[:best]), string(runes[best:])
}

func tail(text string, limit int, counter Counter) string {
	if text == "" || limit == 0 {
		return ""
	}
	runes := []rune(text)
	low, high, best := 0, len(runes), len(runes)
	for low <= high {
		middle := low + (high-low)/2
		candidate := string(runes[middle:])
		if counter.Count(candidate) <= limit {
			best = middle
			high = middle - 1
		} else {
			low = middle + 1
		}
	}
	return strings.TrimLeft(string(runes[best:]), " \t\r\n")
}

func dropFirstRune(text string) string {
	runes := []rune(text)
	if len(runes) <= 1 {
		return ""
	}
	return string(runes[1:])
}
