package main

import (
	"fmt"
	"os/exec"
)

func main() {

	done := make(chan bool, 1)
	go func() {
		cmd := exec.Command("SWAT_abca_150524")
		cmd.Run()
		done <- true

	}()

	fmt.Println("awaiting")

	<-done
	fmt.Println("done")
}
