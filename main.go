package main

import (
	_ "github.com/suxiangdong/indu-scope-backend/internal/logic"

	"context"

	"github.com/suxiangdong/indu-scope-backend/internal/cmd"
)

func main() {
	cmd.Main.Run(context.Background())
}
