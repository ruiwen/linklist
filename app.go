package main

import (
    "encoding/json"
    "log"
    "fmt"
    "strings"
    "time"
    // "regexp"
    // "reflect"
    "database/sql"
    _ "github.com/lib/pq"
    "github.com/gin-gonic/gin"
    // "thoughtmonkeys.com/linklist/models"
)



func main() {

    // Connect to DB
    db, err := sql.Open("postgres", "postgres://linklist:linklist@localhost/linklist")
    if err != nil {
        fmt.Printf("[DB] Error: %s", err)
        log.Fatal(err)
    }
    defer db.Close()
    fmt.Printf("Created DB: %s", db)

    // Create a table
    if _, err = db.Exec("CREATE TABLE IF NOT EXISTS links ( id SERIAL PRIMARY KEY, person jsonb, url jsonb, created timestamp with time zone, chat integer, tags jsonb )"); err != nil {
        log.Fatal(err)
    }

    // Supports queries like:
    // SELECT data->'url' FROM links WHERE data @> '{"domain": "github.com"}'
    if rows, err := db.Query("SELECT to_regclass('links.idxdomains')"); rows == nil {
        if _, err = db.Exec("CREATE INDEX idxdomains ON links USING gin (url jsonb_path_ops)"); err != nil {
            log.Fatal(err)
        }
    }
    // Supports queries like:
    // SELECT data->'url' FROM links WHERE data -> 'tags' ? 'kafka'
    // TODO: Implement tags
    // if _, err = db.Exec("CREATE UNIQUE INDEX idxtags ON links USING gin ((data -> 'tags'))"); err != nil {
    //     log.Fatal(err)
    // }


    router := gin.Default()

    router.POST("/retrieve", func(c *gin.Context) {

        update := Update{}
        e := c.BindJSON(&update)
        if e != nil {
            log.Fatal(e)
        }

        if strings.HasPrefix(update.Message.Text, "/retrieve") {
            components := strings.SplitN(update.Message.Text, " ", 5)
            object := components[1]
            field := components[2]
            value := components[3]

            fmt.Printf("Object: -%s-, Field: -%s-, Value: -%s-\n", object, field, value)

            var rows *sql.Rows
            var err error
            if object == "user" {
                fmt.Println("Object: user\n")
                rows, err = db.Query(`SELECT * FROM links WHERE person->>$1 = $2`, field, value)
            } else {
                fmt.Println("Object: others\n")
                rows, err = db.Query(`SELECT * FROM links WHERE url->>$1 = $2`, field, value)
            }

            fmt.Println("Rows retrieved\n")
            fmt.Printf("Error: %s\n", err)
            if err != nil {
                log.Fatal(e)
            }
            defer rows.Close()

            fmt.Println("Scanning..\n")
            links := []Link{}
            var userTmp json.RawMessage
            var urlTmp json.RawMessage
            var createdTmp time.Time
            var tagsTmp sql.NullString
            for rows.Next() {
                var link Link
                err := rows.Scan(
                    &link.Id,
                    &userTmp,
                    &urlTmp,
                    &createdTmp,
                    &link.Chat.Id,
                    &tagsTmp)

                // err := rows.Scan(
                //     &link.Id,
                //     &link.User,
                //     &link.URL,
                //     &createdTmp,
                //     &link.Chat.Id,
                //     &link.Tags)



                if err != nil {
                    log.Fatal(err)
                }

                // link.User = User(userTmp)
                json.Unmarshal(userTmp, &link.User)
                json.Unmarshal(urlTmp, &link.URL)
                fmt.Printf("createdTmp: -%s-\n", createdTmp)
                link.Created = Timestamp(createdTmp)
                fmt.Printf("tags: %s\n", tagsTmp)
                // json.Unmarshal(createdTmp, &link.Created)
                // userTmp.UnmarshalJSON(&link.User)
                // link.URL = *urlTmp
                // link.Created = *createdTmp
                // link.Tags = *tagsTmp

                fmt.Println("Appending.. %s\n", link)
                links = append(links, link)
            }
            c.JSON(200, gin.H{"links": Model{M: &links}})
        }
    })

    router.POST("/save", func(c *gin.Context) {

        update := Update{}
        fmt.Println("pre-bind")
        e := c.BindJSON(&update)
        if e != nil {
            log.Fatal(e)
        }

        fmt.Println(update.Message.Date)

        // update.Message.URL.Parse()
        if strings.HasPrefix(update.Message.Text, "/save") {

            url := URL{}
            _, err = url.Extract(update.Message.Text)
            if err != nil {
                log.Fatal(err)
            }
            url.Parse()

            // Insert into DB
            if _, err = db.Exec("INSERT INTO links (person, url, created, chat) VALUES ($1, $2, $3, $4)", Model{M: &update.Message.From}.ToJSON(), Model{M: &url}.ToJSON(), update.Message.Date.Time(), update.Message.Chat.Id); err != nil {
                log.Fatal(err)
            }
        }

        c.JSON(200, gin.H{"success": true})

    })

    // router.POST("/save", func(c *gin.Context) {

    //     var url URL
    //     e := c.BindJSON(&url)
    //     if e == nil {

    //         url.Parse()

    //         // Insert into DB
    //         var data []byte
    //         if data, err = json.Marshal(url); err != nil {
    //             log.Fatal(err)
    //         }
    //         fmt.Println(string(data[:]))
    //         if _, err = db.Exec("INSERT INTO links (url) VALUES ($1)", data); err != nil {
    //             fmt.Println(err)
    //             log.Fatal(err)
    //         }


    //         c.JSON(200, url)
    //     } else {
    //         fmt.Println(e)
    //     }

    // })


    router.POST("/message", func(c *gin.Context) {

        var message Message
        e := c.BindJSON(&message)
        if e == nil {
            c.JSON(200, gin.H{"from": message.From.Username})
        } else {
            fmt.Println(e)
        }

    })

    router.POST("/user", func(c *gin.Context) {

        var user User
        e := c.BindJSON(&user)
        if e == nil {
            c.JSON(200, user)
        } else {
            fmt.Println(e)
        }

    })

    // router.GET("/:action/user", func(c *gin.Context) {

    //     action := c.Params.ByName("action")
    //     c.JSON(200, gin.H{"action": action, "end": "user"})
    // })

    // router.GET("/:action/duh", func(c *gin.Context) {

    //     action := c.Params.ByName("action")
    //     c.JSON(200, gin.H{"action": action, "end": "duh"})
    // })


    // router.GET("/ping", func(c *gin.Context) {

    //     c.String(200, "pong")
    // })



    router.Run("0.0.0.0:8000")

}
