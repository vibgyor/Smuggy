package main

import (
//	"fmt"
	"strings"
	"time"
	
	"github.com/bwmarrin/discordgo"
)

var (
	starvedToDeath = false // only true if second place player starves to death
	participatingPlayers []string
)

func startPVP(s *discordgo.Session, m *discordgo.Message) {
	time.Sleep(138 * time.Minute)
	if config.PVP != true {
		config.PVP = true
		s.ChannelMessageSend(m.ChannelID, ":warning: **PVP is now on!** :warning:")
		
		for i := range participatingPlayers {
			for j := range userList.Members {
				if participatingPlayers[i] == userList.Members[j].ID {
					newPoints := userList.Members[j].Points + 5
					
					userList.Members[j].Points = newPoints
					userList.Members[j].Dead = false
					userList.Members[j].Stats.Participations++
				}
			}
		}
		
		saveConfig()
		saveUserList()
	}
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
	checkJoinedServer(m)
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
		
		if _, ok := userList.Members[userID]; ok {
			if userList.Members[userID].Dead != true {
				var newPoints int
				
				deadUsers := getDeadUsers()
				totalUsers := len(userList.Members)
				secondPlace := totalUsers - 2
				
				if deadUsers == 0 {
					newPoints = 0
					userList.Members[userID].Stats.FirstDeaths++
				} else if deadUsers == secondPlace {
					newPoints = userList.Members[userID].Points + 10
					
					if strings.Contains(msg, "starved to death") {
						starvedToDeath = true
					}
				} else {
					newPoints = userList.Members[userID].Points
				}
				
				userList.Members[userID].Points = newPoints
				userList.Members[userID].Dead = true
				userList.Members[userID].Stats.TotalDeaths++
				
				if strings.Contains(msg, "was slain by") {
					userList.Members[userID].Stats.PlayerDeaths++
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
		
		if _, ok := userList.Members[userID]; ok {
			if userList.Members[userID].Dead != true {
				newPoints := userList.Members[userID].Points + 2
				
				deadUsers := getDeadUsers()
				totalUsers := len(userList.Members)
				firstPlace := totalUsers - 1
				
				if deadUsers == firstPlace {
					userList.Members[userID].Stats.Wins++
					if starvedToDeath != true {
						newPoints = userList.Members[userID].Points + 15
					}
				}
			
				userList.Members[userID].Points = newPoints
				userList.Members[userID].Stats.Kills++
				
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
		
		if _, ok := userList.Members[userID]; ok {
			if userList.Members[userID].Dead != true {
				if config.EventStarted == true {
					if config.PVP == true{
						// punish the player for leaving in the middle of the event before dying
						userList.Members[userID].Dead = true
						
						saveUserList()
					}
				}
			}
		}
	}
}

func checkJoinedServer(m *discordgo.Message) {
	msg := strings.ToLower(m.Content)
	
	if strings.Contains(msg, "just joined the server!") {
		msgList := strings.Fields(m.Content)
		nick := msgList[0]
		nick = strings.Trim(nick, "*")
		
		var userID string
		
		if nick != "" {
			userID = getUserID(nick)
		}
		
		participatingPlayers = append(participatingPlayers, userID)
	}
}

func getDeadUsers() int {
	var deadUsers int
	
	for i := range userList.Members {
		if userList.Members[i].Dead == true {
			deadUsers++
		}
	}
	
	return deadUsers
}