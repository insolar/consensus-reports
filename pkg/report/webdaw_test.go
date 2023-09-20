package report

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/insolar/consensus-reports/pkg/middleware"
)

func TestWebdavClient_ReadReportData(t *testing.T) {
	cfg := Config{
		Webdav: middleware.WebDavConfig{
			Host:      "https://webdav.yandex.ru",
			Username:  "fspecter",
			Password:  "awkward20",
			Directory: "/fake102",
			Timeout:   time.Second * 10,
		},
		Git: struct {
			Branch string
			Hash   string
		}{"master", "aabbcc"},
	}

	c := CreateWebdavClient(cfg)
	_, err := c.ReadTemplateData()
	assert.NoError(t, err)
}
