package telegram

func MapUpdateToIncomingMessage(update Update) (IncomingMessage, bool) {
	if update.Message == nil {
		return IncomingMessage{}, false
	}

	return IncomingMessage{
		UpdateID:  update.UpdateID,
		MessageID: update.Message.MessageID,
		ChatID:    update.Message.Chat.ID,
		Text:      update.Message.Text,
	}, true
}
