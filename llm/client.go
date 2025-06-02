package llm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"time"

	log "github.com/charmbracelet/log"
	"github.com/regiellis/chronos-go/chronos" // Ensured chronos is imported
	"github.com/regiellis/chronos-go/config"
)

// OllamaClient implements Parser using a local LLM (7B+ model via Ollama or llama.cpp)
type OllamaClient struct {
	Model string // e.g., "llama2:7b"
	Host  string // e.g., "http://localhost:11434"
}

// NewOllamaClient loads config from .env or uses defaults
func NewOllamaClient() *OllamaClient {
	cfg, _ := config.LoadEnvConfig() // Assuming LoadEnvConfig handles potential errors gracefully
	return &OllamaClient{
		Model: cfg.OllamaModel,
		Host:  cfg.OllamaHost,
	}
}

// llmParsePrompt is the prompt template for parsing time entries.
// Updated to request fields closer to the new chronos.Entry structure.
const llmParsePrompt = `Parse the following time tracking entry and return a JSON object with these fields: summary (string), project_name (string, optional), start_time (RFC3339, optional, defaults to now if not specified), end_time (RFC3339, optional), duration_minutes (integer, optional, used if end_time not specified).\nInput: "%s"`

// ParseEntry uses the LLM to parse a natural language time entry string.
// Returns the new chronos.Entry structure.
func (c *OllamaClient) ParseEntry(input string) (*chronos.Entry, error) {
	prompt := fmt.Sprintf(llmParsePrompt, input)
	cmd := exec.Command("ollama", "run", c.Model, prompt) // Assumes ollama is in PATH
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	if err != nil {
		log.Error("[LLM] ParseEntry command failed", "error", err, "stdout", outBuf.String(), "stderr", errBuf.String())
		return nil, fmt.Errorf("LLM parsing command failed: %w. Stderr: %s", err, errBuf.String())
	}

	// Attempt to clean up LLM output if it includes non-JSON text around the JSON object.
	// A common pattern is for LLMs to wrap JSON in backticks or provide explanations.
	jsonOutput := outBuf.String()
	firstBrace := bytes.IndexByte(outBuf.Bytes(), '{')
	lastBrace := bytes.LastIndexByte(outBuf.Bytes(), '}')

	if firstBrace != -1 && lastBrace != -1 && lastBrace > firstBrace {
		jsonOutput = jsonOutput[firstBrace : lastBrace+1]
	} else {
		log.Warn("[LLM] ParseEntry output does not look like JSON, attempting direct parse.", "output", outBuf.String())
		// No JSON object found, return error or attempt direct parse if it makes sense.
		// For now, error out if no clear JSON.
		return nil, fmt.Errorf("LLM output did not contain a recognizable JSON object: %s", outBuf.String())
	}


	var resp struct {
		Summary         string `json:"summary"`
		ProjectName     string `json:"project_name"` // LLM returns project name
		StartTime       string `json:"start_time"`   // RFC3339
		EndTime         string `json:"end_time"`     // RFC3339
		DurationMinutes int64  `json:"duration_minutes"`
	}

	if err := json.Unmarshal([]byte(jsonOutput), &resp); err != nil {
		log.Error("[LLM] ParseEntry JSON unmarshal failed", "error", err, "json_output", jsonOutput)
		return nil, fmt.Errorf("failed to parse LLM JSON output: %w. Raw output: %s", err, jsonOutput)
	}

	entry := &chronos.Entry{
		Summary:   resp.Summary,
		ProjectID: 0, // ProjectID needs to be looked up based on resp.ProjectName
		// Store ProjectName temporarily if needed, or handle lookup in cmd/add_cmd.go
		// For now, add_cmd.go expects ParseEntry to return a chronos.Entry,
		// but it has fields like `parsedLLMEntry.Summary`, `parsedLLMEntry.StartTime`, `parsedLLMEntry.Duration`.
		// The fields here are now closer to the target chronos.Entry.
	}

	if resp.StartTime != "" {
		st, errSt := time.Parse(time.RFC3339, resp.StartTime)
		if errSt == nil {
			entry.StartTime = st
		} else {
			log.Warn("[LLM] Could not parse StartTime from LLM, defaulting.", "value", resp.StartTime, "error", errSt)
			entry.StartTime = time.Now() // Default if parsing fails
		}
	} else {
		entry.StartTime = time.Now() // Default if not provided
	}

	if resp.EndTime != "" {
		et, errEt := time.Parse(time.RFC3339, resp.EndTime)
		if errEt == nil {
			entry.EndTime = et
		} else {
			log.Warn("[LLM] Could not parse EndTime from LLM.", "value", resp.EndTime, "error", errEt)
			// Fallback to duration if EndTime parsing fails or not provided
			if resp.DurationMinutes > 0 {
				entry.EndTime = entry.StartTime.Add(time.Duration(resp.DurationMinutes) * time.Minute)
			} else {
				entry.EndTime = entry.StartTime.Add(30 * time.Minute) // Default duration
			}
		}
	} else if resp.DurationMinutes > 0 {
		entry.EndTime = entry.StartTime.Add(time.Duration(resp.DurationMinutes) * time.Minute)
	} else {
		// If neither EndTime nor DurationMinutes is provided, set a default duration (e.g., 0 or 30 min)
		entry.EndTime = entry.StartTime // Or add a default duration like 30 minutes
	}
	
	entry.CreatedAt = time.Now()
	entry.UpdatedAt = time.Now()
	entry.Invoiced = false

	// The caller (cmd/add_cmd.go) will need to handle ProjectID lookup using resp.ProjectName
	// One way is to add a temporary field to chronos.Entry like `parsedLLMEntry.ProjectName`
	// or return ProjectName alongside the entry from this function.
	// For now, add_cmd.go has placeholders for ProjectID. This function now provides a more structured entry.
	// To help `cmd/add_cmd.go` with project name, we can pass it back via a field not saved to DB
	// or by returning multiple values. For simplicity, `add_cmd.go` may need to extract it from summary or expect it.
	// Let's assume `add_cmd.go` will get ProjectName from `parsedLLMEntry.ProjectName` if `ParseEntry` were to return that.
	// The current `add_cmd.go` structure does `parsedLLMEntry, errLLM := llmClient.ParseEntry(input)`
	// and then accesses `parsedLLMEntry.Summary`, `parsedLLMEntry.StartTime`, `parsedLLMEntry.Duration`.
	// This means the `chronos.Entry` returned here must have at least those, or `add_cmd.go` adapts.
	// The `Duration` field is not in the new `chronos.Entry`. We return StartTime and EndTime.
	// `add_cmd.go` will calculate duration if needed for display.
	// We will also add ProjectName to the struct for add_cmd.go to use for lookup.
	// This field `RawProjectName` is for temporary holding of name from LLM.
	// type Entry struct { ... RawProjectName string `json:"-"` ... } // Would require changing chronos.Entry
	// For now, cmd/add_cmd.go will have to make do without project name from here directly.

	return entry, nil
}

// SummarizeBlock uses the LLM to generate a summary for a block and its entries.
// Updated to use new chronos.Entry fields.
func (c *OllamaClient) SummarizeBlock(block *chronos.Block, entries []*chronos.Entry) (string, error) {
	prompt := "Summarize the following work block for a client. Block: " + block.Name + " (Client: " + block.Client + ", Project: " + block.Project + ")\n"
	for _, e := range entries {
		duration := e.EndTime.Sub(e.StartTime).Minutes()
		// Project name would ideally come from fetching Project by e.ProjectID
		prompt += fmt.Sprintf("- Summary: %s (ProjectID: %d, Duration: %.0f min)\n", e.Summary, e.ProjectID, duration)
	}
	cmd := exec.Command("ollama", "run", c.Model, prompt)
	out, err := cmd.Output()
	if err != nil {
		log.Error("[LLM] SummarizeBlock failed", "error", err)
		return "", err
	}
	return string(out), nil
}

// FeedbackAfterEntry uses the LLM to provide feedback after a new entry is added.
// Updated to use new chronos.Entry fields.
func (c *OllamaClient) FeedbackAfterEntry(entry *chronos.Entry, entries []*chronos.Entry) (string, error) {
	totalMinutes := 0.0
	for _, e := range entries { // Iterate over all entries to sum up time for the current context (e.g., current block)
		totalMinutes += e.EndTime.Sub(e.StartTime).Minutes()
	}
	entryDuration := entry.EndTime.Sub(entry.StartTime).Minutes()
	// Project name would ideally come from fetching Project by entry.ProjectID
	prompt := fmt.Sprintf("You are a smart time tracker assistant. The user just logged a new entry: '%s' (ProjectID: %d, Duration: %.0f min).\n", entry.Summary, entry.ProjectID, entryDuration)
	prompt += fmt.Sprintf("Total time logged in this context: %.2f hours. Give a concise, friendly feedback message.", totalMinutes/60.0)

	cmd := exec.Command("ollama", "run", c.Model, prompt)
	out, err := cmd.Output()
	if err != nil {
		log.Error("[LLM] FeedbackAfterEntry failed", "error", err)
		return "", err
	}
	return string(out), nil
}

// EnhancedFeedback provides richer feedback after an entry, including progress and warnings.
// Updated to use new chronos.Entry and chronos.Block fields.
func (c *OllamaClient) EnhancedFeedback(entry *chronos.Entry, entries []*chronos.Entry, block *chronos.Block) (string, error) {
	totalMinutesInBlock := 0.0
	for _, e := range entries { // Assuming `entries` are those belonging to the `block`
		if e.BlockID == block.ID { // Filter for entries in the current block
			totalMinutesInBlock += e.EndTime.Sub(e.StartTime).Minutes()
		}
	}

	progress := 0.0
	blockEndStr := "N/A"
	if block != nil {
		if !block.EndTime.IsZero() && !block.StartTime.IsZero() {
			blockDurationHours := block.EndTime.Sub(block.StartTime).Hours()
			if blockDurationHours > 0 {
				progress = (totalMinutesInBlock / 60.0) / blockDurationHours * 100
			}
			blockEndStr = block.EndTime.Format("2006-01-02")
		} else {
			blockEndStr = "Ongoing or not defined"
		}
	}
	entryDuration := entry.EndTime.Sub(entry.StartTime).Minutes()
	// Project name from block.Project (string) is fine here.
	prompt := fmt.Sprintf("You are a time tracking assistant. The user just logged a new entry: '%s' (Project: %s, Duration: %.0f min).\n", entry.Summary, block.Project, entryDuration)
	prompt += fmt.Sprintf("Total time logged in this block ('%s'): %.2f hours. Progress: %.1f%%. Block ends: %s. Warn if over/under target. Suggest balancing if needed.", block.Name, totalMinutesInBlock/60.0, progress, blockEndStr)

	cmd := exec.Command("ollama", "run", c.Model, prompt)
	out, err := cmd.Output()
	if err != nil {
		log.Error("[LLM] EnhancedFeedback failed", "error", err)
		return "", err
	}
	return string(out), nil
}

// AnswerUserQuery uses the LLM to answer a user question about their time tracking data.
// Already accepts []*chronos.Entry and []*chronos.Block. JSON marshaling will use new struct fields.
func (c *OllamaClient) AnswerUserQuery(question string, entries []*chronos.Entry, blocks []*chronos.Block) (string, error) {
	prompt := "You are a smart time tracker assistant. The user asked: '" + question + "'.\n"
	prompt += "Here are the user's blocks and entries in JSON format (entries have 'id', 'block_id', 'project_id', 'summary', 'start_time', 'end_time', 'invoiced'; blocks have 'id', 'name', 'client', 'project', 'start_time', 'end_time', 'active'):\n"
	
	// Marshal entries and blocks. Handle potential errors.
	entriesJson, errEntries := json.MarshalIndent(entries, "", "  ")
	if errEntries != nil {
		log.Error("[LLM] Failed to marshal entries to JSON", "error", errEntries)
		return "", fmt.Errorf("failed to marshal entries for LLM: %w", errEntries)
	}
	blocksJson, errBlocks := json.MarshalIndent(blocks, "", "  ")
	if errBlocks != nil {
		log.Error("[LLM] Failed to marshal blocks to JSON", "error", errBlocks)
		return "", fmt.Errorf("failed to marshal blocks for LLM: %w", errBlocks)
	}

	prompt += "Blocks:\n" + string(blocksJson) + "\nEntries:\n" + string(entriesJson) + "\n"
	prompt += "Answer concisely and helpfully based *only* on the provided JSON data. Do not make assumptions about ProjectID mappings unless explicitly asked."

	cmd := exec.Command("ollama", "run", c.Model, prompt)
	out, err := cmd.Output()
	if err != nil {
		log.Error("[LLM] AnswerUserQuery failed", "error", err)
		return "", err
	}
	return string(out), nil
}

// SuggestNextEntry uses the LLM to suggest the next likely entry/task for the user.
// Already accepts []*chronos.Entry and []*chronos.Block. JSON marshaling will use new struct fields.
func (c *OllamaClient) SuggestNextEntry(entries []*chronos.Entry, blocks []*chronos.Block) (string, error) {
	prompt := "Based on the user's recent time entries and active block (see JSON data), suggest the next likely task or entry. Be concise.\n"
	entriesJson, _ := json.MarshalIndent(entries, "", "  ")
	blocksJson, _ := json.MarshalIndent(blocks, "", "  ")
	prompt += "Blocks:\n" + string(blocksJson) + "\nEntries:\n" + string(entriesJson)
	cmd := exec.Command("ollama", "run", c.Model, prompt)
	out, err := cmd.Output()
	if err != nil {
		log.Error("[LLM] SuggestNextEntry failed", "error", err)
		return "", err
	}
	return string(out), nil
}

// SmartReminder uses the LLM to generate reminders or nudges based on user activity.
// Already accepts []*chronos.Entry and []*chronos.Block. JSON marshaling will use new struct fields.
func (c *OllamaClient) SmartReminder(entries []*chronos.Entry, blocks []*chronos.Block) (string, error) {
	prompt := "You are a time tracking assistant. Based on the user's recent entries and blocks (see JSON data), suggest a smart reminder or nudge (e.g., log time, resume a block, review a sprint, etc). Be concise.\n"
	entriesJson, _ := json.MarshalIndent(entries, "", "  ")
	blocksJson, _ := json.MarshalIndent(blocks, "", "  ")
	prompt += "Blocks:\n" + string(blocksJson) + "\nEntries:\n" + string(entriesJson)
	cmd := exec.Command("ollama", "run", c.Model, prompt)
	out, err := cmd.Output()
	if err != nil {
		log.Error("[LLM] SmartReminder failed", "error", err)
		return "", err
	}
	return string(out), nil
}

// AutoCompleteFields uses the LLM to suggest completions for project, client, or task fields.
// Already accepts []*chronos.Entry and []*chronos.Block. JSON marshaling will use new struct fields.
func (c *OllamaClient) AutoCompleteFields(partial string, entries []*chronos.Entry, blocks []*chronos.Block) (string, error) {
	prompt := "Suggest auto-completions for this partial input (could be project name, client name, or task summary): '" + partial + "'. Use the provided JSON data for context.\n"
	entriesJson, _ := json.MarshalIndent(entries, "", "  ")
	blocksJson, _ := json.MarshalIndent(blocks, "", "  ")
	prompt += "Blocks:\n" + string(blocksJson) + "\nEntries:\n" + string(entriesJson)
	cmd := exec.Command("ollama", "run", c.Model, prompt)
	out, err := cmd.Output()
	if err != nil {
		log.Error("[LLM] AutoCompleteFields failed", "error", err)
		return "", err
	}
	return string(out), nil
}
