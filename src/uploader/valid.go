package uploader

import (
	"temp"
	"errors"
)

var ValidErrTooLarge = errors.New("File too large")

type Valid struct {
	MaxSize    int64    `json:"max_size,omitempty"`
	MimeTypes  []string `json:"mime_types,omitempty"`
}

func (v *Valid) HasError(tf *temp.File) (err error) {
	if v.MaxSize > 0 {
		if tf.Size > v.MaxSize {
			err = ValidErrTooLarge
		}
	}
	return
}