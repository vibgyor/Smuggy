package main

import (
	"github.com/bwmarrin/discordgo"
)

type Command struct {
	Name string
	
	Exec func(*discordgo.Session, *discordgo.Message, []string)
}

type Config struct {
	Status string `json:"status"`
	Token string `json:"token"`
	CreatorID string `json:"creator_id"`
	MCBotID string `json:"mc_bot_id"`
	RelayChannel string `json:"relay_channel"`
	LeaderboardChannel string `json:"leaderboard_channel"`
	EventStarted bool `json:"event_started"`
	PVP bool `json:"pvp"`
}

type UserList struct {
	Members map[string]*User
}

type User struct {
	ID string `json:"id"`
	Username string `json:"username"`
	Nick string `json:"nick"`
	Roles map[string][]GuildRole `json:"roles"`
	Points int `json:"points"`
	Dead bool `json:"dead"`
	Stats UserStats `json:"stats"`
}

type GuildRole struct {
	ID string `json:"id"`
	Name string `json:"name"`
}

type UserStats struct {
	Wins int `json:"wins"`
	Participations int `json:"participations"`
	Kills int `json:"Kills"`
	PlayerDeaths int `json:"player_deaths"`
	TotalDeaths int `json:"total_deaths"`
	FirstDeaths int `json:"first_deaths"`
}