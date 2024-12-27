package templates

const NoticeCommentBody string = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Notice comment notification</title>
    <style>
        body {
            font-family: 'Roboto', Arial, sans-serif;
            margin: 0;
            padding: 0;
        }
        .email-container {
            max-width: 650px;
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
            text-align: center;
            font-size: 16px;
            color: #333333;
        }
        .comment {
            color: #263238;
						margin-top: 25px;
						padding-left: 20px;
						text-align: left;
						line-height: 1.6em;
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
				.button {
					padding: 15px;
					background-color: #263228;
					color: #ffffff;
					font-weight: bold;
					border-radius: 10px;
					text-decoration: none;
				}
    </style>
</head>
<body>
    <div class="email-container">
        <div class="header">
            <h1>New comment notification</h1>
        </div>
        <div class="content">
            <h2>Hello {{Name}},</h2>
            <p><strong>{{Commenter}}</strong> has just commented on your post like below.</p>
						<div class="comment">{{Comment}}</div>
						<p>&nbsp;</p>
      			<p><a href="{{Link}}" class="button">VIEW COMMENT</a></p>
						<p>&nbsp;</p>
						<p>&nbsp;</p>
        </div>
        <div class="footer">
            <p>If you have any questions, contact us at <a href="mailto:{{From}}">{{From}}</a> ⎯ <a href="http://{{Host}}">{{Host}}</a></p>
        </div>
    </div>
</body>
</html>
`
