---
title: Protocol Conformance
layout: default
---

# Protocol Conformance

---

prismAId can check whether a systematic review conforms to a reporting protocol such as PRISMA 2020. The check is **declarative**: the protocol's requirements are expressed as data, and conformance is decided by a machine-checkable engine rather than asserted by the language model.

## A declarative, neurosymbolic approach

Conformance sits on top of prismAId's existing open-science layer and combines two kinds of component:

- **Neural** — the language models do the open-ended work: screening manuscripts and extracting structured data. Their output is written into a structured [RevAIse](review/revaise-integration) review record.
- **Symbolic** — each protocol is published by the RevAIse model as machine-readable **SHACL shapes**. A validator checks the record against those shapes and returns a verdict.
- **The bridge** — the RevAIse record is the shared artifact: the neural side fills it in, the symbolic side checks it.

Two principles make this more than a convenience wrapper, and they are worth stating explicitly:

1. **Requirements live in the protocol, not in the software.** The rules a review must satisfy are the SHACL shapes, which are data. A single engine serves any protocol, and adopting a new or revised protocol requires no change to prismAId.
2. **Conformance is a symbolic verdict, not the model's opinion.** Whether a review meets a standard is decided by validating the record against the shapes. The claim is machine-checked and reproducible, not produced by the model that also did the review.

This is what lets AI contributions be verified against versioned, declarative standards — the payoff for open science.

## How it works

The mechanism is a short, deterministic pipeline:

1. The RevAIse review record (JSON) is framed as an RDF graph using the RevAIse vocabulary, so its fields and object types become the terms the shapes refer to.
2. The graph is validated against the selected protocol's SHACL shapes.
3. The result is a report: whether the record **conforms**, plus a list of **unmet constraints**, each carrying the protocol's own message — for PRISMA 2020, mapped to the numbered checklist items (for example *"PRISMA 2020 (25): funding sources must be declared."*).

prismAId records the stages it actually performs — search, screening, and extraction. The other stages a protocol requires — registration, risk-of-bias assessment, synthesis, reporting — are documented by the reviewer (guided by an agent) in the same RevAIse record. Because conformance is evaluated over the whole record, the report tells you what is still missing across the **entire** protocol, not just the automated part.

## Protocols

Protocols are selected by name. **PRISMA 2020** is included. The design is pluggable: as the RevAIse model publishes shapes for additional protocols, they become available by name, with no change to prismAId's code.

## Usage

Provide the RevAIse review record and a protocol name.

**Command line:**

```bash
./prismaid -conformance review.revaise.json -protocol prisma-2020
```

It prints the verdict and any unmet constraints, and exits with a non-zero status when the record does not conform, so it can be used in scripts.

**Go:**

```go
report, err := prismaid.CheckConformance(recordJSON, "prisma-2020")
// report.Conforms, report.Violations (each with a Message)
```

**Python:**

```python
report = prismaid.check_conformance(record_json, "prisma-2020")
# report["conforms"], report["violations"]
```

**R** (returns the report as a JSON string; parse with `jsonlite::fromJSON`):

```r
report <- CheckConformance(record_json, "prisma-2020")
```

**Julia** (returns the report as a JSON string; parse with `JSON.jl`):

```julia
report = PrismAId.check_conformance(record_json, "prisma-2020")
```

## The report

Every channel returns the same information:

- `conforms` — whether the record satisfies the protocol.
- `violations` — the unmet constraints; each carries the protocol's message (the checklist item), the focus node, and the property path.

The verdict and the messages come entirely from the protocol's shapes, so the report reflects the standard itself, not prismAId's interpretation of it.

See [RevAIse Integration](review/revaise-integration) for how the review record is produced and maintained across stages.

<div id="wcb" class="carbonbadge"></div>
<script src="https://unpkg.com/website-carbon-badges@1.1.3/b.min.js" defer></script>
