package state

import (
	config "github.com/federicoReghini/gator/internal/config"
	"github.com/federicoReghini/gator/internal/database"
)

type State struct {
	Db  *database.Queries
	Cfg *config.Config
}
