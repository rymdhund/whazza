package monitor

import (
	"errors"
	"fmt"
	"net/smtp"
	"time"

	"github.com/rymdhund/whazza/internal/base"
	"github.com/rymdhund/whazza/internal/hubutil"
	. "github.com/rymdhund/whazza/internal/logging"
	"github.com/rymdhund/whazza/internal/persist"
)

// Monitor the checks for timeouts
type Monitor struct {
	cfg          hubutil.HubConfig
	hubStartTime time.Time
}

type Mailer struct {
	host     string
	port     int
	user     string
	password string
	from     string
}

func New(cfg hubutil.HubConfig, hubStartTime time.Time) *Monitor {
	return &Monitor{cfg, hubStartTime}
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
	db, err := persist.Open(m.cfg.Database())
	if err != nil {
		return err
	}
	defer db.Close()

	expChecks, err := getExpiredChecks(db, m.hubStartTime)
	if err != nil {
		return err
	}
	for _, check := range expChecks {

		lastStatus, err := db.LastNotification(check.ID)
		if err != nil {
			return err
		}

		if lastStatus != "expired" {
			err := m.notify(db, check, base.ExpiredResult())
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *Monitor) HandleResult(check persist.CheckModel, res persist.ResultModel) error {
	db, err := persist.Open(m.cfg.Database())
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
	InfoLog.Printf("Notification [%s] %+v", res.Status, check)
	subj := fmt.Sprintf("%s %s", check.Check.Title(), res.Status)
	body := fmt.Sprintf("%+v\n[%s] - %s", check, res.Status, res.Msg)

	if m.cfg.NotifyEmail != "" {
		mailer, err := m.mkMailer()
		if err != nil {
			WarningLog.Printf("Not sending mail since: %s", err)
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
	DebugLog.Printf("Sending notification email to %s", to)
	auth := smtp.PlainAuth("", m.user, m.password, m.host)
	addr := fmt.Sprintf("%s:%d", m.host, m.port)

	msg := fmt.Sprintf("To: %s\r\n", to)
	msg += fmt.Sprintf("Subject: %s\r\n", subject)
	msg += "\r\n"
	msg += fmt.Sprintf("%s\r\n", body)

	err := smtp.SendMail(addr, auth, m.from, []string{to}, []byte(msg))
	if err != nil {
		return fmt.Errorf("sendMail: failed with %w", err)
	}
	return nil
}

func maxTime(t1 time.Time, t2 time.Time) time.Time {
	if t1.After(t2) {
		return t1
	} else {
		return t2
	}
}

func getExpiredChecks(db *persist.DB, hubStart time.Time) ([]persist.CheckModel, error) {
	overviews, err := db.GetCheckOverviews()
	if err != nil {
		return nil, err
	}

	expired := []persist.CheckModel{}
	for _, ov := range overviews {
		if ov.Result.Status == "expired" {
			// If the server is newly started we give checks a chance to report in
			// Do this by pretending we got a result right before the hub started
			t := maxTime(ov.LastReceived.Timestamp, hubStart)
			if ov.CheckModel.Check.IsExpired(t, time.Now()) {
				expired = append(expired, ov.CheckModel)
			}
		}
	}

	return expired, nil
}
