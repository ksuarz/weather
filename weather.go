/*
A simple weather web application. Outputs weather data from OpenWeatherMap via
a REST interface.
*/
package main

import (
    "encoding/json"
    "errors"
    "html/template"
    "io/ioutil"
    "log"
    "math"
    "net/http"
    "regexp"
    "time"
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
    Comparison string
}

type WeatherList struct {
    List []WeatherData `json:"list"`
}

var templates = template.Must(template.ParseFiles("templates/index.html", "templates/weather.html", "templates/notfound.html"))
var validPath = regexp.MustCompile("^/(weather)/([a-zA-Z0-9 ,]+)$")

func getCity(w http.ResponseWriter, r *http.Request) (string, error) {
    m := validPath.FindStringSubmatch(r.URL.Path)
    if m == nil {
        http.NotFound(w, r)
        return "", errors.New("Invalid Page")
    }

    // First subexpression is "weather"; city is second
    return m[2], nil
}

func getTodayFlavorText(hour int) string {
    if hour < 5 || hour > 21 {
        return "Tonight"
    } else if hour >= 9 && hour < 12 {
        return "Today"
    } else if hour >= 12 && hour <= 17 {
        return "This afternoon"
    } else {
        return "This evening"
    }
}

func getYesterdayFlavorText(hour int) string {
    if hour < 5 || hour > 21 {
        return "last night"
    } else {
        return "yesterday"
    }
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

func handleNotFound(w http.ResponseWriter, r *http.Request) {
    renderTemplate(w, "notfound", nil)
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

    // If no data, then city not found
    if len(data.List) == 0 {
        http.Redirect(w, r, "/notfound", http.StatusNotFound)
        return
    }

    // Data sanitization
    var datum WeatherData = data.List[0]
    //data.List[0].Main.Temperature = math.Floor(data.List[0].Main.Temperature+0.5)
    datum.Comparison = getComparison(datum.Main.Temperature, city)
    datum.Main.Temperature = math.Floor(datum.Main.Temperature + 0.5)

    // Render a template
    renderTemplate(w, "weather", datum)
}

func getComparison(today float64, city string) string {
    var hour int = time.Now().Hour()
    var resp *http.Response
    var err error
    var data WeatherList

    // Query the historical data endpoint
    resp, err = http.Get("http://api.openweathermap.org/data/2.5/history/city?q=" + city + "&type=day&units=metric")
    if err != nil {
        log.Printf("Couldn't get yesterday's data - querying failed.")
        log.Printf("%v", err)
        return ""
    }
    defer resp.Body.Close()

    // Read JSON
    var buf []byte
    buf, err = ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Printf("Couldn't get yesterday's data - reading JSON failed.")
        log.Printf("%v", err)
        return ""
    }

    // Unmarshal
    err = json.Unmarshal(buf, &data)
    if err != nil || len(data.List) == 0 {
        log.Printf("Couldn't get yesterday's data - unmarshaling failed.")
        log.Printf("%v", err)
        return ""
    }

    var datum WeatherData = data.List[min(hour, len(data.List)-1)]
    var yesterday float64 = datum.Main.Temperature - 273.15
    var tft, yft string = getTodayFlavorText(hour), getYesterdayFlavorText(hour)
    var diff float64 = today - yesterday

    log.Printf("Detected temperature difference from yesterday: %f", diff)
    if diff < -10.0 {
        return tft + " is much colder than " + yft + "."
    } else if diff < -5.0 {
        return tft + " is colder than " + yft + "."
    } else if diff < -1.0 {
        return tft + " is cooler than " + yft + "."
    } else if diff < 1.0 {
        return tft + "'s temperature is similar to " + yft + "."
    } else if diff < 5.0 {
        return tft + " is warmer than " + yft + "."
    } else if diff < 10.0 {
        return tft + " is hotter than " + yft + "."
    } else {
        return tft + " is much hotter than " + yft + "."
    }
}

func min(x, y int) int {
    if x < y {
        return x
    } else {
        return y
    }
}

func main() {
    http.HandleFunc("/", handleIndex)
    http.HandleFunc("/weather/", handleWeather)
    http.HandleFunc("/notfound/", handleNotFound)

    // Start the server
    http.ListenAndServe(":8080", nil)
}
