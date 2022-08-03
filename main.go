package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/spf13/viper"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
)

var (
	url        = "https://myip.ipip.net/"
	ddns       = "127.0.0.1"
	keyid      string
	keyck      string
	domain     string
	RecordType string
	RecordLine string
	TTL        uint64
	RecordId   uint64
	api        string
)

func main() {
	pz()
	fmt.Println("开始执行DDNS监控!")
	ticker := time.NewTicker(60 * time.Second) //循环查询ip时间
	for {
		<-ticker.C
		aa := iPdz()
		if aa != ddns {
			ddns = aa
			fmt.Println("更改ip:", ddns)
			dnss(ddns)
		}
	}
}

func pz() {
	viper.SetConfigName("ddns") // 文件名 (没有后缀的)
	viper.SetConfigType("ini")  // 文件类型（文件合理的后缀名）
	viper.AddConfigPath("./")   // 文件的目录，支持表达式，也可以增加多个
	err := viper.ReadInConfig() // 对文件进行读取
	if err != nil {             // 读取文件失败
		file, error := os.OpenFile("./ddns.ini", os.O_RDWR|os.O_CREATE, 0766)
		if error != nil {
			fmt.Println(error)
			panic(err)
		}
		fmt.Println(file)
		data := "[api]\napi=\n[apikey]\nkeyid=123\nkeyck=123\n[setdomain]\ndomain=\nRecordType=\nRecordLine=\nTTl=\nRecordId=\n"
		//写入byte的slice数据
		file.Write([]byte(data))
		//写入字符串
		//file.WriteString(data)
		file.Close()
		panic("请配置ini文件")
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("没有找到配置文件")
			panic(err)
		} else {
			fmt.Println("找到配置文件，但产生了另一个错误")
			panic(err)
		}
	}
	// 找到并成功解析了配置文件
	keyid = viper.GetString("apikey.keyid")
	keyck = viper.GetString("apikey.keyck")
	domain = viper.GetString("setdomain.domain")
	RecordType = viper.GetString("setdomain.RecordType")
	RecordLine = viper.GetString("setdomain.RecordLine")
	TTL = viper.GetUint64("setdomain.TTL")
	RecordId = viper.GetUint64("setdomain.RecordId")
	api = viper.GetString("api.api")
}

func iPdz() string {
	res, err := http.Get(url)
	if err != nil {
		//发起ip请求
		ERROR("请求错误\n")
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		//检查到返回数据
		ERROR("返回解析错误\n")
	}
	c := string(body)                                //传递参数
	reg1 := regexp.MustCompile(`\d+\.\d+\.\d+\.\d+`) //设置正则表达式
	if reg1 == nil {                                 //解释失败，返回nil
		ERROR("正则解释失败\n")

	}
	//根据规则提取关键信息
	result1 := reg1.FindAllStringSubmatch(c, -1)
	ba := result1[len(result1)-1]
	baa := ba[len(ba)-1]
	return baa
}

func dnss(ipe string) {
	// 实例化一个认证对象，入参需要传入腾讯云账户secretId，secretKey,此处还需注意密钥对的保密
	// 密钥可前往https://console.cloud.tencent.com/cam/capi网站进行获取
	credential := common.NewCredential(
		keyid,
		keyck,
	)
	// 实例化一个client选项，可选的，没有特殊需求可以跳过
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = api
	// 实例化要请求产品的client对象,clientProfile是可选的
	client, _ := dnspod.NewClient(credential, "", cpf)

	// 实例化一个请求对象,每个接口都会对应一个request对象
	request := dnspod.NewModifyRecordRequest()
	//设置域名处理参数
	request.Domain = common.StringPtr(domain)
	request.RecordType = common.StringPtr(RecordType)
	request.RecordLine = common.StringPtr(RecordLine)
	request.Value = common.StringPtr(ipe)
	request.TTL = common.Uint64Ptr(TTL)
	request.RecordId = common.Uint64Ptr(RecordId)

	// 返回的resp是一个ModifyRecordResponse的实例，与请求对象对应
	response, err := client.ModifyRecord(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		//fmt.Printf("API 错误已返回： %s", err)处理错误返回
		s := err.(*errors.TencentCloudSDKError).Message
		s1 := err.(*errors.TencentCloudSDKError).Code + "。 "
		s2 := err.(*errors.TencentCloudSDKError).RequestId + "!"
		ss := "API 错误已返回[go163:]：" + s + s1 + s2
		ERROR(ss) //写入log
		return
	}
	if err != nil {
		panic(err)
	}
	// 输出json格式的字符串回包
	//fmt.Printf("%s\n", response.ToJsonString())
	logss(response.ToJsonString())
}

func ERROR(msg interface{}) {
	//时间截
	fileName := time.Now().Format("2006-01-02")
	//组合文件名
	fileName += "Error.log"
	//打开文件，并且设置了文件打开的模式
	logFile, _ := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	//多个地方同时写入
	Error := log.New(io.MultiWriter(logFile, os.Stderr),
		"[ERROR]: ", //内容开头
		log.Ldate|log.Ltime|log.Llongfile)

	Error.Printf("%v", msg) //输出
}
func logss(msg interface{}) {
	//时间截
	fileName := time.Now().Format("2006-01-02")
	//组合文件名
	fileName += "修改记录.log"
	//打开文件，并且设置了文件打开的模式
	logFile, _ := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	//多个地方同时写入
	Error := log.New(io.MultiWriter(logFile, os.Stderr),
		"[ok]: ", //内容开头
		log.Ldate|log.Llongfile)

	Error.Printf("%v", msg) //输出
}
