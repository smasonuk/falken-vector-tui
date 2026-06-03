# falken-vector-tui

Note: under active development. Alpha quality.

`falken-vector-tui` is a small terminal interface for the public
`github.com/smasonuk/falken-vector` SDK. It lets you index a local directory,
search the resulting vector store, ask questions against the indexed content,
inspect status, and compact the vector database without switching back to the
CLI for the common workflow.

The TUI operates on the current working directory. Start it from the directory
that contains the knowledgebase you want to index and query.

## Requirements

- Go 1.26.2 or newer, matching the module's `go.mod`.
- A terminal that can run an interactive TUI.
- The same `FALKENGO_*` environment variables required by the
  `falken-vector` SDK.
- Embedding and LLM credentials with access to the models configured in the
  environment.

## Quick start

Build the binary from this module:

```bash
cd /path/to/falken-vector-tui
make build
```

Then run the binary from the directory you want to index:

```bash
cd /path/to/your/knowledgebase
/path/to/falken-vector-tui/bin/falken-vector-tui
```

For a local development run from inside this module:

```bash
go run ./cmd/falken-vector-tui
```

## Configuration

Configuration is read from the same `FALKENGO_*` environment variables as the
SDK. Example values (this example shows working against the portkey api):

```bash
export FALKENGO_EMBEDDING_MODEL=text-embedding-3-small
export FALKENGO_EMBEDDING_MODEL_URL=https://portkey.COMPANYNAME.com/v1
export FALKENGO_EMBEDDING_MODEL_API_KEY=xxxx
export FALKENGO_LLM_API_KEY=xxxxxx
export FALKENGO_LLM_BASE_URL=https://portkey.COMPANYNAME.com/v1
export FALKENGO_LLM_MODEL=gpt-5.2
export FALKENGO_EMBEDDING_MODEL_HEADERS='{"X-Portkey-Provider":"@openai-aifoundry-swc-001"}'
export FALKENGO_LLM_HEADERS='{"X-Portkey-Provider":"@openai-aifoundry-swc-001"}'
```

Do not commit real API keys or provider headers. Keep them in your shell,
profile, dotenv tooling, or secret manager.

Optional startup flags:

```bash
falken-vector-tui --state-dir <path>
falken-vector-tui --retrieval lexical|vector|hybrid
falken-vector-tui --top-k <n>
```

`--state-dir` overrides the SDK state directory. Relative paths are resolved
from the directory where the TUI is launched. If no state directory is provided,
the app uses `FALKENGO_STATE_DIR`; if that is also empty it falls back to
`.falkengo` under the current working directory.

`--retrieval` and `--top-k` configure SDK defaults. The current Ask and Search
screens submit hybrid retrieval requests with the app constant `TOPK = 16`.

## Navigation

Use the navigation bar or these keys:

| Key | Action |
| --- | --- |
| `1` | Ask |
| `2` | Status |
| `3` | Index |
| `4` | Search |
| `5` | Compact |
| `Esc` | Return to Ask, or cancel the active operation |
| `Ctrl-C` | Quit when idle, or cancel the active operation |
| `q` | Quit when idle |

The event log at the bottom of the screen shows SDK activity such as ingest,
embedding, retrieval, LLM requests, warnings, and errors. Long-running
operations can be cancelled with `Esc` or `Ctrl-C`.

## Screens

### Ask

Ask a natural-language question against the indexed directory. Answers appear
in the main panel. When the SDK returns sources, they appear in the source panel
with relative file paths and line ranges. Activating a source opens a preview of
the cited lines from disk.

Current request defaults:

- Retrieval mode: `hybrid`
- Top K: `16`
- Agent mode: enabled

### Status

Shows the index status for the current directory, including document counts,
chunk counts, deleted/error counts, and the last indexed time. If no index is
found, this screen offers a dry run and an immediate index action.

### Index

Build or refresh the index for the current directory. The ingest request uses
`.` as its root, so launch the app from the repository, docs folder, or content
directory you want indexed.

Index options:

- `Extensions`: comma-separated file extensions to include, such as
  `go,md,txt`. Leave blank to use the SDK default.
- `Exclude extensions`: comma-separated extensions to skip, such as `tmp,log`.
- `Exclude directories`: comma-separated directory names to skip, such as
  `fixtures,dist`.
- `Chunk size`: number of characters per chunk. Must be greater than zero.
- `Overlap`: overlap between chunks. Must be at least zero and less than the
  chunk size.
- `Chunker`: one of `auto`, `fixed`, `markdown`, `text`, or `code`.
- `Sync deleted files`: remove files from the index when they no longer exist
  in the source directory.

Use `Dry run` to inspect what would be scanned before committing a full index.
Use `Index this directory` to run ingest with embedding concurrency set to `4`.

### Search

Runs retrieval against the indexed directory and displays matching chunks. The
left panel lists scored results with file and line ranges. The right panel shows
the selected chunk preview and any query plan information returned by the SDK.

### Compact

Rebuilds the vector database from active manifest chunks. This is useful after
deletions, embedding model changes, or maintenance work where inactive chunks
should be cleaned up.

Compact options:

- `Batch size`: number of chunks processed per embedding batch.
- `Keep backup`: keep a backup while compacting.

Use `Dry run compact` before changing the database. `Compact now` requires a
second press as confirmation before it runs.

## Saved State

TUI preferences are stored as JSON in:

```text
<state-dir>/tui/preferences.json
```

The state directory is chosen in this order:

1. `--state-dir <path>`
2. `FALKENGO_STATE_DIR`
3. `.falkengo` under the launch directory

Saved preferences currently include Index screen settings and Compact screen
settings. Volatile values such as draft questions, search results, answers,
events, and errors are not persisted.

## Development

Common Make targets:

```bash
make build   # builds bin/falken-vector-tui with CGO_ENABLED=0
make test    # runs go test ./...
make lint    # runs go vet ./...
make clean   # removes the built binary
```

You can override the Go tool or output path:

```bash
GO=go1.26.2 BINARY=/tmp/falken-vector-tui make build
```

Package layout:

- `cmd/falken-vector-tui`: process entry point and argument parsing.
- `internal/config`: startup flags and SDK engine configuration overrides.
- `internal/app`: view models, request builders, preferences, event formatting,
  operation runner, and SDK adapter.
- `internal/ui`: earlgray components for the screen layout and controls.

The TUI uses `github.com/smasonuk/earlgray` for terminal UI rendering and
`github.com/smasonuk/falken-vector/pkg/falkenvector` for indexing, retrieval,
answering, status, and compaction.

## Troubleshooting

- `No index found for this directory`: open Status or Index and run
  `Index this directory`. Confirm that you launched the TUI from the directory
  you intended to index.
- `question is required`: the Ask and Search screens require non-empty input.
- Chunk validation errors: make sure chunk size is greater than zero and overlap
  is smaller than chunk size.
- Environment/configuration errors: check that all required `FALKENGO_*`
  variables are exported in the shell that launches the TUI.
- Source preview errors: a cited file may have moved, been deleted, or changed
  since the last index. Re-index the directory if source locations are stale.
- Preference warnings: corrupt preference JSON is non-fatal. Remove
  `<state-dir>/tui/preferences.json` to return to defaults.
