package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/orbitool/faucet"
	"github.com/orbitool/httpcaptcha"
	"github.com/rs/cors"
	"github.com/urfave/cli/v2"
	"log"
	"math/big"
	"net/http"
	"os"
	"time"
)

func main() {
	app := cli.NewApp()
	app.Name = "faucet"
	app.Description = "configurable ethereum testnet faucet"
	app.Flags = []cli.Flag{
		&cli.StringFlag{Name: "rpc", Usage: "url of your networks rpc endpoint", Required: true,
			EnvVars: []string{"FAUCET_RPC"}},
		&cli.StringFlag{Name: "key", Usage: "private key of the faucets wallet", Required: true,
			EnvVars: []string{"FAUCET_KEY"}},
		&cli.Float64Flag{Name: "amount", Usage: "amount of eth that can be claimed, value in ETH", Value: 0.01,
			EnvVars: []string{"FAUCET_AMOUNT"}},
		&cli.IntFlag{Name: "port", Usage: "the port number this service will be served on", Value: 8080,
			EnvVars: []string{"PORT"}},
		&cli.IntFlag{Name: "delay", Usage: "amount of time in seconds a user must wait before making another claim", Value: 43200,
			EnvVars: []string{"FAUCET_DELAY"}},
		&cli.BoolFlag{Name: "captcha", Usage: "require a captcha to be solved for each claim request", Value: false,
			EnvVars: []string{"FAUCET_CAPTCHA"}},
		&cli.IntFlag{Name: "chain", Usage: "chain id. If not provided will be fetched from the rpc provider",
			EnvVars: []string{"FAUCET_CHAIN"}},
		&cli.StringFlag{Name: "cors", Usage: "set allowed origin", Value: "*",
			EnvVars: []string{"FAUCET_CORS"}},
		&cli.IntFlag{Name: "proxy-count", Usage: "the number of proxies in front of this service", Value: 0,
			EnvVars: []string{"FAUCET_PROXY_COUNT"}},
	}

	app.Action = Serve

	log.Fatal(app.Run(os.Args))
}

func Serve(ctx *cli.Context) error {

	// 1. Create the facuet config
	config := &faucet.Config{
		PrivateKey: ctx.String("key"),
		Provider:   ctx.String("rpc"),
		Amount:     faucet.ToWei(ctx.Float64("amount"), 18),
		Delay:      time.Duration(ctx.Int("delay")) * time.Second,
		ProxyCount: ctx.Int("proxy-count"),
	}

	if chain := ctx.Int("chain"); chain != 0 {
		config.ChainID = big.NewInt(int64(chain))
	}

	// 2. Initialize the faucet
	f, err := faucet.New(config)
	if err != nil {
		return err
	}

	// 3. Setup the http router
	router := httprouter.New()
	router.GlobalOPTIONS = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Access-Control-Request-Method") != "" {
			// Set CORS headers
			header := w.Header()
			header.Set("Access-Control-Allow-Methods", header.Get("Allow"))
			header.Set("Access-Control-Allow-Origin", "*")
		}

		// Adjust status code to 204
		w.WriteHeader(http.StatusNoContent)
	})

	// 4. Optionally use a captcha to protect the routes
	var handler http.Handler = f
	if ctx.Bool("captcha") {
		captcha := httpcaptcha.New(nil)
		router.HandlerFunc("GET", "/captcha/create", captcha.Create)
		router.HandlerFunc("GET", "/captcha/challenge/:media", captcha.Challenge)
		handler = captcha.Middleware(f)
	}
	router.Handler("GET", "/claim", handler)
	router.Handler("GET", "/health", serveString("healthy"))
	router.Handler("GET", "/address", serveString(f.Address()))
	router.Handler("GET", "/amount", serveString(faucet.ToEth(f.Config.Amount, 18).String()))

	// 5. Start the http server :)
	errChan := make(chan error)
	go func() {
		addr := fmt.Sprintf(":%d", ctx.Int("port"))

		errChan <- http.ListenAndServe(addr, logger(cors.AllowAll().Handler(router)))
	}()

	// 6. Print some usefull information about the server to the console
	fmt.Printf("Faucet: running at http://localhost:%d ...\n", ctx.Int("port"))
	fmt.Println("\nRoutes: ")
	printRoute("GET", "/address", "address of the faucets wallet")
	printRoute("GET", "/claim?address=<address>", "Claim test eth to the provided address")
	printRoute("GET", "/health", "returns status 200 if service is running")
	printRoute("GET", "/amount", "returns the amount of eth given on claim")

	if ctx.Bool("captcha") {
		printRoute("GET", "/captcha/create", "Get a new captcha id")
		printRoute("GET", "/captcha/challenge/<id>.png", "Get a captcha image to solve")
		fmt.Println("\nNote: The '/claim' path is protected by a captcha. Requiring the following headers:")
		fmt.Println("     - 'X-Captha' providing the captcha id ")
		fmt.Println("     - 'X-Captha-Solution' providing the captcha solution")
	}

	fmt.Printf("\nNote: Periodically ensure the address '%s' is topped up to keep the service running.\n", f.Address())
	fmt.Println("\n--- Listening for requests ---")

	// block
	return <-errChan
}

func serveString(s string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, s)
	})
}

func printRoute(method, name, desc string) {
	fmt.Printf(" - Route: [%s] %-27s > %s\n", method, name, desc)
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header()
		rw := &responseWriter{ResponseWriter: w, statusCode: 200}
		next.ServeHTTP(rw, r)
		log.Printf("[%d] %s %s", rw.statusCode, r.Method, r.URL)
	})
}
