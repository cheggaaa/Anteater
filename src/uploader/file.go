package uploader

import (
	"storage"
	"temp"
	"errors"
	"fmt"
	"utils"
	"time"
)

type File struct {
	// name for save to anteater
	Name	string `json:"name,omitempty"`
	// type - image or file
	Type    string `json:"type,omitempty"`
	// field name in form
	Field   string `json:"field,omitempty"`
	// validate
	Valid   *Valid `json:"valid,omitempty"`
	// file state
	State   *FileState `json:"state,omitempty"`
		
	// Only for images
	// GIF, JPG, PNG
	Format  string `json:"format,omitempty"`
	// image width
	Width   int `json:"width,omitempty"`
	// image height
	Height  int `json:"height,omitempty"`
	// image quality (for jpg)
	Quality int `json:"quality,omitempty"`
	// need crop
	Crop    bool `json:"crop,omitempty"`
}


type FileState struct {
	Uploaded bool  `json:"uploaded,omitempty"`
	Size     int64 `json:"size,omitempty"`
	Md5      string `json:"md5,omitempty"`
}

type Error struct {
	Code int `json:"code,omitempty"`
	Msg string `json:"message,omitempty"`
}


func (f *File) Upload(tmpf *TmpFiles) (err error) {
	f.State = &FileState{}
	tf, err := tmpf.GetByField(f.Field)
	if f.checkErr(err) {
		return
	}
	
	if f.Valid != nil {
		if f.checkErr(f.Valid.HasError(tf)) {
			return
		}
	}
	
	var tfr *temp.File
	
	switch f.Type {
		case "file":
			tfr, err = f.uploadFile(tf)
			break;
		case "image":
			tfr, err = f.uploadImage(tf)
			break;
		default:
			err = errors.New("Undefined file type: " + f.Type)		
	}
	
	if f.checkErr(err) {
		return
	}
	
	tmpf.SetResult(f.Name, tfr)	
	return
}

func (f *File) uploadImage(tf *temp.File) (tfr *temp.File, err error) {
	image, err := utils.ImageIdenty(tf.Filename)
	if err != nil {
		return
	}
	
	tfr, err = tf.Clone("." + fmt.Sprintf("%d%d", time.Now().Unix(), time.Now().UnixNano()))
	if err != nil {
		return
	}
	tfr.Disconnect()
	tf.Disconnect()
	if f.Crop {
		image.Crop(tfr.Filename, f.Format, f.Width, f.Height, f.Quality)
	} else {
		image.Resize(tfr.Filename, f.Format, f.Width, f.Height, f.Quality)
	}
	err = tf.Connect()
	if err == nil {
		err = tfr.Connect()
	}
	return
}

func (f *File) uploadFile(tf *temp.File) (tfr *temp.File, err error) {
	tfr = tf
	return
}

func (f *File) SetState(file *storage.File) {
	f.State = &FileState{
		Uploaded : true,
		Size     : file.Size,
		Md5      : fmt.Sprintf("%x", file.Md5),
	}
}

func (f *File) checkErr(err error) bool {
	if err != nil {
		// store err to state later
		return true
	}
	return false
}