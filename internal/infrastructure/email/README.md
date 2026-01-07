# Email Service

This package provides email functionality for the Equitywala backend, specifically for sending OTP verification emails.

## Configuration

The email service requires the following environment variables:

```env
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM_EMAIL=no-reply@equitywala.com
SMTP_FROM_NAME=Team Equitywala
FRONTEND_URL=http://localhost:5173
```

### Gmail Setup

If using Gmail, you'll need to:
1. Enable 2-Step Verification on your Google account
2. Generate an App Password: https://myaccount.google.com/apppasswords
3. Use the App Password as `SMTP_PASSWORD`

### Other SMTP Providers

The service works with any SMTP provider. Common configurations:

**SendGrid:**
- SMTP_HOST: `smtp.sendgrid.net`
- SMTP_PORT: `587`
- SMTP_USERNAME: `apikey`
- SMTP_PASSWORD: Your SendGrid API key

**AWS SES (Amazon Simple Email Service):**

For **Asia Pacific (Mumbai) - ap-south-1** region:
```env
SMTP_HOST=email-smtp.ap-south-1.amazonaws.com
SMTP_PORT=587
SMTP_USERNAME=your-ses-smtp-username
SMTP_PASSWORD=your-ses-smtp-password
SMTP_FROM_EMAIL=verified-email@yourdomain.com
SMTP_FROM_NAME=Team Equitywala
FRONTEND_URL=http://localhost:5173
```

**Setting up AWS SES SMTP credentials:**

1. **Go to AWS SES Console:**
   - Navigate to: https://console.aws.amazon.com/ses/
   - Select your region (e.g., Asia Pacific - Mumbai)

2. **Create SMTP credentials:**
   - Click on "SMTP settings" in the left sidebar
   - Click "Create SMTP credentials" button
   - Enter a name for your IAM user (e.g., "equitywala-smtp-user")
   - Click "Create"
   - **Important:** Download and save the credentials file - you won't be able to see the password again!
   - The credentials file will contain:
     - **SMTP Username** (looks like: `AKIAIOSFODNN7EXAMPLE`)
     - **SMTP Password** (a long random string)
   - **⚠️ These are NOT your regular AWS Access Keys!** They are SMTP-specific credentials.

3. **Verify your email address or domain:**
   - Go to "Verified identities" in SES console
   - Click "Create identity"
   - Choose "Email address" or "Domain"
   - Follow the verification process
   - Use the verified email as `SMTP_FROM_EMAIL`

4. **Port options:**
   - **Port 587 (STARTTLS)** - Recommended (use this)
   - Port 465 (TLS Wrapper) - Alternative
   - Port 25 - Not recommended (often blocked by ISPs)

5. **Important notes:**
   - AWS SES SMTP credentials are **different** from your AWS access keys
   - SMTP credentials are region-specific
   - In sandbox mode, you can only send to verified email addresses
   - Request production access to send to any email address

**For other AWS regions:**
- US East (N. Virginia): `email-smtp.us-east-1.amazonaws.com`
- US West (Oregon): `email-smtp.us-west-2.amazonaws.com`
- EU (Ireland): `email-smtp.eu-west-1.amazonaws.com`
- Replace `{region}` in the hostname with your region code

## Email Template

The OTP email template is based on the Figma design and includes:
- Equitywala branding
- Personalized greeting
- 4-digit OTP code displayed in separate boxes
- Expiration message
- Verification button
- Footer with unsubscribe links and social icons

The template is embedded in the code but can be overridden by placing a file at `templates/email/otp.html`.

## Usage

The email service is automatically integrated with the authentication handlers:
- Signup: Sends OTP email after user registration
- Resend OTP: Sends a new OTP email when requested

The service gracefully handles failures - if email sending fails, the error is logged but doesn't block the user registration process (OTP is still created in the database).

