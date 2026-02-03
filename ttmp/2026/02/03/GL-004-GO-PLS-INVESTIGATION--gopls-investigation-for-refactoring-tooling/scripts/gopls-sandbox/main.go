package main

import (
	"fmt"

	"example.com/gopls-sandbox/lib"
)

func main() {
	widget := lib.NewWidget("alpha")
	if err := lib.UseWidget(widget); err != nil {
		fmt.Println("run error:", err)
		return
	}
	runner := lib.WrapWidget(widget)
	_ = runner
	fmt.Println(widget.String())
}
