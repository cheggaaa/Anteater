package uploader

import (
	"storage"
	"config"
	"net/http"
	"aelog"
	"io"
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
	
	aelog.Debugf("Uploader: token found: %s\n", token)
	
	status = true
	
	write := func(token string, params interface{}, err error) {
		var resp *http.Response
		if err != nil {
			resp, err = u.ctrl.SetStatusErr(token, err)
		} else {
			resp, err = u.ctrl.SetStatus(token, params)
		}
		if err != nil {
			errCode = 500
			return
		}
		w.Header().Set("Content-Type", "application/x-javascript")
		io.Copy(w, resp.Body)
		w.WriteHeader(resp.StatusCode)
		err = nil
	}
	
	files, err := u.ctrl.GetParams(token)
	if err != nil {
		if err == ErrTokenNotFound {
			aelog.Debugln("Uploader: ctrl return not found")
			errCode = 404
		} else {
			errCode = 500
			aelog.Debugf("Uploader: ctrl return error: %v\n", err)
			write(token, nil, err)
		}
		return
	}
	
	
	err = files.Upload(u.conf, u.stor, r, w)
	
	if err != nil {
		aelog.Debugf("Uploader: files.Upload has err: %v\n", err)
		errCode = 500
		write(token, nil, err)
		return
	}
	write(token, files, nil)
	aelog.Debugln("Uploader: files.Upload done")
	return
}