package core

import (
	"encoding/json"
	"fmt"
	"github.com/globalmac/boyar/types"
	"html/template"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type App struct {
	SiteConfig  SiteConfig
	ContentDir  string
	OutputDir   string
	TemplateDir string
	Posts       types.Posts
	Tags        types.Tags
	PostTypes   []string
}

type SiteConfig struct {
	BaseURL         string            `yaml:"baseURL"`
	Name            string            `yaml:"site_name"`
	Title           string            `yaml:"title"`
	Author          string            `yaml:"author"`
	SftpPort        string            `yaml:"sftp_port"`
	SftpServer      string            `yaml:"sftp_server"`
	SftpLogin       string            `yaml:"sftp_login"`
	SftpPassword    string            `yaml:"sftp_pass"`
	Keywords        string            `yaml:"keywords"`
	Description     string            `yaml:"description"`
	Port            string            `yaml:"port"`
	ContentPath     string            `yaml:"content_dir"`
	BuildDir        string            `yaml:"build_dir"`
	SourceDir       string            `yaml:"source_dir"`
	Pages           []string          `yaml:"pages"`
	PostTypesValues map[string]string `yaml:"post_types"`
	PerPageIndex    int               `yaml:"per_page_index"`
	PerPageCategory int               `yaml:"per_page_category"`
	PerPageTag      int               `yaml:"per_page_tag"`
}

func Process(cf string) *App {

	config, err := LoadConfig(cf)
	if err != nil {
		log.Fatal(err)
	}

	err = CreateDir(config.BuildDir)
	if err != nil {
		return nil
	}

	return &App{
		SiteConfig:  config,
		ContentDir:  config.ContentPath,
		OutputDir:   config.BuildDir,
		TemplateDir: config.SourceDir,
	}
}

func (core *App) ScanContent() {
	var paths []string

	filepath.Walk(core.ContentDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, ".md") {
			paths = append(paths, path)
		}

		return nil
	})

	for _, path := range paths {
		dir := filepath.Dir(path)
		folder := strings.Replace(dir, core.ContentDir, "", 1)
		filename := filepath.Base(path)
		filename = strings.Replace(filename, ".md", "", 1)

		var post types.Post

		data, err := os.ReadFile(path)
		if err != nil {
			log.Println(err)
		}

		fmd, body := markdownParser(string(data))

		post.Type = strings.Replace(folder, "/", "", 1)
		post.Slug = filename
		post.Slug = slugify(filename)
		post.Title = fmd.Title
		post.Cover = fmd.Cover
		post.Image = fmd.Image
		post.Date = fmd.Date
		post.Tags = fmd.Tags
		post.Description = fmd.Description
		post.Author = fmd.Author
		post.SourceUrl = fmd.SourceUrl

		if fmd.Draft {
			post.Status = "draft"
		} else {
			post.Status = "published"
		}

		// Collect all post types
		if !sliceContains(core.PostTypes, post.Type) {
			core.PostTypes = append(core.PostTypes, post.Type)
		}

		post.Content, err = markdownRender(body)
		if err != nil {
			log.Println(err)
		}

		post.Summary, post.Reminder = splitContent(post.Content)
		post.SummaryClean = removeHTMLTags(post.Summary)

		if post.Status == "published" {
			core.Posts = append(core.Posts, post)
		}
	}

	// Retrieve tags
	for _, post := range core.Posts {
		for _, t := range post.Tags {
			tag := core.Tags.Find(t)

			if tag.Name == "" {
				tag.Name = t
				tag.Slug = slugify(t)

				core.Tags = append(core.Tags, tag)
			}
		}
	}
}

func splitContent(content string) (summary, remainder string) {
	parts := strings.SplitN(content, "<!--more-->", 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	return content, ""
}

func (core *App) MakeIndexPage() {
	sortedPosts := core.Posts
	sort.Sort(types.PostsByDate(sortedPosts))

	if len(core.SiteConfig.Pages) > 0 {
		for _, page := range core.SiteConfig.Pages {
			if _, err := os.Stat(page); os.IsNotExist(err) {
				err := core.SaveAsHTML(page, page, map[string]interface{}{})
				if err != nil {
					log.Println(err)
				}
			}
		}
	}

	perPage := core.SiteConfig.PerPageIndex
	dividedPosts := DividePosts(sortedPosts, perPage, "all")

	if len(dividedPosts) > 0 {
		for pageNum, pagePosts := range dividedPosts {

			pp := pageNum + 1

			data := map[string]interface{}{
				"Posts":       pagePosts,
				"IsHome":      true,
				"CurrentPage": pp,
				"TotalPages":  len(dividedPosts),
			}

			fileName := fmt.Sprintf("/%d.html", pageNum+1)
			if pp == 1 {
				fileName = fmt.Sprintf("index.html")
			}
			err := core.SaveAsHTML(fileName, "index.html", data)
			if err != nil {
				log.Println(err)
			}
		}

	} else {

		data := map[string]interface{}{
			"Posts":       sortedPosts,
			"IsHome":      true,
			"CurrentPage": 0,
			"TotalPages":  len(dividedPosts),
		}

		fileName := fmt.Sprintf("index.html")
		err := core.SaveAsHTML(fileName, "index.html", data)
		if err != nil {
			log.Println(err)
		}

	}

}

func (core *App) MakeDetailPages() {

	sortedTags := core.Tags
	sort.Sort(types.TagsByName(sortedTags))

	if len(sortedTags) > 0 {
		for _, post := range core.Posts {
			fileName := fmt.Sprintf("%s/%s.html", post.Type, post.Slug)
			data := map[string]interface{}{
				"Post":       post,
				"IsSingular": true,
				"Tags":       sortedTags,
			}
			err := core.SaveAsHTML(fileName, "detail.html", data)
			if err != nil {
				log.Println(err)
			}
		}
	}

}

func (core *App) MakeTagIndexPage() {
	sortedTags := core.Tags
	sort.Sort(types.TagsByName(sortedTags))

	if len(sortedTags) > 0 {
		err := core.SaveAsHTML("tags/index.html", "tags.html", map[string]interface{}{
			"Tags": sortedTags,
		})
		if err != nil {
			log.Println(err)
		}
	}
}

func (core *App) MakeTagPages() {

	sortedTags := core.Tags
	sort.Sort(types.TagsByName(sortedTags))

	if len(sortedTags) > 0 {
		for i, tag := range sortedTags {
			sortedTags[i].CountPosts = len(core.Posts.FindByTag(tag.Name))
		}
		for _, tag := range core.Tags {
			perPage := core.SiteConfig.PerPageTag
			tag.Posts = core.Posts.FindByTag(tag.Name)
			dividedPosts := DividePosts(tag.Posts, perPage, "all")
			for pageNum, tagPosts := range dividedPosts {

				pp := pageNum + 1

				data := map[string]interface{}{
					"TagPosts":    tagPosts,
					"IsArchive":   true,
					"CurrentPage": pp,
					"Tag":         tag,
					"PostTypes":   core.PostTypes,
					"Tags":        sortedTags,
					"TotalPages":  len(dividedPosts),
				}

				fileName := fmt.Sprintf("tags/%d/%s.html", pageNum+1, tag.Slug)
				if pp == 1 {
					fileName = fmt.Sprintf("tags/%s.html", tag.Slug)
				}
				err := core.SaveAsHTML(fileName, "tag.html", data)
				if err != nil {
					log.Println(err)
				}
			}

		}
	}

}

func (core *App) MakeRSS() {
	sortedPosts := core.Posts
	sort.Sort(types.PostsByDate(sortedPosts))

	if len(sortedPosts) > 0 {

		data := map[string]interface{}{
			"Posts": sortedPosts,
			"Site":  core.SiteConfig,
		}

		rssTemplate := `
	{{ $baseURL := .Site.BaseURL }}
	<rss version="2.0">
	<channel>
		<title>{{ .Site.Title }}</title>
		<link>{{ $baseURL }}</link>
		<description>{{ .Site.Description }}</description>
		{{ range .Posts }}
		<item>
			<title>{{ .Title }}</title>
			<link>{{ $baseURL }}{{ .Permarlink }}</link>
			<description>{{ .SummaryClean }}</description>
			<pubDate>{{ .Date.Format "2006-01-02T15:04:05" }}</pubDate>
		</item>
		{{ end }}
	</channel>
	</rss>
	`

		t, err := template.New("").Parse(rssTemplate)
		if err != nil {
			log.Fatalln(err)
		}

		f, err := os.Create(core.OutputDir + "/rss.xml")
		if err != nil {
			log.Fatalln(err)
		}

		t.Execute(f, data)

	}
}

func (core *App) MakeSearchJson() {

	sortedPosts := core.Posts
	sort.Sort(types.PostsByDate(sortedPosts))

	if len(sortedPosts) > 0 {

		type SearchBlock struct {
			Url   string `json:"k"`
			Title string `json:"v"`
		}
		var data []SearchBlock
		if len(sortedPosts) > 0 {
			for _, post := range sortedPosts {
				data = append(data, SearchBlock{
					post.Title, core.SiteConfig.BaseURL + post.Permarlink(),
				})
			}
			if len(data) > 0 {
				f, _ := os.Create(core.OutputDir + "/search.json")
				defer f.Close()
				jsonContent, _ := json.Marshal(data)
				f.Write(jsonContent)
			}
		}

	}

}

func (core *App) MakePostCategories() {

	if len(core.PostTypes) > 0 {

		for _, postType := range core.PostTypes {

			var posts types.Posts

			for _, post := range core.Posts {
				if post.Type == postType {
					posts = append(posts, post)
				} else if strings.HasPrefix(post.Type, postType) {
					posts = append(posts, post)
				}
			}

			perPage := core.SiteConfig.PerPageCategory
			dividedPosts := DividePosts(posts, perPage, postType)

			sortedTags := core.Tags
			sort.Sort(types.TagsByName(sortedTags))

			if len(sortedTags) > 0 {
				for i, tag := range sortedTags {
					sortedTags[i].CountPosts = len(core.Posts.FindByTag(tag.Name))
				}
			}

			for pageNum, pagePosts := range dividedPosts {

				pp := pageNum + 1

				data := map[string]interface{}{
					"Posts":       pagePosts,
					"PostType":    postType,
					"PostTypes":   core.PostTypes,
					"Tags":        sortedTags,
					"PerPage":     perPage,
					"CurrentPage": pp,
					"TotalPages":  len(dividedPosts),
				}

				fileName := fmt.Sprintf("%s/page/%d.html", postType, pageNum+1)
				if pp == 1 {
					fileName = fmt.Sprintf("%s/index.html", postType)
				}

				err := core.SaveAsHTML(fileName, "posts.html", data)
				if err != nil {
					log.Println(err)
				}

			}

		}

	}

}

func (core *App) MakeSiteMap() {
	sortedPosts := core.Posts
	sort.Sort(types.PostsByDate(sortedPosts))

	if len(sortedPosts) > 0 {

		data := map[string]interface{}{
			"Posts": sortedPosts,
			"Site":  core.SiteConfig,
		}

		sitemapTemplate := strings.TrimSpace(`
	{{ $baseURL := .Site.BaseURL }}
	<urlset xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xsi:schemaLocation="http://www.sitemaps.org/schemas/sitemap/0.9 http://www.sitemaps.org/schemas/sitemap/0.9/sitemap.xsd">
	<channel>
		{{ range .Posts }}
		<url>
			<loc>{{ $baseURL }}{{ .Permarlink }}</loc>
			<lastmod>{{ .Date.Format "2006-01-02T15:04:05" }}</lastmod>
		</url>
		{{ end }}
	</channel>
	</urlset>
	`)

		t, err := template.New("").Parse(sitemapTemplate)
		if err != nil {
			log.Fatalln(err)
		}

		f, err := os.Create(core.OutputDir + "/sitemap.xml")
		if err != nil {
			log.Fatalln(err)
		}

		t.Execute(f, data)

	}

}

func (core *App) MakeRobotsTxt() {
	data := map[string]interface{}{
		"Url": core.SiteConfig.BaseURL,
	}

	sitemapTemplate := strings.TrimSpace(`User-agent: *
Disallow:

Sitemap: {{ .Url }}/sitemap.xml`)

	t, err := template.New("").Parse(sitemapTemplate)
	if err != nil {
		log.Fatalln(err)
	}

	f, err := os.Create(core.OutputDir + "/robots.txt")
	if err != nil {
		log.Fatalln(err)
	}

	t.Execute(f, data)

}

func (core *App) SaveAsHTML(fileName, templateName string, data map[string]interface{}) error {
	tpl := compileTemplate(templateName, core)

	fullPath := core.OutputDir + "/" + fileName

	err := CreateDir(filepath.Dir(fullPath))
	if err != nil {
		return err
	}

	f, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer f.Close()

	data["Site"] = map[string]interface{}{
		"BaseURL":     core.SiteConfig.BaseURL,
		"Title":       core.SiteConfig.Title,
		"Name":        core.SiteConfig.Name,
		"Description": core.SiteConfig.Description,
		"Keywords":    core.SiteConfig.Keywords,
		"NowYear":     time.Now().Format("2006"),
		"Timestamp":   time.Now().Unix(),
		"Posts":       core.Posts,
		"Tags":        core.Tags,
	}

	return tpl.ExecuteTemplate(f, templateName, data)
}

func (core *App) CopyStaticFiles() {
	var paths []string

	filepath.Walk(core.SiteConfig.SourceDir+"/static", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			paths = append(paths, path)
		}

		return nil
	})

	if len(paths) > 0 {

		for _, path := range paths {
			dir := filepath.Dir(path)
			destDir := strings.Replace(dir, core.SiteConfig.SourceDir+"/static", core.SiteConfig.BuildDir, 1)
			filename := filepath.Base(path)

			err := CreateDir(destDir)
			if err != nil {
				log.Println(err)
			}

			srcFile, err := os.Open(path)
			if err != nil {
				log.Println(err)
			}
			defer srcFile.Close()

			destFile, err := os.Create(destDir + "/" + filename)
			if err != nil {
				log.Println(err)
			}
			defer destFile.Close()

			_, err = io.Copy(destFile, srcFile)
			if err != nil {
				log.Println(err)
			}
		}

	}

}

func compileTemplate(templateName string, core *App) *template.Template {
	t := template.New("")

	funcMap := template.FuncMap{
		"safe_html": func(s string) template.HTML {
			return template.HTML(s)
		},
		"join": func(a []string, sep string) string {
			return strings.Join(a, sep)
		},
		"slugify":     slugify,
		"slugify_tag": slugifyWithExt,
		"inc": func(i int) int {
			return i + 1
		},
		"dec": func(i int) int {
			return i - 1
		},
		"seq": func(start, end int) []int {
			seq := make([]int, end-start+1)
			for i := range seq {
				seq[i] = start + i
			}
			return seq
		},
		"subtract": func(a, b int) int {
			return a - b
		},
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
		"max": func(a, b int) int {
			if a > b {
				return a
			}
			return b
		},
		"min": func(a, b int) int {
			if a < b {
				return a
			}
			return b
		},
		"mod": func(a, b int) int {
			return a % b
		},
		"len": func(q []types.Post) int {
			return len(q)
		},
		"post_types": func(a string) string {

			vv := core.SiteConfig.PostTypesValues

			if vv[a] != "" {
				return vv[a]
			} else {
				return "-"
			}

		},
	}

	t = template.Must(t.Funcs(funcMap).ParseGlob(core.SiteConfig.SourceDir + "/layouts/*.html"))

	return template.Must(t.ParseFiles(core.SiteConfig.SourceDir + "/" + templateName))
}

func DividePosts(posts types.Posts, perPage int, postType string) [][]types.Post {
	var dividedPosts [][]types.Post
	var allPosts []types.Post

	if len(posts) > 0 {

		for _, post := range posts {

			if strings.HasPrefix(post.Type, postType) {
				allPosts = append(allPosts, post)
			} else {
				if postType == "all" {
					allPosts = append(allPosts, post)
				} else if postType != "pages" && post.Type == postType {
					allPosts = append(allPosts, post)
				} else if post.Type == postType {
					allPosts = append(allPosts, post)
				}
			}

		}

		for i := 0; i < len(allPosts); i += perPage {
			end := i + perPage
			if end > len(allPosts) {
				end = len(allPosts)
			}
			dividedPosts = append(dividedPosts, allPosts[i:end])
		}

	}

	return dividedPosts
}
