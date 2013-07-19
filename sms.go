// sms.go
package sms77

import (
	"io/ioutil"
	//"fmt"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

var Smstype string = TYPE_BASICPLUS
var Debug bool = false
var Status bool = true
var From string = ""
var User string = ""
var Password string = ""

var FillReturnchan bool = true
var Returnchan chan string

var baseurl string = "https://gateway.sms77.de"

func init() {
	Returnchan = make(chan string, 15)
}

var Errors map[string]string = map[string]string{
	"100": "SMS wurde erfolgreich verschickt",
	"101": "Versand an mindestens einen Empfänger fehlgeschlagen",
	"151": "Kontakt nicht gefunden",
	"152": "Änderung gespeichert",
	"153": "Änderung nicht gespeichert",
	"201": "Ländercode für diesen SMS-Typ nicht gültig. Bitte als Basic SMS verschicken.",
	"202": "Empfängernummer ungültig",
	"300": "Bitte Benutzer/Passwort angeben",
	"301": "Variable to nicht gesetzt",
	"304": "Variable type nicht gesetzt",
	"305": "Variable text nicht gesetzt",
	"306": "Absendernummer ungültig, Diese muss vom Format 0049... sein und eine gültige, Handynummer darstellen.",
	"307": "Variable url nicht gesetzt",
	"400": "type ungültig. Siehe erlaubte Werte oben.",
	"401": "Variable text ist zu lang",
	"402": "Reloadsperre – diese SMS wurde bereits innerhalb der letzten 90 Sekunden verschickt",
	"500": "Zu wenig Guthaben vorhanden.",
	"600": "Carrier Zustellung misslungen",
	"700": "Unbekannter Fehler",
	"801": "Logodatei nicht angegeben",
	"802": "Logodatei existiert nicht",
	"803": "Klingelton nicht angegeben",
	"900": "Benutzer/Passwort-Kombination falsch",
	"901": "Ungültige Msg ID",
	"902": "http API für diesen Account deaktiviert",
	"903": "Server IP ist falsch",
}

const (
	TYPE_BASICPLUS = "basicplus"
	TYPE_QUALITY   = "quality"
	TYPE_FESTNETZ  = "festnetz"
	TYPE_FLASH     = "flash"
)

type Sms struct {
	To    string
	Text  string
	Delay string
}

type smsurl struct {
	url  *url.URL
	vals url.Values
}

type PhonebookEntry struct {
	Id         string
	Nick       string
	Empfaenger string
	Email      string
}

func beginUrl() smsurl {
	u, _ := url.Parse(baseurl)
	vals := make(url.Values)
	vals.Add("u", User)
	vals.Add("p", Password)
	if Debug {
		vals.Add("debug", "1")
	}
	return smsurl{u, vals}
}

func (s smsurl) String() string {
	s.url.RawQuery = s.vals.Encode()
	return s.url.String()
}

func Balance() string {
	u := beginUrl()
	u.url.Path = "/balance.php"
	resp, _ := http.Get(u.String())
	b, _ := ioutil.ReadAll(resp.Body)
	return string(b)
}

func Sendsms(SMS *Sms) string {

	u := beginUrl()

	u.vals.Add("to", SMS.To)
	u.vals.Add("text", SMS.Text)
	u.vals.Add("from", From)
	u.vals.Add("type", Smstype)
	if Status {
		u.vals.Add("status", "1")
	}

	resp, _ := http.Get(u.String())
	//split := strings.SplitN(string(b), " ", 2)
	//return Errors[split[0]]
	b, _ := ioutil.ReadAll(resp.Body)
	return string(b)
}

func SmsStatus(msgid string) string {
	u := beginUrl()
	u.url.Path = "/status.php"
	u.vals.Add("msg_id", msgid)

	resp, _ := http.Get(u.String())
	b, _ := ioutil.ReadAll(resp.Body)
	return string(b)
}

func GetPhonebookEntries(searchvalues PhonebookEntry) []PhonebookEntry {
	u := beginUrl()
	u.url.Path = "/adress.php"
	u.vals.Add("action", "read")

	if searchvalues.Email != "" {
		u.vals.Add("email", searchvalues.Email)
	}

	if searchvalues.Empfaenger != "" {
		u.vals.Add("empfaenger", searchvalues.Empfaenger)
	}

	if searchvalues.Id != "" {
		u.vals.Add("id", searchvalues.Id)
	}

	if searchvalues.Nick != "" {
		u.vals.Add("nick", searchvalues.Nick)
	}

	resp, _ := http.Get(u.String())
	b, _ := ioutil.ReadAll(resp.Body)

	if _, err := strconv.Atoi(string(b)); err == nil {
		return nil
	} else {
		ret := make([]PhonebookEntry, 0, 0)
		for _, val := range strings.Split(string(b), "\r\n") {
			v := strings.Split(val, ";")
			p := PhonebookEntry{}
			p.Id = v[0]
			p.Nick = v[1]
			p.Empfaenger = v[2]
			p.Email = v[3]
			ret = append(ret, p)
		}
		return ret
	}
}

func EditOrNewPhonebookEntry(value PhonebookEntry) error {
	u := beginUrl()
	u.url.Path = "/adress.php"
	u.vals.Add("action", "write")

	if value.Email != "" {
		u.vals.Add("email", value.Email)
	}

	if value.Empfaenger != "" {
		u.vals.Add("empfaenger", value.Empfaenger)
	}

	if value.Id != "" {
		u.vals.Add("id", value.Id)
	}

	if value.Nick != "" {
		u.vals.Add("nick", value.Nick)
	}

	resp, _ := http.Get(u.String())
	b, _ := ioutil.ReadAll(resp.Body)

	if strings.HasPrefix(string(b), "152") {
		return nil
	} else {
		return errors.New("153")
	}
}

func DelPhonebookEntry(id string) error {
	u := beginUrl()
	u.url.Path = "/adress.php"
	u.vals.Add("action", "del")
	u.vals.Add("id", id)

	resp, _ := http.Get(u.String())
	b, _ := ioutil.ReadAll(resp.Body)

	if string(b) == "152" {
		return nil
	} else {
		return errors.New("153")
	}
}
