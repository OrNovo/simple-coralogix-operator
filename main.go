package main

import (
	"github.com/golang/glog"
)

func main() {
	operator, err := NewCoralogixOperator()
	if err != nil {
		glog.Errorln(err.Error())
	}
	operator.Run()
}
