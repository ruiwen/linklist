
package main

import (
    "errors"
    "encoding/json"
    "fmt"
    "log"
    "regexp"
    "reflect"
    "strconv"
    "time"
)

// type Model interface {
//     ToJSON() string
// }

type Model struct {
    // Ref: http://dennissuratna.com/using-anonymous-field-to-extend-in-go/
    M interface{}
}

func (m Model) ToJSON() string {
    // var data []byte
    fmt.Printf("Model ToJSON: %s\n", m.M)
    data, err := json.Marshal(m.M)
    if err != nil {
        log.Fatal(err)
    }

    return string(data[:])
}

// Ref: https://medium.com/coding-and-deploying-in-the-cloud/time-stamps-in-golang-abcaf581b72f
type Timestamp time.Time

func (t *Timestamp) MarshalJSON() ([]byte, error) {
    ts := time.Time(*t).Unix()
    stamp := fmt.Sprint(ts)

    return []byte(stamp), nil
}

func (t *Timestamp) UnmarshalJSON(b []byte) error {
    ts, err := strconv.Atoi(string(b))
    if err != nil {
        return err
    }

    *t = Timestamp(time.Unix(int64(ts), 0))

    return nil
}

func (t *Timestamp) Time() time.Time {
    return time.Time(*t)
}

type Update struct {
    // Model
    UpdateId     int      `json:"update_id"`
    Message      Message  `json:"message"`
}

type User struct {
    // Model
    Id           int      `json:"id"`
    FirstName    string   `json:"first_name"`
    LastName     string   `json:"last_name"`
    Username     string   `json:"username"`
}

type Message struct {
    // Model
    MessageId    int        `json:"message_id"`
    From         User       `json:"from"`
    Date         Timestamp  `json:"date,float64"`
    Chat         Chat       `json:"chat"`
    Text         string     `json:"text"`
}

type Chat struct {
    // Model
    Id        int       `json:"id"`
}
// type Chat     int       `json:"id"`

type Link struct {
    // Model
    Id        int       `json:"id"`
    User      User      `json:"user", sql:"json"`
    URL       URL       `json:"url", sql:"json"`
    Created   Timestamp `json:"created"`
    Chat      Chat      `json:"chat"`
    Tags      []string  `json:"tags", sql:"json"`
}

type URL struct {
    // Model
    URL       string    `json:"url"`
    Domain    string    `json:"domain"`
    Path      string    `json:"path"`
    Scheme    string    `json:"scheme"`
    Hash      string    `json:"hash"`
    Search    string    `json:"search"`
}

func (url *URL) Parse () *URL {

    // Receiver is pointer, therefor is *address* of actual URL instance

    var URLRE = regexp.MustCompile(`^(?P<Scheme>https?):\/\/(?P<Domain>[^/?]+)(\/(?P<Path>[^#?]+))?/?(\?(?P<Search>[^#]+))?(#(?P<Hash>.+))?$`)

    match := URLRE.FindStringSubmatch(url.URL)
    if len(match) > 0 {
        for i, name := range URLRE.SubexpNames() {
            if name == "" || match[i] == "" {
                continue
            }

            // Where we'd normally pass &url into reflect.ValueOf(),, since
            // this method's receiver is a pointer to an URL, ie. *URL, we
            // can simply pass that parameter into reflect.ValueOf()
            reflect.ValueOf(url).Elem().FieldByName(name).SetString(match[i])
        }
    }

    return url
}

func (url *URL) Extract (s string) (string, error) {

    var URLEXT = regexp.MustCompile(`https?:\/\/[\w\-.?#/]+`)
    match := URLEXT.FindStringSubmatch(s)
    if len(match) == 0 {
        return "", errors.New("No match")
    } else {
        url.URL = match[0]  // match[0] is the original string
        return match[0], nil
    }

}
