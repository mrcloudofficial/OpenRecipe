// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package terraform

import (
	"context"
	"sync"
	"time"

	"github.com/hashicorp/terraform/internal/addrs"
	"github.com/hashicorp/terraform/internal/checks"
	"github.com/hashicorp/terraform/internal/configs"
	"github.com/hashicorp/terraform/internal/configs/configschema"
	"github.com/hashicorp/terraform/internal/instances"
	"github.com/hashicorp/terraform/internal/moduletest/mocking"
	"github.com/hashicorp/terraform/internal/namedvals"
	"github.com/hashicorp/terraform/internal/plans"
	"github.com/hashicorp/terraform/internal/providers"
	"github.com/hashicorp/terraform/internal/provisioners"
	"github.com/hashicorp/terraform/internal/refactoring"
	"github.com/hashicorp/terraform/internal/states"
	"github.com/hashicorp/terraform/internal/tfdiags"
)

// ContextGraphWalker is the GraphWalker implementation used with the
// Context struct to walk and evaluate the graph.
type ContextGraphWalker struct {
	NullGraphWalker

	// Configurable values
	Context                 *Context
	State                   *states.SyncState   // Used for safe concurrent access to state
	RefreshState            *states.SyncState   // Used for safe concurrent access to state
	PrevRunState            *states.SyncState   // Used for safe concurrent access to state
	Changes                 *plans.ChangesSync  // Used for safe concurrent writes to changes
	Checks                  *checks.State       // Used for safe concurrent writes of checkable objects and their check results
	NamedValues             *namedvals.State    // Tracks evaluation of input variables, local values, and output values
	InstanceExpander        *instances.Expander // Tracks our gradual expansion of module and resource instances
	Imports                 []configs.Import
	MoveResults             refactoring.MoveResults // Read-only record of earlier processing of move statements
	Operation               walkOperation
	StopContext             context.Context
	ExternalProviderConfigs map[addrs.RootProviderConfig]providers.Interface
	Config                  *configs.Config
	PlanTimestamp           time.Time
	Overrides               *mocking.Overrides

	// This is an output. Do not set this, nor read it while a graph walk
	// is in progress.
	NonFatalDiagnostics tfdiags.Diagnostics

	once               sync.Once
	contexts           map[string]*BuiltinEvalContext
	contextLock        sync.Mutex
	providerCache      map[string]providers.Interface
	providerFuncCache  map[string]providers.Interface
	providerSchemas    map[string]providers.ProviderSchema
	providerLock       sync.Mutex
	provisionerCache   map[string]provisioners.Interface
	provisionerSchemas map[string]*configschema.Block
	provisionerLock    sync.Mutex
}

func (w *ContextGraphWalker) EnterPath(path addrs.ModuleInstance) EvalContext {
	w.contextLock.Lock()
	defer w.contextLock.Unlock()

	// If we already have a context for this path cached, use that
	key := path.String()
	if ctx, ok := w.contexts[key]; ok {
		return ctx
	}

	ctx := w.EvalContext().WithPath(path)
	w.contexts[key] = ctx.(*BuiltinEvalContext)
	return ctx
}

func (w *ContextGraphWalker) EvalContext() EvalContext {
	w.once.Do(w.init)

	// Our evaluator shares some locks with the main context and the walker
	// so that we can safely run multiple evaluations at once across
	// different modules.
	evaluator := &Evaluator{
		Meta:          w.Context.meta,
		Config:        w.Config,
		Operation:     w.Operation,
		State:         w.State,
		Changes:       w.Changes,
		Plugins:       w.Context.plugins,
		NamedValues:   w.NamedValues,
		PlanTimestamp: w.PlanTimestamp,
	}

	ctx := &BuiltinEvalContext{
		StopContext:             w.StopContext,
		Hooks:                   w.Context.hooks,
		InputValue:              w.Context.uiInput,
		InstanceExpanderValue:   w.InstanceExpander,
		Plugins:                 w.Context.plugins,
		ExternalProviderConfigs: w.ExternalProviderConfigs,
		MoveResultsValue:        w.MoveResults,
		ProviderCache:           w.providerCache,
		ProviderFuncCache:       w.providerFuncCache,
		ProviderInputConfig:     w.Context.providerInputConfig,
		ProviderLock:            &w.providerLock,
		ProvisionerCache:        w.provisionerCache,
		ProvisionerLock:         &w.provisionerLock,
		ChangesValue:            w.Changes,
		ChecksValue:             w.Checks,
		NamedValuesValue:        w.NamedValues,
		StateValue:              w.State,
		RefreshStateValue:       w.RefreshState,
		PrevRunStateValue:       w.PrevRunState,
		Evaluator:               evaluator,
		OverrideValues:          w.Overrides,
	}

	return ctx
}

func (w *ContextGraphWalker) init() {
	w.contexts = make(map[string]*BuiltinEvalContext)
	w.providerCache = make(map[string]providers.Interface)
	w.providerFuncCache = make(map[string]providers.Interface)
	w.providerSchemas = make(map[string]providers.ProviderSchema)
	w.provisionerCache = make(map[string]provisioners.Interface)
	w.provisionerSchemas = make(map[string]*configschema.Block)
}

func (w *ContextGraphWalker) Execute(ctx EvalContext, n GraphNodeExecutable) tfdiags.Diagnostics {
	// Acquire a lock on the semaphore
	w.Context.parallelSem.Acquire()
	defer w.Context.parallelSem.Release()

	return n.Execute(ctx, w.Operation)
}
