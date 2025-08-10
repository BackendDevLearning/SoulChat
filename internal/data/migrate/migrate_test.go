package migrate

import (
	"kratos-realworld/internal/conf"
	"kratos-realworld/internal/data"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
)

func TestInitDBTable(t *testing.T) {
	myconf, err := LoadConfig("../../../configs/config.yaml")
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	spew.Dump(myconf)
	db := data.NewDatabase(myconf)
	if err := InitDBTable(db); err != nil {
		t.Error(err)
	}
}

func LoadConfig(path string) (*conf.Data, error) {
	c := config.New(
		config.WithSource(
			file.NewSource(path),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		return nil, err
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		return nil, err
	}
	return bc.Data, nil
}
