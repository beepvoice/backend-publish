package main;

import (
  "flag"
  "io/ioutil"
  "net/http"
  "log"
  "regexp"
  "strconv"

  "github.com/golang/protobuf/proto"
  "github.com/julienschmidt/httprouter"
  "github.com/nats-io/go-nats"
  "github.com/dgrijalva/jwt-go"
  "github.com/aiden0z/go-jwt-middleware"
)

const MaxBiteSize = 1024 * 1024 * 10

var listen string
var natsHost string
var secret []byte

var nats_conn *nats.Conn

func main() {
  // Parse flags
  var s string
	flag.StringVar(&listen, "listen", ":8080", "host and port to listen on")
  flag.StringVar(&natsHost, "nats", "nats://localhost:4222", "host and port of NATS")
  flag.StringVar(&s, "secret", "secret", "JWT secret")
  flag.Parse()

  secret = []byte(s)

  //NATS
  n, err := nats.Connect(natsHost)
  if err != nil {
    log.Fatal(err)
    return
  }
  nats_conn = n

  // JWT Middleware
  jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options {
    ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
      return secret, nil
    },
    SigningMethod: jwt.SigningMethodHS256,
  })

  // Routes
	router := httprouter.New()

  router.PUT("/conversation/:key/start/:start", PutBite) // bites
  router.PUT("/conversation/:key/start/:start/user", PutBiteUser) // bite_users

  // Start server
  log.Printf("starting server on %s", listen)
	log.Fatal(http.ListenAndServe(listen, jwtMiddleware.Handler(router)))
}

// TODO: ensure security of regexp
var validConversationRegexp = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)

func validConversation(conversation string) bool {
	return validConversationRegexp.MatchString(conversation)
}

func ParseStartString(start string) (uint64, error) {
	return strconv.ParseUint(start, 10, 64)
}

// Route handlers
func PutBite(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
  user := r.Context().Value("user")
  userClaims := user.(*jwt.Token).Claims.(jwt.MapClaims)
  client := Client {
    Key: userClaims["id"].(string),
    Client: userClaims["client"].(string),
  }

  start, err := ParseStartString(p.ByName("start"))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	key := p.ByName("key")
	if !validConversation(key) {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	reader := http.MaxBytesReader(w, r.Body, MaxBiteSize)
	body, err := ioutil.ReadAll(reader)
  if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

  b := Bite {
    Start: start,
    Key: key,
    Data: body,
    Client: &client,
  }
  out, err := proto.Marshal(&b)
  if err != nil {
    http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
    log.Print(err)
    return
	}
  nats_conn.Publish("new_bite", out)

  w.WriteHeader(200)
}

func PutBiteUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
  user := r.Context().Value("user")
  userClaims := user.(*jwt.Token).Claims.(jwt.MapClaims)
  client := Client {
    Key: userClaims["id"].(string),
    Client: userClaims["client"].(string),
  }

  start, err := ParseStartString(p.ByName("start"))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

  key := p.ByName("key")
	if !validConversation(key) {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	reader := http.MaxBytesReader(w, r.Body, MaxBiteSize)
	body, err := ioutil.ReadAll(reader)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

  b := Bite {
    Start: start,
    Key: key,
    Data: body,
    Client: &client,
  }
  out, err := proto.Marshal(&b)
  if err != nil {
    http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
    log.Print(err)
    return
	}
  nats_conn.Publish("new_bite_user", out)

  w.WriteHeader(200)
}
