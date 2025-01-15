from selenium import webdriver
import sys,time
import argparse
from pathlib import Path
from bs4 import BeautifulSoup
import struct
from datetime import datetime
#from selenium.webdriver.firefox.options import Options
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.chrome.service import Service
from selenium.webdriver.common.by import By

def write_binary(shuru,filename):
    txt_path=Path.cwd()/"FMLDATA"
    if not txt_path.exists():
        txt_path.mkdir ()
    txt_path=txt_path /  str(filename+".DAY")
    morning_8am = datetime.now().replace(hour=8, minute=0, second=0, microsecond=0)
    # now=morning_8am.strftime("%Y-%m-%d %H:%M:%S")
    with open(txt_path,'wb') as f:
        # f.write(bytes([0x00, 0x63, 0x80, 0x67]))
        f.write(struct.pack('<i',int(morning_8am.timestamp())))
        f.write(struct.pack('<f', float(shuru)))
    print("DAY二进制文件成功写入",txt_path)

def web_spider(driver_path,browser_path):
    options = Options()
    options.add_argument("--headless")
    service = Service(driver_path)
    options.binary_location = browser_path  # 设置 浏览器的路径
    if 'chromedriver.exe' in str(driver_path):
        driver = webdriver.Chrome(options=options,service=service)
    if "geckodriver.exe" in str(driver_path):
        driver=webdriver.Firefox(options=options,service=service)
    # driver = webdriver.Firefox(options=options)
    # 需要访问的地址
    driver.get("https://emdata.eastmoney.com/appdc/zjlx/index.html")
    
    
    #获取当前窗口的高度
    window_height = driver.execute_script("return window.outerHeight;")
    # 向下滚动两行
    driver.execute_script(f"window.scrollBy(0, {window_height // 3});")
    #回到原来的滚动位置
    # driver.execute_script("window.scrollTo(0, document.body.scrollHeight - %s);" %(window_height//2 ))
    
    #点击 下拉框
    dianji0 = driver.find_element(By.XPATH, "//div[@class='tab-item tab-item_market active']")
    dianji0.click()
    time.sleep(1)
    
    #点击 北交所
    dianji1 = driver.find_element(By.XPATH, "//div[@class='market-item' and text()='北交所']")
    dianji1.click()
    time.sleep(1)

    dianji1 = driver.find_element(By.XPATH, "//div[@class='tab-item' and text()='DDE排名']")
    dianji1.click()
    time.sleep(1)


    for i in range(5):
        # 下滑窗口到底部
        driver.execute_script("window.scrollTo(0, document.body.scrollHeight);")
        flag=driver.find_element(By.XPATH, "//div[@class='stock-more-section']")
        if (flag.text=="没有更多数据"):
            break
        # 等待加载完成，可以自定义等待时间
        time.sleep(2)

    div_element = driver.find_element(By.XPATH, "//div[@class='zjlx-sticky-table-wrapper']")
    data=div_element.find_element(By.XPATH, "//div[@class='stock-zjlx-list-wrapper table-right list-theader']/table/tbody") #查找内容
    return(data.get_attribute('outerHTML'))


    # 输出 <div> 元素的文本内容
    # print("长度:\n", len(div_element.text),"\n内容：\n",div_element.text)

    driver.quit()    # 运行结束关闭整个浏览器窗口

def get_data(shuru):
    soup = BeautifulSoup(shuru, 'lxml')

    # 解析 tbody 中的所有 tr
    result = []
    for tr in soup.find_all('tr'):
        row_data = {}
        row_data['code'] = tr['code']
        for td in tr.find_all('td'):
            field = td['field']  # 获取 field 属性
            value = td.get_text(strip=True)  # 获取文本内容
            

            if field == "f88":
                field ="当日DDX"
            if field == "f89":
                field ="当日DDY"
            if field == "f90":
                field ="当日DDZ"
            if field == "f91":
                field ="5日DDX"
            if field == "f92":
                field ="5日DDY"
            if field == "f94":
                field ="10日DDX"
            if field == "f95":
                field ="10日DDY"
            if field == "f97":
                field ="连续飘红天数"
            if field == "f98":
                field ="5日内飘红天数"
            if field == "f99":
                field ="10日内飘红天数"

            if field == 'f2':
                field = '最新'
            if field == "f3":
                field ="涨幅"
            if field == "f62":
                field ="净流入"
            if field == "f605":
                field ="净流速"

            
            row_data[field] = value  # 存入字典
        
        # 添加代码属性
        result.append(row_data)  # 将每行的数据添加到列表中
    # print(result)
    return result


def write_txt(shuru,filename):
    txt_path=Path.cwd()
    # txt_path=Path("D:\study\daima\go_spider")
    txt_path=txt_path /  str(time.strftime("%Y-%m-%d", time.localtime())+"_"+filename+".txt")
    try:
        with open(txt_path,'w',encoding='utf-8') as f:
            f.write(shuru)
        print("成功写入",txt_path)
    except:
        print("error!")


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument('filename', type=str, nargs='?', default='C:\Program Files\Google\Chrome\Application\chrome.exe',help='浏览器路径')
    # 解析命令行参数
    args = parser.parse_args()


    # browser="C:\Program Files\Google\Chrome\Application\chrome.exe"
    # browser="D:/program/firefox/firefox.exe"
    browser=args.filename

    if "firefox.exe" in browser:
        driver=Path.cwd()/"geckodriver.exe"
    if "chrome.exe" in browser:
        driver=Path.cwd()/'chromedriver.exe'
    if not driver.exists() or not Path(browser).exists():
        sys.exit("浏览器或driver路径错误，程序退出")

    data=web_spider(driver,browser)
    result1=get_data(data)


    xieru1=''
    xieru2=''
    xieru3=''

    for row in result1:
        tmp=row['当日DDX']
        if tmp=="-":
            tmp="0"
        xieru1=xieru1+("BJ"+str(row['code'])+"\t"+time.strftime("%Y-%m-%d", time.localtime()) +"\t"+tmp)+"\n"

        write_binary(tmp,str(row['code'])+".1")

        tmp=row['当日DDY']
        if tmp=="-":
            tmp="0"
        xieru2=xieru2+("BJ"+str(row['code'])+"\t"+time.strftime("%Y-%m-%d", time.localtime()) +"\t"+tmp)+"\n"
        write_binary(tmp,str(row['code'])+".2")

        tmp=row['当日DDZ']
        if tmp=="-":
            tmp="0"
        xieru3=xieru3+("BJ"+str(row['code'])+"\t"+time.strftime("%Y-%m-%d", time.localtime()) +"\t"+tmp)+"\n"
        write_binary(tmp,str(row['code'])+".3")
        
    write_txt(xieru1,"北交所-当日DDX")
    write_txt(xieru2,"北交所-当日DDY")
    write_txt(xieru3,"北交所-当日DDZ")
