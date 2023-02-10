package main

import (
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
)

var templates = template.Must(template.ParseFiles("edit.html", "view.html"))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

type Page struct {
	Title string
	Body  []byte
}

func (p *Page) save() error {
	filename := p.Title + ".txt"
	return os.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func main() {
	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/edit/", editHandler)
	http.HandleFunc("/save/", saveHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func viewHandler(responseWriter http.ResponseWriter, request *http.Request) {
	title, err := getTitle(responseWriter, request)
	if err != nil {
		return
	}
	page, err := loadPage(title)
	if err != nil {
		http.Redirect(responseWriter, request, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(responseWriter, "view", page)
}

func editHandler(responseWriter http.ResponseWriter, request *http.Request) {
	title, err := getTitle(responseWriter, request)
	if err != nil {
		return
	}
	page, err := loadPage(title)
	if err != nil {
		page = &Page{Title: title}
	}
	renderTemplate(responseWriter, "edit", page)
}

func saveHandler(responseWriter http.ResponseWriter, request *http.Request) {
	title, err := getTitle(responseWriter, request)
	if err != nil {
		return
	}
	body := request.FormValue("body")
	page := &Page{Title: title, Body: []byte(body)}
	err = page.save()
	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(responseWriter, request, "/view/"+title, http.StatusFound)
}

func renderTemplate(responseWriter http.ResponseWriter, viewFileName string, page *Page) {
	err := templates.ExecuteTemplate(responseWriter, viewFileName+".html", page)
	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}
}

func getTitle(responseWriter http.ResponseWriter, request *http.Request) (string, error) {
	m := validPath.FindStringSubmatch(request.URL.Path)
	if m == nil {
		http.NotFound(responseWriter, request)
		return "", errors.New("invalid Page Title")
	}
	return m[2], nil
}
