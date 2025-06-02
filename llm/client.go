package llm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"time"

	log "github.com/charmbracelet/log"
	"github.com/regiellis/chronos-go/chronos"
	"github.com/regiellis/chronos-go/config"
)

// OllamaClient implements Parser using a local LLM (7B+ model via Ollama or llama.cpp)
type OllamaClient struct {
	Model string // e.g., "llama2:7b"
	Host  string // e.g., "http://localhost:11434"
}

// NewOllamaClient loads config from .env or uses defaults
func NewOllamaClient() *OllamaClient {
	cfg, _ := config.LoadEnvConfig()
	return &OllamaClient{
		Model: cfg.OllamaModel,
		Host:  cfg.OllamaHost,
	}
}

// llmPrompt is the prompt template for parsing time entries.
const llmPrompt = `Parse the following time tracking entry and return a JSON object with these fields: project, client, task, description, duration (in minutes), entry_time (RFC3339).\nInput: "%s"`

// ParseEntry uses the LLM to parse a natural language time entry string.
func (c *OllamaClient) ParseEntry(input string) (*chronos.Entry, error) {
	prompt := fmt.Sprintf(llmPrompt, input)
	cmd := exec.Command("ollama", "run", c.Model, prompt)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	// Only log errors below
	err := cmd.Run()
	if err != nil {
		log.Error("[LLM] ParseEntry failed: %s", out.String())
		return nil, errors.New("LLM parsing failed: " + out.String())
	}
	// Parse JSON output
	var resp struct {
		Project     string `json:"project"`
		Client      string `json:"client"`
		Task        string `json:"task"`
		Description string `json:"description"`
		Duration    int64  `json:"duration"`
		EntryTime   string `json:"entry_time"`
	}
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		return nil, errors.New("Failed to parse LLM output: " + err.Error() + ": " + out.String())
	}
	entryTime, err := time.Parse(time.RFC3339, resp.EntryTime)
	if err != nil {
		entryTime = time.Now()
	}
	return &chronos.Entry{
		Project:     resp.Project,
		Client:      resp.Client,
		Task:        resp.Task,
		Description: resp.Description,
		Duration:    resp.Duration,
		EntryTime:   entryTime,
		CreatedAt:   time.Now(),
	}, nil
}

// SummarizeBlock uses the LLM to generate a summary for a block and its entries.
func (c *OllamaClient) SummarizeBlock(block *chronos.Block, entries []*chronos.Entry) (string, error) {
	// Compose a prompt for the LLM
	prompt := "Summarize the following work block for a client. Block: " + block.Name + "\n"
	for _, e := range entries {
		prompt += "- " + e.Project + ": " + e.Description + " (" + fmt.Sprintf("%d min", e.Duration) + ")\n"
	}
	cmd := exec.Command("ollama", "run", c.Model, prompt)
	out, err := cmd.Output()
	if err != nil {
		log.Error("[LLM] SummarizeBlock failed: %v", err)
		return "", err
	}
	return string(out), nil
}

// FeedbackAfterEntry uses the LLM to provide feedback after a new entry is added.
func (c *OllamaClient) FeedbackAfterEntry(entry *chronos.Entry, entries []*chronos.Entry) (string, error) {
	// Compose a prompt for the LLM
	totalMinutes := int64(0)
	for _, e := range entries {
		totalMinutes += e.Duration
	}
	prompt := "You are a smart time tracker assistant. The user just logged a new entry: '" + entry.Description + "' (" + entry.Project + ", " + fmt.Sprintf("%d min", entry.Duration) + ").\n"
	prompt += fmt.Sprintf("Total time logged in this block: %.2f hours. Block ends: %s. Give a concise, friendly feedback message.", float64(totalMinutes)/60.0, entry.EntryTime.Format("2006-01-02"))
	cmd := exec.Command("ollama", "run", c.Model, prompt)
	out, err := cmd.Output()
	if err != nil {
		log.Error("[LLM] FeedbackAfterEntry failed: %v", err)
		return "", err
	}
	return string(out), nil
}

// EnhancedFeedback provides richer feedback after an entry, including progress and warnings.
func (c *OllamaClient) EnhancedFeedback(entry *chronos.Entry, entries []*chronos.Entry, block *chronos.Block) (string, error) {
	totalMinutes := int64(0)
	for _, e := range entries {
		totalMinutes += e.Duration
	}
	progress := 0.0
	blockEnd := ""
	if block != nil {
		blockDuration := block.EndTime.Sub(block.StartTime).Hours()
		if blockDuration > 0 {
			progress = float64(totalMinutes) / 60.0 / blockDuration * 100
		}
		blockEnd = block.EndTime.Format("2006-01-02")
	}
	prompt := "You are a time tracking assistant. The user just logged a new entry: '" + entry.Description + "' (" + entry.Project + ", " + fmt.Sprintf("%d min", entry.Duration) + ").\n"
	prompt += fmt.Sprintf("Total time logged in this block: %.2f hours. Progress: %.1f%%. Block ends: %s. Warn if over/under target. Suggest balancing if needed.", float64(totalMinutes)/60.0, progress, blockEnd)
	cmd := exec.Command("ollama", "run", c.Model, prompt)
	out, err := cmd.Output()
	if err != nil {
		log.Error("[LLM] EnhancedFeedback failed: %v", err)
		return "", err
	}
	return string(out), nil
}

// AnswerUserQuery uses the LLM to answer a user question about their time tracking data.
func (c *OllamaClient) AnswerUserQuery(question string, entries []*chronos.Entry, blocks []*chronos.Block) (string, error) {
	prompt := "You are a smart time tracker assistant. The user asked: '" + question + "'.\n"
	prompt += "Here are the user's blocks and entries in JSON:\n"
	entriesJson, _ := json.Marshal(entries)
	blocksJson, _ := json.Marshal(blocks)
	prompt += string(blocksJson) + "\n" + string(entriesJson) + "\n"
	prompt += "Answer concisely and helpfully."
	cmd := exec.Command("ollama", "run", c.Model, prompt)
	out, err := cmd.Output()
	if err != nil {
		log.Error("[LLM] AnswerUserQuery failed: %v", err)
		return "", err
	}
	return string(out), nil
}

// SuggestNextEntry uses the LLM to suggest the next likely entry/task for the user.
func (c *OllamaClient) SuggestNextEntry(entries []*chronos.Entry, blocks []*chronos.Block) (string, error) {
	prompt := "Based on the user's recent time entries and active block, suggest the next likely task or entry. Be concise.\n"
	entriesJson, _ := json.Marshal(entries)
	blocksJson, _ := json.Marshal(blocks)
	prompt += string(blocksJson) + "\n" + string(entriesJson)
	cmd := exec.Command("ollama", "run", c.Model, prompt)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// SmartReminder uses the LLM to generate reminders or nudges based on user activity.
func (c *OllamaClient) SmartReminder(entries []*chronos.Entry, blocks []*chronos.Block) (string, error) {
	prompt := "You are a time tracking assistant. Based on the user's recent entries and blocks, suggest a smart reminder or nudge (e.g., log time, resume a block, review a sprint, etc). Be concise.\n"
	entriesJson, _ := json.Marshal(entries)
	blocksJson, _ := json.Marshal(blocks)
	prompt += string(blocksJson) + "\n" + string(entriesJson)
	cmd := exec.Command("ollama", "run", c.Model, prompt)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// AutoCompleteFields uses the LLM to suggest completions for project, client, or task fields.
func (c *OllamaClient) AutoCompleteFields(partial string, entries []*chronos.Entry, blocks []*chronos.Block) (string, error) {
	prompt := "Suggest auto-completions for this partial input (project/client/task): '" + partial + "'.\n"
	entriesJson, _ := json.Marshal(entries)
	blocksJson, _ := json.Marshal(blocks)
	prompt += string(blocksJson) + "\n" + string(entriesJson)
	cmd := exec.Command("ollama", "run", c.Model, prompt)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
