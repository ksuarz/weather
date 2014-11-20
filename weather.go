/*
A simple weather web application. Outputs weather data from OpenWeatherMap via
a REST interface.
*/
package main

import (
    "strings"
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
    Id int `json:"id"`
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
    FullDescription string
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

func getWeatherDescription(weather WeatherDesc) string {
    switch weather.Id {
        case 200, 230: return "thunderstorms with light rain"
        case 201, 231: return "thunderstorms with rain"
        case 202, 232: return "thunderstorms with heavy rain"
        case 210: return "light thunderstorms"
        case 211: return "thunderstorms"
        case 212: return "heavy thunderstorms"
        case 221: return "ragged thunderstorms"
        case 300, 310: return "light drizzle"
        case 301, 311: return "drizzling rain"
        case 302, 312: return "heavy drizzle"
        case 313, 321: return "showers"
        case 502, 314, 521: return "heavy rain"
        case 520, 522: return "light showers"
        case 531: return "ragged showers"
        case 620: return "light rain and snow"
        case 621: return "rain and snow"
        case 622: return "heavy rain and snow"
        case 731: return "sand and dust whirls"
        case 781: return "tornadoes"
        case 800: return "clear skies"
        case 801: return "a few clouds"
        case 803: return "some broken clouds"
        case 804: return "overcast skies"
        case 900: return "tornadoes"
        case 901: return "tropical storms"
        case 902, 962: return "hurricane conditions"
        case 903: return "extreme cold"
        case 904: return "extreme heat"
        case 905: return "extreme winds"
        case 906: return "extreme hail"
        case 951: return "calm weather"
        case 952: return "light breezes"
        case 953: return "gentle breezes"
        case 954: return "moderate breezes"
        case 955: return "fresh breezes"
        case 956: return "strong breezes"
        case 958: return "windy, gale-like conditions"
        case 959: return "severe gales"
        case 960: return "storms"
        case 961: return "violent storms"
        default: return weather.Description
    }
}

func getFullWeatherDescription(weather []WeatherDesc) string {
    var descs []string = make([]string, len(weather))
    for i := 0; i < len(weather); i = i + 1 {
        descs[i] = getWeatherDescription(weather[i])
    }
    return strings.Join(descs, ", ")
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
    datum.Comparison = getComparison(datum.Main.Temperature, city)
    datum.Main.Temperature = math.Floor(datum.Main.Temperature + 0.5)
    datum.FullDescription = getFullWeatherDescription(datum.Weather)

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
    if err != nil {
        log.Printf("Couldn't get yesterday's data - unmarshaling failed.")
        log.Printf("%v", err)
        return ""
    } else if len(data.List) == 0 {
        log.Printf("API response found no data for yesterday :(")
        return ""
    }

    var datum WeatherData = data.List[min(hour, len(data.List)-1)]
    var yesterday float64 = datum.Main.Temperature - 273.15
    var tft, yft string = getTodayFlavorText(hour), getYesterdayFlavorText(hour)
    var diff float64 = today - yesterday

    log.Printf("Detected temperature difference from yesterday: %f", diff)
    if diff < -5 {
        return tft + " is much colder than " + yft + "."
    } else if diff < -2.5 {
        return tft + " is colder than " + yft + "."
    } else if diff < -1.0 {
        return tft + " is cooler than " + yft + "."
    } else if diff < 1.0 {
        return tft + "'s temperature is similar to " + yft + "."
    } else if diff < 2.5 {
        return tft + " is warmer than " + yft + "."
    } else if diff < 5.0 {
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
