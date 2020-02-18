package main

import (
	"net/http"
	"testing"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/stretchr/testify/assert"
)


func TestWritePostActionIntegrationResponse(t *testing.T) {
	for name, test := range map[string]struct {
		ExpectedHeader     http.Header
		ExpectedStatusCode []int
	}{
		"successfully": {
			ExpectedHeader: http.Header{
				"Content-Type": []string{"application/json"},
			},
			ExpectedStatusCode: []int{http.StatusOK},
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			response := &model.PostActionIntegrationResponse{}
			w := NewTestResponseWriter()
			r := &http.Request{}
			writePostActionIntegrationResponse(response, w, r)

			assert.Equal(test.ExpectedHeader, w.header)
		})
	}
}
