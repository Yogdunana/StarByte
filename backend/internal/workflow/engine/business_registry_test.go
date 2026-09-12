package engine

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
)

type policyFixture struct {
	user uuid.UUID
	db   *gorm.DB
}

func (p *policyFixture) Resolve(context.Context, *model.FlowInstance, *FlowNode) ([]uuid.UUID, error) {
	return []uuid.UUID{p.user}, nil
}
func (p *policyFixture) ForTransaction(tx *gorm.DB) BusinessApprover {
	return &policyFixture{user: p.user, db: tx}
}
func TestBusinessPolicyRegistrySeparatesTypesAndTransactions(t *testing.T) {
	r := NewBusinessApprovers()
	a, b := uuid.New(), uuid.New()
	if err := r.Register("member_application", &policyFixture{user: a}); err != nil {
		t.Fatal(err)
	}
	if err := r.Register(TaskBusinessType, &policyFixture{user: b}); err != nil {
		t.Fatal(err)
	}
	if err := r.Register("member_application", &policyFixture{user: b}); err == nil {
		t.Fatal("registration replaced admission policy")
	}
	tx := &gorm.DB{}
	bound := r.ForTransaction(tx).(*BusinessApprovers)
	for kind, want := range map[string]uuid.UUID{"member_application": a, TaskBusinessType: b} {
		ids, err := bound.Resolve(context.Background(), &model.FlowInstance{BusinessType: kind}, &FlowNode{Config: map[string]interface{}{"businessType": kind}})
		if err != nil || len(ids) != 1 || ids[0] != want {
			t.Fatalf("wrong policy: %s %v %v", kind, ids, err)
		}
		if bound.resolvers[kind].(*policyFixture).db != tx || r.resolvers[kind].(*policyFixture).db != nil {
			t.Fatal("transaction mutated shared resolver")
		}
	}
	if _, err := r.Resolve(context.Background(), &model.FlowInstance{BusinessType: TaskBusinessType}, &FlowNode{Config: map[string]interface{}{"businessType": "member_application"}}); err == nil {
		t.Fatal("mismatched business accepted")
	}
	if _, err := r.Resolve(context.Background(), &model.FlowInstance{BusinessType: "unknown"}, &FlowNode{Config: map[string]interface{}{"businessType": "unknown"}}); err == nil {
		t.Fatal("unknown policy fell back")
	}
}
func taskGraphFixture() *FlowGraph {
	stages := []string{"start", "assignment", "execution", "review", "acceptance", "end"}
	g := &FlowGraph{Nodes: map[string]*FlowNode{}}
	for i, stage := range stages {
		kind := "approval"
		if i == 0 || i == 5 {
			kind = stage
		}
		g.Nodes[stage] = &FlowNode{ID: stage, Type: kind, Config: map[string]interface{}{"taskStage": stage, "businessType": TaskBusinessType, "assigneeStrategy": "business_role", "approvalType": "single"}}
		if i > 0 {
			g.Edges = append(g.Edges, &FlowEdge{Source: stages[i-1], Target: stage})
		}
	}
	return g
}
func TestTaskGraphCannotSkipReviewOrAcceptance(t *testing.T) {
	if err := ValidateBusinessDefinition(TaskDefinitionKey, taskGraphFixture()); err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func(*FlowGraph){
		func(g *FlowGraph) { g.Edges[2].Target = "end" },
		func(g *FlowGraph) { delete(g.Nodes, "review") },
		func(g *FlowGraph) { g.Nodes["acceptance"].Config["assigneeStrategy"] = "initiator" },
		func(g *FlowGraph) { g.Nodes["review"].Config["businessType"] = "member_application" },
		func(g *FlowGraph) { g.Nodes["review"].Config["approvalType"] = "any" },
	} {
		g := taskGraphFixture()
		mutate(g)
		if err := ValidateTaskDefinition(g); err == nil {
			t.Fatal("unsafe graph accepted")
		}
	}
}

func TestTaskTransferGraphsMatchPublishedKeys(t *testing.T) {
	for _, kind := range []string{"internal", "department", "center"} {
		key := TaskTransferDefinitionKey(kind)
		graph, err := ParseGraph(TaskTransferBPMN(kind))
		if err != nil {
			t.Fatal(err)
		}
		if err := ValidateBusinessDefinition(key, graph); err != nil {
			t.Fatalf("%s: %v", kind, err)
		}
	}
	if TaskTransferDefinitionKey("unknown") != "" || len(TaskTransferBPMN("unknown")) < 10 {
		t.Fatal("unknown transfer kind produced a real definition")
	}
}
