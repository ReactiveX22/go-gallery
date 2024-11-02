package controllers

import (
	"net/http"
)

func StaticHanlder(tpl Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tpl.Execute(w, r, nil)
	}
}

func FAQ(tpl Template) http.HandlerFunc {
	questions := []struct {
		Question string
		Answer   string
	}{
		{
			Question: "Is there a free version?",
			Answer:   "Yes, we offer a free trial of 30 days.",
		},
		{
			Question: "Can I cancel anytime?",
			Answer:   "Yes, you can cancel your subscription at any time without additional charges.",
		},
		{
			Question: "Do you offer customer support?",
			Answer:   "Yes, we provide 24/7 customer support through live chat and email.",
		},
		{
			Question: "What payment methods are accepted?",
			Answer:   "We accept all major credit cards, PayPal, and direct bank transfers.",
		},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		tpl.Execute(w, r, questions)
	}
}
