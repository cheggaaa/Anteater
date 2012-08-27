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

package anteater


import (
	"os"
	"os/exec"
	"strings"
	"errors"
	"strconv"
	"fmt"
)

type Image struct {
	Filename string
	Type     string
	Width    int
	Height   int
}

func ImageIdenty(filename string) (*Image, error) {
	var err error
	
	identify, err := exec.LookPath("identify")
	
	if err != nil {
		return nil, err
	}
	Log.Debugln(identify)
	cmd := exec.Command(identify, filename)
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
	
	Log.Debugln("Identy:", image)
	
	return image, nil
}


func (i *Image) Resize(format string, w, h, q int) error {
	wh := fmt.Sprintf("%dx%d", w, h)
	if q > 0 {
		wh = fmt.Sprintf("%s -quality %d", q)
	}
	
	dst := i.Filename + "." + strings.ToLower(format)
	
	Log.Debugln("Convert from:", i.Filename, "to:", dst, "w:", w, "h:", h, "q:", q)
	
	convert, err := exec.LookPath("convert")
	if err != nil {
		return err
	}
	cmd := exec.Command(convert, i.Filename, "-resize", wh, dst)
	_, err = cmd.Output()
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
	return nil
} 