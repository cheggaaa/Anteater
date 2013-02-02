package nstorage

import (
	"config"
)

type Storage struct {
	Conf *config.Config
	Index *Index
}