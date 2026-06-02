package app

import (
	"time"

	"github.com/smasonuk/falken-vector/pkg/falkenvector"
)

type Screen string

const (
	ScreenStatus  Screen = "status"
	ScreenIndex   Screen = "index"
	ScreenSearch  Screen = "search"
	ScreenAsk     Screen = "ask"
	ScreenCompact Screen = "compact"
)

type OperationKind string

const (
	OperationNone    OperationKind = ""
	OperationStatus  OperationKind = "status"
	OperationIngest  OperationKind = "ingest"
	OperationSearch  OperationKind = "search"
	OperationAsk     OperationKind = "ask"
	OperationCompact OperationKind = "compact"
)

type AppModel struct {
	Directory    string
	ActiveScreen Screen

	Status  StatusViewModel
	Index   IndexViewModel
	Search  SearchViewModel
	Ask     AskViewModel
	Compact CompactViewModel

	Events    []EventLine
	Busy      bool
	BusyKind  OperationKind
	LastError string
}

type StatusViewModel struct {
	Loaded        bool
	NoIndex       bool
	Directory     string
	Documents     falkenvector.StatusDocumentCounts
	Chunks        falkenvector.StatusChunkCounts
	LastIndexedAt *time.Time
	Text          string
}

type IndexViewModel struct {
	ExtensionsInput        string
	ExcludeExtensionsInput string
	ExcludeDirsInput       string
	Chunker                falkenvector.ChunkerMode
	ChunkSizeInput         string
	ChunkOverlapInput      string
	SyncSource             bool
	InlineError            string
	LastResult             *falkenvector.IngestResult
	LastDryRunComplete     bool
}

type SearchViewModel struct {
	Question    string
	Mode        falkenvector.RetrievalMode
	TopKInput   string
	Results     []SearchResultView
	Selected    int
	QueryPlan   string
	InlineError string
}

type SearchResultView struct {
	Title   string
	Preview string
	Chunk   falkenvector.RetrievedChunk
}

type AskViewModel struct {
	Question          string
	Mode              falkenvector.RetrievalMode
	TopKInput         string
	Agent             bool
	Answer            string
	Sources           []string
	SourceItems       []AskSourceView
	SelectedSource    int
	SourceDialogOpen  bool
	SourceDialogTitle string
	SourceDialogText  string
	SourceDialogError string
	Warnings          []string
	ToolCalls         []string
	InlineError       string
}

type AskSourceView struct {
	Label       string
	Path        string
	DisplayPath string
	StartLine   int
	EndLine     int
}

type CompactViewModel struct {
	KeepBackup     bool
	BatchSizeInput string
	Confirming     bool
	DryRunComplete bool
	LastResult     *falkenvector.CompactResult
	InlineError    string
}

type EventLine struct {
	At   time.Time
	Text string
}

func NewModel(directory string) AppModel {
	return AppModel{
		Directory:    directory,
		ActiveScreen: ScreenAsk,
		Status: StatusViewModel{
			Directory: directory,
		},
		Index: IndexViewModel{
			Chunker:           falkenvector.ChunkerAuto,
			ChunkSizeInput:    "1200",
			ChunkOverlapInput: "200",
		},
		Search: SearchViewModel{
			Mode:      falkenvector.RetrievalHybrid,
			TopKInput: "8",
		},
		Ask: AskViewModel{
			Mode:      falkenvector.RetrievalHybrid,
			TopKInput: "8",
			Agent:     true,
		},
		Compact: CompactViewModel{
			BatchSizeInput: "100",
		},
	}
}

func WithEventLine(model AppModel, line EventLine) AppModel {
	if line.Text == "" {
		return model
	}
	model.Events = append(model.Events, line)
	const maxEvents = 200
	if len(model.Events) > maxEvents {
		model.Events = append([]EventLine(nil), model.Events[len(model.Events)-maxEvents:]...)
	}
	return model
}
