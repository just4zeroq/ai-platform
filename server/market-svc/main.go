package main

import (
	_ "market-svc/internal/packed"

	"github.com/gogf/gf/v2/os/gctx"

	"market-svc/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}
