// Package redis implements primitives for yanakipre to use redis.
// Underlying libraries are not expected to be used, use these wrappers instead.
package redis

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/samber/lo"
	"go.uber.org/zap"

	"github.com/yanakipre/bot/internal/logger"
	"github.com/yanakipre/bot/internal/secret"
)

type AuthType string

const (
	// AuthPlain is a regular way to authenticate, supplying a URL with user:password in it
	AuthPlain AuthType = "plain"
	// AuthAWSElasticacheIRSA is IAM Roles for Service Accounts to authenticate in Elasticache.
	// Application obtains short-lived token from Simple Token Service.
	// Redis client establishes a TCP connection that is reset on the elasticache side each hour or so.
	// When connection is reset Redis client reconnects using CredentialsProvider with new token.
	//
	// The service account for the application must have permissions to connect,
	// and elasticache should allow connections for this service account.
	//
	// The following links describe in more details how it works:
	//
	// 1. https://community.aws/content/2ZCKrwaaaTglCCWISSaKv1d7bI3/using-iam-authentication-for-redis-on-aws
	// 2. https://github.com/build-on-aws/aws-redis-iam-auth-golang?tab=readme-ov-file
	//	contains an example for memory db in golang.
	// 3. https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/auth-iam.html
	//	contains some Java code examples.
	// 3. https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/IAM.IdentityBasedPolicies.html#iam-connect-policy
	// 4. https://docs.aws.amazon.com/emr/latest/EMR-on-EKS-DevelopmentGuide/setting-up-enable-IAM.html
	AuthAWSElasticacheIRSA AuthType = "aws_elasticache_irsa"
)

// Config for Redis client.
type Config struct {
	AuthType AuthType `yaml:"auth_type"`
	// AwsIRSA is required when AuthType is AuthAWSElasticacheIRSA
	AwsIRSA    *AWSIRSAConfig `yaml:"aws_elasticache_irsa,omitempty"`
	URL        secret.String  `yaml:"redis_url"`
	ClientName string         `yaml:"client_name"`
}

func (c Config) Validate() error {
	switch c.AuthType {
	case AuthAWSElasticacheIRSA:
		if c.AwsIRSA == nil {
			return fmt.Errorf(
				"chosen %q authentication requires 'aws_irsa' to be set",
				AuthAWSElasticacheIRSA,
			)
		}
		return c.AwsIRSA.Validate()
	case AuthPlain:
		return nil
	default:
		return fmt.Errorf("unknown auth type %q", c.AuthType)
	}
}

func DefaultConfig() Config {
	return Config{
		AuthType:   AuthPlain,
		AwsIRSA:    lo.ToPtr(DefaultAWSIRSAConfig()),
		URL:        secret.NewString("redis://10.30.42.54:6379"),
		ClientName: "yanakipre",
	}
}

// Redis is a Yanakipre Redis client.
// It is a wrapper around redis implementation.
type Redis struct {
	*redis.Client
	tokenProvider *awsIamTokenGenerator
}

type UniversalClient redis.UniversalClient

func (r *Redis) Ready(ctx context.Context) error {
	if r.tokenProvider != nil {
		if err := r.tokenProvider.Ready(ctx); err != nil {
			return fmt.Errorf("token provider is not ready: %w", err)
		}
	}
	resp := r.Ping(ctx)
	if err := resp.Err(); err != nil {
		return err
	}
	_, err := resp.Result()
	return err
}

func (r *Redis) Close() error {
	return r.Client.Close()
}

func New(cfg Config, shutdownCtx context.Context) (*Redis, error) {
	redis.SetLogger(&Log{})
	switch cfg.AuthType {
	case AuthPlain:
		options, err := redis.ParseURL(cfg.URL.Unmask())
		if err != nil {
			return nil, err
		}
		options.ClientName = cfg.ClientName
		c := redis.NewClient(options)
		return &Redis{Client: c}, nil
	case AuthAWSElasticacheIRSA:
		options, err := redis.ParseURL(cfg.URL.Unmask())
		if err != nil {
			return nil, fmt.Errorf("cannot parse redis URL: %w", err)
		}
		options.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
		tokenProvider, err := newAuthTokenGenerator(*cfg.AwsIRSA)
		if err != nil {
			return nil, fmt.Errorf("cannot create awsIamTokenGenerator: %w", err)
		}
		lg := logger.FromContext(shutdownCtx).Named("aws_irsa_creds_provider")
		options.CredentialsProvider = func() (username string, password string) {
			generated, err := tokenProvider.Generate(shutdownCtx)
			if err != nil {
				lg.Error("cannot generated creds", zap.Error(err))
			}
			return cfg.AwsIRSA.UserID, generated
		}
		c := redis.NewClient(options)
		return &Redis{Client: c, tokenProvider: tokenProvider}, nil
	default:
		return nil, fmt.Errorf("unsupported auth method: %q", cfg.AuthType)
	}
}
