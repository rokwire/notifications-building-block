package rest

import "notifications/core/model"

// RecipientCriteria Type
func recipientsCriteriaListFromDef(items []model.InputRecipientCriteria) []model.RecipientCriteria {
	criteriaList := make([]model.RecipientCriteria, len(items))
	for i, item := range items {
		criteriaList[i] = model.RecipientCriteria{AppVersion: item.AppVersion, AppPlatform: item.AppPlatform}
	}
	return criteriaList
}

// MessageRecipient Type
func messagesRecipientsListFromDef(items []model.InputMessageRecipient) []model.MessageRecipient {
	result := make([]model.MessageRecipient, len(items))
	for i, item := range items {
		result[i] = model.MessageRecipient{UserID: item.UserID, Mute: item.Mute}
	}
	return result
}
