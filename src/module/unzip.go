package module

import (
	"net/http"
	"storage"
	"strings"
	"archive/zip"
	"aelog"
	"utils"
	"fmt"
)

const (
	UNZIP_NO = 0
	UNZIP_NORMAL = 1
	UNZIP_FORCE = 2
)

type unZip struct {}


func (u unZip) OnSave(file *storage.File, w http.ResponseWriter, r *http.Request, s *storage.Storage) (e error) {
	mode := u.needUnZip(r)
	if mode == UNZIP_NO {
		return
	}
	aelog.Debugf("UnZip: start unzip(%d) files to: %s (%s)", mode, file.Name, utils.HumanBytes(file.FSize))	
	if mode == UNZIP_FORCE {
		s.DeleteChilds(file.Name)
	}
	
	filesCount, filesSize, err := u.unZipTo(file.Name, file, s)
	if err != nil {
		aelog.Debugf("UnZip: error: %v", err)
		w.Header().Add("X-Ae-Unzip-Error", err.Error())
		return
	}
	aelog.Debugf("UnZip: unziped: %s (%d files, %s files size)", file.Name, filesCount, utils.HumanBytes(filesSize))
	w.Header().Add("X-Ae-Unzip-Count", fmt.Sprint(filesCount))
	w.Header().Add("X-Ae-Unzip-Size", utils.HumanBytes(filesSize))
	return
}

func (u unZip) unZipTo(to string, f *storage.File, s *storage.Storage) (filesCount, filesSize int64, err error) {
	
	if err = f.Open(); err != nil {
		return
	}
	defer f.Close()
	
	z, err := zip.NewReader(f.GetReader(), f.FSize)
	if err != nil {
		return
	}
	
	for _, zf := range z.File {
		if ! zf.FileInfo().IsDir() {
			fs, e := u.saveFile(to, zf, s)
			if e != nil {
				err = e
				return
			}
			filesSize += fs
			filesCount++
		}
	}	
	
	return
}

func (u unZip) saveFile(to string, zf *zip.File, s *storage.Storage) (fs int64, err error) {
	fname := strings.Trim(to, "/") + "/" + strings.Trim(zf.Name, "/")
	reader, err := zf.Open()
	if err != nil {
		return
	}
	defer reader.Close()
	if f, err := s.Add(fname, reader, zf.FileInfo().Size()); err == nil {
		return f.FSize, nil
	}
	return 
}

func (u unZip) needUnZip(r *http.Request) int {
	switch strings.ToLower(r.Header.Get("X-Ae-Unzip")) {
		case "1", "true", "yes":
			return UNZIP_NORMAL
		case "force":
			return UNZIP_FORCE
	}
	return UNZIP_NO
}