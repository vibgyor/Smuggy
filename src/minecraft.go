package main

import (
//	"fmt"
	"strings"
	"time"
	
	"github.com/bwmarrin/discordgo"
)

var (
	starvedToDeath = false // only true if second place player starves to death
)

func startPVP(s *discordgo.Session, m *discordgo.Message) {
	time.Sleep(138 * time.Minute)
	if config.PVP != true {
		config.PVP = true
		s.ChannelMessageSend(m.ChannelID, ":warning: **PVP is now on!** :warning:")
		
		// award all players 5 participation points
		for i := range userList.Member {
			newPoints := userList.Member[i].Points + 5
			userList.Member[i] = &User{
				ID: userList.Member[i].ID,
				Username: userList.Member[i].Username,
				Nick: userList.Member[i].Nick,
				Roles: userList.Member[i].Roles,
				Points: newPoints,
				Dead: userList.Member[i].Dead,
			}
		}
	}
	
	saveConfig()
}

func parseMCBot(s *discordgo.Session, m *discordgo.Message) {
	// ignore messages not from the minecraft bot and not in the relay channel
	if m.Author.ID != config.MCBotID {
		return
	}
	
	if m.ChannelID != config.RelayChannel {
		return
	}
	
	// TODO: prevent players from being able to exploit the bot from Minecraft chat
	checkPlayerDied(m)
	checkKilledPlayer(m)
	checkLeftServer(m)
	
	return
}

func checkPlayerDied(m *discordgo.Message) {
	msg := strings.ToLower(m.Content)
	
	if strings.Contains(msg, "just died due to") {
		msgList := strings.Fields(m.Content)
		nick := msgList[0]
		nick = strings.Trim(nick, "*")
		
		var userID string
		
		if nick != "" {
			userID = getUserID(nick)
		}
		
		if _, ok := userList.Member[userID]; ok {
			if userList.Member[userID].Dead != true {
				var newPoints int
				
				deadUsers := getDeadUsers()
				totalUsers := len(userList.Member)
				secondPlace := totalUsers - 2
				
				if deadUsers == 0 {
					newPoints = 0
				} else if deadUsers == secondPlace {
					newPoints = userList.Member[userID].Points + 10
					
					if strings.Contains(msg, "starved to death") {
						starvedToDeath = true
					}
				} else {
					newPoints = userList.Member[userID].Points
				}
				
				userList.Member[userID] = &User{
					ID: userList.Member[userID].ID,
					Username: userList.Member[userID].Username,
					Nick: userList.Member[userID].Nick,
					Roles: userList.Member[userID].Roles,
					Points: newPoints,
					Dead: true,
				}
				
				saveUserList()
			}
		}
	}
}

func checkKilledPlayer(m *discordgo.Message) {
	msg := strings.ToLower(m.Content)
	
	if strings.Contains(msg, "was slain by") {
		msgList := strings.Fields(m.Content)
		idx := len(msgList) - 1
		nick := msgList[idx]
		
		var userID string
		
		if nick != "" {
			userID = getUserID(nick)
		}
		
		if _, ok := userList.Member[userID]; ok {
			if userList.Member[userID].Dead != true {
				newPoints := userList.Member[userID].Points + 2
				
				deadUsers := getDeadUsers()
				totalUsers := len(userList.Member)
				firstPlace := totalUsers - 1
				
				if deadUsers == firstPlace {
					if starvedToDeath == true {
						// first place player gets no points
						newPoints = userList.Member[userID].Points
					} else {
						newPoints = userList.Member[userID].Points + 15
					}
				}
			
				userList.Member[userID] = &User{
					ID: userList.Member[userID].ID,
					Username: userList.Member[userID].Username,
					Nick: userList.Member[userID].Nick,
					Roles: userList.Member[userID].Roles,
					Points: newPoints,
					Dead: userList.Member[userID].Dead,
				}
				
				saveUserList()
			}
		}
	}
}

func checkLeftServer(m *discordgo.Message) {
	msg := strings.ToLower(m.Content)
	
	if strings.Contains(msg, "just left the server!") {
		msgList := strings.Fields(m.Content)
		nick := msgList[0]
		nick = strings.Trim(nick, "*")
		
		var userID string
		
		if nick != "" {
			userID = getUserID(nick)
		}
		
		if _, ok := userList.Member[userID]; ok {
			if userList.Member[userID].Dead != true {
				if config.EventStarted == true {
					if config.PVP == true{
						// punish the player for leaving in the middle of the event before dying
						userList.Member[userID] = &User{
							ID: userList.Member[userID].ID,
							Username: userList.Member[userID].Username,
							Nick: userList.Member[userID].Nick,
							Roles: userList.Member[userID].Roles,
							Points: userList.Member[userID].Points,
							Dead: true,
						}
						
						saveUserList()
					}
				}
			}
		}
	}
}

func getDeadUsers() int {
	var deadUsers int
	
	for i := range userList.Member {
		if userList.Member[i].Dead == true {
			deadUsers++
		}
	}
	
	return deadUsers
}