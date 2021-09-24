package main

type SCPayload struct { // Shortcut Payload
	Type                  string     `json:"type,omitempty"`
	CallbackId            string     `json:"callback_id,omitempty"`
	User                  *SCUser    `json:"user,omitempty"`
	Is_Enterprise_Install string     `json:"is_enterprise_install,omitempty"`
	ActionTs              string     `json:"action_ts,omitempty"`
	Team                  *SCTeam    `json:"team,omitempty"`
	Token                 string     `json:"token,omitempty"`
	TriggerId             string     `json:"trigger_id,omitempty"`
	MessageTs             string     `json:"message_ts,omitempty"`   // empty if the shortcut is global
	Message               *SCMessage `json:"message,omitempty"`      // empty if the shortcut is global
	ResponseUrl           string     `json:"response_url,omitempty"` // empty if the shortcut is global
	Channel               *SCChannel `json:"channel,omitempty"`      // empty if the shortcut is global
}

type SCMessage struct {
	ClientMsgId string          `json:"client_msg_id,omitempty"`
	Type        string          `json:"type,omitempty"`
	Text        string          `json:"text,omitempty"`
	User        string          `json:"user,omitempty"`
	Ts          string          `json:"ts,omitempty"`
	Team        string          `json:"team,omitempty"`
	Blocks      SCMessageBlocks `json:"blocks,omitempty"`
}

type SCMessageBlocks struct {
	Type     string         `json:"type,omitempty"`
	BlockId  string         `json:"block_id,omitempty"`
	Elements []MessageBlock `json:"lements,omitempty"`
}

type SCChannel struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type SCUser struct {
	Id       string `json:"id,omitempty"`
	Username string `json:"username,omitempty"`
	Name     string `json:"name,omitempty"`
	TeamId   string `json:"team_id,omitempty"`
}

type SCTeam struct {
	Id     string `json:"id,omitempty"`
	Domain string `json:"domain,omitempty"`
}

type SlashCommand struct {
	Token                 string `json:"token,omitempty"`
	Team_Id               string `json:"team_id,omitempty"`
	Team_Domain           string `json:"team_domain,omitempty"`
	Channel_Id            string `json:"channel_id,omitempty"`
	Channel_Name          string `json:"channel_name,omitempty"`
	User_Id               string `json:"user_id,omitempty"`
	User_Name             string `json:"user_name,omitempty"`
	Command               string `json:"command,omitempty"`
	Text                  string `json:"text,omitempty"`
	Api_App_Id            string `json:"api_app_id,omitempty"`
	Is_Enterprise_Install string `json:"is_enterprise_install,omitempty"`
	Response_Url          string `json:"response_url,omitempty"`
	Trigger_Id            string `json:"trigger_id,omitempty"`
}

type ActionClient struct{}

func (bc ActionClient) SendButton() (messageBlocks MessageBlocks) {

	messageBlocks = MessageBlocks{
		Blocks: []MessageBlock{
			MessageBlock{
				Type: "actions",
				Elements: []Element{
					Element{
						Type: "button",
						Text: ElementText{
							Type: "plain_text",
							Text: "Click Me",
						},
						Value:    "click_me_123",
						ActionId: "actionId-0",
					},
				},
			},
		},
	}

	return
}

// var ac = ActionClient{}
