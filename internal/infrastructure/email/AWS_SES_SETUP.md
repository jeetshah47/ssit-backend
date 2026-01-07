# AWS SES SMTP Setup Guide

## ⚠️ Important: SMTP Credentials vs Access Keys

**AWS SES SMTP requires special SMTP credentials that are DIFFERENT from your AWS access keys.**

- ❌ **DON'T use**: AWS Access Key ID and Secret Access Key from IAM
- ✅ **DO use**: SMTP credentials created specifically for SES SMTP

## Step-by-Step Setup

### Step 1: Get SMTP Credentials

1. **Go to AWS SES Console:**
   - Navigate to: https://console.aws.amazon.com/ses/
   - Make sure you're in the correct region (e.g., **Asia Pacific - Mumbai**)

2. **Create SMTP Credentials:**
   - Click on **"SMTP settings"** in the left sidebar
   - Click **"Create SMTP credentials"** button (orange button)
   - Enter a name for your IAM user (e.g., "equitywala-smtp-user")
   - Click **"Create"**
   - **Download the credentials file immediately!** You won't be able to see the password again.

3. **What's in the credentials file:**
   ```
   SMTP Username: AKIAIOSFODNN7EXAMPLE
   SMTP Password: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
   ```
   - The **SMTP Username** goes in `SMTP_USERNAME`
   - The **SMTP Password** goes in `SMTP_PASSWORD`

### Step 2: Verify Your Email Address

1. **Go to "Verified identities"** in SES console
2. Click **"Create identity"**
3. Choose **"Email address"** or **"Domain"**
4. Enter your email address
5. Check your email and click the verification link
6. Use this verified email as `SMTP_FROM_EMAIL` in your `.env`

### Step 3: Configure Your `.env` File

```env
# AWS SES SMTP Configuration (Mumbai Region)
SMTP_HOST=email-smtp.ap-south-1.amazonaws.com
SMTP_PORT=587
SMTP_USERNAME=AKIAIOSFODNN7EXAMPLE  # From credentials file
SMTP_PASSWORD=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY  # From credentials file
SMTP_FROM_EMAIL=your-verified-email@yourdomain.com
SMTP_FROM_NAME=Team Equitywala
FRONTEND_URL=http://localhost:5173
```

### Step 4: Test the Configuration

1. Restart your backend server
2. Try signing up with a test email
3. Check the logs for any errors
4. If successful, you should receive the OTP email

## Common Questions

### Q: Can I use my AWS Access Key ID and Secret Access Key?
**A:** No! AWS SES SMTP interface requires SMTP-specific credentials. Regular AWS access keys won't work for SMTP authentication.

### Q: I lost my SMTP password, what do I do?
**A:** You need to create new SMTP credentials. Go to SMTP settings → Manage existing SMTP credentials → Create new credentials.

### Q: Can I use the same SMTP credentials in different regions?
**A:** No, SMTP credentials are region-specific. You need to create separate credentials for each region.

### Q: What if I'm in sandbox mode?
**A:** In sandbox mode, you can only send emails to verified email addresses. Request production access to send to any email address.

## Troubleshooting

### Error: "535 Authentication Credentials Invalid"
- **Cause**: Using wrong credentials (access keys instead of SMTP credentials)
- **Fix**: Create proper SMTP credentials through SES Console

### Error: "Email address not verified"
- **Cause**: Trying to send from an unverified email address
- **Fix**: Verify your email address in SES Console → Verified identities

### Error: "Message rejected: Email address is not verified"
- **Cause**: In sandbox mode, recipient email is not verified
- **Fix**: Verify the recipient email or request production access

## Port Configuration

- **Port 587 (STARTTLS)** - ✅ Recommended (already configured)
- **Port 465 (TLS Wrapper)** - Alternative if 587 doesn't work
- **Port 25** - Not recommended (often blocked by ISPs)

## Security Best Practices

1. **Never commit credentials to git** - Use `.env` file (already in `.gitignore`)
2. **Rotate credentials regularly** - Create new SMTP credentials periodically
3. **Use IAM policies** - Limit SMTP credentials to only send emails (not manage SES)
4. **Monitor usage** - Check CloudWatch for email sending metrics

