// SPDX-FileCopyrightText: AnmiTaliDev <anmitalidev@nuros.org>
// SPDX-License-Identifier: AGPL-3.0-only

package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') ||
			(value[0] == '\'' && value[len(value)-1] == '\'')) {
			value = value[1 : len(value)-1]
		}
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
}

type Config struct {
	BotToken     string
	ModeratorIDs []int64
	ChannelID    int64
}

func Load() (*Config, error) {
	loadDotEnv(".env")

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("BOT_TOKEN environment variable is required")
	}

	modIDsStr := os.Getenv("MODERATOR_IDS")
	var modIDs []int64
	for _, s := range strings.Split(modIDsStr, ",") {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		id, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid moderator ID %q: %w", s, err)
		}
		modIDs = append(modIDs, id)
	}

	channelIDStr := os.Getenv("CHANNEL_ID")
	if channelIDStr == "" {
		return nil, fmt.Errorf("CHANNEL_ID environment variable is required")
	}
	channelID, err := strconv.ParseInt(channelIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid CHANNEL_ID %q: %w", channelIDStr, err)
	}

	return &Config{
		BotToken:     token,
		ModeratorIDs: modIDs,
		ChannelID:    channelID,
	}, nil
}

func (c *Config) IsModerator(id int64) bool {
	for _, m := range c.ModeratorIDs {
		if m == id {
			return true
		}
	}
	return false
}
