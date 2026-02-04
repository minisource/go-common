package grpcclient

import (
	"context"
	"fmt"
	"time"

	"github.com/minisource/go-common/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Client is a reusable gRPC client with retry, logging, and error handling
type Client struct {
	conn        *grpc.ClientConn
	logger      logging.Logger
	retryConfig RetryConfig
	target      string
	serviceName string
}

// Config holds gRPC client configuration
type Config struct {
	Target             string
	ServiceName        string
	RetryConfig        RetryConfig
	Logger             logging.Logger
	Interceptors       []grpc.UnaryClientInterceptor
	StreamInterceptors []grpc.StreamClientInterceptor
}

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxRetries     int
	InitialDelay   time.Duration
	MaxDelay       time.Duration
	BackoffFactor  float64
	RetryableCodes []codes.Code
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:    3,
		InitialDelay:  500 * time.Millisecond,
		MaxDelay:      5 * time.Second,
		BackoffFactor: 2.0,
		RetryableCodes: []codes.Code{
			codes.Unavailable,
			codes.ResourceExhausted,
			codes.DeadlineExceeded,
			codes.Internal,
		},
	}
}

// NewClient creates a new gRPC client
func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	if cfg.RetryConfig.MaxRetries == 0 {
		cfg.RetryConfig = DefaultRetryConfig()
	}

	// Add logging interceptor first
	interceptors := []grpc.UnaryClientInterceptor{
		createLoggingInterceptor(cfg.Logger, cfg.ServiceName),
	}
	interceptors = append(interceptors, cfg.Interceptors...)

	// Add retry interceptor last
	interceptors = append(interceptors, createRetryInterceptor(cfg.Logger, cfg.ServiceName, cfg.RetryConfig))

	// Add logging stream interceptor
	streamInterceptors := []grpc.StreamClientInterceptor{
		createStreamLoggingInterceptor(cfg.Logger, cfg.ServiceName),
	}
	streamInterceptors = append(streamInterceptors, cfg.StreamInterceptors...)

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(interceptors...),
		grpc.WithChainStreamInterceptor(streamInterceptors...),
	}

	conn, err := grpc.DialContext(ctx, cfg.Target, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", cfg.Target, err)
	}

	cfg.Logger.Info(logging.General, logging.ExternalService, "gRPC connection established", map[logging.ExtraKey]interface{}{
		"service": cfg.ServiceName,
		"target":  cfg.Target,
	})

	return &Client{
		conn:        conn,
		logger:      cfg.Logger,
		retryConfig: cfg.RetryConfig,
		target:      cfg.Target,
		serviceName: cfg.ServiceName,
	}, nil
}

// Conn returns the underlying gRPC connection
func (c *Client) Conn() *grpc.ClientConn {
	return c.conn
}

// Close closes the gRPC connection
func (c *Client) Close() error {
	c.logger.Info(logging.General, logging.ExternalService, "Closing gRPC connection", map[logging.ExtraKey]interface{}{
		"service": c.serviceName,
	})
	return c.conn.Close()
}

// createLoggingInterceptor creates a unary interceptor for logging
func createLoggingInterceptor(logger logging.Logger, serviceName string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		startTime := time.Now()

		logger.Debug(logging.General, logging.ExternalService, "Starting gRPC request", map[logging.ExtraKey]interface{}{
			"service": serviceName,
			"method":  method,
		})

		err := invoker(ctx, method, req, reply, cc, opts...)

		duration := time.Since(startTime)
		if err != nil {
			st, _ := status.FromError(err)
			logger.Error(logging.General, logging.ExternalService, "gRPC request failed", map[logging.ExtraKey]interface{}{
				"service":  serviceName,
				"method":   method,
				"duration": duration.String(),
				"code":     st.Code().String(),
				"error":    err.Error(),
			})
		} else {
			logger.Info(logging.General, logging.ExternalService, "gRPC request completed", map[logging.ExtraKey]interface{}{
				"service":  serviceName,
				"method":   method,
				"duration": duration.String(),
			})
		}

		return err
	}
}

// createStreamLoggingInterceptor creates a stream interceptor for logging
func createStreamLoggingInterceptor(logger logging.Logger, serviceName string) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		logger.Debug(logging.General, logging.ExternalService, "Starting gRPC stream", map[logging.ExtraKey]interface{}{
			"service": serviceName,
			"method":  method,
		})

		stream, err := streamer(ctx, desc, cc, method, opts...)
		if err != nil {
			st, _ := status.FromError(err)
			logger.Error(logging.General, logging.ExternalService, "gRPC stream failed", map[logging.ExtraKey]interface{}{
				"service": serviceName,
				"method":  method,
				"code":    st.Code().String(),
				"error":   err.Error(),
			})
		}

		return stream, err
	}
}

// createRetryInterceptor creates a unary interceptor for retry logic
func createRetryInterceptor(logger logging.Logger, serviceName string, cfg RetryConfig) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		var lastErr error

		for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
			if attempt > 0 {
				delay := calculateBackoff(cfg, attempt)
				logger.Debug(logging.General, logging.ExternalService, "Retrying gRPC request", map[logging.ExtraKey]interface{}{
					"service": serviceName,
					"method":  method,
					"attempt": attempt,
					"delay":   delay.String(),
				})

				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(delay):
				}
			}

			err := invoker(ctx, method, req, reply, cc, opts...)
			if err == nil {
				if attempt > 0 {
					logger.Info(logging.General, logging.ExternalService, "gRPC request succeeded after retry", map[logging.ExtraKey]interface{}{
						"service": serviceName,
						"method":  method,
						"attempt": attempt + 1,
					})
				}
				return nil
			}

			lastErr = err
			st, _ := status.FromError(err)

			if !shouldRetryCode(st.Code(), cfg.RetryableCodes) {
				logger.Debug(logging.General, logging.ExternalService, "gRPC error is not retryable", map[logging.ExtraKey]interface{}{
					"service": serviceName,
					"method":  method,
					"code":    st.Code().String(),
				})
				return err
			}

			logger.Warn(logging.General, logging.ExternalService, "gRPC request failed, will retry", map[logging.ExtraKey]interface{}{
				"service": serviceName,
				"method":  method,
				"attempt": attempt + 1,
				"code":    st.Code().String(),
				"error":   err.Error(),
			})
		}

		logger.Error(logging.General, logging.ExternalService, "gRPC request failed after retries", map[logging.ExtraKey]interface{}{
			"service":  serviceName,
			"method":   method,
			"attempts": cfg.MaxRetries + 1,
		})

		return NewServiceUnavailableError(serviceName, lastErr)
	}
}

func calculateBackoff(cfg RetryConfig, attempt int) time.Duration {
	delay := float64(cfg.InitialDelay) * pow(cfg.BackoffFactor, float64(attempt-1))
	if delay > float64(cfg.MaxDelay) {
		return cfg.MaxDelay
	}
	return time.Duration(delay)
}

func shouldRetryCode(code codes.Code, retryableCodes []codes.Code) bool {
	for _, c := range retryableCodes {
		if c == code {
			return true
		}
	}
	return false
}

func pow(base, exp float64) float64 {
	result := 1.0
	for i := 0; i < int(exp); i++ {
		result *= base
	}
	return result
}

// BearerAuthInterceptor creates an interceptor that adds bearer token to requests
func BearerAuthInterceptor(token string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// BearerAuthStreamInterceptor creates a stream interceptor that adds bearer token to requests
func BearerAuthStreamInterceptor(token string) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
		return streamer(ctx, desc, cc, method, opts...)
	}
}
