package main

import (
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"

	friendlycaptcha "github.com/friendlycaptcha/friendly-captcha-go"
)

type FormMessage struct {
    Subject string
    Message string
}

type TemplateData struct {
	Submitted bool
	Message string
	Sitekey string
}

func main() {
	sitekey := os.Getenv("FRC_SITEKEY")
	apikey := os.Getenv("FRC_APIKEY")

	if sitekey == "" || apikey == "" {
		log.Fatalf("Please set the FRC_SITEKEY and FRC_APIKEY environment values before running this example to your Friendly Captcha sitekey and apikey respectively.")
	}

	frcClient := friendlycaptcha.NewClient(apikey, sitekey)
	tmpl := template.Must(template.ParseFiles("form.html"))

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// GET - the user is requesting the form, not submitting it.
        if r.Method != http.MethodPost {
            tmpl.Execute(w, TemplateData{Submitted: false, Message: "", Sitekey: sitekey})
            return
        }

		formMessage := FormMessage{
			Subject: r.FormValue("subject"),
			Message: r.FormValue("message"),
		}

		solution := r.FormValue(friendlycaptcha.SolutionFormFieldName)
		shouldAccept, err := frcClient.CheckCaptchaSolution(r.Context(), solution)

		if err != nil {
			// Note that there can be errors but we still want to accept the form.
			// The reason is that if Friendly Captcha's API ever goes down, we would rather accept
			// also spammy messages than lock everybody out.

			if errors.Is(err, friendlycaptcha.ErrVerificationFailedDueToClientError) {
				log.Printf("!!!!!\nFriendlyCaptcha is misconfigured! Check your Friendly Captcha API key and sitekey: %s\n", err.Error())
				// Send yourself an alert - the captcha won't be able to do its job to prevent spam.
			} else if (errors.Is(err, friendlycaptcha.ErrVerificationRequest)) {
				log.Printf("Could not talk to the Friendly Captcha API: %s\n", err.Error())
				// Maybe also alert yourself, maybe the Friendly Captcha API is down?
			}
		}

		if !shouldAccept { // The captcha was invalid
			tmpl.Execute(w, TemplateData{Submitted: false, Message: "Anti-robot verification failed, please try again.", Sitekey: sitekey})
			return
		}

		// do something with the data in the form
		_ = formMessage

        tmpl.Execute(w, TemplateData{Submitted: true, Message: "", Sitekey: sitekey})
    })
	log.Printf("Starting server on localhost port 8844 (http://localhost:8844)")

    http.ListenAndServe(":8844", nil)
}