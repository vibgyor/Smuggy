package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	
	"github.com/bwmarrin/discordgo"
)

var (
	commMap = make(map[string]Command)
	
	changeStatus = Command{"status", commandChangeStatus}.add()
	addUser = Command{"adduser", commandAddUser}.add()
	deleteUser = Command{"deleteuser", commandDeleteUser}.add()
	addPoints = Command{"addpoints", commandAddPoints}.add()
	removePoints = Command{"removepoints", commandRemovePoints}.add()
	getPoints = Command{"getpoints", commandGetPoints}.add()
	getLeaderboard = Command{"leaderboard", commandLeaderboard}.add()
	startEvent = Command{"start", commandStartEvent}.add()
	endEvent = Command{"end", commandEndEvent}.add()
)

func (c Command) add() Command {
	commMap[l(c.Name)] = c
	return c
}

func l(s string) string {
	return strings.ToLower(s)
}

func regexUserID(msgList []string) string {
	var userID string
	submatch := userIDRegex.FindStringSubmatch(msgList[1])
	if len(submatch) != 0 {
		userID = submatch[1]
	}
	return userID
}

func parseCommand(s *discordgo.Session, m *discordgo.Message, message string) {
	msgList := strings.Fields(message)
	command := l(func() string {
		if strings.HasPrefix(message, "!") {
			return " " + msgList[0]
		}
		return msgList[0]
	}())
	
	if command == strings.ToLower(commMap[command].Name) {
		commMap[command].Exec(s, m, msgList)
		return
	}
	
	return
}

func commandChangeStatus(s *discordgo.Session, m *discordgo.Message, msgList []string) {
	if len(msgList) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Cannot change status without a message")
		return
	}
	
	// only the bot creator can change its status
	if m.Author.ID != config.CreatorID {
		s.ChannelMessageSend(m.ChannelID, ":no_entry: You are not allowed to change my status")
		return
	}
	
	s.ChannelMessageSend(m.ChannelID, "ok, changing status")
	
	game := strings.Join(msgList[1:], " ")
	
	s.UpdateStatus(0, game)
	config.Status = game
	saveConfig()
	
	return
}

func commandAddUser(s *discordgo.Session, m *discordgo.Message, msgList []string) {
	var userID string
	
	checkRole := checkRolesInGuild(s, m, "Admin")
	if checkRole != true {
		s.ChannelMessageSend(m.ChannelID, ":no_entry: You are not allowed to use this command")
		return
	}
	
	if len(msgList) > 1 {
		userID = regexUserID(msgList)
	} else {
		userID = m.Author.ID
	}
	
	if userID == "" {
		s.ChannelMessageSend(m.ChannelID, ":no_entry: User does not exist")
		return
	}
	
	if userID == s.State.User.ID {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("I cannot add myself to the user list"))
		return
	}
	
	c, _ := s.Channel(m.ChannelID)
	member, _ := s.State.Member(c.GuildID, userID)
	
	if _, ok := userList.Member[userID]; !ok {
		userList.Member[userID] = &User{}
		
		userList.Member[userID].Roles = make(map[string][]GuildRole)
		g, _ := s.State.Guild(c.GuildID)
		roles := g.Roles
		for _, role := range roles {
			for _, roleID := range member.Roles {
				if role.ID == roleID {
					userList.Member[userID].Roles[role.ID] = append(userList.Member[userID].Roles[role.ID], GuildRole{
					ID: role.ID,
					Name: role.Name,
					})
				}
			}
		}
		
		userList.Member[userID] = &User{
			ID: userID,
			Username: member.User.Username,
			Nick: member.Nick,
			Roles: userList.Member[userID].Roles,
			Points: 0,
			Dead: false,
		}
		
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Successfully added <@%s> to user list", userID))
	} else {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("<@%s> is already on the user list", userID))
		return
	}
	
	saveUserList()
	return
}

func commandDeleteUser(s *discordgo.Session, m *discordgo.Message, msgList []string) {
	var userID string
	
	checkRole := checkRolesInGuild(s, m, "Admin")
	if checkRole != true {
		s.ChannelMessageSend(m.ChannelID, ":no_entry: You are not allowed to use this command")
		return
	}
	
	if len(msgList) > 1 {
		userID = regexUserID(msgList)
	} else {
		userID = m.Author.ID
	}
	
	if userID == "" {
		s.ChannelMessageSend(m.ChannelID, ":no_entry: User does not exist")
		return
	}
	
	if _, ok := userList.Member[userID]; ok {
		delete(userList.Member, userID)
		
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Successfully deleted <@%s> from user list", userID))
	} else {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Unable to delete <@%s> as they are not on the user list", userID))
		return
	}
	
	saveUserList()
	return
}

func commandAddPoints(s *discordgo.Session, m *discordgo.Message, msgList []string) {
	var userID string
	var pointsToAdd int
	
	checkRole := checkRolesInGuild(s, m, "Admin")
	if checkRole != true {
		s.ChannelMessageSend(m.ChannelID, ":no_entry: You are not allowed to use this command")
		return
	}
	
	if len(msgList) > 2 {
		userID = regexUserID(msgList)
		pointsToAdd, _ = strconv.Atoi(msgList[2])
	} else {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("I cannot add points without a user mentioned"))
		return
	}
	
	if userID == s.State.User.ID {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("I cannot add points to myself"))
		return
	}
	
	if userID == "" {
		s.ChannelMessageSend(m.ChannelID, ":no_entry: User does not exist")
		return
	}
	
	if _, ok := userList.Member[userID]; ok {
		newPoints := userList.Member[userID].Points + pointsToAdd
		userList.Member[userID] = &User{
			ID: userList.Member[userID].ID,
			Username: userList.Member[userID].Username,
			Nick: userList.Member[userID].Nick,
			Roles: userList.Member[userID].Roles,
			Points: newPoints,
			Dead: userList.Member[userID].Dead,
		}
		
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Added %d points to <@%s>", pointsToAdd, userID))
	} else {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Unable to add points to <@%s>", userID))
		return
	}
	
	saveUserList()
	return
}

func commandRemovePoints(s *discordgo.Session, m *discordgo.Message, msgList []string) {
	var userID string
	var pointsToRemove int
	
	checkRole := checkRolesInGuild(s, m, "Admin")
	if checkRole != true {
		s.ChannelMessageSend(m.ChannelID, ":no_entry: You are not allowed to use this command")
		return
	}
	
	if len(msgList) > 2 {
		userID = regexUserID(msgList)
		pointsToRemove, _ = strconv.Atoi(msgList[2])
	} else {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("I cannot remove points without a user mentioned"))
		return
	}
	
	if userID == s.State.User.ID {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("I cannot remove points from myself"))
		return
	}
	
	if userID == "" {
		s.ChannelMessageSend(m.ChannelID, ":no_entry: User does not exist")
		return
	}
	
	if _, ok := userList.Member[userID]; ok {
		newPoints := userList.Member[userID].Points - pointsToRemove
		
		if newPoints < 0 {
			newPoints = 0
		}
		
		userList.Member[userID] = &User{
			ID: userList.Member[userID].ID,
			Username: userList.Member[userID].Username,
			Nick: userList.Member[userID].Nick,
			Roles: userList.Member[userID].Roles,
			Points: newPoints,
			Dead: userList.Member[userID].Dead,
		}
		
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Removed %d points from <@%s>", pointsToRemove, userID))
	} else {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Unable to remove points from <@%s>", userID))
		return
	}
	
	saveUserList()
	return
}

func commandGetPoints(s *discordgo.Session, m *discordgo.Message, msgList []string) {
	var userID string
	
	if len(msgList) > 1 {
		userID = regexUserID(msgList)
	} else {
		userID = m.Author.ID
	}
	
	if userID == s.State.User.ID {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("I do not have any points to retrieve"))
		return
	}
	
	if userID == "" {
		s.ChannelMessageSend(m.ChannelID, ":no_entry: User does not exist")
		return
	}
	
	if _, ok := userList.Member[userID]; ok {
		totalPoints := userList.Member[userID].Points
		
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Total points for <@%s>: %d", userID, totalPoints))
	} else {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Unable to retrieve points from <@%s>", userID))
		return
	}
	
	return
}

func commandLeaderboard(s *discordgo.Session, m *discordgo.Message, msgList []string) {
	if m.ChannelID != config.LeaderboardChannel {
		s.ChannelMessageSend(m.ChannelID, "I can only display the leaderboard in your leaderboards channel. Try !leaderboard again there.")
		return
	}
	
	var leaderboard []int
	
	for i := range userList.Member {
		leaderboard = append(leaderboard, userList.Member[i].Points)
	}
	
	sort.Sort(sort.Reverse(sort.IntSlice(leaderboard)))
	
	for i := range leaderboard {
		for j := range userList.Member {
			if leaderboard[i] == userList.Member[j].Points {
				spot := i+1
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%d. <@%s>: %d points", spot, userList.Member[j].ID, userList.Member[j].Points))
			}
		}
	}
	
	return 
}

func commandStartEvent(s *discordgo.Session, m *discordgo.Message, msgList []string) {
	// we check roles in the guild in case an admin is not on the user list
	checkRole := checkRolesInGuild(s, m, "Admin")
	
	if checkRole == true {
		if config.EventStarted != true {
			config.EventStarted = true
			go startPVP(s, m)
		} else {
			s.ChannelMessageSend(m.ChannelID, "The event has already been started")
			return
		}
	} else {
		s.ChannelMessageSend(m.ChannelID, ":no_entry: You are not allowed to use this command")
		return
	}
	
	s.ChannelMessageSend(m.ChannelID, ":warning: **The event has been started!** :warning:")
	
	saveConfig()
	return 
}

func commandEndEvent(s *discordgo.Session, m *discordgo.Message, msgList []string) {
	checkRole := checkRolesInGuild(s, m, "Admin")
	
	if checkRole == true {
		if config.EventStarted == true {
			config.EventStarted = false
			if config.PVP == true {
				config.PVP = false
			}
		} else {
			s.ChannelMessageSend(m.ChannelID, "You cannot end an event that has not been started, dummy")
			return
		}
	} else {
		s.ChannelMessageSend(m.ChannelID, ":no_entry: You are not allowed to use this command")
		return
	}
	
	s.ChannelMessageSend(m.ChannelID, ":warning: **The event has ended!** :warning:")
	
	saveConfig()
	return
}