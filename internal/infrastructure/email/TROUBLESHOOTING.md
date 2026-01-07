# Email Service Troubleshooting

## Error: "535 Authentication Credentials Invalid"

This error means your SMTP credentials are incorrect or not properly configured.

### Quick Fix Steps

1. **Check your `.env` file** in the `ssit-backend` directory:
   ```env
   SMTP_HOST=smtp.gmail.com
   SMTP_PORT=587
   SMTP_USERNAME=your-email@gmail.com
   SMTP_PASSWORD=your-app-password-here
   SMTP_FROM_EMAIL=no-reply@equitywala.com
   SMTP_FROM_NAME=Team Equitywala
   ```

2. **For Gmail specifically:**
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

3. **Verify your settings:**
   - `SMTP_HOST`: Should be `smtp.gmail.com` for Gmail
   - `SMTP_PORT`: Should be `587` (TLS) or `465` (SSL)
   - `SMTP_USERNAME`: Your full Gmail address
   - `SMTP_PASSWORD`: The 16-character App Password (not your regular password)

### Common Issues

**Issue: "535 Authentication Credentials Invalid"**
- **Cause**: Wrong password or using regular password instead of App Password
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

