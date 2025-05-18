package type_helper

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
)

// Constraints for supported integer types
type SignedInteger interface {
	~int | ~int32 | ~int64
}

type UnsignedInteger interface {
	~uint | ~uint32 | ~uint64
}

type AnyInteger interface {
	SignedInteger | UnsignedInteger
}

// ParseError represents a custom error for failed parsing attempts.
type ParseError[T AnyInteger] struct {
	Input string
	Err   error
}

func (e *ParseError[T]) Error() string {
	return fmt.Sprintf("error parsing %T from string %q: %v", *new(T), e.Input, e.Err)
}

func (e *ParseError[T]) Unwrap() error {
	return e.Err
}

// ParseIntegerFromString parses a string into any supported signed or unsigned integer type.
func ParseIntegerFromString[T AnyInteger](s string) (T, error) {
	var zero T
	var bitSize int

	switch any(zero).(type) {
	case int32, uint32:
		bitSize = 32
	case int64, uint64:
		bitSize = 64
	default:
		// int or uint (architecture dependent)
		bitSize = 0
	}

	switch any(zero).(type) {
	case int, int32, int64:
		val, err := strconv.ParseInt(s, 10, bitSize)
		if err != nil {
			slog.Error("Error parsing signed integer", slog.String("string", s), slog.String("target_type", fmt.Sprintf("%T", zero)))
			return zero, &ParseError[T]{Input: s, Err: err}
		}
		return T(val), nil

	case uint, uint32, uint64:
		val, err := strconv.ParseUint(s, 10, bitSize)
		if err != nil {
			slog.Error("Error parsing unsigned integer", slog.String("string", s), slog.String("target_type", fmt.Sprintf("%T", zero)))
			return zero, &ParseError[T]{Input: s, Err: err}
		}
		return T(val), nil

	default:
		return zero, errors.New("unsupported integer type")
	}
}
