package main

import "net/http"

type weatherData struct {
    Name string `json:"name"`
    Main struct {
        Kelvin float64 `json:"temp"`
    } `json:"main"`
}

func main() {
    http.HandleFunc("/weather/", func(w http.ResponseWriter, r *http.Request) {
        var 
        city := strings.SplitN(r.URL.Path, "/", 3)[]
        
    })
    http.ListenAndServe(":8080", nil)
}

func loadPage(title string) (*Page, error) {
    filename := title + ".txt"
    body, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    return &Page(title, body), nil
}
