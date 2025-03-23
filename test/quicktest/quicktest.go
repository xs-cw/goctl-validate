package main

import (
	"flag"
	"fmt"

	"github.com/linabellbiu/goctl-validate/test/quicktest/internal/config"
	"github.com/linabellbiu/goctl-validate/test/quicktest/internal/handler"
	"github.com/linabellbiu/goctl-validate/test/quicktest/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/quicktest-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
