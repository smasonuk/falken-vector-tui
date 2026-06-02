# falken-vector-tui

A small terminal interface for the public `github.com/smasonuk/falken-vector`
SDK.

```bash
go run ./cmd/falken-vector-tui
```

Configuration is read from the same `FALKENGO_*` environment variables as the
SDK. Optional startup flags:

```bash
--state-dir <path>
--retrieval lexical|vector|hybrid
--top-k <n>
```
