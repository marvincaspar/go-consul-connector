# Consul connector

This package registers your go application to consul for service discovery in a microservice architecture.

## Usage

Import this package to your project.

```
import consulconnector "github.com/marvincaspar/go-consul-connector"
```

Define the consul endpoint and the service name.

```
consul := flag.String("consul", "consul:8500", "Consul host")
flag.Parse()

servicename := "hello"
```

Register your service to consul

```
id := fmt.Sprintf("%v-%v-%v", servicename, hostname, *port)
consulClient, _ := consulconnector.NewConsulClient(*consul)
health := fmt.Sprintf("http://%v:%v/api/v1/health", hostname, *port)
consulClient.Register(id, servicename + "-service", hostname, *port, "/api", health)
```

De-register the service on shutdown

```
c := make(chan os.Signal)
signal.Notify(c, os.Kill, os.Interrupt, syscall.SIGTERM)
go func() {
  for sig := range c {
    log.Println("Shutting Down...", sig)
    consulClient.DeRegister(id)
    server.Shutdown(context.Background())
    log.Println("Done...Bye")
    os.Exit(0)
  }
}()
```

## Example

```
package main

import (
	"context"
	"flag"
	"fmt"
	"html"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
  consulconnector "github.com/marvincaspar/go-consul-connector"
)

func main() {
	consul := flag.String("consul", "consul:8500", "Consul host")
	port := flag.Int("port", 8080, "this service port")
	flag.Parse()

  servicename := "hello"
	hostname, _ := os.Hostname()
	log.Println("Starting up... ", hostname, " consul host", *consul, " service  ", *port)

	// Register Service
	id := fmt.Sprintf("%v-%v-%v", servicename, hostname, *port)
	consulClient, _ := consulconnector.NewConsulClient(*consul)
	health := fmt.Sprintf("http://%v:%v/api/v1/health", hostname, *port)
	consulClient.Register(id, servicename + "-service", hostname, *port, "/api", health)

	router := mux.NewRouter().StrictSlash(true)

	// Define Health Endpoint
	router.Methods("GET").Path("/api/v1/health").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		str := fmt.Sprintf("{ 'status':'ok', 'host':'%v:%v' }", hostname, *port)
		fmt.Fprintf(w, str)
		log.Println("/api/v1/health called")
	})

	// The Hello endpoint for the service
	router.Methods("GET").Path("/api/v1/hello").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		str := fmt.Sprintf("Hello, %q at %v:%v\n", html.EscapeString(r.URL.Path), hostname, *port)
		rt := rand.Intn(100)
		time.Sleep(time.Duration(rt) * time.Millisecond)
		fmt.Fprintf(w, str)
		log.Println(str)
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%v", *port),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      router,
	}

	// De-register service at shutdown.
	c := make(chan os.Signal)
	signal.Notify(c, os.Kill, os.Interrupt, syscall.SIGTERM)
	go func() {
		for sig := range c {
			log.Println("Shutting Down...", sig)
			consulClient.DeRegister(id)
			server.Shutdown(context.Background())
			log.Println("Done...Bye")
			os.Exit(0)
		}
	}()

	log.Fatal(server.ListenAndServe())
}
