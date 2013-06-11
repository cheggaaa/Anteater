package module

import (
	"net/http"
	"storage"
)

type Module interface {
	OnSave(file *storage.File, w http.ResponseWriter, r *http.Request, s *storage.Storage) (err error)
}

var modules = make([]Module, 0)

func RegisterModules() {
	modules = append(modules, unZip{})
}

func OnSave(file *storage.File, w http.ResponseWriter, r *http.Request, s *storage.Storage) (err error) {
	for _, m := range modules {
		err = m.OnSave(file, w, r, s)
		if err != nil {
			break
		}
	}
	return
}
