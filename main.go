package main

import (
	"encoding/json"
	"fmt"
	"inspur.com/cmdb-consumer/cmdb"
	"inspur.com/cmdb-consumer/options"
	"io/ioutil"
	"k8s.io/klog/v2"
	"net/http"
	"strconv"
)

func main()  {

	klog.InitFlags(nil)
	defer klog.Flush()
	opts := options.NewOptions()
	klog.Infof("启动参数：%v\n", opts)

	cookieStr := cmdb.Login(opts)
	cmdbClient := cmdb.NewClient(opts)
	cmdbClient.CookieStr = cookieStr
	eh := eventsHandler{
		cmdb:    cmdbClient,
	}


	mux := http.NewServeMux()
	mux.HandleFunc("/", index)
	mux.Handle("/cmdb/v1/actions/events", &eh)

	server := &http.Server{
		Addr: "0.0.0.0:8080",
		Handler: mux,
	}
	server.ListenAndServe()
}

func index(w http.ResponseWriter, r *http.Request)  {
	fmt.Fprintf(w, "Hello, world!, %s!", r.URL.Path[1:])
}

type eventsHandler struct {
	cmdb    *cmdb.Client
}

func (ih *eventsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	productcode := r.Header.Get("ce-productcode")
	operation := r.Header.Get("ce-operation")

	klog.V(4).Infof("获取到事件的productcode：%v\n", productcode)
	klog.V(4).Infof("事件类型：%s\n", operation)


	res, err := ParseResponse(r)
	if err != nil {
		klog.Errorf("ParseResponse: %v\n", err)
		return
	}
	if _, ok := res["ref"]; ok {
		delete(res, "ref")
	}
	res["bk_inst_name"] = "test"
	fmt.Printf("body:%v\n", r)
	fmt.Printf("res:%v\n", res["account"])

	if operation == "created"{
		resp, err := ih.cmdb.AddInstance("POST", productcode, res)
		if err != nil {
			klog.Errorf("AddInstance: %v\n", err)
			return

		}
		fmt.Printf("cmdb resp: %v\n", resp)
	}
	if operation == "deleted"{
		instCon := cmdb.InstCondition{
			Field:    "id",
			Operator: "$eq",
			Value:    res["id"].(string),
		}
		var temp []cmdb.InstCondition
		temp = append(temp, instCon)
		condition := cmdb.Condition{
			Condition: map[string]interface{}{
				productcode: temp,
			},
		}

		res, err := ih.cmdb.GetInstance(productcode, &condition)
		if err != nil{
			klog.Errorf("GetInstance: %v\n", err)
			return
		}
		if res["bk_error_msg"].(string) != "success" {
			klog.Errorf("getInstance: %v\n", res["bk_error_msg"])
			return
		}
		data := res["data"].(map[string]interface{})
		for _, item := range data["info"].([]interface{}) {
			bkInstId := int(item.(map[string]interface{})["bk_inst_id"].(float64))
			id := strconv.Itoa(bkInstId)
			res1, err := ih.cmdb.DelInstance(productcode, id)
			if err != nil{
				klog.Errorf("DelInstance: %v\n", err)
				return
			}
			klog.V(4).Infof("删除实例结果：%v,\n", res1)
		}

	}



}

func ParseResponse(r *http.Request) (map[string]interface{}, error) {
	var result map[string]interface{}
	body, err := ioutil.ReadAll(r.Body)
	if err == nil {
		err = json.Unmarshal(body, &result)
	}

	return result, err
}
