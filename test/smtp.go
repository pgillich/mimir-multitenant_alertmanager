package test

import (
	"fmt"
	"io"
	"log/slog"

	smtp "github.com/emersion/go-smtp"
)

type SmtpBackend struct {
	Logger *slog.Logger
}

func (bkd *SmtpBackend) NewSession(conn *smtp.Conn) (smtp.Session, error) {
	bkd.Logger.Info("SMTPD_NewSession", "Hostname", conn.Hostname(), "LocalAddr", conn.Conn().LocalAddr().String(), "RemoteAddr", conn.Conn().RemoteAddr().String())
	return &SmtpSession{Logger: bkd.Logger}, nil
}

type SmtpSession struct {
	Logger *slog.Logger
	From   string
	To     []string
}

func (s *SmtpSession) Mail(from string, opts *smtp.MailOptions) error {
	s.Logger.Info("SMTPD_NewSession", "From", from)
	s.From = from
	return nil
}

func (s *SmtpSession) Rcpt(to string, opts *smtp.RcptOptions) error {
	s.Logger.Info("SMTPD_Rcpt", "To", to)
	s.To = append(s.To, to)
	return nil
}

func (s *SmtpSession) Data(r io.Reader) error {
	if b, err := io.ReadAll(r); err != nil {
		return err
	} else {
		s.Logger.Info("SMTPD_Data", "Message", string(b))

		// Here you would typically process the email
		return nil
	}
}

func (s *SmtpSession) AuthPlain(username, password string) error {
	s.Logger.Info("SMTPD_AuthPlain", "username", username, "password", password)
	if username != "testuser" || password != "testpass" {
		return fmt.Errorf("invalid username or password")
	}

	return nil
}

func (s *SmtpSession) Reset() {
	s.Logger.Info("SMTPD_Reset")
	s.From = ""
	s.To = []string{}
}

func (s *SmtpSession) Logout() error {
	s.Logger.Info("SMTPD_Logout")
	return nil
}
