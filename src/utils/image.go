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
	"errors"
	"fmt"
	"github.com/cheggaaa/Anteater/src/aelog"
	"math"
	"os/exec"
	"strconv"
	"strings"
)

type Image struct {
	Filename string
	Type     string
	Width    int
	Height   int
}

func Identify(filename string) (*Image, error) {
	var err error
	cmd := exec.Command("identify", filename)
	res, err := cmd.CombinedOutput()
	if err != nil {
		e := parseError(string(res))
		if e != "" {
			err = fmt.Errorf("%s", e)
		} else {
			err = fmt.Errorf("%s: %v", filename, err)
		}
		return nil, err
	}

	image := &Image{
		Filename: filename,
	}

	ress := string(res)
	ress = strings.Replace(ress, filename, "filename", 1)
	params := strings.Split(ress, " ")
	if len(params) < 3 {
		return nil, errors.New("Indetify return ivalid data: " + ress)
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

func (i *Image) Resize(dst, format string, w, h, q int, optimize bool) error {
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

	dstImagick := format + ":" + dst

	var cmd *exec.Cmd
	if q > 0 {
		cmd = exec.Command("convert", i.Filename, "-strip", "-resize", wh, "-quality", fmt.Sprintf("%d", q), dstImagick)
	} else {
		cmd = exec.Command("convert", i.Filename, "-strip", "-resize", wh, dstImagick)
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	i.Width = w
	i.Height = h
	i.Type = format

	if optimize {
		i.optimize(dst)
	}

	return nil
}

func (i *Image) Crop(dst, format string, w, h, q int, optimize bool) error {
	cw, ch := w, h
	kc := float64(w) / float64(h)
	ki := float64(i.Width) / float64(i.Height)

	if ki > kc {
		ch = i.Height
		cw = int(math.Ceil(float64(i.Height) * kc))
	} else {
		cw = i.Width
		ch = int(math.Ceil(float64(i.Width) / kc))
	}

	crop := fmt.Sprintf("%dx%d+0+0", cw, ch)
	cmd := exec.Command("convert", i.Filename, "-gravity", "Center", "-crop", crop, dst)

	res, err := cmd.CombinedOutput()
	if err != nil {
		e := parseError(string(res))
		if e != "" {
			err = fmt.Errorf("%s", e)
		}
		return err
	}
	i.Filename = dst

	err = i.Resize(dst, format, w, h, q, optimize)
	if err != nil {
		return err
	}
	i.Width = w
	i.Height = h
	return nil
}

func (i *Image) optimize(dst string) {
	if i.Type != "png" {
		return
	}
	command := "pngquant"
	if _, err := exec.LookPath(command); err != nil {
		aelog.Warnln("Optimize image: Command", command, "not found:", err)
		return
	}

	cmd := exec.Command(command, dst, "--force", "--output", dst)
	if res, err := cmd.CombinedOutput(); err != nil {
		aelog.Warnf("Optimize image: %s return error: %v (%s)", command, err, string(res))
	}
	aelog.Debugln("Optimize image:", dst, "success!")
}

func parseError(probe string) (err string) {
	parts := strings.Split(probe, "\n")
	l := len(parts)
	if l < 2 {
		return
	}
	return strings.TrimSpace(parts[l - 2])
}