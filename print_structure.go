package goutils

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// PrintStruct can print any data structure like json format
func PrintStruct(in interface{}) {
	b, err := json.Marshal(in)
	if err != nil {
		fmt.Printf("%+v", in)
		fmt.Println()
		return
	}
	var out bytes.Buffer
	err = json.Indent(&out, b, "", "    ")
	if err != nil {
		fmt.Printf("%+v", in)
		fmt.Println()
		return
	}
	fmt.Println(out.String())
}
