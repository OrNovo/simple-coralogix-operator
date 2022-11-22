package main

import "log"

//////
func main() {
	operator, err := NewCoralogixOperator()
	if err != nil {
		log.Fatal(err.Error())
	}
	operator.Run()
}
