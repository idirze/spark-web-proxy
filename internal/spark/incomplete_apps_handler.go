/*
 *    Copyright 2025 The OKDP Authors.
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

// Package spark provides handlers and response rewriting logic for Spark UI
// and Spark History endpoints proxied through the application.
package spark

import (
	"bytes"
	"compress/gzip"
	"io"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/okdp/spark-web-proxy/internal/logging"
	"github.com/okdp/spark-web-proxy/internal/spark/proxy"
)

// sparkVersionRe matches the Spark version span in Spark UI/History HTML pages
// and captures the major version number (e.g., "3" from "3.3.1").
var sparkVersionRe = regexp.MustCompile(`(?is)<span[^>]*class=["'][^"']*\bversion\b[^"']*["'][^>]*>\s*([0-9]+)`)

// IncompleteAppsHandler implements proxy.ReverseProxyHandler and injects the
// required scripts into the Spark History "incomplete applications" page.
type IncompleteAppsHandler struct {
}

// NewIncompleteAppsHandler creates a reverse proxy configured to handle
// Spark History incomplete applications pages.
func NewIncompleteAppsHandler(upstreamURL *url.URL, appID string) *proxy.SparkReverseProxy {
	return proxy.NewSparkReverseProxy(IncompleteAppsHandler{}, upstreamURL, appID)
}

// ServeSparkHistoryIncompleteApps proxies Spark History incomplete applications
// requests to the configured upstream.
func ServeSparkHistoryIncompleteApps(c *gin.Context, upstreamURL *url.URL, appID string) {
	NewIncompleteAppsHandler(upstreamURL, appID).
		ServeHTTP(c.Writer, c.Request)
}

// ModifyRequest returns a function that rewrites the incoming request URL to
// target the provided upstream URL.
func (c IncompleteAppsHandler) ModifyRequest(upstreamURL *url.URL) func(*http.Request) {
	return func(req *http.Request) {
		req.URL.Scheme = upstreamURL.Scheme
		req.URL.Host = upstreamURL.Host
		req.Host = upstreamURL.Host
		upstreamURL.RawQuery = req.URL.RawQuery
		upstreamURL.RawFragment = req.URL.RawFragment
		req.URL = upstreamURL
	}
}

// ModifyResponse returns a function that rewrites the Spark History incomplete
// applications page when it contains the "No incomplete applications found!" message.
func (c IncompleteAppsHandler) ModifyResponse() func(*http.Response) error {
	return func(resp *http.Response) error {
		resp.TransferEncoding = []string{"identity"}
		// spark.history.ui.maxApplications = math.MaxInt32
		// https://spark.apache.org/docs/latest/monitoring.html#spark-history-server-configuration-options
		return handleIncompleteApplicationsPage(resp, math.MaxInt32)
	}
}

// handleIncompleteApplicationsPage injects the Spark History page scripts when
// the response is an HTML "no incomplete applications" page.
func handleIncompleteApplicationsPage(resp *http.Response, limit int) error {

	log.Debug("Handle incomplete applications pages")

	ct := strings.ToLower(resp.Header.Get("Content-Type"))
	if !strings.Contains(ct, "text/html") {
		return nil
	}

	plain, raw, isGzip, err := readBodyMaybeGunzip(resp)
	if err != nil {
		log.Warn("Failed to read HTML response body: %v", err)
		return nil
	}

	// Intercept "no incomplete applications" page
	if !bytes.Contains(plain, []byte("No incomplete applications found!")) {
		restoreBody(resp, raw, isGzip)
		return nil
	}

	modified := replaceNoIncompleteBlock(plain, limit)

	if err := writeBody(resp, modified, isGzip); err != nil {
		log.Warn("Failed to write modified HTML response body: %v", err)
		restoreBody(resp, raw, isGzip)
		return nil
	}

	log.Debug("Add Spark historypage scripts into 'incomplete applications' page")
	return nil
}

func replaceNoIncompleteBlock(html []byte, limit int) []byte {
	major, ok := sparkMajorFromHTML(html)

	var repl []byte
	if ok && major >= 4 {
		log.Debug("Spark version parsed successfully (major=%d); using Spark 4+ ES module call", major)
		repl = []byte(
			`<script src="/static/dataTables.rowsGroup.js"></script>` + "\n" +
				`<script type="module" src="/static/historypage.js"></script>` + "\n" +
				`<script type="module">` + "\n" +
				`  import { setAppLimit } from "/static/historypage.js";` + "\n" +
				`  setAppLimit(` + strconv.Itoa(limit) + `);` + "\n" +
				`</script>` + "\n" +
				`<div id="history-summary" class="row-fluid"></div>` + "\n",
		)
	} else {
		log.Debug("Spark version parsed successfully (major=%d); using Spark 3+ classic js call", major)
		repl = []byte(
			`<script src="/static/dataTables.rowsGroup.js"></script>` + "\n" +
				`<div id="history-summary" class="row-fluid"></div>` + "\n" +
				`<script src="/static/historypage.js"></script>` + "\n" +
				`<script>setAppLimit(` + strconv.Itoa(limit) + `)</script>` + "\n",
		)
	}

	needle := []byte("<h4>No incomplete applications found!</h4>")
	if bytes.Contains(html, needle) {
		return bytes.Replace(html, needle, repl, 1)
	}

	// fallback: replace plain text
	return bytes.Replace(html, []byte("No incomplete applications found!"), repl, 1)
}

func readBodyMaybeGunzip(resp *http.Response) (plain []byte, raw []byte, isGzip bool, err error) {
	isGzip = strings.Contains(strings.ToLower(resp.Header.Get("Content-Encoding")), "gzip")

	raw, err = io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return nil, nil, isGzip, err
	}

	if !isGzip {
		return raw, raw, false, nil
	}

	gr, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		return nil, nil, true, err
	}

	defer func() {
		if err := gr.Close(); err != nil {
			log.Warn("failed to close gzip reader: %v", err)
		}
	}()

	plain, err = io.ReadAll(gr)
	if err != nil {
		return nil, nil, true, err
	}
	return plain, raw, true, nil
}

func restoreBody(resp *http.Response, raw []byte, isGzip bool) {
	resp.Body = io.NopCloser(bytes.NewReader(raw))
	resp.ContentLength = int64(len(raw))
	resp.Header.Set("Content-Length", strconv.Itoa(len(raw)))
	if isGzip {
		resp.Header.Set("Content-Encoding", "gzip")
	} else {
		resp.Header.Del("Content-Encoding")
	}
}

func writeBody(resp *http.Response, plain []byte, gzipIt bool) error {
	if !gzipIt {
		resp.Body = io.NopCloser(bytes.NewReader(plain))
		resp.ContentLength = int64(len(plain))
		resp.Header.Set("Content-Length", strconv.Itoa(len(plain)))
		resp.Header.Del("Content-Encoding")
		return nil
	}

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write(plain); err != nil {
		_ = gw.Close()
		return err
	}
	if err := gw.Close(); err != nil {
		return err
	}

	b := buf.Bytes()
	resp.Body = io.NopCloser(bytes.NewReader(b))
	resp.ContentLength = int64(len(b))
	resp.Header.Set("Content-Length", strconv.Itoa(len(b)))
	resp.Header.Set("Content-Encoding", "gzip")
	return nil
}

/*
sparkMajorFromHTML extracts the Apache Spark *major version* from
the Spark UI HTML page.

It looks for a <span> element whose class attribute contains the token
"version", for example:

	<span class="version">3.3.1</span>
	<span class="version" style="margin-right: 15px;">4.0.0</span>
	<span class="foo version bar">4.1.1</span>

Only the *major* version (the number before the first dot) is returned.

Returns:

	(major, true)  if a version span is found and parsed successfully
	(0, false)     if the version cannot be found or parsed

This function intentionally uses a lightweight regex instead of a full
HTML parser for performance and robustness in a reverse-proxy context,
where Spark’s HTML structure is stable and well-known.
*/
func sparkMajorFromHTML(b []byte) (int, bool) {
	m := sparkVersionRe.FindSubmatch(b)
	if len(m) < 2 {
		return 0, false
	}
	major, err := strconv.Atoi(string(m[1]))
	if err != nil {
		return 0, false
	}
	return major, true
}
