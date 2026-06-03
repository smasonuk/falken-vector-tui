# falken-vector-tui

Note: this tool is under active development. Treat it as alpha-quality software.

`falken-vector-tui` is a terminal app for asking questions about a local
directory. Point it at a folder of source code, notes, Markdown files, or other
text-like content, build an index, then search or ask questions without leaving
your terminal.

It is the interactive UI for the `github.com/smasonuk/falken-vector` library.
The library also provides the `falkengo` CLI. The TUI is meant for the common
daily workflow:

1. Open the tool from the directory you care about.
2. Index that directory.
3. Search the index for matching source chunks.
4. Ask natural-language questions and get answers with cited sources.
5. Check status or compact the local vector database when needed.

## What It Does

Falken Vector is a local RAG system. RAG means "retrieval-augmented generation":
before asking a language model to answer, the tool first retrieves relevant
source material from your own files, then gives that source material to the
model as context.

The important idea is that the model is not magically reading your whole
project. Falken Vector builds a local index, searches that index, and sends only
the selected source chunks to the configured model when you ask for an answer.

Typical uses:

- Find where a behavior, API, decision, or error is described.
- Ask questions about a codebase or documentation folder.
- Summarize a topic across many local notes.
- Inspect the exact files and line ranges behind an answer.
- Keep a project-local search database refreshed as files change.

## How It Works

When you run `Index this directory`, Falken Vector scans the current directory
and indexes text-like files. It skips common generated, dependency, hidden, and
state directories unless you narrow or override discovery with index options.

Indexing has a few steps:

```text
file on disk
  -> text chunks
  -> contextual indexed text
  -> embedding vectors
  -> vector database
  -> SQLite manifest with paths, lines, text, and metadata
```

A chunk is a smaller piece of a file. Large files are split because embedding
models and retrieval work better with smaller passages than with entire
repositories.

An embedding is a list of numbers that represents a piece of text. Text with
similar meaning tends to produce vectors that are close together. A vector
database stores those number lists and can quickly answer: "which chunks are
closest to this question?"

Falken Vector keeps two local stores:

- A vector database for semantic similarity search.
- A SQLite manifest for file paths, chunk text, line ranges, active/deleted
  status, and lexical full-text search.

The vector database is not the source of truth for your text. It stores vectors
plus chunk IDs. The manifest maps those IDs back to the real source text and
file locations.

## Retrieval Modes

Retrieval is the step that turns a question into matching chunks.

Falken Vector supports three retrieval styles:

- `vector`: embeds the question and searches for semantically similar chunks.
  This can find relevant text even when the exact words differ.
- `lexical`: uses SQLite full-text search. This is closer to keyword search and
  is useful for exact names, errors, symbols, and IDs.
- `hybrid`: runs both vector and lexical search, then fuses the rankings.

The TUI's Ask and Search screens currently use hybrid retrieval with `TOPK = 16`
for each request. That means they try to get the best of both worlds: exact
keyword matches plus semantic matches.

## Agent Mode

The Ask screen uses Falken Vector's agentic ask mode.

In a one-shot RAG flow, the program retrieves chunks once, sends them to the
model, and returns the model's answer. Agent mode is more flexible: it starts a
Falken agent with access to search tools such as `search_index`. The agent can
decide what to search for, make more than one search, and use the retrieved
source set to produce an answer.

In practice, agent mode is useful when a question is broad or exploratory, such
as "summarize everything about indexing" or "where is citation validation
implemented?" The agent may try multiple searches, collect evidence from
different files, and then cite the sources it used.

Agent answers should include source citations like `[source 1]`. The source
panel in the TUI lists sources made available to the answer path, including the
files and line ranges behind cited sources. Open a source to preview those lines
directly from disk.

The event log shows agent activity at a safe summary level, including tool calls
and tool results. It does not dump full prompts, retrieved source text, or model
responses into the log by default.

## What Leaves Your Machine

The index and state files are local, but model calls go wherever your
`FALKENGO_*` environment variables point.

During indexing:

- Indexed chunk text is sent to the configured `/embeddings` endpoint.
- The returned vectors are stored locally.

During Ask:

- The question and retrieved source chunks are sent to the configured
  `/chat/completions` endpoint.
- If you enable whole-document attachment for selected Ask sources, the full
  text of those selected documents is sent to the chat endpoint instead of
  retrieved chunks.
- Agent mode may perform several retrieval/tool steps before the final answer.

During Search:

- The TUI uses hybrid retrieval. Vector retrieval needs an embedding call for
  the question; lexical retrieval is local.

Use a local OpenAI-compatible provider if your content must stay on your
machine. Do not commit API keys or provider routing headers.

## Requirements

- Go 1.26.2 or newer, matching this module's `go.mod`.
- A terminal that can run an interactive TUI.
- An OpenAI-compatible embeddings endpoint for indexing and hybrid search.
- An OpenAI-compatible chat completions endpoint for Ask.
- The required `FALKENGO_*` environment variables exported in the shell that
  launches the TUI.

## Quick Start

Build the app:

```bash
cd /path/to/falken-vector-tui
make build
```

Start it from the directory you want to index:

```bash
cd /path/to/your/project-or-notes
/path/to/falken-vector-tui/bin/falken-vector-tui
```

The launch directory matters. The TUI treats the current working directory as
the knowledgebase.

For local development from inside this module:

```bash
go run ./cmd/falken-vector-tui
```

## Downloads / Releases

Versioned binaries are published on GitHub Releases when a `v*` tag is pushed.
Download the archive that matches your OS and CPU architecture, then run the
included `falken-vector-tui` executable.

To publish a new release:

```bash
git tag v0.1.0
git push origin v0.1.0
```

## First Session Workflow

1. Export the model environment variables shown below.
2. Start the TUI from your project, docs folder, or notes folder.
3. Press `3` to open Index.
4. Run `Dry run` to see what will be scanned.
5. Run `Index this directory`.
6. Press `4` to Search for exact topics, symbols, or concepts.
7. Press `1` to Ask natural-language questions with cited answers.
8. Use Status and Compact for maintenance.

If the Ask or Search screen says no index exists, go to Index and build one.

## Configuration

Falken Vector expects OpenAI-compatible `/embeddings` and `/chat/completions`
APIs. The TUI reads the same environment variables as the SDK and `falkengo`
CLI.

Example:

```bash
export FALKENGO_EMBEDDING_MODEL=text-embedding-3-small
export FALKENGO_EMBEDDING_MODEL_URL=https://api.example.com/v1
export FALKENGO_EMBEDDING_MODEL_API_KEY=xxxx

export FALKENGO_LLM_MODEL=gpt-compatible-model
export FALKENGO_LLM_BASE_URL=https://api.example.com/v1
export FALKENGO_LLM_API_KEY=xxxx
```

Provider routing headers can be supplied as JSON objects:

```bash
export FALKENGO_EMBEDDING_MODEL_HEADERS='{"Header-Name":"value"}'
export FALKENGO_LLM_HEADERS='{"Header-Name":"value"}'
```

Portkey-style example:

```bash
export FALKENGO_EMBEDDING_MODEL=text-embedding-3-small
export FALKENGO_EMBEDDING_MODEL_URL=https://portkey.COMPANYNAME.com/v1
export FALKENGO_EMBEDDING_MODEL_API_KEY=xxxx
export FALKENGO_EMBEDDING_MODEL_HEADERS='{"X-Portkey-Provider":"@openai-aifoundry-swc-001"}'

export FALKENGO_LLM_MODEL=gpt-5.2
export FALKENGO_LLM_BASE_URL=https://portkey.COMPANYNAME.com/v1
export FALKENGO_LLM_API_KEY=xxxx
export FALKENGO_LLM_HEADERS='{"X-Portkey-Provider":"@openai-aifoundry-swc-001"}'
```

Optional startup flags:

```bash
falken-vector-tui --state-dir <path>
falken-vector-tui --retrieval lexical|vector|hybrid
falken-vector-tui --top-k <n>
```

`--state-dir` controls where Falken Vector stores its local state. Relative
paths are resolved from the launch directory.

State directory selection order:

1. `--state-dir <path>`
2. `FALKENGO_STATE_DIR`
3. `.falkengo` under the launch directory

`--retrieval` and `--top-k` set SDK defaults. The current TUI Ask and Search
screens send explicit hybrid retrieval requests with `TOPK = 16`.

## Navigation

Use the top navigation bar or these keys:

| Key | Action |
| --- | --- |
| `1` | Ask |
| `2` | Status |
| `3` | Index |
| `4` | Search |
| `5` | Compact |
| `Esc` | Return to Ask, close dialogs, or cancel the active operation |
| `Ctrl-C` | Quit when idle, or cancel the active operation |
| `q` | Quit when idle |

The event log at the bottom of the screen shows indexing, embedding, retrieval,
LLM, warning, and error events. Long-running operations can be cancelled with
`Esc` or `Ctrl-C`.

## Screens

### Ask

Ask a natural-language question about the indexed directory. The answer appears
in the main panel. Sources appear on the right when the SDK returns them.

Open a source to preview the returned file lines from disk. If the file changed
after indexing, the source may be stale; re-index to refresh file locations and
line ranges.

Use `Select sources...` to limit an Ask request to a subset of indexed files or
directories. The picker shows a filesystem-shaped tree built from indexed
documents. Selecting a file includes that file. Selecting a directory includes
all indexed files beneath that directory recursively.

If no sources are selected, Ask uses the full indexed corpus. If sources are
selected, normal Ask keeps agentic RAG enabled but restricts retrieval and
agent tools to the selected files/directories. The agent is also told that the
selected sources are the complete accessible corpus for that answer, so if the
answer is not supported there it should say it was not found in the selected
sources.

The source picker also has `Attach selected documents whole`. When enabled, Ask
reads the selected files from disk and sends the complete selected documents as
context instead of letting the agent search the index. Directory selections are
expanded recursively. The dialog shows the selected document count and a rough
token estimate before you send; large selections warn but are not blocked.

Current request behavior:

- Retrieval mode: hybrid
- Top K: 16
- Agent mode: enabled
- Source filtering: optional, from `Select sources...`
- Whole-document attachment: optional, only for selected sources

### Status

Shows whether the current directory has an index, where the state lives, how
many documents and chunks are indexed, how many are deleted or errored, and when
the index last changed.

If no index exists, Status offers quick actions to dry-run or index the
directory.

### Index

Builds or refreshes the index for the current directory. The ingest root is `.`,
so launch the TUI from the directory you want indexed.

Index options:

- `Extensions`: file extensions to include, such as `go,md,txt`. Leave blank
  for SDK defaults.
- `Exclude extensions`: file extensions to skip, such as `tmp,log`.
- `Exclude directories`: directory names to skip, such as `fixtures,dist`.
- `Chunk size`: target characters per chunk. Must be greater than zero.
- `Overlap`: repeated characters between neighboring chunks. Must be at least
  zero and smaller than chunk size.
- `Chunker`: `auto`, `fixed`, `markdown`, `text`, or `code`.
- `Sync deleted files`: mark missing files as deleted in the index when they
  were previously indexed from this source root.

Use `Dry run` first when indexing a large or unfamiliar directory. Use
`Index this directory` to run the actual ingest. The TUI uses embedding
concurrency `4` for indexing.

### Search

Retrieves matching chunks without asking the chat model to write an answer.

Use Search when you want evidence, not prose. The left panel lists scored
results with file and line ranges. The right panel previews the selected chunk
and any query plan information returned by the SDK.

### Compact

Rebuilds the vector database from active manifest chunks.

Compaction is useful after deletions, embedding model changes, interrupted
maintenance, or when you want to remove inactive vector data. It can also
recreate a missing vector database from the manifest.

Compact options:

- `Batch size`: number of chunks processed per embedding batch.
- `Keep backup`: keep a backup of the old vector database while compacting.

Use `Dry run compact` to see what would happen. `Compact now` requires a second
press as confirmation before it changes the database.

## Local State

By default, the TUI stores Falken Vector state inside the directory you launch
it from:

```text
.falkengo/
  manifest.sqlite
  vecgo-data/
  locks/
  tui/preferences.json
```

The manifest tracks documents, chunks, source paths, line ranges, active/deleted
state, lexical search data, and vector references. `vecgo-data` stores the
vector database. `preferences.json` stores TUI settings such as Index and
Compact options.

Draft questions, answers, search results, event logs, and errors are not saved.

## Safety Notes

- The TUI indexes local files, but it does not train the model.
- Indexing sends chunk text to your configured embedding endpoint.
- Ask sends retrieved source chunks to your configured chat endpoint.
- Ask can send complete selected documents when `Attach selected documents
  whole` is enabled.
- Search uses hybrid retrieval, so it may send the question to the embedding
  endpoint.
- The local index can become stale when files move or change; re-index after
  significant edits.
- Do not store real API keys in committed shell scripts or README examples.

## Troubleshooting

- `No index found for this directory`: open Index and run `Index this
  directory`. Confirm you launched the TUI from the intended directory.
- `question is required`: Ask and Search require non-empty input.
- Embedding configuration errors: check `FALKENGO_EMBEDDING_MODEL`,
  `FALKENGO_EMBEDDING_MODEL_URL`, API key, and headers.
- LLM configuration errors: check `FALKENGO_LLM_MODEL`,
  `FALKENGO_LLM_BASE_URL`, API key, and headers.
- Chunk validation errors: make sure chunk size is greater than zero and
  overlap is smaller than chunk size.
- Source preview errors: the source file may have moved, been deleted, or
  changed since indexing. Re-index the directory.
- Whole-document attachment errors: one of the selected files may have moved,
  been deleted, or become unreadable since indexing. Re-index or adjust the
  selected sources.
- Build errors mentioning missing `falkenvector` symbols such as
  `IndexedDocument`, `AttachedDocument`, `SourceScopeNote`, or
  `ListIndexedDocuments` usually mean the TUI is compiling against an older
  `github.com/smasonuk/falken-vector` module than the one in `go.mod`.
  Refresh modules with `go mod download` or `go mod tidy`, and avoid relying on
  a local `go.work` file to hide stale published dependencies.
- Slow indexing: narrow extensions or exclude noisy directories before
  indexing.
- Preference warnings: corrupt preferences are non-fatal. Remove
  `<state-dir>/tui/preferences.json` to return to defaults.

## Relationship To `falkengo`

`falken-vector-tui` wraps the public Falken Vector SDK. The `falkengo` CLI in
the Falken Vector project exposes more flags and advanced workflows, including
retrieval evaluation, repair, reset, and detailed agent debug output.

Use the TUI for interactive indexing, search, asking, status, and compaction.
Use `falkengo` when you need scripting, JSON output, evaluation datasets, or
advanced debug controls that are not exposed in the TUI yet.

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
