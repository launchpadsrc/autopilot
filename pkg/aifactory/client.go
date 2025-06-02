package aifactory

import (
	"log"
	"os"

	"github.com/sashabaranov/go-openai"
)

func Client() *openai.Client {
	var (
		kind     = os.Getenv("OPENAI_TYPE")
		endpoint = os.Getenv("OPENAI_ENDPOINT")
		version  = os.Getenv("OPENAI_API_VERSION")
		key      = os.Getenv("OPENAI_API_KEY")
	)
	switch kind {
	case "azure":
		config := openai.DefaultAzureConfig(key, endpoint)
		config.APIVersion = version
		return openai.NewClientWithConfig(config)
	case "openai":
		return openai.NewClient(key)
	default:
		log.Fatalln("unknown OPENAI_TYPE:", kind)
	}
	return nil
}
