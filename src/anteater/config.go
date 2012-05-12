package anteater

import (
	"github.com/kless/goconfig/config"
	"errors"
	"strconv"
)

var Conf *Config

type Config struct {
	// Data path
	DataPath      string
	ContainerSize int64
	MinEmptySpace  int64
	
	// Http
	HttpWriteAddr string
	HttpReadAddr  string
	
	// Http Headers
	Headers       map[string]string
	
	// Mime Types
	MimeTypes     map[string]string
	
	// Log
	LogLevel      int
	LogFile		  string
}


func LoadConfig(filename string) error {
	c, err := config.ReadDefault(filename)
	if err != nil {
		return err
	}
	
	// Data path
	dataPath, err := c.String("data", "path")
	if err != nil {
		return err
	}
	if len(dataPath) == 0 {
		return errors.New("Empty data path in config " + filename)
	}
	
	// Container size
	s, err := c.String("data", "container_size")
	if err != nil {
		return err
	}
	containerSize, err := GetSizeFromString(s)
	
	// Min empty space
	s, err = c.String("data", "min_empty_space")
	if err != nil {
		return err
	}
	minEmptySpace, err := GetSizeFromString(s)
	
	// Http write addr
	httpWriteAddr, err := c.String("http", "write_addr")
	if err != nil {
		return err
	}
	
	// Http read addr
	httpReadAddr, err := c.String("http", "read_addr")
	if err != nil || len(httpReadAddr) == 0 {
		httpReadAddr = httpWriteAddr
	}
	
	// Headers	
	headers := make(map[string]string, 0)
	hOpts, err := c.Options("http-headers")
	if err == nil {
		for _, o := range(hOpts) {
			v, err := c.String("http-headers", o)
			if err == nil && len(v) > 0 {
				headers[o] = v
			} 
		}
	}
	
	// Mime	
	mimeTypes := make(map[string]string, 0)
	mOpts, err := c.Options("mime-types")
	if err == nil {
		for _, o := range(mOpts) {
			v, err := c.String("mime-types", o)
			if err == nil && len(v) > 0 {
				mimeTypes["." + o] = v
			} 
		}
	}
	
	// Log level
	levels := map[string]int {
		"debug" : LOG_DEBUG,
		"info"  : LOG_INFO,
		"warn"  : LOG_WARN,
	}
	llv, err := c.String("log", "level")
	if err != nil {
		llv = "info"
	}
	logLevel, ok := levels[llv]
	if ! ok {
		logLevel = levels["info"]
	}
	
	// Log file
	logFile, err := c.String("log", "file")
	
	Conf = &Config{dataPath, containerSize, minEmptySpace, httpWriteAddr, httpReadAddr, headers, mimeTypes, logLevel, logFile}
	
	return nil
}

func GetSizeFromString(s string) (int64, error) {
	if s == "" {
		return 0, nil
	}

	var m int64 = 1

	switch s[len(s)-1] {
	case 'K', 'k':
		m = 1024
	case 'M', 'm':
		m = 1024 * 1024
	case 'G', 'g':
		m = 1024 * 1024 * 1024
	case 'T', 't':
		m = 1024 * 1024 * 1024 * 1024
	}

	if m != 1 {
		s = s[0 : len(s)-1]
	}

	res, err := strconv.ParseInt(s, 0, 64)
	
	if err != nil {
		return 0, err
	}
	
	return res * m, nil
}