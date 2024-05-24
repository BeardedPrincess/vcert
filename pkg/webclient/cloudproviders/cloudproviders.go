package cloudproviders

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Khan/genqlient/graphql"
	"github.com/google/uuid"

	"github.com/Venafi/vcert/v5/pkg/domain"
	"github.com/Venafi/vcert/v5/pkg/util"
)

//go:generate go run -mod=mod github.com/Khan/genqlient genqlient.yaml

type CloudProvidersClient struct {
	graphqlClient graphql.Client
}

func NewCloudProvidersClient(url string, httpClient *http.Client) *CloudProvidersClient {
	return &CloudProvidersClient{
		graphqlClient: graphql.NewClient(url, httpClient),
	}
}

func (c *CloudProvidersClient) GetCloudProvider(ctx context.Context, request domain.GetCloudProviderRequest) (*domain.CloudProvider, error) {
	if request.Name == "" {
		return nil, fmt.Errorf("cloud provider name cannot be empty")
	}

	status := cloudProviderStatusFromDomain(request.Status)
	providerType, err := cloudProviderTypeFromDomain(request.Type)
	if err != nil {
		return nil, err
	}

	resp, err := GetCloudProviders(ctx, c.graphqlClient, &status, &providerType, request.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve Cloud Provider with name %s: %w", request.Name, err)
	}
	if resp == nil || resp.GetCloudProviders() == nil || len(resp.GetCloudProviders().Nodes) != 1 {
		return nil, fmt.Errorf("could not find Cloud Provider with name %s", request.Name)
	}

	cp := resp.GetCloudProviders().Nodes[0]

	statusDetails := ""
	if cp.GetStatusDetails() != nil {
		statusDetails = *cp.GetStatusDetails()
	}

	return &domain.CloudProvider{
		ID:             cp.GetId(),
		Name:           cp.GetName(),
		Type:           string(cp.GetType()),
		Status:         string(cp.GetStatus()),
		StatusDetails:  statusDetails,
		KeystoresCount: cp.GetKeystoresCount(),
	}, nil
}

func (c *CloudProvidersClient) GetCloudKeystore(ctx context.Context, request domain.GetCloudKeystoreRequest) (*domain.CloudKeystore, error) {
	if request.CloudKeystoreID == nil {
		if request.CloudKeystoreName == nil || (request.CloudProviderID == nil && request.CloudProviderName == nil) {
			return nil, fmt.Errorf("following combinations are accepted for provisioning: keystore ID, or both provider Name and keystore Name, or both provider ID and keystore Name")
		}
	}

	resp, err := GetCloudKeystores(ctx, c.graphqlClient, request.CloudKeystoreID, request.CloudKeystoreName, request.CloudProviderID, request.CloudProviderName)
	msg := util.GetKeystoreOptionsString(request.CloudProviderID, request.CloudKeystoreID, request.CloudProviderName, request.CloudKeystoreName)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve Cloud Keystore with %s: %w", msg, err)
	}

	if resp == nil || resp.CloudKeystores == nil {
		return nil, fmt.Errorf("could not find keystore with %s", msg)
	}

	if len(resp.CloudKeystores.Nodes) != 1 {
		return nil, fmt.Errorf("could not find keystore with with %s", msg)
	}

	ck := resp.CloudKeystores.Nodes[0]

	return &domain.CloudKeystore{
		ID:                     ck.GetId(),
		Name:                   ck.GetName(),
		Type:                   string(ck.GetType()),
		MachineIdentitiesCount: ck.MachineIdentitiesCount,
	}, nil
}

func (c *CloudProvidersClient) GetMachineIdentity(ctx context.Context, request domain.GetCloudMachineIdentityRequest) (*domain.CloudMachineIdentity, error) {
	if request.MachineIdentityID == nil {
		return nil, fmt.Errorf("machine identity ID missing")
	}

	resp, err := GetMachineIdentities(ctx, c.graphqlClient, request.KeystoreID, request.MachineIdentityID, request.Fingerprints, request.NewlyDiscovered, request.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve cloud machine identity with id %s: %w", *request.MachineIdentityID, err)
	}
	if len(resp.CloudMachineIdentities.Nodes) != 1 {
		return nil, fmt.Errorf("could not find cloud machine identity with with ID %s", *request.MachineIdentityID)
	}

	mi := resp.CloudMachineIdentities.Nodes[0]

	return mi.toDomain()
}

func (c *CloudProvidersClient) ProvisionCertificate(ctx context.Context, certificateID string, cloudKeystoreID string, wsClientID string, options *CertificateProvisioningOptionsInput) (*domain.ProvisioningResponse, error) {
	if certificateID == "" {
		return nil, fmt.Errorf("certificateID cannot be empty")
	}
	if cloudKeystoreID == "" {
		return nil, fmt.Errorf("cloudKeystoreID cannot be empty")
	}
	if wsClientID == "" {
		return nil, fmt.Errorf("wsClientID cannot be empty")
	}
	resp, err := ProvisionCertificate(ctx, c.graphqlClient, certificateID, cloudKeystoreID, wsClientID, options)
	if err != nil {
		return nil, fmt.Errorf("failed to provision certificate with certificate ID %s, keystore ID %s and websocket ID %s: %w", certificateID, cloudKeystoreID, wsClientID, err)
	}

	if resp == nil || resp.ProvisionToCloudKeystore == nil {
		return nil, fmt.Errorf("failed to provision certificate with certificate ID %s, keystore ID %s and websocket ID %s", certificateID, cloudKeystoreID, wsClientID)
	}

	return &domain.ProvisioningResponse{
		WorkflowId:   resp.GetProvisionToCloudKeystore().GetWorkflowId(),
		WorkflowName: resp.GetProvisionToCloudKeystore().GetWorkflowName(),
	}, nil
}

func (v *GetMachineIdentitiesCloudMachineIdentitiesMachineIdentityConnectionNodesMachineIdentity) toDomain() (*domain.CloudMachineIdentity, error) {
	id, err := uuid.Parse(v.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to parse cloud machine identity id %s: %w", v.Id, err)
	}
	keystoreID, err := uuid.Parse(v.CloudKeystoreId)
	if err != nil {
		return nil, fmt.Errorf("failed to parse cloud key store id %s: %w", v.CloudKeystoreId, err)
	}
	certificateID, err := uuid.Parse(v.CertificateId)
	if err != nil {
		return nil, fmt.Errorf("failed to parse cloud certificate id %s: %w", v.CertificateId, err)
	}
	providerIDStr := ""
	if v.CloudProviderId != nil {
		providerIDStr = *v.CloudProviderId
	}
	providerID, err := uuid.Parse(providerIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse cloud provider id %s: %w", providerIDStr, err)
	}

	keystoreName := ""
	if v.CloudKeystoreName != nil {
		keystoreName = *v.CloudKeystoreName
	}
	providerName := ""
	if v.CloudProviderName != nil {
		providerName = *v.CloudProviderName
	}
	statusDetails := ""
	if v.StatusDetails != nil {
		statusDetails = *v.StatusDetails
	}
	metadata, err := v.metadataToDomain()
	if err != nil {
		return nil, fmt.Errorf("failed to parse cloud certificate metadata: %w", err)
	}

	return &domain.CloudMachineIdentity{
		ID:                id,
		CloudKeystoreID:   keystoreID,
		CloudKeystoreName: keystoreName,
		CloudProviderID:   providerID,
		CloudProviderName: providerName,
		CertificateID:     certificateID,
		Metadata:          metadata,
		Status:            v.Status.toDomain(),
		StatusDetails:     statusDetails,
	}, nil
}

func (mis MachineIdentityStatus) toDomain() domain.MachineIdentityStatus {
	switch mis {
	case MachineIdentityStatusNew:
		return domain.MachineIdentityStatusNew
	case MachineIdentityStatusPending:
		return domain.MachineIdentityStatusPending
	case MachineIdentityStatusInstalled:
		return domain.MachineIdentityStatusInstalled
	case MachineIdentityStatusDiscovered:
		return domain.MachineIdentityStatusDiscovered
	case MachineIdentityStatusValidated:
		return domain.MachineIdentityStatusValidated
	case MachineIdentityStatusMissing:
		return domain.MachineIdentityStatusMissing
	case MachineIdentityStatusFailed:
		return domain.MachineIdentityStatusFailed
	default:
		return domain.MachineIdentityStatusUnknown
	}
}

func (v *GetMachineIdentitiesCloudMachineIdentitiesMachineIdentityConnectionNodesMachineIdentity) metadataToDomain() (*domain.CertificateCloudMetadata, error) {
	if v.Metadata == nil {
		return nil, nil
	}
	m := *v.Metadata

	data, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal cloud certificate metadata: %w", err)
	}

	values := make(map[string]interface{})
	err = json.Unmarshal(data, &values)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal cloud certificate metadata: %w", err)
	}

	certMetadata := domain.NewCertificateCloudMetadata(values)
	return &certMetadata, nil
}

func cloudProviderStatusFromDomain(status domain.CloudProviderStatus) CloudProviderStatus {
	switch status {
	case domain.CloudProviderStatusValidated:
		return CloudProviderStatusValidated
	case domain.CloudProviderStatusNotValidated:
		return CloudProviderStatusNotValidated
	default:
		return CloudProviderStatusNotValidated
	}
}

func cloudProviderTypeFromDomain(providerType domain.CloudProviderType) (CloudProviderType, error) {
	switch providerType {
	case domain.CloudProviderTypeAWS:
		return CloudProviderTypeAws, nil
	case domain.CloudProviderTypeAzure:
		return CloudProviderTypeAzure, nil
	case domain.CloudProviderTypeGCP:
		return CloudProviderTypeGcp, nil
	default:
		return "UNKNOWN", fmt.Errorf("failed to determine cloud provider type for %s", providerType)
	}
}
