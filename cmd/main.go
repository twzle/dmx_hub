package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"git.miem.hse.ru/hubman/dmx-executor/config"
	"git.miem.hse.ru/hubman/dmx-executor/internal"

	"git.miem.hse.ru/hubman/hubman-lib"
	"git.miem.hse.ru/hubman/hubman-lib/core"
	"git.miem.hse.ru/hubman/hubman-lib/executor"
)

func main() {
	systemConfig := &core.SystemConfig{}
	userConfig := &config.UserConfig{}

	err := core.ReadConfig(systemConfig, userConfig)
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
			System:          systemConfig,
			User:            userConfig,
			ParseUserConfig: func(data []byte) (core.Configuration, error) { return config.ParseConfigFromBytes(data) },
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
			update, ok := configuration.User.(*config.UserConfig)
			if !ok {
				panic(
					fmt.Sprintf(
						"Refresh config error: expected type %T, received %T",
						config.UserConfig{},
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
