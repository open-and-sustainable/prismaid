---
title: MCP Server
layout: default
---

# MCP Server

---

The optional MCP server exposes prismAId's tools to agents over `stdio`, so an assistant can help a user design, validate, run, and check a systematic review through the same primitives the library and CLI use.

The server supports three usage patterns:

- local installation from the Go source
- container-based execution from the GHCR image
- registry-based use through the MCP Registry

## Available tools

The tools fall into four groups.

**Design and setup** — offline, no API keys, safe on drafts:

- `prismaid_validate_config` — validate a `review`, `screening`, or `zotero` configuration.
- `prismaid_generate_review_config` — build a review-tool TOML from structured parameters.
- `prismaid_generate_screening_config` — build a screening-tool TOML from structured parameters.
- `prismaid_generate_zotero_config` — build a Zotero-download TOML from structured parameters.

**Protocol conformance** — symbolic check against the latest shapes RevAIse publishes; needs network, but no API keys or file access (see [Protocol Conformance](conformance)):

- `prismaid_check_conformance` — check a [RevAIse record](review/revaise-integration) against a protocol's SHACL shapes.
- `prismaid_list_protocols` — list the accepted protocol identifiers.
- `prismaid_protocol_guidance` — return a protocol's full requirement checklist, grouped by record class, so a user can plan a conforming review before running anything (see [Protocol Guidance](guidance)). Advisory; it does not constrain the order in which tools are used.

**RevAIse records** — build and check a review record by hand (see [RevAIse Integration](review/revaise-integration)); no API keys or file access, and the last two fetch the released data model live (network):

- `prismaid_generate_revaise_record` — seed a review record with a valid header and optional stubs for the stages prismAId does not run (registration, search, risk of bias, synthesis).
- `prismaid_merge_record_stage` — merge a stage into an existing record, or fill a seeded stub.
- `prismaid_revaise_schema` — describe the released RevAIse data model (classes, enums, required slots, enum values), fetched live — never the LinkML source.
- `prismaid_validate_record` — validate a record against the released data-model JSON Schema (structural validity, distinct from conformance).

**Execution** — read and write files, use the network, and read LLM API keys from the environment:

- `prismaid_review` — run a systematic review.
- `prismaid_screening` — screen manuscripts.
- `prismaid_convert` — convert PDF/DOCX/HTML files to plain text.
- `prismaid_download_zotero` — download attachments from a Zotero collection.
- `prismaid_download_url_list` — download files from a list of URLs.

Agents discover tool schemas via `tools/list` and call them with `tools/call`. The generator tools accept the same structured parameters as prismAId's Go configuration generators, so an agent can author a configuration field by field, validate it, and then run it — all in one session.

## Running execution tools

The design, generator, and conformance tools are self-contained. The execution tools are not: they read and write files and call LLM providers. Two conventions apply when running the server in a container:

- **File paths are resolved inside the server's own filesystem.** A configuration's `input_directory`, `results_file_name`, and similar paths must refer to paths the server can see. When running the container, bind-mount the working directory (for example `-v "$PWD":/work`) and use the mounted paths in the configuration.
- **API keys come from the environment.** Provider keys placed in the configuration are used as-is; otherwise the standard provider environment variables must be passed into the server process (for example `-e OPENAI_API_KEY`).
- **Host services are not reachable by default.** The container is network-isolated from the host, so a self-hosted LLM endpoint on the host (for example Ollama on `127.0.0.1:11434`) cannot be reached from the containerized server. To reach it, run the container on the host network — `docker run --rm -i --network host -v "$PWD":/work ghcr.io/open-and-sustainable/prismaid-mcp:<version>` — or point the configuration at an endpoint the container can resolve.

Because of these constraints, a practical pattern is to **use the MCP server to design, validate, and check** (the offline design, RevAIse-record, and conformance tools need none of this setup) and to **run the execution tools with the native binary or a language package on the host**, where local files and services are directly available. Configuration authored through the MCP server runs unchanged there.

## Use from Go source

Use this when you want a local binary built directly from the project source (see also the [Go Package](installation/go) installation page).

```sh
go install github.com/open-and-sustainable/prismaid/cmd/prismaid-mcp@latest
prismaid-mcp
```

## Use from the GHCR container image

Use this when you want to run the MCP server without a local Go toolchain.

```sh
docker pull ghcr.io/open-and-sustainable/prismaid-mcp:0.15.0
docker run --rm -i \
  -v "$PWD":/work -e OPENAI_API_KEY \
  ghcr.io/open-and-sustainable/prismaid-mcp:0.15.0
```

Replace `0.15.0` with the released version you want to run. The bind-mount and environment flags are only needed for the execution tools.

## Use from the MCP Registry

Use this when your agent platform supports MCP Registry server discovery and installation.

The prismAId MCP server is published by GitHub Actions on pushed version tags such as `v0.15.0`, using GitHub OIDC authentication and the registry publisher CLI. The registry entry points to the published OCI package, so agents resolve a versioned package rather than a repository source tree.

Registry references:

- Discovery page: `https://registry.modelcontextprotocol.io/?q=prismaid`
- Server ID: `io.github.open-and-sustainable/prismaid-mcp`

## Example requests

Example `tools/list` request:

```json
{ "jsonrpc": "2.0", "id": 1, "method": "tools/list" }
```

Example `tools/call` request (validate a review configuration):

```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/call",
  "params": {
    "name": "prismaid_validate_config",
    "arguments": {
      "config_type": "review",
      "toml": "[project]\nname = \"demo\"\n..."
    }
  }
}
```

In all cases, clients interact with the server through standard MCP `tools/list` and `tools/call` requests.

<div id="wcb" class="carbonbadge"></div>
<script src="https://unpkg.com/website-carbon-badges@1.1.3/b.min.js" defer></script>
