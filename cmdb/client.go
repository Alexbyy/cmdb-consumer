package cmdb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"inspur.com/cmdb-consumer/options"
	"io/ioutil"
	"k8s.io/klog/v2"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	BaseUrl     string
	CookieStr   string
	ContentType string
	HttpClient  *http.Client
	objs        []map[string]string
}

//用于使用条件查询某个实例
type Condition struct {
	Condition map[string]interface{} `json:"condition"`
}
type InstCondition struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}


func Login(opts *options.Options) string{
	//user := os.Getenv("SECRET_USERNAME")
	//pass := os.Getenv("SECRET_PASSWORD")
	user := "admin"
	pass := "QY62w7QGwJ"
	str := fmt.Sprintf("username=%s&password=%s", user, pass)
	httpClientTrans := &http.Transport{}
	cookieStr := ""
	//登录获取token

	req, err := http.NewRequest("POST", opts.CmdbBaseUrl + "/login",
		strings.NewReader(str))
	if err != nil {
		klog.Fatalf("登录cmdb失败1：%v\n", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", "32")
	resp, err := httpClientTrans.RoundTrip(req)
	for err != nil {
		klog.Errorf("登录cmdb失败3, 二十秒后重试：%v\n", err)
		time.Sleep(time.Duration(20) * time.Second)
		resp, err = httpClientTrans.RoundTrip(req)

	}
	fmt.Printf("resp cookie:%v\n", resp)
	defer resp.Body.Close()

	// 拼接cookie
	for _, cookie := range resp.Cookies() {
		cookieStr += cookie.Name + "=" + cookie.Value + ";"
	}
	return cookieStr
}

func NewClient(opts *options.Options) *Client {

	httpClient := &http.Client{}
	newClient := &Client{
		BaseUrl:     opts.CmdbBaseUrl,
		ContentType: "application/json;charset=UTF-8",
		CookieStr:   "",
		HttpClient:  httpClient,
	}
	return newClient
}

func (c *Client) AddInstance(method string, url string, body interface{}) (map[string]interface{}, error) {
	ms, err := json.Marshal(body)
	if err != nil {
		fmt.Errorf("json 编译错误：%v\n", err)
	}
	payload := bytes.NewBuffer([]byte(ms))
	url = c.BaseUrl + "/api/v3/create/instance/object/" + url
	req, err := http.NewRequest(method, url, payload)
	req.Header.Set("Content-Type", c.ContentType)
	req.Header.Set("Cookie", c.CookieStr)
	resp, err := c.HttpClient.Do(req)
	defer resp.Body.Close()
	res, err := ParseResponse(resp)
	if err != nil {
		// handle error
		fmt.Errorf("请求错误：%v\n", err)
	}
	//_ = PrintJson(res)
	return res, nil
}


func (c *Client) DelInstance(objId string, instId string) (map[string]interface{}, error) {

	url := c.BaseUrl + "/api/v3/delete/instance/object/" + objId + "/inst/" + instId
	fmt.Printf("url: %s\n", url)
	req, err := http.NewRequest("DELETE", url, nil)
	req.Header.Set("Content-Type", c.ContentType)
	req.Header.Set("Cookie", c.CookieStr)
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	res, err := ParseResponse(resp)
	if err != nil {
		return nil, err
	}
	return res, nil

}

//Operator: 取值为：$regex $eq $ne
func (c *Client) GetInstance(objId string, body *Condition) (map[string]interface{}, error) {
	ms, err := json.Marshal(*body)
	if err != nil {
		return nil, err
	}
	payload := bytes.NewBuffer([]byte(ms))
	url := c.BaseUrl + "/api/v3/find/instassociation/object/" + objId
	req, err := http.NewRequest("POST", url, payload)
	req.Header.Set("Content-Type", c.ContentType)
	req.Header.Set("Cookie", c.CookieStr)
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	res, err := ParseResponse(resp)
	if err != nil {
		return nil, err
	}
	return res, nil

}


func ParseResponse(r *http.Response) (map[string]interface{}, error) {
	var result map[string]interface{}
	body, err := ioutil.ReadAll(r.Body)
	if err == nil {
		err = json.Unmarshal(body, &result)
	}

	return result, err
}
