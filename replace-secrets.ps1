# Script to replace secrets with placeholders in task-definition.json
$file = "aws-ecs/task-definition.json"
if (Test-Path $file) {
    $content = Get-Content $file -Raw
    
    # Replace all secrets with placeholders
    $content = $content -replace '"gAkErm5VehfXQQ9q9Xyy2PyFJ3sAzn3v"', '"REPLACE_WITH_DB_PASSWORD"'
    $content = $content -replace '"AKIAR4B5EDV66U7RZ7WL"', '"REPLACE_WITH_AWS_ACCESS_KEY_ID"'
    $content = $content -replace '"xo5apP\+ccqeI5VAW2FvFOIJYd3m28Paps2bFb2Xr"', '"REPLACE_WITH_AWS_SECRET_ACCESS_KEY"'
    $content = $content -replace '"AKIAR4B5EDV6ZVVQ7CXV"', '"REPLACE_WITH_SMTP_USERNAME"'
    $content = $content -replace '"BMn7Ua6UXajIN6A\+Wnb4/DJSwYghUuGLDjdG9fxF5KKl"', '"REPLACE_WITH_SMTP_PASSWORD"'
    $content = $content -replace '"HWN%I0rMvNLd#VnA"', '"REPLACE_WITH_PAYTM_MERCHANT_KEY"'
    $content = $content -replace '"vhdlgu1jn1mgr5e6"', '"REPLACE_WITH_ZERODHA_API_KEY"'
    $content = $content -replace '"qr9kjdttkg3hvggrfevwqcxgox3irdrc"', '"REPLACE_WITH_ZERODHA_API_SECRET"'
    $content = $content -replace '"36019972afbfac75d648995098bbbd712ad26036519a43bf8921f509fbaebbb42971e9d10ffbd4e611fc4bae0616d1cdfa57c105781a6f78e079336a1d29ae77"', '"REPLACE_WITH_JWT_SECRET"'
    $content = $content -replace 'mongodb\+srv://go-backend:H3V2LS32ATeb@cluster0\.chqvn9u\.mongodb\.net/\?appName=Cluster0', 'REPLACE_WITH_MONGODB_URI'
    
    Set-Content -Path $file -Value $content -NoNewline
    git add $file
}
