package show

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"tracker/server/auth"
	"tracker/templates"
	"tracker/trackable/common"

	"github.com/gorilla/mux"
)

const DEVMODE = false

// Frontend implemnts server.Frontend
type Frontend struct {
	name      string
	host      *common.Host
	handler   Handler
	templates *template.Template
}

func (f *Frontend) RegisterHandlers(subdomain string) {
	rtr := mux.NewRouter()
	rtr.HandleFunc(fmt.Sprintf("/%s/", subdomain), f.listRequest)
	rtr.HandleFunc(fmt.Sprintf("/%s/request", subdomain), f.addShowRequest)
	rtr.HandleFunc(fmt.Sprintf("/%s/schedule", subdomain), f.scheduleRequest)
	rtr.HandleFunc(fmt.Sprintf("/%s/{type:[a-z]+}", subdomain), f.listRequest)
	rtr.HandleFunc(fmt.Sprintf("/%s/{id:[0-9]+}", subdomain), f.detailRequest)
	rtr.HandleFunc(fmt.Sprintf("/%s/request/list", subdomain), f.requestedShowsRequest)

	http.Handle(fmt.Sprintf("/%s/", subdomain), rtr)
}

func (f *Frontend) Init(host *common.Host) error {
	fmt.Println("Show Frontend Initialised")
	f.host = host

	// Define all template functions
	funcMap := template.FuncMap{
		"mod":          templates.Mod,
		"doubleDigits": templates.DoubleDigits,
	}

	// Load all templates
	f.templates = template.Must(template.New("main").Funcs(funcMap).ParseGlob(
		"templates/shows/*.html"))

	return nil
}

// Reload is only used for debugging/dev purposes. It will reinitialize the frontend each time it's
// called. This helps with development as we don't have to restart the server to see updates in
// the templates
func (f *Frontend) Reload() {
	if DEVMODE {
		f.Init(f.host)
	}
}

func (f *Frontend) listRequest(w http.ResponseWriter, r *http.Request) {
	f.Reload()

	params := mux.Vars(r)
	listType, ok := params["type"]
	if !ok {
		listType = "all"
	}

	apiURL := fmt.Sprintf("%s/api/show/get/list/%s", f.host.Address(), listType)
	resp, err := http.Get(apiURL)
	if err != nil {
		fmt.Println(err)
		return
	}

	decode := json.NewDecoder(resp.Body)
	var jsonRep ShowList
	decode.Decode(&jsonRep)

	user, err := auth.CurrentUser(r)
	if err != nil {
		fmt.Println("Error getting current user: %v\n", err)
	}

	data := struct {
		ShowList
		User auth.User
	}{jsonRep, user}

	err = f.templates.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		serveError(err, w)
	}
}

func (f *Frontend) detailRequest(w http.ResponseWriter, r *http.Request) {
	f.Reload()

	params := mux.Vars(r)
	id := params["id"]

	apiURL := fmt.Sprintf("%s/api/show/get/%s", f.host.Address(), id)
	resp, err := http.Get(apiURL)
	if err != nil {
		serveError(err, w)
		return
	}

	decode := json.NewDecoder(resp.Body)
	var jsonRep ShowFull
	decode.Decode(&jsonRep)

	user, err := auth.CurrentUser(r)
	if err != nil {
		fmt.Println("Error getting current user: %v\n", err)
	}

	data := struct {
		ShowFull
		User auth.User
	}{jsonRep, user}

	err = f.templates.ExecuteTemplate(w, "detail.html", data)
	if err != nil {
		serveError(err, w)
	}
}

func (f *Frontend) scheduleRequest(w http.ResponseWriter, r *http.Request) {
	f.Reload()

	curDate := common.CurrentDate()
	startDate := curDate.Minus(7 + curDate.Weekday())
	endDate := startDate.Plus((7 * 7) - 1)

	apiUrl := fmt.Sprintf("%s/api/show/get/schedule/%s/%s", f.host.Address(), startDate, endDate)
	resp, err := http.Get(apiUrl)
	if err != nil {
		fmt.Println(err)
		return
	}

	decode := json.NewDecoder(resp.Body)
	var jsonRep Schedule
	decode.Decode(&jsonRep)

	user, err := auth.CurrentUser(r)
	if err != nil {
		fmt.Println("Error getting current user: %v\n", err)
	}

	data := struct {
		Schedule
		User auth.User
	}{jsonRep, user}

	err = f.templates.ExecuteTemplate(w, "schedule.html", data)
	if err != nil {
		serveError(err, w)
	}
}

func (f *Frontend) loginRequest(w http.ResponseWriter, r *http.Request) {
	f.Reload()

	err := f.templates.ExecuteTemplate(w, "login.html", nil)
	if err != nil {
		serveError(err, w)
	}
}

func (f *Frontend) addShowRequest(w http.ResponseWriter, r *http.Request) {
	f.Reload()

	user, err := auth.CurrentUser(r)
	if err != nil {
		fmt.Printf("Error getting current user: %v\n", err)
	}

	data := struct {
		User auth.User
	}{user}

	err = f.templates.ExecuteTemplate(w, "add_show.html", data)
	if err != nil {
		serveError(err, w)
	}
}

func (f *Frontend) requestedShowsRequest(w http.ResponseWriter, r *http.Request) {
	f.Reload()

	user, err := auth.CurrentUser(r)
	if err != nil {
		fmt.Printf("Error getting current user: %v\n", err)
	}

	showRequests := make([]*ShowRequest, 0)
	if user.Username != "" && user.Admin {
		apiUrl := fmt.Sprintf("%s/api/show/request/list", f.host.Address())
		resp, err := http.Get(apiUrl)
		if err != nil {
			fmt.Println(err)
			return
		}

		decode := json.NewDecoder(resp.Body)
		decode.Decode(&showRequests)
	}

	data := struct {
		User     auth.User
		Requests []*ShowRequest
	}{user, showRequests}

	err = f.templates.ExecuteTemplate(w, "requested_shows.html", data)
	if err != nil {
		serveError(err, w)
	}
}
