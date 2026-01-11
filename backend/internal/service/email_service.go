package service

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

// SESEmailService implements EmailService using AWS SES
type SESEmailService struct {
	client       *ses.Client
	fromAddress  string
	replyTo      string
	templates    map[string]EmailTemplate
}

// EmailTemplate represents an email template
type EmailTemplate struct {
	Subject     string
	HTMLBody    string
	TextBody    string
}

// NewSESEmailService creates a new SES email service
func NewSESEmailService(client *ses.Client, fromAddress, replyTo string) *SESEmailService {
	service := &SESEmailService{
		client:      client,
		fromAddress: fromAddress,
		replyTo:     replyTo,
		templates:   make(map[string]EmailTemplate),
	}
	
	// Initialize default templates
	service.initializeTemplates()
	
	return service
}

// SendEmail sends a simple email
func (s *SESEmailService) SendEmail(
	ctx context.Context,
	to []string,
	subject string,
	body string,
	templateData map[string]interface{},
) error {
	log.Printf("Sending email to %v with subject: %s", to, subject)

	// Replace placeholders in body if templateData provided
	finalBody := body
	if templateData != nil {
		finalBody = s.replacePlaceholders(body, templateData)
	}

	input := &ses.SendEmailInput{
		Source: aws.String(s.fromAddress),
		Destination: &types.Destination{
			ToAddresses: to,
		},
		Message: &types.Message{
			Subject: &types.Content{
				Data:    aws.String(subject),
				Charset: aws.String("UTF-8"),
			},
			Body: &types.Body{
				Text: &types.Content{
					Data:    aws.String(finalBody),
					Charset: aws.String("UTF-8"),
				},
				Html: &types.Content{
					Data:    aws.String(s.wrapInHTML(finalBody)),
					Charset: aws.String("UTF-8"),
				},
			},
		},
		ReplyToAddresses: []string{s.replyTo},
	}

	result, err := s.client.SendEmail(ctx, input)
	if err != nil {
		log.Printf("ERROR: Failed to send email: %v", err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("Email sent successfully. Message ID: %s", *result.MessageId)
	return nil
}

// SendTemplatedEmail sends an email using a predefined template
func (s *SESEmailService) SendTemplatedEmail(
	ctx context.Context,
	to []string,
	templateName string,
	data map[string]interface{},
) error {
	log.Printf("Sending templated email (%s) to %v", templateName, to)

	template, exists := s.templates[templateName]
	if !exists {
		log.Printf("WARNING: Template '%s' not found, using default template", templateName)
		template = s.getDefaultTemplate()
	}

	// Replace placeholders in template
	subject := s.replacePlaceholders(template.Subject, data)
	htmlBody := s.replacePlaceholders(template.HTMLBody, data)
	textBody := s.replacePlaceholders(template.TextBody, data)

	input := &ses.SendEmailInput{
		Source: aws.String(s.fromAddress),
		Destination: &types.Destination{
			ToAddresses: to,
		},
		Message: &types.Message{
			Subject: &types.Content{
				Data:    aws.String(subject),
				Charset: aws.String("UTF-8"),
			},
			Body: &types.Body{
				Text: &types.Content{
					Data:    aws.String(textBody),
					Charset: aws.String("UTF-8"),
				},
				Html: &types.Content{
					Data:    aws.String(htmlBody),
					Charset: aws.String("UTF-8"),
				},
			},
		},
		ReplyToAddresses: []string{s.replyTo},
	}

	result, err := s.client.SendEmail(ctx, input)
	if err != nil {
		log.Printf("ERROR: Failed to send templated email: %v", err)
		return fmt.Errorf("failed to send templated email: %w", err)
	}

	log.Printf("Templated email sent successfully. Message ID: %s", *result.MessageId)
	return nil
}

// Helper methods

func (s *SESEmailService) replacePlaceholders(text string, data map[string]interface{}) string {
	result := text
	for key, value := range data {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}
	return result
}

func (s *SESEmailService) wrapInHTML(text string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
		.container { max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { background-color: #3b82f6; color: white; padding: 20px; text-align: center; }
		.content { padding: 20px; background-color: #f9fafb; }
		.footer { padding: 20px; text-align: center; font-size: 12px; color: #666; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h2>Your Company HRMS</h2>
		</div>
		<div class="content">
			%s
		</div>
		<div class="footer">
			<p>This is an automated message. Please do not reply to this email.</p>
		</div>
	</div>
</body>
</html>
	`, strings.ReplaceAll(text, "\n", "<br>"))
}

func (s *SESEmailService) getDefaultTemplate() EmailTemplate {
	return EmailTemplate{
		Subject:  "Background Check Notification",
		HTMLBody: "<p>{{message}}</p>",
		TextBody: "{{message}}",
	}
}

// Initialize email templates
func (s *SESEmailService) initializeTemplates() {
	// Background Check Initiated Template
	s.templates[TemplateCheckInitiated] = EmailTemplate{
		Subject: "Background Check Initiated",
		HTMLBody: `
<h2>Background Check Initiated</h2>
<p>Dear {{first_name}} {{last_name}},</p>
<p>We are writing to inform you that a background check has been initiated as part of your employment process with {{company_name}}.</p>
<h3>What to Expect:</h3>
<ul>
	<li><strong>Check Types:</strong> {{check_types}}</li>
	<li><strong>Estimated Completion:</strong> {{estimated_eta}}</li>
</ul>
<p>You may be contacted by our background check provider to verify information or provide additional details. Please respond promptly to any requests to avoid delays.</p>
<p>If you have any questions, please contact us at {{support_email}}.</p>
<p>Thank you for your cooperation.</p>
<p>Best regards,<br>Human Resources Team</p>
		`,
		TextBody: `Background Check Initiated

Dear {{first_name}} {{last_name}},

We are writing to inform you that a background check has been initiated as part of your employment process with {{company_name}}.

What to Expect:
- Check Types: {{check_types}}
- Estimated Completion: {{estimated_eta}}

You may be contacted by our background check provider to verify information or provide additional details. Please respond promptly to any requests to avoid delays.

If you have any questions, please contact us at {{support_email}}.

Thank you for your cooperation.

Best regards,
Human Resources Team`,
	}

	// Background Check Completed Template
	s.templates[TemplateCheckCompleted] = EmailTemplate{
		Subject: "Background Check Completed",
		HTMLBody: `
<h2>Background Check Completed</h2>
<p>Dear {{first_name}} {{last_name}},</p>
<p>Your background check has been completed on {{completed_date}}.</p>
<p><strong>Result:</strong> {{result}}</p>
<p>You can view your full report at: <a href="{{report_url}}">View Report</a></p>
<p>If you have any questions or concerns about your results, please contact us at {{support_email}}.</p>
<p>Best regards,<br>Human Resources Team</p>
		`,
		TextBody: `Background Check Completed

Dear {{first_name}} {{last_name}},

Your background check has been completed on {{completed_date}}.

Result: {{result}}

You can view your full report at: {{report_url}}

If you have any questions or concerns about your results, please contact us at {{support_email}}.

Best regards,
Human Resources Team`,
	}

	// Pre-Adverse Action Template (FCRA Required)
	s.templates[TemplateAdverseActionPre] = EmailTemplate{
		Subject: "Important: Pre-Adverse Action Notice",
		HTMLBody: `
<h2>Pre-Adverse Action Notice</h2>
<p>Dear {{first_name}} {{last_name}},</p>
<p>We are writing to inform you that information in a consumer report may result in an adverse employment decision. This is a <strong>pre-adverse action notice</strong> as required by the Fair Credit Reporting Act (FCRA).</p>

<h3>What This Means:</h3>
<p>Before we make a final decision, you have the right to review the report and dispute any information you believe is inaccurate or incomplete.</p>

<h3>Your Rights:</h3>
<ul>
	<li>You have the right to obtain a copy of the report</li>
	<li>You have the right to dispute inaccurate information</li>
	<li>You have {{dispute_deadline}} to respond</li>
</ul>

<h3>Report Information:</h3>
<p><strong>Consumer Reporting Agency:</strong> {{provider_name}}<br>
<strong>Address:</strong> {{provider_address}}<br>
<strong>Phone:</strong> {{provider_phone}}<br>
<strong>Website:</strong> <a href="{{provider_website}}">{{provider_website}}</a></p>

<p><a href="{{report_url}}">View Your Report</a></p>

<h3>What You Should Do:</h3>
<ol>
	<li>Review the report carefully</li>
	<li>If you find any errors, contact the consumer reporting agency directly</li>
	<li>Contact us at {{support_email}} if you have questions</li>
</ol>

<p><strong>Important:</strong> The consumer reporting agency did not make the decision to take adverse action and cannot explain why. They only provided the report.</p>

<h3>Summary of Your Rights Under FCRA:</h3>
<div style="background-color: #f3f4f6; padding: 15px; margin: 15px 0; font-size: 12px;">
	<pre>{{fcra_summary_of_rights}}</pre>
</div>

<p>Please contact us at {{support_email}} if you have any questions.</p>

<p>Sincerely,<br>Human Resources Team<br>{{company_name}}</p>
		`,
		TextBody: `PRE-ADVERSE ACTION NOTICE

Dear {{first_name}} {{last_name}},

We are writing to inform you that information in a consumer report may result in an adverse employment decision. This is a pre-adverse action notice as required by the Fair Credit Reporting Act (FCRA).

WHAT THIS MEANS:
Before we make a final decision, you have the right to review the report and dispute any information you believe is inaccurate or incomplete.

YOUR RIGHTS:
- You have the right to obtain a copy of the report
- You have the right to dispute inaccurate information  
- You have until {{dispute_deadline}} to respond

REPORT INFORMATION:
Consumer Reporting Agency: {{provider_name}}
Address: {{provider_address}}
Phone: {{provider_phone}}
Website: {{provider_website}}

View Your Report: {{report_url}}

WHAT YOU SHOULD DO:
1. Review the report carefully
2. If you find any errors, contact the consumer reporting agency directly
3. Contact us at {{support_email}} if you have questions

IMPORTANT: The consumer reporting agency did not make the decision to take adverse action and cannot explain why. They only provided the report.

SUMMARY OF YOUR RIGHTS UNDER FCRA:
{{fcra_summary_of_rights}}

Please contact us at {{support_email}} if you have any questions.

Sincerely,
Human Resources Team
{{company_name}}`,
	}

	// Background Check Failed Template
	s.templates[TemplateCheckFailed] = EmailTemplate{
		Subject: "Background Check Status Update",
		HTMLBody: `
<h2>Background Check Status Update</h2>
<p>Dear {{first_name}} {{last_name}},</p>
<p>We wanted to update you on the status of your background check.</p>
<p>Unfortunately, we encountered an issue during the verification process. Our HR team will contact you shortly to discuss next steps.</p>
<p>If you have any immediate questions, please contact us at {{support_email}}.</p>
<p>Best regards,<br>Human Resources Team</p>
		`,
		TextBody: `Background Check Status Update

Dear {{first_name}} {{last_name}},

We wanted to update you on the status of your background check.

Unfortunately, we encountered an issue during the verification process. Our HR team will contact you shortly to discuss next steps.

If you have any immediate questions, please contact us at {{support_email}}.

Best regards,
Human Resources Team`,
	}
}