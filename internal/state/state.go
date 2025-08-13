package state

import (
	config "github.com/federicoReghini/Blog-aggregator/internal/config"
	"github.com/federicoReghini/Blog-aggregator/internal/database"
)

type State struct {
	Db  *database.Queries
	Cfg *config.Config
}
