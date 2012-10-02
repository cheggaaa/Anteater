package uploader

import (
	"storage"
	"config"
	"net/http"
)

type Uploader struct {
	conf *config.Config
	stor *storage.Storage
	ctrl *Ctrl
}

func NewUploader(c *config.Config, s *storage.Storage) *Uploader {
	return &Uploader{conf : c, stor : s, ctrl : NewCtrl(c.UploaderCtrlUrl)}
}

func (u *Uploader) TryRequest(r *http.Request, w http.ResponseWriter) (status bool, err error, errCode int) {
	if ! u.conf.UploaderEnable {
		return
	}
	token := r.URL.Query().Get(u.conf.UploaderParamName)
	if token == "" {
		return
	}
	
	status = true
	
	files, err := u.ctrl.GetParams(token)
	if err != nil {
		if err == ErrTokenNotFound {
			errCode = 404
		} else {
			errCode = 500
		}
		return
	}
	
	err = files.Upload(u.conf, u.stor, r, w)
	if err != nil {
		errCode = 500
		return
	}
	
	return
}