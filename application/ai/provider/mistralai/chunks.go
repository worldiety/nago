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
	"log/slog"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/file"
	"go.wdy.de/nago/application/ai/message"
	"go.wdy.de/nago/pkg/xjson"
	"go.wdy.de/nago/pkg/xtime"
	"go.wdy.de/nago/presentation/core"
)

var chunkVariants = []xjson.VariantOption{
	xjson.Variant[TextChunk]("text"),
	xjson.Variant[ToolFileChunk]("tool_file"),
	xjson.Variant[DocumentURLChunk]("document_url"),
	xjson.Variant[FileChunk]("file"),
}

type ChunkBox struct {
	Values []Chunk
}

func (c ChunkBox) IntoMessages(id string, role message.Role) []message.Message {
	var tmp []message.Message
	now := xtime.Now()
	for _, value := range c.Values {
		switch v := value.(type) {
		case TextChunk:
			tmp = append(tmp, message.Message{
				ID:           message.ID(id),
				CreatedAt:    now,
				Role:         role,
				MessageInput: option.Pointer[string](&v.Text),
			})
		case ToolFileChunk:
			var t file.Type
			switch v.FileType {
			case "png":
				t = file.PNG
			default:
				slog.Warn("unknown bogus mistral filetype", "type", v.FileType)
				t = file.Binary
			}

			f := file.File{
				ID:       file.ID(v.FileId),
				Name:     v.FileName,
				MimeType: t,
			}

			tmp = append(tmp, message.Message{
				ID:        message.ID(id),
				CreatedAt: now,
				Role:      role,
				File:      option.Pointer(&f),
			})
		case DocumentURLChunk:
			doc := message.DocumentURL{
				Name: v.Name,
				URL:  core.URI(v.Url),
			}
			tmp = append(tmp, message.Message{
				ID:          message.ID(id),
				CreatedAt:   now,
				Role:        role,
				DocumentURL: option.Pointer(&doc),
			})
		default:
			panic(fmt.Errorf("implement me %T", v))
		}
	}

	return tmp
}

func (c ChunkBox) MarshalJSON() ([]byte, error) {
	return xjson.MarshalInternally("type", c.Values, chunkVariants...)
}

func (c *ChunkBox) UnmarshalJSON(data []byte) error {
	if len(data) > 1 && data[0] == '"' {
		// the dreaded string content literal
		var tmp string
		if err := json.Unmarshal(data, &tmp); err != nil {
			return err
		}

		c.Values = nil
		c.Values = append(c.Values, TextChunk{
			Type: "text",
			Text: tmp,
		})

		return nil
	}

	tmp, err := xjson.UnmarshalInternally("type", data, chunkVariants...)
	if err != nil {
		return err
	}
	c.Values = nil
	switch t := tmp.(type) {
	case []any:
		for _, a := range t {
			c.Values = append(c.Values, a.(Chunk))
		}
	default:
		c.Values = append(c.Values, tmp.(Chunk))
	}

	return nil
}

type Chunk interface {
	isChunk()
}

type TextChunk struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (TextChunk) isChunk() {}

type ToolFileChunk struct {
	Type     string `json:"type,omitempty"`
	Tool     string `json:"tool,omitempty"`
	FileId   string `json:"file_id,omitempty"`
	FileName string `json:"file_name,omitempty"`
	FileType string `json:"file_type,omitempty"`
}

func (ToolFileChunk) isChunk() {}

// FileChunk does not exist for beta conversation API, however the ToolFileChunk as input does not work either
// and also does not make sense at all.
type FileChunk struct {
	FileId string `json:"file_id,omitempty"`
}

func (FileChunk) isChunk() {}

type DocumentURLChunk struct {
	Name string `json:"document_name"`
	Url  string `json:"document_url"`
}

func (DocumentURLChunk) isChunk() {}
