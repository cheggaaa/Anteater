package stats

import (
	"storage"
)

type Storage struct {
	ContainersCount int
	FilesCount int
	FilesSize int64
	TotalSize int64
	FreeSpace int64
	HoleCount int
	HoleSize int64
	IndexVersion uint64
}

func (ss *Storage) Refresh(s *storage.Storage) {
	ss.ContainersCount = len(s.Containers.Containers)
	ss.FilesCount = len(s.Index.Files)
	ss.IndexVersion = s.Index.Version()
	ss.FilesSize = 0
	ss.TotalSize = 0
	ss.HoleCount = 0
	ss.HoleSize = 0
	for _, c := range s.Containers.Containers {
		ss.TotalSize += c.Size
		hc, hs := c.Spaces.Stats()
		ss.HoleCount += hc
		ss.HoleSize += hs
	}
	
	for _, f := range s.Index.Files {
		ss.FilesSize += f.Size
	}
}