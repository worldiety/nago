package uilegacy

import "io"

type DownloadToken string

// DownloadSource can be implemented by any core.Component to provide a dynamic download source for this concrete
// component and page instance.
type DownloadSource interface {
	DownloadSource() func() (io.Reader, error)
	DownloadToken() DownloadToken
}
