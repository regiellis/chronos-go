package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type UserConfig struct {
	DefaultRate     float64 `json:"default_rate"`
	DefaultBillable bool    `json:"default_billable"`
	Theme           string  `json:"theme"`
}

// EnvConfig holds LLM/Ollama config
type EnvConfig struct {
	OllamaHost  string
	OllamaModel string
}

func LoadConfig(path string) (*UserConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		return &UserConfig{}, nil // fallback to defaults
	}
	defer f.Close()
	var cfg UserConfig
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return &UserConfig{}, nil
	}
	return &cfg, nil
}

func SaveConfig(path string, cfg *UserConfig) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(cfg)
}

// FindEnvPath checks for .env in XDG config, home, or local dir
func FindEnvPath() string {
	home, _ := os.UserHomeDir()
	cfgPaths := []string{
		filepath.Join(home, ".config", "chronos", ".env"),
		".env",
		filepath.Join(filepath.Dir(os.Args[0]), ".env"),
	}
	for _, p := range cfgPaths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return cfgPaths[0] // default location
}

// LoadEnvConfig loads Ollama config from .env
func LoadEnvConfig() (*EnvConfig, error) {
	path := FindEnvPath()
	f, err := os.Open(path)
	if err != nil {
		return &EnvConfig{OllamaHost: "http://localhost:11434", OllamaModel: "llama2:7b"}, nil
	}
	defer f.Close()
	cfg := &EnvConfig{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		var k, v string
		if n, _ := fmt.Sscanf(line, "%[^=]=%s", &k, &v); n == 2 {
			switch k {
			case "OLLAMA_HOST":
				cfg.OllamaHost = v
			case "OLLAMA_MODEL":
				cfg.OllamaModel = v
			}
		}
	}
	if cfg.OllamaHost == "" {
		cfg.OllamaHost = "http://localhost:11434"
	}
	if cfg.OllamaModel == "" {
		cfg.OllamaModel = "llama2:7b"
	}
	return cfg, nil
}

// PromptEnvConfig interactively creates a .env file with user input or sane defaults
func PromptEnvConfig() error {
	path := FindEnvPath()
	fmt.Printf("No .env found. Let's create one at %s\n", path)
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Ollama host [http://localhost:11434]: ")
	host, _ := reader.ReadString('\n')
	host = string([]byte(host)[:len(host)-1])
	if host == "" {
		host = "http://localhost:11434"
	}
	fmt.Print("Ollama model [llama2:7b]: ")
	model, _ := reader.ReadString('\n')
	model = string([]byte(model)[:len(model)-1])
	if model == "" {
		model = "llama2:7b"
	}
	os.MkdirAll(filepath.Dir(path), 0700)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	fmt.Fprintf(f, "OLLAMA_HOST=%s\nOLLAMA_MODEL=%s\n", host, model)
	fmt.Println(".env created!")
	return nil
}
