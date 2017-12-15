package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var (
	config = &Config{}
	userList = &UserList{}
	userIDRegex = regexp.MustCompile("<@!?([0-9]{18})>")
)

func loadJSON(path string, v interface{}) error {
	f, err := os.OpenFile(path, os.O_RDONLY, 0600)
	if err != nil {
		return err
	}
	
	return json.NewDecoder(f).Decode(v)
}

func saveJSON(path string, data interface{}) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

func loadConfig() error {
	err := loadJSON("config.json", config)
	if err != nil {
		fmt.Println("Error loading config: ", err)
		return err
	}
	
	return nil
}

func saveConfig() {
	err := saveJSON("config.json", config)
	if err != nil {
		fmt.Println("Error saving config", err)
		return
	}
}

func loadUserList() error {
	userList.Members = make(map[string]*User)
	err := loadJSON("userlist.json", userList)
	if err != nil {
		fmt.Println("Error loading user list: ", err)
	}
	
	return err
}

func saveUserList() {
	err := saveJSON("userlist.json", userList)
	if err != nil {
		fmt.Println("Error saving user list: ", err)
	}
}

func main() {
	loadConfig()
	loadUserList()
	
	defer saveConfig()
	defer saveUserList()
	
	dg, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	dg.AddHandler(messageCreate)
	dg.AddHandler(messageUpdate)

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}
	
	dg.UpdateStatus(0, config.Status)

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	dg.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	
	processCommands(s, m.Message)
	parseMCBot(s, m.Message)
}

func messageUpdate(s *discordgo.Session, m *discordgo.MessageUpdate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	
	processCommands(s, m.Message)
}

func processCommands(s *discordgo.Session, m *discordgo.Message) {
	if strings.HasPrefix(m.Content, "!") {
		parseCommand(s, m, strings.TrimPrefix(m.Content, "!"))
	}
}