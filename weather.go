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
)

/*
Describes individual weather descriptions:
  - Id: The ID number of the weather condition
  - Type: A string containing the official weather type
  - Description: A longer description of the weather type
  - Icon: The name of an icon available via the API
*/
type WeatherDesc struct {
    Id int `json:"id"`
    Type string `json:"main"`
    Description string `json:"description"`
    Icon string `json:"icon"`
}

/*
A complete data structure describing the weather for a given time.
  - Name: The name of the city
  - CityID: A unique ID number for the city
  - Time: The time, expressed as seconds since the epoch
  - Weather: A list of individual WeatherDesc structures detailing the
    individual weather conditions
  - Sys: An embedded document containing:
    + Country: Either the full country name or a two-letter country code
    + Sunrise: The time of sunrise, expressed as Unix time
    + Sunset: The time of sunset, expressed as Unix time
  - Wind: an embedded document containing:
    + Speed: The wind speed in meters per second
  - Main: an embedded document containing:
    + Temperature: The temperature in either Celsius or Kelvin
    + Humidity: The humidity, as a percentage from 0% to 100$
    + Pressure: The pressure in hPa.
*/
type WeatherData struct {
    Name string `json:"name"`
    CityId int32 `json:"id"`
    Time int64 `json:"dt"`
    Weather []WeatherDesc
    Sys struct {
        Country string `json:"country"`
        Sunrise int64 `json:"sunrise"`
        Sunset int64 `json:"sunset"`
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

/*
A list of weather data points.
*/
type WeatherList struct {
    List []WeatherData `json:"list"`
}

var templates = template.Must(template.ParseFiles("templates/index.html", "templates/weather.html", "templates/notfound.html"))
var validPath = regexp.MustCompile("^/(weather)/([a-zA-Z0-9 ,]+)$")

// Given a URL, returns the city portion of it and an error if it occurs.
func getCity(w http.ResponseWriter, r *http.Request) (string, error) {
    m := validPath.FindStringSubmatch(r.URL.Path)
    if m == nil {
        http.NotFound(w, r)
        return "", errors.New("Invalid Page")
    }

    // First subexpression is "weather"; city is second
    return m[2], nil
}

// Given the weather data, determines whether or not the data was taken during the
// day or the night.
func isDaytime(weather WeatherDesc) bool {
    return strings.Contains(weather.Icon, "d")
}

// Returns a human-readable string that will be grammatically correct for the
// sentences we are constructing.
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

// Given a list of weather descriptions, return their combination in a
// properly-punctuated fashion.
func getFullWeatherDescription(weather []WeatherDesc) string {
    var descs []string = make([]string, len(weather))
    for i := 0; i < len(weather); i = i + 1 {
        descs[i] = getWeatherDescription(weather[i])
    }
    if len(descs) == 1 {
        return descs[0]
    } else {
        return strings.Join(descs[:len(descs)-1], ", ") + " and " + descs[len(descs)-1]
    }
}

// Renders the template found at 'templates/${tmpl}.html'.
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
    datum.Comparison = getComparison(datum.Main.Temperature, isDaytime(datum.Weather[0]), city)
    datum.Main.Temperature = math.Floor(datum.Main.Temperature + 0.5)
    datum.FullDescription = getFullWeatherDescription(datum.Weather)

    // Render a template
    renderTemplate(w, "weather", datum)
}

// Takes today's weather and returns a comparison string determining whether or
// not it is warmer or cooler than yesterday.
func getComparison(today float64, todayIsDay bool, city string) string {
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
    var diff float64 = today - yesterday
    var tft, yft string
    if todayIsDay {
        tft = "Today"
        yft = "yesterday"
    } else {
        tft = "Tonight"
        yft = "last night"
    }

    log.Printf("Detected temperature difference from yesterday: %f", diff)
    if diff < -5 {
        // (-inf, -5)
        return tft + " is much cooler than " + yft + "."
    } else if diff < -2.5 {
        // [-5, -2.5)
        return tft + " is cooler than " + yft + "."
    } else if diff < -1.0 {
        // [-2.5, -1.0)
        return tft + " is slightly cooler than " + yft + "."
    } else if diff < 1.0 {
        // [-1.0, 1.0)
        return tft + "'s temperature is similar to " + yft + "."
    } else if diff < 2.5 {
        // [1.0, 2.5)
        return tft + " is slightly warmer than " + yft + "."
    } else if diff < 5.0 {
        // [2.5, 5.0)
        return tft + " is warmer than " + yft + "."
    } else {
        // [5.0, inf)
        return tft + " is much warmer than " + yft + "."
    }
}

// Returns the minimum of two integers.
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
