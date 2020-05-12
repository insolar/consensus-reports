package report

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/insolar/consensus-reports/pkg/middleware"
)

func TestWebdavClient_ReadReportData(t *testing.T) {
	cfg := middleware.WebDavConfig{
		URL: "https://webdav.yandex.ru",
		User: "fspecter",
		Password: "awkward20",
		Timeout: time.Second * 10,
	}
	path := "/fake102"

	c := CreateWebdavClient(cfg, path)
	err := c.ReadReportData()
	assert.NoError(t, err)
}
