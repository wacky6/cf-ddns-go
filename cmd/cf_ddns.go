package main

import (
	"fmt"

	"../util"
)

func main() {
	result, err := util.DetectByInterface(util.IP4, []string{})
	fmt.Printf("err = %v, IP4 = %v\n", err, result)

	result, err = util.DetectByInterface(util.IP6, []string{})
	fmt.Printf("err = %v, IP6 = %v\n", err, result)
}
