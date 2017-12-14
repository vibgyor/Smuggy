package main

import (
	"github.com/bwmarrin/discordgo"
)

func checkRolesInGuild(s *discordgo.Session, m *discordgo.Message, roleName string) bool {
	channelInGuild, _ := s.State.Channel(m.ChannelID)
	guild, _ := s.State.Guild(channelInGuild.GuildID)
	roles := guild.Roles
	user, _ := s.User(m.Author.ID)
	members, _ := s.State.Member(channelInGuild.GuildID, user.ID)
	
	var roleNames []string
	
	for _, role := range members.Roles {
		for _, guildRole := range roles {
			if guildRole.ID == role {
				roleNames = append(roleNames, guildRole.Name)
			}
		}
	}
	
	for i := range roleNames {
		if roleNames[i] == roleName {
			return true
		}
	}
	
	return false
}

func getUserID(name string) string {
	var userID string
	for i := range userList.Member {
		if userList.Member[i].Nick == "" {
			if name == userList.Member[i].Username {
				userID = userList.Member[i].ID
			}
		} else {
			if name == userList.Member[i].Nick {
				userID = userList.Member[i].ID
			}
		}
	}
	
	return userID
}