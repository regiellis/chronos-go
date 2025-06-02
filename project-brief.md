
# Project Brief: Smart CLI Time Tracker for Freelancers

**Objective:**
Build a modern, display-rich terminal time tracking app for freelancers, with smart features powered by a local LLM (gemma3).
The app should streamline time entry, support project/client organization, and provide actionable feedback and summaries. It
will use Go with Charm’s Huh, and Lipgloss for a rich CLI experience, and SQLite3 for storage.

## Core Requirements

### 1. **User Experience & Workflow**
- **CLI-First:** All interactions occur in the terminal, with a visually rich, keyboard-driven UI (using Huh and Lipgloss).
- **Quick Entry:** Users can log time in a single line, using natural language (e.g., `2h today on UI Design - Form updates`).
- **Block-Based Tracking:**
  - Users can start a “block” (e.g., a 2-week sprint for a client).
  - Time entries default to the active block.
  - If no block is active, prompt the user to select or create one.
- **Project/Client/Task Tagging:** Each entry can be tagged for organization and reporting.

### 2. **Views**
- **List View:** Chronological list of entries, filterable by project, client, or block.
- **Block View:** Shows active blocks, progress, and summaries.
- **Summary View:** Totals and breakdowns by project/client/block.
- **Invoice View:** Invoice-ready summary of billable hours, rates, and totals.
- **JSON Export:** Any view can be exported as JSON for use in other tools.

### 3. **LLM Integration (gemma3)**
- **Natural Language Parsing:**  
  - Parse freeform user input into structured time entries.
- **Entry Suggestions:**  
  - Suggest auto-completions or new entries based on recent activity.
- **Project Summaries:**  
  - Generate concise, human-readable summaries for projects or blocks.
- **Feedback & Guidance:**  
  - After each entry, provide feedback (e.g., “You have 15h logged, 2 days left in this block”).
  - Respond to user queries (e.g., “How much time left in this block?”).

### 4. **Storage**
- **SQLite3:**  
  - Use SQLite3 as the embedded database for all entries, blocks, and metadata.
  - Ensure data can be easily queried and exported.

### 5. **Reporting & Export**
- **JSON Output:**  
  - Enable export of summaries, invoices, and logs in JSON format.
- **Invoice View:**  
  - Present a formatted, ready-to-bill summary for any block or client.

---

## Example User Stories

- As a freelancer, I want to quickly log time using natural language, so I don’t waste time on data entry.
- As a user, I want to organize my work into blocks (e.g., sprints or projects) and see progress at a glance.
- As a user, I want the app to give me feedback and summaries, so I always know how much time I’ve logged and what’s left.
- As a user, I want to export my data as JSON for use in other tools or for invoicing.

---

## Example Commands

```bash
# Start a new block
chronos block start "Client X – May Sprint" --duration 2w

# Add a time entry (defaults to active block)
chronos add 2h today on "UI Design – Form updates"

# View block progress
chronos view block

# Generate invoice view for a block
chronos view invoice --block "Client X – May Sprint"

# Export summary as JSON
chronos export summary --block "Client X – May Sprint" --format json

# Get LLM project summary
chronos summarize --block "Client X – May Sprint"
```

---

## Key Features Table

| Feature                  | Description                                         |
|--------------------------|-----------------------------------------------------|
| Quick Entry & Parsing    | Fast, natural language time logging                 |
| Block-Based Tracking     | Organize work into client-focused, time-bound blocks|
| Smart Views              | List, block, summary, and invoice views             |
| LLM Feedback             | Suggestions, summaries, and contextual responses    |
| Sqlite3 Storage           | Fast, reliable, and queryable local database        |
| JSON Export              | Easy integration with other tools                   |

---

## Technical Stack

- **Language:** Go
- **UI:** Charm Bubble Tea, Lipgloss, Huh
- **Database:** Sqlite3
- **LLM Integration:** Local gemma3 model (e.g., via Ollama or llama.cpp)

---

## Notes

- The LLM is opt-in and provides feedback, suggestions, and summaries.
- The app should remain fast and responsive in the terminal.
- Focus on freelancer needs: simplicity, clarity, and actionable insights.

---
