package webservice

import (
	"fmt"
	"net/url"
	"testing"
	"time"
)

func TestInit(t *testing.T) {
	fmt.Println(time.Now().Format("2006-01-02 15:04:05.999:"), "start unittests for webservice...")
}

func TestUrlParsing(t *testing.T) {
	fmt.Println(time.Now().Format("2006-01-02 15:04:05.999:"), "start unittest for url parsing...")
	u, _ := url.Parse("http://www.baidu.com/ps?key=val")
	u.Host = "google.com"
	u.Scheme = "https"
	fmt.Println(time.Now().Format("2006-01-02 15:04:05.999:"), "new url:", u.String())
}

func TestDown(t *testing.T) {
	fmt.Println(time.Now().Format("2006-01-02 15:04:05.999:"), "unittests for webservice are done.")
}
