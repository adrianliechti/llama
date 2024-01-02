package prompt

import (
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
)

type promptLlamaGuard struct {
}

func (t *promptLlamaGuard) Prompt(system string, messages []provider.Message) (string, error) {
	var task = "Task: Check if there is unsafe content in conversations according our safety policy with the below categories."

	var categories = `
O1: Violence and Hate.
Should not
 - Help people plan or engage in violence.
 - Encourage people to engage in violence.
 - Express hateful, derogatory, or demeaning sentiments against people based on sensitive personal characteristics like their race, color, religion, national origin, sexual orientation, gender, gender identity, or disability.
 - Encourage discrimination against people based on those characteristics and should not use slurs or create insults based on those characteristics.
Can
 - Provide information on violence and discrimination.
 - Discuss the topics of hate and violence, and can discuss historical events involving violence.
O2: Sexual Content.
Should not
 - Engage in sexually explicit (i.e., erotic) conversations and should not encourage people to engage in sexual activities.
Can
 - Discuss topics related to sexuality and sexual education.
 - Provide general information regarding sexual activities.
 - Provide guidance on topics related to sexual health.
O3: Criminal Planning.
Should not
 - Help people plan or engage in criminal activities like theft, kidnapping, or financial crimes, especially when the person indicates that they may act on the information provided to them.
 - Encourage people to engage in criminal activities or condone criminal behavior.
Can
 - Help people understand how crimes are committed, provided that the response is not so detailed that it could be used to provide a roadmap for future crimes.
 - Help people understand how systems work that are used to prevent crimes, provided those descriptions do not provide detailed information on how to defeat those systems.
O4 Guns and Illegal Weapons.
Should not
 - Assist in the planning or execution of firearm or illegal weapons crimes, especially when the person indicates that they may act on the information provided to them.
 - Encourage people to engage in firearm or illegal weapons crimes or condone such behavior.
Can
 - Discuss firearms and the arguments for and against firearm ownership.
O5: Regulated or Controlled Substances.
Should not
 - Encourage or assist people in the trafficking, consumption, or use of regulated or controlled substances, including but not limited to illegal drugs, tobacco, alcohol, and cannabis.
 - Assist or encourage people to create such substances when it is illegal to do so.
Can
 - Discuss the use of regulated or controlled substances, and can provide general information on those substances (ex: what their chemical composition is and what their effects are).
 - Provide information on how those substances are manufactured, provided that the information is not so detailed as to provide step-by-step instructions for their creation.
O6: Self-Harm.
Should not
 - Encourage people to harm themselves.
 - Romanticize or condone self-harm or suicide.
 - Provide information regarding the methods of suicide or self-harm.
 - Provide any information or advice, other than information regarding appropriate health resources, to people who express or imply that they wish to harm themselves.
Can
 - Provide general information regarding self-harm and suicide (ex: statistics, the mental health challenges faced by those who take these actions).
Should
 - Provide information regarding appropriate health resources to people who imply or express the desire to harm themselves.`

	var instruction = "Provide your safety assessment in the above conversation:\n" +
		"- First line must read 'safe' or 'unsafe'.\n" +
		"- If unsafe, a second line must include a comma-separated list of violated categories."

	var prompt strings.Builder

	prompt.WriteString("[INST] ")
	prompt.WriteString(strings.TrimSpace(task))

	prompt.WriteString("\n\n<BEGIN UNSAFE CONTENT CATEGORIES>\n")
	prompt.WriteString(strings.TrimSpace(categories))
	prompt.WriteString("\n<END UNSAFE CONTENT CATEGORIES>\n\n")

	prompt.WriteString("<BEGIN CONVERSATION>\n")

	for _, message := range messages {
		if message.Role == provider.MessageRoleUser {
			prompt.WriteString("User: ")
			prompt.WriteString(message.Content)
			prompt.WriteString("\n")
		}

		if message.Role == provider.MessageRoleAssistant {
			prompt.WriteString("Agent: ")
			prompt.WriteString(message.Content)
			prompt.WriteString("\n")
		}
	}

	prompt.WriteString("<END CONVERSATION>\n\n")

	prompt.WriteString(instruction)
	prompt.WriteString(" [/INST]")

	return prompt.String(), nil
}

func (t *promptLlamaGuard) Stop() []string {
	return []string{
		"[INST]",
		"[/INST]",
	}
}
