package main

import (
	mail "email-service/emailModel"
	"email-service/service"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

var con = service.Connection{}

func init() {
	con.Server = "mongodb://localhost:27017"
	con.Database = "EmailService"
	con.Collection = "email_data"

	con.Connect()
}

func sendEmail(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != "POST" {
		respondWithError(w, http.StatusBadRequest, "Invalid method")
		return
	}

	var mailDetails mail.Mail

	if err := json.NewDecoder(r.Body).Decode(&mailDetails); err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("%v", err))
		return
	}

	if len(mailDetails.MailSendTo) == 0 || mailDetails.MailBody == nil {
		respondWithError(w, http.StatusBadGateway, "Please enter emailTo and email body")
		return
	}

	if result, err := con.EmailWithoutAttachment(mailDetails); err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("%v", err))
	} else {
		respondWithJson(w, http.StatusAccepted, map[string]string{
			"message": result,
		})
	}
}

func sendEmailWithAttachment(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != "POST" {
		respondWithError(w, http.StatusBadRequest, "Invalid method")
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "The uploaded file is too big. Please choose an file that's less than 1MB in size", http.StatusBadRequest)
		return

	}
	files := r.MultipartForm.File["file"]
	request := r.MultipartForm.Value["request"][0]
	attachment := false
	if r.MultipartForm.Value["attachment"][0] == "yes" {
		attachment = true
	}

	var mailDetails mail.Mail
	err := json.Unmarshal([]byte(request), &mailDetails)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("%v", err))
		return
	}

	if len(mailDetails.MailSendTo) == 0 || mailDetails.MailBody == nil {
		respondWithError(w, http.StatusBadGateway, "Please enter email to and email body")
		return
	}

	if result, err := con.EmailWithAttachMent(mailDetails, files, attachment); err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("%v", err))
	} else {
		respondWithJson(w, http.StatusAccepted, map[string]string{
			"message": result,
		})
	}
}

func searchFilter(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != "POST" {
		respondWithError(w, http.StatusBadRequest, "Invalid method")
		return
	}

	var search mail.Search

	if err := json.NewDecoder(r.Body).Decode(&search); err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("%v", err))
		return
	}

	if result, err := con.SearchFilter(search); err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("%v", err))
	} else {
		respondWithJson(w, http.StatusAccepted, result)
	}
}

func search(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != "GET" {
		respondWithError(w, http.StatusBadRequest, "Invalid method")
		return
	}

	path := r.URL.Path
	segments := strings.Split(path, "/")
	id := segments[len(segments)-1]

	if result, err := con.SearchByEmailId(id); err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("%v", err))
	} else {
		respondWithJson(w, http.StatusAccepted, result)
	}
}

func respondWithJson(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	respondWithJson(w, code, map[string]string{"error": msg})
}

func main() {
	// http.HandleFunc("/send-email", sendEmail)
	http.HandleFunc("/search", searchFilter)
	http.HandleFunc("/search-by-emailId/", search)
	http.HandleFunc("/send-email", sendEmailWithAttachment)
	fmt.Println("Service Started at 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
