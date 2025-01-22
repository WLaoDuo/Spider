package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	"github.com/tidwall/gjson"
)

var (
	ErrFalse   = errors.New("false")
	ErrNotExit = errors.New("文件不存在")
)

// 定义 Cookie 结构
type Cookie struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Path     string `json:"path"`
	Domain   string `json:"domain"`
	HttpOnly bool   `json:"httpOnly"`
	Secure   bool   `json:"secure"`
	// Expiry   int64  `json:"expires"`
}

func exits(filepath string) bool {
	// 获取文件状态信息
	_, err := os.Stat(filepath)

	// 根据错误判断文件是否存在
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("文件 %s 不存在。\n", filepath)
		} else {
			// 其他错误，如权限不足等
			fmt.Printf("获取文件 %s 状态时发生错误：%v\n", filepath, err)
		}
		return false
	} else {
		// fmt.Printf("文件 %s 存在。\n", filepath)
		return true
	}
}

func checkFile(driver_path, browser_path string) bool {
	if driver_path == "" {
		// 获取当前工作目录
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Println("获取当前工作目录失败:", err)
			return false
		}

		// 拼接文件名
		driver_path = filepath.Join(cwd, "chromedriver.exe")
	}

	if !exits(driver_path) || !exits(browser_path) {
		return false
	} else {
		return true
	}
}

func readfile(json_path string) (string, error) {
	if !exits(json_path) {
		return "", ErrNotExit
	}
	content, err := ioutil.ReadFile(json_path)
	if err != nil {
		fmt.Printf("无法读取文件: %v\n", err)
		return "", err
	}

	// 将文件内容转换为字符串
	jsonData := string(content)
	// fmt.Println(jsonData)
	return jsonData, nil
}

func getcookie(cookie_path2 string) []*selenium.Cookie {
	// 创建一个切片来存储 Cookie
	var cookies []*selenium.Cookie

	var name, cookieValue, path, domain string
	var secure bool
	var expires uint
	if !exits(cookie_path2) {
		return nil
	}
	data, _ := readfile(cookie_path2)
	if data == "" {
		return nil
	}
	len0 := gjson.Get(data, "#").Int()
	for i := 0; i <= int(len0); i++ {
		result := gjson.Get(data, strconv.Itoa(i))
		result.ForEach(func(key, value gjson.Result) bool {
			switch {
			case key.String() == "name":
				name = value.String()
			case key.String() == "value":
				cookieValue = value.String()
			case key.String() == "path":
				path = value.String()
			case key.String() == "domain":
				domain = value.String()
			// case key.String() == "httpOnly":
			// 	httpOnly = value.Bool()
			case key.String() == "secure":
				secure = value.Bool()
			case key.String() == "expirationDate":
				expires = uint(value.Float())
			}
			return true
		})
		// 将 Cookie 转换为 selenium.Cookie 并添加到切片中
		// _ = httpOnly
		seleniumCookie := &selenium.Cookie{
			Name:   name,
			Value:  cookieValue,
			Path:   path,
			Domain: domain,
			// HttpOnly: httpOnly,
			Secure: secure,
			Expiry: expires,
		}
		cookies = append(cookies, seleniumCookie)
	}
	return cookies
}

func web_spider(driver_path, browser_path, cookie_path string) (string, error) {
	if !checkFile(driver_path, browser_path) || !exits(cookie_path) {
		fmt.Println(ErrNotExit)
		return "", ErrNotExit
	} else {
		fmt.Println("开启浏览器")
	}

	// chrome设置
	opts := []selenium.ServiceOption{
		// selenium.Output(os.Stderr), // Output debug information to STDERR.
	}
	caps := selenium.Capabilities{
		"browserName":  "chrome",
		"chromeBinary": browser_path,
	}

	// 禁止加载图片，加快渲染速度
	imagCaps := map[string]interface{}{
		"profile.managed_default_content_settings.images": 2,
	}

	chromeCaps := chrome.Capabilities{
		Prefs: imagCaps,
		Path:  "",
		Args: []string{
			// "--headless", // 设置Chrome无头模式，在linux下运行，需要设置这个参数，否则会报错
			"--no-sandbox",
			"--user-agent=Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:126.0) Gecko/20100101 Firefox/126.1",
			"--disable-blink-features=AutomationControlled", // 从 Chrome 88 开始，它的 V8 引擎升级了，加了这个参数，window.navigator.driver=false
			// "--proxy-server=socks5://127.0.0.1:7890",
		},
	}

	caps.AddChrome(chromeCaps)

	// 获取一个随机可用端口
	// listener, err := net.Listen("tcp", ":0")
	// if err != nil {
	// 	panic(err)
	// }

	// port := listener.Addr().(*net.TCPAddr).Port
	// listener.Close()
	port := 18125

	service, err := selenium.NewChromeDriverService(driver_path, port, opts...)
	if err != nil {
		return "", err
	}
	defer service.Stop()

	driver, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", port))
	if err != nil {
		return "", err
	}
	defer driver.Quit()

	url := "https://quote.eastmoney.com/zixuan/"
	driver.Get(url)

	// 添加多个 Cookie
	cookies := getcookie(cookie_path)

	for _, cookie := range cookies {
		err := driver.AddCookie(cookie)
		if err != nil {
			fmt.Println("添加 Cookie 失败:", err)
			return "", err
		}
	}
	// time.Sleep(1 * time.Second)
	driver.Refresh()
	time.Sleep(1 * time.Second)

	element, _ := driver.FindElement(selenium.ByXPATH, "//li[contains(text(),'资金流向')]")
	element.Click()
	time.Sleep(1 * time.Second)

	scrollbar, _ := driver.FindElement(selenium.ByXPATH, "//tbody/tr[1]/td[10]/span[1]")
	scrollbar.Click()
	time.Sleep(1 * time.Second)

	for i := 0; i <= 30; i++ {
		driver.KeyDown(selenium.PageDownKey)
		time.Sleep(500 * time.Millisecond)
	}
	time.Sleep(1 * time.Second)
	// element, err = driver.FindElement(selenium.ByXPATH, "//div[@class='ui-droppable']/table/tbody")
	element, err = driver.FindElement(selenium.ByXPATH, "//div[@class='ui-droppable']/table")
	if err != nil {
		return "", err
	}
	text, err := element.GetAttribute("outerHTML")
	if err != nil {
		return "", err
	}
	driver.Quit()
	return text, nil
}

func write_txt(shuru, filename string) error {

	filePath2, _ := os.Getwd()
	filePath := filepath.Join(filePath2, filename+".txt")
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Println("文件打开失败", err)
		return err
	}
	//及时关闭file句柄
	defer file.Close()
	//写入文件时，使用带缓存的 *Writer
	write := bufio.NewWriter(file)
	write.WriteString(shuru)
	//Flush将缓存的文件真正写入到文件中
	write.Flush()
	fmt.Println("成功写入", filePath)
	return nil
}

func get_data(shuru string) []map[string]string {
	html, _ := readfile(shuru)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		log.Fatal(err)
	}

	var tableTitle []string
	doc.Find("thead tr th").Each(func(i int, cell *goquery.Selection) {
		if len(cell.Text()) != 0 {
			tableTitle = append(tableTitle, cell.Text())
		}
	})

	// 创建二维切片来存储表格数据
	// var tableData [][]string
	var tableData []map[string]string
	doc.Find("tbody tr").Each(func(i int, s *goquery.Selection) {
		// var rowData []string
		var testMap = make(map[string]string) // 声明并 make 初始化 map
		// 遍历每行中的每一列
		s.Find("td").Each(func(j int, cell *goquery.Selection) {
			if j != len(tableTitle) {
				// rowData = append(rowData, cell.Text())
				testMap[tableTitle[j]] = cell.Text()
			}
		})
		tableData = append(tableData, testMap)
		// for key, value := range tableTitle {
		// 	// fmt.Println(key, value, rowData[key])
		// 	testMap[value] = rowData[key]
		// }
		// fmt.Println(testMap)

		// 将每一行的数据添加到二维切片中
		// tableData = append(tableData, rowData)
	})

	// 打印二维切片中的数据
	// for _, row := range tableData {
	// 	fmt.Println(row)
	// }

	return tableData
}
func main() {
	cookie_path, _ := os.Getwd()
	cookie_path2 := filepath.Join(cookie_path, "cookie.txt")
	// data, err := (web_spider("D:/study/daima/pachong/chromedriver.exe", "C:/Program Files/Google/Chrome/Application/chrome.exe", cookie_path2))
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// write_txt(data, "test")

	cookie_path2 = filepath.Join(cookie_path, "test.txt.html")
	data := get_data(cookie_path2)
	fmt.Print("len(data)=", len(data), "\tlen(data)[0]=", len(data[0]), data[0], data[0]["小单净占比"])
}
