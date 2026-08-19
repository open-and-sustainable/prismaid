// Package chunking plans and merges explicitly configured review document chunks.
//
// Chunking is used only when project.configuration.chunking.enabled is true and
// a generated review prompt exceeds the user-provided input_context_tokens
// limit. Split keeps source text in order, prefers paragraph and sentence
// boundaries, and can repeat a bounded tail as overlap. Every extraction prompt
// is bound to its source document and chunk index so Merge can restore one
// primary response per source document without relying on response order.
//
// Token counting uses a compatible OpenAI tiktoken encoding only for a single
// supported OpenAI model. Other providers and ensembles use UTF-8 byte length
// as a conservative estimate. The selected method is included in the review
// plan and log so users can assess the resulting chunk plan before execution.
package chunking
