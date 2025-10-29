package registry

import (
	"github.com/PROJECT_NAME/internal/errors"
	"github.com/nayla-finance/go-nayla/clients/rest/kyc"
	"github.com/nayla-finance/go-nayla/clients/rest/los"
)

func (r *Registry) InitializeClients() error {
	var err error

	r.kycClient, err = kyc.NewClient(
		kyc.WithBaseURL(r.Config().KYC.BaseURL),
		kyc.WithAPIKey(r.Config().KYC.APIKey),
		kyc.WithLogger(r.Logger()),
		kyc.WithErrorResponseMapper(func(er kyc.ErrorResponse) error {
			// Do any mapping here
			return r.NewError(errors.ErrInternal, er.Message)
		}),
	)
	if err != nil {
		return err
	}

	r.losClient, err = los.NewClient(
		los.WithBaseURL(r.Config().LOS.BaseURL),
		los.WithAPIKey(r.Config().LOS.APIKey),
		los.WithLogger(r.Logger()),
		los.WithErrorResponseMapper(func(er los.ErrorResponse) error {
			// Do any mapping here
			return r.NewError(errors.ErrInternal, er.Message)
		}),
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *Registry) KYCClient() kyc.Client {
	return r.kycClient
}

func (r *Registry) LOSClient() los.Client {
	return r.losClient
}
