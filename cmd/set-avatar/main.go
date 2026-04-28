// set-avatar is a one-shot utility for updating the bot's Discord avatar.
// Animated GIFs are supported for bots without Nitro.
//
// Usage:
//
//	go run ./cmd/set-avatar --config configs/config.yaml --avatar avatar.gif
//
// Rate limit: Discord allows roughly 2 avatar changes per hour.
package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/ak1m1tsu/barman/internal/infrastructure/config"
)

func main() {
	cfgPath := flag.String("config", "configs/config.yaml", "path to config.yaml")
	avatarPath := flag.String("avatar", "", "path to avatar image (gif, png, jpg, webp)")
	flag.Parse()

	if *avatarPath == "" {
		log.Fatal("--avatar is required")
	}

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	data, err := os.ReadFile(*avatarPath)
	if err != nil {
		log.Fatalf("read avatar: %v", err)
	}

	mime, err := mimeType(*avatarPath)
	if err != nil {
		log.Fatal(err)
	}

	dataURI := fmt.Sprintf("data:%s;base64,%s", mime, base64.StdEncoding.EncodeToString(data))

	s, err := discordgo.New("Bot " + cfg.Discord.Token)
	if err != nil {
		log.Fatalf("create session: %v", err)
	}

	if _, err = s.UserUpdate("", dataURI, ""); err != nil {
		log.Fatalf("update avatar: %v", err)
	}

	fmt.Printf("avatar updated successfully (%s, %d KB)\n", mime, len(data)/1024)
}

// mimeType returns the MIME type for the given image file path based on its extension.
func mimeType(path string) (string, error) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".gif":
		return "image/gif", nil
	case ".png":
		return "image/png", nil
	case ".jpg", ".jpeg":
		return "image/jpeg", nil
	case ".webp":
		return "image/webp", nil
	default:
		return "", fmt.Errorf("unsupported image format %q; use gif, png, jpg, or webp", filepath.Ext(path))
	}
}
