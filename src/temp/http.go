package temp

import (
	"net/http"
	"io"
	"errors"
	"fmt"
)

func (f *File) LoadFromUrl(url string) (err error) {
	r, err := http.Get(url)
	if err != nil {
		return
	}
	defer r.Body.Close()
	
	err = f.Create()
	if err != nil {
		return
	}
	
	if r.StatusCode != 200 {
		err = errors.New(fmt.Sprintf("server return non-200 status: %s", r.Status))
		return
	}
	_, err = io.Copy(f.File, r.Body)
	
	if err == nil {
		err = f.setState()
	}
	
	return
}