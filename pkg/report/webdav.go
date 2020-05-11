package report

import (
	"fmt"

	"github.com/studio-b12/gowebdav"
)

type Config struct {
	URL string
	User string
	Password string
}


type WebdavClient struct {

}

func() ReadReportData() {
	path := "/fake102"
	cfg := Config{URL: "https://webdav.yandex.ru", User: "fspecter", Password: "awkward20"}
	client := gowebdav.NewClient(cfg.URL, cfg.User, cfg.Password)

	files, _ := client.ReadDir(path)
	for _, file := range files {
		// notice that [file] has os.FileInfo type
		fmt.Println(file)
		// sort by date
	}
}