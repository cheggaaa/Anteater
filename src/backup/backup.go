package backup

import (
	"github.com/cheggaaa/Anteater/src/storage"
	/*"github.com/cheggaaa/Anteater/src/config"
	"time"
	"errors"
	"github.com/cheggaaa/Anteater/src/aelog"
	"strings"
	*/)

func CreateBackup(s *storage.Storage, toPath string) (err error) {
	/*
			toPath = strings.TrimRight(toPath, "/") + "/"

			var backup *storage.Storage
			defer func() {
				if backup != nil {
			        backup.Close()
			        backup = nil
		        }
				if r := recover(); r != nil {
		        	aelog.Warnf("Panic while backup: %v\n", r)
		        	err = r.(error)
		        }
			}()

			// create config for backup storage
			conf := &config.Config{}
			conf.ContainerSize = s.Conf.ContainerSize
			conf.DumpTime = 0
			conf.DataPath = toPath
			if conf.DataPath == s.Conf.DataPath {
				return errors.New("Can't backup data to same dir")
			}
			// init backup storage
			backup = storage.GetStorage(conf)

			syncFile := func(name string) (sync bool) {
				file, ok := s.Index.Get(name)
				if ! ok {
					sync = backup.Delete(name)
					return
				}
				file.Open()
				defer file.Close()

				bfile, bok := backup.Get(name)
				if bok {
					if bfile.Size == file.Size && bfile.Md5S() == file.Md5S() {
						return false
					} else {
						backup.Delete(name)
					}
				}

				bfile = backup.Add(name, file.GetReader(), file.Size)
				bfile.Time = file.Time
				sync = true
				return
			}


			aelog.Infof("BACKUP: Started %v", time.Now())
			if backup.Index.Version() > 0 {
				i := 0
				for name, _ := range backup.Index.Files {
					_, ok := s.Index.Get(name)
					if ! ok {
						backup.Delete(name)
						i++
					}
				}
				aelog.Infof("BACKUP: Deleted %d files from backup", i)
			}

			i := 0

			for name, _ := range s.Index.Files {
				if syncFile(name) {
					i++
				}
			}
			aelog.Infof("BACKUP: %d files was sync", i)

			err = backup.Dump()
	*/
	return
}
