package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"git.miem.hse.ru/hubman/dmx-executor/internal"
	"git.miem.hse.ru/hubman/dmx-executor/internal/config"
	"git.miem.hse.ru/hubman/dmx-executor/internal/models"

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


	agentConf := core.AgentConfiguration{
		System:          systemConfig,
		User:            userConfig,
		ParseUserConfig: func(data []byte) (core.Configuration, error) { return config.ParseConfigFromBytes(data) },
	}

	app := core.NewContainer(agentConf.System.Logging)
	logger := app.Logger()

	manager := internal.NewManager(logger)
	signals := manager.GetSignals()

	app.RegisterPlugin(
		hubman.NewAgentPlugin(
			app.Logger(),
			agentConf,
			hubman.WithManipulator(
				hubman.WithSignal[models.SceneChanged](),
				hubman.WithChannel(signals),
			),
			hubman.WithExecutor(
				hubman.WithCommand(models.SetChannel{}, func(command core.SerializedCommand, parser executor.CommandParser) {
					var cmd models.SetChannel // json-like api
					parser(&cmd)                // enriches your command with data from redis

					err := manager.ProcessSetChannel(ctx, cmd)
					if err != nil {
						logger.Error(fmt.Sprintf("error while execute move command: %v", err))
					}
				}),
				hubman.WithCommand(models.Blackout{}, func(command core.SerializedCommand, parser executor.CommandParser) {
					var cmd models.Blackout // json-like api
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
				manager.UpdateDevices(ctx, *update)
			}),
		),
	)

	manager.UpdateDevices(ctx, *userConfig)
	<-app.WaitShutdown()
	os.Exit(0)
}
