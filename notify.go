package gammu

import (
	"fmt"
	"github.com/go-pg/pg/v10"
	"time"
)

type NewMsg struct {
	Phone string // Phone Number or SIM IMSI
	Text  string
	Date  time.Time
	From  string
}

func (g *Gammu) ConcatSMS(n *RunOnMsgBody) (*NewMsg, error) {
	var messages []Inbox
	err := g.DB.Model(&messages).Where("\"ID\" in (?)", pg.In(n.MessageIDs)).Select()
	if err != nil {
		return nil, fmt.Errorf("неудача получения сообщений с id: %s по причине: %s", n.MessageIDs, err.Error())
	}

	if len(messages) == 0 {
		return nil, fmt.Errorf("сообщения в базе данных не найдены %s %s", n.PhoneID, n.MessageIDs)
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

	text := ""
	for _, message := range messages {
		text += message.TextDecoded + " "
	}

	return &NewMsg{
		Phone: phoneNumber,
		Text:  text,
		Date:  m.ReceivingDateTime,
		From:  m.SenderNumber,
	}, nil
}
