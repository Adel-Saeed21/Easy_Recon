package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)


const configFile = ".easyrecon_config.json"

type Config struct {
	BotToken string `json:"bot_token"`
	ChatID   int64  `json:"chat_id"`
}

func loadConfig() (*Config, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func saveConfig(cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configFile, data, 0600)
}

// ─── Telegram API types ───────────────────────────────────────────────────────

type TGUpdate struct {
	UpdateID int64      `json:"update_id"`
	Message  *TGMessage `json:"message"`
}

type TGMessage struct {
	MessageID int64  `json:"message_id"`
	Chat      TGChat `json:"chat"`
	Text      string `json:"text"`
}

type TGChat struct {
	ID int64 `json:"id"`
}

type TGResponse struct {
	OK     bool       `json:"ok"`
	Result []TGUpdate `json:"result"`
}


type TelegramBot struct {
	token  string
	chatID int64

	InputCh chan string

	OutputCh chan string

	offset int64
	mu     sync.Mutex
}

func NewTelegramBot(token string, chatID int64) *TelegramBot {
	return &TelegramBot{
		token:    token,
		chatID:   chatID,
		InputCh:  make(chan string, 10),
		OutputCh: make(chan string, 100),
	}
}

func (b *TelegramBot) apiURL(method string) string {
	return fmt.Sprintf("https://api.telegram.org/bot%s/%s", b.token, method)
}

func (b *TelegramBot) SendMessage(text string) error {
	if text == "" {
		return nil
	}
	params := url.Values{}
	params.Set("chat_id", strconv.FormatInt(b.chatID, 10))
	params.Set("text", text)

	resp, err := http.PostForm(b.apiURL("sendMessage"), params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// getUpdates polls Telegram for new messages.
func (b *TelegramBot) getUpdates() ([]TGUpdate, error) {
	params := url.Values{}
	params.Set("offset", strconv.FormatInt(b.offset, 10))
	params.Set("timeout", "10")

	resp, err := http.PostForm(b.apiURL("getUpdates"), params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var tgResp TGResponse
	if err := json.Unmarshal(body, &tgResp); err != nil {
		return nil, err
	}
	if !tgResp.OK {
		return nil, fmt.Errorf("telegram API returned not-ok")
	}
	return tgResp.Result, nil
}

// Start launches the two goroutines: one polls for input, one forwards output.
func (b *TelegramBot) Start() {
	// Goroutine 1: poll Telegram → put text into InputCh
	go func() {
		for {
			updates, err := b.getUpdates()
			if err != nil {
				time.Sleep(2 * time.Second)
				continue
			}
			for _, u := range updates {
				b.mu.Lock()
				b.offset = u.UpdateID + 1
				b.mu.Unlock()

				if u.Message == nil {
					continue
				}
				// Only accept messages from the authorised chat
				if u.Message.Chat.ID != b.chatID {
					continue
				}
				b.InputCh <- strings.TrimSpace(u.Message.Text)
			}
		}
	}()

	// Goroutine 2: read OutputCh → send to Telegram
	go func() {
		for line := range b.OutputCh {
			_ = b.SendMessage(line)
		}
	}()
}

// ─── Interactive stdio bridge ─────────────────────────────────────────────────
//
// TelegramIO replaces the normal os.Stdin / fmt.Print flow so that easyRecon
// can ask questions and receive answers through Telegram instead of the terminal.

type TelegramIO struct {
	bot *TelegramBot
}

// Ask sends a question to Telegram and waits for the user's reply.
func (t *TelegramIO) Ask(question string) string {
	t.bot.OutputCh <- question
	answer := <-t.bot.InputCh
	return answer
}

// Print forwards a line of output to Telegram (and also prints to stdout).
func (t *TelegramIO) Print(line string) {
	fmt.Println(line)
	if strings.TrimSpace(line) != "" {
		t.bot.OutputCh <- line
	}
}

// AskYesNo sends a yes/no question and returns true for "y"/"yes".
func (t *TelegramIO) AskYesNo(question string) bool {
	answer := t.Ask(question + "\n\nReply y or n")
	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes"
}

func SetupBot() (*TelegramBot, *TelegramIO, error) {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Println()
		fmt.Println("  ┌─────────────────────────────────────────────────┐")
		fmt.Println("  │  Telegram Bot Setup (first run)                 │")
		fmt.Println("  │                                                 │")
		fmt.Println("  │  1. Open Telegram and message @BotFather        │")
		fmt.Println("  │  2. Send /newbot and follow the instructions    │")
		fmt.Println("  │  3. Copy the token and paste it below           │")
		fmt.Println("  └─────────────────────────────────────────────────┘")
		fmt.Println()

		scanner := bufio.NewScanner(os.Stdin)

		fmt.Print("  Enter Bot Token: ")
		scanner.Scan()
		token := strings.TrimSpace(scanner.Text())
		if token == "" {
			return nil, nil, fmt.Errorf("bot token cannot be empty")
		}

		// Ask the user to send /start so we can learn their chat ID
		fmt.Println()
		fmt.Println("  [*] Now open your bot in Telegram and send:  /start")
		fmt.Println("  [*] Waiting for your message...")
		fmt.Println()

		chatID, err := waitForFirstMessage(token)
		if err != nil {
			return nil, nil, fmt.Errorf("could not get chat ID: %w", err)
		}

		cfg = &Config{BotToken: token, ChatID: chatID}
		if err := saveConfig(cfg); err != nil {
			fmt.Printf("  [!] Warning: could not save config: %v\n", err)
		}

		fmt.Printf("  [✓] Linked to chat ID %d — config saved to %s\n\n", chatID, configFile)
	}

	bot := NewTelegramBot(cfg.BotToken, cfg.ChatID)
	bot.Start()

	tio := &TelegramIO{bot: bot}
	return bot, tio, nil
}

// waitForFirstMessage polls until a /start message arrives; returns the chat ID.
func waitForFirstMessage(token string) (int64, error) {
	tmpBot := &TelegramBot{token: token}
	deadline := time.Now().Add(2 * time.Minute)

	for time.Now().Before(deadline) {
		updates, err := tmpBot.getUpdates()
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}
		for _, u := range updates {
			if u.Message != nil {
				return u.Message.Chat.ID, nil
			}
		}
		time.Sleep(2 * time.Second)
	}
	return 0, fmt.Errorf("timed out waiting for /start message")
}