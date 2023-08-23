// Code generated by smithy-go-codegen DO NOT EDIT.

package sts

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	internalauth "github.com/aws/aws-sdk-go-v2/internal/auth"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

// Returns a set of temporary credentials for an Amazon Web Services account or
// IAM user. The credentials consist of an access key ID, a secret access key, and
// a security token. Typically, you use GetSessionToken if you want to use MFA to
// protect programmatic calls to specific Amazon Web Services API operations like
// Amazon EC2 StopInstances . MFA-enabled IAM users must call GetSessionToken and
// submit an MFA code that is associated with their MFA device. Using the temporary
// security credentials that the call returns, IAM users can then make programmatic
// calls to API operations that require MFA authentication. An incorrect MFA code
// causes the API to return an access denied error. For a comparison of
// GetSessionToken with the other API operations that produce temporary
// credentials, see Requesting Temporary Security Credentials (https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_temp_request.html)
// and Comparing the Amazon Web Services STS API operations (https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_temp_request.html#stsapi_comparison)
// in the IAM User Guide. No permissions are required for users to perform this
// operation. The purpose of the sts:GetSessionToken operation is to authenticate
// the user using MFA. You cannot use policies to control authentication
// operations. For more information, see Permissions for GetSessionToken (https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_temp_control-access_getsessiontoken.html)
// in the IAM User Guide. Session Duration The GetSessionToken operation must be
// called by using the long-term Amazon Web Services security credentials of an IAM
// user. Credentials that are created by IAM users are valid for the duration that
// you specify. This duration can range from 900 seconds (15 minutes) up to a
// maximum of 129,600 seconds (36 hours), with a default of 43,200 seconds (12
// hours). Credentials based on account credentials can range from 900 seconds (15
// minutes) up to 3,600 seconds (1 hour), with a default of 1 hour. Permissions The
// temporary security credentials created by GetSessionToken can be used to make
// API calls to any Amazon Web Services service with the following exceptions:
//   - You cannot call any IAM API operations unless MFA authentication
//     information is included in the request.
//   - You cannot call any STS API except AssumeRole or GetCallerIdentity .
//
// The credentials that GetSessionToken returns are based on permissions
// associated with the IAM user whose credentials were used to call the operation.
// The temporary credentials have the same permissions as the IAM user. Although it
// is possible to call GetSessionToken using the security credentials of an Amazon
// Web Services account root user rather than an IAM user, we do not recommend it.
// If GetSessionToken is called using root user credentials, the temporary
// credentials have root user permissions. For more information, see Safeguard
// your root user credentials and don't use them for everyday tasks (https://docs.aws.amazon.com/IAM/latest/UserGuide/best-practices.html#lock-away-credentials)
// in the IAM User Guide For more information about using GetSessionToken to
// create temporary credentials, see Temporary Credentials for Users in Untrusted
// Environments (https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_temp_request.html#api_getsessiontoken)
// in the IAM User Guide.
func (c *Client) GetSessionToken(ctx context.Context, params *GetSessionTokenInput, optFns ...func(*Options)) (*GetSessionTokenOutput, error) {
	if params == nil {
		params = &GetSessionTokenInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "GetSessionToken", params, optFns, c.addOperationGetSessionTokenMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*GetSessionTokenOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type GetSessionTokenInput struct {

	// The duration, in seconds, that the credentials should remain valid. Acceptable
	// durations for IAM user sessions range from 900 seconds (15 minutes) to 129,600
	// seconds (36 hours), with 43,200 seconds (12 hours) as the default. Sessions for
	// Amazon Web Services account owners are restricted to a maximum of 3,600 seconds
	// (one hour). If the duration is longer than one hour, the session for Amazon Web
	// Services account owners defaults to one hour.
	DurationSeconds *int32

	// The identification number of the MFA device that is associated with the IAM
	// user who is making the GetSessionToken call. Specify this value if the IAM user
	// has a policy that requires MFA authentication. The value is either the serial
	// number for a hardware device (such as GAHT12345678 ) or an Amazon Resource Name
	// (ARN) for a virtual device (such as arn:aws:iam::123456789012:mfa/user ). You
	// can find the device for an IAM user by going to the Amazon Web Services
	// Management Console and viewing the user's security credentials. The regex used
	// to validate this parameter is a string of characters consisting of upper- and
	// lower-case alphanumeric characters with no spaces. You can also include
	// underscores or any of the following characters: =,.@:/-
	SerialNumber *string

	// The value provided by the MFA device, if MFA is required. If any policy
	// requires the IAM user to submit an MFA code, specify this value. If MFA
	// authentication is required, the user must provide a code when requesting a set
	// of temporary security credentials. A user who fails to provide the code receives
	// an "access denied" response when requesting resources that require MFA
	// authentication. The format for this parameter, as described by its regex
	// pattern, is a sequence of six numeric digits.
	TokenCode *string

	noSmithyDocumentSerde
}

// Contains the response to a successful GetSessionToken request, including
// temporary Amazon Web Services credentials that can be used to make Amazon Web
// Services requests.
type GetSessionTokenOutput struct {

	// The temporary security credentials, which include an access key ID, a secret
	// access key, and a security (or session) token. The size of the security token
	// that STS API operations return is not fixed. We strongly recommend that you make
	// no assumptions about the maximum size.
	Credentials *types.Credentials

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationGetSessionTokenMiddlewares(stack *middleware.Stack, options Options) (err error) {
	err = stack.Serialize.Add(&awsAwsquery_serializeOpGetSessionToken{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsAwsquery_deserializeOpGetSessionToken{}, middleware.After)
	if err != nil {
		return err
	}
	if err = addlegacyEndpointContextSetter(stack, options); err != nil {
		return err
	}
	if err = addSetLoggerMiddleware(stack, options); err != nil {
		return err
	}
	if err = awsmiddleware.AddClientRequestIDMiddleware(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddComputeContentLengthMiddleware(stack); err != nil {
		return err
	}
	if err = addResolveEndpointMiddleware(stack, options); err != nil {
		return err
	}
	if err = v4.AddComputePayloadSHA256Middleware(stack); err != nil {
		return err
	}
	if err = addRetryMiddlewares(stack, options); err != nil {
		return err
	}
	if err = addHTTPSignerV4Middleware(stack, options); err != nil {
		return err
	}
	if err = awsmiddleware.AddRawResponseToMetadata(stack); err != nil {
		return err
	}
	if err = awsmiddleware.AddRecordResponseTiming(stack); err != nil {
		return err
	}
	if err = addClientUserAgent(stack, options); err != nil {
		return err
	}
	if err = smithyhttp.AddErrorCloseResponseBodyMiddleware(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddCloseResponseBodyMiddleware(stack); err != nil {
		return err
	}
	if err = addGetSessionTokenResolveEndpointMiddleware(stack, options); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opGetSessionToken(options.Region), middleware.Before); err != nil {
		return err
	}
	if err = awsmiddleware.AddRecursionDetection(stack); err != nil {
		return err
	}
	if err = addRequestIDRetrieverMiddleware(stack); err != nil {
		return err
	}
	if err = addResponseErrorMiddleware(stack); err != nil {
		return err
	}
	if err = addRequestResponseLogging(stack, options); err != nil {
		return err
	}
	if err = addendpointDisableHTTPSMiddleware(stack, options); err != nil {
		return err
	}
	return nil
}

func newServiceMetadataMiddleware_opGetSessionToken(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		SigningName:   "sts",
		OperationName: "GetSessionToken",
	}
}

type opGetSessionTokenResolveEndpointMiddleware struct {
	EndpointResolver EndpointResolverV2
	BuiltInResolver  builtInParameterResolver
}

func (*opGetSessionTokenResolveEndpointMiddleware) ID() string {
	return "ResolveEndpointV2"
}

func (m *opGetSessionTokenResolveEndpointMiddleware) HandleSerialize(ctx context.Context, in middleware.SerializeInput, next middleware.SerializeHandler) (
	out middleware.SerializeOutput, metadata middleware.Metadata, err error,
) {
	if awsmiddleware.GetRequiresLegacyEndpoints(ctx) {
		return next.HandleSerialize(ctx, in)
	}

	req, ok := in.Request.(*smithyhttp.Request)
	if !ok {
		return out, metadata, fmt.Errorf("unknown transport type %T", in.Request)
	}

	if m.EndpointResolver == nil {
		return out, metadata, fmt.Errorf("expected endpoint resolver to not be nil")
	}

	params := EndpointParameters{}

	m.BuiltInResolver.ResolveBuiltIns(&params)

	var resolvedEndpoint smithyendpoints.Endpoint
	resolvedEndpoint, err = m.EndpointResolver.ResolveEndpoint(ctx, params)
	if err != nil {
		return out, metadata, fmt.Errorf("failed to resolve service endpoint, %w", err)
	}

	req.URL = &resolvedEndpoint.URI

	for k := range resolvedEndpoint.Headers {
		req.Header.Set(
			k,
			resolvedEndpoint.Headers.Get(k),
		)
	}

	authSchemes, err := internalauth.GetAuthenticationSchemes(&resolvedEndpoint.Properties)
	if err != nil {
		var nfe *internalauth.NoAuthenticationSchemesFoundError
		if errors.As(err, &nfe) {
			// if no auth scheme is found, default to sigv4
			signingName := "sts"
			signingRegion := m.BuiltInResolver.(*builtInResolver).Region
			ctx = awsmiddleware.SetSigningName(ctx, signingName)
			ctx = awsmiddleware.SetSigningRegion(ctx, signingRegion)

		}
		var ue *internalauth.UnSupportedAuthenticationSchemeSpecifiedError
		if errors.As(err, &ue) {
			return out, metadata, fmt.Errorf(
				"This operation requests signer version(s) %v but the client only supports %v",
				ue.UnsupportedSchemes,
				internalauth.SupportedSchemes,
			)
		}
	}

	for _, authScheme := range authSchemes {
		switch authScheme.(type) {
		case *internalauth.AuthenticationSchemeV4:
			v4Scheme, _ := authScheme.(*internalauth.AuthenticationSchemeV4)
			var signingName, signingRegion string
			if v4Scheme.SigningName == nil {
				signingName = "sts"
			} else {
				signingName = *v4Scheme.SigningName
			}
			if v4Scheme.SigningRegion == nil {
				signingRegion = m.BuiltInResolver.(*builtInResolver).Region
			} else {
				signingRegion = *v4Scheme.SigningRegion
			}
			if v4Scheme.DisableDoubleEncoding != nil {
				// The signer sets an equivalent value at client initialization time.
				// Setting this context value will cause the signer to extract it
				// and override the value set at client initialization time.
				ctx = internalauth.SetDisableDoubleEncoding(ctx, *v4Scheme.DisableDoubleEncoding)
			}
			ctx = awsmiddleware.SetSigningName(ctx, signingName)
			ctx = awsmiddleware.SetSigningRegion(ctx, signingRegion)
			break
		case *internalauth.AuthenticationSchemeV4A:
			v4aScheme, _ := authScheme.(*internalauth.AuthenticationSchemeV4A)
			if v4aScheme.SigningName == nil {
				v4aScheme.SigningName = aws.String("sts")
			}
			if v4aScheme.DisableDoubleEncoding != nil {
				// The signer sets an equivalent value at client initialization time.
				// Setting this context value will cause the signer to extract it
				// and override the value set at client initialization time.
				ctx = internalauth.SetDisableDoubleEncoding(ctx, *v4aScheme.DisableDoubleEncoding)
			}
			ctx = awsmiddleware.SetSigningName(ctx, *v4aScheme.SigningName)
			ctx = awsmiddleware.SetSigningRegion(ctx, v4aScheme.SigningRegionSet[0])
			break
		case *internalauth.AuthenticationSchemeNone:
			break
		}
	}

	return next.HandleSerialize(ctx, in)
}

func addGetSessionTokenResolveEndpointMiddleware(stack *middleware.Stack, options Options) error {
	return stack.Serialize.Insert(&opGetSessionTokenResolveEndpointMiddleware{
		EndpointResolver: options.EndpointResolverV2,
		BuiltInResolver: &builtInResolver{
			Region:       options.Region,
			UseDualStack: options.EndpointOptions.UseDualStackEndpoint,
			UseFIPS:      options.EndpointOptions.UseFIPSEndpoint,
			Endpoint:     options.BaseEndpoint,
		},
	}, "ResolveEndpoint", middleware.After)
}
