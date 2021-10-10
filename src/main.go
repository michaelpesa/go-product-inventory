package main

import (
	"fmt"
  "os"
	"html/template"
	"log"
	"net/http"
  "time"

	"github.com/newrelic/go-agent/v3/newrelic"
)

func main() {
	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName(os.Getenv("APP_NAME") + "_" + os.Getenv("ENV_NAME")),
		newrelic.ConfigLicense(os.Getenv("NR_LICENSE_KEY")),
		newrelic.ConfigDistributedTracerEnabled(true),
	)
	if err != nil {
		panic(err)
	}

	db := database{"shoes": 50, "socks": 5}
	http.HandleFunc(newrelic.WrapHandleFunc(app, "/list", db.list))
	http.HandleFunc(newrelic.WrapHandleFunc(app, "/price", db.price))
	http.HandleFunc(newrelic.WrapHandleFunc(app, "/update", db.update))
	log.Fatal(http.ListenAndServe("0.0.0.0:8000", nil))
}

type dollars float32

func (d dollars) String() string {
	return fmt.Sprintf("$%.2f", d)
}

type database map[string]dollars

const dbListTemplate = `
<table>
<tr style='text-align: left'>
  <th>Item</th>
  <th>Price</th>
</tr>
{{range $item, $price := . -}}
<tr>
  <td>{{$item}}</td>
  <td>{{$price}}</td>
</tr>
{{end -}}
</table>
`

func (db database) list(w http.ResponseWriter, req *http.Request) {
	t := template.Must(template.New("dbListTemplate").Parse(dbListTemplate))
	if err := t.Execute(w, db); err != nil {
		fmt.Fprintf(w, "failed to parse template: %v", err)
	}
}

func (db database) price(w http.ResponseWriter, req *http.Request) {
	item := req.URL.Query().Get("item")
	price, ok := db[item]
	if !ok {
    time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "no such item: %q", item)
		return
	}
	fmt.Fprintf(w, "%s\n", price)
}

func (db database) update(w http.ResponseWriter, req *http.Request) {
	item := req.URL.Query().Get("item")
	_, ok := db[item]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "no such item: %q", item)
		return
	}
	var price dollars
	fmt.Sscanf(req.URL.Query().Get("price"), "%f", &price)
	db[item] = price
}
