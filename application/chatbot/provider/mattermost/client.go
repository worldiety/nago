// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mattermost

import (
	"errors"
	"net/http"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/chatbot/channel"
	"go.wdy.de/nago/application/chatbot/user"
	"go.wdy.de/nago/pkg/xhttp"
)

type Client struct {
	token string
	cl    *http.Client
	group *xhttp.RequestGroup
	base  string
}

func NewClient(settings Settings) *Client {
	return &Client{
		token: settings.Token,
		cl: &http.Client{
			Timeout: time.Second * 30,
		},
		group: xhttp.NewRequestGroup().RateLimit(settings.RPS),
		base:  settings.URL,
	}
}

type User struct {
	Id            string `json:"id"`
	CreateAt      int    `json:"create_at"`
	UpdateAt      int    `json:"update_at"`
	DeleteAt      int    `json:"delete_at"`
	Username      string `json:"username"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Nickname      string `json:"nickname"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	AuthService   string `json:"auth_service"`
	Roles         string `json:"roles"`
	Locale        string `json:"locale"`
	NotifyProps   struct {
		Email                string `json:"email"`
		Push                 string `json:"push"`
		Desktop              string `json:"desktop"`
		DesktopSound         string `json:"desktop_sound"`
		MentionKeys          string `json:"mention_keys"`
		Channel              string `json:"channel"`
		FirstName            string `json:"first_name"`
		AutoResponderMessage string `json:"auto_responder_message"`
		PushThreads          string `json:"push_threads"`
		Comments             string `json:"comments"`
		DesktopThreads       string `json:"desktop_threads"`
		EmailThreads         string `json:"email_threads"`
	} `json:"notify_props"`
	Props struct {
	} `json:"props"`
	LastPasswordUpdate int  `json:"last_password_update"`
	LastPictureUpdate  int  `json:"last_picture_update"`
	FailedAttempts     int  `json:"failed_attempts"`
	MfaActive          bool `json:"mfa_active"`
	Timezone           struct {
		UseAutomaticTimezone string `json:"useAutomaticTimezone"`
		ManualTimezone       string `json:"manualTimezone"`
		AutomaticTimezone    string `json:"automaticTimezone"`
	} `json:"timezone"`
	TermsOfServiceId       string `json:"terms_of_service_id"`
	TermsOfServiceCreateAt int    `json:"terms_of_service_create_at"`
}

func (c User) IntoUser() user.User {
	return user.User{
		ID:        user.ID(c.Id),
		Firstname: c.FirstName,
		Lastname:  c.LastName,
		Nickname:  c.Nickname,
		Email:     user.Email(c.Email),
	}
}

func (c *Client) Users() ([]User, error) {
	var resp []User
	err := xhttp.NewRequest().
		Group(c.group).
		BaseURL(c.base).
		Query("per_page", "1000").
		URL("api/v4/users").
		Assert2xx(true).
		BearerAuthentication(c.token).
		ToJSON(&resp).
		Get()

	if err != nil {
		return resp, err
	}

	return resp, nil
}

type ChannelCreatedResponse struct {
	Id            string `json:"id"`
	CreateAt      int64  `json:"create_at"`
	UpdateAt      int64  `json:"update_at"`
	DeleteAt      int64  `json:"delete_at"`
	TeamId        string `json:"team_id"`
	Type          string `json:"type"`
	DisplayName   string `json:"display_name"`
	Name          string `json:"name"`
	Header        string `json:"header"`
	Purpose       string `json:"purpose"`
	LastPostAt    int64  `json:"last_post_at"`
	TotalMsgCount int    `json:"total_msg_count"`
	ExtraUpdateAt int64  `json:"extra_update_at"`
	CreatorId     string `json:"creator_id"`
}

func (c ChannelCreatedResponse) IntoChannel() channel.Channel {
	return channel.Channel{
		ID:   channel.ID(c.Id),
		Name: c.Name,
	}
}

func (c *Client) CreateChannelDirect(userA ...string) (ChannelCreatedResponse, error) {
	var resp ChannelCreatedResponse
	err := xhttp.NewRequest().
		Group(c.group).
		BaseURL(c.base).
		URL("api/v4/channels/direct").
		Assert2xx(true).
		BearerAuthentication(c.token).
		ToJSON(&resp).
		BodyJSON(userA).
		Post()

	if err != nil {
		return resp, err
	}

	return resp, nil
}

func (c *Client) UsersMe() (User, error) {
	var resp User
	err := xhttp.NewRequest().
		Group(c.group).
		BaseURL(c.base).
		URL("api/v4/users/me").
		Assert2xx(true).
		BearerAuthentication(c.token).
		ToLimit(1024 * 1024).
		ToJSON(&resp).
		Get()

	if err != nil {
		return resp, err
	}

	return resp, nil
}

type CreatePostRequest struct {
	ChannelId string   `json:"channel_id"`
	Message   string   `json:"message"`
	RootId    string   `json:"root_id,omitempty"`
	FileIds   []string `json:"file_ids,omitempty"`
	Props     *struct {
	} `json:"props,omitempty"`
	Metadata *struct {
		Priority struct {
			Priority     string `json:"priority"`
			RequestedAck bool   `json:"requested_ack"`
		} `json:"priority"`
	} `json:"metadata,omitempty"`
}

type PostResponse struct {
	Id         string `json:"id"`
	CreateAt   int64  `json:"create_at"`
	UpdateAt   int64  `json:"update_at"`
	DeleteAt   int64  `json:"delete_at"`
	EditAt     int64  `json:"edit_at"`
	UserId     string `json:"user_id"`
	ChannelId  string `json:"channel_id"`
	RootId     string `json:"root_id"`
	OriginalId string `json:"original_id"`
	Message    string `json:"message"`
	Type       string `json:"type"`
	Props      struct {
	} `json:"props"`
	Hashtag       string   `json:"hashtag"`
	FileIds       []string `json:"file_ids"`
	PendingPostId string   `json:"pending_post_id"`
	Metadata      struct {
		Embeds []struct {
			Type string `json:"type"`
			Url  string `json:"url"`
			Data struct {
			} `json:"data"`
		} `json:"embeds"`
		Emojis []struct {
			Id        string `json:"id"`
			CreatorId string `json:"creator_id"`
			Name      string `json:"name"`
			CreateAt  int    `json:"create_at"`
			UpdateAt  int    `json:"update_at"`
			DeleteAt  int    `json:"delete_at"`
		} `json:"emojis"`
		Files []struct {
			Id              string `json:"id"`
			UserId          string `json:"user_id"`
			PostId          string `json:"post_id"`
			CreateAt        int    `json:"create_at"`
			UpdateAt        int    `json:"update_at"`
			DeleteAt        int    `json:"delete_at"`
			Name            string `json:"name"`
			Extension       string `json:"extension"`
			Size            int    `json:"size"`
			MimeType        string `json:"mime_type"`
			Width           int    `json:"width"`
			Height          int    `json:"height"`
			HasPreviewImage bool   `json:"has_preview_image"`
		} `json:"files"`
		Images struct {
		} `json:"images"`
		Reactions []struct {
			UserId    string `json:"user_id"`
			PostId    string `json:"post_id"`
			EmojiName string `json:"emoji_name"`
			CreateAt  int    `json:"create_at"`
		} `json:"reactions"`
		Priority struct {
			Priority     string `json:"priority"`
			RequestedAck bool   `json:"requested_ack"`
		} `json:"priority"`
		Acknowledgements []struct {
			UserId         string `json:"user_id"`
			PostId         string `json:"post_id"`
			AcknowledgedAt int    `json:"acknowledged_at"`
		} `json:"acknowledgements"`
	} `json:"metadata"`
}

func (c *Client) Post(req CreatePostRequest) (PostResponse, error) {
	var resp PostResponse
	err := xhttp.NewRequest().
		Group(c.group).
		BaseURL(c.base).
		URL("api/v4/posts").
		Assert2xx(true).
		BearerAuthentication(c.token).
		ToJSON(&resp).
		BodyJSON(req).
		Post()

	if err != nil {
		return resp, err
	}

	return resp, nil
}

func (c *Client) UserByEmail(mail string) (option.Opt[User], error) {
	if !user.Email(mail).Valid() {
		// security note: do not allow url path injection
		return option.None[User](), errors.New("invalid email")
	}

	var resp User
	err := xhttp.NewRequest().
		Group(c.group).
		BaseURL(c.base).
		URL("api/v4/users/email/" + mail).
		Assert2xx(true).
		BearerAuthentication(c.token).
		ToLimit(1024 * 1024).
		ToJSON(&resp).
		Get()

	if err != nil {
		var stat xhttp.UnexpectedStatusCodeError
		if errors.As(err, &stat) && stat.StatusCode == http.StatusNotFound {
			return option.None[User](), nil
		}

		return option.None[User](), err
	}

	return option.Some(resp), nil
}
