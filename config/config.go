package config

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"strings"

	"github.com/diamondburned/arikawa/v3/discord"
)

func NewConfig() Config {
	return Config{}
}

type Config struct {
	DiscordToken            string            `koanf:"discord.token" description:"discord bot token from the discord developer website"`
	DiscordChannel          string            `koanf:"discord.channel" description:"channel id that the bot works on"`
	DiscordChannelSnowflake discord.ChannelID `koanf:"-"`
	EconAddress             string            `koanf:"econ.address" description:"ip:port"`
	EconPassword            string            `koanf:"econ.password" description:"econ server password"`
}

func (cfg *Config) Validate() error {

	if cfg.DiscordToken == "" {
		return errors.New("discord token must not be empty")
	}

	parts := strings.SplitN(cfg.EconAddress, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid address, expected host:port: %s", cfg.EconAddress)
	}
	host := parts[0]
	port := parts[1]

	ips, err := net.LookupHost(host)
	if err != nil {
		return fmt.Errorf("failed to resolve host: %s: %w", host, err)
	}

	// we try to select ipv4
	var selectedAddr netip.Addr
	for _, ip := range ips {
		addr, err := netip.ParseAddr(ip)
		if err != nil {
			return fmt.Errorf("failed to parse resolved address with port: %s: %w", addr, err)
		}

		if !selectedAddr.IsValid() {
			selectedAddr = addr
			if selectedAddr.Is4() {
				break
			}
		} else if addr.Is4() || addr.Is4In6() {
			selectedAddr = addr
		}
	}

	if !selectedAddr.IsValid() {
		return fmt.Errorf("could not select any resolved address for %s in %s", host, strings.Join(ips, ", "))
	}

	cfg.EconAddress = net.JoinHostPort(selectedAddr.String(), port)

	if cfg.EconPassword == "" {
		return errors.New("econ password must not be empty")
	}

	chanSnowflake, err := discord.ParseSnowflake(cfg.DiscordChannel)
	if err != nil {
		return fmt.Errorf("invalid discord channel: %w", err)
	}
	cfg.DiscordChannelSnowflake = discord.ChannelID(chanSnowflake)

	return nil
}
