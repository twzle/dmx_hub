package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"

	"git.miem.hse.ru/hubman/dmx-executor/config"
	"git.miem.hse.ru/hubman/dmx-executor/internal"

	"git.miem.hse.ru/hubman/hubman-lib"
	"git.miem.hse.ru/hubman/hubman-lib/core"
	"git.miem.hse.ru/hubman/hubman-lib/executor"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("error while init config: %v", err)
	}
	ctx := context.Background()

	logger, err := internal.NewLogger()
	if err != nil {
		log.Fatalf("error while init logger: %v", err)
	}

	manager := internal.NewManager(logger)

	app := hubman.NewAgentApp(
		core.AgentConfiguration{
			System: &core.SystemConfig{
				Server: &core.InterfaceConfig{
					IP:   net.ParseIP(cfg.Host),
					Port: cfg.Port,
				},
				RedisUrl: cfg.RedisUrl,
			},
			User: &internal.RefreshConfig{},
			ParseUserConfig: func(jsonBuf []byte) (core.Configuration, error) {
				var conf internal.RefreshConfig
				err := json.Unmarshal(jsonBuf, &conf)
				return &conf, err
			},
		},
		hubman.WithExecutor(
			hubman.WithCommand(internal.SetChannel{}, func(command core.SerializedCommand, parser executor.CommandParser) {
				var cmd internal.SetChannel // json-like api
				parser(&cmd)                // enriches your command with data from redis

				err := manager.ProcessSetChannel(ctx, cmd)
				if err != nil {
					logger.Error(fmt.Sprintf("error while execute move command: %v", err))
				}
			}),
			hubman.WithCommand(internal.Blackout{}, func(command core.SerializedCommand, parser executor.CommandParser) {
				var cmd internal.Blackout // json-like api
				parser(&cmd)              // enriches your command with data from redis

				err := manager.ProcessBlackout(ctx, cmd)
				if err != nil {
					logger.Error(fmt.Sprintf("error while execute update speed command: %v", err))
				}
			}),
		),
		hubman.WithOnConfigRefresh(func(configuration core.AgentConfiguration) {
			update, ok := configuration.User.(*internal.RefreshConfig)
			if !ok {
				panic(
					fmt.Sprintf(
						"Refresh config error: expected type %T, received %T",
						internal.RefreshConfig{},
						configuration.User,
					),
				)
			}
			manager.UpdateDMXDevices(ctx, update.DMXDevices)
		}),
	)

	<-app.WaitShutdown()
	os.Exit(0)
}
