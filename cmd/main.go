package main

import (
	"fmt"
	"time"

	"github.com/sigrdrifa/go-htmx-websockets-example/internal/hardware"
	"github.com/sigrdrifa/go-htmx-websockets-example/internal/utils"
)

func main() {
	fmt.Print("Starting system monitor..\n\n")

	go func() {
		for {
			systemSection, err := hardware.GetSystemSection()
			utils.ThrowOnError("systemSection", err)

			diskSection, err := hardware.GetDiskSection()
			utils.ThrowOnError("diskSection", err)

			cpuSection, err := hardware.GetCpuSection()
			utils.ThrowOnError("cpuSection", err)

			fmt.Println(systemSection)
			fmt.Println(diskSection)
			fmt.Println(cpuSection)
			fmt.Println()
			fmt.Println()
			time.Sleep(time.Second * 3)
		}
	}()

	time.Sleep(time.Minute * 5)

}
