package service

import (
	"bytes"
	"context"
	"embed"
	"html/template"
	"net/url"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

type EmailService struct {
	client   *ses.Client
	sender   string
	template *template.Template
}

type EmailData struct {
	Link string
}

var TemplateFS embed.FS

func NewEmailService() (*EmailService, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(os.Getenv("AWS_REGION")),
	)
	if err != nil {
		return nil, err
	}

	client := ses.NewFromConfig(cfg)
	sender := os.Getenv("EMAIL_AUTH_SENDER")
	if sender == "" {
		sender = "noreply@luna4.me"
	}

	tmpl, err := template.ParseFS(TemplateFS, "assets/templates/email-auth.html")
	if err != nil {
		return nil, err
	}

	return &EmailService{
		client:   client,
		sender:   sender,
		template: tmpl,
	}, nil
}

func (e *EmailService) SendAuthEmail(ctx context.Context, email, token string, redirect string) error {
	serviceURL := os.Getenv("SERVICE_URL")
	if serviceURL == "" {
		serviceURL = "localhost:8080"
	}

	authPath := os.Getenv("EMAIL_AUTH_PATH")
	if authPath == "" {
		authPath = "/auth/email/verify"
	}

	link := "https://" + serviceURL + authPath + "?token=" + token + "&email=" + url.QueryEscape(email)
	if redirect != "" {
		link += "&redirect=" + url.QueryEscape(redirect)
	}

	var body bytes.Buffer
	err := e.template.Execute(&body, EmailData{Link: link})
	if err != nil {
		return err
	}

	input := &ses.SendEmailInput{
		Source: aws.String(e.sender),
		Destination: &types.Destination{
			ToAddresses: []string{email},
		},
		Message: &types.Message{
			Subject: &types.Content{
				Data:    aws.String("Luna4 Authentication Request"),
				Charset: aws.String("UTF-8"),
			},
			Body: &types.Body{
				Html: &types.Content{
					Data:    aws.String(body.String()),
					Charset: aws.String("UTF-8"),
				},
			},
		},
	}

	_, err = e.client.SendEmail(ctx, input)
	return err
}
