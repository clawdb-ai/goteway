package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type config struct {
	Host       string
	Port       int
	Upstream   *url.URL
	LogFile    string
	LogHeaders bool
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	logger, closeLogFile, err := newLogger(cfg.LogFile)
	if err != nil {
		log.Fatalf("init logger: %v", err)
	}
	defer closeLogFile()

	proxy := httputil.NewSingleHostReverseProxy(cfg.Upstream)
	proxy.ErrorLog = log.New(logger.Writer(), "proxy-error: ", log.LstdFlags|log.Lmicroseconds)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		rid := requestID(r)
		logger.Printf("rid=%s upstream_error method=%s path=%s err=%q", rid, r.Method, r.URL.RequestURI(), err.Error())
		http.Error(w, "bad gateway", http.StatusBadGateway)
	}

	handler := loggingMiddleware(logger, proxy, cfg.LogHeaders)

	srv := &http.Server{
		Addr:              net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)),
		Handler:           handler,
		ReadHeaderTimeout: 15 * time.Second,
		IdleTimeout:       120 * time.Second,
		ErrorLog:          log.New(logger.Writer(), "http-server: ", log.LstdFlags|log.Lmicroseconds),
	}

	logger.Printf("openclaw-go-proxy start listen=%s upstream=%s log_headers=%t", srv.Addr, cfg.Upstream.String(), cfg.LogHeaders)

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
		logger.Printf("openclaw-go-proxy stop reason=%s", ctx.Err())
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("server failed: %v", err)
		}
	}
}

func loadConfig() (config, error) {
	host := env("GATEWAY_HOST", "127.0.0.1")
	if host == "" {
		return config{}, fmt.Errorf("GATEWAY_HOST must not be empty")
	}
	port, err := envInt("GATEWAY_PORT", 18789)
	if err != nil {
		return config{}, err
	}
	if port <= 0 || port > 65535 {
		return config{}, fmt.Errorf("GATEWAY_PORT must be between 1 and 65535")
	}

	upstream := env("OPENCLAW_UPSTREAM_URL", "http://127.0.0.1:18790")
	upstreamURL, err := url.Parse(upstream)
	if err != nil || upstreamURL.Scheme == "" || upstreamURL.Host == "" {
		return config{}, fmt.Errorf("OPENCLAW_UPSTREAM_URL must be a valid absolute URL")
	}

	logHeaders, err := envBool("OPENCLAW_PROXY_LOG_HEADERS", false)
	if err != nil {
		return config{}, err
	}

	return config{
		Host:       host,
		Port:       port,
		Upstream:   upstreamURL,
		LogFile:    os.Getenv("OPENCLAW_PROXY_LOG_FILE"),
		LogHeaders: logHeaders,
	}, nil
}

func newLogger(logFile string) (*log.Logger, func(), error) {
	writer := io.Writer(os.Stdout)
	closeFn := func() {}

	if logFile != "" {
		if err := os.MkdirAll(filepathDir(logFile), 0o755); err != nil {
			return nil, nil, fmt.Errorf("create log dir: %w", err)
		}
		f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return nil, nil, fmt.Errorf("open log file: %w", err)
		}
		writer = io.MultiWriter(os.Stdout, f)
		closeFn = func() { _ = f.Close() }
	}

	return log.New(writer, "", log.LstdFlags|log.Lmicroseconds), closeFn, nil
}

func loggingMiddleware(logger *log.Logger, next http.Handler, logHeaders bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := requestID(r)
		rec := &statusRecorder{ResponseWriter: w}
		start := time.Now()
		next.ServeHTTP(rec, r)
		d := time.Since(start)

		status := rec.status
		if status == 0 {
			// Upgraded HTTP connections (WebSocket) may bypass WriteHeader.
			status = http.StatusSwitchingProtocols
		}

		if logHeaders {
			logger.Printf(
				"rid=%s remote=%s method=%s path=%s status=%d bytes=%d duration_ms=%d ua=%q headers=%q",
				rid,
				clientIP(r.RemoteAddr),
				r.Method,
				r.URL.RequestURI(),
				status,
				rec.bytes,
				d.Milliseconds(),
				r.UserAgent(),
				flattenHeaders(r.Header),
			)
			return
		}

		logger.Printf(
			"rid=%s remote=%s method=%s path=%s status=%d bytes=%d duration_ms=%d ua=%q",
			rid,
			clientIP(r.RemoteAddr),
			r.Method,
			r.URL.RequestURI(),
			status,
			rec.bytes,
			d.Milliseconds(),
			r.UserAgent(),
		)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int64
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	n, err := r.ResponseWriter.Write(b)
	r.bytes += int64(n)
	return n, err
}

func (r *statusRecorder) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (r *statusRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("response writer does not support hijacking")
	}
	if r.status == 0 {
		r.status = http.StatusSwitchingProtocols
	}
	return h.Hijack()
}

func (r *statusRecorder) Push(target string, opts *http.PushOptions) error {
	p, ok := r.ResponseWriter.(http.Pusher)
	if !ok {
		return http.ErrNotSupported
	}
	return p.Push(target, opts)
}

func (r *statusRecorder) ReadFrom(src io.Reader) (int64, error) {
	if rf, ok := r.ResponseWriter.(io.ReaderFrom); ok {
		if r.status == 0 {
			r.status = http.StatusOK
		}
		n, err := rf.ReadFrom(src)
		r.bytes += n
		return n, err
	}
	return io.Copy(r, src)
}

func requestID(r *http.Request) string {
	if v := strings.TrimSpace(r.Header.Get("X-Request-Id")); v != "" {
		return v
	}
	if v := strings.TrimSpace(r.Header.Get("X-Correlation-Id")); v != "" {
		return v
	}
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	return hex.EncodeToString(b[:])
}

func flattenHeaders(h http.Header) string {
	pairs := make([]string, 0, len(h))
	for k, v := range h {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, strings.Join(v, ",")))
	}
	return strings.Join(pairs, ";")
}

func clientIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}

func env(k, def string) string {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return def
	}
	return v
}

func envInt(k string, def int) (int, error) {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return def, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid integer", k)
	}
	return n, nil
}

func envBool(k string, def bool) (bool, error) {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return def, nil
	}
	switch strings.ToLower(v) {
	case "1", "true", "yes", "on":
		return true, nil
	case "0", "false", "no", "off":
		return false, nil
	default:
		return false, fmt.Errorf("%s must be one of true/false/1/0/yes/no/on/off", k)
	}
}

func filepathDir(path string) string {
	i := strings.LastIndex(path, "/")
	if i <= 0 {
		return "."
	}
	return path[:i]
}
