// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mistralai

import (
	"time"

	"go.wdy.de/nago/application/ai/library"
	"go.wdy.de/nago/pkg/xhttp"
)

type CreateLibraryRequest struct {
	ChunkSize   int    `json:"chunk_size,omitempty"`
	Description string `json:"description,omitempty"`
	Name        string `json:"name"`
}

type LibraryInfo struct {
	ChunkSize   *int      `json:"chunk_size"`
	CreatedAt   time.Time `json:"created_at"`
	Id          string    `json:"id"`
	Name        string    `json:"name"`
	NbDocuments int       `json:"nb_documents"`
	OwnerId     string    `json:"owner_id"`
	OwnerType   string    `json:"owner_type"`
	TotalSize   int       `json:"total_size"`
	UpdatedAt   time.Time `json:"updated_at"`
	Description string    `json:"description"`
}

func (i LibraryInfo) IntoLibrary() library.Library {
	return library.Library{
		ID:          library.ID(i.Id),
		Name:        i.Name,
		Description: i.Description,
	}
}

func (c *Client) CreateLibrary(req CreateLibraryRequest) (LibraryInfo, error) {
	var resp LibraryInfo
	err := xhttp.NewRequest().
		Client(c.c).
		BaseURL(c.base).
		Retry(c.retry).
		URL("libraries").
		Assert2xx(true).
		BearerAuthentication(c.token).
		BodyJSON(req).
		ToJSON(&resp).
		ToLimit(1024 * 1024).
		Post()

	return resp, err
}

func (c *Client) GetLibrary(id string) (LibraryInfo, error) {
	var resp LibraryInfo
	err := xhttp.NewRequest().
		Client(c.c).
		BaseURL(c.base).
		Retry(c.retry).
		URL("libraries/" + id).
		Assert2xx(true).
		BearerAuthentication(c.token).
		ToJSON(&resp).
		ToLimit(1024 * 1024).
		Get()

	return resp, err
}

func (c *Client) DeleteLibrary(id string) error {
	err := xhttp.NewRequest().
		Client(c.c).
		BaseURL(c.base).
		Retry(c.retry).
		URL("libraries/" + id).
		Assert2xx(true).
		BearerAuthentication(c.token).
		ToLimit(1024 * 1024).
		Delete()

	return err
}

func (c *Client) ListAllLibraries() ([]LibraryInfo, error) {
	var resp struct {
		Data []LibraryInfo `json:"data"`
	}
	err := xhttp.NewRequest().
		Client(c.c).
		BaseURL(c.base).
		Retry(c.retry).
		URL("libraries").
		Assert2xx(true).
		BearerAuthentication(c.token).
		ToJSON(&resp).
		ToLimit(1024 * 1024).
		Get()

	return resp.Data, err
}
