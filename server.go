package main

import (
  "bytes"
  "encoding/json"
  "github.com/labstack/echo"
  "github.com/gomodule/redigo/redis"
  "net/http"
  "os"
  "time"
)

var (
  ConnPool = newPool()
)

type jsonMELI struct {
  Id string `json:"id"`
  Data json.RawMessage `json:"data"`
}

func newPool() *redis.Pool {
  return &redis.Pool{
    MaxIdle: 100,
    IdleTimeout: 240 * time.Second,
    Dial: func() (redis.Conn, error) {
      return redis.Dial("tcp", os.Getenv("REDIS_HOST") + ":6379")
    },
  }
}

func checkExistance(id string) (bool, bool) {
  conn := ConnPool.Get()
  r, err := redis.String(conn.Do("JSON.GET", id))
  defer conn.Close()

  switch err {
    case nil:
      // Exists 
      return true, false
    default:
      if r == "" {
	// Not exists
        return false, false
      }
      // Something crashed
      return false, true
  }
}

func getJSON(c echo.Context) error {
  id := c.Param("id")
  test, crash := checkExistance(id)

  // TODO: redigo crashing should be resolved somewhere else
  if crash {
    return c.JSON(http.StatusInternalServerError, json.RawMessage("{ \"error\": \"Existance check in Redis failed.\" }"))
  }

  if !test {
    return c.JSON(http.StatusNotFound, json.RawMessage("{ \"error\": \"Element not found.\" }"))
  }

  if test {
    // TODO: querying twice seems wrong 
    conn := ConnPool.Get()
    r, _ := redis.String(conn.Do("JSON.GET", id))
    defer conn.Close()

    // Using ReJSON is not that nice, RawMessage needs a JSON formatted string
    rawp := []byte("{ \"id\": \"" + id + "\"," + "\"data\": " + r + "}")

    return c.JSON(http.StatusOK, json.RawMessage(rawp))
  }

  // TODO: this will never be used but conditionals could be better
  return c.NoContent(http.StatusOK)
}

func postJSON(c echo.Context) error {
  // bytes buffer to get request body
  buf := new(bytes.Buffer)
  buf.ReadFrom(c.Request().Body)

  // json raw message to accept any JSON
  rawp := json.RawMessage(buf.String())

  var jm jsonMELI

  // Try to unmarshal into struct var
  err := json.Unmarshal(rawp, &jm)
  if err != nil {
    return c.JSON(http.StatusInternalServerError, json.RawMessage("{ \"error\": \"" + err.Error() + "\" }"))
  }

  test, crash := checkExistance(jm.Id)

  if test {
    return c.JSON(http.StatusForbidden, json.RawMessage("{ \"error\": \"Element already exists.\" }"))
  }

  if crash {
    return c.JSON(http.StatusInternalServerError, json.RawMessage("{ \"error\": \"Existance check in Redis failed.\" }"))
  }

  if (!test && !crash) {
    conn := ConnPool.Get()
    _, err = conn.Do("JSON.SET", jm.Id, ".", string(jm.Data[:]))
    defer conn.Close()
    if err != nil {
      return c.JSON(http.StatusInternalServerError, json.RawMessage("{ \"error\": \"" + err.Error() + "\" }"))
    }
    return c.JSON(http.StatusCreated, rawp)
  }

  return c.NoContent(http.StatusOK)
}

func putJSON(c echo.Context) error {
  id := c.Param("id")
  test, crash := checkExistance(id)

  if crash {
    return c.JSON(http.StatusInternalServerError, json.RawMessage("{ \"error\": \"Existance check in Redis failed.\" }"))
  }

  if !test {
    return c.JSON(http.StatusNotFound, json.RawMessage("{ \"error\": \"Element not found.\" }"))
  }

  // bytes buffer to get request body
  buf := new(bytes.Buffer)
  buf.ReadFrom(c.Request().Body)

  rawp := jsonMELI{
    Id: id,
    Data: json.RawMessage(buf.String()),
  }

  if test {
    conn := ConnPool.Get()
    _, err := conn.Do("JSON.SET", id, ".", string(rawp.Data[:]))
    defer conn.Close()
    if err != nil {
      return c.JSON(http.StatusInternalServerError, json.RawMessage("{ \"error\": \"" + err.Error() + "\" }"))
    }
    return c.JSON(http.StatusAccepted, rawp)
  }

  return c.NoContent(http.StatusOK)
}

func delJSON(c echo.Context) error {
  id := c.Param("id")
  test, crash := checkExistance(id)

  if crash {
    return c.JSON(http.StatusInternalServerError, json.RawMessage("{ \"error\": \"Existance check in Redis failed.\" }"))
  }

  if !test {
    return c.JSON(http.StatusNotFound, json.RawMessage("{ \"error\": \"Element not found.\" }"))
  }

  if test {
   conn := ConnPool.Get()
    _, err := conn.Do("JSON.DEL", id)
    defer conn.Close()
    if err != nil {
      return c.JSON(http.StatusInternalServerError, json.RawMessage("{ \"error\": \"" + err.Error() + "\" }"))
    }
    return c.JSON(http.StatusOK, json.RawMessage("{ \"message\": \"OK\" }"))
  }

  // This will never be used
  return c.NoContent(http.StatusOK)
}

func status(c echo.Context) error {
  return c.NoContent(http.StatusOK)
}

func main() {
  e := echo.New()
  e.GET("/json/:id", getJSON)
  e.POST("/json", postJSON)
  e.PUT("/json/:id", putJSON)
  e.DELETE("/json/:id", delJSON)
  e.GET("/status", status)
  e.Logger.Fatal(e.Start(":1323"))
}
