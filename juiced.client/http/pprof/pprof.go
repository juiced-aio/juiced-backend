// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package pprof serves via its HTTP server runtime profiling data
// in the format expected by the pprof visualization tool.
//
// The package is typically only imported for the side effect of
// registering its HTTP handlers.
// The handled paths all begin with /debug/pprof/.
//
// To use pprof, link this package into your program:
//	import _ "backend.juicedbot.io/juiced.client/http/pprof"
//
// If your application is not already running an http server, you
// need to start one. Add "backend.juicedbot.io/juiced.client/http" and "log" to your imports and
// the following code to your main function:
//
// 	go func() {
// 		log.Println(http.ListenAndServe("localhost:6060", nil))
// 	}()
//
// If you are not using DefaultServeMux, you will have to register handlers
// with the mux you are using.
//
// Then use the pprof tool to look at the heap profile:
//
//	go tool pprof http://localhost:6060/debug/pprof/heap
//
// Or to look at a 30-second CPU profile:
//
//	go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
//
// Or to look at the goroutine blocking profile, after calling
// runtime.SetBlockProfileRate in your program:
//
//	go tool pprof http://localhost:6060/debug/pprof/block
//
// Or to look at the holders of contended mutexes, after calling
// runtime.SetMutexProfileFraction in your program:
//
//	go tool pprof http://localhost:6060/debug/pprof/mutex
//
// The package also exports a handler that serves execution trace data
// for the "go tool trace" command. To collect a 5-second execution trace:
//
//	wget -O trace.out http://localhost:6060/debug/pprof/trace?seconds=5
//	go tool trace trace.out
//
// To view all available profiles, open http://localhost:6060/debug/pprof/
// in your browser.
//
// For a study of the facility in action, visit
//
//	https://blog.golang.org/2011/06/profiling-go-programs.html
//
package pprof

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"html"
	"io"
	"log"
	"net/url"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"sort"
	"strconv"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.client/http"

	"github.com/google/pprof/profile"
)

func init() {
	http.HandleFunc("/debug/pprof/", Index)
	http.HandleFunc("/debug/pprof/cmdline", Cmdline)
	http.HandleFunc("/debug/pprof/profile", Profile)
	http.HandleFunc("/debug/pprof/visualprofile", VisualProfile)
	http.HandleFunc("/debug/pprof/visualheap", VisualHeap)
	http.HandleFunc("/debug/pprof/symbol", Symbol)
	http.HandleFunc("/debug/pprof/trace", Trace)
}

// Cmdline responds with the running program's
// command line, with arguments separated by NUL bytes.
// The package initialization registers it as /debug/pprof/cmdline.
func Cmdline(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, strings.Join(os.Args, "\x00"))
}

func sleep(r *http.Request, d time.Duration) {
	select {
	case <-time.After(d):
	case <-r.Context().Done():
	}
}

func durationExceedsWriteTimeout(r *http.Request, seconds float64) bool {
	srv, ok := r.Context().Value(http.ServerContextKey).(*http.Server)
	return ok && srv.WriteTimeout != 0 && seconds >= srv.WriteTimeout.Seconds()
}

func serveError(w http.ResponseWriter, status int, txt string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Go-Pprof", "1")
	w.Header().Del("Content-Disposition")
	w.WriteHeader(status)
	fmt.Fprintln(w, txt)
}

type Sample struct {
	Percent     float64
	Usage       int64
	Functions   []profile.Function
	FunctionIDs []uint64
	Line        int64
}

func IsInside(x uint64, ySlice []uint64) bool {
	for _, y := range ySlice {
		if y == x {
			return true
		}
	}
	return false
}

// VisualProfile responds with the pprof-formatted cpu profile.
// Profiling lasts for duration specified in seconds GET parameter, or for 30 seconds if not specified.
// The package initialization registers it as /debug/pprof/visualprofile.
func VisualProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	sec, err := strconv.ParseInt(r.FormValue("seconds"), 10, 64)
	if sec <= 0 || err != nil {
		sec = 30
	}

	if durationExceedsWriteTimeout(r, float64(sec)) {
		serveError(w, http.StatusBadRequest, "profile duration exceeds server's WriteTimeout")
		return
	}

	// Set Content Type assuming StartCPUProfile will work,
	// because if it does it starts writing.
	buf := new(bytes.Buffer)
	if err := pprof.StartCPUProfile(buf); err != nil {
		// StartCPUProfile failed, so no writes yet.
		serveError(w, http.StatusInternalServerError,
			fmt.Sprintf("Could not enable CPU profiling: %s", err))
		return
	}

	sleep(r, time.Duration(sec)*time.Second)
	pprof.StopCPUProfile()

	parsed, err := profile.ParseData(buf.Bytes())
	if err != nil {
		serveError(w, http.StatusInternalServerError, "Could not parse cpu profile")
	}

	var cpuTotal int64
	var samples []*Sample
	for _, sample := range parsed.Sample {
		tempSample := Sample{Usage: sample.Value[1]}
		cpuTotal += sample.Value[1]
		for _, location := range sample.Location {
			tempSample.FunctionIDs = append(tempSample.FunctionIDs, location.ID)
			tempSample.Line = location.Line[0].Line
		}

		samples = append(samples, &tempSample)
	}

	for _, function := range parsed.Function {
		for _, sample := range samples {
			if IsInside(function.ID, sample.FunctionIDs) {
				sample.Functions = append(sample.Functions, *function)
			}
		}
	}

	sort.SliceStable(samples, func(i, j int) bool {
		return samples[i].Usage > samples[j].Usage
	})

	w.Write([]byte("<h1>Cpu Usage:</h1>"))
	for _, sample := range samples {
		sample.Percent = float64(sample.Usage) / float64(cpuTotal)
		if len(sample.Functions) > 0 {
			w.Write([]byte(fmt.Sprintf("<div>%%%v&nbsp;&nbsp;|&nbsp;%v&nbsp;%v:%v</div>", sample.Percent*100, sample.Functions[0].Name, sample.Functions[0].Filename, sample.Line)))
			w.Write([]byte("<div>&nbsp;</div>"))
		}
	}

}

// VisualHeap responds with the pprof-formatted heap profile.
// The package initialization registers it as /debug/pprof/visualheap.
func VisualHeap(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	p := pprof.Lookup("heap")
	if p == nil {
		serveError(w, http.StatusNotFound, "Unknown profile")
		return
	}

	w.Header().Set("Content-Type", "text/html")
	var buf bytes.Buffer

	p.WriteTo(&buf, 0)
	parsed, err := profile.ParseData(buf.Bytes())
	if err != nil {
		serveError(w, http.StatusInternalServerError, "Could not parse heap profile")
	}
	var memoryTotal int64
	var samples []*Sample
	for _, sample := range parsed.Sample {
		tempSample := Sample{Usage: sample.Value[1]}
		memoryTotal += sample.Value[1]
		for _, location := range sample.Location {
			tempSample.FunctionIDs = append(tempSample.FunctionIDs, location.ID)

		}

		samples = append(samples, &tempSample)
	}

	for _, function := range parsed.Function {
		for _, sample := range samples {
			if IsInside(function.ID, sample.FunctionIDs) {
				sample.Functions = append(sample.Functions, *function)
			}
		}
	}

	sort.SliceStable(samples, func(i, j int) bool {
		return samples[i].Usage > samples[j].Usage
	})

	w.Write([]byte("<h1>Memory Usage:</h1>"))
	for _, sample := range samples {
		sample.Percent = float64(sample.Usage) / float64(memoryTotal)
		if len(sample.Functions) > 0 {
			w.Write([]byte(fmt.Sprintf("<div>%%%v&nbsp;&nbsp;|&nbsp;%v&nbsp;%v</div>", sample.Percent*100, sample.Functions[0].Name, sample.Functions[0].Filename)))
			w.Write([]byte("<div>&nbsp;</div>"))
		}
	}

}

// Profile responds with the pprof-formatted cpu profile.
// Profiling lasts for duration specified in seconds GET parameter, or for 30 seconds if not specified.
// The package initialization registers it as /debug/pprof/profile.
func Profile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	sec, err := strconv.ParseInt(r.FormValue("seconds"), 10, 64)
	if sec <= 0 || err != nil {
		sec = 30
	}

	if durationExceedsWriteTimeout(r, float64(sec)) {
		serveError(w, http.StatusBadRequest, "profile duration exceeds server's WriteTimeout")
		return
	}

	// Set Content Type assuming StartCPUProfile will work,
	// because if it does it starts writing.
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", `attachment; filename="profile"`)
	if err := pprof.StartCPUProfile(w); err != nil {
		// StartCPUProfile failed, so no writes yet.
		serveError(w, http.StatusInternalServerError,
			fmt.Sprintf("Could not enable CPU profiling: %s", err))
		return
	}
	sleep(r, time.Duration(sec)*time.Second)
	pprof.StopCPUProfile()
}

// Trace responds with the execution trace in binary form.
// Tracing lasts for duration specified in seconds GET parameter, or for 1 second if not specified.
// The package initialization registers it as /debug/pprof/trace.
func Trace(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	sec, err := strconv.ParseFloat(r.FormValue("seconds"), 64)
	if sec <= 0 || err != nil {
		sec = 1
	}

	if durationExceedsWriteTimeout(r, sec) {
		serveError(w, http.StatusBadRequest, "profile duration exceeds server's WriteTimeout")
		return
	}

	// Set Content Type assuming trace.Start will work,
	// because if it does it starts writing.
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", `attachment; filename="trace"`)
	if err := trace.Start(w); err != nil {
		// trace.Start failed, so no writes yet.
		serveError(w, http.StatusInternalServerError,
			fmt.Sprintf("Could not enable tracing: %s", err))
		return
	}
	sleep(r, time.Duration(sec*float64(time.Second)))
	trace.Stop()
}

// Symbol looks up the program counters listed in the request,
// responding with a table mapping program counters to function names.
// The package initialization registers it as /debug/pprof/symbol.
func Symbol(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	// We have to read the whole POST body before
	// writing any output. Buffer the output here.
	var buf bytes.Buffer

	// We don't know how many symbols we have, but we
	// do have symbol information. Pprof only cares whether
	// this number is 0 (no symbols available) or > 0.
	fmt.Fprintf(&buf, "num_symbols: 1\n")

	var b *bufio.Reader
	if r.Method == "POST" {
		b = bufio.NewReader(r.Body)
	} else {
		b = bufio.NewReader(strings.NewReader(r.URL.RawQuery))
	}

	for {
		word, err := b.ReadSlice('+')
		if err == nil {
			word = word[0 : len(word)-1] // trim +
		}
		pc, _ := strconv.ParseUint(string(word), 0, 64)
		if pc != 0 {
			f := runtime.FuncForPC(uintptr(pc))
			if f != nil {
				fmt.Fprintf(&buf, "%#x %s\n", pc, f.Name())
			}
		}

		// Wait until here to check for err; the last
		// symbol will have an err because it doesn't end in +.
		if err != nil {
			if err != io.EOF {
				fmt.Fprintf(&buf, "reading request: %v\n", err)
			}
			break
		}
	}

	w.Write(buf.Bytes())
}

// Handler returns an HTTP handler that serves the named profile.
func Handler(name string) http.Handler {
	return handler(name)
}

type handler string

func (name handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	p := pprof.Lookup(string(name))
	if p == nil {
		serveError(w, http.StatusNotFound, "Unknown profile")
		return
	}
	if sec := r.FormValue("seconds"); sec != "" {
		name.serveDeltaProfile(w, r, p, sec)
		return
	}
	gc, _ := strconv.Atoi(r.FormValue("gc"))
	heap := handler("heap")
	if name == heap && gc > 0 {
		runtime.GC()
	}
	debug, _ := strconv.Atoi(r.FormValue("debug"))
	if debug != 0 {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, name))
	}
	p.WriteTo(w, debug)
}

func (name handler) serveDeltaProfile(w http.ResponseWriter, r *http.Request, p *pprof.Profile, secStr string) {
	sec, err := strconv.ParseInt(secStr, 10, 64)
	if err != nil || sec <= 0 {
		serveError(w, http.StatusBadRequest, `invalid value for "seconds" - must be a positive integer`)
		return
	}
	if !profileSupportsDelta[name] {
		serveError(w, http.StatusBadRequest, `"seconds" parameter is not supported for this profile type`)
		return
	}
	// 'name' should be a key in profileSupportsDelta.
	if durationExceedsWriteTimeout(r, float64(sec)) {
		serveError(w, http.StatusBadRequest, "profile duration exceeds server's WriteTimeout")
		return
	}
	debug, _ := strconv.Atoi(r.FormValue("debug"))
	if debug != 0 {
		serveError(w, http.StatusBadRequest, "seconds and debug params are incompatible")
		return
	}
	p0, err := collectProfile(p)
	if err != nil {
		serveError(w, http.StatusInternalServerError, "failed to collect profile")
		return
	}

	t := time.NewTimer(time.Duration(sec) * time.Second)
	defer t.Stop()

	select {
	case <-r.Context().Done():
		err := r.Context().Err()
		if err == context.DeadlineExceeded {
			serveError(w, http.StatusRequestTimeout, err.Error())
		} else { // TODO: what's a good status code for cancelled requests? 400?
			serveError(w, http.StatusInternalServerError, err.Error())
		}
		return
	case <-t.C:
	}

	p1, err := collectProfile(p)
	if err != nil {
		serveError(w, http.StatusInternalServerError, "failed to collect profile")
		return
	}
	ts := p1.TimeNanos
	dur := p1.TimeNanos - p0.TimeNanos

	p0.Scale(-1)

	p1, err = profile.Merge([]*profile.Profile{p0, p1})
	if err != nil {
		serveError(w, http.StatusInternalServerError, "failed to compute delta")
		return
	}

	p1.TimeNanos = ts // set since we don't know what profile.Merge set for TimeNanos.
	p1.DurationNanos = dur

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s-delta"`, name))
	p1.Write(w)
}

func collectProfile(p *pprof.Profile) (*profile.Profile, error) {
	var buf bytes.Buffer
	if err := p.WriteTo(&buf, 0); err != nil {
		return nil, err
	}
	ts := time.Now().UnixNano()
	p0, err := profile.Parse(&buf)
	if err != nil {
		return nil, err
	}
	p0.TimeNanos = ts
	return p0, nil
}

var profileSupportsDelta = map[handler]bool{
	handler("allocs"):       true,
	handler("block"):        true,
	handler("goroutine"):    true,
	handler("heap"):         true,
	handler("mutex"):        true,
	handler("threadcreate"): true,
}

var profileDescriptions = map[string]string{
	"allocs":       "A sampling of all past memory allocations",
	"block":        "Stack traces that led to blocking on synchronization primitives",
	"cmdline":      "The command line invocation of the current program",
	"goroutine":    "Stack traces of all current goroutines",
	"heap":         "A sampling of memory allocations of live objects. You can specify the gc GET parameter to run GC before taking the heap sample.",
	"mutex":        "Stack traces of holders of contended mutexes",
	"profile":      "CPU profile. You can specify the duration in the seconds GET parameter. After you get the profile file, use the go tool pprof command to investigate the profile.",
	"threadcreate": "Stack traces that led to the creation of new OS threads",
	"trace":        "A trace of execution of the current program. You can specify the duration in the seconds GET parameter. After you get the trace file, use the go tool trace command to investigate the trace.",
}

type profileEntry struct {
	Name  string
	Href  string
	Desc  string
	Count int
}

// Index responds with the pprof-formatted profile named by the request.
// For example, "/debug/pprof/heap" serves the "heap" profile.
// Index responds to a request for "/debug/pprof/" with an HTML page
// listing the available profiles.
func Index(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/debug/pprof/") {
		name := strings.TrimPrefix(r.URL.Path, "/debug/pprof/")
		if name != "" {
			handler(name).ServeHTTP(w, r)
			return
		}
	}

	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var profiles []profileEntry
	for _, p := range pprof.Profiles() {
		profiles = append(profiles, profileEntry{
			Name:  p.Name(),
			Href:  p.Name(),
			Desc:  profileDescriptions[p.Name()],
			Count: p.Count(),
		})
	}

	// Adding other profiles exposed from within this package
	for _, p := range []string{"cmdline", "profile", "trace"} {
		profiles = append(profiles, profileEntry{
			Name: p,
			Href: p,
			Desc: profileDescriptions[p],
		})
	}

	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].Name < profiles[j].Name
	})

	if err := indexTmplExecute(w, profiles); err != nil {
		log.Print(err)
	}
}

func indexTmplExecute(w io.Writer, profiles []profileEntry) error {
	var b bytes.Buffer
	b.WriteString(`<html>
<head>
<title>/debug/pprof/</title>
<style>
.profile-name{
	display:inline-block;
	width:6rem;
}
</style>
</head>
<body>
/debug/pprof/<br>
<br>
Types of profiles available:
<table>
<thead><td>Count</td><td>Profile</td></thead>
`)

	for _, profile := range profiles {
		link := &url.URL{Path: profile.Href, RawQuery: "debug=1"}
		fmt.Fprintf(&b, "<tr><td>%d</td><td><a href='%s'>%s</a></td></tr>\n", profile.Count, link, html.EscapeString(profile.Name))
	}

	b.WriteString(`</table>
<a href="goroutine?debug=2">full goroutine stack dump</a>
<br/>
<p>
Profile Descriptions:
<ul>
`)
	for _, profile := range profiles {
		fmt.Fprintf(&b, "<li><div class=profile-name>%s: </div> %s</li>\n", html.EscapeString(profile.Name), html.EscapeString(profile.Desc))
	}
	b.WriteString(`</ul>
</p>
</body>
</html>`)

	_, err := w.Write(b.Bytes())
	return err
}
