package templates

const MainPageBody = `
<!--
  ___________ ____  ____  ___    ____  ____  
 /_  __/ ___// __ )/ __ \/   |  / __ \/ __ \ 
  / /  \__ \/ __  / / / / /| | / /_/ / / / / 
 / /  ___/ / /_/ / /_/ / ___ |/ _, _/ /_/ /  
/_/  /____/_____/\____/_/  |_/_/ |_/_____/  

{{.Version}} | Powered by tsboard.dev
This page was created to assist search engine                  
-->
<!doctype html>
<html lang="ko">
  <head>
    <link rel="preconnect" href="https://fonts.gstatic.com/" crossorigin="anonymous" />
    <link
      rel="preload"
      as="style"
      onload="this.rel='stylesheet'"
      href="https://fonts.googleapis.com/css2?family=Roboto:wght@100;300;400;500;700;900&display=swap"
    />
    <meta charset="UTF-8" />
    <link rel="icon" href="/favicon.ico" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <link rel="preconnect" href="https://fonts.googleapis.com" />
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin />
    <link
      href="https://fonts.googleapis.com/css2?family=IBM+Plex+Sans+KR:wght@100;200;300;400;500;600;700&family=Protest+Strike&display=swap"
      rel="stylesheet"
    />
    <style type="text/css">
		body {
			font-family: "IBM Plex Sans KR", sans-serif;
			font-weight: 400;
			margin: 0px;
			font-size: 1em;
			background-color: #eceff1;
		}

		a {
			text-decoration: none;
		}

		#tsboardToolbar {
			position: fixed;
			z-index: 100;
			top: 0px;
			width: 100%;
			height: 64px;
			padding-left: 30px;
			background-color: #37474f;
		}

		#tsboardToolbar .title {
			cursor: pointer;
			font-family: "Protest Strike", sans-serif;
			font-size: 1.6em;
			color: white;
		}

		#tsboardToolbar .title a {
			color: white;
		}

		#tsboardContainer {
			padding-top: 64px;
			width: 1200px;
			min-height: 300px;
			margin: auto;
		}

		#tsboardInfo {
			padding-top: 10px;
			font-size: 0.9em;
			color: #b0bec5;
			text-align: center;
		}

		#tsboardContainer .article {
			position: relative;
			margin-top: 25px;
			box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
			border-radius: 10px;
			transition: all 0.3s ease;
			background-color: white;
			overflow: hidden;
		}

		#tsboardContainer .article:hover {
			box-shadow: 0 4px 6px rgba(0, 0, 0, 0.5);
		}

		#tsboardContainer .tag {
			border-top: #dddddd 1px solid;
			padding: 15px;
		}

		#tsboardContainer ul {
			margin: 0px;
			padding: 0px;
			margin-top: 10px;
			margin-bottom: 10px;
			list-style-type: none;
		}

		#tsboardContainer .tag li {
			display: inline;
			margin-top: 5px;
			margin-right: 5px;
			background-color: #cfd8dc;
			border-radius: 5px;
			padding: 10px;
		}

		#tsboardContainer .comment {
			border-top: #eeeeee 1px solid;
			padding: 15px;
			max-height: 300px;
			overflow: hidden;
		}

		#tsboardContainer .pa {
			padding: 15px;
		}

		#tsboardContainer .image {
			position: absolute;
			top: 0px;
			left: 0px;
			width: 300px;
			height: 600px;
			overflow: hidden;
			background-color: #f7f7f7;
			border-right: #eeeeee 1px solid;
		}

		#tsboardContainer .post {
			margin-left: 320px;
			height: 600px;
			overflow: hidden;
		}

		#tsboardContainer .post img {
			max-width: 100%;
			height: auto;
		}

		#tsboardContainer .post .title {
			margin: 0px;
			padding-top: 10px;
			padding-bottom: 10px;
		}

		#tsboardContainer .post .title a {
			color: #263238;
		}

		#tsboardContainer .post .content {
			line-height: 1.7em;
			padding-bottom: 10px;
			padding-right: 10px;
		}

		#tsboardContainer .additional {
			font-size: 0.9em;
			color: #b0bec5;
			padding-bottom: 10px;
		}

		#tsboardContainer .comment .commentContent {
			line-height: 1.7em;
			padding-bottom: 10px;
		}

		#tsboardFooter {
			text-align: center;
			padding-top: 15px;
			padding-bottom: 15px;
			margin-top: 30px;
			color: #b0bec5;
			font-size: 0.9em;
			line-height: 1.7em;
		}

		#tsboardFooter a {
			color: #455a64;
		}

		#tsboardFooter .poweredBy {
			padding-top: 10px;
			color: #b0bec5;
			font-size: 0.9em;
		}

		#tsboardFooter .poweredBy a {
			color: #b0bec5;
		}
		</style>
    <title>{{.PageTitle}}</title>
  </head>
  <body>
    <div id="tsboardApp">
      <header id="tsboardToolbar">
        <h1 class="title">
          <a
            href="{{.PageUrl}}"
            target="_blank"
            >{{.PageTitle}}</a
          >
        </h1>
      </header>

      <div id="tsboardContainer">

				<div id="tsboardInfo">
        <p>
					This page has been created to help search engines easily index the {{.PageTitle}} site.<br />
					For the page intended for actual users, please <a href="{{.PageUrl}}" target="_blank">click here</a> to visit.
        </p>
				</div>

				{{- range .Articles }}
				<article class="article">
					<div class="image">
						<img src="{{.Cover}}" width="300" />
					</div>
					<div class="post">
						<h2 class="title">
							<a href="{{.Url}}" target="_blank">{{.Title}}</a>
						</h2>
						<div class="content">
							<div class="additional">
								<span class="date">{{.Date}}</span> / 
								<span class="like">{{.Like}} like(s)</span> /
								<span class="writer">written by <strong>{{.Name}}</strong></span>
							</div>
							<div class="text">{{.Content}}</div>
						</div>
					</div>
					<section class="tag">
						<ul>
							{{- range .Hashtags }}
							<li>{{.Name}}</li>
							{{- end }}
						</ul>
					</section>
					<section class="comment">
						<ul>
							{{- range .Comments }}
							<li>
								<div class="commentContent">{{.Content}}</div>
								<div class="additional">
									<span class="date">{{.Date}}</span> /
									<span class="like">{{.Like}} like(s)</span> /
									<span class="writer">written by <strong>{{.Name}}</strong></span>
								</div>
							</li>
							{{- end }}
						</ul>
					</section>
				</article>
				{{- end }}
			</div>

      <footer id="tsboardFooter">
        <p>
					This page has been created to help search engines easily index the {{.PageTitle}} site.<br />
					For the page intended for actual users, please <a href="{{.PageUrl}}" target="_blank">click here</a> to visit.
        </p>
        <p class="poweredBy">
          <a
            href="https://tsboard.dev"
            target="_blank"
            title="This website was built using TSBOARD"
            >tsboard.dev</a
          >
        </p>
      </footer>
    </div>
  </body>
</html>
`
