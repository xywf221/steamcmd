package main

import (
	"fmt"
	"github.com/xywf221/steamcmd"
)

func main() {
	cmd, err := steamcmd.NewSteamCmd("/root/steam")
	if err != nil {
		panic(err)
	}
	fmt.Println(cmd.Run())
}
