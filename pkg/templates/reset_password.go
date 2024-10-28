package templates

var ResetPasswordTitle string = "[{{Host}}] Reset Your Password"
var ResetPasswordBody string = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Password Reset</title>
    <style>
        body {
            font-family: 'Roboto', Arial, sans-serif;
            margin: 0;
            padding: 0;
        }
        .container {
            max-width: 450px;
            margin: 0 auto;
            background-color: #ECEFF1;
            padding: 25px;
            border-radius: 20px;
        }
        .header {
            text-align: center;
            padding-bottom: 20px;
        }
        .header h1 {
            color: #263238;
        }
        .content {
            padding: 20px;
            color: #333333;
        }
        .content p {
            line-height: 1.6;
            margin: 10px 0;
        }
        .button {
            display: block;
            width: fit-content;
            margin: 20px auto;
            padding: 12px 24px;
            background-color: #263238;
            color: #ffffff;
            text-decoration: none;
            border-radius: 10px;
            font-size: 16px;
            font-weight: 500;
            text-align: center;
        }
        .footer {
            text-align: center;
            font-size: 12px;
            color: #888888;
            margin-top: 20px;
        }
        .footer a {
            color: #546E7A;
            text-decoration: none;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Password Reset Request</h1>
        </div>
        <div class="content">
            <h2>Hello!</h2>
            <p>We received a request to reset your password. Click the button below to set up a new password for your account.</p>
            <a href="{{Host}}/changepassword/{{Uid}}/{{Code}}" class="button" target="_blank">Reset Password</a>
            <p>If you didn't request a password reset, please ignore this email or contact support if you have any concerns.</p>
            <p>For security reasons, this link will expire in 24 hours.</p>
        </div>
        <div class="footer">
            <p>If you have any questions, contact us at <a href="mailto:{{From}}">{{From}}</a> ⎯ <a href="http://{{Host}}">{{Host}}</a></p>
        </div>
    </div>
</body>
</html>
`

var ResetPasswordChat string = "Request to reset password from {{Id}} ({{Uid}})"
