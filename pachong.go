package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
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
			"--headless", // 设置Chrome无头模式，在linux下运行，需要设置这个参数，否则会报错
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
	time.Sleep(500 * time.Millisecond)

	element, _ := driver.FindElement(selenium.ByXPATH, "//li[contains(text(),'我的自选')]")
	// element.Click()
	// time.Sleep(1 * time.Second)

	scrollbar, _ := driver.FindElement(selenium.ByXPATH, "//tbody/tr[1]/td[10]/span[1]")
	scrollbar.Click()
	time.Sleep(500 * time.Millisecond)

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

func writeBinary(value float32, filename string) error {
	filePath2, _ := os.Getwd()
	dirPath := filepath.Join(filePath2, "FMLDATA")
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		err := os.Mkdir(dirPath, os.ModePerm)
		if err != nil {
			fmt.Println("无法创建目录:", err)
			return err
		}
	}

	// 2. 生成完整文件路径
	filePath := filepath.Join(dirPath, filename+".DAY")

	// 3. 生成当天上午8点时间戳
	timestamp, err := getToday8AMTimestamp()
	if err != nil {
		return fmt.Errorf("时间计算错误: %v", err)
	}

	// 4. 创建并写入二进制文件
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0755) //覆盖写
	if err != nil {
		return fmt.Errorf("文件创建失败: %v", err)
	}
	defer file.Close()

	// 写入时间戳（小端32位整数）
	if err := binary.Write(file, binary.LittleEndian, int32(timestamp)); err != nil {
		return fmt.Errorf("写入时间戳失败: %v", err)
	}

	// 写入数值（小端单精度浮点数）
	if err := binary.Write(file, binary.LittleEndian, value); err != nil {
		return fmt.Errorf("写入数值失败: %v", err)
	}
	// fmt.Println("DAY二进制文件成功写入: ", filePath)
	return nil
}

// getToday8AMTimestamp 获取当天上午8点的Unix时间戳
func getToday8AMTimestamp() (int64, error) {
	now := time.Now()

	// 构造当天8点时间
	eightAM := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		8, // 小时
		0, // 分钟
		0, // 秒
		0, // 纳秒
		now.Location(),
	)

	// 处理跨天情况（如果当前时间早于8点）
	if now.Before(eightAM) {
		eightAM = eightAM.Add(-24 * time.Hour)
	}

	return eightAM.Unix(), nil
}

func get_data(shuru string) []map[string]string {
	// html, _ := readfile(shuru)
	html := shuru

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

func extractNumberEnhanced(input string) float32 {
	if input == "-" {
		return float32(0)
	}
	re := regexp.MustCompile(`([+-]?)(\d+\.?\d*)([万亿]?)`)
	matches := re.FindStringSubmatch(input)
	if len(matches) < 4 {
		return float32(0)
	}

	sign := matches[1]
	number := matches[2]
	unit := matches[3]

	// 单位转换
	multiplier := 1.0
	switch unit {
	case "万":
		multiplier = 1
	case "亿":
		// multiplier = 100000000
		multiplier = 10000
	}

	// 转换为浮点数并应用单位
	value, err := strconv.ParseFloat(sign+number, 64)
	if err != nil {
		return float32(0)
	}
	return float32(value * multiplier)
}
func getBrowserPath() string {
	if path := os.Getenv("CHROME_PATH"); path != "" {
		return path
	}
	switch runtime.GOOS {
	case "windows":
		return `C:\Program Files\Google\Chrome\Application\chrome.exe`
	case "darwin":
		return "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
	default:
		return "/usr/bin/google-chrome"
	}
}
func main() {
	workDir, _ := os.Getwd()
	cookie_path2 := filepath.Join(workDir, "cookies.json.txt")
	driver := filepath.Join(workDir, "chromedriver.exe")

	html, err := web_spider(driver, getBrowserPath(), cookie_path2)
	if err != nil {
		fmt.Println(err)
		return
	}
	// write_txt(html, "test")

	data2 := get_data(html)
	// fmt.Println("len(data)=", len(data2), "\tlen(data)[0]=", len(data2[0]), data2[0], data2[0]["小单净占比"])

	for _, row := range data2 {
		// fmt.Printf("--- 第 %d 行 ---\n", rowIndex)
		// for key, value := range row {
		// 	fmt.Println("字段", key, ":", value, "转换=", extractNumberEnhanced(value))
		// }
		err := writeBinary(extractNumberEnhanced(row["超大单流入"]), row["代码"]+".1")
		err = writeBinary(extractNumberEnhanced(row["大单流入"]), row["代码"]+".2")
		err = writeBinary(extractNumberEnhanced(row["超大单流出"]), row["代码"]+".3")
		err = writeBinary(extractNumberEnhanced(row["大单流出"]), row["代码"]+".4")
		err = writeBinary(extractNumberEnhanced(row["中单净占比"]), row["代码"]+".5")
		err = writeBinary(extractNumberEnhanced(row["小单净占比"]), row["代码"]+".6")
		err = writeBinary(extractNumberEnhanced(row["当日DDX"]), row["代码"]+".7")
		err = writeBinary(extractNumberEnhanced(row["当日DDY"]), row["代码"]+".8")
		err = writeBinary(extractNumberEnhanced(row["当日DDZ"]), row["代码"]+".9")
		if err != nil {
			fmt.Println("发生错误:", err)
			return
		}
	}
	fmt.Println("成功保存至/FMLDATA目录")
}
