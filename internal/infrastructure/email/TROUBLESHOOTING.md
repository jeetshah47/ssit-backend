# Email Service Troubleshooting

## Error: "535 Authentication Credentials Invalid"

This error means your SMTP credentials are incorrect or not properly configured.

### Quick Fix Steps

1. **Check your `.env` file** in the `ssit-backend` directory and verify all SMTP variables are set:
   ```env
   SMTP_HOST=email-smtp.ap-south-1.amazonaws.com  # For AWS SES (or smtp.gmail.com for Gmail)
   SMTP_PORT=587
   SMTP_USERNAME=your-smtp-username
   SMTP_PASSWORD=your-smtp-password
   SMTP_FROM_EMAIL=your-verified-email@yourdomain.com
   SMTP_FROM_NAME=Team Equitywala
   FRONTEND_URL=http://localhost:5173
   ```

2. **For AWS SES specifically:**
   - ❌ **DON'T use your AWS Access Key ID and Secret Access Key**
   - ✅ **DO use SMTP credentials created through SES Console**
   
   **How to get AWS SES SMTP credentials:**
   1. Go to AWS SES Console: https://console.aws.amazon.com/ses/
   2. Select your region (e.g., Asia Pacific - Mumbai)
   3. Click "SMTP settings" in the left sidebar
   4. Click "Create SMTP credentials" or "Manage my existing SMTP credentials"
   5. Download the credentials file immediately (you won't see the password again!)
   6. Use the **SMTP Username** (looks like `AKIAIOSFODNN7EXAMPLE`) as `SMTP_USERNAME`
   7. Use the **SMTP Password** (long random string) as `SMTP_PASSWORD`
   8. Make sure your `SMTP_FROM_EMAIL` is verified in SES Console → Verified identities
   
   **See detailed guide:** `internal/infrastructure/email/AWS_SES_SETUP.md`

3. **For Gmail specifically:**
   - ❌ **DON'T use your regular Gmail password**
   - ✅ **DO use an App Password**
   
   **How to get a Gmail App Password:**
   1. Go to https://myaccount.google.com/apppasswords
   2. Sign in with your Google account
   3. Select "Mail" and "Other (Custom name)"
   4. Enter "Equitywala Backend" as the name
   5. Click "Generate"
   6. Copy the 16-character password (no spaces)
   7. Use this as your `SMTP_PASSWORD` in `.env`

4. **Verify your settings:**
   - **For AWS SES:**
     - `SMTP_HOST`: Should be `email-smtp.{region}.amazonaws.com` (e.g., `email-smtp.ap-south-1.amazonaws.com`)
     - `SMTP_PORT`: Should be `587` (STARTTLS) - recommended
     - `SMTP_USERNAME`: Your SES SMTP username (from credentials file)
     - `SMTP_PASSWORD`: Your SES SMTP password (from credentials file)
     - `SMTP_FROM_EMAIL`: Must be a verified email address in SES
   - **For Gmail:**
     - `SMTP_HOST`: Should be `smtp.gmail.com`
     - `SMTP_PORT`: Should be `587` (TLS) or `465` (SSL)
     - `SMTP_USERNAME`: Your full Gmail address
     - `SMTP_PASSWORD`: The 16-character App Password (not your regular password)

### Common Issues

**Issue: "535 Authentication Credentials Invalid"**
- **For AWS SES**: Using AWS access keys instead of SMTP credentials, or wrong SMTP credentials
  - **Fix**: Create proper SMTP credentials through SES Console → SMTP settings
- **For Gmail**: Using regular password instead of App Password
  - **Fix**: Generate a new App Password and update `.env`

**Issue: "Connection timeout"**
- **Cause**: Firewall or network blocking SMTP port
- **Fix**: Check if port 587 is open, or try port 465 with SSL

**Issue: "TLS handshake failed"**
- **Cause**: SMTP server doesn't support the TLS version
- **Fix**: The service uses `MandatoryStartTLS`, ensure your SMTP server supports it

### Testing Email Configuration

You can test if your email configuration is correct by:

1. **Check logs**: The service logs detailed error messages
2. **Verify environment variables are loaded**: Check that your `.env` file is in the `ssit-backend` directory
3. **Test with a simple SMTP client**: Use a tool like `telnet` or an email testing tool

### Alternative: Use a Different SMTP Provider

If Gmail is causing issues, consider:

**SendGrid (Recommended for production):**
```env
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USERNAME=apikey
SMTP_PASSWORD=your-sendgrid-api-key
```

**AWS SES:**
```env
SMTP_HOST=email-smtp.ap-south-1.amazonaws.com
SMTP_PORT=587
SMTP_USERNAME=your-ses-smtp-username
SMTP_PASSWORD=your-ses-smtp-password
```

**⚠️ Important:** AWS SES SMTP credentials are **DIFFERENT** from your AWS access keys!
- **SMTP Username**: Created through SES Console → SMTP Settings → Create SMTP credentials
- **SMTP Password**: Generated when you create SMTP credentials (download the file!)
- **NOT** your AWS Access Key ID and Secret Access Key from IAM

**How to get SMTP credentials:**
1. Go to AWS SES Console → SMTP settings
2. Click "Create SMTP credentials" or "Manage my existing SMTP credentials"
3. Download the credentials file (contains SMTP username and password)
4. Use those values in your `.env` file

### Important Notes

- The email service runs **asynchronously** - errors won't block the signup process
- OTP codes are still created and stored even if email fails
- Check application logs for detailed error messages
- In development, you can continue testing even if emails fail (OTP is in the database)

### Next Steps

After fixing your SMTP credentials:
1. Restart your backend server
2. Try signing up again
3. Check logs to confirm email is sent successfully
4. Verify the email arrives in the recipient's inbox

