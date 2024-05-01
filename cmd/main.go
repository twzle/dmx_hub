package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"git.miem.hse.ru/hubman/dmx-executor/internal"
	"git.miem.hse.ru/hubman/dmx-executor/internal/device"
	"git.miem.hse.ru/hubman/dmx-executor/internal/models"
	"git.miem.hse.ru/hubman/hubman-lib"
	"git.miem.hse.ru/hubman/hubman-lib/core"
	"git.miem.hse.ru/hubman/hubman-lib/executor"
)

func main() {
	systemConfig := &core.SystemConfig{}
	userConfig := &device.UserConfig{}

	err := core.ReadConfig(systemConfig, userConfig)
	if err != nil {
		log.Fatalf("error while init config: %v", err)
	}
	ctx := context.Background()

	agentConf := core.AgentConfiguration{
		System:          systemConfig,
		User:            userConfig,
		ParseUserConfig: func(data []byte) (core.Configuration, error) { return device.ParseConfigFromBytes(data) },
	}

	app := core.NewContainer(agentConf.System.Logging)
	checkManager := core.NewCheckManager()
	logger := app.Logger()

	manager := internal.NewManager(logger, checkManager)
	signals := manager.GetSignals()

	app.RegisterPlugin(
		hubman.NewAgentPlugin(
			app.Logger(),
			agentConf,
			hubman.WithManipulator(
				hubman.WithSignal[models.SceneChanged](),
				hubman.WithSignal[models.SceneSaved](),
				hubman.WithChannel(signals),
			),
			hubman.WithExecutor(
				hubman.WithCommand(models.SetChannel{}, func(command core.SerializedCommand, parser executor.CommandParser) error {
					var cmd models.SetChannel // json-like api
					parser(&cmd)              // enriches your command with data from redis

					return manager.ProcessSetChannel(ctx, cmd)
				}),
				hubman.WithCommand(models.IncrementChannel{}, func(command core.SerializedCommand, parser executor.CommandParser) error {
					var cmd models.IncrementChannel // json-like api
					parser(&cmd)                    // enriches your command with data from redis

					return manager.ProcessIncrementChannel(ctx, cmd)
				}),
				hubman.WithCommand(models.Blackout{}, func(command core.SerializedCommand, parser executor.CommandParser) error {
					var cmd models.Blackout // json-like api
					parser(&cmd)            // enriches your command with data from redis

					return manager.ProcessBlackout(ctx, cmd)
				}),
				hubman.WithCommand(models.SetScene{}, func(command core.SerializedCommand, parser executor.CommandParser) error {
					var cmd models.SetScene // json-like api
					parser(&cmd)            // enriches your command with data from redis

					return manager.ProcessSetScene(ctx, cmd)
				}),
				hubman.WithCommand(models.SaveScene{}, func(command core.SerializedCommand, parser executor.CommandParser) error {
					var cmd models.SaveScene // json-like api
					parser(&cmd)             // enriches your command with data from redis

					return manager.ProcessSaveScene(ctx, cmd)
				}),
			),
			hubman.WithOnConfigRefresh(func(configuration core.AgentConfiguration) {
				update, ok := configuration.User.(*device.UserConfig)
				if !ok {
					panic(
						fmt.Sprintf(
							"Refresh config error: expected type %T, received %T",
							device.UserConfig{},
							configuration.User,
						),
					)
				}
				manager.UpdateDevices(ctx, *update)
			}),
			hubman.WithCheckRegistry(checkManager),
		),
	)

	manager.UpdateDevices(ctx, *userConfig)
	<-app.WaitShutdown()
	os.Exit(0)
}
