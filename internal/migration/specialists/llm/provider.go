package llm

import (
	"fmt"
	"os"

	"github.com/arianlopezc/Trabuco/internal/ai"
	"github.com/arianlopezc/Trabuco/internal/auth"
)

// defaultProvider returns an Anthropic provider built from the user's
// stored credentials (or the ANTHROPIC_API_KEY env var). All migration
// specialists share this provider.
//
// We resolve lazily so importers don't need to wire auth at registration
// time; specialists are registered in init() before main() loads any auth.
func defaultProvider() (ai.Provider, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		manager, err := auth.NewManager()
		if err != nil {
			return nil, fmt.Errorf("credential manager: %w", err)
		}
		cred, err := manager.GetCredential(auth.ProviderAnthropic)
		if err != nil {
			return nil, fmt.Errorf("no Anthropic API key configured: %w", err)
		}
		apiKey = cred.APIKey
	}
	return ai.NewAnthropicProvider(&ai.ProviderConfig{
		APIKey: apiKey,
		Model:  ai.ModelClaudeSonnet.ID, // Sonnet is the default; opus too expensive for routine specialist calls
	})
}

func readFileBest(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
