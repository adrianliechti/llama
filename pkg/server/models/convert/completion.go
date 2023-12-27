package convert

import (
	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/server/models"
)

func MessageRole(role provider.MessageRole) models.MessageRole {
	switch role {
	case provider.MessageRoleSystem:
		return models.MessageRoleSystem

	case provider.MessageRoleUser:
		return models.MessageRoleUser

	case provider.MessageRoleAssistant:
		return models.MessageRoleAssistant

	default:
		return ""
	}
}

func MessageResult(r provider.MessageResult) models.FinishReason {
	switch r {
	case "":
		return ""

	case provider.MessageResultStop:
		return models.FinishReasonStop

	case provider.MessageResultLength:
		return models.FinishReasonLength

	default:
		return ""
	}
}
