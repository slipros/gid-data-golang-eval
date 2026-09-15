package service

import (
	"io"
)

type Upload struct {
	r UploadReader
	c UploadCloser
}

type UploadReader = io.Reader // want `GID-276: interface UploadReader is declared below type Upload \(line 7\)`

type UploadCloser io.Closer // want `GID-276: interface UploadCloser is declared below type Upload \(line 7\)`

type UploadSize int
