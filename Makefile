# Makefile - compiles the weather server

all: weather

weather: weather.go
	go build weather.go

clean:
	rm -f weather
