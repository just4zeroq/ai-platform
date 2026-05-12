package main

import (
	_ "asset-svc/internal/packed"

	"github.com/gogf/gf/v2/os/gctx"

	"asset-svc/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}
