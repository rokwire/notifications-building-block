// Package Def provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.12.4 DO NOT EDIT.
package Def

const (
	BearerAuthScopes = "bearerAuth.Scopes"
)

// CoreAccountRef defines model for CoreAccountRef.
type CoreAccountRef struct {
	Name   *string `json:"name,omitempty"`
	UserId *string `json:"user_id,omitempty"`
}

// FirebaseToken defines model for FirebaseToken.
type FirebaseToken struct {
	AppPlatform *string `json:"app_platform,omitempty"`
	AppVersion  *string `json:"app_version,omitempty"`
	DateCreated *string `json:"date_created,omitempty"`
	DateUpdated *string `json:"date_updated,omitempty"`
	Token       *string `json:"token,omitempty"`
}

// Message defines model for Message.
type Message struct {
	Id                       *string                 `json:"_id,omitempty"`
	AppId                    *string                 `json:"app_id,omitempty"`
	Body                     *string                 `json:"body,omitempty"`
	Data                     *[]string               `json:"data,omitempty"`
	DateCreated              *string                 `json:"date_created,omitempty"`
	DateUpdated              *string                 `json:"date_updated,omitempty"`
	OrgId                    *string                 `json:"org_id,omitempty"`
	Priority                 *string                 `json:"priority,omitempty"`
	RecipientAccountCriteria *map[string]interface{} `json:"recipient_account_criteria,omitempty"`
	Recipients               *Recipient              `json:"recipients,omitempty"`
	RecipientsCriteriaList   *RecipientCriteria      `json:"recipients_criteria_list,omitempty"`
	Sender                   *Sender                 `json:"sender,omitempty"`
	Subject                  *string                 `json:"subject,omitempty"`
	Topic                    *string                 `json:"topic,omitempty"`
}

// Recipient defines model for Recipient.
type Recipient struct {
	Mute                 *bool   `json:"mute,omitempty"`
	Name                 *string `json:"name,omitempty"`
	NotificationDisabled *bool   `json:"notification_disabled,omitempty"`
	Read                 *bool   `json:"read,omitempty"`
	UserId               *string `json:"user_id,omitempty"`
}

// RecipientCriteria defines model for RecipientCriteria.
type RecipientCriteria struct {
	AppPlatform *string `json:"app_platform,omitempty"`
	AppVersion  *string `json:"app_version,omitempty"`
}

// Sender defines model for Sender.
type Sender struct {
	Type *string         `json:"type,omitempty"`
	User *CoreAccountRef `json:"user,omitempty"`
}

// Topic defines model for Topic.
type Topic struct {
	AppId       *string `json:"app_id,omitempty"`
	DateCreated *string `json:"date_created,omitempty"`
	DateUpdated *string `json:"date_updated,omitempty"`
	Description *string `json:"description,omitempty"`
	Name        *string `json:"name,omitempty"`
	OrgId       *string `json:"org_id,omitempty"`
}

// User defines model for User.
type User struct {
	Id                    *string        `json:"_id,omitempty"`
	DateCreated           *string        `json:"date_created,omitempty"`
	DateUpdated           *string        `json:"date_updated,omitempty"`
	FirebaseTokens        *FirebaseToken `json:"firebase_tokens,omitempty"`
	NotificationsDisabled *string        `json:"notifications_disabled,omitempty"`
	Topics                *[]interface{} `json:"topics,omitempty"`
	UserId                *string        `json:"user_id,omitempty"`
}

// ClientReqMail defines model for _client_req_mail.
type ClientReqMail struct {
	Body    *string `json:"body,omitempty"`
	Subject *string `json:"subject,omitempty"`
	ToMail  *string `json:"to_mail,omitempty"`
}

// ClientReqMessage defines model for _client_req_message.
type ClientReqMessage struct {
	Ids []string `json:"_ids"`
}

// ClientReqMessageV2 defines model for _client_req_messageV2.
type ClientReqMessageV2 struct {
	Async   *bool                   `json:"async,omitempty"`
	Message *SharedReqCreateMessage `json:"message,omitempty"`
}

// ClientReqToken defines model for _client_req_token.
type ClientReqToken struct {
	AppPlatform   *string `json:"app_platform,omitempty"`
	AppVersion    *string `json:"app_version,omitempty"`
	PreviousToken *string `json:"previous_token,omitempty"`
	Token         string  `json:"token"`
}

// ClientReqUser defines model for _client_req_user.
type ClientReqUser struct {
	NotificationsDisabled bool `json:"notifications_disabled"`
}

// RecipientsReqAddRecipients defines model for _recipients_req_AddRecipients.
type RecipientsReqAddRecipients struct {
	Body *[]struct {
		Mute   bool   `json:"mute"`
		UserId string `json:"user_id"`
	} `json:"body,omitempty"`
}

// SharedReqCreateMessage defines model for _shared_req_CreateMessage.
type SharedReqCreateMessage struct {
	AppId string                 `json:"app_id"`
	Body  string                 `json:"body"`
	Data  map[string]interface{} `json:"data"`

	// Id optional
	Id                       *string                                        `json:"id,omitempty"`
	OrgId                    string                                         `json:"org_id"`
	Priority                 int                                            `json:"priority"`
	RecipientAccountCriteria map[string]interface{}                         `json:"recipient_account_criteria"`
	Recipients               []SharedReqCreateMessageInputMessageRecipient  `json:"recipients"`
	RecipientsCriteriaList   []SharedReqCreateMessageInputRecipientCriteria `json:"recipients_criteria_list"`
	Subject                  string                                         `json:"subject"`
	Time                     *int64                                         `json:"time,omitempty"`
	Topic                    *string                                        `json:"topic,omitempty"`
}

// SharedReqCreateMessageInputMessageRecipient defines model for _shared_req_CreateMessage_InputMessageRecipient.
type SharedReqCreateMessageInputMessageRecipient struct {
	Mute   bool   `json:"mute"`
	UserId string `json:"user_id"`
}

// SharedReqCreateMessageInputRecipientCriteria defines model for _shared_req_CreateMessage_InputRecipientCriteria.
type SharedReqCreateMessageInputRecipientCriteria struct {
	AppPlatform *string `json:"app_platform,omitempty"`
	AppVersion  *string `json:"app_version,omitempty"`
}

// SharedReqCreateMessages defines model for _shared_req_CreateMessages.
type SharedReqCreateMessages = []SharedReqCreateMessage

// GetApiAdminMessagesParams defines parameters for GetApiAdminMessages.
type GetApiAdminMessagesParams struct {
	// Offset offset
	Offset string `json:"offset"`

	// Limit limit
	Limit string `json:"limit"`

	// Order order - Possible values: asc, desc. Default: desc
	Order string `json:"order"`

	// StartDate start_date - Start date filter in milliseconds as an integer epoch value
	StartDate string `json:"start_date"`

	// EndDate end_date - End date filter in milliseconds as an integer epoch value
	EndDate string `json:"end_date"`
}

// DeleteApiBbsMessagesParams defines parameters for DeleteApiBbsMessages.
type DeleteApiBbsMessagesParams struct {
	// Ids ids of the messages for deletion separated with comma
	Ids string `json:"ids"`
}

// PostApiBbsMessagesMessageIdRecipientsParams defines parameters for PostApiBbsMessagesMessageIdRecipients.
type PostApiBbsMessagesMessageIdRecipientsParams struct {
	// Mute mute
	Mute bool `json:"mute"`
}

// GetApiMessagesParams defines parameters for GetApiMessages.
type GetApiMessagesParams struct {
	// Read read
	Read *bool `json:"read,omitempty"`

	// Mute mute
	Mute *bool `json:"mute,omitempty"`

	// Offset offset
	Offset string `json:"offset"`

	// Limit limit
	Limit string `json:"limit"`

	// Order order - Possible values: asc, desc. Default: desc
	Order string `json:"order"`

	// StartDate start_date - Start date filter in milliseconds as an integer epoch value
	StartDate string `json:"start_date"`

	// EndDate end_date - End date filter in milliseconds as an integer epoch value
	EndDate string `json:"end_date"`
}

// GetApiTopicTopicMessagesParams defines parameters for GetApiTopicTopicMessages.
type GetApiTopicTopicMessagesParams struct {
	// Offset offset
	Offset string `json:"offset"`

	// Limit limit - limit the result
	Limit string `json:"limit"`

	// Order order - Possible values: asc, desc. Default: desc
	Order string `json:"order"`

	// StartDate start_date - Start date filter in milliseconds as an integer epoch value
	StartDate string `json:"start_date"`

	// EndDate end_date - End date filter in milliseconds as an integer epoch value
	EndDate string `json:"end_date"`
}

// PostApiAdminMessageJSONRequestBody defines body for PostApiAdminMessage for application/json ContentType.
type PostApiAdminMessageJSONRequestBody = SharedReqCreateMessage

// PutApiAdminMessageJSONRequestBody defines body for PutApiAdminMessage for application/json ContentType.
type PutApiAdminMessageJSONRequestBody = SharedReqCreateMessage

// GetApiAdminMessagesJSONRequestBody defines body for GetApiAdminMessages for application/json ContentType.
type GetApiAdminMessagesJSONRequestBody = ClientReqMessage

// PutApiAdminTopicJSONRequestBody defines body for PutApiAdminTopic for application/json ContentType.
type PutApiAdminTopicJSONRequestBody = Topic

// PostApiBbsMailJSONRequestBody defines body for PostApiBbsMail for application/json ContentType.
type PostApiBbsMailJSONRequestBody = ClientReqMail

// PostApiBbsMessageJSONRequestBody defines body for PostApiBbsMessage for application/json ContentType.
type PostApiBbsMessageJSONRequestBody = ClientReqMessageV2

// PostApiBbsMessagesJSONRequestBody defines body for PostApiBbsMessages for application/json ContentType.
type PostApiBbsMessagesJSONRequestBody = SharedReqCreateMessages

// PostApiBbsMessagesMessageIdRecipientsJSONRequestBody defines body for PostApiBbsMessagesMessageIdRecipients for application/json ContentType.
type PostApiBbsMessagesMessageIdRecipientsJSONRequestBody = RecipientsReqAddRecipients

// PostApiIntMailJSONRequestBody defines body for PostApiIntMail for application/json ContentType.
type PostApiIntMailJSONRequestBody = ClientReqToken

// PostApiIntMessageJSONRequestBody defines body for PostApiIntMessage for application/json ContentType.
type PostApiIntMessageJSONRequestBody = SharedReqCreateMessage

// PostApiIntV2MessageJSONRequestBody defines body for PostApiIntV2Message for application/json ContentType.
type PostApiIntV2MessageJSONRequestBody = ClientReqMessageV2

// PostApiMessageJSONRequestBody defines body for PostApiMessage for application/json ContentType.
type PostApiMessageJSONRequestBody = SharedReqCreateMessage

// DeleteApiMessagesJSONRequestBody defines body for DeleteApiMessages for application/json ContentType.
type DeleteApiMessagesJSONRequestBody = ClientReqMessage

// GetApiMessagesJSONRequestBody defines body for GetApiMessages for application/json ContentType.
type GetApiMessagesJSONRequestBody = ClientReqMessage

// PostApiTokenJSONRequestBody defines body for PostApiToken for application/json ContentType.
type PostApiTokenJSONRequestBody = ClientReqToken

// PostApiTopicTopicSubscribeJSONRequestBody defines body for PostApiTopicTopicSubscribe for application/json ContentType.
type PostApiTopicTopicSubscribeJSONRequestBody = ClientReqToken

// PostApiTopicTopicUnsubscribeJSONRequestBody defines body for PostApiTopicTopicUnsubscribe for application/json ContentType.
type PostApiTopicTopicUnsubscribeJSONRequestBody = ClientReqToken

// DeleteApiUserJSONRequestBody defines body for DeleteApiUser for application/json ContentType.
type DeleteApiUserJSONRequestBody = ClientReqUser

// PutApiUserJSONRequestBody defines body for PutApiUser for application/json ContentType.
type PutApiUserJSONRequestBody = ClientReqUser
