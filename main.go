package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/api/cmdroute"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/jxsl13/cli-config-boilerplate/cliconfig"
	"github.com/jxsl13/tw-discord-econ-fifo/config"
	"github.com/spf13/cobra"
	"github.com/teeworlds-go/econ"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cmd := NewRootCmd(ctx)
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func NewRootCmd(ctx context.Context) *cobra.Command {
	cli := &CLI{
		ctx: ctx,
		cfg: config.NewConfig(),
	}

	cmd := cobra.Command{
		Use: filepath.Base(os.Args[0]),
	}
	cmd.PreRunE = cli.PrerunE(&cmd)
	cmd.RunE = cli.RunE
	return &cmd
}

type CLI struct {
	ctx context.Context
	cfg config.Config
}

func (cli *CLI) PrerunE(cmd *cobra.Command) func(*cobra.Command, []string) error {
	parser := cliconfig.RegisterFlags(&cli.cfg, false, cmd)
	return func(cmd *cobra.Command, args []string) error {
		log.SetOutput(cmd.OutOrStdout()) // redirect log output to stdout
		return parser()                  // parse registered commands
	}
}

func (cli *CLI) RunE(cmd *cobra.Command, args []string) error {
	log.Println("starting application...")
	var (
		address  = cli.cfg.EconAddress
		password = cli.cfg.EconPassword
		wg       sync.WaitGroup
		ctx      = cli.ctx
	)
	defer func() {
		log.Println("waiting for all goroutines to finish")
		wg.Wait()
		log.Println("all goroutines finished")
	}()

	log.Printf("connecting to %s", address)
	conn, err := econ.DialTo(address, password, econ.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to connect to econ: %w", err)
	}
	// safeguard in order to always close the connection
	defer func() {
		log.Println("closing econ connection...")
		_ = conn.Close()
		log.Println("econ connection closed")
	}()

	var (
		commandChan = make(chan string, 64)
	)

	wg.Add(1)
	go asyncWriteLine(ctx, &wg, conn, commandChan)

	var commands = []api.CreateCommandData{
		{
			Name:        "exec",
			Description: "execute rcon command",
			Options: discord.CommandOptions{
				discord.NewStringOption("command", "command to execute", true),
			},
		},
	}

	r := cmdroute.NewRouter()
	r.AddFunc("exec", func(ctx context.Context, data cmdroute.CommandData) *api.InteractionResponseData {
		if data.Event.ChannelID != cli.cfg.DiscordChannelSnowflake {
			return nil
		}

		commandOption := data.Options.Find("command")
		commandString := ""
		err = commandOption.Value.UnmarshalTo(&commandString)
		if err != nil {
			return &api.InteractionResponseData{
				Content: option.NewNullableString(fmt.Sprintf("execution failed: %v", err)),
				Flags:   discord.EphemeralMessage,
			}
		}

		select {
		case <-ctx.Done():
			return &api.InteractionResponseData{
				Content: option.NewNullableString("did not execute command, application shutdown."),
				Flags:   discord.EphemeralMessage,
			}
		case commandChan <- commandString:
			return &api.InteractionResponseData{
				Content: option.NewNullableString(fmt.Sprintf("executed `%s`", commandString)),
				Flags:   discord.EphemeralMessage,
			}
		default:
			return &api.InteractionResponseData{
				Content: option.NewNullableString("execution failed, command queue blocked."),
				Flags:   discord.EphemeralMessage,
			}
		}
	})

	s := state.New("Bot " + cli.cfg.DiscordToken)
	s.AddInteractionHandler(r)
	s.AddIntents(gateway.IntentGuilds)

	if err := cmdroute.OverwriteCommands(s, commands); err != nil {
		return fmt.Errorf("cannot update bot commands: %w", err)
	}

	log.Println("started application")
	err = s.Connect(ctx)
	if err != nil {
		return fmt.Errorf("cannot connect to discord: %w", err)
	}
	log.Println("discord connection closed")
	return nil
}
