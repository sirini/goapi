package templates

const WelcomeBody = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Welcome Email</title>
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
            <h1>Welcome Aboard!</h1>
        </div>
        <div class="content">
            <h2>Congratulations on Joining, {{Name}}!</h2>
            <p>Thank you for signing up with us. You’re now all set to explore and make the most out of our platform's features and services.</p>
            <p>To get started, click the button below to log in and begin your journey.</p>
            <a href="{{Host}}/login" class="button" target="_blank">Go to Login</a>
        </div>
        <div class="footer">
            <p>If you have any questions, contact us at <a href="mailto:{{From}}">{{From}}</a> ⎯ <a href="http://{{Host}}">{{Host}}</a></p>
        </div>
    </div>
</body>
</html>
`
