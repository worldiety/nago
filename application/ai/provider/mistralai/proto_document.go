// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mistralai

import (
	"bytes"
	"io"
	"mime/multipart"
	"time"

	"go.wdy.de/nago/application/ai/document"
	"go.wdy.de/nago/application/ai/library"
	"go.wdy.de/nago/pkg/xhttp"
	"go.wdy.de/nago/pkg/xtime"
)

type DocumentInfo struct {
	CreatedAt             time.Time `json:"created_at"`
	Extension             string    `json:"extension"`
	Hash                  string    `json:"hash"`
	Id                    string    `json:"id"`
	LibraryId             string    `json:"library_id"`
	MimeType              string    `json:"mime_type"`
	Name                  string    `json:"name"`
	ProcessingStatus      string    `json:"processing_status"`
	Size                  int64     `json:"size"`
	TokensProcessingTotal int64     `json:"tokens_processing_total"`
	UploadedById          string    `json:"uploaded_by_id"`
	UploadedByType        string    `json:"uploaded_by_type"`
	Summary               string    `json:"summary"`
}

func (i DocumentInfo) IntoDocument() document.Document {
	return document.Document{
		ID:               document.ID(i.Id),
		CreatedAt:        xtime.UnixMilliseconds(i.CreatedAt.UnixMilli()),
		Hash:             i.Hash,
		Library:          library.ID(i.LibraryId),
		MimeType:         i.MimeType,
		Name:             i.Name,
		ProcessingStatus: document.ProcessingStatus(i.ProcessingStatus),
		Size:             i.Size,
		Summary:          i.Summary,
	}
}

type File struct {
	Content  []byte `json:"content"`
	FileName string `json:"fileName"`
}

type CreateDocumentRequest struct {
	File File `json:"file"`
}

func (c *Client) CreateDocument(libId string, filename string, reader io.Reader) (DocumentInfo, error) {
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return DocumentInfo{}, err
	}

	_, err = io.Copy(part, reader)
	if err != nil {
		return DocumentInfo{}, err
	}

	if err := writer.Close(); err != nil {
		return DocumentInfo{}, err
	}

	var resp DocumentInfo
	err = xhttp.NewRequest().
		Client(c.c).
		BaseURL(c.base).
		Retry(c.retry).
		Header("Content-Type", writer.FormDataContentType()).
		URL("libraries/" + libId + "/documents").
		Assert2xx(true).
		BearerAuthentication(c.token).
		Body(func() (io.Reader, error) {
			return &requestBody, nil
		}).
		ToJSON(&resp).
		ToLimit(1024 * 1024).
		Post()

	return resp, err
}

func (c *Client) ListDocuments(id string) ([]DocumentInfo, error) {
	var resp struct {
		Data        []DocumentInfo `json:"data"`
		CurrentPage int            `json:"current_page"`
		HasMore     bool           `json:"has_more"`
		PageSize    int            `json:"page_size"`
		TotalItems  int            `json:"total_items"`
		TotalPages  int            `json:"total_pages"`
	}
	err := xhttp.NewRequest().
		Client(c.c).
		BaseURL(c.base).
		Retry(c.retry).
		URL("libraries/"+id+"/documents").
		Query("page", "0").
		Query("page_size", "100000").
		Assert2xx(true).
		BearerAuthentication(c.token).
		ToJSON(&resp).
		ToLimit(1024 * 1024 * 1024).
		Get()

	return resp.Data, err
}

func (c *Client) DeleteDocument(lib, doc string) error {
	return xhttp.NewRequest().
		Client(c.c).
		BaseURL(c.base).
		Retry(c.retry).
		Assert2xx(true).
		URL("libraries/" + lib + "/documents/" + doc).
		BearerAuthentication(c.token).
		Delete()
}
