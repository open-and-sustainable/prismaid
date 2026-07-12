---
title: Protocol Guidance
layout: default
---

# Protocol Guidance

---

Before checking a review for conformance, prismAId can tell you what a conforming review will need. **Guidance** returns a reporting protocol's full requirement checklist — the same requirements the [conformance check](conformance) enforces, but listed up front so you can plan for them.

## Proactive and reactive

Conformance works in two directions:

- **Reactive** — [checking a record](conformance) reports the requirements it does not yet meet.
- **Proactive** — guidance answers *what will a conforming review need?* before any record exists.

Guidance returns the checklist extracted from the protocol's published SHACL shapes, grouped by the record class each requirement applies to (for example the review as a whole, the screening stage, or the synthesis). Each item carries the protocol's own message — for PRISMA 2020, the numbered checklist wording. An agent can use it to help plan a conforming review from the outset, then use conformance checking to track what remains.

Like conformance checking, guidance is **declarative**: the requirements come from the shapes RevAIse publishes, pulled at call time (see [Protocol Conformance](conformance) for the shared mechanism). It is **advisory** — it describes what a protocol requires and does not constrain the order in which prismAId's tools are used.

## Usage

Guidance is available on every channel, taking just the protocol name:

- **Command line:** `./prismaid -guidance prisma-2020`
- **Go:** `prismaid.ProtocolGuidance("prisma-2020")` — returns the requirements, each with a `TargetClass` and the protocol's `Message`.
- **Python:** `prismaid.protocol_guidance("prisma-2020")` — returns a dict.
- **R:** `ProtocolGuidance("prisma-2020")` — returns the guidance as a JSON string.
- **Julia:** `PrismAId.protocol_guidance("prisma-2020")` — returns the guidance as a JSON string.
- **MCP:** the `prismaid_protocol_guidance` tool on the [MCP server](mcp-server).

The protocol identifier is the same one accepted by the [conformance check](conformance) — nothing is bundled with prismAId, so the checklist reflects the latest shapes RevAIse publishes. Because those shapes are fetched on demand, guidance requires a network connection.

<div id="wcb" class="carbonbadge"></div>
<script src="https://unpkg.com/website-carbon-badges@1.1.3/b.min.js" defer></script>
