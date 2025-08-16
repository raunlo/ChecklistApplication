package connection

import (
	"sync"

	"github.com/raunlo/pgx-with-automapper/pool"
)

var (
	once sync.Once
	db   pool.Conn
)

// NewDatabaseConnection returns a singleton database connection pool.
// Creating the pool once avoids the overhead of establishing new
// database connections for every request.
func NewDatabaseConnection(cfg pool.DatabaseConfiguration) pool.Conn {
	once.Do(func() {
		db = pool.NewDatabasePool(cfg)
	})
	return db
}
