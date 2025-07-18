package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/arran4/golang-twickets/twitchets"
	"gopkg.in/yaml.v2"
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
	// ✅ Start minimal HTTP server for Render port binding
	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "Twitchets bot is running.")
		})

		port := os.Getenv("PORT")
		if port == "" {
			port = "10000"
		}

		fmt.Println("Starting HTTP server on port " + port)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// 🔧 Load config from environment variable
	cfgYaml := os.Getenv("CONFIG")
	if cfgYaml == "" {
		log.Fatal("CONFIG environment variable not set")
	}

	var cfg Config
	if err := yaml.Unmarshal([]byte(cfgYaml), &cfg); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	client := twitchets.NewClient(cfg.Country)
	notifier := twitchets.NewTelegramNotifier(cfg.Notification.Telegram.Token, cfg.Notification.Telegram.ChatID)

	fmt.Println("Starting Twitchets bot...")
	interval := time.Duration(cfg.Polling.Interval) * time.Second

	for {
		for _, ticket := range cfg.Tickets {
			fmt.Println("Polling Twickets event...")

			items, err := client.GetTickets(ticket.EventID)
			if err != nil {
				log.Printf("Error fetching tickets: %v", err)
				continue
			}

			for _, item := range items {
				if item.NumTickets >= ticket.NumTickets && item.DiscountPercent() >= ticket.Discount {
					err := notifier.Notify(item)
					if err != nil {
						log.Printf("Notify error: %v", err)
					} else {
						log.Printf("Notified for event %s: %d tickets at £%.2f", ticket.EventID, item.NumTickets, item.Price)
					}
				}
			}
		}

		time.Sleep(interval)
	}
}
