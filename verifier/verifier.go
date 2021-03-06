// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package verifier

import (
	"fmt"
	"github.com/hashicorp/go-plugin"

	"veraison/common"
	"veraison/endorsement"
	"veraison/policy"
)

type Verifier struct {
	pm        *policy.PolicyManager
	em        *endorsement.EndorsementManager
	pe        common.IPolicyEngine
	rpcClient plugin.ClientProtocol
	client    *plugin.Client
}

func NewVerifier() (*Verifier, error) {
	v := new(Verifier)

	v.pm = policy.NewPolicyManager()
	v.em = endorsement.NewEndorsementManager()

	return v, nil
}

func (v *Verifier) Initialize(vc VerifierConfig) error {
	v.em.InitializeStore(vc.PluginLocations, vc.EndorsementStoreName, vc.EndorsementStoreParams)
	v.pm.InitializeStore(vc.PluginLocations, vc.PolicyStoreName, vc.PolicyStoreParams)

	engineName := common.Canonize(vc.PolicyEngineName)

	lp, err := common.LoadPlugin(vc.PluginLocations, "policyengine", engineName)
	if err != nil {
		return err
	}

	v.pe = lp.Raw.(common.IPolicyEngine)
	v.client = lp.PluginClient
	v.rpcClient = lp.RpcClient

	if v.client == nil {
		return fmt.Errorf("Failed to find policy engine with name '%v'", engineName)
	}

	err = v.pe.Init(vc.PolicyEngineParams)
	if err != nil {
		v.client.Kill()
		return err
	}

	return nil
}

func (v *Verifier) Verify(ec *common.EvidenceContext, simple bool) (*common.AttestationResult, error) {
	policy, err := v.pm.GetPolicy(ec.TenantID, ec.Format)
	if err != nil {
		return nil, err
	}

	if err = v.pe.LoadPolicy(policy.Rules); err != nil {
		return nil, err
	}

	qds, err := policy.GetQueryDesriptors(ec.Evidence, common.QcNone)
	if err != nil {
		return nil, err
	}

	matches, err := v.em.GetEndorsements(qds...)
	if err != nil {
		return nil, err
	}

	endorsements := make(map[string]interface{})
	for name, qr := range matches {
		if len(qr) == 1 {
			endorsements[name] = qr[0]
		} else if len(qr) == 0 {
			return nil, fmt.Errorf("No matches for '%v'", name)
		} else {
			return nil, fmt.Errorf("Too many matches for '%v'", name)
		}
	}

	result := new(common.AttestationResult)

	if err := v.pe.GetAttetationResult(ec.Evidence, endorsements, simple, result); err != nil {
		return nil, err
	}

	return result, nil
}

func (v *Verifier) Close() {
	v.em.Close()
	v.pm.Close()
	v.client.Kill()
	v.rpcClient.Close()
}
