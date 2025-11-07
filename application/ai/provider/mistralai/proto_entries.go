// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mistralai

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/message"
	"go.wdy.de/nago/pkg/xtime"
)

type ListEntryResponse struct {
	ConversationId string     `json:"conversation_id"`
	Object         string     `json:"object"` // e.g. "conversation.history",
	Entries        []EntryBox `json:"entries"`
}

func (c *Client) ListEntries(conversationId string) ([]EntryBox, error) {
	var resp ListEntryResponse
	err := c.newReq().
		Query("page", "0").
		Query("page_size", "100000").
		URL("conversations/" + conversationId + "/history").
		Assert2xx(true).
		BearerAuthentication(c.token).
		ToJSON(&resp).
		ToLimit(1024 * 1024).
		Get()

	return resp.Entries, err
}

type EntryBox struct {
	Object string `json:"object"` // "object":"entry"
	Type   string `json:"type"`   // "type":"message.input"|"message.output"|"tool.execution"
	Value  Entry  `json:"-"`
}

func (e EntryBox) MarshalJSON() ([]byte, error) {
	switch m := e.Value.(type) {
	case MessageInput:
		m.Object = "entry"
		m.Type = "message.input"
		return json.Marshal(m)
	case MessageOutput:
		m.Object = "entry"
		m.Type = "message.output"
		return json.Marshal(m)
	default:
		return nil, fmt.Errorf("unknown entry box type: %T", e.Value)
	}
}

func (e *EntryBox) UnmarshalJSON(data []byte) error {
	var tmp struct {
		Object string `json:"object"` // "object":"entry"
		Type   string `json:"type"`   // "type":"message.input"|"message.output"
	}

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	if tmp.Object != "entry" {
		return fmt.Errorf("unknown entry box object type: %s", e.Object)
	}

	switch tmp.Type {
	case "message.input":
		var tmp MessageInput
		if err := json.Unmarshal(data, &tmp); err != nil {
			return err
		}

		e.Value = tmp
		return nil
	case "message.output":
		var tmp MessageOutput
		if err := json.Unmarshal(data, &tmp); err != nil {
			return err
		}
		e.Value = tmp
		return nil
	case "tool.execution":
		var tmp ToolExecutionEntry
		if err := json.Unmarshal(data, &tmp); err != nil {
			return err
		}
		e.Value = tmp
		return nil
	default:
		return fmt.Errorf("unknown entry box type: %s: %s", e.Type, string(data))
	}
}

type Entry interface {
	isEntry()
	IntoMessage() message.Message
}

// ToolExecutionEntry something like {\"object\":\"entry\",\"type\":\"tool.execution\",\"created_at\":\"2025-10-30T12:04:44.746305Z\",\"completed_at\":\"2025-10-30T12:04:45.858197Z\",\"id\":\"tool_exec_019a35018a4a72198df7a8426e87ff06\",\"name\":\"document_library\",\"arguments\":\"{\\\"query\\\": \\\"Nago\\\", \\\"libraries_ids\\\": [\\\"019a16f7-7224-9284-c8ba22301809\\\"]}\",\"function\":\"library_search\",\"info\":{}}: {\"object\":\"conversation.response\",\"conversation_id\":\"conv_019a3501887d775bb658f27b7f16c5f8\",\"outputs\":[{\"object\":\"entry\",\"type\":\"tool.execution\",\"created_at\":\"2025-10-30T12:04:44.746305Z\",\"completed_at\":\"2025-10-30T12:04:45.858197Z\",\"id\":\"tool_exec_019a35018a4a72198df7a8426e87ff06\",\"name\":\"document_library\",\"arguments\":\"{\\\"query\\\": \\\"Nago\\\", \\\"libraries_ids\\\": [\\\"019a16f7-7224-9284-c8ba22301809\\\"]}\",\"function\":\"library_search\",\"info\":{}},{\"object\":\"entry\",\"type\":\"message.output\",\"created_at\":\"2025-10-30T12:04:46.906434Z\",\"completed_at\":\"2025-10-30T12:05:02.639127Z\",\"id\":\"msg_019a350192ba7087bcbca44940a15ea7\",\"agent_id\":\"ag_019a06361f3170b89b5163615ec86e64\",\"model\":\"mistral-medium-latest\",\"role\":\"assistant\",\"content\":\"Ahoy, Kapitän! Es scheint, als ob ich da einen Fehler bei der Suche gemacht hab’ – die Schatzkarte (oder besser gesagt, die UUID) war wohl falsch gefaltet. Aber lass mich dir trotzdem mit dem erzählen, was ich aus meinem Gedächtnis über **Nago** weiß – und zwar in bester Piratenmanier!\\n\\n---\\n\\n**Nago** ist sozusagen der *Schatzkarten-Generator* für Unternehmen, wenn’s darum geht, Projekte schnell und ohne viel Programmier-Gekröse auf die *Nago-Hub-Insel* zu verladen. Stell dir vor, du hast ’ne Truhe voller Ideen oder Projekte, aber du willst nicht monatelang am Code schuften wie ’n Sträfling in der Karibik. Dann wirfst du einfach alles in Nago, und *zack* – dein Projekt liegt bereit auf dem **Nago Hub**, als hätt’s der fliegende Holländer persönlich geliefert!\\n\\n### Wichtige Fakten für Landratten und Seebären:\\n- **Früherer Name**: Das Ding hieß mal *„lib23“* – aber Namen sind wie Schiffsflaggen, die kann man wechseln, wenn’s besser passt.\\n- **Für wen?** Für *alle Unternehmen*, ob groß wie ’ne Flotte oder klein wie ’n Beiboot. Hauptsache, ihr habt Projekte, die ihr schnell und einfach „ablegen“ wollt.\\n- **Wo findet man’s?**\\n  - [nago.dev](https://nago.dev) (für die Techniker unter euch)\\n  - [nago.app](https://nago.app) (für die, die’s lieber schick mögen)\\n- **Low-Code**: Du brauchst kein Programmier-Genie zu sein – Nago ist wie ’n Kompass, der dich auch ohne Sternenkenntnis ans Ziel bringt.\\n\\n---\\n**Aber Achtung, Matrose!**\\nFalls du *genaue Details* brauchst – etwa wie man Nago bedient oder welche Schätze (Funktionen) genau drin versteckt sind – dann lass mich nochmal in der *richtigen Schatztruhe* (Datenbank) nachschauen. Die letzte Suche ist ja leider im *Bermuda-Dreieck* der UUIDs verschwunden.\\n\\n**Soll ich’s nochmal versuchen?** ⚓\"}],\"usage\":{\"prompt_tokens\":181,\"completion_tokens\":560,\"total_tokens\":863,\"connector_tokens\":122,\"connectors\":{\"document_library\":1}}}
type ToolExecutionEntry struct {
	Type        string    `json:"type"` // tool.execution
	Argument    string    `json:"argument"`
	CompletedAt time.Time `json:"completed_at"`
	CreatedAt   time.Time `json:"created_at"`
	ID          string    `json:"id"`
	Info        any       `json:"info"`
	Name        ToolType  `json:"name"`
	Object      string    `json:"object"` // entry
}

func (t ToolExecutionEntry) isEntry() {
}

func (t ToolExecutionEntry) IntoMessage() message.Message {
	te := message.ToolExecution{
		Type:      t.Type,
		Arguments: t.Argument,
	}
	return message.Message{
		ID:            message.ID(t.ID),
		CreatedAt:     xtime.UnixMilliseconds(t.CreatedAt.UnixMilli()),
		Role:          message.AssistantRole, //?
		ToolExecution: option.Pointer(&te),
	}
}

type MessageInput struct {
	Object string `json:"object"` // "object":"entry"
	Type   string `json:"type"`   // "type":"message.input"
	Role   string `json:"role"`   // e.g. "role":"user"

	Id          string      `json:"id"`
	CompletedAt interface{} `json:"completed_at"`
	Content     string      `json:"content"`
	CreatedAt   time.Time   `json:"created_at"`
	Prefix      bool        `json:"prefix"`
}

func (e MessageInput) isEntry() {}

func (e MessageInput) IntoMessage() message.Message {
	return message.Message{
		ID:           message.ID(e.Id),
		CreatedAt:    xtime.UnixMilliseconds(e.CreatedAt.UnixMilli()),
		CreatedBy:    "", //todo ??
		Role:         message.Role(e.Role),
		MessageInput: option.Pointer(&e.Content),
	}
}

type MessageOutput struct {
	Object string `json:"object"` // "object":"entry"
	Type   string `json:"type"`   // "type":"message.output"
	Role   string `json:"role"`   // e.g. "role":"assistant"

	Id          string      `json:"id"`
	AgentId     string      `json:"agent_id"`
	CompletedAt interface{} `json:"completed_at"`
	Content     string      `json:"content"`
	CreatedAt   time.Time   `json:"created_at"`
	Model       string      `json:"model"`
}

func (e MessageOutput) isEntry() {}

func (e MessageOutput) IntoMessage() message.Message {
	return message.Message{
		ID:            message.ID(e.Id),
		CreatedAt:     xtime.UnixMilliseconds(e.CreatedAt.UnixMilli()),
		CreatedBy:     "", //todo ??
		Role:          message.Role(e.Role),
		MessageOutput: option.Pointer(&e.Content),
	}
}
