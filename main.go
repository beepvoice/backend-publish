package main;

import (
  "encoding/binary"
  "errors"
  "flag"
  "io/ioutil"
  "net/http"
  "log"
  "regexp"
  "strconv"

  "github.com/golang/protobuf/proto"
  "github.com/julienschmidt/httprouter"
  "github.com/nats-io/go-nats"
)

const MaxBiteSize = 1024 * 1024 * 10

var listen string
var natsHost string

var nats_conn *nats.Conn

func main() {
  // Parse flags
	flag.StringVar(&listen, "listen", ":8080", "host and port to listen on")
  flag.StringVar(&natsHost, "nats", "nats://localhost:4222", "host and port of NATS")
  flag.Parse()

  //NATS
  n, err := nats.Connect(natsHost)
  if err != nil {
    log.Fatal(err)
    return
  }
  nats_conn = n

  // Routes
	router := httprouter.New()

  router.PUT("/conversation/:key/start/:start", PutBite) // bites
  router.PUT("/conversation/:key/start/:start/user", PutBiteUser) // bite_users

  // Start server
  log.Printf("starting server on %s", listen)
	log.Fatal(http.ListenAndServe(listen, router))
}

// Marshalling keys and assorted helper functions
func validObj(obj string) bool {
	return obj == "bite" || obj == "user"
}

// TODO: ensure security of regexp
var validConversationRegexp = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)

func validConversation(conversation string) bool {
	return validConversationRegexp.MatchString(conversation)
}

const conversationSeprator = '@'
const objSeprator = '+'

func MarshalKey(obj, conversation string, start uint64) ([]byte, error) {
	prefixBytes, err := MarshalKeyPrefix(obj, conversation)
	if err != nil {
		return nil, err
	}

	startBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(startBytes, start)

	return append(prefixBytes, startBytes...), nil
}

func MarshalKeyPrefix(obj, conversation string) ([]byte, error) {
	if !validObj(obj) || !validConversation(conversation) {
		return nil, errors.New("main: FormatKey: bad obj or conversation")
	}
	return []byte(obj + string(objSeprator) + conversation + string(conversationSeprator)), nil
}

func ParseStartString(start string) (uint64, error) {
	return strconv.ParseUint(start, 10, 64)
}

// Route handlers
func PutBite(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
  start, err := ParseStartString(p.ByName("start"))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	key, err := MarshalKey("bite", p.ByName("key"), start)
	if err != nil {
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
  start, err := ParseStartString(p.ByName("start"))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	key, err := MarshalKey("user", p.ByName("key"), start)
	if err != nil {
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
