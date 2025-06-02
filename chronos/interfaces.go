package chronos

// EntryStore defines the interface for entry data operations.
// The *Entry types refer to chronos.Entry.
type EntryStore interface {
	CreateEntry(entry *Entry) error
	GetEntryByID(id int64) (*Entry, error)
	UpdateEntry(entry *Entry) error
	DeleteEntry(id int64) error
	ListEntries(filters map[string]interface{}) ([]*Entry, error)
}

// BlockStore defines the interface for block data operations.
// The *Block types refer to chronos.Block.
type BlockStore interface {
	CreateBlock(block *Block) error
	GetBlockByID(id int64) (*Block, error)
	UpdateBlock(block *Block) error
	DeleteBlock(id int64) error
	ListBlocks(filters map[string]interface{}) ([]*Block, error)
	GetActiveBlock() (*Block, error)
	SetActiveBlock(id int64) error
}

// ProjectStore defines the interface for project data operations.
// The *Project types refer to chronos.Project.
type ProjectStore interface {
	CreateProject(project *Project) error
	GetProjectByID(id int64) (*Project, error)
	UpdateProject(project *Project) error
	DeleteProject(id int64) error
	ListProjects(clientID *int64) ([]*Project, error)
}

// ClientStore defines the interface for client data operations.
// The *Client types refer to chronos.Client.
type ClientStore interface {
	CreateClient(client *Client) error
	GetClientByID(id int64) (*Client, error)
	UpdateClient(client *Client) error
	DeleteClient(id int64) error
	ListClients() ([]*Client, error)
}

// TemplateCreator defines methods for creating/managing templates.
type TemplateCreator interface {
	EnsureTemplatesTable() error
	SaveTemplate(name string, entryText string) error
}

// TemplateRetriever defines methods for retrieving templates.
type TemplateRetriever interface {
	GetTemplate(name string) (string, error)
}

// TemplateStore combines template creation and retrieval operations.
type TemplateStore interface {
	TemplateCreator
	TemplateRetriever
}

// Note: The actual SQL execution logic (e.g., using store.DB.Exec)
// will reside in the concrete implementation of these interfaces,
// presumably in the db package (e.g., on db.Store).
// The functions in chronos/entries.go, chronos/block.go, etc.,
// will call these interface methods.
// For example, a chronos.CreateEntry function would be:
// func CreateEntryInChronosPackage(store EntryStore, entry *Entry) error {
//     // ... any pre-processing or validation specific to chronos package ...
//     return store.CreateEntry(entry) // Calls the method on the interface
// }
// However, the current task is to change the existing chronos functions
// (which contain the DB logic) to take interface types. This means their
// internal db.Store.DB.Exec calls will temporarily break.
