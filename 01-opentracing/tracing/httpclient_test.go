package tracing

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
)

func TestHTTPClient(t *testing.T) {
	tracer, err := NewTracer("test")

	client := &http.Client{Transport: &nethttp.Transport{}}
	req, err := http.NewRequest("GET", "http://qq.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	// req = req.WithContext(ctx) // extend existing trace, if any

	req, ht := nethttp.TraceRequest(tracer, req)
	defer ht.Finish()

	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	respBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(respBody))
}
