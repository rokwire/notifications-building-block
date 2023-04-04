// Package Def provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.11.0 DO NOT EDIT.
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

// MessageRecipient defines model for MessageRecipient.
type MessageRecipient struct {
	AppId     *string `json:"app_id,omitempty"`
	Id        *string `json:"id,omitempty"`
	MessageId *string `json:"message_id,omitempty"`
	Mute      *bool   `json:"mute,omitempty"`
	OrgId     *string `json:"org_id,omitempty"`
	Read      *bool   `json:"read,omitempty"`
	UserId    *string `json:"user_id,omitempty"`
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

// AdminResGetMessagesStats defines model for _admin_res_GetMessagesStats.
type AdminResGetMessagesStats struct {
	FieldId *string `json:"field_id,omitempty"`
}

// BbsReqAddRecipients defines model for _bbs_req_AddRecipients.
type BbsReqAddRecipients = []struct {
	Mute   bool   `json:"mute"`
	UserId string `json:"user_id"`
}

// BbsReqRemoveRecipients defines model for _bbs_req_RemoveRecipients.
type BbsReqRemoveRecipients struct {
	UsersIds []string `json:"users_ids"`
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

// SharedReqCreateMessage defines model for _shared_req_CreateMessage.
type SharedReqCreateMessage struct {
	AppId string                 `json:"app_id"`
	Body  string                 `json:"body"`
	Data  map[string]interface{} `json:"data"`

	// optional
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

// PostApiAdminMessageJSONBody defines parameters for PostApiAdminMessage.
type PostApiAdminMessageJSONBody = SharedReqCreateMessage

// PutApiAdminMessageJSONBody defines parameters for PutApiAdminMessage.
type PutApiAdminMessageJSONBody = SharedReqCreateMessage

// GetApiAdminMessagesJSONBody defines parameters for GetApiAdminMessages.
type GetApiAdminMessagesJSONBody = ClientReqMessage

// GetApiAdminMessagesParams defines parameters for GetApiAdminMessages.
type GetApiAdminMessagesParams struct {
	// offset
	Offset string `json:"offset"`

	// limit
	Limit string `json:"limit"`

	// order - Possible values: asc, desc. Default: desc
	Order string `json:"order"`

	// start_date - Start date filter in milliseconds as an integer epoch value
	StartDate string `json:"start_date"`

	// end_date - End date filter in milliseconds as an integer epoch value
	EndDate string `json:"end_date"`
}

// PutApiAdminTopicJSONBody defines parameters for PutApiAdminTopic.
type PutApiAdminTopicJSONBody = Topic

// PostApiBbsMailJSONBody defines parameters for PostApiBbsMail.
type PostApiBbsMailJSONBody = ClientReqMail

// PostApiBbsMessageJSONBody defines parameters for PostApiBbsMessage.
type PostApiBbsMessageJSONBody = ClientReqMessageV2

// DeleteApiBbsMessagesParams defines parameters for DeleteApiBbsMessages.
type DeleteApiBbsMessagesParams struct {
	// ids of the messages for deletion separated with comma
	Ids string `json:"ids"`
}

// PostApiBbsMessagesJSONBody defines parameters for PostApiBbsMessages.
type PostApiBbsMessagesJSONBody = SharedReqCreateMessages

// DeleteApiBbsMessagesMessageIdRecipientsJSONBody defines parameters for DeleteApiBbsMessagesMessageIdRecipients.
type DeleteApiBbsMessagesMessageIdRecipientsJSONBody = BbsReqRemoveRecipients

// PostApiBbsMessagesMessageIdRecipientsJSONBody defines parameters for PostApiBbsMessagesMessageIdRecipients.
type PostApiBbsMessagesMessageIdRecipientsJSONBody = BbsReqAddRecipients

// PostApiIntMailJSONBody defines parameters for PostApiIntMail.
type PostApiIntMailJSONBody = ClientReqToken

// PostApiIntMessageJSONBody defines parameters for PostApiIntMessage.
type PostApiIntMessageJSONBody = SharedReqCreateMessage

// PostApiIntV2MessageJSONBody defines parameters for PostApiIntV2Message.
type PostApiIntV2MessageJSONBody = ClientReqMessageV2

// PostApiMessageJSONBody defines parameters for PostApiMessage.
type PostApiMessageJSONBody = SharedReqCreateMessage

// DeleteApiMessagesJSONBody defines parameters for DeleteApiMessages.
type DeleteApiMessagesJSONBody = ClientReqMessage

// GetApiMessagesJSONBody defines parameters for GetApiMessages.
type GetApiMessagesJSONBody = ClientReqMessage

// GetApiMessagesParams defines parameters for GetApiMessages.
type GetApiMessagesParams struct {
	// read
	Read *bool `json:"read,omitempty"`

	// mute
	Mute *bool `json:"mute,omitempty"`

	// offset
	Offset string `json:"offset"`

	// limit
	Limit string `json:"limit"`

	// order - Possible values: asc, desc. Default: desc
	Order string `json:"order"`

	// start_date - Start date filter in milliseconds as an integer epoch value
	StartDate string `json:"start_date"`

	// end_date - End date filter in milliseconds as an integer epoch value
	EndDate string `json:"end_date"`
}

// PostApiTokenJSONBody defines parameters for PostApiToken.
type PostApiTokenJSONBody = ClientReqToken

// GetApiTopicTopicMessagesParams defines parameters for GetApiTopicTopicMessages.
type GetApiTopicTopicMessagesParams struct {
	// offset
	Offset string `json:"offset"`

	// limit - limit the result
	Limit string `json:"limit"`

	// order - Possible values: asc, desc. Default: desc
	Order string `json:"order"`

	// start_date - Start date filter in milliseconds as an integer epoch value
	StartDate string `json:"start_date"`

	// end_date - End date filter in milliseconds as an integer epoch value
	EndDate string `json:"end_date"`
}

// PostApiTopicTopicSubscribeJSONBody defines parameters for PostApiTopicTopicSubscribe.
type PostApiTopicTopicSubscribeJSONBody = ClientReqToken

// PostApiTopicTopicUnsubscribeJSONBody defines parameters for PostApiTopicTopicUnsubscribe.
type PostApiTopicTopicUnsubscribeJSONBody = ClientReqToken

// DeleteApiUserJSONBody defines parameters for DeleteApiUser.
type DeleteApiUserJSONBody = ClientReqUser

// PutApiUserJSONBody defines parameters for PutApiUser.
type PutApiUserJSONBody = ClientReqUser

// PostApiAdminMessageJSONRequestBody defines body for PostApiAdminMessage for application/json ContentType.
type PostApiAdminMessageJSONRequestBody = PostApiAdminMessageJSONBody

// PutApiAdminMessageJSONRequestBody defines body for PutApiAdminMessage for application/json ContentType.
type PutApiAdminMessageJSONRequestBody = PutApiAdminMessageJSONBody

// GetApiAdminMessagesJSONRequestBody defines body for GetApiAdminMessages for application/json ContentType.
type GetApiAdminMessagesJSONRequestBody = GetApiAdminMessagesJSONBody

// PutApiAdminTopicJSONRequestBody defines body for PutApiAdminTopic for application/json ContentType.
type PutApiAdminTopicJSONRequestBody = PutApiAdminTopicJSONBody

// PostApiBbsMailJSONRequestBody defines body for PostApiBbsMail for application/json ContentType.
type PostApiBbsMailJSONRequestBody = PostApiBbsMailJSONBody

// PostApiBbsMessageJSONRequestBody defines body for PostApiBbsMessage for application/json ContentType.
type PostApiBbsMessageJSONRequestBody = PostApiBbsMessageJSONBody

// PostApiBbsMessagesJSONRequestBody defines body for PostApiBbsMessages for application/json ContentType.
type PostApiBbsMessagesJSONRequestBody = PostApiBbsMessagesJSONBody

// DeleteApiBbsMessagesMessageIdRecipientsJSONRequestBody defines body for DeleteApiBbsMessagesMessageIdRecipients for application/json ContentType.
type DeleteApiBbsMessagesMessageIdRecipientsJSONRequestBody = DeleteApiBbsMessagesMessageIdRecipientsJSONBody

// PostApiBbsMessagesMessageIdRecipientsJSONRequestBody defines body for PostApiBbsMessagesMessageIdRecipients for application/json ContentType.
type PostApiBbsMessagesMessageIdRecipientsJSONRequestBody = PostApiBbsMessagesMessageIdRecipientsJSONBody

// PostApiIntMailJSONRequestBody defines body for PostApiIntMail for application/json ContentType.
type PostApiIntMailJSONRequestBody = PostApiIntMailJSONBody

// PostApiIntMessageJSONRequestBody defines body for PostApiIntMessage for application/json ContentType.
type PostApiIntMessageJSONRequestBody = PostApiIntMessageJSONBody

// PostApiIntV2MessageJSONRequestBody defines body for PostApiIntV2Message for application/json ContentType.
type PostApiIntV2MessageJSONRequestBody = PostApiIntV2MessageJSONBody

// PostApiMessageJSONRequestBody defines body for PostApiMessage for application/json ContentType.
type PostApiMessageJSONRequestBody = PostApiMessageJSONBody

// DeleteApiMessagesJSONRequestBody defines body for DeleteApiMessages for application/json ContentType.
type DeleteApiMessagesJSONRequestBody = DeleteApiMessagesJSONBody

// GetApiMessagesJSONRequestBody defines body for GetApiMessages for application/json ContentType.
type GetApiMessagesJSONRequestBody = GetApiMessagesJSONBody

// PostApiTokenJSONRequestBody defines body for PostApiToken for application/json ContentType.
type PostApiTokenJSONRequestBody = PostApiTokenJSONBody

// PostApiTopicTopicSubscribeJSONRequestBody defines body for PostApiTopicTopicSubscribe for application/json ContentType.
type PostApiTopicTopicSubscribeJSONRequestBody = PostApiTopicTopicSubscribeJSONBody

// PostApiTopicTopicUnsubscribeJSONRequestBody defines body for PostApiTopicTopicUnsubscribe for application/json ContentType.
type PostApiTopicTopicUnsubscribeJSONRequestBody = PostApiTopicTopicUnsubscribeJSONBody

// DeleteApiUserJSONRequestBody defines body for DeleteApiUser for application/json ContentType.
type DeleteApiUserJSONRequestBody = DeleteApiUserJSONBody

// PutApiUserJSONRequestBody defines body for PutApiUser for application/json ContentType.
type PutApiUserJSONRequestBody = PutApiUserJSONBody
