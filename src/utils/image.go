package utils

import (
	"os"
	"os/exec"
	"strings"
	"errors"
	"strconv"
	"fmt"
	"aelog"
)

type Image struct {
	Filename string
	Type     string
	Width    int
	Height   int
}

func ImageIdenty(filename string) (*Image, error) {
	var err error
	cmd := exec.Command("identify", filename)
	res, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	
	image := &Image{
		Filename: filename,
	}
	
	params := strings.Split(string(res), " ")
	if len(params) < 3 {
		return nil, errors.New("Indetify return ivalid data: " + string(res))
	}
	
	image.Type = strings.ToLower(params[1])
	
	wh := strings.Split(params[2], "x")
	if len(wh) < 2 {
		return nil, errors.New("Can't decode identify width/height: " + params[2])
	}	
	image.Width, err = strconv.Atoi(wh[0])
	if err != nil {
		return nil, err
	}
	image.Height, err = strconv.Atoi(wh[1])
	if err != nil {
		return nil, err
	}	
	
	return image, nil
}


func (i *Image) Resize(format string, w, h, q int) error {
	wh := fmt.Sprintf("%dx%d", w, h)
	if w == 0 {
		wh = fmt.Sprintf("x%d", h)
	}	
	if h == 0 {
		wh = fmt.Sprintf("%d", w)
	}
	
	if format == "" {
		format = i.Type
	}
	dst := i.Filename + "." + strings.ToLower(format)
	
	aelog.Debugln("Convert from:", i.Filename, "to:", dst, "w:", w, "h:", h, "q:", q)
	var cmd *exec.Cmd
	if q > 0 {
		cmd = exec.Command("convert", i.Filename, "-strip",  "-resize", wh, "-quality",  fmt.Sprintf("%d", q), dst)
	} else {
		cmd = exec.Command("convert", i.Filename, "-strip",  "-resize", wh, dst)
	}
	_, err := cmd.Output()
	if err != nil {
		return err
	}
	
	os.Remove(i.Filename)
	err = os.Rename(dst, i.Filename)
	if err != nil {
		return err
	}
	i.Width = w
	i.Height = h
	i.Type = format
	return nil
} 

func (i *Image) Crop(format string, w, h, q int) error {
	var s int 
	rw, rh := w, h
	if w > h {
		s = w
	} else {
		s = h
	}
	
	if i.Width > i.Height {
		rh = s
		rw = 0
	} else {
		rw = s
		rh = 0
	}
	err := i.Resize(format, rw, rh, q)
	if err != nil {
		return err
	}
	crop := fmt.Sprintf("%dx%d+0+0", w, h)
	cmd := exec.Command("convert", i.Filename, "-gravity", "Center", "-crop", crop, i.Filename)
	_, err = cmd.Output()
	if err != nil {
		return err
	}
	i.Width = s
	i.Height = s
	return nil
} 
