package response

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// TranslateGORMError converts a GORM error into a domain-level AppError.
//
// Mapping:
//   - gorm.ErrRecordNotFound  → AppError(404 Not Found)
//   - gorm.ErrDuplicatedKey   → AppError(409 Conflict)
//   - nil                     → nil
//   - anything else           → wrapped internal error
//
// Usage in repo layer:
//
//	user, err := s.repo.GetByID(ctx, id)
//	if err != nil {
//	    return nil, response.TranslateGORMError(err)
//	}
func TranslateGORMError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return NewNotFoundError("资源")
	}

	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return NewConflictError("资源已存在")
	}

	// Wrap unrecognised GORM errors as internal errors so the caller
	// gets a 500 response with the original error preserved for logging.
	return fmt.Errorf("database error: %w", err)
}
