package main

import (
	"context"
	"github.com/gontikr99/bidbot2/controller/bot"
	"github.com/gontikr99/bidbot2/controller/discord"
	"github.com/gontikr99/bidbot2/controller/everquest"
	"github.com/gontikr99/bidbot2/controller/gui"
	"github.com/gontikr99/bidbot2/controller/plugin"
	"github.com/gontikr99/bidbot2/controller/storage"
	"log"
	"runtime"
	"sync"
)

func main() {
	runtime.GOMAXPROCS(4)
	gui.RunMainWindow(&storage.BoltholdBackedConfig{},
		func(ctx context.Context, cc storage.ControllerConfig) {
			log.Println("Starting")
			gp, err := plugin.NewGuildPlugin(ctx, &storage.DatabaseWebCache{}, cc.RulesLua())
			if err != nil {
				log.Printf("Failed to create plugin: %v", err)
				return
			}

			// Connecting to Discord and setting up EverQuest both take some time, so do both
			// in parallel.
			var eqc *everquest.Client
			var dc *discord.Client
			errChan := make(chan error, 2)
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				var errT error
				eqc, errT = everquest.NewEqClient(ctx, cc)
				if errT != nil {
					errChan <- errT
					log.Printf("Failed to create EverQuest context: %v", err)
					return
				}

				errT = eqc.SetupCommandAndControl()
				if errT != nil {
					errChan <- errT
					log.Printf("Failed to set up Command & Control: %v", err)
					return
				}
				gp.SetEqClient(eqc)
				bot.RegisterLinkCommands(eqc)
			}()

			wg.Add(1)
			go func() {
				defer wg.Done()
				var errT error
				dc, errT = discord.NewDiscordClient(ctx, cc)
				if errT != nil {
					errChan <- errT
					log.Printf("Failed to connect to Discord: %v", err)
					return
				}
				gp.SetDiscordClient(dc)
				bot.RegisterDiscordBindCommands(dc)
			}()
			wg.Wait()

			select {
			case err = <-errChan:
				return
			case <-ctx.Done():
				return
			default:
				break
			}

			bot.RegisterDKPCommands(dc, eqc, gp)
			bot.RegisterAuctionCommand(eqc, dc, gp)
			bot.RegisterSayCommands(eqc, dc)
			bot.StartPeriodicRaidDumps(eqc, dc)
			log.Println("Initialization completed")
			log.Println("------------------------------")
			<-ctx.Done()
		})
}
