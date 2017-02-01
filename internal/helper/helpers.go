package helper

import (
	"errors"
	"log"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

const NginxResponseBufferHeader = "X-Accel-Buffering"

func Fail500(w http.ResponseWriter, r *http.Request, err error) {
	http.Error(w, "Internal server error", 500)
	captureRavenError(r, err)
	printError(r, err)
}

func LogError(r *http.Request, err error) {
	captureRavenError(r, err)
	printError(r, err)
}

func ServiceUnavailable(w http.ResponseWriter, r *http.Request, err error) {
	http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
	captureRavenError(r, err)
	printError(r, err)
}

func TooManyRequests(w http.ResponseWriter, r *http.Request, err error) {
	http.Error(w, "Too Many Requests", 429) // http.StatusTooManyRequests was added in go1.6
	captureRavenError(r, err)
	printError(r, err)
}

func printError(r *http.Request, err error) {
	if r != nil {
		log.Printf("error: %s %q: %v", r.Method, r.RequestURI, err)
	} else {
		log.Printf("error: %v", err)
	}
}

func SetNoCacheHeaders(header http.Header) {
	header.Set("Cache-Control", "no-cache, no-store, max-age=0, must-revalidate")
	header.Set("Pragma", "no-cache")
	header.Set("Expires", "Fri, 01 Jan 1990 00:00:00 GMT")
}

func OpenFile(path string) (file *os.File, fi os.FileInfo, err error) {
	file, err = os.Open(path)
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			file.Close()
		}
	}()

	fi, err = file.Stat()
	if err != nil {
		return
	}

	// The os.Open can also open directories
	if fi.IsDir() {
		err = &os.PathError{
			Op:   "open",
			Path: path,
			Err:  errors.New("path is directory"),
		}
		return
	}

	return
}

func URLMustParse(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		log.Fatalf("urlMustParse: %q %v", s, err)
	}
	return u
}

func HTTPError(w http.ResponseWriter, r *http.Request, error string, code int) {
	if r.ProtoAtLeast(1, 1) {
		// Force client to disconnect if we render request error
		w.Header().Set("Connection", "close")
	}

	http.Error(w, error, code)
}

func HeaderClone(h http.Header) http.Header {
	h2 := make(http.Header, len(h))
	for k, vv := range h {
		vv2 := make([]string, len(vv))
		copy(vv2, vv)
		h2[k] = vv2
	}
	return h2
}

func CleanUpProcessGroup(cmd *exec.Cmd) {
	if cmd == nil {
		return
	}

	process := cmd.Process
	if process != nil && process.Pid > 0 {
		// Send SIGTERM to the process group of cmd
		syscall.Kill(-process.Pid, syscall.SIGTERM)
	}

	// reap our child process
	cmd.Wait()
}

func ExitStatus(err error) (int, bool) {
	exitError, ok := err.(*exec.ExitError)
	if !ok {
		return 0, false
	}

	waitStatus, ok := exitError.Sys().(syscall.WaitStatus)
	if !ok {
		return 0, false
	}

	return waitStatus.ExitStatus(), true
}

func DisableResponseBuffering(w http.ResponseWriter) {
	w.Header().Set(NginxResponseBufferHeader, "no")
}

func AllowResponseBuffering(w http.ResponseWriter) {
	w.Header().Del(NginxResponseBufferHeader)
}

func SetForwardedFor(newHeaders *http.Header, originalRequest *http.Request) {
	if clientIP, _, err := net.SplitHostPort(originalRequest.RemoteAddr); err == nil {
		var header string

		// If we aren't the first proxy retain prior
		// X-Forwarded-For information as a comma+space
		// separated list and fold multiple headers into one.
		if prior, ok := originalRequest.Header["X-Forwarded-For"]; ok {
			header = strings.Join(prior, ", ") + ", " + clientIP
		} else {
			header = clientIP
		}
		newHeaders.Set("X-Forwarded-For", header)
	}
}

func IsContentType(expected, actual string) bool {
	parsed, _, err := mime.ParseMediaType(actual)
	return err == nil && parsed == expected
}
