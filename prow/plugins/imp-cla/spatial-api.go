package imp_cla

import (
	"crypto/x509"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

var (
	apiHost = "api.spatial.improbable.io:10104"
)

func GetApiConnection(serviceAccountFile string) (*grpc.ClientConn, error) {
	pool, err := x509.SystemCertPool()
	if err != nil {
		return nil, errors.Wrap(err, "faield to get cert pool")
	}
	creds := credentials.NewClientTLSFromCert(pool, "")
	perRPC, err := oauth.NewServiceAccountFromFile(serviceAccountFile)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create service account from file %s", serviceAccountFile)
	}
	conn, err := grpc.Dial(apiHost, grpc.WithTransportCredentials(creds), grpc.WithPerRPCCredentials(perRPC))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to connect to Improbable api at %s", apiHost)
	}

	return conn, nil
}
