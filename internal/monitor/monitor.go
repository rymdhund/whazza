package monitor

import (
	"errors"
	"fmt"
	"log"
	"net/smtp"
	"time"

	"github.com/rymdhund/whazza/internal/base"
	"github.com/rymdhund/whazza/internal/hubutil"
	"github.com/rymdhund/whazza/internal/persist"
)

type Monitor struct {
	cfg hubutil.HubConfig
}

type Mailer struct {
	host     string
	port     int
	user     string
	password string
	from     string
}

func New(cfg hubutil.HubConfig) *Monitor {
	return &Monitor{cfg}
}

func (m *Monitor) mkMailer() (Mailer, error) {
	if m.cfg.SMTPHost == "" {
		return Mailer{}, errors.New("No smtphost configured")
	}
	if m.cfg.SMTPPort == 0 {
		return Mailer{}, errors.New("No smtpport configured")
	}
	if m.cfg.SMTPUser == "" {
		return Mailer{}, errors.New("No smtpuser configured")
	}
	if m.cfg.SMTPPassword == "" {
		return Mailer{}, errors.New("No smtppassword configured")
	}
	if m.cfg.SMTPFrom == "" {
		return Mailer{}, errors.New("No smtpfrom configured")
	}
	return Mailer{
		host:     m.cfg.SMTPHost,
		port:     m.cfg.SMTPPort,
		user:     m.cfg.SMTPUser,
		password: m.cfg.SMTPPassword,
		from:     m.cfg.SMTPFrom,
	}, nil
}

func (m *Monitor) CheckForExpired() error {
	db, err := persist.Open("./whazza.db")
	if err != nil {
		return err
	}
	defer db.Close()

	expChecks, err := db.GetExpiredChecks()
	if err != nil {
		return err
	}
	for _, chk := range expChecks {
		lastStatus, err := db.LastNotification(chk.ID)
		if err != nil {
			return err
		}

		if lastStatus != "expired" {
			err := m.notify(db, chk, base.Result{Status: "expired", Timestamp: time.Now()})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *Monitor) HandleResult(check persist.CheckModel, res persist.ResultModel) error {
	db, err := persist.Open("./whazza.db")
	if err != nil {
		return err
	}
	defer db.Close()

	oldStatus, err := db.LastNotification(check.ID)
	if err != nil {
		return err
	}

	// we treat no notificated statuses yet as "good"
	if oldStatus == "" {
		oldStatus = "good"
	}

	if res.Status != oldStatus {
		err := m.notify(db, check, res.Result)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Monitor) notify(db *persist.DB, check persist.CheckModel, res base.Result) error {
	log.Printf("Notification [%s] %+v", res.Status, check)
	subj := fmt.Sprintf("%s %s", check.Name(), res.Status)
	body := fmt.Sprintf("Notification\n\n [%s] %+v", res.Status, check)

	if m.cfg.NotifyEmail != "" {
		mailer, err := m.mkMailer()
		if err != nil {
			log.Printf("Not sending mail since: %s", err)
		} else {
			err := mailer.sendMail(m.cfg.NotifyEmail, subj, body)
			if err != nil {
				return err
			}
		}
	}
	err := db.AddNotification(check.ID, res.Status)
	return err
}

func (m Mailer) sendMail(to, subject, body string) error {
	log.Printf("Sending notification email to %s", to)
	auth := smtp.PlainAuth("", m.user, m.password, m.host)
	addr := fmt.Sprintf("%s:%d", m.host, m.port)

	msg := fmt.Sprintf("To: %s\r\n", to)
	msg += fmt.Sprintf("Subject: %s\r\n", subject)
	msg += "\r\n"
	msg += fmt.Sprintf("%s\r\n", body)

	err := smtp.SendMail(addr, auth, m.from, []string{to}, []byte(msg))
	if err != nil {
		return err
	}
	return nil
}
