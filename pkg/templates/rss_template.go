package templates

const RssBody = `
<?xml version="1.0" encoding="UTF-8" ?>
<rss version="2.0">
    <channel>
        <title>#BLOG.TITLE#</title>
        <link>#BLOG.LINK#</link>
        <description>#BLOG.INFO#</description>
        <language>#BLOG.LANG#</language>
        <pubDate>#BLOG.DATE#</pubDate>
        <lastBuildDate>#BLOG.DATE#</lastBuildDate>
        <docs>http://blogs.law.harvard.edu/tech/rss</docs>
        <generator>#BLOG.GENERATOR#</generator>
        #BLOG.ITEM#
    </channel>
</rss>
`