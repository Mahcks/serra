package email

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"net"
	"net/smtp"

	"github.com/mahcks/serra/pkg/structures"
)

type Service struct {
	settings *structures.EmailSettings
	auth     smtp.Auth
}

func NewService(settings *structures.EmailSettings) *Service {
	if !settings.Enabled {
		return &Service{settings: settings}
	}

	var auth smtp.Auth
	if settings.SMTPUsername != "" && settings.SMTPPassword != "" {
		auth = smtp.PlainAuth("", settings.SMTPUsername, settings.SMTPPassword, settings.SMTPHost)
	}

	return &Service{
		settings: settings,
		auth:     auth,
	}
}

func (s *Service) IsEnabled() bool {
	return s.settings != nil && s.settings.Enabled
}

func (s *Service) SendInvitation(data structures.InvitationEmailData) error {
	if !s.IsEnabled() {
		return fmt.Errorf("email service is not enabled")
	}

	subject := fmt.Sprintf("You're invited to join %s", data.AppName)
	
	htmlBody, err := s.renderInvitationHTML(data)
	if err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	textBody := s.renderInvitationText(data)

	return s.sendEmail(data.Username, subject, htmlBody, textBody)
}

func (s *Service) sendEmail(to, subject, htmlBody, textBody string) error {
	from := s.settings.SenderAddress
	fromName := s.settings.SenderName
	if fromName == "" {
		fromName = "Serra"
	}

	// Construct message
	msg := bytes.Buffer{}
	msg.WriteString(fmt.Sprintf("From: %s <%s>\r\n", fromName, from))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: multipart/alternative; boundary=\"boundary123\"\r\n")
	msg.WriteString("\r\n")
	
	// Text part
	msg.WriteString("--boundary123\r\n")
	msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(textBody)
	msg.WriteString("\r\n")
	
	// HTML part
	msg.WriteString("--boundary123\r\n")
	msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(htmlBody)
	msg.WriteString("\r\n")
	msg.WriteString("--boundary123--\r\n")

	addr := fmt.Sprintf("%s:%d", s.settings.SMTPHost, s.settings.SMTPPort)
	
	// Handle different encryption methods
	if s.settings.EncryptionMethod == "starttls" || s.settings.UseSTARTTLS {
		return s.sendMailWithSTARTTLS(addr, s.auth, from, []string{to}, msg.Bytes())
	}
	
	return smtp.SendMail(addr, s.auth, from, []string{to}, msg.Bytes())
}

func (s *Service) sendMailWithSTARTTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	// Connect to the SMTP server
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	// Create SMTP client
	client, err := smtp.NewClient(conn, s.settings.SMTPHost)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Quit()

	// Start TLS if supported
	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsConfig := &tls.Config{
			ServerName:         s.settings.SMTPHost,
			InsecureSkipVerify: s.settings.AllowSelfSigned,
		}
		if err = client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("failed to start TLS: %w", err)
		}
	}

	// Authenticate if credentials are provided
	if auth != nil {
		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("failed to authenticate: %w", err)
		}
	}

	// Set sender
	if err = client.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipients
	for _, recipient := range to {
		if err = client.Rcpt(recipient); err != nil {
			return fmt.Errorf("failed to set recipient %s: %w", recipient, err)
		}
	}

	// Send message
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}
	defer writer.Close()

	if _, err = writer.Write(msg); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}

func (s *Service) renderInvitationHTML(data structures.InvitationEmailData) (string, error) {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Welcome to {{.AppName}}</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 30px; text-align: center; border-radius: 8px; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 8px; margin: 20px 0; }
        .button { display: inline-block; background: #667eea; color: white; padding: 12px 30px; text-decoration: none; border-radius: 6px; margin: 20px 0; }
        .button:hover { background: #5a6fd8; }
        .info-box { background: #e3f2fd; border-left: 4px solid #2196f3; padding: 15px; margin: 20px 0; }
        .footer { text-align: center; color: #666; font-size: 12px; margin-top: 30px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>🎬 Welcome to {{.AppName}}!</h1>
        <p>You've been invited by {{.InviterName}}</p>
    </div>
    
    <div class="content">
        <h2>Hi {{.Username}}!</h2>
        
        <p>Great news! {{.InviterName}} has invited you to join {{.AppName}}, our media request and management system.</p>
        
        <p><strong>What you'll be able to do:</strong></p>
        <ul>
            <li>🔍 Search and discover movies and TV shows</li>
            <li>📝 Request content you want to watch</li>
            <li>📊 Track your request status</li>
            <li>🎯 Access our {{.MediaServerName}} media server</li>
        </ul>
        
        <div style="text-align: center;">
            <a href="{{.AcceptURL}}" class="button">Accept Invitation & Set Password</a>
        </div>
        
        <div class="info-box">
            <strong>Account Details:</strong><br>
            <strong>Username:</strong> {{.Username}}<br>
            <strong>Media Server:</strong> {{.MediaServerURL}}<br>
            <strong>Invitation expires:</strong> {{.ExpiresAt}}
        </div>
        
        <p><strong>Next Steps:</strong></p>
        <ol>
            <li>Click the button above to accept your invitation</li>
            <li>Set your password (this will work for both systems)</li>
            <li>Start exploring and requesting content!</li>
        </ol>
        
        <p>If you didn't expect this invitation, you can safely ignore this email.</p>
    </div>
    
    <div class="footer">
        <p>This invitation was sent by {{.AppName}}<br>
        If you have any questions, please contact {{.InviterName}}</p>
    </div>
</body>
</html>`

	t, err := template.New("invitation").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, data)
	return buf.String(), err
}

func (s *Service) renderInvitationText(data structures.InvitationEmailData) string {
	return fmt.Sprintf(`Welcome to %s!

Hi %s,

You've been invited by %s to join %s, our media request and management system.

What you'll be able to do:
- Search and discover movies and TV shows
- Request content you want to watch  
- Track your request status
- Access our %s media server

Account Details:
Username: %s
Media Server: %s
Invitation expires: %s

To accept your invitation and set your password, visit:
%s

Next Steps:
1. Click the link above to accept your invitation
2. Set your password (this will work for both systems)
3. Start exploring and requesting content!

If you didn't expect this invitation, you can safely ignore this email.

--
This invitation was sent by %s
If you have any questions, please contact %s`,
		data.AppName,
		data.Username,
		data.InviterName,
		data.AppName,
		data.MediaServerName,
		data.Username,
		data.MediaServerURL,
		data.ExpiresAt,
		data.AcceptURL,
		data.AppName,
		data.InviterName)
}