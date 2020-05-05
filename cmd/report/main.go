package main;

import (
	"fmt"
	"os"

	"github.com/studio-b12/gowebdav"

	"github.com/insolar/consensus-reports/pkg/report"
)

type Config struct {
	URL string;
	User string;
	Password string;
}

func main() {
	err := report.MakeReport(report.ReportConfig{}, os.Stdout)
	if err != nil {
		panic(err)
	}

	path := "/cef3ab1e"
	cfg := Config{URL: "https://webdav.yandex.ru", User: "", Password: ""}
	client := gowebdav.NewClient(cfg.URL, cfg.User, cfg.Password)

	files, _ := client.ReadDir(path)
	for _, file := range files {
		// notice that [file] has os.FileInfo type
		fmt.Println(file)
		// sort by date
	}


}