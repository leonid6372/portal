package models

import (
	"io"
)

type Image struct {
	Payload   io.Reader
	Name      string
	Size      int64
	Extension string
}
