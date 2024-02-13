package merge

import (
	"os"
	"path/filepath"
	initial "sing2catweb/Initial"
	"sync"

	"github.com/bitly/go-simplejson"
)

func GenerateConfigJson() {
	outbound_channel := make(chan *simplejson.Json,50)
	outbounds := []*simplejson.Json{}
	var jobs sync.WaitGroup
	jobs.Add(1)
	// 获取固定信息
	log := initial.GetValue("template").(*simplejson.Json).Get("log")
	dns := initial.GetValue("template").(*simplejson.Json).Get("dns")
	inbounds := initial.GetValue("template").(*simplejson.Json).Get("inbounds")
	experimental := initial.GetValue("template").(*simplejson.Json).Get("experimental")
	// 获取会变化的信息,出站和路由
	route := MergeRoute()
	go func(){
		outbounds := MergeOutbounds()
		for _,outbound := range(outbounds){
			outbound_channel <- outbound
		}
		
		defer jobs.Done()
		defer close(outbound_channel)
	}()
	// 设置json
	config := simplejson.New()
	config.Set("log",log)
	config.Set("dns",dns)
	config.Set("inbounds",inbounds)
	config.Set("route",route)
	config.Set("experimental",experimental)
	
	for outbound := range(outbound_channel){
		outbounds = append(outbounds, outbound)
	}
	jobs.Wait()

	config.Set("outbounds",outbounds)

	config_file,_ := config.EncodePretty()
	// 写入文件
	dst_dir := filepath.Join(initial.GetValue("base_dir").(string),"/temp/config.json")

	_,err := os.Stat(dst_dir)
	if err != nil{
		if os.IsNotExist(err){
			os.Remove(dst_dir)
		}
	}else{
		os.Remove(dst_dir)
	}
	file ,_ := os.OpenFile(dst_dir,os.O_CREATE|os.O_RDWR,0777)
	
	file.WriteString(string(config_file))
	defer file.Close()
}