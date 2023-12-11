package gammu

import (
	"fmt"
	"github.com/go-pg/pg/v10"
	"time"
)

func (g *Gammu) ConcatSMS(n *Notify) (string, error) {
	var messages []Inbox
	err := g.DB.Model(&messages).Where("\"ID\" in (?)", pg.In(n.MessageIDs)).Select()
	if err != nil {
		return "", fmt.Errorf("неудача получения сообщений с id: %s по причине: %s", n.MessageIDs, err.Error())
	}

	if len(messages) == 0 {
		return "", fmt.Errorf("сообщения в базе данных не найдены %s %s", n.PhoneID, n.MessageIDs)
	}

	m := messages[0]

	phoneNumber := n.PhoneID

	var model PhonesIMSI
	err = g.DB.Model(&model).Where("\"IMSI\" = ?", n.PhoneID).Select()
	if err == nil && model.Phone != "" {
		phoneNumber = model.Phone
	} else {
		modem, ok := g.Store.GetModemByIMSI(m.RecipientID)
		if ok {
			phoneNumber = fmt.Sprintf("Modem%d", modem.Num)
		}
	}

	smsTime := m.ReceivingDateTime.Add(7 * time.Hour).Format("2006.01.02 15:04:05")
	smsFromTo := fmt.Sprintf("%s  >  %s", m.SenderNumber, phoneNumber)

	combinedText := fmt.Sprintf("%s\n%s\n\n", smsTime, smsFromTo)
	for _, message := range messages {
		combinedText += message.TextDecoded + " "
	}

	return combinedText, nil
}
