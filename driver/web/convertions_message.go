package web

import (
	"notifications/core/model"
	Def "notifications/driver/web/docs/gen"
)

// RecipientCriteria Type
func recipientsCriteriaListFromDef(items []Def.SharedReqCreateMessageInputRecipientCriteria) []model.RecipientCriteria {
	criteriaList := make([]model.RecipientCriteria, len(items))
	for i, item := range items {
		criteriaList[i] = model.RecipientCriteria{AppVersion: item.AppVersion, AppPlatform: item.AppPlatform}
	}
	return criteriaList
}

// MessageRecipient Type
func messagesRecipientsListFromDef(items []Def.SharedReqCreateMessageInputMessageRecipient) []model.MessageRecipient {
	result := make([]model.MessageRecipient, len(items))
	for i, item := range items {
		result[i] = model.MessageRecipient{UserID: item.UserId, Mute: item.Mute}
	}
	return result
}
