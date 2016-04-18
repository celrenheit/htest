package htest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
)

type data struct {
	Mail     string
	Password string
}

func TestRequests(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/path", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			w.Header().Set("foo", "bar")
			fmt.Fprintf(w, "Response")
		}
	})

	mux.HandleFunc("/cookie", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:  "batman",
			Value: "htest",
		})
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/json", func(w http.ResponseWriter, r *http.Request) {
		d := &data{
			Mail:     "test@test.com",
			Password: "pass",
		}

		b, _ := json.Marshal(d)
		w.Write(b)
	})

	mux.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		var d data
		err := json.NewDecoder(r.Body).Decode(&d)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
		status := http.StatusOK
		if d.Password != "pass" {
			status = http.StatusUnauthorized
		}
		w.WriteHeader(status)
		w.Write([]byte(http.StatusText(status)))
	})

	test := New(t, mux)
	test.Request("POST", "/path").
		Do().
		ExpectBody("Response").
		ExpectHeader("foo", "bar").
		ExpectStatus(http.StatusOK)

	test.Request("GET", "/cookie").Do().
		ExpectCookie("batman", "htest")

	test.Request("HEAD", "/path").Do().
		ExpectStatus(http.StatusOK).
		ExpectBody("").
		ExpectHeader("foo", "")

	test.Request("GET", "/unknownpath").Do().ExpectStatus(http.StatusNotFound)

	test.Request("GET", "/json").Do().
		ExpectJSON(&data{
			Mail:     "test@test.com",
			Password: "pass",
		})

	test.Post("/auth").
		Send(&data{
			Mail:     "test@test.com",
			Password: "pass",
		}).
		Do().
		ExpectStatus(http.StatusOK)

	test.Post("/auth").
		Send(&data{
			Mail:     "test@test.com",
			Password: "pass222222",
		}).
		Do().
		ExpectStatus(http.StatusUnauthorized).
		ExpectBody(http.StatusText(http.StatusUnauthorized))
}

func TestExample(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/path", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			w.Header().Set("foo", "bar")
			fmt.Fprintf(w, "Response")
		}
	})

	test := New(t, mux)
	test.Request("POST", "/path").
		Do().
		ExpectBody("Response").
		ExpectHeader("foo", "bar").
		ExpectStatus(http.StatusOK)
}

func TestHandler(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("foo", "bar")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "You are not authorized")
	})

	test := New(t, mux)

	test.Get("/admin").Do().
		ExpectHeader("foo", "bar").
		ExpectStatus(http.StatusUnauthorized).
		ExpectBody("You are not authorized")
}

func TestBuildingRequest(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/path", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("foo", r.Header.Get("foo"))
		io.Copy(w, r.Body)
	})

	test := New(t, mux)

	test.Get("/path").
		// Set a header to the request
		AddHeader("foo", "barbar").
		// Send a string to the request's body
		SendString("my data").

		// Executes the request
		Do().

		// The body sent should stay the same
		ExpectBody("my data").
		// The header sent should stay the same
		ExpectHeader("foo", "barbar")
}
