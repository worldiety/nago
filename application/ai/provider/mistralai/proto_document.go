// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mistralai

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"strconv"
	"time"

	"go.wdy.de/nago/application/ai/document"
	"go.wdy.de/nago/application/ai/library"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/pkg/xhttp"
	"go.wdy.de/nago/pkg/xtime"
	"go.wdy.de/nago/presentation/core"
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
	err = c.newReq().
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

	var statErr xhttp.UnexpectedStatusCodeError
	if errors.As(err, &statErr) {
		if statErr.StatusCode == 429 {
			return resp, fmt.Errorf("%w: %w", err, provider.TooManyRequests)
		}
	}

	return resp, err
}

type Pagination struct {
	TotalItems  int  `json:"total_items"`
	TotalPages  int  `json:"total_pages"`
	CurrentPage int  `json:"current_page"`
	PageSize    int  `json:"page_size"`
	HasMore     bool `json:"has_more"`
}

func (c *Client) ListDocuments(id string) ([]DocumentInfo, error) {
	var resp struct {
		Data       []DocumentInfo `json:"data"`
		Pagination Pagination     `json:"pagination"`
	}
	err := c.newReq().
		URL("libraries/"+id+"/documents").
		Query("page", "0").
		Query("page_size", "100").
		Assert2xx(true).
		BearerAuthentication(c.token).
		ToJSON(&resp).
		ToLimit(1024 * 1024 * 1024).
		Get()

	// TODO their paging logic is broken anyway and the API changes on a daily basis, that is not beta, just WIP

	if !resp.Pagination.HasMore {
		return resp.Data, nil
	}

	var tmp []DocumentInfo
	tmp = append(tmp, resp.Data...)
	for resp.Pagination.HasMore {
		err = c.newReq().
			URL("libraries/"+id+"/documents").
			Query("page", strconv.Itoa(resp.Pagination.CurrentPage+1)).
			Query("page_size", "100").
			Assert2xx(true).
			BearerAuthentication(c.token).
			ToJSON(&resp).
			ToLimit(1024 * 1024 * 1024).
			Get()

		if err != nil {
			return nil, err
		}

		tmp = append(tmp, resp.Data...)
	}

	return tmp, err
}

func (c *Client) DeleteDocument(lib, doc string) error {
	return c.newReq().
		Assert2xx(true).
		URL("libraries/" + lib + "/documents/" + doc).
		BearerAuthentication(c.token).
		Delete()
}

func (c *Client) GetDocument(libId, docId string) (DocumentInfo, error) {
	var resp DocumentInfo
	err := c.newReq().
		URL("libraries/" + libId + "/documents/" + docId).
		Assert2xx(true).
		BearerAuthentication(c.token).
		ToJSON(&resp).
		ToLimit(1024 * 1024).
		Get()

	return resp, err
}

func (c *Client) GetDocumentText(libId, docId string) (string, error) {
	var resp struct {
		Text string `json:"text"`
	}
	err := c.newReq().
		URL("libraries/" + libId + "/documents/" + docId + "/text_content").
		Assert2xx(true).
		BearerAuthentication(c.token).
		ToJSON(&resp).
		Get()

	return resp.Text, err
}

func (c *Client) GetDocumentStatus(libId, docId string) (document.ProcessingStatus, error) {
	var resp struct {
		DocumentId       string                    `json:"document_id"`
		ProcessingStatus document.ProcessingStatus `json:"processing_status"`
	}
	err := c.newReq().
		URL("libraries/" + libId + "/documents/" + docId + "/status").
		Assert2xx(true).
		BearerAuthentication(c.token).
		ToJSON(&resp).
		Get()

	return resp.ProcessingStatus, err
}

func (c *Client) GetDocumentSignedURL(libID, fileID string) (core.URI, error) {
	var resp string
	err := c.newReq().
		URL("libraries/" + libID + "/documents/" + fileID + "/signed-url").
		Assert2xx(true).
		BearerAuthentication(c.token).
		ToJSON(&resp).
		ToLimit(1024 * 1024).
		Get()

	return core.URI(resp), err
}

func (c *Client) GetDocumentDownload(libID, fileID string) (io.ReadCloser, error) {
	// TODO this endpoint does not exist, thus emulate it via download link
	uri, err := c.GetDocumentSignedURL(libID, fileID)
	if err != nil {
		return nil, err
	}

	var resp io.ReadCloser
	err = c.newReq().
		BaseURL("").
		URL(string(uri)). // note that we got an absolute URL
		Assert2xx(true).
		//BearerAuthentication(c.token). // note that we got a signed request
		ToCloser(func(readCloser io.ReadCloser) {
			resp = readCloser
		}).
		Get()

	return resp, err
}
