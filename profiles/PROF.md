По данны профилирования heap вижу что помять выделяется на CompressGZIPResponseMW (go-pool PutWriter).

После замены реализации CompressGZIPResponseMW на CompressGZIPBufferResponseMW пофит в -1291.74kB

```go
go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof
File: server
Type: inuse_space
Time: May 19, 2024 at 3:17pm (+04)
Showing nodes accounting for -1291.74kB, 45.66% of 2829.23kB total
      flat  flat%   sum%        cum   cum%
-1805.17kB 63.80% 63.80% -1805.17kB 63.80%  compress/flate.NewWriter (inline)
 1024.18kB 36.20% 27.60%  1024.18kB 36.20%  net/textproto.readMIMEHeader
-1024.06kB 36.20% 63.80% -1024.06kB 36.20%  github.com/go-chi/chi/v5.endpoints.Value (inline)
  513.31kB 18.14% 45.66%   513.31kB 18.14%  compress/flate.NewReader
         0     0% 45.66%   513.31kB 18.14%  compress/gzip.(*Reader).Reset
         0     0% 45.66%   513.31kB 18.14%  compress/gzip.(*Reader).readHeader
         0     0% 45.66% -1805.17kB 63.80%  compress/gzip.(*Writer).Close
         0     0% 45.66% -1805.17kB 63.80%  compress/gzip.(*Writer).Write
         0     0% 45.66%   513.31kB 18.14%  compress/gzip.NewReader
         0     0% 45.66%  -512.03kB 18.10%  github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/handler.AddMetricOperations
         0     0% 45.66%  -512.03kB 18.10%  github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/handler.AddPProfOperations
         0     0% 45.66%  -512.03kB 18.10%  github.com/go-chi/chi/v5.(*Mux).Handle (inline)
         0     0% 45.66%  -512.03kB 18.10%  github.com/go-chi/chi/v5.(*Mux).Mount
         0     0% 45.66%  -512.03kB 18.10%  github.com/go-chi/chi/v5.(*Mux).Route
         0     0% 45.66% -1291.86kB 45.66%  github.com/go-chi/chi/v5.(*Mux).ServeHTTP
         0     0% 45.66% -1024.06kB 36.20%  github.com/go-chi/chi/v5.(*Mux).handle
         0     0% 45.66% -1024.06kB 36.20%  github.com/go-chi/chi/v5.(*node).InsertRoute
         0     0% 45.66% -1024.06kB 36.20%  github.com/go-chi/chi/v5.(*node).setEndpoint
         0     0% 45.66%  -512.03kB 18.10%  github.com/go-chi/chi/v5/middleware.Profiler
         0     0% 45.66% -1805.17kB 63.80%  github.com/ungerik/go-pool.(*GzipPool).PutWriter
         0     0% 45.66%   513.31kB 18.14%  main.createMiddleWareList.NewCompressGZIPBufferResponseMW.func3.1
         0     0% 45.66% -1805.17kB 63.80%  main.createMiddleWareList.NewCompressGZIPResponseMW.func3.1
         0     0% 45.66% -1291.86kB 45.66%  main.createMiddleWareList.NewLoggingResponseMW.func1.1
         0     0% 45.66%   513.31kB 18.14%  main.createMiddleWareList.NewUncompressGZIPRequestMW.func4.1
         0     0% 45.66% -1024.06kB 36.20%  main.main
         0     0% 45.66%  1024.18kB 36.20%  net/http.(*conn).readRequest
         0     0% 45.66%  -267.68kB  9.46%  net/http.(*conn).serve
         0     0% 45.66% -1291.86kB 45.66%  net/http.HandlerFunc.ServeHTTP
         0     0% 45.66%  1024.18kB 36.20%  net/http.readRequest
         0     0% 45.66% -1291.86kB 45.66%  net/http.serverHandler.ServeHTTP
         0     0% 45.66%  1024.18kB 36.20%  net/textproto.(
```
