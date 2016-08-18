package main

import (
	"encoding/json"
	//"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

// Set up a global string for our secret
var mySigningKey []byte

// = []byte("secret")

// Product will contain information about the products
type Product struct {
	ID          int
	Name        string
	Slug        string
	Description string
}

/* We will create our catalog of VR experiences and store them in a slice. */
var products = []Product{
	Product{ID: 1, Name: "Hover Shooters", Slug: "hover-shooters", Description: "Shoot your way to the top on 14 different hoverboards"},
	Product{ID: 2, Name: "Ocean Explorer", Slug: "ocean-explorer", Description: "Explore the depths of the sea in this one of a kind underwater experience"},
	Product{ID: 3, Name: "Dinosaur Park", Slug: "dinosaur-park", Description: "Go back 65 million years in the past and ride a T-Rex"},
	Product{ID: 4, Name: "Cars VR", Slug: "cars-vr", Description: "Get behind the wheel of the fastest cars in the world."},
	Product{ID: 5, Name: "Robin Hood", Slug: "robin-hood", Description: "Pick up the bow and arrow and master the art of archery"},
	Product{ID: 6, Name: "Real World VR", Slug: "real-world-vr", Description: "Explore the seven wonders of the world in VR"},
}

func init() {
	err := godotenv.Load()
	if err != nil {
		fmt.Print("Error loading .env file")
		return
	}
	mySigningKey = []byte(os.Getenv("CLIENT_SECRET"))
}

func main() {
	// Here we are instantiating the gorilla/mux router
	r := mux.NewRouter()

	// On the default page we will simply serve our static index page.
	r.Handle("/", http.FileServer(http.Dir("./views/")))
	// We will setup our server so we can serve static assest like images, css from the /static/{file} route
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	// Our API is going to consist of three routes
	// /status - which we will call to make sure that our API is up and running
	// /stuff - which will retrieve a list of products that the user can leave feedback on
	// /stuff/{slug}/feedback - which will capture user feedback on products
	r.Handle("/status", StatusHandler).Methods("GET")
	r.Handle("/stuff", jwtMiddleware.Handler(ProductsHandler)).Methods("GET")
	r.Handle("/stuff/{slug}/feedback", jwtMiddleware.Handler(AddFeedbackHandler)).Methods("POST")

	r.Handle("/get-token", GetTokenHandler).Methods("GET")

	// Our application will run on port 3000. Here we declare the port and pass in our router.
	http.ListenAndServe(":3000", handlers.LoggingHandler(os.Stdout, r))
}

// NotImplemented is called whenever an API endpoint is hit we will simply
// return the message "Not Implemented"
var NotImplemented = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Not Implemented"))
})

// StatusHandler will be invoked when the user calls the /status route. It will
// simply return a string with the message "API is up and running"
var StatusHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("API is up and running"))
})

// ProductsHandler will be called when the user makes a GET request to the
// products endpoint. This handler will return a list of products available for
// users to review.
var ProductsHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// Here we are converting the slice of products to json
	payload, _ := json.Marshal(products)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(payload))
})

// AddFeedbackHandler will add either positive or negative feedback to the
// product We would normally save this data to the database - but for this demo
// we'll fake it so that as long as the request is successful and we can match a
// product to our catalog of products we'll return an OK status.
var AddFeedbackHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	var product Product
	vars := mux.Vars(r)
	slug := vars["slug"]

	for _, p := range products {
		if p.Slug == slug {
			product = p
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if product.Slug != "" {
		payload, _ := json.Marshal(product)
		w.Write([]byte(payload))
	} else {
		w.Write([]byte("Product Not Found"))
	}
})

// GetTokenHandler returns a token for secured endpoints
var GetTokenHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	/* Create the token */
	token := jwt.New(jwt.SigningMethodHS256)
	//token := jwt.New(jwt.SigningMethodRS512)

	claims := make(jwt.MapClaims)
	claims["admin"] = false
	claims["name"] = "bob"
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix() // settings.Get().JWTExpirationDelta
	claims["iat"] = time.Now().Unix()
	token.Claims = claims

	// Set token claims
	//token.Claims["admin"] = false
	//token.Claims["name"] = "bob"
	//token.Claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	// Sign the token with our secret
	tokenString, _ := token.SignedString(mySigningKey)

	// Write the token to the browser
	w.Write([]byte(tokenString))
})

var jwtMiddleware = jwtmiddleware.New(jwtmiddleware.Options{
	ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
		return mySigningKey, nil
		//fmt.Println("Client secret: ", os.Getenv("CLIENT_SECRET"))
		// decoded, err := base64.URLEncoding.DecodeString(os.Getenv("CLIENT_SECRET"))
		// if err != nil {
		//     return nil, err
		// }
		//sec := os.Getenv("CLIENT_SECRET")
		//return []byte(sec), nil
	},
	SigningMethod: jwt.SigningMethodHS256,
})
