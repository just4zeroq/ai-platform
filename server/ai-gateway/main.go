package main

import (
	_ "ai-gateway/internal/packed"

	"github.com/gogf/gf/v2/os/gctx"

	"ai-gateway/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}
