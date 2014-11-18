/*
A simple weather web application. Outputs weather data from OpenWeatherMap via
a REST interface.
*/
package main

import "net/http"

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
        Sunset int `json:"subset"`
    } `json:"sys"`
    Wind struct {
        Speed float64 `json:"speed"` // in km/hr
    } `json:"wind"`
    Main struct {
        Kelvin float64 `json:"temp"`
        MinKelvin float64 `json:"temp_min"`
        MaxKelvin float64 `json:"temp_max"`
        Humidity int `json:"humidity"`
        Pressure int `json:"pressure"`
    } `json:"main"`
}

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        // TODO a nice home page
    })

    // TODO the main weather handling
    http.HandleFunc("/weather/", func(w http.ResponseWriter, r *http.Request) {
        var data weatherData
        var city = strings.SplitN(r.URL.Path, "/", 3)[]

        err = Unmarshal(, &data) // TODO
        if err != nil {
            // TODO Redirect to a custom error page
            http.Redirect(w, r, "/error", http.StatusFound)
            return
        }

        // Render a template?
        renderTemplate(w, "view", ?)
    })

    // Start the server
    http.ListenAndServe(":8080", nil)
}

func loadPage(filename string) (*Page, error) {
    body, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    return &Page(title, body), nil
}
