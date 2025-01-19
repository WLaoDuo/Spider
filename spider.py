from selenium import webdriver
import sys,time,pandas
import argparse
from pathlib import Path
from bs4 import BeautifulSoup
import struct,json
from datetime import datetime
from selenium.webdriver.common.action_chains import ActionChains
from selenium.webdriver.common.keys import Keys
#from selenium.webdriver.firefox.options import Options
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.chrome.service import Service
from selenium.webdriver.common.by import By


def web_spider(driver_path,browser_path,cookie_path):
    options = Options()
    # 设置ua头
    custom_user_agent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
    options.add_argument(f'--user-agent={custom_user_agent}')
    options.add_argument("--headless") #无头模式，可隐藏浏览器界面

    service = Service(driver_path)
    options.binary_location = browser_path  # 设置 浏览器的路径
    if 'chromedriver.exe' in str(driver_path):
        driver = webdriver.Chrome(options=options,service=service)
    if "geckodriver.exe" in str(driver_path):
        driver=webdriver.Firefox(options=options,service=service)
    

    # 需要访问的地址
    driver.get("https://quote.eastmoney.com/zixuan/")

    # 加载 cookies 需要替换字段为"sameSite": "None",
    with open(cookie_path, "r") as f:
        cookies = json.load(f)
        for cookie in cookies:
            driver.add_cookie(cookie)
        time.sleep(1)
    driver.refresh()
    time.sleep(1)

    #点击 "资金流向"
    dianji1 = driver.find_element(By.XPATH, "//li[contains(text(),'资金流向')]")
    dianji1.click()
    time.sleep(2)

    #向下滚动，加载所有数据
    actions = ActionChains(driver)
    scrollbar = driver.find_element(By.XPATH, "//tbody/tr[1]/td[10]/span[1]")
    scrollbar.click()
    time.sleep(2)
    for i in range(40):
        actions.send_keys(Keys.PAGE_DOWN).perform()
        time.sleep(0.5)
    actions.send_keys(Keys.END).perform()
    time.sleep(1)

    #查找表格内容
    data=driver.find_element(By.XPATH, "//div[@class='ui-droppable']/table/tbody")
    result=data.get_attribute('outerHTML')
    driver.quit()    # 运行结束关闭整个浏览器窗口
    return(result)

def get_data(shuru):
    soup = BeautifulSoup(shuru, 'lxml')

    result = []
    for tr in soup.find_all('tr'):
        row_data = {}
        # 提取各个 <td> 中的内容
        data = [td.get_text(strip=True) for td in tr.find_all('td')]
        # print("data:\n",data)
        # 输出结果
        # for index, value in enumerate(data):
        #     row_data[index]=value
        row_data['code']=data[1]
        # row_data['name']=data[2]
        # row_data['股吧']=data[3]
        row_data['最新价']=data[4]
        row_data['涨跌幅']=data[5]
        row_data['主力净流入']=data[6]
        row_data['集合竞价']=data[7]
        row_data['超大单流入']=data[8]
        row_data['超大单流出']=data[9]
        row_data['超大单净额']=data[10]
        row_data['超大单净占比']=data[11]
        row_data['大单流入']=data[12]
        row_data['大单流出']=data[13]
        row_data['大单净额']=data[14]
        row_data['大单净占比']=data[15]
        row_data['中单流入']=data[16]
        row_data['中单流出']=data[17]
        row_data['中单净额']=data[18]
        row_data['中单净占比']=data[19]
        row_data['小单流入']=data[20]
        row_data['小单流出']=data[21]
        row_data['小单净额']=data[22]
        row_data['小单净占比']=data[23]
        row_data['date']=str(time.strftime("%Y-%m-%d", time.localtime()))
        result.append(row_data)
    return result

# def write_txt(shuru,filename):
#     txt_path=Path.cwd()
#     # txt_path=Path("D:\study\daima\go_spider")
#     txt_path=txt_path /  str(time.strftime("%Y-%m-%d", time.localtime())+"_"+filename+".txt")
#     try:
#         with open(txt_path,'w',encoding='utf-8') as f:
#             f.write(shuru)
#         print("成功写入",txt_path)
#     except:
#         print("error!")

def process_value(x):# 使用 applymap 检查并处理所有单元格中的“万”和“亿”
    if isinstance(x, str):
        if '亿' in x:
            return (float(x.replace('亿', '').strip())) * 10000
        elif '万' in x:
            return (x.replace('万', ''))
    return x

# def write_binary(shuru,filename):
#     txt_path=Path.cwd()/"FMLDATA"
#     if not txt_path.exists():
#         txt_path.mkdir ()
#     txt_path=txt_path /  str(filename+".DAY")
#     morning_8am = datetime.now().replace(hour=8, minute=0, second=0, microsecond=0)
#     # now=morning_8am.strftime("%Y-%m-%d %H:%M:%S")
#     with open(txt_path,'wb') as f:
#         # f.write(bytes([0x00, 0x63, 0x80, 0x67]))
#         f.write(struct.pack('<i',int(morning_8am.timestamp())))
#         f.write(struct.pack('<f', float(shuru)))
#     print("DAY二进制文件成功写入",txt_path)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument('browser', type=str, nargs='?', default='C:\Program Files\Google\Chrome\Application\chrome.exe',help='浏览器路径')
    parser.add_argument('cookie', type=str, nargs='?', default=Path.cwd()/'cookie.txt',help='cookie路径')
    # 解析命令行参数
    args = parser.parse_args()


    # browser="C:\Program Files\Google\Chrome\Application\chrome.exe"
    # browser="D:/program/firefox/firefox.exe"
    browser=args.browser
    if "firefox.exe" in browser:
        driver=Path.cwd()/"geckodriver.exe"
    if "chrome.exe" in browser:
        driver=Path.cwd()/'chromedriver.exe'
    if not driver.exists() or not Path(browser).exists() or not Path(args.cookie).exists():
        sys.exit("浏览器或driver路径或cookie.json路径错误，程序退出")

    html=web_spider(driver,browser,args.cookie)
    # write_txt(str(html),"data")
    df = pandas.DataFrame(get_data(html))
    df.replace({"—": "0"}, inplace=True) #清洗空值
    df = df.map(process_value) #去除“万”，“亿”中文并乘10000


    txt_path=Path.cwd()/"spider4"
    if not txt_path.exists():
        txt_path.mkdir ()
    df.to_csv(txt_path/"CDB.TXT", sep='\t', header=False,index=False, columns=['code', 'date','超大单流入'], encoding='utf-8')
    df.to_csv(txt_path/"CDS.TXT", sep='\t', header=False,index=False, columns=['code', 'date','超大单流出'], encoding='utf-8')
    df.to_csv(txt_path/"DB.TXT", sep='\t', header=False,index=False, columns=['code', 'date','大单流入'], encoding='utf-8')
    df.to_csv(txt_path/"DS.TXT", sep='\t', header=False,index=False, columns=['code', 'date','大单流出'], encoding='utf-8')
    # df.to_csv(txt_path/"ZB.TXT", sep='\t', header=False,index=False, columns=['code', 'date','中单流入'], encoding='utf-8')
    # df.to_csv(txt_path/"ZS.TXT", sep='\t', header=False,index=False, columns=['code', 'date','中单流出'], encoding='utf-8')
    # df.to_csv(txt_path/"CDJ.TXT", sep='\t', header=False,index=False, columns=['code', 'date','超大单净占比'], encoding='utf-8')
    # df.to_csv(txt_path/"DJ.TXT", sep='\t', header=False,index=False, columns=['code', 'date','大单净占比'], encoding='utf-8')
    df.to_csv(txt_path/"ZJ.TXT", sep='\t', header=False,index=False, columns=['code', 'date','中单净占比'], encoding='utf-8')
    df.to_csv(txt_path/"XJ.TXT", sep='\t', header=False,index=False, columns=['code', 'date','小单净占比'], encoding='utf-8')
    print("txt成功保存至",txt_path)
