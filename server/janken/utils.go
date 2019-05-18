package janken

import (
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-server/model"
)

func (p *Plugin) SendEphemeralPost(channelId, userId, message string) {
	post := &model.Post{}
	post.ChannelId = channelId
	post.UserId = userId
	post.Message = message
	post.AddProp("sent_by_plugin", true)
	_ = p.API.SendEphemeralPost(userId, post)
}

func (p *Plugin) AppendMessage(post *model.Post, format string, args ...interface{}) (*model.Post) {
	message := fmt.Sprintf(format, args...)
	post.Message = fmt.Sprintf("%s\n%s", post.Message, message)
	return post
}

func (p *Plugin) writePostActionIntegrationResponse(response *model.PostActionIntegrationResponse, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(response.ToJson())
}
