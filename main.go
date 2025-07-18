package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/arran4/golang-twickets/twitchets"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Country     string `yaml:"country"`
	Tickets     []struct {
		EventID    string `yaml:"eventId"`
		NumTickets int    `yaml:"numTickets"`
		Discount   int    `yaml:"discount"`
	} `yaml:"tickets"`
	Notification struct {
		Telegram struct {
			Token  string `yaml:"token"`
			ChatID int64  `yaml:"chatId"`
		} `yaml:"telegram"`
	} `yaml:"notification"`
	Polling struct {
		Interval int `yaml:"interval"`
	} `yaml:"polling"`
}

func main() {
	// Start HTTP server on port 10000
	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "Twitchets bot is running")
		})
		log.Println("HTTP server listening on :10000")
		log.Fatal(http.ListenAndServe(":10000", nil))
	}()

	// Load config from environment variable
	raw := os.Getenv("CONFIG")
	if raw == "" {
		log.Fatal("Missing CONFIG environment variable")
	}

	var cfg Config
	if err := yaml.Unmarshal([]byte(raw), &cfg); err != nil {
		log.Fatalf("Failed to parse CONFIG: %v", err)
	}

	log.Printf("Starting Twickets polling (%d events)...", len(cfg.Tickets))

	for _, t := range cfg.Tickets {
		go func(ticketCfg struct {
			EventID    string `yaml:"eventId"`
			NumTickets int    `yaml:"numTickets"`
			Discount   int    `yaml:"discount"`
		}) {
			client := twickets.NewTwicketsClient(cfg.Country)
			for {
				tickets, err := client.GetTickets(ticketCfg.EventID)
				if err == nil {
					for _, ticket := range tickets {
						if ticket.Tickets >= ticketCfg.NumTickets &&
							ticket.DiscountPercent() >= ticketCfg.Discount {
							message := fmt.Sprintf("🎟️ %dx tickets listed for £%.2f – %s", ticket.Tickets, ticket.Price, ticket.URL())
							twickets.SendTelegram(cfg.Notification.Telegram.Token, cfg.Notification.Telegram.ChatID, message)
						}
					}
				}
				time.Sleep(time.Duration(cfg.Polling.Interval) * time.Second)
			}
		}(t)
	}

	select {} // block forever
}
