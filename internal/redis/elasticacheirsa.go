package redis

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"go.uber.org/zap"

	"github.com/yanakipre/bot/internal/encodingtooling"
	"github.com/yanakipre/bot/internal/logger"
)

const (
	// autodetectRegion is a word for the config to instruct the module
	// to get the region of the Elasticache from aws.Config automatically.
	autodetectRegion = "auto-detect-region"
	connectAction    = "connect"

	// If the request has no payload you should use the hex encoded SHA-256 of an empty string as the payloadHash value.
	hexEncodedSHA256EmptyString = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
)

// awsIamTokenGenerator generates signed AWS tokens to access Elasticache.
type awsIamTokenGenerator struct {
	cfg             AWSIRSAConfig
	aws             aws.Config
	targetAWSRegion string
	req             *http.Request

	signer *v4.Signer
}

func newAuthTokenGenerator(cfg AWSIRSAConfig) (*awsIamTokenGenerator, error) {
	queryParams := url.Values{
		"Action":        {connectAction},
		"User":          {cfg.UserID},
		"X-Amz-Expires": {strconv.FormatInt(int64(cfg.TokenLifeSpan.Duration.Seconds()), 10)},
		//"ResourceType":  {"ServerlessCache"}, - use it for serverless cache.
	}

	authURL := url.URL{
		Host:     cfg.ClusterName,
		Scheme:   "http",
		Path:     "/",
		RawQuery: queryParams.Encode(),
	}

	req, err := http.NewRequest(http.MethodGet, authURL.String(), nil)
	if err != nil {
		return &awsIamTokenGenerator{}, err
	}

	return &awsIamTokenGenerator{
		cfg:    cfg,
		req:    req,
		signer: v4.NewSigner(),
	}, nil
}

func (atg *awsIamTokenGenerator) Ready(ctx context.Context) error {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("cannot load AWS config: %w", err)
	}

	region := atg.cfg.Region
	if region == autodetectRegion {
		region = cfg.Region
	}
	// if no region was set or autodetected, fail
	if region == "" {
		return fmt.Errorf("region cannot be empty")
	}
	atg.targetAWSRegion = region
	logger.Debug(ctx, "aws config is loaded", zap.String("region", atg.targetAWSRegion))

	credentials, err := cfg.Credentials.Retrieve(ctx)
	if err != nil {
		return fmt.Errorf("cannot creds from generator: %w", err)
	}

	if credentials.AccessKeyID == "" || credentials.SecretAccessKey == "" {
		return fmt.Errorf("credentials are empty")
	}
	atg.aws = cfg
	return nil
}

func (atg *awsIamTokenGenerator) Generate(ctx context.Context) (string, error) {
	credentials, err := atg.aws.Credentials.Retrieve(ctx)
	if err != nil {
		return "", fmt.Errorf("cannot retrieve AWS credentials: %w", err)
	}

	signedURL, _, err := atg.signer.PresignHTTP(
		ctx,
		credentials,
		atg.req,
		hexEncodedSHA256EmptyString,
		atg.cfg.ServiceName,
		atg.targetAWSRegion,
		time.Now().UTC(),
	)
	if err != nil {
		return "", fmt.Errorf("request signing failed - %w", err)
	}

	signedURL = strings.Replace(signedURL, "http://", "", 1)

	return signedURL, nil
}

type AWSIRSAConfig struct {
	Region                    string `yaml:"region"`
	ElasticacheClusterEnabled bool   `yaml:"-"`
	// How long it is allowed to establish connection with this token, after application got it.
	TokenLifeSpan encodingtooling.Duration `yaml:"token_life_span"`
	ServiceName   string                   `yaml:"service_name"`
	ClusterName   string                   `yaml:"cluster_name"`
	// UserID is not AWS IAM starting with "arn:"!
	// It's "User ID" that is configured in elasticache, for example "yanakipre"
	UserID string `yaml:"user_id"`
}

func (c AWSIRSAConfig) Validate() error {
	if c.ElasticacheClusterEnabled {
		// supporting cluster configurations SHOULD be simple:
		// use redis.NewClusterClient instead of redis.NewClient
		// but it was not tested.
		return fmt.Errorf("cluster enabled configurations are not implemented")
	}
	return nil
}

func DefaultAWSIRSAConfig() AWSIRSAConfig {
	return AWSIRSAConfig{
		Region:                    autodetectRegion,
		ElasticacheClusterEnabled: false,
		// "The IAM authentication token is valid for 15 minutes"
		// https://docs.aws.amazon.com/memorydb/latest/devguide/auth-iam.html#auth-iam-limits
		TokenLifeSpan: encodingtooling.NewDuration(time.Second * 900),
		UserID:        "yanakipre",
		ServiceName:   "elasticache",
		ClusterName:   "regional-redis",
	}
}
