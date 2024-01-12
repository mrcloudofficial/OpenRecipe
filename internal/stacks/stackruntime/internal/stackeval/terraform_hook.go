// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package stackeval

import (
	"context"
	"sync"

	"github.com/hashicorp/terraform/internal/addrs"
	"github.com/hashicorp/terraform/internal/plans"
	"github.com/hashicorp/terraform/internal/stacks/stackaddrs"
	"github.com/hashicorp/terraform/internal/stacks/stackruntime/hooks"
	"github.com/hashicorp/terraform/internal/terraform"
	"github.com/zclconf/go-cty/cty"
)

// componentInstanceTerraformHook implements terraform.Hook for plan and apply
// operations on a specified component instance. It connects the standard
// terraform.Hook callbacks to the given stackruntime.Hooks callbacks.
//
// We unfortunately must embed a context.Context in this type, as the existing
// OpenRecipe core hook interface does not support threading a context through.
// The lifetime of this hook instance is strictly smaller than its surrounding
// context, but we should migrate away from this for clarity when possible.
type componentInstanceTerraformHook struct {
	terraform.NilHook

	ctx   context.Context
	seq   *hookSeq
	hooks *Hooks
	addr  stackaddrs.AbsComponentInstance

	mu                                 sync.Mutex
	resourceInstanceObjectApplyAction  addrs.Map[addrs.AbsResourceInstanceObject, plans.Action]
	resourceInstanceObjectApplySuccess addrs.Set[addrs.AbsResourceInstanceObject]
}

func (h *componentInstanceTerraformHook) resourceInstanceAddr(addr addrs.AbsResourceInstance) stackaddrs.AbsResourceInstance {
	return stackaddrs.AbsResourceInstance{
		Component: h.addr,
		Item:      addr,
	}
}

func (h *componentInstanceTerraformHook) resourceInstanceObjectAddr(riAddr addrs.AbsResourceInstance, dk addrs.DeposedKey) stackaddrs.AbsResourceInstanceObject {
	return stackaddrs.AbsResourceInstanceObject{
		Component: h.addr,
		Item: addrs.AbsResourceInstanceObject{
			ResourceInstance: riAddr,
			DeposedKey:       dk,
		},
	}
}

func (h *componentInstanceTerraformHook) PreDiff(addr addrs.AbsResourceInstance, dk addrs.DeposedKey, priorState, proposedNewState cty.Value) (terraform.HookAction, error) {
	hookMore(h.ctx, h.seq, h.hooks.ReportResourceInstanceStatus, &hooks.ResourceInstanceStatusHookData{
		Addr:   h.resourceInstanceObjectAddr(addr, dk),
		Status: hooks.ResourceInstancePlanning,
	})
	return terraform.HookActionContinue, nil
}

func (h *componentInstanceTerraformHook) PostDiff(addr addrs.AbsResourceInstance, dk addrs.DeposedKey, action plans.Action, priorState, plannedNewState cty.Value) (terraform.HookAction, error) {
	hookMore(h.ctx, h.seq, h.hooks.ReportResourceInstanceStatus, &hooks.ResourceInstanceStatusHookData{
		Addr:   h.resourceInstanceObjectAddr(addr, dk),
		Status: hooks.ResourceInstancePlanned,
	})
	return terraform.HookActionContinue, nil
}

func (h *componentInstanceTerraformHook) PreApply(addr addrs.AbsResourceInstance, dk addrs.DeposedKey, action plans.Action, priorState, plannedNewState cty.Value) (terraform.HookAction, error) {
	if action != plans.NoOp {
		hookMore(h.ctx, h.seq, h.hooks.ReportResourceInstanceStatus, &hooks.ResourceInstanceStatusHookData{
			Addr:   h.resourceInstanceObjectAddr(addr, dk),
			Status: hooks.ResourceInstanceApplying,
		})
	}

	h.mu.Lock()
	if h.resourceInstanceObjectApplyAction.Len() == 0 {
		h.resourceInstanceObjectApplyAction = addrs.MakeMap[addrs.AbsResourceInstanceObject, plans.Action]()
	}
	h.resourceInstanceObjectApplyAction.Put(
		addrs.AbsResourceInstanceObject{
			ResourceInstance: addr,
			DeposedKey:       dk,
		},
		action,
	)
	h.mu.Unlock()

	return terraform.HookActionContinue, nil
}

func (h *componentInstanceTerraformHook) PostApply(addr addrs.AbsResourceInstance, dk addrs.DeposedKey, newState cty.Value, err error) (terraform.HookAction, error) {
	objAddr := h.resourceInstanceObjectAddr(addr, dk)
	localObjAddr := addr.DeposedObject(dk)

	h.mu.Lock()
	action, ok := h.resourceInstanceObjectApplyAction.GetOk(localObjAddr)
	h.mu.Unlock()
	if !ok {
		// Weird, but we'll just tolerate it to be robust.
		return terraform.HookActionContinue, nil
	}

	if action == plans.NoOp {
		// We don't emit starting hooks for no-op changes and so we shouldn't
		// emit ending hooks for them either.
		return terraform.HookActionContinue, nil
	}

	status := hooks.ResourceInstanceApplied
	if err != nil {
		status = hooks.ResourceInstanceErrored
	} else {
		h.mu.Lock()
		if h.resourceInstanceObjectApplySuccess == nil {
			h.resourceInstanceObjectApplySuccess = addrs.MakeSet[addrs.AbsResourceInstanceObject]()
		}
		h.resourceInstanceObjectApplySuccess.Add(localObjAddr)
		h.mu.Unlock()
	}

	hookMore(h.ctx, h.seq, h.hooks.ReportResourceInstanceStatus, &hooks.ResourceInstanceStatusHookData{
		Addr:   objAddr,
		Status: status,
	})
	return terraform.HookActionContinue, nil
}

func (h *componentInstanceTerraformHook) PreProvisionInstanceStep(addr addrs.AbsResourceInstance, typeName string) (terraform.HookAction, error) {
	// NOTE: We assume provisioner events are always about the "current"
	// object for the given resource instance, because the hook API does
	// not include a DeposedKey argument in this case.
	hookMore(h.ctx, h.seq, h.hooks.ReportResourceInstanceProvisionerStatus, &hooks.ResourceInstanceProvisionerHookData{
		Addr:   h.resourceInstanceObjectAddr(addr, addrs.NotDeposed),
		Name:   typeName,
		Status: hooks.ProvisionerProvisioning,
	})
	return terraform.HookActionContinue, nil
}

func (h *componentInstanceTerraformHook) ProvisionOutput(addr addrs.AbsResourceInstance, typeName string, msg string) {
	// TODO: determine whether we should continue line splitting as we do with jsonHook

	// NOTE: We assume provisioner events are always about the "current"
	// object for the given resource instance, because the hook API does
	// not include a DeposedKey argument in this case.
	output := msg
	hookMore(h.ctx, h.seq, h.hooks.ReportResourceInstanceProvisionerStatus, &hooks.ResourceInstanceProvisionerHookData{
		Addr:   h.resourceInstanceObjectAddr(addr, addrs.NotDeposed),
		Name:   typeName,
		Status: hooks.ProvisionerProvisioning,
		Output: &output,
	})
}

func (h *componentInstanceTerraformHook) PostProvisionInstanceStep(addr addrs.AbsResourceInstance, typeName string, err error) (terraform.HookAction, error) {
	// NOTE: We assume provisioner events are always about the "current"
	// object for the given resource instance, because the hook API does
	// not include a DeposedKey argument in this case.
	status := hooks.ProvisionerProvisioned
	if err != nil {
		status = hooks.ProvisionerErrored
	}
	hookMore(h.ctx, h.seq, h.hooks.ReportResourceInstanceProvisionerStatus, &hooks.ResourceInstanceProvisionerHookData{
		Addr:   h.resourceInstanceObjectAddr(addr, addrs.NotDeposed),
		Name:   typeName,
		Status: status,
	})
	return terraform.HookActionContinue, nil
}

func (h *componentInstanceTerraformHook) ResourceInstanceObjectAppliedAction(addr addrs.AbsResourceInstanceObject) plans.Action {
	h.mu.Lock()
	ret, ok := h.resourceInstanceObjectApplyAction.GetOk(addr)
	h.mu.Unlock()
	if !ok {
		return plans.NoOp
	}
	return ret
}

func (h *componentInstanceTerraformHook) ResourceInstanceObjectsSuccessfullyApplied() addrs.Set[addrs.AbsResourceInstanceObject] {
	return h.resourceInstanceObjectApplySuccess
}
