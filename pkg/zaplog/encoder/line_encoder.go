package encoder

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"math"
	"sync"
	"time"
	"unicode/utf8"

	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

const (
	_hex             = "0123456789abcdef"
	_log_field_split = "|"
)

var _ zapcore.Encoder = (*LineEncoder)(nil)

var (
	_linePool = sync.Pool{
		New: func() interface{} {
			return &LineEncoder{}
		},
	}
	_bufferPool      = buffer.NewPool()
	nullLiteralBytes = []byte("null")
)

// LineEncoder is a custom encoder that produces line-format logs
// Format: [datetime]|[level]|[service]|[trace-id]|[span-id]|[parent-span-id]|[file:line]|[function]|message
type LineEncoder struct {
	*zapcore.EncoderConfig
	buf    *buffer.Buffer
	spaced bool // include spaces after colons and commas

	// for encoding generic values by reflection
	reflectBuf *buffer.Buffer
	reflectEnc zapcore.ReflectedEncoder

	// Additional fields for tracing
	appName     string
	traceID     string
	spanID      string
	parentSpanID string
	processID   int
	threadID    int
}

// EncoderConfig extends zapcore.EncoderConfig with additional options
type EncoderConfig struct {
	zapcore.EncoderConfig
	EnableCaller     bool
	EnableStacktrace bool
	FieldSeparator   string
}

func putLineEncoder(enc *LineEncoder) {
	if enc.reflectBuf != nil {
		enc.reflectBuf.Free()
	}
	enc.EncoderConfig = nil
	enc.buf = nil
	enc.reflectBuf = nil
	enc.reflectEnc = nil
	_linePool.Put(enc)
}

func defaultReflectedEncoder(w io.Writer) zapcore.ReflectedEncoder {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc
}

// NewLineEncoder creates a new LineEncoder with the given configuration
func NewLineEncoder(cfg zapcore.EncoderConfig, spaced bool) *LineEncoder {
	if cfg.SkipLineEnding {
		cfg.LineEnding = ""
	} else if cfg.LineEnding == "" {
		cfg.LineEnding = zapcore.DefaultLineEnding
	}
	if cfg.NewReflectedEncoder == nil {
		cfg.NewReflectedEncoder = defaultReflectedEncoder
	}
	return &LineEncoder{
		EncoderConfig: &cfg,
		buf:           _bufferPool.Get(),
		spaced:        spaced,
	}
}

// SetTraceInfo sets the trace information for this encoder
func (enc *LineEncoder) SetTraceInfo(appName, traceID, spanID, parentSpanID string, processID, threadID int) {
	enc.appName = appName
	enc.traceID = traceID
	enc.spanID = spanID
	enc.parentSpanID = parentSpanID
	enc.processID = processID
	enc.threadID = threadID
}

// Clone creates a copy of the encoder
func (enc *LineEncoder) Clone() zapcore.Encoder {
	clone := enc.clone()
	clone.buf.Write(enc.buf.Bytes())
	return clone
}

func (enc *LineEncoder) clone() *LineEncoder {
	clone := _linePool.Get().(*LineEncoder)
	clone.EncoderConfig = enc.EncoderConfig
	clone.buf = _bufferPool.Get()
	return clone
}

// EncodeEntry encodes a log entry and its fields into the line format
func (enc *LineEncoder) EncodeEntry(ent zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	final := enc.clone()

	// time
	final.AddTime(final.TimeKey, ent.Time)
	final.AppendString(enc.ConsoleSeparator)

	// level
	if final.LevelKey != "" && final.EncodeLevel != nil {
		cur := final.buf.Len()
		final.EncodeLevel(ent.Level, final)
		if cur == final.buf.Len() {
			final.AppendString(ent.Level.String())
		}
		final.AppendString(enc.ConsoleSeparator)
	}

	// app name (service)
	if enc.appName != "" {
		final.AppendString(enc.appName)
		final.AppendString(enc.ConsoleSeparator)
	}

	// trace information
	if enc.traceID != "" {
		final.AppendString(enc.traceID)
	}
	final.AppendString(enc.ConsoleSeparator)

	if enc.spanID != "" {
		final.AppendString(enc.spanID)
	}
	final.AppendString(enc.ConsoleSeparator)

	// parent span id (new field)
	if enc.parentSpanID != "" {
		final.AppendString(enc.parentSpanID)
	}
	final.AppendString(enc.ConsoleSeparator)

	// caller information
	if ent.Caller.Defined {
		if final.CallerKey != "" {
			final.addKey(final.CallerKey)
			cur := final.buf.Len()
			final.EncodeCaller(ent.Caller, final)
			if cur == final.buf.Len() {
				final.AppendString(ent.Caller.String())
			}
		}
		if final.FunctionKey != "" {
			final.addKey(final.FunctionKey)
			final.AppendString(ent.Caller.Function)
		}
		final.AppendString(enc.ConsoleSeparator)
	} else {
		final.AppendString(enc.ConsoleSeparator)
	}

	// process and thread info (optional)
	if enc.processID > 0 {
		final.AppendInt(enc.processID)
	}
	final.AppendString(enc.ConsoleSeparator)

	if enc.threadID > 0 {
		final.AppendInt(enc.threadID)
	}
	final.AppendString(enc.ConsoleSeparator)

	// message
	final.addKey(enc.MessageKey)
	final.AppendString(ent.Message)

	// additional fields
	addFields(final, fields)

	// stack trace
	final.AddString(final.StacktraceKey, ent.Stack)
	final.buf.AppendString(final.LineEnding)

	ret := final.buf
	putLineEncoder(final)
	return ret, nil
}

// Implement zapcore.Encoder interface methods

func (enc *LineEncoder) AddArray(key string, arr zapcore.ArrayMarshaler) error {
	enc.addKey(key)
	return enc.AppendArray(arr)
}

func (enc *LineEncoder) AddObject(key string, obj zapcore.ObjectMarshaler) error {
	enc.addKey(key)
	return enc.AppendObject(obj)
}

func (enc *LineEncoder) AddBinary(key string, val []byte) {
	enc.AddString(key, base64.StdEncoding.EncodeToString(val))
}

func (enc *LineEncoder) AddByteString(key string, val []byte) {
	enc.addKey(key)
	enc.AppendByteString(val)
}

func (enc *LineEncoder) AddBool(key string, val bool) {
	enc.addKey(key)
	enc.AppendBool(val)
}

func (enc *LineEncoder) AddComplex128(key string, val complex128) {
	enc.addKey(key)
	enc.AppendComplex128(val)
}

func (enc *LineEncoder) AddComplex64(key string, val complex64) {
	enc.addKey(key)
	enc.AppendComplex64(val)
}

func (enc *LineEncoder) AddDuration(key string, val time.Duration) {
	enc.addKey(key)
	enc.AppendDuration(val)
}

func (enc *LineEncoder) AddFloat64(key string, val float64) {
	enc.addKey(key)
	enc.AppendFloat64(val)
}

func (enc *LineEncoder) AddFloat32(key string, val float32) {
	enc.addKey(key)
	enc.AppendFloat32(val)
}

func (enc *LineEncoder) AddInt64(key string, val int64) {
	enc.addKey(key)
	enc.AppendInt64(val)
}

func (enc *LineEncoder) AddInt(k string, v int)         { enc.AddInt64(k, int64(v)) }
func (enc *LineEncoder) AddInt32(k string, v int32)     { enc.AddInt64(k, int64(v)) }
func (enc *LineEncoder) AddInt16(k string, v int16)     { enc.AddInt64(k, int64(v)) }
func (enc *LineEncoder) AddInt8(k string, v int8)       { enc.AddInt64(k, int64(v)) }
func (enc *LineEncoder) AddUint(k string, v uint)       { enc.AddUint64(k, uint64(v)) }
func (enc *LineEncoder) AddUint32(k string, v uint32)   { enc.AddUint64(k, uint64(v)) }
func (enc *LineEncoder) AddUint16(k string, v uint16)   { enc.AddUint64(k, uint64(v)) }
func (enc *LineEncoder) AddUint8(k string, v uint8)     { enc.AddUint64(k, uint64(v)) }
func (enc *LineEncoder) AddUintptr(k string, v uintptr) { enc.AddUint64(k, uint64(v)) }
func (enc *LineEncoder) AddUint64(key string, val uint64) {
	enc.addKey(key)
	enc.AppendUint64(val)
}

func (enc *LineEncoder) AddReflected(key string, obj interface{}) error {
	valueBytes, err := enc.encodeReflected(obj)
	if err != nil {
		return err
	}
	enc.addKey(key)
	_, err = enc.buf.Write(valueBytes)
	return err
}

func (enc *LineEncoder) AddString(key, val string) {
	enc.addKey(key)
	enc.AppendString(val)
}

func (enc *LineEncoder) AddTime(key string, val time.Time) {
	enc.addKey(key)
	enc.AppendTime(val)
}

func (enc *LineEncoder) OpenNamespace(key string) {
	enc.addKey(key)
	enc.buf.AppendByte('{')
}

// Append methods

func (enc *LineEncoder) AppendArray(arr zapcore.ArrayMarshaler) error {
	enc.buf.AppendByte('[')
	err := arr.MarshalLogArray(enc)
	enc.buf.AppendByte(']')
	return err
}

func (enc *LineEncoder) AppendObject(obj zapcore.ObjectMarshaler) error {
	return obj.MarshalLogObject(enc)
}

func (enc *LineEncoder) AppendBool(val bool) {
	enc.buf.AppendBool(val)
}

func (enc *LineEncoder) AppendByteString(val []byte) {
	enc.safeAddByteString(val)
}

func (enc *LineEncoder) AppendComplex128(v complex128) {
	enc.appendComplex(v, 64)
}

func (enc *LineEncoder) AppendComplex64(v complex64) {
	enc.appendComplex(complex128(v), 32)
}

func (enc *LineEncoder) AppendFloat64(v float64) {
	enc.appendFloat(v, 64)
}

func (enc *LineEncoder) AppendFloat32(v float32) {
	enc.appendFloat(float64(v), 32)
}

func (enc *LineEncoder) AppendInt(v int) {
	enc.AppendInt64(int64(v))
}

func (enc *LineEncoder) AppendInt64(val int64) {
	enc.buf.AppendInt(val)
}

func (enc *LineEncoder) AppendInt32(v int32) {
	enc.AppendInt64(int64(v))
}

func (enc *LineEncoder) AppendInt16(v int16) {
	enc.AppendInt64(int64(v))
}

func (enc *LineEncoder) AppendInt8(v int8) {
	enc.AppendInt64(int64(v))
}

func (enc *LineEncoder) AppendString(val string) {
	enc.safeAddString(val)
}

func (enc *LineEncoder) AppendTimeLayout(time time.Time, layout string) {
	enc.buf.AppendTime(time, layout)
}

func (enc *LineEncoder) AppendTime(val time.Time) {
	cur := enc.buf.Len()
	if e := enc.EncodeTime; e != nil {
		e(val, enc)
	}
	if cur == enc.buf.Len() {
		enc.AppendInt64(val.UnixNano())
	}
}

func (enc *LineEncoder) AppendUint(v uint) {
	enc.AppendUint64(uint64(v))
}

func (enc *LineEncoder) AppendUint64(val uint64) {
	enc.buf.AppendUint(val)
}

func (enc *LineEncoder) AppendUint32(v uint32) {
	enc.AppendUint64(uint64(v))
}

func (enc *LineEncoder) AppendUint16(v uint16) {
	enc.AppendUint64(uint64(v))
}

func (enc *LineEncoder) AppendUint8(v uint8) {
	enc.AppendUint64(uint64(v))
}

func (enc *LineEncoder) AppendUintptr(v uintptr) {
	enc.AppendUint64(uint64(v))
}

func (enc *LineEncoder) AppendDuration(val time.Duration) {
	cur := enc.buf.Len()
	if e := enc.EncodeDuration; e != nil {
		e(val, enc)
	}
	if cur == enc.buf.Len() {
		enc.AppendInt64(int64(val))
	}
}

func (enc *LineEncoder) AppendReflected(val interface{}) error {
	valueBytes, err := enc.encodeReflected(val)
	if err != nil {
		return err
	}
	_, err = enc.buf.Write(valueBytes)
	return err
}

// Helper methods

func (enc *LineEncoder) addKey(key string) {
	if key != "" && key != zapcore.OmitKey {
		enc.buf.AppendString(key)
	}
}

func (enc *LineEncoder) resetReflectBuf() {
	if enc.reflectBuf == nil {
		enc.reflectBuf = _bufferPool.Get()
		enc.reflectEnc = enc.NewReflectedEncoder(enc.reflectBuf)
	} else {
		enc.reflectBuf.Reset()
	}
}

func (enc *LineEncoder) encodeReflected(obj interface{}) ([]byte, error) {
	if obj == nil {
		return nullLiteralBytes, nil
	}
	enc.resetReflectBuf()
	if err := enc.reflectEnc.Encode(obj); err != nil {
		return nil, err
	}
	enc.reflectBuf.TrimNewline()
	return enc.reflectBuf.Bytes(), nil
}

func (enc *LineEncoder) appendComplex(val complex128, precision int) {
	r, i := float64(real(val)), float64(imag(val))
	enc.buf.AppendFloat(r, precision)
	if i >= 0 {
		enc.buf.AppendByte('+')
	}
	enc.buf.AppendFloat(i, precision)
	enc.buf.AppendByte('i')
}

func (enc *LineEncoder) appendFloat(val float64, bitSize int) {
	switch {
	case math.IsNaN(val):
		enc.buf.AppendString(`"NaN"`)
	case math.IsInf(val, 1):
		enc.buf.AppendString(`"+Inf"`)
	case math.IsInf(val, -1):
		enc.buf.AppendString(`"-Inf"`)
	default:
		enc.buf.AppendFloat(val, bitSize)
	}
}

func (enc *LineEncoder) safeAddByteString(s []byte) {
	for i := 0; i < len(s); {
		if enc.tryAddRuneSelf(s[i]) {
			i++
			continue
		}
		r, size := utf8.DecodeRune(s[i:])
		if enc.tryAddRuneError(r, size) {
			i++
			continue
		}
		enc.buf.Write(s[i : i+size])
		i += size
	}
}

func (enc *LineEncoder) safeAddString(s string) {
	for i := 0; i < len(s); {
		if enc.tryAddRuneSelf(s[i]) {
			i++
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		if enc.tryAddRuneError(r, size) {
			i++
			continue
		}
		enc.buf.AppendString(s[i : i+size])
		i += size
	}
}

func (enc *LineEncoder) tryAddRuneSelf(b byte) bool {
	if b >= utf8.RuneSelf {
		return false
	}
	if b >= 0x20 && b != '\\' && b != '"' {
		enc.buf.AppendByte(b)
		return true
	}
	switch b {
	case '\\', '"':
		enc.buf.AppendByte('\\')
		enc.buf.AppendByte(b)
	case '\n':
		enc.buf.AppendByte('\\')
		enc.buf.AppendByte('n')
	case '\r':
		enc.buf.AppendByte('\\')
		enc.buf.AppendByte('r')
	case '\t':
		enc.buf.AppendByte('\\')
		enc.buf.AppendByte('t')
	default:
		// Encode bytes < 0x20, except for the escape sequences above.
		enc.buf.AppendString(`\u00`)
		enc.buf.AppendByte(_hex[b>>4])
		enc.buf.AppendByte(_hex[b&0xF])
	}
	return true
}

func (enc *LineEncoder) tryAddRuneError(r rune, size int) bool {
	if r == utf8.RuneError && size == 1 {
		enc.buf.AppendString(`\ufffd`)
		return true
	}
	return false
}

func addFields(enc zapcore.ObjectEncoder, fields []zapcore.Field) {
	for i := range fields {
		fields[i].AddTo(enc)
	}
}
