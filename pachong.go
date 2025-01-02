package main

import (
	"fmt"
	"net"
	"time"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

func web_spider(driver_path, browser_path string) (string, error) {
	url := "https://xz.aliyun.com/feed"

	opts := []selenium.ServiceOption{}
	caps := selenium.Capabilities{
		"browserName":  "chrome",
		"chromeBinary": browser_path,
	}

	// 禁止加载图片，加快渲染速度
	imagCaps := map[string]interface{}{
		"profile.managed_default_content_settings.images": 2,
	}

	args := []string{
		// "--headless",
		"--no-sandbox",
		"--user-agent=Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.55 Safari/537.36",
	}
	// if config.Cfg.Proxy.CrawlerProxyEnabled {
	// 	// 设置代理
	// 	proxyArgs := fmt.Sprintf("--proxy-server=%s", config.Cfg.Proxy.ProxyUrl)
	// 	args = append(args, proxyArgs)
	// }
	chromeCaps := chrome.Capabilities{
		Prefs: imagCaps,
		Path:  "",
		Args:  args,
	}

	caps.AddChrome(chromeCaps)

	// 获取一个随机可用端口
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	service, err := selenium.NewChromeDriverService(driver_path, port, opts...)
	if err != nil {
		return "", err
	}
	defer service.Stop()

	webDriver, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", port))
	if err != nil {
		return "", err
	}
	defer webDriver.Quit()

	webDriver.AddCookie(&selenium.Cookie{
		Name:  "defaultJumpDomain",
		Value: "www",
	})

	err = webDriver.Get(url)
	if err != nil {
		return "", err
	}

	element, err := webDriver.FindElement(selenium.ByXPATH, "/html/body/pre")
	if err != nil {
		return "", err
	}
	text, err := element.Text()
	if err != nil {
		return "", err
	}

	time.Sleep(3 * time.Second)
	// fmt.Print(text)
	return text, nil
}

func main() {
	web_spider("D:/study/daima/python3_spider/chromedriver.exe", "C:/Program Files/Google/Chrome/Application/chrome.exe")
}
