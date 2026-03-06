// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import "go.wdy.de/nago/pkg/sitemap"

// AddSitemapURL adds a sitemap URL to the sitemap. If no sitemap is configured, a new one is created and
// will be provided at the /sitemap.xml endpoint.
func (c *Configurator) AddSitemapURL(url sitemap.URL) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.sitemap == nil {
		c.sitemap = sitemap.NewSitemap()
	}

	c.sitemap.AddURL(url)
}
