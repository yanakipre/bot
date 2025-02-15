package clouderr

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/samber/lo"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ctxkey struct{}

var (
	logFieldsContextKey       ctxkey = struct{}{}
	logFieldKeysMapContextKey ctxkey = struct{}{}
)

// ContextWithFields lets you persiste log fields in context directly as a slice.
// These fields can then be retrieved, e.g. to enrich errors.
// Nested calls preserve parent log fields.
func ContextWithFields(ctx context.Context, fields ...zap.Field) context.Context {
	// dedupe log fields and persist in context
	lkm := logFieldKeyMapFromContext(ctx)
	lf := FieldsFromContext(ctx)
	for _, f := range fields {
		if idx, ok := lkm[f.Key]; ok {
			lf[idx] = f
		} else {
			lf = append(lf, f)
			lkm[f.Key] = len(lf) - 1
		}
	}
	return context.WithValue(
		context.WithValue(
			ctx,
			logFieldKeysMapContextKey,
			lkm,
		),
		logFieldsContextKey,
		lf,
	)
}

func FieldsFromContext(ctx context.Context) []zap.Field {
	if fields, ok := ctx.Value(logFieldsContextKey).([]zap.Field); ok {
		return fields
	}
	return nil
}

func logFieldKeyMapFromContext(ctx context.Context) map[string]int {
	if m, ok := ctx.Value(logFieldKeysMapContextKey).(map[string]int); ok {
		return m
	}
	return map[string]int{}
}

type ErrorWithFields struct { //nolint:errname
	error
	fields []zap.Field
}

func (err ErrorWithFields) Fields() []zap.Field {
	return err.fields
}

func (err ErrorWithFields) Unwrap() error {
	return err.error
}

// WithFields constructs error with fields
func WithFields(text string, fields ...zap.Field) error {
	return WrapWithFields(errors.New(text), fields...)
}

// WithFieldsFromContext constructs error with fields from context
func WithFieldsFromContext(ctx context.Context, text string, fields ...zap.Field) error {
	return WithFields(text, append(FieldsFromContext(ctx), fields...)...)
}

// WrapWithFieldsFromContext constructs error with fields from context that wraps passed error
func WrapWithFieldsFromContext(ctx context.Context, err error, fields ...zap.Field) error {
	return WrapWithFields(err, append(FieldsFromContext(ctx), fields...)...)
}

// WithFields constructs error with fields that wraps passed error
func WrapWithFields(err error, fields ...zap.Field) error {
	if err == nil {
		panic("cannot wrap nil error")
	}
	if len(fields) == 0 {
		return err
	}
	return ErrorWithFields{
		error:  err,
		fields: fields,
	}
}

// UnwrapFields returns a unique list of all fields in the error and wrapped/joined errors
// Field is unique if it's Key is unique
// In case of field keys collision, the field from the deepest-wrapped error wins
func UnwrapFields(err error) []zap.Field {
	fields := make(map[string]zap.Field)
	errs := []error{err}
	for len(errs) > 0 {
		e := errs[0]
		errs = errs[1:]
		if wf, ok := e.(interface{ Fields() []zap.Field }); ok {
			for _, f := range wf.Fields() {
				fields[f.Key] = f
			}
		}
		if uw, ok := e.(interface{ Unwrap() []error }); ok {
			errs = append(errs, uw.Unwrap()...)
		}
		if uw, ok := e.(interface{ Unwrap() error }); ok {
			errs = append(errs, uw.Unwrap())
		}
	}
	return lo.MapToSlice(fields, func(_ string, f zap.Field) zap.Field {
		return f
	})
}

// FieldToString returns field representation of form
// Key:"Value"
func FieldToString(f zap.Field) string {
	var v any

	switch f.Type {
	case zapcore.ArrayMarshalerType:
		v = f.Interface.(zapcore.ArrayMarshaler)
	case zapcore.ObjectMarshalerType:
		v = f.Interface.(zapcore.ObjectMarshaler)
	case zapcore.InlineMarshalerType:
		v = f.Interface.(zapcore.ObjectMarshaler)
	case zapcore.BinaryType:
		v = hex.EncodeToString(f.Interface.([]byte))
	case zapcore.BoolType:
		v = f.Integer == 1
	case zapcore.ByteStringType:
		v = string(f.Interface.([]byte))
	case zapcore.Complex128Type:
		v = f.Interface.(complex128)
	case zapcore.Complex64Type:
		v = f.Interface.(complex64)
	case zapcore.DurationType:
		v = time.Duration(f.Integer)
	case zapcore.Float64Type:
		v = math.Float64frombits(uint64(f.Integer))
	case zapcore.Float32Type:
		v = math.Float32frombits(uint32(f.Integer))
	case zapcore.Int64Type:
		v = f.Integer
	case zapcore.Int32Type:
		v = int32(f.Integer)
	case zapcore.Int16Type:
		v = int16(f.Integer)
	case zapcore.Int8Type:
		v = int8(f.Integer)
	case zapcore.StringType:
		v = f.String
	case zapcore.TimeType:
		if f.Interface != nil {
			v = time.Unix(0, f.Integer).In(f.Interface.(*time.Location))
		} else {
			// Fall back to UTC if location is nil.
			v = time.Unix(0, f.Integer)
		}
	case zapcore.TimeFullType:
		v = f.Interface.(time.Time)
	case zapcore.Uint64Type:
		v = uint64(f.Integer)
	case zapcore.Uint32Type:
		v = uint32(f.Integer)
	case zapcore.Uint16Type:
		v = uint16(f.Integer)
	case zapcore.Uint8Type:
		v = uint8(f.Integer)
	case zapcore.UintptrType:
		v = uintptr(f.Integer)
	case zapcore.ReflectType,
		zapcore.NamespaceType,
		zapcore.StringerType:
		v = f.Interface
	case zapcore.ErrorType:
		v = f.Interface.(error)
	case zapcore.SkipType:
		break
	default:
		panic(WithFields("unknown field type", zap.String("key", f.Key), zap.Int("type", int(f.Type))))
	}

	return fmt.Sprintf(`%s:"%v"`, f.Key, v)
}

// Errors creates a field containing an array of errors, preserving their fields
func Errors(key string, errs []error) zap.Field {
	return zap.Array(key, zapcore.ArrayMarshalerFunc(func(arr zapcore.ArrayEncoder) error {
		for _, err := range errs {
			if err == nil {
				continue
			}
			if appendErr := arr.AppendObject(zapcore.ObjectMarshalerFunc(func(enc zapcore.ObjectEncoder) error {
				enc.AddString("error", err.Error())
				for _, f := range UnwrapFields(err) {
					f.AddTo(enc)
				}
				return nil
			})); appendErr != nil {
				return appendErr
			}
		}
		return nil
	}))
}
