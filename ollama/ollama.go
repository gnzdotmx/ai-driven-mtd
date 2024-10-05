package ollama

import (
	"context"
	"log"

	"github.com/teilomillet/gollm"
)

// AskOllama sends a prompt to Ollama and retrieves the response
func AskOllama(promptStr string) (string, error) {
	// return ollamaResponse.Answer, nil
	llm, err := gollm.NewLLM(
		gollm.SetProvider("ollama"),
		gollm.SetModel("llama3:latest"),
		gollm.SetDebugLevel(gollm.LogLevelWarn),
	)
	if err != nil {
		log.Fatalf("Failed to create LLM: %v", err)
	}

	// Create a prompt using NewPrompt function
	prompt := gollm.NewPrompt(promptStr)

	// Generate a response
	ctx := context.Background()
	response, err := llm.Generate(ctx, prompt)
	if err != nil {
		log.Fatalf("Failed to generate response: %v", err)
	}

	return response, nil

}
