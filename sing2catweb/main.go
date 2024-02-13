package main

import (
	"fmt"
	initial "sing2catweb/Initial"
	"sing2catweb/merge"
)
func main() {
	initial.GetBaseDir()
	initial.LoadConfig(fmt.Sprintf("%s/config/config.yaml",initial.GetValue("base_dir")))
	initial.LoadTemplate(fmt.Sprintf("%s/config/template.json",initial.GetValue("base_dir")))
	
	initial.InitialLogger()
	merge.GenerateConfigJson()
	
}