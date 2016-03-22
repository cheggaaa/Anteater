package module

import (
	"encoding/json"
	"net/http"
	"storage"
	"strconv"
	//"strings"
)

const (
	fileListCommand = "filelist"
)

type FileList struct {
	Err  *string    `json:"error,omitempty"`
	List []FileInfo `json:"list"`
}

type FileInfo struct {
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	ContentType string `json:"content-type,omitmepty"`
	MD5         string `json:"md5,omitempty"`
	Modified    int64  `json:"modified,omitempty"`
}

type fileList struct{}

func (fl fileList) OnSave(file *storage.File, w http.ResponseWriter, r *http.Request, s *storage.Storage) (err error) {
	return
}

func (fl fileList) OnCommand(command, filename string, w http.ResponseWriter, r *http.Request, s *storage.Storage) (cont bool, err error) {
	if command != fileListCommand {
		return true, nil
	}
	var response = FileList{}
	list, err := s.Index.List(filename, fl.parseNested(r))
	if err != nil {
		errString := err.Error()
		response.Err = &errString
		fl.jsonResponse(w, response)
		return
	}
	response.List = make([]FileInfo, 0)
	for _, fn := range list {
		if f, ok := s.Get(fn); ok {
			response.List = append(response.List, FileInfo{
				Name:        f.Name[len(filename)+1:],
				Size:        f.FSize,
				ContentType: f.ContentType(),
				MD5:         f.Md5S(),
				Modified:    f.Time.Unix(),
			})
		}
	}
	fl.jsonResponse(w, response)
	return false, nil
}

func (fl fileList) parseNested(r *http.Request) int {
	if nestedS := r.Header.Get("X-Ae-Filelist-Depth"); nestedS != "" {
		if nested, _ := strconv.Atoi(nestedS); nested > 0 {
			return nested
		}
	}
	return 1
}

func (fl fileList) jsonResponse(w http.ResponseWriter, resp FileList) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(resp)
}
