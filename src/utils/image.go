/*
  Copyright 2012 Sergey Cherepanov (https://github.com/cheggaaa)

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.
*/

package utils

import (
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


func (i *Image) Resize(dst, format string, w, h, q int) error {
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
	
	aelog.Debugln("Image: convert from:", i.Filename, "to:", dst, "w:", w, "h:", h, "q:", q)
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
	
	if err != nil {
		return err
	}
	i.Width = w
	i.Height = h
	i.Type = format
	return nil
} 

func (i *Image) Crop(dst, format string, w, h, q int) error {
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
	err := i.Resize(dst, format, rw, rh, q)
	if err != nil {
		return err
	}
	crop := fmt.Sprintf("%dx%d+0+0", w, h)
	cmd := exec.Command("convert", dst, "-gravity", "Center", "-crop", crop, dst)
	_, err = cmd.Output()
	if err != nil {
		return err
	}
	i.Width = s
	i.Height = s
	return nil
} 
