package main

import (
	mergesingboxconfig "sing2cat_web/MergeSingboxConfig"
	sing2catconfig "sing2cat_web/Sing2catConfig"
	"sync"
)

// func init()  {
// 	sing2catconfig.Set_value("project_dir",sing2catconfig.Get_Sing2cat_dir())
// 	sing2catconfig.Get_logger_Core()
// 	sing2catconfig.Sing2cat_Init()
// }
func main() {
	var jobs sync.WaitGroup
	jobs.Add(1)
	go func() {	
		defer jobs.Done()	
		sing2catconfig.Set_value("project_dir",sing2catconfig.Get_Sing2cat_dir())
		sing2catconfig.Get_logger_Core()	
		defer sing2catconfig.Sing2cat_logger.Sync()
		sing2catconfig.Sing2cat_Init()
		if err := mergesingboxconfig.GenerateConfigJson();err != nil{
			sing2catconfig.Logger_caller("Generate Config failed!",err)
		}else{
			sing2catconfig.Logger_caller("Generate Config success!",err)
		}
	}()
	jobs.Wait()
}