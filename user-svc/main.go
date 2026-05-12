package main

import (
	_ "user-svc/internal/packed"

	"github.com/gogf/gf/v2/os/gctx"

	"user-svc/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}
