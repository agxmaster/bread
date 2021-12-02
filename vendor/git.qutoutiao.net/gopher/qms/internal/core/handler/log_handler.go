package handler

import (
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"git.qutoutiao.net/gopher/qms/internal/config"
	"git.qutoutiao.net/gopher/qms/internal/core/invocation"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/httputil"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/iputil"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/protocol"
	"git.qutoutiao.net/gopher/qms/pkg/errors"
	"git.qutoutiao.net/gopher/qms/pkg/qlog"
	"github.com/golib/zerolog/diode"
	"google.golang.org/grpc/codes"
)

// LogProviderHandler tracing provider handler
type LogProviderHandler struct {
	logger qlog.Logger
}

// Handle is to handle the provider tracing related things
func (t *LogProviderHandler) Handle(chain *Chain, i *invocation.Invocation, cb invocation.ResponseCallBack) {
	if !config.Get().AccessLog.Enabled {
		chain.Next(i, cb)
		return
	}

	l, err := newLogParams(i)
	if err != nil {
		chain.Next(i, cb)
		return
	}

	chain.Next(i, func(r *invocation.Response) (err error) {
		err = cb(r)
		l.format(t.logger, r.RequestID, r.Status, r.Err)
		return
	})
}

// Name returns tracing-provider string
func (t *LogProviderHandler) Name() string {
	return LogProvider
}

func newLogProviderHandler() Handler {
	logHandler := &LogProviderHandler{}

	accessLog := config.Get().AccessLog
	if accessLog.Enabled {
		output := io.Writer(os.Stdout)
		if accessLog.FileName != "" && filepath.IsAbs(accessLog.FileName) {
			if err := createLogDir(accessLog.FileName); err != nil {
				qlog.WithError(err).Errorf("create accesslog file failed")
				return logHandler
			}

			file, err := os.OpenFile(accessLog.FileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
			if err != nil {
				qlog.WithError(err).Errorf("open accesslog file failed")
				return logHandler
			}
			output = file
		}
		if accessLog.AsyncEnabled {
			output = diode.NewWriter(output, 10000, 10*time.Millisecond, func(missed int) {
				output.Write([]byte("[WARN] Logger dropped " + strconv.Itoa(missed) + " messages."))
			})
		}

		logHandler.logger = qlog.NewWithOption(&qlog.Option{
			Output:    output,
			Level:     qlog.InfoLevel,
			Formatter: qlog.JsonFormatter,
		})
	}

	return logHandler
}

func createLogDir(out string) error {
	_, err := os.Stat(filepath.Dir(out))
	if err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Dir(out), os.ModePerm)
		if err != nil {
			return errors.WithStack(err)
		}
	} else if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func init() {
	RegisterHandler(LogProvider, newLogProviderHandler)
}

type logParams struct {
	protocol protocol.Protocol
	start    time.Time
	fields   qlog.Fields
}

func newLogParams(i *invocation.Invocation) (*logParams, error) {
	l := &logParams{
		protocol: i.Protocol,
		start:    time.Now(),
		fields:   make(qlog.Fields, 10),
	}

	switch i.Protocol {
	case protocol.ProtocHTTP:
		request, err := httputil.HTTPRequest(i)
		if err != nil {
			qlog.Error("extract request from invocation failed")
			return nil, err
		}

		path := request.URL.Path
		raw := request.URL.RawQuery
		if raw != "" {
			l.fields["query"] = raw
		}
		l.fields["component"] = "net/http"
		l.fields["clientIP"] = iputil.ClientIP(request)
		l.fields["path"] = path
		l.fields["method"] = request.Method
	case protocol.ProtocGrpc:
		l.fields["component"] = "grpc"
		l.fields["service"] = path.Dir(i.SchemaID)[1:]
		l.fields["method"] = path.Base(i.SchemaID)
	}
	return l, nil
}

func (l *logParams) format(logger qlog.Logger, requestid string, code int, err error) {
	if logger == nil {
		qlog.Errorf("qlog.Logger is nil")
		return
	}
	l.fields["request_id"] = requestid
	l.fields["duration_ms"] = time.Since(l.start).Milliseconds()
	var level qlog.Level

	switch l.protocol {
	case protocol.ProtocHTTP:
		l.fields["code"] = code
		level = l.httpCode2Level(code)
	case protocol.ProtocGrpc:
		code := codes.Code(code)
		l.fields["code"] = code.String()
		level = l.grpcCode2Level(code)
	}
	if err != nil {
		level = qlog.ErrorLevel
		l.fields["error"] = err
	}
	logger.WithFields(l.fields).Logf(level, "finished call with code %v", code)
}

// StatusCodeColor is the ANSI color for appropriately logging http status code to a terminal.
func (l *logParams) httpCode2Level(code int) qlog.Level {
	switch {
	case code >= http.StatusOK && code < http.StatusMultipleChoices:
		return qlog.InfoLevel
	case code >= http.StatusMultipleChoices && code < http.StatusBadRequest:
		return qlog.WarnLevel
	case code >= http.StatusBadRequest && code < http.StatusInternalServerError:
		return qlog.WarnLevel
	default:
		return qlog.ErrorLevel
	}
}

// code2Level is the default implementation of gRPC return codes to log levels for server side.
func (l *logParams) grpcCode2Level(code codes.Code) qlog.Level {
	switch code {
	case codes.OK, codes.Canceled, codes.InvalidArgument, codes.NotFound, codes.AlreadyExists, codes.Unauthenticated:
		return qlog.InfoLevel
	case codes.DeadlineExceeded, codes.PermissionDenied, codes.ResourceExhausted, codes.FailedPrecondition, codes.Aborted, codes.OutOfRange, codes.Unavailable:
		return qlog.WarnLevel
	default:
		return qlog.ErrorLevel
	}
}
