# Working with error values

## Enriching errors

Use static strings for error texts and fields from zap package to enrich error objects. Otherwise errlint linter won't
let you merge.

### How to enrich

If you want to enrich error object with variable data (like project_id and other ids, invalid input params, timestamps,
etc etc everything you used fmt.Sprintf for before), there are 2 constructors currently in clouderr package:

- `clouderr.WithFields(string, ...zap.Field)` error accepts string text for the new error and typed fields from `zap`
  package (like `zap.String`, `zap.Int` etc etc)
- `clouderr.WrapWithFields(error, ...zap.Field)` error wraps existing error object and accepts fields from `zap` package
- Also, constructors from semerr package (like `semerr.Internal` etc) now also accept fields from `zap` package

### Linting rules for error creation expressions/functions from semerr, clouderr, errors, fmt packages:

- `fmt.Sprintf` is prohibited
- String concatenation + is prohibited
- Any formating directives except for `%w` (wrap another error) inside `fmt.Errorf` are prohibited

### How to get enrichment data from error values

You can get a complete list of fields from the error, including all the fields in all of the wrapped errors using
`clouderr.UnwrapFields(error) []zap.Field`.

This list of fields will be logged with the error.

Also it will be formatted in a human-readable form whenever such an error is returned from the API presenter (this part
relies on `*semerr.Error.MessageWithFields() string` method, and should work as long as the presenter casts errors to
semerr, or wraps into `semerr`. All presenters by 2024-07-01 do that and should continue to do that.

## Handling panics

### There are also a few helper facilities to deal with panics, please use those:

- `semerr.UnwrapPanic(r interface{}) error` accepts the value from `recover()` and returns it if it is already an error,
  otherwise it wraps whatever value there is into `semerr.Internal` so you can always treat values from `recover()` in a
  standard way as regular `error` values
- `logger.Panic(r interface{})` logs values from `recover()` in a uniform way with `stack_trace` and such. It relies on
  `semerr.UnwrapPanic` and `clouderr.UnwrapFields` internally

## Examples

### Creating errors

```go
err := clouderr.WithFields("error message", zap.String("project_id", projectID))
```

```go
err := semerr.Internal("error message", zap.String("project_id", projectID))
```

### Wrapping errors

```go
err = clouderr.WrapWithFields(err, zap.String("project_id", projectID))
```

```go
err = clouderr.WrapWithFields(
    fmt.Errorf("additional error context message: %w", err),
    zap.String("project_id", projectID),
)
```

```go
err = semerr.WrapWithInternal(
    err
    "additional error context message",
    zap.String("project_id", projectID),
)
```

```go
err = semerr.WrapWithInternal(
    err
    "additional error context message",
    zap.String("project_id", projectID),
    zap.NamedError("app_error", appErr),
)
```

### Working with fields

```go
ff := clouderr.UnwrapFields(err)
idFields := lo.Filter(ff, func(f zap.Field, _ int) bool {
    return strings.HasSuffix(f.Key, "_id") || strings.HasSuffix(f.Key, "_ids")
})

ids := make([]string, 0, len(idFields))
for _, f := range idFields {
    if f.Type == zapcore.StringType {
        ids = append(ids, )
    }

    switch f.Type {
        case zapcore.StringType:
            ids = append(ids, f.String)
        case zapcore.StringerType:
            ids = append(ids, f.Interface.(interface{String()}).String())
        case zapcore.ArrayMarshalerType:
            if ss, ok := f.Interface.([]string); ok {
                ids = append(ids, ss...)
            }
    }
}
```

### Handling panic

```go
// simple handling
if p := recover(); p != nil {
    someLocalVar = "panicked"

    logger.Panic(p)
}
```

```go
// custom handling
if p := recover(); p != nil {
    logger.Panic(p)
    err := semerr.UnwrapPanic(err)

    logger.LogPanic(
        logger.FromContext(ctx),
        fmt.Errorf("additional error context message: %w", err),
        zap.String("some additional fields", "..."),
    )

    panic(p)
}
```
