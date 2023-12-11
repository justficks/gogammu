package gammu

import (
	"fmt"
	"time"
)

type Inbox struct {
	tableName         struct{}  `pg:"inbox"`
	UpdatedInDB       time.Time `pg:"UpdatedInDB,default:now()"`
	ReceivingDateTime time.Time `pg:"ReceivingDateTime,default:now()"`
	//Text              string    `pg:"Text,notnull"`
	SenderNumber string `pg:"SenderNumber,type:varchar(20),default:'',notnull"`
	Coding       string `pg:"Coding,type:varchar(255),default:'Default_No_Compression',notnull,use_zero"`
	UDH          string `pg:"UDH,notnull"`
	SMSCNumber   string `pg:"SMSCNumber,type:varchar(20),default:'',notnull"`
	Class        int    `pg:"Class,default:-1,notnull"`
	TextDecoded  string `pg:"TextDecoded,default:'',notnull"`
	ID           int    `pg:"ID,pk"`
	RecipientID  string `pg:"RecipientID,notnull"`
	Processed    bool   `pg:"Processed,default:false,notnull"`
	Status       int    `pg:"Status,default:-1,notnull"`
}

type Sentitems struct {
	tableName         struct{}   `pg:"sentitems"`
	UpdatedInDB       time.Time  `pg:"UpdatedInDB,default:now()"`
	InsertIntoDB      time.Time  `pg:"InsertIntoDB,default:now()"`
	SendingDateTime   time.Time  `pg:"SendingDateTime,default:now()"`
	DeliveryDateTime  *time.Time `pg:"DeliveryDateTime"`
	Text              string     `pg:"Text,notnull"`
	DestinationNumber string     `pg:"DestinationNumber,type:varchar(20),default:'',notnull"`
	Coding            string     `pg:"Coding,type:varchar(255),default:'Default_No_Compression',notnull,use_zero"`
	UDH               string     `pg:"UDH,notnull"`
	SMSCNumber        string     `pg:"SMSCNumber,type:varchar(20),default:'',notnull"`
	Class             int        `pg:"Class,default:-1,notnull"`
	TextDecoded       string     `pg:"TextDecoded,default:'',notnull"`
	ID                int        `pg:"ID,pk"`
	SenderID          string     `pg:"SenderID,notnull"`
	SequencePosition  int        `pg:"SequencePosition,default:1,notnull"`
	Status            string     `pg:"Status,type:varchar(255),default:'SendingOK',notnull"`
	StatusError       int        `pg:"StatusError,default:-1,notnull"`
	TPMR              int        `pg:"TPMR,default:-1,notnull"`
	RelativeValidity  int        `pg:"RelativeValidity,default:-1,notnull"`
	CreatorID         string     `pg:"CreatorID,notnull"`
	StatusCode        int        `pg:"StatusCode,default:-1,notnull"`
}

type Phones struct {
	tableName    struct{}  `pg:"phones"`
	ID           int       `pg:"ID,notnull"`
	UpdatedInDB  time.Time `pg:"UpdatedInDB,default:now()"`
	InsertIntoDB time.Time `pg:"InsertIntoDB,default:now()"`
	TimeOut      time.Time `pg:"TimeOut,default:now()"`
	Send         bool      `pg:"Send,default:false,notnull"`
	Receive      bool      `pg:"Receive,default:false,notnull"`
	IMEI         string    `pg:"IMEI,type:varchar(35),pk,notnull"`
	IMSI         string    `pg:"IMSI,type:varchar(35),notnull"`
	NetCode      string    `pg:"NetCode,type:varchar(10),default:'ERROR'"`
	NetName      string    `pg:"NetName,type:varchar(35),default:'ERROR'"`
	Client       string    `pg:"Client,notnull"`
	Battery      int       `pg:"Battery,default:-1,notnull"`
	Signal       int       `pg:"Signal,default:-1,notnull"`
	Sent         int       `pg:"Sent,default:0,notnull"`
	Received     int       `pg:"Received,default:0,notnull"`
}

type PhonesIMSI struct {
	tableName struct{} `pg:"phones_imsi"`
	ID        int      `pg:"ID,notnull"`
	IMSI      string   `pg:"IMSI,notnull"`
	Phone     string   `pg:"Phone,notnull"`
}

func (g *Gammu) GetInbox(page int, pageSize int) ([]Inbox, error) {
	var items []Inbox
	offset := (page - 1) * pageSize
	err := g.DB.Model(&items).
		Where("\"TextDecoded\" != ?", "").
		Order("ID DESC").
		Offset(offset).
		Limit(pageSize).
		Select()
	if err != nil {
		return nil, err
	}
	return items, err
}

func (g *Gammu) GetInboxCount() (int, error) {
	count, err := g.DB.Model((*Inbox)(nil)).
		Where("\"TextDecoded\" != ?", "").
		Count()
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (g *Gammu) DeleteInbox(id int) error {
	inbox := &Inbox{ID: id}
	_, err := g.DB.Model(inbox).Where("\"ID\" = ?", id).Delete()
	if err != nil {
		return err
	}
	return nil
}

func (g *Gammu) DeleteMonitor(id int) error {
	inbox := &Phones{ID: id}
	_, err := g.DB.Model(inbox).Where("\"ID\" = ?::text", id).Delete()
	if err != nil {
		return err
	}
	return nil
}

func (g *Gammu) GetOutbox() ([]Sentitems, error) {
	var items []Sentitems
	err := g.DB.Model(&items).Select()
	if err != nil {
		return nil, err
	}
	return items, err
}

func (g *Gammu) DeleteOutbox(id int) error {
	outbox := &Sentitems{ID: id}
	_, err := g.DB.Model(outbox).Where("\"ID\" = ?", id).Delete()
	if err != nil {
		return err
	}
	return nil
}

func (g *Gammu) GetPhones() ([]Phones, error) {
	var items []Phones
	err := g.DB.Model(&items).Select()
	if err != nil {
		return nil, err
	}
	return items, err
}

func (g *Gammu) GetPhonesIMSI() ([]PhonesIMSI, error) {
	var items []PhonesIMSI
	err := g.DB.Model(&items).Select()
	if err != nil {
		return nil, err
	}
	return items, err
}

func (g *Gammu) AddPhoneIMSI(input PhonesIMSI) error {
	if input.Phone == "" || input.IMSI == "" {
		return fmt.Errorf("phone or IMSI is empty")
	}
	_, err := g.DB.Model(&input).Insert()
	if err != nil {
		return err
	}
	return nil
}

func (g *Gammu) UpdatePhoneIMSI(id int, newPhone string) error {
	var phone PhonesIMSI
	_, err := g.DB.Model(&phone).Set("\"Phone\" = ?", newPhone).Where("\"ID\" = ?", id).Update()
	if err != nil {
		return err
	}
	return nil
}

func (g *Gammu) DeletePhoneIMSI(id int) error {
	phone := &PhonesIMSI{ID: id}
	_, err := g.DB.Model(phone).Where("\"ID\" = ?", id).Delete()
	if err != nil {
		return err
	}
	return nil
}
