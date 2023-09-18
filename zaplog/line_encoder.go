package zaplog

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

var _ zapcore.Encoder = (*lineEncoder)(nil)

var (
	_linePool = sync.Pool{
		New: func() interface{} {
			return &lineEncoder{}
		},
	}
	_bufferPool      = buffer.NewPool()
	nullLiteralBytes = []byte("null")
)

type lineEncoder struct {
	*zapcore.EncoderConfig
	buf    *buffer.Buffer
	spaced bool // include spaces after colons and commas

	// for encoding generic values by reflection
	reflectBuf *buffer.Buffer
	reflectEnc zapcore.ReflectedEncoder
}

func putLineEncoder(enc *lineEncoder) {
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

func NewLineEncoder(cfg zapcore.EncoderConfig, spaced bool) *lineEncoder {
	if cfg.SkipLineEnding {
		cfg.LineEnding = ""
	} else if cfg.LineEnding == "" {
		cfg.LineEnding = zapcore.DefaultLineEnding
	}
	if cfg.NewReflectedEncoder == nil {
		cfg.NewReflectedEncoder = defaultReflectedEncoder
	}
	return &lineEncoder{
		EncoderConfig: &cfg,
		buf:           _bufferPool.Get(),
		spaced:        spaced,
	}
}

func (enc *lineEncoder) AddArray(key string, arr zapcore.ArrayMarshaler) error {
	enc.addKey(key)
	return enc.AppendArray(arr)
}

func (enc *lineEncoder) AddObject(key string, obj zapcore.ObjectMarshaler) error {
	enc.addKey(key)
	return enc.AppendObject(obj)
}
func (enc *lineEncoder) AddBinary(key string, val []byte) {
	enc.AddString(key, base64.StdEncoding.EncodeToString(val))
}
func (enc *lineEncoder) AddByteString(key string, val []byte) {
	enc.addKey(key)
	enc.AppendByteString(val)
}
func (enc *lineEncoder) AddBool(key string, val bool) {
	enc.addKey(key)
	enc.AppendBool(val)
}
func (enc *lineEncoder) AddComplex128(key string, val complex128) {
	enc.addKey(key)
	enc.AppendComplex128(val)
}
func (enc *lineEncoder) AddComplex64(key string, val complex64) {
	enc.addKey(key)
	enc.AppendComplex64(val)
}
func (enc *lineEncoder) AddDuration(key string, val time.Duration) {
	enc.addKey(key)
	enc.AppendDuration(val)
}
func (enc *lineEncoder) AddFloat64(key string, val float64) {
	enc.addKey(key)
	enc.AppendFloat64(val)
}
func (enc *lineEncoder) AddFloat32(key string, val float32) {
	enc.addKey(key)
	enc.AppendFloat32(val)
}
func (enc *lineEncoder) AddInt64(key string, val int64) {
	enc.addKey(key)
	enc.AppendInt64(val)
}
func (enc *lineEncoder) resetReflectBuf() {
	if enc.reflectBuf == nil {
		enc.reflectBuf = _bufferPool.Get()
		enc.reflectEnc = enc.NewReflectedEncoder(enc.reflectBuf)
	} else {
		enc.reflectBuf.Reset()
	}
}
func (enc *lineEncoder) encodeReflected(obj interface{}) ([]byte, error) {
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
func (enc *lineEncoder) AddReflected(key string, obj interface{}) error {
	valueBytes, err := enc.encodeReflected(obj)
	if err != nil {
		return err
	}
	enc.addKey(key)
	_, err = enc.buf.Write(valueBytes)
	return err
}
func (enc *lineEncoder) OpenNamespace(key string) {
	enc.addKey(key)
	enc.buf.AppendByte('{')
}
func (enc *lineEncoder) AddString(key, val string) {
	enc.addKey(key)
	enc.AppendString(val)
}
func (enc *lineEncoder) AddTime(key string, val time.Time) {
	enc.addKey(key)
	enc.AppendTime(val)
}
func (enc *lineEncoder) AddUint64(key string, val uint64) {
	enc.addKey(key)
	enc.AppendUint64(val)
}
func (enc *lineEncoder) AppendArray(arr zapcore.ArrayMarshaler) error {
	enc.buf.AppendByte('[')
	err := arr.MarshalLogArray(enc)
	enc.buf.AppendByte(']')
	return err
}
func (enc *lineEncoder) AppendObject(obj zapcore.ObjectMarshaler) error {
	return obj.MarshalLogObject(enc)
}
func (enc *lineEncoder) AppendBool(val bool)         { enc.buf.AppendBool(val) }
func (enc *lineEncoder) AppendByteString(val []byte) { enc.safeAddByteString(val) }
func (enc *lineEncoder) appendComplex(val complex128, precision int) {
	r, i := float64(real(val)), float64(imag(val))
	enc.buf.AppendFloat(r, precision)
	if i >= 0 {
		enc.buf.AppendByte('+')
	}
	enc.buf.AppendFloat(i, precision)
	enc.buf.AppendByte('i')
}
func (enc *lineEncoder) AppendDuration(val time.Duration) {
	cur := enc.buf.Len()
	if e := enc.EncodeDuration; e != nil {
		e(val, enc)
	}
	if cur == enc.buf.Len() {
		enc.AppendInt64(int64(val))
	}
}
func (enc *lineEncoder) AppendInt64(val int64) { enc.buf.AppendInt(val) }
func (enc *lineEncoder) AppendReflected(val interface{}) error {
	valueBytes, err := enc.encodeReflected(val)
	if err != nil {
		return err
	}
	_, err = enc.buf.Write(valueBytes)
	return err
}
func (enc *lineEncoder) AppendString(val string) { enc.safeAddString(val) }
func (enc *lineEncoder) AppendTimeLayout(time time.Time, layout string) {
	enc.buf.AppendTime(time, layout)
}
func (enc *lineEncoder) AppendTime(val time.Time) {
	cur := enc.buf.Len()
	if e := enc.EncodeTime; e != nil {
		e(val, enc)
	}
	if cur == enc.buf.Len() {
		enc.AppendInt64(val.UnixNano())
	}
}
func (enc *lineEncoder) AppendUint64(val uint64)        { enc.buf.AppendUint(val) }
func (enc *lineEncoder) AddInt(k string, v int)         { enc.AddInt64(k, int64(v)) }
func (enc *lineEncoder) AddInt32(k string, v int32)     { enc.AddInt64(k, int64(v)) }
func (enc *lineEncoder) AddInt16(k string, v int16)     { enc.AddInt64(k, int64(v)) }
func (enc *lineEncoder) AddInt8(k string, v int8)       { enc.AddInt64(k, int64(v)) }
func (enc *lineEncoder) AddUint(k string, v uint)       { enc.AddUint64(k, uint64(v)) }
func (enc *lineEncoder) AddUint32(k string, v uint32)   { enc.AddUint64(k, uint64(v)) }
func (enc *lineEncoder) AddUint16(k string, v uint16)   { enc.AddUint64(k, uint64(v)) }
func (enc *lineEncoder) AddUint8(k string, v uint8)     { enc.AddUint64(k, uint64(v)) }
func (enc *lineEncoder) AddUintptr(k string, v uintptr) { enc.AddUint64(k, uint64(v)) }
func (enc *lineEncoder) AppendComplex64(v complex64)    { enc.appendComplex(complex128(v), 32) }
func (enc *lineEncoder) AppendComplex128(v complex128)  { enc.appendComplex(complex128(v), 64) }
func (enc *lineEncoder) AppendFloat64(v float64)        { enc.appendFloat(v, 64) }
func (enc *lineEncoder) AppendFloat32(v float32)        { enc.appendFloat(float64(v), 32) }
func (enc *lineEncoder) AppendInt(v int)                { enc.AppendInt64(int64(v)) }
func (enc *lineEncoder) AppendInt32(v int32)            { enc.AppendInt64(int64(v)) }
func (enc *lineEncoder) AppendInt16(v int16)            { enc.AppendInt64(int64(v)) }
func (enc *lineEncoder) AppendInt8(v int8)              { enc.AppendInt64(int64(v)) }
func (enc *lineEncoder) AppendUint(v uint)              { enc.AppendUint64(uint64(v)) }
func (enc *lineEncoder) AppendUint32(v uint32)          { enc.AppendUint64(uint64(v)) }
func (enc *lineEncoder) AppendUint16(v uint16)          { enc.AppendUint64(uint64(v)) }
func (enc *lineEncoder) AppendUint8(v uint8)            { enc.AppendUint64(uint64(v)) }
func (enc *lineEncoder) AppendUintptr(v uintptr)        { enc.AppendUint64(uint64(v)) }

func (enc *lineEncoder) appendFloat(val float64, bitSize int) {
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

func (enc *lineEncoder) clone() *lineEncoder {
	clone := _linePool.Get().(*lineEncoder)
	clone.EncoderConfig = enc.EncoderConfig
	clone.buf = _bufferPool.Get()
	return clone
}

func (enc *lineEncoder) Clone() zapcore.Encoder {
	clone := enc.clone()
	clone.buf.Write(enc.buf.Bytes())
	return clone
}

func (enc *lineEncoder) addKey(key string) {
	// ignore
	if key != "" {
		enc.buf.AppendString(key)
	}

}

func (enc *lineEncoder) safeAddByteString(s []byte) {
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

func (enc *lineEncoder) tryAddRuneSelf(b byte) bool {
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

func (enc *lineEncoder) tryAddRuneError(r rune, size int) bool {
	if r == utf8.RuneError && size == 1 {
		enc.buf.AppendString(`\ufffd`)
		return true
	}
	return false
}

func (enc *lineEncoder) addElementSeparator() {
	last := enc.buf.Len() - 1
	if last < 0 {
		return
	}
	switch enc.buf.Bytes()[last] {
	case '{', '[', ':', ',', ' ':
		return
	default:
		enc.buf.AppendByte(',')
		if enc.spaced {
			enc.buf.AppendByte(' ')
		}
	}
}

func (enc *lineEncoder) safeAddString(s string) {
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

func (enc *lineEncoder) EncodeEntry(ent zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
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

	// final log
	if enc.buf.Len() > 0 {
		final.buf.Write(enc.buf.Bytes())
	}

	// caller
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
	}
	// hierachy_id, ignore field
	final.AppendString(enc.ConsoleSeparator)

	// message
	final.addKey(enc.MessageKey)
	final.AppendString(ent.Message)

	// another fields
	addFields(final, fields)

	// strack
	final.AddString(final.StacktraceKey, ent.Stack)
	final.buf.AppendString(final.LineEnding)

	ret := final.buf
	putLineEncoder(final)
	return ret, nil
}

func addFields(enc zapcore.ObjectEncoder, fields []zapcore.Field) {
	for i := range fields {
		fields[i].AddTo(enc)
	}
}
