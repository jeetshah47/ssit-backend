package email

// otpEmailTemplate is the HTML template for OTP verification emails
// Based on the Figma design provided
const otpEmailTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Verify your email - Equitywala</title>
    <style>
        body {
            margin: 0;
            padding: 0;
            font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            background-color: #fafafa;
            color: #414651;
        }
        .email-container {
            max-width: 640px;
            margin: 0 auto;
            background-color: #ffffff;
            padding: 32px;
        }
        .email-header {
            padding: 24px;
            background-color: #ffffff;
        }
        .logo {
            font-size: 20px;
            font-weight: 600;
            color: #202124;
            text-decoration: none;
        }
        .email-body {
            padding: 32px 24px;
        }
        .greeting {
            font-size: 16px;
            line-height: 24px;
            color: #414651;
            margin-bottom: 16px;
        }
        .otp-container {
            margin: 24px 0;
        }
        .otp-boxes {
            display: flex;
            gap: 8px;
            margin-bottom: 6px;
        }
        .otp-box {
            width: 64px;
            height: 64px;
            border: 1px solid #d5d7da;
            border-radius: 10px;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 32px;
            font-weight: 500;
            color: #202124;
            background-color: #ffffff;
            box-shadow: 0px 1px 2px 0px rgba(10, 13, 18, 0.05);
            letter-spacing: -0.64px;
        }
        .expiry-text {
            font-size: 16px;
            line-height: 24px;
            color: #414651;
            margin: 24px 0;
        }
        .verify-button {
            display: inline-block;
            background-color: #2a2525;
            color: #ffffff;
            padding: 10px 16px;
            border-radius: 8px;
            text-decoration: none;
            font-size: 16px;
            font-weight: 600;
            line-height: 24px;
            margin: 24px 0;
            border: 2px solid rgba(255, 255, 255, 0.12);
            box-shadow: 0px 1px 2px 0px rgba(10, 13, 18, 0.05);
        }
        .verify-button:hover {
            background-color: #1a1a1a;
        }
        .closing {
            font-size: 16px;
            line-height: 24px;
            color: #414651;
            margin-top: 24px;
        }
        .email-footer {
            padding: 32px 24px;
            border-top: 1px solid #e5e7eb;
        }
        .footer-text {
            font-size: 14px;
            line-height: 20px;
            color: #535862;
            margin-bottom: 14px;
        }
        .footer-link {
            color: #6941c6;
            text-decoration: underline;
        }
        .footer-links {
            margin-bottom: 48px;
        }
        .footer-bottom {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-top: 48px;
        }
        .footer-logo {
            font-size: 16px;
            font-weight: 600;
            color: #202124;
        }
        .social-icons {
            display: flex;
            gap: 16px;
        }
        .social-icon {
            width: 20px;
            height: 20px;
            opacity: 0.7;
        }
        .copyright {
            font-size: 14px;
            line-height: 20px;
            color: #535862;
            margin-top: 14px;
        }
    </style>
</head>
<body>
    <div style="background-color: #fafafa; padding: 32px 0;">
        <div class="email-container">
            <!-- Header -->
            <div class="email-header">
                <a href="https://equitywala.com" class="logo">equitywala.com</a>
            </div>

            <!-- Body -->
            <div class="email-body">
                <div class="greeting">
                    <p>Hi {{.RecipientName}},</p>
                    <p>This is your verification code:</p>
                </div>

                <div class="otp-container">
                    <div class="otp-boxes">
                        {{range .OTPDigits}}
                        <div class="otp-box">{{.}}</div>
                        {{end}}
                    </div>
                </div>

                <div class="expiry-text">
                    <p>This code will only be valid for the next {{.ExpiresInMinutes}} minutes. If the code does not work, you can use this login verification link:</p>
                </div>

                <div>
                    <a href="{{.VerificationLink}}" class="verify-button">Verify email</a>
                </div>

                <div class="closing">
                    <p>Thanks,<br>The team</p>
                </div>
            </div>

            <!-- Footer -->
            <div class="email-footer">
                <div class="footer-text">
                    <p>This email was sent to <a href="mailto:{{.RecipientEmail}}" class="footer-link">{{.RecipientEmail}}</a>. If you'd rather not receive this kind of email, you can <a href="{{.UnsubscribeLink}}" class="footer-link">unsubscribe</a> or <a href="{{.ManagePreferencesLink}}" class="footer-link">manage your email preferences</a>.</p>
                </div>
                <div class="copyright">
                    <p>© {{.CompanyName}}</p>
                </div>
                <div class="footer-bottom">
                    <div class="footer-logo">equitywala.com</div>
                    <div class="social-icons">
                        <!-- Social media icons would go here -->
                        <span style="color: #a4a7ae;">Twitter</span>
                        <span style="color: #a4a7ae;">Facebook</span>
                        <span style="color: #a4a7ae;">Instagram</span>
                    </div>
                </div>
            </div>
        </div>
    </div>
</body>
</html>
`

