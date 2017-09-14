package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ericaro/frontmatter"
	"github.com/julienschmidt/httprouter"
	"github.com/microcosm-cc/bluemonday"
	"github.com/oxtoacart/bpool"
	"github.com/russross/blackfriday"
)

var templates map[string]*template.Template
var article map[string]*Post
var allposts Posts
var bufpool *bpool.BufferPool

// AboutMe : About Me structure
type AboutMe struct {
	Name        string
	City        string
	Nationality string
}

// Post : Post structure
type Post struct {
	Title       string        `yaml:"Title"`
	Author      string        `yaml:"Author"`
	Description string        `yaml:"Description"`
	Date        string        `yaml:"Date"`
	Tag         []string      `yaml:"Tag,flow"`
	Content     template.HTML `fm:"content" yaml:"-"`
	Link        string
	date        time.Time
}

// Posts : posts sorted by date
type Posts []Post

func (slice Posts) Len() int {
	return len(slice)
}

func (slice Posts) Less(i, j int) bool {
	return slice[i].date.After(slice[j].date)
}

func (slice Posts) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

// PathConfig : Template configurations
type PathConfig struct {
	TemplateLayoutPath  string
	TemplateIncludePath string
	ArticlePath         string
}

var mainTmpl = `{{ define "main" }} {{ template "base" . }} {{end}}`
var pathConfig PathConfig

func loadConfiguration() {
	pathConfig.ArticlePath = "posts/"
	pathConfig.TemplateLayoutPath = "templates/layout/"
	pathConfig.TemplateIncludePath = "templates/"
}

func loadArticles() {
	if article == nil {
		article = make(map[string]*Post)
	}
	if allposts != nil {
		allposts = Posts{}
	}
	articlePath, err := filepath.Glob(pathConfig.ArticlePath + "*.md")
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range articlePath {
		baseName := filepath.Base(file)
		n := strings.LastIndexByte(baseName, '.')
		var fileName string
		if n > 0 {
			fileName = baseName[:n]
		}
		postStruct := &Post{}
		postFile, err := ioutil.ReadFile(file)
		if err != nil {
			panic(err)
		}
		err = frontmatter.Unmarshal(postFile, postStruct)
		if err != nil {
			log.Fatal(err)
		}
		postStruct.Link = fileName
		dateParse, _ := time.Parse("Monday, 2 January 2006", postStruct.Date)
		postStruct.date = dateParse
		postStruct.Date = dateParse.Format("2 Jan 2006")
		article[fileName] = postStruct
	}
	// Sorting by date
	for key, value := range article {
		allposts = append(allposts, Post{
			Title: value.Title,
			Link:  key,
			date:  article[key].date, // Important for ordering post by date
			Date:  article[key].Date,
		})
	}
	sort.Sort(allposts)

	log.Println("article load successfully")
	log.Println(article)

}

func loadTemplates() {
	if templates == nil {
		templates = make(map[string]*template.Template)
	}

	layoutFiles, err := filepath.Glob(pathConfig.TemplateLayoutPath + "*.tmpl")
	if err != nil {
		log.Fatal(err)
	}

	includeFiles, err := filepath.Glob(pathConfig.TemplateIncludePath + "*.tmpl")
	if err != nil {
		log.Fatal(err)
	}

	mainTemplate := template.New("main")

	mainTemplate, err = mainTemplate.Parse(mainTmpl)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range includeFiles {
		fileName := filepath.Base(file)
		files := append(layoutFiles, file)
		templates[fileName], err = mainTemplate.Clone()
		if err != nil {
			log.Fatal(err)
		}
		templates[fileName] = template.Must(templates[fileName].ParseFiles(files...))
	}

	log.Println("templates loading successful")
	bufpool = bpool.NewBufferPool(64)
	log.Println("buffer allocation successful")
}

func renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	tmpl, ok := templates[name]
	if !ok {
		http.Error(w, fmt.Sprintf("The template %s does not exist.", name), http.StatusInternalServerError)
	}

	buf := bufpool.Get()
	defer bufpool.Put(buf)

	err := tmpl.Execute(buf, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
}

func index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	renderTemplate(w, "index.tmpl", nil)
}

func articlesIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// allpost := Posts{}
	// for key, value := range article {
	// 	allposts = append(allposts, Post{
	// 		Title: value.Title,
	// 		Link:  key,
	// 		date:  article[key].date, // Important for ordering post by date
	// 		Date:  article[key].Date,
	// 	})
	// }
	// sort.Sort(allposts)
	// posts := make([]Post, 0, len(article))
	// for key := range article {
	// 	posts = append(posts, Post{
	// 		Title: article[key].Title,
	// 		Link:  key,
	// 	})
	// }
	data := struct {
		Post []Post
	}{
		allposts,
	}
	renderTemplate(w, "indexarticle.tmpl", data)
}

func articles(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	postStructure := article[p.ByName("article")]
	unsafe := blackfriday.MarkdownCommon([]byte(postStructure.Content))
	safeContent := bluemonday.UGCPolicy().SanitizeBytes(unsafe)
	postStructure.Content = template.HTML(safeContent)
	renderTemplate(w, "post.tmpl", postStructure)
}

func aboutMe(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	aboutme := &AboutMe{Name: "Ari Maulana", City: "Jakarta", Nationality: "Indonesia"}
	renderTemplate(w, "aboutme.tmpl", aboutme)
}

func main() {
	loadConfiguration()
	loadArticles()
	loadTemplates()

	router := httprouter.New()
	router.ServeFiles("/static/*filepath", http.Dir("public/"))
	router.GET("/", index)
	router.GET("/articles", articlesIndex)
	router.GET("/articles/:article", articles)
	router.GET("/aboutme", aboutMe)
	http.ListenAndServe(":8080", router)
}
