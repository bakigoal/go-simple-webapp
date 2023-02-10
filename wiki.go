package main

import (
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
)

var templates = template.Must(template.ParseFiles("templates/edit.html", "templates/view.html"))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

type Page struct {
	Title string
	Body  []byte
}

func (p *Page) save() error {
	filename := "pages/" + p.Title + ".txt"
	return os.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := "pages/" + title + ".txt"
	body, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func main() {
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func viewHandler(responseWriter http.ResponseWriter, request *http.Request, title string) {
	page, err := loadPage(title)
	if err != nil {
		http.Redirect(responseWriter, request, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(responseWriter, "view", page)
}

func editHandler(responseWriter http.ResponseWriter, _ *http.Request, title string) {
	page, err := loadPage(title)
	if err != nil {
		page = &Page{Title: title}
	}
	renderTemplate(responseWriter, "edit", page)
}

func saveHandler(responseWriter http.ResponseWriter, request *http.Request, title string) {
	body := request.FormValue("body")
	page := &Page{Title: title, Body: []byte(body)}
	err := page.save()
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
