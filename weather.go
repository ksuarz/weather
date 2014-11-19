/*
A simple weather web application. Outputs weather data from OpenWeatherMap via
a REST interface.
*/
package main

import (
    "errors"
    "encoding/json"
    "html/template"
    "io/ioutil"
    "log"
    "net/http"
    "regexp"
)

type WeatherDesc struct {
    Type string `json:"main"`
    Description string `json:"description"`
}

type WeatherData struct {
    Name string `json:"name"`
    Weather []WeatherDesc
    Sys struct {
        Country string `json:"country"`
        Sunrise int `json:"sunrise"`
        Sunset int `json:"sunset"`
    } `json:"sys"`
    Wind struct {
        Speed float64 `json:"speed"`
    } `json:"wind"`
    Main struct {
        Temperature float64 `json:"temp"`
        Humidity float64 `json:"humidity"`
        Pressure float64 `json:"pressure"`
    } `json:"main"`
}

type WeatherList struct {
    List []WeatherData `json:"list"`
}

var templates = template.Must(template.ParseFiles("templates/index.html", "templates/weather.html"))
var validPath = regexp.MustCompile("^/(weather)/([a-zA-Z0-9]+)$")

func getCity(w http.ResponseWriter, r *http.Request) (string, error) {
    m := validPath.FindStringSubmatch(r.URL.Path)
    if m == nil {
        http.NotFound(w, r)
        return "", errors.New("Invalid Page")
    }

    // First subexpression is "weather"; city is second
    return m[2], nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
    var err error = templates.ExecuteTemplate(w, tmpl+".html", data)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        log.Fatal(err)
    }
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
    renderTemplate(w, "index", nil)
}

func handleWeather(w http.ResponseWriter, r *http.Request) {
    var city string
    var data WeatherList
    var resp *http.Response
    var err error

    // Validate the city name
    city, err = getCity(w, r)
    if err != nil {
        return
    }

    // Query the OpenWeatherMap endpoint
    resp, err = http.Get("http://api.openweathermap.org/data/2.5/find?q=" + city + "&units=metric")
    if err != nil {
        log.Fatal(err)
        return
    }
    // TODO handle *resp.StatusCode
    defer resp.Body.Close()

    // Read in the JSON response
    var buf []byte
    buf, err = ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Fatal(err)
        return
    }

    // Unmarshaling
    err = json.Unmarshal(buf, &data)
    if err != nil {
        log.Fatal(err)
        return
    }

    // Render a template
    renderTemplate(w, "weather", data.List[0])
}

func main() {
    http.HandleFunc("/", handleIndex)
    http.HandleFunc("/weather/", handleWeather)

    // Start the server
    http.ListenAndServe(":8080", nil)
}
