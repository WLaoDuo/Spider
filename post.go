package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/tidwall/gjson"
)

type json1 struct {
	Source  string `json:"source"`
	Version string `json:"version"`
	// Query_area       string `json:"query_area"`
	// Block_list       string `json:"block_list"`
	Add_info         string `json:"add_info"`
	Question         string `json:"question"`
	Perpage          int    `json:"perpage"`
	Page             int    `json:"page"`
	Secondary_intent string `json:"secondary_intent"`
	// Log_info         string `json:"log_info"`
	// Rsh              string `json:"rsh"`
}

func createJSONData(chaxun string) ([]byte, error) {
	out := json1{
		Source:  "Ths_iwencai_Xuangu",
		Version: "2.0",
		// Query_area:       "",
		// Block_list:       "",
		Add_info:         "{\"urp\":{\"scene\":1,\"company\":1,\"business\":1},\"contentType\":\"json\",\"searchInfo\":true}",
		Question:         "酒",
		Perpage:          100,
		Page:             1,
		Secondary_intent: "stock",
		// Log_info: "{\"input_type\":\"click\"}",
		// Rsh: "761544030",
	}
	out.Question = chaxun

	jsonData, err := json.Marshal(out)
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}

func sousuo(url2 string, hexinV string, jsonData []byte) (string, error) {
	req, err := http.NewRequest("POST", url2, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:126.0) Gecko/20100101 Firefox/124.1")
	req.Header.Set("Content-Type", "application/json")

	req.Header.Set("Hexin-V", hexinV)
	// req.Header.Set("Cookie", "other_uid=Ths_iwencai_Xuangu_nchr3bxgleswsknqk048ptptchvdkst7; ta_random_userid=06xguc4qov; cid=362a1b071ecfc9d5347c78eac7df996e1735896530; cid=362a1b071ecfc9d5347c78eac7df996e1735896530; ComputerID=362a1b071ecfc9d5347c78eac7df996e1735896530; WafStatus=0; v=A9OoclAOsIK4u3xWxN35vqFQYlz4iGb8IR2rfoXkL1MpGf0CDVj3mjHsO9CW")

	// 设置代理
	// proxyURL, err := url.Parse("http://127.0.0.1:8080")
	// if err != nil {
	// 	return "", err
	// }

	// // 创建一个自定义的 HTTP 客户端
	// transport := &http.Transport{
	// 	Proxy: http.ProxyURL(proxyURL),
	// }
	// client := &http.Client{Transport: transport}
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	// fmt.Println("Response Status:", resp.Status)
	// fmt.Println("Response Body:", string(body))
	if strings.Contains(string(body), "Nginx forbidden") {
		return "", errors.New("Hexin-V失效")
	}
	return string(body), nil

}

func chulishuju(shuru gjson.Result) (string, error) {
	if shuru.String() == "[]" {
		fmt.Println("未查询到结果")
		return "", errors.New("null data")
	} else {
		a := ""
		fmt.Println("len=", shuru.Get("#"))
		for _, j := range shuru.Array() {
			a = a + (j.Get("code").String() + "\t" + j.Get("股票简称").String() + "\n")
		}
		return a, nil
	}
}

func writeTxt(txtname, data string) error {
	if data == "" {
		return errors.New("null data")
	}
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("获取当前工作目录失败:", err)
		return err
	}
	// 拼接文件名
	txt_path := filepath.Join(cwd, txtname+".txt")

	// f, err := os.Create(txt_path)
	f, err := os.OpenFile(txt_path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		return err
	}
	l, err := f.WriteString(data)
	if err != nil {
		fmt.Println(err)
		f.Close()
		return err
	}
	_ = l
	fmt.Println("成功写入", txt_path)
	err = f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func main() {
	cmd := exec.Command("./Hexin-V.exe") // 使用相对路径或绝对路径

	// 获取标准输出
	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("Command execution failed: %s", err)
	}

	// 打印输出结果
	HexinV := strings.TrimSpace(string(output)) //去除字符串尾部空格，回车
	// fmt.Println("'" + HexinV + "'")

	var chaxun = flag.String("chaxun", "长鑫", "查询关键字")
	flag.Parse()
	// chaxun := "长鑫"
	// chaxun := "jiu"

	url := "https://search.10jqka.com.cn/customized/chart/get-robot-data"

	jsonData, err := createJSONData(*chaxun)
	if err != nil {
		fmt.Println("Error creating JSON data:", err)
	}

	result, err := sousuo(url, HexinV, jsonData)
	if err != nil {
		fmt.Println("Error sending POST request:", err)
	}
	temp := gjson.Get(result, "data.answer.0.txt.0.content.components.0.data.datas")
	length := gjson.Get(result, "data.answer.0.txt.0.content.components.0.data.meta.extra.row_count")
	fmt.Println("总数", length)
	data, err := chulishuju(temp)
	// fmt.Println(data)
	writeTxt(*chaxun, data)
}
