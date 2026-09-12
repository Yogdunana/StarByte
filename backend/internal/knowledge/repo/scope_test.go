package repo

import (
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/knowledge/model"
	"github.com/google/uuid"
)

func TestVisible(t *testing.T) {
	author := uuid.New()
	pub := &model.Doc{AuthorID: author, Status: model.StatusPublished, Visibility: model.VisibilityPublic}
	authn := &model.Doc{AuthorID: author, Status: model.StatusPublished, Visibility: model.VisibilityAuthenticated}
	perm := &model.Doc{AuthorID: author, Status: model.StatusPublished, Visibility: model.VisibilityPermission, PermissionCode: "doc:secret"}
	draft := &model.Doc{AuthorID: author, Status: model.StatusDraft, Visibility: model.VisibilityPublic}

	anon := ReadableScope{}
	member := ReadableScope{UserID: uuid.New()}
	holder := ReadableScope{UserID: uuid.New(), Perms: []string{"doc:secret"}}
	staff := ReadableScope{Staff: true, UserID: uuid.New()}

	if !Visible(pub, anon) || Visible(authn, anon) || Visible(perm, anon) || Visible(draft, anon) {
		t.Fatal("anon window")
	}
	if !Visible(pub, member) || !Visible(authn, member) || Visible(perm, member) || Visible(draft, member) {
		t.Fatal("member window")
	}
	if !Visible(perm, holder) || !Visible(draft, ReadableScope{UserID: author}) {
		t.Fatal("perm / author window")
	}
	if !Visible(draft, staff) || !Visible(perm, staff) {
		t.Fatal("staff window")
	}
}
