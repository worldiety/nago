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
	"time"

	"go.wdy.de/nago/application/ai/file"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/pkg/xhttp"
	"go.wdy.de/nago/pkg/xtime"
	"go.wdy.de/nago/presentation/core"
)

type FileSchema struct {
	Id         string      `json:"id"`
	Object     string      `json:"object"`
	Bytes      *int64      `json:"bytes"`
	CreatedAt  int64       `json:"created_at"`
	Filename   string      `json:"filename"`
	Purpose    string      `json:"purpose"`
	SampleType string      `json:"sample_type"`
	Source     string      `json:"source"`
	NumLines   int         `json:"num_lines"`
	Mimetype   string      `json:"mimetype"`
	Signature  interface{} `json:"signature"`
}

func (s FileSchema) IntoFile() file.File {
	return file.File{
		ID:        file.ID(s.Id),
		Name:      s.Filename,
		MimeType:  file.Type(s.Mimetype),
		CreatedAt: xtime.UnixMilliseconds(time.Unix(s.CreatedAt, 0).UnixMilli()),
	}
}

type ListFileResponse struct {
	Data   []FileSchema
	Object string `json:"object"` // list
	Total  int    `json:"total"`
}

func (c *Client) ListFiles() (ListFileResponse, error) {
	var resp ListFileResponse
	err := c.newReq().
		URL("conversations/files").
		Assert2xx(true).
		BearerAuthentication(c.token).
		ToJSON(&resp).
		ToLimit(1024 * 1024 * 128).
		Get()

	return resp, err
}

func (c *Client) GetFile(id string) (FileSchema, error) {
	var resp FileSchema
	err := c.newReq().
		URL("files/" + id).
		Assert2xx(true).
		BearerAuthentication(c.token).
		ToJSON(&resp).
		ToLimit(1024 * 1024).
		Get()

	return resp, err
}

func (c *Client) UploadFile(filename string, reader io.Reader) (FileSchema, error) {
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return FileSchema{}, err
	}

	_, err = io.Copy(part, reader)
	if err != nil {
		return FileSchema{}, err
	}

	// TODO important to tell the API what this is. OCR is the only thing which accepts arbitrary files
	if err := writer.WriteField("purpose", "ocr"); err != nil {
		return FileSchema{}, err
	}

	if err := writer.Close(); err != nil {
		return FileSchema{}, err
	}

	var resp FileSchema
	err = c.newReq().
		Header("Content-Type", writer.FormDataContentType()).
		URL("files").
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

func (c *Client) DeleteFile(fileId string) error {
	return c.newReq().
		Assert2xx(true).
		URL("files/" + fileId).
		BearerAuthentication(c.token).
		Delete()
}

func (c *Client) DownloadFile(fileId string) (io.ReadCloser, error) {
	var resp io.ReadCloser
	err := c.newReq().
		URL("files/" + fileId + "/content").
		Assert2xx(true).
		BearerAuthentication(c.token).
		ToCloser(func(readCloser io.ReadCloser) {
			resp = readCloser
		}).
		Get()

	return resp, err
}

func (c *Client) GetSignedURL(fileId string) (core.URI, error) {
	var resp struct {
		URL core.URI `json:"url"`
	}
	err := c.newReq().
		URL("files/" + fileId + "/url").
		Assert2xx(true).
		BearerAuthentication(c.token).
		ToJSON(&resp).
		ToLimit(1024 * 1024).
		Get()

	return resp.URL, err
}
