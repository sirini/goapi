package templates

const SitemapBody = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
{{- range . }}
  <url>
    <loc>{{ .Loc }}</loc>
    <lastmod>{{ .LastMod }}</lastmod>
    <changefreq>{{ .ChangeFreq }}</changefreq>
    <priority>{{ .Priority }}</priority>
  </url>
{{- end }}
</urlset>`
