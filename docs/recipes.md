---
title: Recipes
layout: default
---

# Recipes

---

Short, task-oriented guides. Each step links to the page with the full options.

## Run a review end to end

1. [Install prismAId](installation/setup-overview), or connect the [MCP server](mcp-server) and let an agent do the rest.
2. Acquire papers with the [Download tool](tools/download-tool), then turn them into text with the [Convert tool](tools/convert-tool).
3. Create a configuration with the [Review Configurator](review/review-configurator) or the `-init` command, then run the [Review tool](tools/review-tool).
4. Optionally record the run in a [RevAIse](review/revaise-integration) file and [check it for conformance](conformance).

## Screen a manuscript list

1. Prepare your records (for example a Zotero export or a CSV).
2. Configure the [Screening tool](tools/screening-tool) with the filters you need: [deduplication](filters/deduplication), [language](filters/language), [article type](filters/article-type), [topic relevance](filters/topic-relevance).
3. Run screening to tag items for inclusion or exclusion.

## Check a review for PRISMA conformance

1. See what the protocol requires up front with [Protocol Guidance](guidance).
2. Maintain your review as a [RevAIse record](review/revaise-integration) across stages.
3. [Check conformance](conformance): `prismaid -conformance review.revaise.json -protocol prisma-2020`, then fix the reported gaps and re-check.

<div id="wcb" class="carbonbadge"></div>
<script src="https://unpkg.com/website-carbon-badges@1.1.3/b.min.js" defer></script>
