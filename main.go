package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type Config struct {
	APIType  string `json:"api_type"` // "openai" or "anthropic"
	APIKey   string `json:"api_key"`
	BaseURL  string `json:"base_url"`
	Model    string `json:"model"`
	Thinking bool   `json:"thinking"` // default false
}

const (
	Version = "1.01"

	SystemPromptTemplate = `You are a command-line assistant. The user's working platform is %s (%s).
Your task is to transform the user's natural language request into a single executable command line.

Rules:
1. Output ONLY the command itself, without any explanation, markdown formatting, or code blocks.
2. If the request cannot be reasonably expressed as a single command, output exactly: echo "Sorry, I can't convert that to a one-liner command"
3. Prefer standard, portable commands that work across different shells.
4. If multiple approaches exist, choose the safest and most commonly available one.
5. Never include destructive commands (rm -rf /, etc.) without explicit confirmation from the user.
6. Use the most appropriate tool for the task (e.g., use 'find' for file searches, 'grep' for text filtering, etc.).`

	SystemPromptRegenerateTemplate = `You are a command-line assistant. The user's working platform is %s (%s).
Your task is to transform the user's natural language request into a single executable command line.

Previous commands generated (DO NOT use these same commands again):
%s

Rules:
1. Output ONLY the command itself, without any explanation, markdown formatting, or code blocks.
2. If the request cannot be reasonably expressed as a single command, output exactly: echo "Sorry, I can't convert that to a one-liner command"
3. Prefer standard, portable commands that work across different shells.
4. If multiple approaches exist, choose the safest and most commonly available one.
5. Never include destructive commands (rm -rf /, etc.) without explicit confirmation from the user.
6. Use the most appropriate tool for the task (e.g., use 'find' for file searches, 'grep' for text filtering, etc.).
7. CRITICAL: Generate a DIFFERENT command from all previously listed commands. Be creative and find an alternative approach.`

	OpenAIModel    = "gpt-4o-mini"
	AnthropicModel = "claude-3-5-haiku-20241022"

	OpenAIBaseURL    = "https://api.openai.com/v1/chat/completions"
	AnthropicBaseURL = "https://api.anthropic.com/v1/messages"
)

var configPath string

func init() {
	usr, err := user.Current()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current user: %v\n", err)
		os.Exit(1)
	}
	configPath = filepath.Join(usr.HomeDir, ".ol_config.json")
}

func main() {
	if len(os.Args) < 2 {
		showHelp()
		os.Exit(1)
	}

	arg := os.Args[1]

	switch arg {
	case "--help", "-h":
		showHelp()
	case "--init":
		if err := runInit(); err != nil {
			fmt.Fprintf(os.Stderr, "Initialization failed: %v\n", err)
			os.Exit(1)
		}
	case "--version", "-v":
		fmt.Printf("ol version %s\n", Version)
	default:
		// Check if config exists
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			fmt.Println("First time setup required.")
			if err := runInit(); err != nil {
				fmt.Fprintf(os.Stderr, "Initialization failed: %v\n", err)
				os.Exit(1)
			}
		}

		config, err := loadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		// Join all arguments as the user request
		userRequest := strings.Join(os.Args[1:], " ")

		// Track command history for regeneration
		var history []string

		for {
			// Start spinner
			stopSpinner := make(chan bool)
			go showSpinner(stopSpinner)

			// Get command from AI
			var command string
			if len(history) == 0 {
				command, err = getCommand(config, userRequest)
			} else {
				command, err = getCommandWithHistory(config, userRequest, history)
			}

			// Stop spinner
			stopSpinner <- true
			<-stopSpinner

			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			// Add to history
			history = append(history, command)

			// Show the command and ask for confirmation
			fmt.Printf("%s\n", command)
			if len(history) > 1 {
				fmt.Printf("(Regeneration %d) ", len(history)-1)
			}
			fmt.Print("Press Enter to execute, 'r' to regenerate, or Ctrl+C to cancel...")

			// Read input
			reader := bufio.NewReader(os.Stdin)
			input, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("\nCancelled.")
				os.Exit(0)
			}

			input = strings.TrimSpace(strings.ToLower(input))

			if input == "" {
				// Enter pressed - execute
				executeCommand(command)
				return
			} else if input == "r" || input == "regen" {
				// Regenerate
				clearLine()
				fmt.Println("Regenerating...")
				continue
			} else {
				// Anything else - cancel
				fmt.Println("Cancelled.")
				os.Exit(0)
			}
		}
	}
}

func showHelp() {
	fmt.Println(`ol - One-liner command assistant

Usage:
  ol <your request>    Convert natural language to shell command
  ol --help            Show this help message
  ol --init            Reinitialize configuration
  ol --version         Show version

Examples:
  ol list all files modified in the last 24 hours
  ol find files larger than 100MB in current directory
  ol show git log with graph visualization

Controls:
  Enter                Execute the generated command
  r                    Regenerate a different command
  Ctrl+C               Cancel

Configuration file: ~/.ol_config.json`)
}

func runInit() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("=== ol Configuration ===")
	fmt.Println()
	fmt.Println("Select your API provider:")
	fmt.Println("1) OpenAI Compatible API")
	fmt.Println("2) Anthropic Compatible API")
	fmt.Print("Enter choice (1 or 2): ")

	choice, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	choice = strings.TrimSpace(choice)

	config := &Config{
		Thinking: false, // default off
	}

	switch choice {
	case "1":
		config.APIType = "openai"
		config.Model = OpenAIModel
		fmt.Println()
		fmt.Print("Enter API Key: ")
		apiKey, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		config.APIKey = strings.TrimSpace(apiKey)

		fmt.Print("Enter Base URL (press Enter for default: https://api.openai.com/v1): ")
		baseURL, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		baseURL = strings.TrimSpace(baseURL)
		if baseURL == "" {
			baseURL = "https://api.openai.com/v1"
		}
		config.BaseURL = baseURL

		fmt.Print("Enter Model (press Enter for default: gpt-4o-mini): ")
		model, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		model = strings.TrimSpace(model)
		if model != "" {
			config.Model = model
		}

	case "2":
		config.APIType = "anthropic"
		config.Model = AnthropicModel
		config.BaseURL = AnthropicBaseURL
		fmt.Println()
		fmt.Print("Enter API Key: ")
		apiKey, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		config.APIKey = strings.TrimSpace(apiKey)

		fmt.Print("Enter Model (press Enter for default: claude-3-5-haiku-20241022): ")
		model, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		model = strings.TrimSpace(model)
		if model != "" {
			config.Model = model
		}

	default:
		return fmt.Errorf("invalid choice")
	}

	return saveConfig(config)
}

func loadConfig() (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func saveConfig(config *Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}

func showSpinner(stop chan bool) {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	i := 0
	for {
		select {
		case <-stop:
			clearLine()
			stop <- true
			return
		default:
			fmt.Printf("\r%s Generating command...", frames[i%len(frames)])
			time.Sleep(80 * time.Millisecond)
			i++
		}
	}
}

func clearLine() {
	if runtime.GOOS == "windows" {
		fmt.Print("\r                                                    \r")
	} else {
		fmt.Print("\r\033[K")
	}
}

func getCommand(config *Config, userRequest string) (string, error) {
	switch config.APIType {
	case "openai":
		return getCommandOpenAI(config, userRequest, nil)
	case "anthropic":
		return getCommandAnthropic(config, userRequest, nil)
	default:
		return "", fmt.Errorf("unknown API type: %s", config.APIType)
	}
}

func getCommandWithHistory(config *Config, userRequest string, history []string) (string, error) {
	switch config.APIType {
	case "openai":
		return getCommandOpenAI(config, userRequest, history)
	case "anthropic":
		return getCommandAnthropic(config, userRequest, history)
	default:
		return "", fmt.Errorf("unknown API type: %s", config.APIType)
	}
}

func getCommandOpenAI(config *Config, userRequest string, history []string) (string, error) {
	var systemPrompt string
	if len(history) > 0 {
		historyStr := ""
		for i, cmd := range history {
			historyStr += fmt.Sprintf("%d. %s\n", i+1, cmd)
		}
		systemPrompt = fmt.Sprintf(SystemPromptRegenerateTemplate, runtime.GOOS, runtime.GOARCH, historyStr)
	} else {
		systemPrompt = fmt.Sprintf(SystemPromptTemplate, runtime.GOOS, runtime.GOARCH)
	}

	payload := map[string]interface{}{
		"model": config.Model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userRequest},
		},
		"temperature": 0.1,
	}

	if config.Thinking {
		payload["temperature"] = 0.7
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	url := config.BaseURL + "/chat/completions"
	if config.BaseURL == "https://api.openai.com/v1" {
		url = "https://api.openai.com/v1/chat/completions"
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.APIKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API error: %s", string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	command := strings.TrimSpace(result.Choices[0].Message.Content)
	command = cleanCommand(command)

	return command, nil
}

func getCommandAnthropic(config *Config, userRequest string, history []string) (string, error) {
	var systemPrompt string
	if len(history) > 0 {
		historyStr := ""
		for i, cmd := range history {
			historyStr += fmt.Sprintf("%d. %s\n", i+1, cmd)
		}
		systemPrompt = fmt.Sprintf(SystemPromptRegenerateTemplate, runtime.GOOS, runtime.GOARCH, historyStr)
	} else {
		systemPrompt = fmt.Sprintf(SystemPromptTemplate, runtime.GOOS, runtime.GOARCH)
	}

	payload := map[string]interface{}{
		"model":      config.Model,
		"max_tokens": 1024,
		"system":     systemPrompt,
		"messages": []map[string]string{
			{"role": "user", "content": userRequest},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", config.BaseURL, bytes.NewReader(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", config.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API error: %s", string(body))
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
			Type string `json:"type"`
		} `json:"content"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if len(result.Content) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	command := strings.TrimSpace(result.Content[0].Text)
	command = cleanCommand(command)

	return command, nil
}

func cleanCommand(command string) string {
	command = strings.TrimPrefix(command, "```bash")
	command = strings.TrimPrefix(command, "```sh")
	command = strings.TrimPrefix(command, "```shell")
	command = strings.TrimPrefix(command, "```")
	command = strings.TrimSuffix(command, "```")
	return strings.TrimSpace(command)
}

func executeCommand(command string) {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd.exe", "/C", command)
	} else {
		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/bash"
		}
		cmd = exec.Command(shell, "-c", command)
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
		os.Exit(1)
	}
}
