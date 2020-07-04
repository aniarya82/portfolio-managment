package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"text/template"

	"github.com/gorilla/mux"
)

// ContactDetail is structure to define contact
type ContactDetail struct {
	Email   string
	Subject string
	Message string
}

func submitcontact(w http.ResponseWriter, r *http.Request) {
	forms := template.Must(template.ParseFiles("static/forms.html"))
	details := ContactDetail{
		Email:   r.FormValue("email"),
		Subject: r.FormValue("subject"),
		Message: r.FormValue("message"),
	}
	fmt.Printf("Email : %s\n", details.Email)
	fmt.Printf("Subject : %s\n", details.Subject)
	fmt.Printf("Message : %s\n", details.Message)
	_ = details
	forms.Execute(w, struct{ Success bool }{true})
}

func getcontact(w http.ResponseWriter, r *http.Request) {
	forms := template.Must(template.ParseFiles("static/forms.html"))
	forms.Execute(w, nil)
}

// Todo is structure of defining task
type Todo struct {
	Title string
	Done  bool
}

// TodoPageData is structure to define the template
type TodoPageData struct {
	PageTitle string
	Todos     []Todo
}

func todo(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("static/layout.html"))
	data := TodoPageData{
		PageTitle: "My TODO List",
		Todos: []Todo{
			{Title: "Task 11", Done: false},
			{Title: "Task 21", Done: false},
			{Title: "Task 31", Done: true},
		},
	}
	tmpl.Execute(w, data)
}

func books(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	title := vars["title"]
	page := vars["page"]

	fmt.Fprintf(w, "Your request is %s book at page %s numer", title, page)
}

type pageRoutes struct {
	Src  string
	Text string
}

type pageVariable struct {
	Title  string
	Routes []pageRoutes
}

func indexPage(w http.ResponseWriter, r *http.Request) {
	index := template.Must(template.ParseFiles("static/index.html"))
	allRoutes := []pageRoutes{
		pageRoutes{"/books/hello/page/2", "testing-variables"},
		pageRoutes{"/todo", "todo tasks"},
		pageRoutes{"/contactus", "Contact Us"},
		pageRoutes{"/stocks", "List available stocks"},
	}
	variable := pageVariable{
		Title:  "Index page",
		Routes: allRoutes,
	}
	index.Execute(w, variable)
}

type pageStocks struct {
	Label string
	Text  string
}

type pageVar struct {
	Title  string
	Stocks []pageStocks
}

func listAllStocks(w http.ResponseWriter, r *http.Request) {
	tmplt := template.Must(template.ParseFiles("static/listStocks.html"))
	res, err := http.Get("http://127.0.0.1:5000/api/nifty50/")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	fmt.Println(res.Status)
	resData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	resMod := string(resData)
	allStocks := []pageStocks{}
	json.Unmarshal([]byte(resMod), &allStocks)

	// fmt.Println(allStocks)
	resStr := pageVar{
		Title:  "List Stocks",
		Stocks: allStocks,
	}
	// fmt.Fprintf(w, resStr)
	tmplt.Execute(w, resStr)
}

type portfolio struct {
	Tickers  []string
	Duration []string
	Weights  []string
}

type standardTicker struct {
	Ticker string
	Weight float32
}

type standardPortfolio struct {
	Return    float64
	Variance  float64
	Portfolio []standardTicker
}

var finalPort standardPortfolio

func submitStocks(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fmt.Println(r.Form["stocks"])
	fmt.Println(r.Form["duration"])
	fmt.Println(r.Form["weights"])
	submitPort := portfolio{
		Tickers:  r.Form["stocks"],
		Duration: r.Form["duration"],
		Weights:  r.Form["weights"],
	}
	jsonVal, _ := json.Marshal(submitPort)
	u := bytes.NewBuffer(jsonVal)
	req, err := http.NewRequest("POST", "http://127.0.0.1:5000/api/savePortfolio", u)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	fmt.Println("Response Status : ", res.Status)
	body, _ := ioutil.ReadAll(res.Body)
	resBody := string(body)
	fmt.Println("Response body: ", resBody)
	json.Unmarshal([]byte(resBody), &finalPort)
	// fmt.Fprintf(w, "Testing")
	http.Redirect(w, r, "/portfolioStats", 302)
	// http.Redirect(w, r, "/selected", 302)
}

func portfolioStats(w http.ResponseWriter, r *http.Request) {
	tmplt := template.Must(template.ParseFiles("static/portfolioStats.html"))
	tmplt.Execute(w, finalPort)
}

func main() {
	r := mux.NewRouter()
	// Static and assets dir loading
	fs := http.FileServer(http.Dir("static/"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	r.HandleFunc("/books/{title}/page/{page}", books)
	r.HandleFunc("/todo", todo)
	r.HandleFunc("/", indexPage)
	r.HandleFunc("/stocks", listAllStocks).Methods("GET")
	r.HandleFunc("/stocks", submitStocks).Methods("POST")
	r.HandleFunc("/contactus", getcontact).Methods("GET")
	r.HandleFunc("/contactus", submitcontact).Methods("POST")
	// r.HandleFunc("/selected", displaySelected).Methods("GET")
	// r.HandleFunc("/selected", submitSelected).Methods("POST")
	r.HandleFunc("/portfolioStats", portfolioStats).Methods("GET")

	http.ListenAndServe(":80", r)
}
