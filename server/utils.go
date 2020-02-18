package main

import (
	"fmt"

	"github.com/mattermost/mattermost-server/v5/model"
)

func (p *Plugin) sendEphemeralPost(channelID, userID, message string) *model.Post {
	post := &model.Post{}
	post.ChannelId = channelID
	post.UserId = userID
	post.Message = message
	post.AddProp("sent_by_plugin", true)
	return p.API.SendEphemeralPost(userID, post)
}

func appendMessage(post *model.Post, format string, args ...interface{}) *model.Post {
	message := fmt.Sprintf(format, args...)
	post.Message = fmt.Sprintf("%s\n%s", post.Message, message)
	return post
}
