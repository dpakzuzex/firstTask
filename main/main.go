package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"

	"../config"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

var (
	htmlPage *template.Template
)

func main() {

	htmlPage = template.Must(template.ParseGlob("index.html"))
	r := mux.NewRouter()
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("error get config: ", err)
	}
	fmt.Println("Server start at :8080")

	err = config.PostgresConnect()
	if err != nil {
		log.Printf("Error database postgres connect: %+v", err)
		return
	}
	fmt.Println("Database postgres connect:", viper.GetString("postgres.host"))

	_, err = config.NewRedisClient()
	if err != nil {
		log.Printf("Error database redis connect: %+v", err)
		return
	}
	fmt.Println("Database redis connect:", viper.GetString(`redis.host`))

	r.HandleFunc("/", handler).Methods("GET")
	r.HandleFunc("/api/first", buttonHandler).Methods("POST")
	r.HandleFunc("/api/second", jwtHandler).Methods("GET")
	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	// fmt.Fprint(w, "Hello World!")
	htmlPage.ExecuteTemplate(w, "index.html", nil)
}

func buttonHandler(w http.ResponseWriter, r *http.Request) {
	tokenString, err := config.GenerateJWT()
	if err != nil {
		fmt.Println("Error buttonHandler #1", err)
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://localhost:8080/api/second", nil)
	if err != nil {
		fmt.Fprintf(w, "Error: %s", err.Error())
	}

	req.Header.Set("Token", tokenString)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error buttonHandler #2: ", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error buttonHandler #3:", err)
	}

	fmt.Fprintf(w, string(body))

}

func jwtHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header["Token"] != nil {
		token, err := jwt.Parse(r.Header["Token"][0], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("error while parsing")
			}
			return []byte(viper.GetString("keyJWT")), nil
		})
		if err != nil {
			fmt.Fprintf(w, err.Error())
		}

		if token.Valid {
			fmt.Fprint(w, "Hello World!")
		} else {
			fmt.Fprint(w, "You don't have a token")
		}
	} else {
		fmt.Fprint(w, "You don't have a token")
	}
}
